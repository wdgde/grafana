import { FC } from 'react';

import { FeatureToggles } from '@grafana/data';
import { DashboardPageRouteSearchParams } from 'app/features/dashboard/containers/types';

export interface DashboardBannersProps {
  queryParams: DashboardPageRouteSearchParams;
  path?: string;
  route?: string;
  slug?: string;
}

export interface BannerConfig {
  key: string;
  Component: FC<DashboardBannerProps>;
  featureFlag?: keyof FeatureToggles;
  routes?: string[];
  showInKiosk?: true;
  check?: () => boolean;
}

export interface DashboardBannerProps extends DashboardBannersProps {
  onDismiss: () => void;
}

export const commonAlertProps = {
  severity: 'info' as const,
  style: { flex: 0 } as const,
  bottomSpacing: 0,
};
