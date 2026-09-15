#!/usr/bin/env node
// Postinstall: downloads the platform binary from GitHub releases.
// Exits 0 on failure so `npm install` never breaks due to download issues.

const os = require('os');
const fs = require('fs');
const path = require('path');
const https = require('https');
const { execSync } = require('child_process');

const { version } = require('./package.json');
const REPO = 'wso2/integration-platform-tools';
const BINARY = 'wso2-integration-platform';
const BIN_DIR = path.join(__dirname, 'bin');

const PLATFORMS = { linux: 'linux', darwin: 'darwin', win32: 'windows' };
const ARCHS = { x64: 'amd64', arm64: 'arm64', arm: 'arm', ia32: '386' };

const platform = PLATFORMS[os.platform()];
const arch = ARCHS[os.arch()];

if (!platform || !arch) {
  console.warn(`[${BINARY}] Unsupported platform: ${os.platform()}/${os.arch()}`);
  console.warn(`Install manually: https://github.com/${REPO}/releases`);
  process.exit(0);
}

const isWindows = platform === 'windows';
const binFile = isWindows ? `${BINARY}.exe` : BINARY;
const binPath = path.join(BIN_DIR, binFile);

// GitHub release tags use a "v" prefix.
const tag = `v${version}`;

// Linux releases ship a standalone CLI archive; Mac/Windows releases only ship
// the .mcpb Claude extension (which is a zip containing server/<binary>).
const isLinux = platform === 'linux';
const artifact = isLinux
  ? `${BINARY}-${tag}-${platform}-${arch}.tar.gz`
  : `${BINARY}-${tag}-${platform}-${arch}.mcpb`;
const url = `https://github.com/${REPO}/releases/download/${tag}/${artifact}`;

if (fs.existsSync(binPath)) {
  process.exit(0);
}

function download(url, dest) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest);
    function get(u) {
      https.get(u, (res) => {
        if (res.statusCode === 301 || res.statusCode === 302) {
          return get(res.headers.location);
        }
        if (res.statusCode !== 200) {
          return reject(new Error(`HTTP ${res.statusCode} downloading ${artifact}`));
        }
        res.pipe(file);
        file.on('finish', () => { file.close(); resolve(); });
        file.on('error', reject);
      }).on('error', reject);
    }
    get(url);
  });
}

function extract(archive, destDir) {
  if (archive.endsWith('.tar.gz')) {
    // Linux CLI archive: binary is at the root of the tar
    execSync(`tar -xzf "${archive}" -C "${destDir}"`);
  } else {
    // .mcpb (zip): binary lives inside server/ subdirectory
    const tmpDir = path.join(os.tmpdir(), `mcpb-extract-${process.pid}`);
    fs.mkdirSync(tmpDir, { recursive: true });
    if (isWindows) {
      execSync(`tar -xf "${archive}" -C "${tmpDir}"`);
    } else {
      execSync(`unzip -o "${archive}" -d "${tmpDir}"`);
    }
    const extracted = path.join(tmpDir, 'server', binFile);
    fs.copyFileSync(extracted, path.join(destDir, binFile));
    fs.rmSync(tmpDir, { recursive: true, force: true });
  }
}

async function main() {
  fs.mkdirSync(BIN_DIR, { recursive: true });
  const tmp = path.join(os.tmpdir(), artifact);

  console.log(`[${BINARY}] Downloading ${artifact}...`);
  await download(url, tmp);
  extract(tmp, BIN_DIR);
  fs.unlinkSync(tmp);

  if (!isWindows) {
    fs.chmodSync(binPath, '755');
  }
  console.log(`[${BINARY}] Installed successfully.`);
}

main().catch((err) => {
  console.warn(`[${BINARY}] Download failed: ${err.message}`);
  console.warn(`Install manually: https://github.com/${REPO}/releases`);
  process.exit(0);
});
