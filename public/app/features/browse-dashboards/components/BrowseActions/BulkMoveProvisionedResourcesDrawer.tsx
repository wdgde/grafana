import { useEffect, useState } from 'react';
import { FormProvider, useForm } from 'react-hook-form';

import { t, Trans } from '@grafana/i18n';
import { Box, Button, Drawer, Field, Stack } from '@grafana/ui';
import {
  RepositoryView,
  useCreateRepositoryFilesWithPathMutation,
  useDeleteRepositoryFilesWithPathMutation,
  useGetRepositoryFilesWithPathQuery,
} from 'app/api/clients/provisioning/v0alpha1';
import { FolderPicker } from 'app/core/components/Select/FolderPicker';
import { AnnoKeySourcePath } from 'app/features/apiserver/types';
import { ResourceEditFormSharedFields } from 'app/features/dashboard-scene/components/Provisioned/ResourceEditFormSharedFields';
import { getDefaultWorkflow, getWorkflowOptions } from 'app/features/dashboard-scene/saving/provisioned/defaults';
import { generateTimestamp } from 'app/features/dashboard-scene/saving/provisioned/utils/timestamp';
import { useGetResourceRepositoryView } from 'app/features/provisioning/hooks/useGetResourceRepositoryView';
import { WorkflowOption } from 'app/features/provisioning/types';

import { useGetFolderQuery } from '../../api/browseDashboardsAPI';
import { DashboardTreeSelection } from '../../types';

import { DescendantCount } from './DescendantCount';

interface BulkMoveFormData {
  comment: string;
  ref: string;
  workflow?: WorkflowOption;
}

interface FormProps extends BulkMoveProvisionResourceProps {
  initialValues: BulkMoveFormData;
  repository: RepositoryView;
  workflowOptions: Array<{ label: string; value: string }>;
  isGitHub: boolean;
  folderPath?: string;
}

interface BulkMoveProvisionResourceProps {
  folderUid?: string;
  selectedItems: Omit<DashboardTreeSelection, 'panel' | '$all'>;
  onClose: () => void;
}

export function FormContent({
  selectedItems,
  initialValues,
  folderUid,
  folderPath,
  workflowOptions,
  isGitHub,
  repository,
  onClose,
}: FormProps) {
  const [createFile] = useCreateRepositoryFilesWithPathMutation();
  const [deleteFile] = useDeleteRepositoryFilesWithPathMutation();
  const [isLoading, setIsLoading] = useState(false);

  const [targetFolderUID, setTargetFolderUID] = useState<string>(folderUid || '');
  const { data: targetFolder } = useGetFolderQuery(targetFolderUID || '', { skip: !targetFolderUID });

  const methods = useForm<BulkMoveFormData>({ defaultValues: initialValues });
  const { handleSubmit, watch } = methods;
  const [ref, workflow] = watch(['ref', 'workflow']);

  const onFolderChange = (folderUid?: string) => {
    setTargetFolderUID(folderUid || '');
  };

  const handleSubmitForm = async (data: BulkMoveFormData) => {
    if (!targetFolder || !repository) {
      return;
    }

    setIsLoading(true);

    try {
      // TODO: Implement actual bulk move logic here
      // This is where you'll implement the actual API calls to:
      // 1. Get each resource's current data and source path
      // 2. Delete from source location
      // 3. Create in target location

      console.log('Moving resources to:', targetFolder.title);
      console.log(
        'Selected folders:',
        Object.keys(selectedItems.folder).filter((uid) => selectedItems.folder[uid])
      );
      console.log(
        'Selected dashboards:',
        Object.keys(selectedItems.dashboard).filter((uid) => selectedItems.dashboard[uid])
      );
      console.log('Form data:', data);

      // Simulate work for now
      // await new Promise(resolve => setTimeout(resolve, 2000));

      console.log('Move completed successfully');
      onClose();
    } catch (error) {
      console.error('Move failed:', error);
      // TODO: Add proper error handling and user feedback
    } finally {
      setIsLoading(false);
    }
  };

  const canSubmit = targetFolderUID && !isLoading;

  return (
    <Drawer onClose={onClose} title="Bulk Move Resources">
      <FormProvider {...methods}>
        <form onSubmit={handleSubmit(handleSubmitForm)}>
          <Stack direction="column" gap={2}>
            <Box paddingBottom={2}>
              <Trans i18nKey="browse-dashboards.bulk-move-resources-form.move-warning">
                This will move selected resources and their descendants. In total, this will affect:
              </Trans>
              <DescendantCount selectedItems={{ ...selectedItems, panel: {}, $all: false }} />
            </Box>

            {/* Target folder selection */}
            <Field label={t('dashboard-settings.general.folder-label', 'Target Folder')}>
              <FolderPicker value={targetFolderUID} onChange={onFolderChange} />
            </Field>

            <ResourceEditFormSharedFields
              resourceType="dashboard"
              readOnly={isLoading}
              workflow={workflow}
              workflowOptions={workflowOptions}
              isGitHub={isGitHub}
            />

            <Stack gap={2}>
              <Button variant="primary" type="submit" disabled={!canSubmit}>
                {isLoading
                  ? t('browse-dashboards.bulk-move-resources-form.moving', 'Moving...')
                  : t('browse-dashboards.bulk-move-resources-form.move-action', 'Move Resources')}
              </Button>
              <Button variant="secondary" onClick={onClose} fill="outline" disabled={isLoading}>
                <Trans i18nKey="browse-dashboards.bulk-move-resources-form.cancel-action">Cancel</Trans>
              </Button>
            </Stack>
          </Stack>
        </form>
      </FormProvider>
    </Drawer>
  );
}

export function BulkMoveProvisionedResourceDrawer({
  folderUid,
  selectedItems,
  onClose,
}: BulkMoveProvisionResourceProps) {
  const { repository, folder } = useGetResourceRepositoryView({ folderName: folderUid });

  const workflowOptions = getWorkflowOptions(repository);
  const isGitHub = repository?.type === 'github';
  const folderPath = folder?.metadata?.annotations?.[AnnoKeySourcePath] || '';
  const timestamp = generateTimestamp();

  const initialValues = {
    comment: '',
    ref: `bulk-move/${timestamp}`,
    workflow: getDefaultWorkflow(repository),
  };

  if (!repository) {
    return null;
  }

  return (
    <FormContent
      repository={repository}
      selectedItems={selectedItems}
      initialValues={initialValues}
      folderUid={folderUid}
      folderPath={folderPath}
      workflowOptions={workflowOptions}
      isGitHub={isGitHub}
      onClose={onClose}
    />
  );
}
