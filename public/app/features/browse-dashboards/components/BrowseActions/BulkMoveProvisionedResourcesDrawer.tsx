import { useEffect, useState } from 'react';
import { FormProvider, useForm } from 'react-hook-form';

import { t, Trans } from '@grafana/i18n';
import { Box, Button, Drawer, Field, Stack } from '@grafana/ui';
import { useGetFolderQuery } from 'app/api/clients/folder/v1beta1';
import {
  RepositoryView,
  useCreateRepositoryFilesWithPathMutation,
  useDeleteRepositoryFilesWithPathMutation,
} from 'app/api/clients/provisioning/v0alpha1';
import { FolderPicker } from 'app/core/components/Select/FolderPicker';
import { AnnoKeySourcePath } from 'app/features/apiserver/types';
import { ResourceEditFormSharedFields } from 'app/features/dashboard-scene/components/Provisioned/ResourceEditFormSharedFields';
import { getDefaultWorkflow, getWorkflowOptions } from 'app/features/dashboard-scene/saving/provisioned/defaults';
import { generateTimestamp } from 'app/features/dashboard-scene/saving/provisioned/utils/timestamp';
import { useGetResourceRepositoryView } from 'app/features/provisioning/hooks/useGetResourceRepositoryView';
import { WorkflowOption } from 'app/features/provisioning/types';

import { DashboardTreeSelection } from '../../types';

import { DescendantCount } from './DescendantCount';
import { executeBulkMove } from './utils';

export type BulkMoveFormData = {
  comment: string;
  ref: string;
  workflow?: WorkflowOption;
  path?: string;
};

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
  const { data: targetFolder } = useGetFolderQuery({ name: targetFolderUID! }, { skip: !targetFolderUID });

  const methods = useForm<BulkMoveFormData>({ defaultValues: initialValues });
  const { handleSubmit, watch } = methods;
  const [workflow] = watch(['ref', 'workflow']);

  const onFolderChange = (folderUid?: string) => {
    setTargetFolderUID(folderUid || '');
  };

  const handleSubmitForm = async (formData: BulkMoveFormData) => {
    if (!targetFolder || !repository) {
      return;
    }

    setIsLoading(true);

    try {
      const folderAnnotations = targetFolder?.metadata.annotations || {};

      const result = await executeBulkMove({
        selectedItems,
        targetFolderPath: folderAnnotations[AnnoKeySourcePath],
        repository,
        mutations: { createFile, deleteFile },
        options: { ...formData },
      });

      // Handle results
      if (result.failed.length === 0) {
        console.log(`Successfully moved all ${result.summary.successCount} dashboards`);
        onClose();
      } else {
        console.log(`Partial success: ${result.summary.successCount} successful, ${result.summary.failedCount} failed`);
        console.error('Failed moves:', result.failed);
        // TODO: Show partial success dialog
      }
    } catch (error) {
      console.error('Bulk move failed:', error);
      // TODO: Show error notification
    } finally {
      setIsLoading(false);
    }
  };

  const canSubmit = targetFolderUID && !isLoading;

  return (
    <Drawer onClose={onClose} title={t('browse-dashboards.bulk-move-resources-form.title', 'Bulk Move Resources')}>
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
            <Field label={t('browse-dashboards.bulk-move-resources-form.target-folder', 'Target Folder')}>
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
