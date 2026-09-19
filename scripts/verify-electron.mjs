import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const repoRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const electronDir = path.join(repoRoot, 'node_modules', 'electron');

const nodeMajor = Number(process.versions.node.split('.')[0]);
if (nodeMajor >= 24) {
  console.warn(
    `[postinstall] Node ${process.versions.node} can prevent Electron from installing correctly. ` +
      'Use Node 20 or 22 LTS (see package.json engines).',
  );
}

if (!fs.existsSync(path.join(electronDir, 'package.json'))) {
  process.exit(0);
}

let platformPath;
try {
  platformPath = fs.readFileSync(path.join(electronDir, 'path.txt'), 'utf8').trim();
} catch {
  platformPath = '';
}

const binary = platformPath ? path.join(electronDir, 'dist', platformPath) : '';

if (!platformPath || !fs.existsSync(binary)) {
  console.error(
    '[postinstall] Electron failed to install correctly (missing path.txt or binary).\n' +
      '  Use Node 20 or 22 LTS, then:\n' +
      '    rm -rf node_modules/electron && yarn install',
  );
  process.exit(1);
}
