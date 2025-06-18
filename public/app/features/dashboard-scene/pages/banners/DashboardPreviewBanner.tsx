import { textUtil } from '@grafana/data';
import { Trans, t } from '@grafana/i18n';
import { Alert, Icon, Stack } from '@grafana/ui';
import { useGetRepositoryFilesWithPathQuery } from 'app/api/clients/provisioning/v0alpha1';
import { usePullRequestParam } from 'app/features/provisioning/hooks/usePullRequestParam';

import { commonAlertProps, DashboardBannerProps, DashboardBannersProps } from './utils';

function DashboardPreviewBannerContent({ queryParams, slug, path }: Required<Omit<DashboardBannersProps, 'route'>>) {
  const prParam = usePullRequestParam();
  const file = useGetRepositoryFilesWithPathQuery({ name: slug, path, ref: queryParams.ref });

  if (file.data?.errors) {
    return (
      <Alert
        {...commonAlertProps}
        title={t('dashboard-scene.dashboard-preview-banner.title-error-loading-dashboard', 'Error loading dashboard')}
        severity="error"
      >
        {file.data.errors.map((error, index) => (
          <div key={index}>{error}</div>
        ))}
      </Alert>
    );
  }

  // This page was loaded with a `pull_request_url` in the URL
  if (prParam?.length) {
    return (
      <Alert
        {...commonAlertProps}
        title={t(
          'dashboard-scene.dashboard-preview-banner.title-dashboard-loaded-request-git-hub',
          'This dashboard is loaded from a pull request in GitHub.'
        )}
        buttonContent={
          <Stack alignItems="center">
            <Trans i18nKey="dashboard-scene.dashboard-preview-banner.view-pull-request-in-git-hub">
              View pull request in GitHub
            </Trans>
            <Icon name="external-link-alt" />
          </Stack>
        }
        onRemove={() => window.open(textUtil.sanitizeUrl(prParam), '_blank')}
      >
        <Trans i18nKey="dashboard-scene.dashboard-preview-banner.value-not-saved">
          The value is not yet saved in the Grafana database
        </Trans>
      </Alert>
    );
  }

  // Check if this is a GitHub link
  const githubURL = file.data?.urls?.newPullRequestURL ?? file.data?.urls?.compareURL;
  if (githubURL) {
    return (
      <Alert
        {...commonAlertProps}
        title={t(
          'dashboard-scene.dashboard-preview-banner.title-dashboard-loaded-branch-git-hub',
          'This dashboard is loaded from a branch in GitHub.'
        )}
        buttonContent={
          <Stack alignItems="center">
            <Trans i18nKey="dashboard-scene.dashboard-preview-banner.open-pull-request-in-git-hub">
              Open pull request in GitHub
            </Trans>
            <Icon name="external-link-alt" />
          </Stack>
        }
        onRemove={() => window.open(textUtil.sanitizeUrl(githubURL), '_blank')}
      >
        <Trans i18nKey="dashboard-scene.dashboard-preview-banner.not-saved">
          The value is not yet saved in the Grafana database
        </Trans>
      </Alert>
    );
  }

  return (
    <Alert
      {...commonAlertProps}
      title={t(
        'dashboard-scene.dashboard-preview-banner.title-dashboard-loaded-external-repository',
        'This dashboard is loaded from an external repository'
      )}
    >
      <Trans i18nKey="dashboard-scene.dashboard-preview-banner.not-yet-saved">
        The value is not saved in the Grafana database
      </Trans>
    </Alert>
  );
}

export function DashboardPreviewBanner({ queryParams, slug, path }: DashboardBannerProps) {
  return <DashboardPreviewBannerContent queryParams={queryParams} slug={slug!} path={path!} />;
}
