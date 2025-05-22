import { ScenarioRegistry } from './registry';

export type UseScenarioRegistryOptions = {};

export type UseScenarioRegistry = (options: UseScenarioRegistryOptions) => ScenarioRegistry;

let singleton: UseScenarioRegistry | undefined;

export function setScenarioRegistryHook(hook: UseScenarioRegistry): void {
  // We allow overriding the hook in tests
  if (singleton && process.env.NODE_ENV !== 'test') {
    throw new Error('setScenarioRegistryHook() function should only be called once, when Grafana is starting.');
  }
  singleton = hook;
}

export function useScenarioRegistry(options: UseScenarioRegistryOptions): ScenarioRegistry {
  if (!singleton) {
    throw new Error('setScenarioRegistryHook(options) can only be used after the Grafana instance has started.');
  }
  return singleton(options);
}
