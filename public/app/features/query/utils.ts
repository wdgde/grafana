import { createMonitoringLogger } from '@grafana/runtime';

export const queryLogger: ReturnType<typeof createMonitoringLogger> = createMonitoringLogger('features.query');
