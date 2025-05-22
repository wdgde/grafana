import { ScenarioRegistry } from '@grafana/test-utils';

export function useScenarioRegistry(): ScenarioRegistry {
  const registry = new ScenarioRegistry();
  return registry;
}
