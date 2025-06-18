import { Trans } from '@grafana/i18n';
import { Alert, Icon } from '@grafana/ui';

import { commonAlertProps, DashboardBannerProps } from './utils';

export function DynamicDashboardsWarningBanner({ onDismiss }: DashboardBannerProps) {
  return (
    <Alert {...commonAlertProps} severity="warning" title="" buttonContent={<Icon name="times" />} onRemove={onDismiss}>
      <Trans i18nKey="">Lorem ipsum dolores sit amet</Trans>
    </Alert>
  );
}
