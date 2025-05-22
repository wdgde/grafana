import { HttpHandler } from 'msw';
import { SetupWorker } from 'msw/browser';

import worker from './worker';

export type Scenario = HttpHandler[];

type ScenarioRegistryOptions = {
  worker?: SetupWorker; // optionally use a worker that is different from the default one
};

export class ScenarioRegistry {
  scenarios: Map<string, Scenario>;
  worker: SetupWorker;

  constructor(options?: ScenarioRegistryOptions) {
    this.scenarios = new Map();
    this.worker = options?.worker ?? worker;
  }

  registerScenario(name: string, scenario: Scenario) {
    this.scenarios.set(name, scenario);
  }

  // return an array of handlers from the registry of scenarios so we can pass this into the MSW worker
  handlers(): HttpHandler[] {
    return Array.from(this.scenarios.values()).flat();
  }

  availableScenarios(): string[] {
    return Array.from(this.scenarios.keys());
  }
}
