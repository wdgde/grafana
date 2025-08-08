package recordingrule

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	model "github.com/grafana/grafana/apps/alerting/rules/pkg/apis/alerting/v0alpha1"
	"github.com/grafana/grafana/pkg/apimachinery/utils"
	"github.com/grafana/grafana/pkg/services/apiserver/endpoints/request"
	ngmodels "github.com/grafana/grafana/pkg/services/ngalert/models"
	"github.com/grafana/grafana/pkg/util"
	prom_model "github.com/prometheus/common/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var (
	errInvalidRule = fmt.Errorf("rule is not a recording rule")
)

func ConvertToK8sResource(
	orgID int64,
	rule *ngmodels.AlertRule,
	namespaceMapper request.NamespaceMapper,
) (*model.RecordingRule, error) {
	if rule.Type() != ngmodels.RuleTypeRecording {
		return nil, errInvalidRule
	}
	k8sRule := &model.RecordingRule{
		ObjectMeta: metav1.ObjectMeta{
			UID:       types.UID(rule.UID),
			Name:      rule.UID,
			Namespace: namespaceMapper(orgID),
			Labels:    make(map[string]string),
		},
		Spec: model.RecordingRuleSpec{
			Title:  rule.Title,
			Paused: util.Pointer(rule.IsPaused),
			Data:   make(map[string]model.RecordingRuleQuery),
			Trigger: model.RecordingRuleIntervalTrigger{
				Interval: model.RecordingRulePromDuration(strconv.FormatInt(rule.IntervalSeconds, 10)),
			},
			Labels: make(map[string]model.RecordingRuleTemplateString),

			Metric:              rule.Record.Metric,
			TargetDatasourceUID: rule.Record.TargetDatasourceUID,
		},
	}

	if rule.RuleGroup != "" && !ngmodels.IsNoGroupRuleGroup(rule.RuleGroup) {
		k8sRule.Labels["group"] = rule.RuleGroup
	}

	for k, v := range rule.Labels {
		k8sRule.Spec.Labels[k] = model.RecordingRuleTemplateString(v)
	}

	for _, query := range rule.Data {
		k8sQuery := model.RecordingRuleQuery{
			QueryType:     query.QueryType,
			Model:         query.Model,
			DatasourceUID: model.RecordingRuleDatasourceUID(query.DatasourceUID),
			Source:        util.Pointer(rule.Condition == query.RefID),
		}
		if time.Duration(query.RelativeTimeRange.From) > 0 || time.Duration(query.RelativeTimeRange.To) > 0 {
			k8sQuery.RelativeTimeRange = &model.RecordingRuleRelativeTimeRange{
				From: model.RecordingRulePromDurationWMillis(query.RelativeTimeRange.From.String()),
				To:   model.RecordingRulePromDurationWMillis(query.RelativeTimeRange.To.String()),
			}
		}
		k8sRule.Spec.Data[query.RefID] = k8sQuery
	}

	// TODO: add the common metadata fields
	meta, err := utils.MetaAccessor(k8sRule)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}
	meta.SetFolder(rule.NamespaceUID)
	if rule.UpdatedBy != nil {
		meta.SetUpdatedBy(string(*rule.UpdatedBy))
		k8sRule.SetUpdatedBy(string(*rule.UpdatedBy))
	}
	meta.SetUpdatedTimestamp(&rule.Updated)
	k8sRule.SetUpdateTimestamp(rule.Updated)

	// FIXME: we don't have a creation timestamp in the domain model, so we can't set it here.
	// We should consider adding it to the domain model. Migration can set it to the Updated timestamp for existing
	// k8sRule.SetCreationTimestamp(rule.)

	return k8sRule, nil
}

func ConvertToK8sResources(
	orgID int64,
	rules []*ngmodels.AlertRule,
	namespaceMapper request.NamespaceMapper,
	continueToken string,
) (*model.RecordingRuleList, error) {
	k8sRules := &model.RecordingRuleList{
		ListMeta: metav1.ListMeta{
			Continue: continueToken,
		},
		Items: make([]model.RecordingRule, 0, len(rules)),
	}
	for _, rule := range rules {
		k8sRule, err := ConvertToK8sResource(orgID, rule, namespaceMapper)
		if err != nil {
			return nil, fmt.Errorf("failed to convert to k8s resource: %w", err)
		}
		k8sRules.Items = append(k8sRules.Items, *k8sRule)
	}
	return k8sRules, nil
}

func ConvertToDomainModel(orgID int64, k8sRule *model.RecordingRule) (*ngmodels.AlertRule, error) {
	if k8sRule.UID != types.UID(k8sRule.Name) {
		return nil, fmt.Errorf("object name (%s) does not match object UID (%s)", k8sRule.Name, k8sRule.UID)
	}
	domainRule := &ngmodels.AlertRule{
		OrgID:    orgID,
		UID:      string(k8sRule.UID),
		Title:    k8sRule.Spec.Title,
		Data:     make([]ngmodels.AlertQuery, 0, len(k8sRule.Spec.Data)),
		IsPaused: k8sRule.Spec.Paused != nil && *k8sRule.Spec.Paused,
		Labels:   make(map[string]string),

		Record: &ngmodels.Record{
			Metric:              k8sRule.Spec.Metric,
			TargetDatasourceUID: k8sRule.Spec.TargetDatasourceUID,
		},
	}

	meta, err := utils.MetaAccessor(k8sRule)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}

	domainRule.NamespaceUID = meta.GetFolder()

	interval, err := strconv.Atoi(string(k8sRule.Spec.Trigger.Interval))
	if err != nil {
		return nil, fmt.Errorf("failed to parse interval: %w", err)
	}
	domainRule.IntervalSeconds = int64(interval)

	for k, v := range k8sRule.Spec.Labels {
		domainRule.Labels[k] = string(v)
	}
	for refID, query := range k8sRule.Spec.Data {
		modelJson, err := json.Marshal(query.Model)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal model: %w", err)
		}
		domainQuery := ngmodels.AlertQuery{
			RefID:         refID,
			QueryType:     query.QueryType,
			DatasourceUID: string(query.DatasourceUID),
			Model:         modelJson,
		}
		if query.RelativeTimeRange != nil {
			from, err := prom_model.ParseDuration(string(query.RelativeTimeRange.From))
			if err != nil {
				return nil, fmt.Errorf("failed to parse duration: %w", err)
			}
			to, err := prom_model.ParseDuration(string(query.RelativeTimeRange.To))
			if err != nil {
				return nil, fmt.Errorf("failed to parse duration: %w", err)
			}
			domainQuery.RelativeTimeRange = ngmodels.RelativeTimeRange{
				From: ngmodels.Duration(from),
				To:   ngmodels.Duration(to),
			}
		}

		domainRule.Data = append(domainRule.Data, domainQuery)

		if query.Source != nil && *query.Source {
			if domainRule.Condition != "" {
				return nil, fmt.Errorf("multiple queries marked as source: %s and %s", domainRule.Condition, refID)
			}
			domainRule.Condition = refID
		}
	}
	return domainRule, nil
}
