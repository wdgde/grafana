import { t } from '@grafana/i18n';
import { Alert } from '@grafana/ui';

import { alertRuleApi } from '../../../api/alertRuleApi';
import { GRAFANA_RULER_CONFIG } from '../../../api/featureDiscoveryApi';
import { useSettingsPageNav } from '../../../settings/navigation';
import { stringifyErrorLike } from '../../../utils/misc';
import { withPageErrorBoundary } from '../../../withPageErrorBoundary';
import { AlertingPageWrapper } from '../../AlertingPageWrapper';

import { DeletedRules } from './DeletedRules';

function DeletedrulesPage() {
  const { navId, pageNav } = useSettingsPageNav();

  const {
    currentData = [],
    isLoading,
    error,
  } = alertRuleApi.endpoints.getDeletedRules.useQuery({
    rulerConfig: GRAFANA_RULER_CONFIG,
    filter: {}, // todo: add filters, and limit?????
  });

  return (
    <AlertingPageWrapper navId={navId} isLoading={isLoading} pageNav={pageNav}>
      <>
        {error && (
          <Alert title={t('alerting.deleted-rules.errorloading', 'Failed to load alert deleted rules')}>
            {stringifyErrorLike(error)}
          </Alert>
        )}
        <DeletedRules deletedRules={currentData} />
      </>
    </AlertingPageWrapper>
  );
}

export default withPageErrorBoundary(DeletedrulesPage);
