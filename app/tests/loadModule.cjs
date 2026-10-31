const { readFileSync } = require('node:fs');
const { dirname, resolve } = require('node:path');
const ts = require('typescript');

// Transpiles a TypeScript module of src/ to CommonJS and runs it, resolving its
// relative imports the same way, so the pure form and filter helpers are tested
// without a bundler. Type-only imports vanish in the transpile.
function loadModule(path) {
  const source = readFileSync(path, 'utf8');
  const compiled = ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS } });
  const exports = {};
  const requireSibling = (specifier) => loadModule(resolve(dirname(path), `${specifier}.ts`));
  new Function('exports', 'require', compiled.outputText)(exports, requireSibling);
  return exports;
}

module.exports = { loadModule };
