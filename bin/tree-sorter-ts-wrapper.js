#!/usr/bin/env node

const { spawn } = require('child_process');
const path = require('path');
const fs = require('fs');

// Map Node.js platform/arch to npm package names
const PLATFORM_MAP = {
  darwin: 'darwin',
  linux: 'linux',
  win32: 'win32'
};

const ARCH_MAP = {
  x64: 'x64',
  arm64: 'arm64'
};

function getBinaryPath() {
  const platform = PLATFORM_MAP[process.platform];
  const arch = ARCH_MAP[process.arch];
  
  if (!platform || !arch) {
    console.error(`Unsupported platform: ${process.platform} ${process.arch}`);
    process.exit(1);
  }

  const packageName = `tree-sorter-ts-${platform}-${arch}`;
  const binaryName = `tree-sorter-ts${process.platform === 'win32' ? '.exe' : ''}`;
  
  try {
    // Try to resolve the platform-specific package
    const packagePath = require.resolve(`${packageName}/package.json`);
    const packageDir = path.dirname(packagePath);
    const binaryPath = path.join(packageDir, binaryName);
    
    if (fs.existsSync(binaryPath)) {
      return binaryPath;
    }
  } catch (e) {
    // Package not found
  }
  
  console.error(`
tree-sorter-ts binary not found!

It looks like the optional dependency for your platform (${platform}-${arch}) was not installed.

Try running:
  npm install ${packageName}

Or reinstall tree-sorter-ts:
  npm install -g tree-sorter-ts
`);
  process.exit(1);
}

// Get the binary path
const binaryPath = getBinaryPath();

// Run the binary with all arguments
const child = spawn(binaryPath, process.argv.slice(2), {
  stdio: 'inherit',
  shell: false
});

child.on('error', (err) => {
  console.error('Failed to execute tree-sorter-ts:', err.message);
  process.exit(1);
});

child.on('exit', (code, signal) => {
  if (signal) {
    process.kill(process.pid, signal);
  } else {
    process.exit(code !== null ? code : 1);
  }
});