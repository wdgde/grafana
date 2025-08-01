import commonjs from '@rollup/plugin-commonjs';
import json from '@rollup/plugin-json';
import { createRequire } from 'node:module';

import { plugins as basePlugins, cjsOutput, entryPoint, esmOutput } from '../rollup.config.parts';

const rq = createRequire(import.meta.url);
const pkg = rq('./package.json');

const plugins = [
  ...basePlugins,
  json(),
  commonjs({
    include: /node_modules/,
    requireReturnsDefault: 'auto',
  }),
];

export default [
  {
    input: entryPoint,
    plugins,
    output: [cjsOutput(pkg), esmOutput(pkg, 'grafana-alerting')],
    external: ['react', 'react-dom', 'react-router-dom'],
    treeshake: false,
  },
  {
    input: 'src/unstable.ts',
    plugins,
    output: [cjsOutput(pkg), esmOutput(pkg, 'grafana-alerting')],
    external: ['react', 'react-dom', 'react-router-dom'],
    treeshake: false,
  },
  {
    input: 'src/testing.ts',
    plugins,
    output: [cjsOutput(pkg), esmOutput(pkg, 'grafana-alerting')],
    external: ['react', 'react-dom', 'react-router-dom'],
  },
];
