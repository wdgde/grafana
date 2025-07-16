package correlations

import (
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"net/http"
	"time"

	authlib "github.com/grafana/authlib/types"
	correlation "github.com/grafana/grafana/apps/correlation/pkg/apis/correlation/v0alpha1"
	"github.com/grafana/grafana/pkg/api/response"
	"github.com/grafana/grafana/pkg/api/routing"
	"github.com/grafana/grafana/pkg/apimachinery/utils"
	"github.com/grafana/grafana/pkg/middleware"
	ac "github.com/grafana/grafana/pkg/services/accesscontrol"
	grafanaapiserver "github.com/grafana/grafana/pkg/services/apiserver"
	"github.com/grafana/grafana/pkg/services/apiserver/endpoints/request"
	gapiutil "github.com/grafana/grafana/pkg/services/apiserver/utils"
	contextmodel "github.com/grafana/grafana/pkg/services/contexthandler/model"
	"github.com/grafana/grafana/pkg/services/datasources"
	"github.com/grafana/grafana/pkg/services/featuremgmt"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/grafana/grafana/pkg/util/errhttp"
	"github.com/grafana/grafana/pkg/web"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

func (s *CorrelationsService) registerAPIEndpoints() {
	uidScope := datasources.ScopeProvider.GetResourceScopeUID(ac.Parameter(":uid"))
	authorize := ac.Middleware(s.AccessControl)

	s.RouteRegister.Get("/api/datasources/correlations", middleware.ReqSignedIn, authorize(ac.EvalPermission(datasources.ActionRead)), routing.Wrap(s.getCorrelationsHandler))

	s.RouteRegister.Group("/api/datasources/uid/:uid/correlations", func(entities routing.RouteRegister) {
		entities.Get("/", authorize(ac.EvalPermission(datasources.ActionRead)), routing.Wrap(s.getCorrelationsBySourceUIDHandler))
		entities.Post("/", authorize(ac.EvalPermission(datasources.ActionWrite, uidScope)), routing.Wrap(s.createHandler))

		entities.Group("/:correlationUID", func(entities routing.RouteRegister) {
			entities.Get("/", authorize(ac.EvalPermission(datasources.ActionRead)), routing.Wrap(s.getCorrelationHandler))
			entities.Delete("/", authorize(ac.EvalPermission(datasources.ActionWrite, uidScope)), routing.Wrap(s.deleteHandler))
			entities.Patch("/", authorize(ac.EvalPermission(datasources.ActionWrite, uidScope)), routing.Wrap(s.updateHandler))
		})
	}, middleware.ReqSignedIn)
}

// swagger:route POST /datasources/uid/{sourceUID}/correlations correlations createCorrelation
//
// Add correlation.
//
// Responses:
// 200: createCorrelationResponse
// 400: badRequestError
// 401: unauthorisedError
// 403: forbiddenError
// 404: notFoundError
// 500: internalServerError
func (s *CorrelationsService) createHandler(c *contextmodel.ReqContext) response.Response {
	if s.featuremgmt.IsEnabledGlobally(featuremgmt.FlagKubernetesCorrelations) {
		s.k8sHandler.createCorrelation(c)
		return nil
	}

	cmd := CreateCorrelationCommand{}
	if err := web.Bind(c.Req, &cmd); err != nil {
		return response.Error(http.StatusBadRequest, "bad request data", err)
	}
	cmd.SourceUID = web.Params(c.Req)[":uid"]
	cmd.OrgId = c.GetOrgID()

	correlation, err := s.CreateCorrelation(c.Req.Context(), cmd)
	if err != nil {
		if errors.Is(err, ErrSourceDataSourceDoesNotExists) || errors.Is(err, ErrTargetDataSourceDoesNotExists) {
			return response.Error(http.StatusNotFound, "Data source not found", err)
		}
		return response.Error(http.StatusInternalServerError, "Failed to add correlation", err)
	}

	return response.JSON(http.StatusOK, CreateCorrelationResponseBody{Result: correlation, Message: "Correlation created"})
}

// swagger:parameters createCorrelation
type CreateCorrelationParams struct {
	// in:body
	// required:true
	Body CreateCorrelationCommand `json:"body"`
	// in:path
	// required:true
	SourceUID string `json:"sourceUID"`
}

// swagger:response createCorrelationResponse
type CreateCorrelationResponse struct {
	// in: body
	Body CreateCorrelationResponseBody `json:"body"`
}

// swagger:route DELETE /datasources/uid/{uid}/correlations/{correlationUID} correlations deleteCorrelation
//
// Delete a correlation.
//
// Responses:
// 200: deleteCorrelationResponse
// 401: unauthorisedError
// 403: forbiddenError
// 404: notFoundError
// 500: internalServerError
func (s *CorrelationsService) deleteHandler(c *contextmodel.ReqContext) response.Response {
	if s.featuremgmt.IsEnabledGlobally(featuremgmt.FlagKubernetesCorrelations) {
		s.k8sHandler.deleteCorrelation(c)
		return nil
	}

	cmd := DeleteCorrelationCommand{
		UID:       web.Params(c.Req)[":correlationUID"],
		SourceUID: web.Params(c.Req)[":uid"],
		OrgId:     c.GetOrgID(),
	}

	err := s.DeleteCorrelation(c.Req.Context(), cmd)
	if err != nil {
		if errors.Is(err, ErrSourceDataSourceDoesNotExists) {
			return response.Error(http.StatusNotFound, "Data source not found", err)
		}

		if errors.Is(err, ErrCorrelationNotFound) {
			return response.Error(http.StatusNotFound, "Correlation not found", err)
		}

		if errors.Is(err, ErrCorrelationReadOnly) {
			return response.Error(http.StatusForbidden, "Correlation can only be edited via provisioning", err)
		}

		return response.Error(http.StatusInternalServerError, "Failed to delete correlation", err)
	}

	return response.JSON(http.StatusOK, DeleteCorrelationResponseBody{Message: "Correlation deleted"})
}

// swagger:parameters deleteCorrelation
type DeleteCorrelationParams struct {
	// in:path
	// required:true
	DatasourceUID string `json:"uid"`
	// in:path
	// required:true
	CorrelationUID string `json:"correlationUID"`
}

//swagger:response deleteCorrelationResponse
type DeleteCorrelationResponse struct {
	// in: body
	Body DeleteCorrelationResponseBody `json:"body"`
}

// swagger:route PATCH /datasources/uid/{sourceUID}/correlations/{correlationUID} correlations updateCorrelation
//
// Updates a correlation.
//
// Responses:
// 200: updateCorrelationResponse
// 400: badRequestError
// 401: unauthorisedError
// 403: forbiddenError
// 404: notFoundError
// 500: internalServerError
func (s *CorrelationsService) updateHandler(c *contextmodel.ReqContext) response.Response {
	if s.featuremgmt.IsEnabledGlobally(featuremgmt.FlagKubernetesCorrelations) {
		s.k8sHandler.updateCorrelation(c)
		return nil
	}

	cmd := UpdateCorrelationCommand{}
	if err := web.Bind(c.Req, &cmd); err != nil {
		if errors.Is(err, ErrUpdateCorrelationEmptyParams) {
			return response.Error(http.StatusBadRequest, "At least one of label, description or config is required", err)
		}

		return response.Error(http.StatusBadRequest, "bad request data", err)
	}

	cmd.UID = web.Params(c.Req)[":correlationUID"]
	cmd.SourceUID = web.Params(c.Req)[":uid"]
	cmd.OrgId = c.GetOrgID()

	correlation, err := s.UpdateCorrelation(c.Req.Context(), cmd)
	if err != nil {
		if errors.Is(err, ErrSourceDataSourceDoesNotExists) {
			return response.Error(http.StatusNotFound, "Data source not found", err)
		}

		if errors.Is(err, ErrCorrelationNotFound) {
			return response.Error(http.StatusNotFound, "Correlation not found", err)
		}

		if errors.Is(err, ErrCorrelationReadOnly) {
			return response.Error(http.StatusForbidden, "Correlation can only be edited via provisioning", err)
		}

		return response.Error(http.StatusInternalServerError, "Failed to update correlation", err)
	}

	return response.JSON(http.StatusOK, UpdateCorrelationResponseBody{Message: "Correlation updated", Result: correlation})
}

// swagger:parameters updateCorrelation
type UpdateCorrelationParams struct {
	// in:path
	// required:true
	DatasourceUID string `json:"sourceUID"`
	// in:path
	// required:true
	CorrelationUID string `json:"correlationUID"`
	// in: body
	Body UpdateCorrelationCommand `json:"body"`
}

// swagger:response updateCorrelationResponse
type UpdateCorrelationResponse struct {
	// in: body
	Body UpdateCorrelationResponseBody `json:"body"`
}

// swagger:route GET /datasources/uid/{sourceUID}/correlations/{correlationUID} correlations getCorrelation
//
// Gets a correlation.
//
// Responses:
// 200: getCorrelationResponse
// 401: unauthorisedError
// 404: notFoundError
// 500: internalServerError
func (s *CorrelationsService) getCorrelationHandler(c *contextmodel.ReqContext) response.Response {
	if s.featuremgmt.IsEnabledGlobally(featuremgmt.FlagKubernetesCorrelations) {
		// NOTE: does not get by source uid yet!
		s.k8sHandler.getCorrelation(c)
		return nil
	}

	query := GetCorrelationQuery{
		UID:       web.Params(c.Req)[":correlationUID"],
		SourceUID: web.Params(c.Req)[":uid"],
		OrgId:     c.GetOrgID(),
	}

	correlation, err := s.getCorrelation(c.Req.Context(), query)
	if err != nil {
		if errors.Is(err, ErrCorrelationNotFound) {
			return response.Error(http.StatusNotFound, "Correlation not found", err)
		}
		if errors.Is(err, ErrSourceDataSourceDoesNotExists) {
			return response.Error(http.StatusNotFound, "Source data source not found", err)
		}

		return response.Error(http.StatusInternalServerError, "Failed to get correlation", err)
	}

	return response.JSON(http.StatusOK, correlation)
}

// swagger:parameters getCorrelation
type GetCorrelationParams struct {
	// in:path
	// required:true
	DatasourceUID string `json:"sourceUID"`
	// in:path
	// required:true
	CorrelationUID string `json:"correlationUID"`
}

//swagger:response getCorrelationResponse
type GetCorrelationResponse struct {
	// in: body
	Body Correlation `json:"body"`
}

// swagger:route GET /datasources/uid/{sourceUID}/correlations correlations getCorrelationsBySourceUID
//
// Gets all correlations originating from the given data source.
//
// Responses:
// 200: getCorrelationsBySourceUIDResponse
// 401: unauthorisedError
// 404: notFoundError
// 500: internalServerError
func (s *CorrelationsService) getCorrelationsBySourceUIDHandler(c *contextmodel.ReqContext) response.Response {
	query := GetCorrelationsBySourceUIDQuery{
		SourceUID: web.Params(c.Req)[":uid"],
		OrgId:     c.GetOrgID(),
	}

	correlations, err := s.getCorrelationsBySourceUID(c.Req.Context(), query)
	if err != nil {
		if errors.Is(err, ErrCorrelationNotFound) {
			return response.Error(http.StatusNotFound, "No correlation found", err)
		}
		if errors.Is(err, ErrSourceDataSourceDoesNotExists) {
			return response.Error(http.StatusNotFound, "Source data source not found", err)
		}

		return response.Error(http.StatusInternalServerError, "Failed to get correlations", err)
	}

	return response.JSON(http.StatusOK, correlations)
}

// swagger:parameters getCorrelationsBySourceUID
type GetCorrelationsBySourceUIDParams struct {
	// in:path
	// required:true
	DatasourceUID string `json:"sourceUID"`
}

// swagger:response getCorrelationsBySourceUIDResponse
type GetCorrelationsBySourceUIDResponse struct {
	// in: body
	Body []Correlation `json:"body"`
}

// swagger:route GET /datasources/correlations correlations getCorrelations
//
// Gets all correlations.
//
// Responses:
// 200: getCorrelationsResponse
// 401: unauthorisedError
// 404: notFoundError
// 500: internalServerError
func (s *CorrelationsService) getCorrelationsHandler(c *contextmodel.ReqContext) response.Response {
	if s.featuremgmt.IsEnabledGlobally(featuremgmt.FlagKubernetesCorrelations) {
		// NOTE: does not get by source uid yet!
		s.k8sHandler.listCorrelations(c)
		return nil
	}

	limit := c.QueryInt64("limit")
	if limit <= 0 {
		limit = 100
	} else if limit > 1000 {
		limit = 1000
	}

	page := c.QueryInt64("page")
	if page <= 0 {
		page = 1
	}

	sourceUIDs := c.QueryStrings("sourceUID")

	query := GetCorrelationsQuery{
		OrgId:      c.GetOrgID(),
		Limit:      limit,
		Page:       page,
		SourceUIDs: sourceUIDs,
	}

	correlations, err := s.getCorrelations(c.Req.Context(), query)
	if err != nil {
		if errors.Is(err, ErrCorrelationNotFound) {
			return response.Error(http.StatusNotFound, "No correlation found", err)
		}

		return response.Error(http.StatusInternalServerError, "Failed to get correlations", err)
	}

	return response.JSON(http.StatusOK, correlations)
}

// swagger:parameters getCorrelations
type GetCorrelationsParams struct {
	// Limit the maximum number of correlations to return per page
	// in:query
	// required:false
	// default:100
	// maximum: 1000
	Limit int64 `json:"limit"`
	// Page index for starting fetching correlations
	// in:query
	// required:false
	// default:1
	Page int64 `json:"page"`
	// Source datasource UID filter to be applied to correlations
	// in:query
	// type: array
	// collectionFormat: multi
	// required:false
	SourceUIDs []string `json:"sourceUID"`
}

//swagger:response getCorrelationsResponse
type GetCorrelationsResponse struct {
	// in: body
	Body []Correlation `json:"body"`
}

//-----------------------------------------------------------------------------------------
// Correlation k8s wrapper functions
//-----------------------------------------------------------------------------------------

type correlationK8sHandler struct {
	namespacer           request.NamespaceMapper
	gvr                  schema.GroupVersionResource
	clientConfigProvider grafanaapiserver.DirectRestConfigProvider
}

func newCorrelationK8sHandler(cfg *setting.Cfg, clientConfigProvider grafanaapiserver.DirectRestConfigProvider) *correlationK8sHandler {
	gvr := schema.GroupVersionResource{
		Group:    correlation.CorrelationKind().Group(),
		Version:  correlation.CorrelationKind().Version(),
		Resource: correlation.CorrelationKind().Plural(),
	}
	return &correlationK8sHandler{
		gvr:                  gvr,
		namespacer:           request.GetNamespaceMapper(cfg),
		clientConfigProvider: clientConfigProvider,
	}
}

func (ck8s *correlationK8sHandler) getCorrelation(c *contextmodel.ReqContext) {
	client, ok := ck8s.getClient(c)
	if !ok {
		return // error is already sent
	}
	uid := web.Params(c.Req)[":uid"]
	out, err := client.Get(c.Req.Context(), uid, v1.GetOptions{})
	if err != nil {
		ck8s.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, UnstructuredToLegacyCorrelation(out, ck8s.namespacer))
}

func (ck8s *correlationK8sHandler) listCorrelations(c *contextmodel.ReqContext) {
	client, ok := ck8s.getClient(c)
	if !ok {
		return // error is already sent
	}
	out, err := client.List(c.Req.Context(), v1.ListOptions{})
	if err != nil {
		ck8s.writeError(c, err)
		return
	}
	correlations := make([]Correlation, 0, len(out.Items))
	for _, item := range out.Items {
		correlations = append(correlations, UnstructuredToLegacyCorrelation(&item, ck8s.namespacer))
	}

	result := GetCorrelationsResponseBody{
		Correlations: correlations,
		Page:         1,
		Limit:        100000,
		TotalCount:   int64(len(correlations)),
	}
	c.JSON(http.StatusOK, result)
}

func (ck8s *correlationK8sHandler) deleteCorrelation(c *contextmodel.ReqContext) {
	client, ok := ck8s.getClient(c)
	if !ok {
		return // error is already sent
	}
	uid := web.Params(c.Req)[":correlationUID"]
	err := client.Delete(c.Req.Context(), uid, v1.DeleteOptions{})
	if err != nil {
		ck8s.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, "")
}

func (ck8s *correlationK8sHandler) updateCorrelation(c *contextmodel.ReqContext) {
	client, ok := ck8s.getClient(c)
	if !ok {
		return // error is already sent
	}
	uid := web.Params(c.Req)[":correlationUID"]
	cmd := UpdateCorrelationCommand{}
	if err := web.Bind(c.Req, &cmd); err != nil {
		c.JsonApiErr(http.StatusBadRequest, "bad request data", err)
		return
	}
	cmd.OrgId = c.GetOrgID()
	existing, err := client.Get(c.Req.Context(), uid, v1.GetOptions{})
	if err != nil {
		ck8s.writeError(c, err)
		return
	}

	existingSpec := existing.Object["spec"].(map[string]interface{})
	if cmd.SourceUID == "" {
		cmd.SourceUID = existingSpec["source_uid"].(string)
	}

	obj := LegacyUpdateCommandToUnstructured(cmd, ck8s.namespacer)

	if targetUid, ok := existingSpec["target_uid"].(string); ok {
		obj.Object["spec"].(map[string]interface{})["target_uid"] = targetUid
	}

	obj.SetResourceVersion(existing.GetResourceVersion())
	obj.SetNamespace(ck8s.namespacer(cmd.OrgId))
	obj.SetName(uid)
	out, err := client.Update(c.Req.Context(), &obj, v1.UpdateOptions{})
	if err != nil {
		ck8s.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, UnstructuredToLegacyCorrelation(out, ck8s.namespacer))
}

func (ck8s *correlationK8sHandler) createCorrelation(c *contextmodel.ReqContext) {
	client, ok := ck8s.getClient(c)
	if !ok {
		return // error is already sent
	}
	cmd := CreateCorrelationCommand{}
	if err := web.Bind(c.Req, &cmd); err != nil {
		c.JsonApiErr(http.StatusBadRequest, "bad request data", err)
		return
	}
	cmd.SourceUID = web.Params(c.Req)[":uid"]
	cmd.OrgId = c.GetOrgID()
	obj := LegacyCreateCommandToUnstructured(cmd, ck8s.namespacer)
	out, err := client.Create(c.Req.Context(), &obj, v1.CreateOptions{})
	if err != nil {
		ck8s.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, UnstructuredToLegacyCorrelation(out, ck8s.namespacer))
}

func (ck8s *correlationK8sHandler) getClient(c *contextmodel.ReqContext) (dynamic.ResourceInterface, bool) {
	dyn, err := dynamic.NewForConfig(ck8s.clientConfigProvider.GetDirectRestConfig(c))
	if err != nil {
		c.JsonApiErr(500, "client", err)
		return nil, false
	}
	return dyn.Resource(ck8s.gvr).Namespace(ck8s.namespacer(c.OrgID)), true
}

func (ck8s *correlationK8sHandler) writeError(c *contextmodel.ReqContext, err error) {
	//nolint:errorlint
	statusError, ok := err.(*k8serrors.StatusError)
	if ok {
		c.JsonApiErr(int(statusError.Status().Code), statusError.Status().Message, err)
		return
	}
	errhttp.Write(c.Req.Context(), err, c.Resp)
}

func UnstructuredToLegacyCorrelation(item *unstructured.Unstructured, namespacer request.NamespaceMapper) Correlation {
	meta, err := utils.MetaAccessor(item)
	if err != nil {
		return Correlation{}
	}
	info, _ := authlib.ParseNamespace(meta.GetNamespace())
	if info.OrgID < 0 {
		info.OrgID = 1 // This resolves all test cases that assume org 1
	}

	spec := item.Object["spec"].(map[string]interface{})

	c := Correlation{
		UID:         meta.GetName(),
		Label:       spec["label"].(string),
		Description: spec["description"].(string),
		Type:        CorrelationType(spec["type"].(string)),
		SourceUID:   spec["source_uid"].(string),
		OrgID:       info.OrgID,
	}

	if provisioned, ok := spec["provisioned"].(bool); ok {
		c.Provisioned = provisioned
	}

	if targetUid, ok := spec["target_uid"].(string); ok {
		c.TargetUID = &targetUid
	}

	if config, ok := spec["config"].(map[string]interface{}); ok {
		configBytes, _ := json.Marshal(config)
		_ = json.Unmarshal(configBytes, &c.Config)
	}

	return c
}
func ConvertToK8sResource(c Correlation, namespacer request.NamespaceMapper) correlation.Correlation {
	h := fnv.New64a()
	h.Write([]byte(fmt.Sprintf("%d-%s", c.OrgID, c.UID)))
	hash := h.Sum64()

	cor := correlation.Correlation{
		ObjectMeta: metav1.ObjectMeta{
			Name:              c.UID,
			UID:               types.UID(c.UID),
			ResourceVersion:   fmt.Sprintf("%d", hash), // typically will be the creation timestamp, but no timestamps in correlations. see playlists for example.
			CreationTimestamp: metav1.NewTime(time.UnixMilli(time.Now().UnixMilli())),
			Namespace:         namespacer(c.OrgID),
		},
		Spec: correlation.CorrelationSpec{
			SourceUid:   c.SourceUID,
			Label:       c.Label,
			Description: c.Description,
			Type:        string(c.Type),
		},
	}

	if c.TargetUID != nil {
		cor.Spec.TargetUid = *c.TargetUID
	}

	if c.Provisioned {
		cor.Spec.Provisioned = 1
	}

	if cfg, err := json.Marshal(c.Config); err == nil {
		cor.Spec.Config = string(cfg)
	}

	cor.UID = gapiutil.CalculateClusterWideUID(&cor)
	return cor
}

func ConvertToLegacyCreateCommand(c *correlation.Correlation, orgId int64) CreateCorrelationCommand {
	var config CorrelationConfig
	if err := json.Unmarshal([]byte(c.Spec.Config), &config); err != nil {
		config = CorrelationConfig{}
	}

	return CreateCorrelationCommand{
		OrgId:       orgId,
		SourceUID:   c.Spec.SourceUid,
		TargetUID:   &c.Spec.TargetUid,
		Label:       c.Spec.Label,
		Description: c.Spec.Description,
		Config:      config,
		Provisioned: c.Spec.Provisioned != 0,
		Type:        CorrelationType(c.Spec.Type),
	}
}

func ConvertToLegacyUpdateCommand(c *correlation.Correlation, orgId int64) UpdateCorrelationCommand {
	var config CorrelationConfig
	if err := json.Unmarshal([]byte(c.Spec.Config), &config); err != nil {
		config = CorrelationConfig{}
	}
	correlationType := CorrelationType(c.Spec.Type)

	return UpdateCorrelationCommand{
		OrgId:       orgId,
		SourceUID:   c.Spec.SourceUid,
		Label:       &c.Spec.Label,
		Description: &c.Spec.Description,
		Config: &CorrelationConfigUpdateDTO{
			Field:           &config.Field,
			Target:          &config.Target,
			Transformations: config.Transformations,
		},
		Type: &correlationType,
	}
}

func LegacyUpdateCommandToUnstructured(cmd UpdateCorrelationCommand, namespacer request.NamespaceMapper) unstructured.Unstructured {
	spec := map[string]interface{}{}

	if cmd.Label != nil {
		spec["label"] = *cmd.Label
	}
	if cmd.Description != nil {
		spec["description"] = *cmd.Description
	}
	if cmd.Type != nil {
		spec["type"] = *cmd.Type
	}
	if cmd.Config != nil {
		configBytes, _ := json.Marshal(cmd.Config)
		spec["config"] = string(configBytes)
	}

	if cmd.SourceUID != "" {
		spec["source_uid"] = cmd.SourceUID
	}

	finalObj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"spec": spec,
		},
	}
	finalObj.SetName(cmd.UID)
	finalObj.SetNamespace(namespacer(cmd.OrgId))
	finalObj.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   correlation.CorrelationKind().Group(),
		Version: correlation.CorrelationKind().Version(),
		Kind:    correlation.CorrelationKind().Kind(),
	})

	return *finalObj
}

func LegacyCreateCommandToUnstructured(cmd CreateCorrelationCommand, namespacer request.NamespaceMapper) unstructured.Unstructured {
	spec := map[string]interface{}{
		"label":       cmd.Label,
		"description": cmd.Description,
		"type":        cmd.Type,
		"source_uid":  cmd.SourceUID,
	}

	if cmd.TargetUID != nil {
		spec["target_uid"] = *cmd.TargetUID
	}

	if cmd.Provisioned {
		spec["provisioned"] = 1
	}

	configBytes, _ := json.Marshal(cmd.Config)
	spec["config"] = string(configBytes)

	finalObj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"spec": spec,
		},
	}
	uid := string(gapiutil.CalculateClusterWideUID(&correlation.Correlation{}))
	finalObj.SetName(uid)
	finalObj.SetNamespace(namespacer(cmd.OrgId))
	finalObj.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   correlation.CorrelationKind().Group(),
		Version: correlation.CorrelationKind().Version(),
		Kind:    correlation.CorrelationKind().Kind(),
	})

	return *finalObj
}
