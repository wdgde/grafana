import store from 'app/core/store';

import { sendAppNotification } from './core/copy/appNotification';
import { AppNotificationSeverity } from './types';

export const STORAGE_MOCK_API_KEY = 'grafana.dev.mockApi';
export const SCENARIO_SEARCH_PARAM = '__scenario';

export const getMockEnabledStateFromLocalstorage = () => store.getBool(STORAGE_MOCK_API_KEY, false);

export const toggleMockApiAndReload = () => {
  const currentState = getMockEnabledStateFromLocalstorage();
  store.set(STORAGE_MOCK_API_KEY, String(!currentState));

  const action = currentState ? 'Disabling' : 'Enabling';

  sendAppNotification(`${action} Mock API`, 'Reloading...', AppNotificationSeverity.Info);

  setTimeout(() => {
    window.location.reload();
  }, 200);
};

export const potentiallySetupMockApi = async () => {
  const scenarioNames = new URLSearchParams(window.location.search).getAll(SCENARIO_SEARCH_PARAM);

  if (wantsMockingEnabled()) {
    const { default: worker } = await import('@grafana/test-utils/worker');

    // TODO: Generalise and move Alerting handlers into @grafana/test-utils or @grafana/alerting package
    const { default: alertingHandlers } = await import('./features/alerting/unified/mocks/server/all-handlers');
    worker.use(...alertingHandlers);

    // alerting package scenarios
    const { scenarios: ALERTING_PKG_SCENARIOS } = await import('@grafana/alerting/internal');
    scenarioNames.forEach((name) => {
      if (name in ALERTING_PKG_SCENARIOS) {
        // @ts-expect-error
        worker.use(...ALERTING_PKG_SCENARIOS[name]);
      }
    });

    worker.start({ onUnhandledRequest: 'bypass' });
  }
};

function wantsMockingEnabled() {
  const isDevEnv = process.env.NODE_ENV === 'development';
  const enabledInLocalStorage = getMockEnabledStateFromLocalstorage();
  const scenarioNames = new URLSearchParams(window.location.search).getAll(SCENARIO_SEARCH_PARAM);

  return isDevEnv && (enabledInLocalStorage || scenarioNames.length > 0);
}

export const notifyIfMockApiEnabled = () => {
  if (wantsMockingEnabled()) {
    sendAppNotification(
      'Mock API currently enabled',
      'Some network requests will be intercepted',
      AppNotificationSeverity.Info
    );
  }
};
