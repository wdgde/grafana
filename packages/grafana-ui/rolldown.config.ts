import { createRequire } from 'node:module';
import { dirname, resolve } from 'node:path';
import { dts } from 'rolldown-plugin-dts';
import copy from 'rollup-plugin-copy';
import { nodeExternals } from 'rollup-plugin-node-externals';
import svg from 'rollup-plugin-svg-import';
const rq = createRequire(import.meta.url);
const icons = rq('../../public/app/core/icons/cached.json');
const pkg = rq('./package.json');

const projectCwd = process.env.PROJECT_CWD ?? '../../';

const iconSrcPaths = icons.map((iconSubPath) => {
  // eslint-disable-next-line @grafana/no-restricted-img-srcs
  return `../../public/img/icons/${iconSubPath}.svg`;
});
const commonPlugins = [nodeExternals({ deps: true, packagePath: './package.json' }), svg({ stringify: true })];
const inputs = ['./src/index.ts', './src/unstable.ts'];

export default [
  {
    input: inputs,
    output: {
      format: 'esm',
      sourcemap: true,
      dir: dirname(pkg.publishConfig.module),
      entryFileNames: '[name].mjs',
      // preserveModules: true,
      // preserveModulesRoot: resolve(projectCwd, `packages/grafana-ui/src`),
    },
    plugins: [
      ...commonPlugins,
      dts(),
      copy({ targets: [{ src: iconSrcPaths, dest: './dist/public/' }], flatten: false }),
    ],
  },
  {
    input: inputs,
    output: {
      format: 'cjs',
      sourcemap: true,
      dir: dirname(pkg.publishConfig.main),
      entryFileNames: '[name].cjs',
    },
    plugins: [...commonPlugins],
  },
  {
    input: inputs,
    output: {
      dir: dirname(pkg.publishConfig.main),
    },
    plugins: [...commonPlugins, dts({ emitDtsOnly: true })],
  },
];
