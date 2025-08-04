import alias from '@rollup/plugin-alias';
import commonjs from '@rollup/plugin-commonjs';
import json from '@rollup/plugin-json';
import { createRequire } from 'node:module';
import { dirname, resolve as pathResolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import path from 'path';
// import tsConfigPaths from 'rollup-plugin-tsconfig-paths';
// import { typescriptPaths } from 'rollup-plugin-typescript-paths';

import { plugins as basePlugins, cjsOutput, entryPoint, esmOutput } from '../rollup.config.parts';

const __dirname = dirname(fileURLToPath(import.meta.url));

console.log('__dirname', __dirname);
console.log('pathResolve(__dirname, "public/app")', pathResolve(__dirname, '../../public/app'));

const rq = createRequire(import.meta.url);
const pkg = rq('./package.json');

// Custom plugin to transform app/ imports to relative paths
// const transformAppImports = () => {
//   return {
//     name: 'transform-app-imports',
//     renderChunk(code, chunk, options) {
//       // Transform app/ imports to relative paths
//       return code.replace(/from ['"]app\/([^'"]+)['"]/g, (match, importPath) => {
//         // Create a shorter relative path that goes directly to the public/app directory
//         // Since all files are in dist/esm/public/app/..., we can use a shorter path
//         const relativePath = '../../../../../../../../../public/app';
//         return `from '${relativePath}/${importPath}'`;
//       });
//     },
//   };
// };

// console.log(typescriptPaths);

const plugins = [
  ...basePlugins,
  // typescriptPaths({
  //   tsConfigPath: './tsconfig.build.json',
  // }),
  // alias({
  //   entries: [{ find: 'app', replacement: pathResolve(__dirname, '../../public/app') }],
  // }),
  // transformAppImports(),
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
