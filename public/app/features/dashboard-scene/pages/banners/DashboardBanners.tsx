import { css } from '@emotion/css';
import { useMemo } from 'react';
import { useLocalStorage } from 'react-use';

import { config } from '@grafana/runtime';
import { useStyles2 } from '@grafana/ui';
import { DashboardRoutes } from 'app/types';

import { DashboardPreviewBanner } from './DashboardPreviewBanner';
import { DynamicDashboardsWarningBanner } from './DynamicDashboardsWarningBanner';
import { BannerConfig, DashboardBannersProps } from './utils';

export function DashboardBanners(props: DashboardBannersProps) {
  const styles = useStyles2(getStyles);

  const [dismissed, setDismissed] = useLocalStorage<string[]>('grafana.dashboard.banners.dismissed', []);

  const banners = useMemo(() => {
    if (!props.route || !props.path) {
      return [];
    }

    const bannersConfig: BannerConfig[] = [
      {
        key: 'dashboard-preview',
        Component: DashboardPreviewBanner,
        featureFlag: 'provisioning',
        routes: [DashboardRoutes.Provisioning],
        check: () => !!props.slug,
      },
      {
        key: 'dynamic-dashboards-warning',
        Component: DynamicDashboardsWarningBanner,
        featureFlag: 'dashboardNewLayouts',
        routes: [DashboardRoutes.New, DashboardRoutes.Normal],
      },
    ];

    const isKiosk = 'kiosk' in props.queryParams;

    return bannersConfig.filter(
      (bannerConfig) =>
        (!bannerConfig.featureFlag || config.featureToggles[bannerConfig.featureFlag]) &&
        (!isKiosk || bannerConfig.showInKiosk) &&
        (!bannerConfig.routes || bannerConfig.routes.includes(props.route!)) &&
        (!dismissed || !dismissed.includes(bannerConfig.key)) &&
        (bannerConfig.check?.() ?? true)
    );
  }, [props.route, props.path, props.queryParams, props.slug, dismissed]);

  if (banners.length === 0) {
    return null;
  }

  return (
    <div className={styles.container}>
      {banners.map(({ key, Component }) => (
        <Component
          key={key}
          {...props}
          onDismiss={() => {
            if (!dismissed) {
              setDismissed([key]);
              return;
            }

            if (!dismissed.includes(key)) {
              setDismissed([...dismissed, key]);
            }
          }}
        />
      ))}
    </div>
  );
}

const getStyles = () => ({
  container: css({
    display: 'flex',
    flexDirection: 'column',
  }),
});
