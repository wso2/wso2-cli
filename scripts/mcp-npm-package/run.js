#!/usr/bin/env node
const path = require('path');
const { spawnSync } = require('child_process');
const os = require('os');

const BINARY = 'wso2-integration-platform';
const isWindows = os.platform() === 'win32';
const binPath = path.join(__dirname, 'bin', isWindows ? `${BINARY}.exe` : BINARY);

if (!require('fs').existsSync(binPath)) {
  console.error(`[${BINARY}] Binary not found at ${binPath}`);
  console.error(`Try reinstalling: npm install @pcnfernando-wso2/integration-platform-mcp`);
  process.exit(1);
}

const result = spawnSync(binPath, process.argv.slice(2), { stdio: 'inherit' });
process.exit(result.status ?? 1);
