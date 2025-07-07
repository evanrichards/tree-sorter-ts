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

async function getRelease(version) {
  return new Promise((resolve, reject) => {
    // If no version specified, get latest
    const path = version 
      ? `/repos/${REPO}/releases/tags/v${version}`
      : `/repos/${REPO}/releases/latest`;
    
    const options = {
      hostname: 'api.github.com',
      path: path,
      headers: {
        'User-Agent': 'tree-sorter-ts-installer'
      }
    };

    https.get(options, (res) => {
      let data = '';
      res.on('data', (chunk) => data += chunk);
      res.on('end', () => {
        try {
          if (res.statusCode !== 200) {
            reject(new Error(`Failed to fetch release${version ? ` v${version}` : ''}: ${res.statusCode}`));
          } else {
            resolve(JSON.parse(data));
          }
        } catch (e) {
          reject(e);
        }
      });
    }).on('error', reject);
  });
}

function getInstalledVersion() {
  try {
    const packagePath = path.join(__dirname, '..', 'package.json');
    const packageJson = JSON.parse(fs.readFileSync(packagePath, 'utf8'));
    return packageJson.version;
  } catch (e) {
    console.error('Could not read package.json:', e.message);
    return null;
  }
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

    // Get the installed package version
    const installedVersion = getInstalledVersion();
    let downloadUrl;
    let version = installedVersion || 'latest';
    
    try {
      // Fetch the specific release for the installed version
      const release = await getRelease(installedVersion);
      version = release.tag_name || version;
      
      // Find the right asset
      const assetName = `tree-sorter-ts_${platform}_${arch}.tar.gz`;
      const asset = release.assets.find(a => a.name === assetName);
      
      if (asset) {
        downloadUrl = asset.browser_download_url;
      } else {
        throw new Error(`Binary not found for ${platform}-${arch} in release ${version}`);
      }
    } catch (e) {
      if (installedVersion) {
        console.error(`Failed to fetch release v${installedVersion}: ${e.message}`);
        console.error('Please ensure this version has been released with binaries.');
        process.exit(1);
      } else {
        console.log('Could not determine installed version, downloading latest release...');
        // Fallback to latest release
        downloadUrl = `https://github.com/${REPO}/releases/latest/download/tree-sorter-ts_${platform}_${arch}.tar.gz`;
      }
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

    console.log(`Successfully installed tree-sorter-ts ${installedVersion ? `v${installedVersion}` : version}`);
  } catch (error) {
    console.error('Installation failed:', error.message);
    console.error('You can manually download the binary from:');
    console.error(`https://github.com/${REPO}/releases`);
    process.exit(1);
  }
}

install();