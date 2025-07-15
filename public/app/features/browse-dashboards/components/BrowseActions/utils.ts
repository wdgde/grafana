import { t } from '@grafana/i18n';
import { config, getBackendSrv } from '@grafana/runtime';
import { RepositoryView } from 'app/api/clients/provisioning/v0alpha1';
import { AnnoKeySourcePath } from 'app/features/apiserver/types';
import { getDashboardAPI } from 'app/features/dashboard/api/dashboard_api';

import { DashboardTreeSelection } from '../../types';

import { BulkMoveFormData } from './BulkMoveProvisionedResourcesDrawer';

export function buildBreakdownString(
  folderCount: number,
  dashboardCount: number,
  libraryPanelCount: number,
  alertRuleCount: number
) {
  const total = folderCount + dashboardCount + libraryPanelCount + alertRuleCount;
  const parts = [];
  if (folderCount) {
    parts.push(t('browse-dashboards.counts.folder', '{{count}} folder', { count: folderCount }));
  }
  if (dashboardCount) {
    parts.push(t('browse-dashboards.counts.dashboard', '{{count}} dashboard', { count: dashboardCount }));
  }
  if (libraryPanelCount) {
    parts.push(t('browse-dashboards.counts.libraryPanel', '{{count}} library panel', { count: libraryPanelCount }));
  }
  if (alertRuleCount) {
    parts.push(t('browse-dashboards.counts.alertRule', '{{count}} alert rule', { count: alertRuleCount }));
  }
  let breakdownString = t('browse-dashboards.counts.total', '{{count}} item', { count: total });
  if (parts.length > 0) {
    breakdownString += `: ${parts.join(', ')}`;
  }
  return breakdownString;
}

interface BulkMoveRequest {
  selectedItems: Omit<DashboardTreeSelection, 'panel' | '$all'>;
  targetFolderPath: string;
  repository: RepositoryView;
  mutations: {
    createFile: any;
    deleteFile: any;
  };
  options: BulkMoveFormData;
}

interface BulkMoveResult {
  successful: Array<{
    uid: string;
    status: 'success';
    oldPath: string;
    newPath: string;
    title?: string;
  }>;
  failed: Array<{
    uid: string;
    status: 'failed';
    error: string;
    title?: string;
  }>;
  summary: {
    total: number;
    successCount: number;
    failedCount: number;
  };
}

/**
 * Execute bulk move operation for selected dashboards
 * TODO: This will be replaced with backend job queue system
 */
export const executeBulkMove = async ({
  selectedItems,
  targetFolderPath,
  repository,
  mutations,
  options,
}: BulkMoveRequest): Promise<BulkMoveResult> => {
  console.log('Starting bulk move operation...');

  // 1. Get dashboard data for all selected dashboards
  const dashboardsToMove = Object.keys(selectedItems.dashboard).filter((uid) => selectedItems.dashboard[uid]);

  console.log(`Fetching data for ${dashboardsToMove.length} dashboards...`);

  const dashboardDataResults = await Promise.allSettled(
    dashboardsToMove.map(async (uid) => {
      try {
        const dto = await getDashboardAPI().getDashboardDTO(uid);
        const sourcePath = dto.meta.k8s?.annotations?.[AnnoKeySourcePath] || dto.meta.provisionedExternalId;

        if (!sourcePath) {
          throw new Error(`No source path found for dashboard ${uid}`);
        }

        const baseUrl = `/apis/provisioning.grafana.app/v0alpha1/namespaces/${config.namespace}`;
        const url = `${baseUrl}/repositories/${repository.name}/files/${sourcePath}`;
        const fileResponse = await getBackendSrv().get(url);

        return {
          uid,
          data: fileResponse,
          title: dto.dashboard.title,
        };
      } catch (error) {
        console.error(`Failed to get dashboard data for ${uid}:`, error);
        throw error;
      }
    })
  );

  // 2. Filter successful data fetches
  const fulfilledResources = dashboardDataResults
    .filter((result): result is PromiseFulfilledResult<any> => result.status === 'fulfilled')
    .map((result) => result.value);

  const dataFetchFailures = dashboardDataResults
    .map((result, index) => {
      if (result.status === 'rejected') {
        return {
          uid: dashboardsToMove[index],
          status: 'failed' as const,
          error: result.reason?.message || 'Failed to fetch dashboard data',
          title: undefined,
        };
      }
      return null;
    })
    .filter(Boolean);

  console.log(`Successfully fetched data for ${fulfilledResources.length}/${dashboardsToMove.length} dashboards`);

  // 3. Execute moves for dashboards with valid data
  const moveResults = await Promise.allSettled(
    fulfilledResources.map(async ({ uid, data, title }) => {
      try {
        console.log(`Moving dashboard ${uid}`);

        const fileName = data.resource?.dryRun?.metadata?.annotations?.[AnnoKeySourcePath];
        const newPath = `${targetFolderPath}/${fileName}`;
        const body = data.resource.file;

        // Create in target location
        await mutations
          .createFile({
            name: repository.name,
            path: newPath,
            ref: options.workflow === 'write' ? undefined : options.ref,
            message: options.comment || `Move dashboard: ${title || uid}`,
            body: body,
          })
          .unwrap();

        // Delete from current location
        await mutations
          .deleteFile({
            name: repository.name,
            path: `${options.path}/${fileName}`,
            ref: options.workflow === 'write' ? undefined : options.ref,
            message: options.comment || `Move dashboard: ${title || uid}`,
          })
          .unwrap();

        console.log(`Successfully moved dashboard ${uid}`);

        return {
          uid,
          status: 'success' as const,
          oldPath: `${options.path}/${fileName}`,
          newPath,
          title,
        };
      } catch (error) {
        console.error(`Failed to move dashboard ${uid}:`, error);
        return {
          uid,
          status: 'failed' as const,
          error: error instanceof Error ? error.message : 'Unknown error',
          title,
        };
      }
    })
  );

  // 4. Process results
  const successful = [];
  const failed = [...dataFetchFailures];

  moveResults.forEach((result) => {
    if (result.status === 'fulfilled') {
      if (result.value.status === 'success') {
        successful.push(result.value);
      } else {
        failed.push(result.value);
      }
    } else {
      failed.push({
        uid: 'unknown',
        status: 'failed' as const,
        error: result.reason?.message || 'Unknown error',
        title: undefined,
      });
    }
  });

  const summary = {
    total: dashboardsToMove.length,
    successCount: successful.length,
    failedCount: failed.length,
  };

  console.log(`Bulk move completed: ${summary.successCount} successful, ${summary.failedCount} failed`);

  return {
    successful,
    failed,
    summary,
  };
};
