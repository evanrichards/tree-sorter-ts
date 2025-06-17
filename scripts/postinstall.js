#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const https = require('https');
const { promisify } = require('util');
const { pipeline } = require('stream');
const streamPipeline = promisify(pipeline);

const REPO = 'evanrichards/tree-sorter-ts';
const BINARY_NAME = 'tree-sorter-ts';

const PLATFORM_MAP = {
  darwin: 'darwin',
  linux: 'linux',
  win32: 'windows'
};

const ARCH_MAP = {
  x64: 'amd64',
  arm64: 'arm64'
};

async function getLatestRelease() {
  return new Promise((resolve, reject) => {
    const options = {
      hostname: 'api.github.com',
      path: `/repos/${REPO}/releases/latest`,
      headers: {
        'User-Agent': 'tree-sorter-ts-installer'
      }
    };

    https.get(options, (res) => {
      let data = '';
      res.on('data', (chunk) => data += chunk);
      res.on('end', () => {
        try {
          resolve(JSON.parse(data));
        } catch (e) {
          reject(e);
        }
      });
    }).on('error', reject);
  });
}

async function downloadFile(url, destPath) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(destPath);
    
    https.get(url, (response) => {
      if (response.statusCode === 302 || response.statusCode === 301) {
        // Follow redirect
        https.get(response.headers.location, (redirectResponse) => {
          streamPipeline(redirectResponse, file)
            .then(resolve)
            .catch(reject);
        }).on('error', reject);
      } else {
        streamPipeline(response, file)
          .then(resolve)
          .catch(reject);
      }
    }).on('error', reject);
  });
}

async function install() {
  try {
    const platform = PLATFORM_MAP[process.platform];
    const arch = ARCH_MAP[process.arch];
    
    if (!platform || !arch) {
      console.error(`Unsupported platform: ${process.platform} ${process.arch}`);
      process.exit(1);
    }

    console.log(`Installing tree-sorter-ts for ${platform} ${arch}...`);

    // Create bin directory
    const binDir = path.join(__dirname, '..', 'bin');
    if (!fs.existsSync(binDir)) {
      fs.mkdirSync(binDir, { recursive: true });
    }

    // Try to get latest release
    let downloadUrl;
    let version = 'latest';
    
    try {
      const release = await getLatestRelease();
      version = release.tag_name;
      
      // Find the right asset
      const assetName = `tree-sorter-ts_${platform}_${arch}.tar.gz`;
      const asset = release.assets.find(a => a.name === assetName);
      
      if (asset) {
        downloadUrl = asset.browser_download_url;
      }
    } catch (e) {
      console.log('Could not fetch latest release, falling back to direct download...');
    }

    // Fallback to direct download if release API fails
    if (!downloadUrl) {
      downloadUrl = `https://github.com/${REPO}/releases/latest/download/tree-sorter-ts_${platform}_${arch}.tar.gz`;
    }

    const tempFile = path.join(binDir, 'temp.tar.gz');
    
    console.log(`Downloading from ${downloadUrl}...`);
    await downloadFile(downloadUrl, tempFile);

    // Extract the binary
    const tar = require('tar');
    await tar.x({
      file: tempFile,
      cwd: binDir,
      filter: (path) => path.endsWith(BINARY_NAME) || path.endsWith(`${BINARY_NAME}.exe`)
    });

    // Clean up
    fs.unlinkSync(tempFile);

    // Make binary executable on Unix
    if (platform !== 'windows') {
      const binaryPath = path.join(binDir, BINARY_NAME);
      fs.chmodSync(binaryPath, 0o755);
    }

    console.log(`Successfully installed tree-sorter-ts ${version}`);
  } catch (error) {
    console.error('Installation failed:', error.message);
    console.error('You can manually download the binary from:');
    console.error(`https://github.com/${REPO}/releases`);
    process.exit(1);
  }
}

install();