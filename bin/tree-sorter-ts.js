#!/usr/bin/env node

const { spawn } = require('child_process');
const path = require('path');
const fs = require('fs');

const isWindows = process.platform === 'win32';
const binaryName = isWindows ? 'tree-sorter-ts.exe' : 'tree-sorter-ts';
const binaryPath = path.join(__dirname, binaryName);

// Check if binary exists
if (!fs.existsSync(binaryPath)) {
  console.error('Binary not found. Please run npm install again.');
  console.error(`Expected binary at: ${binaryPath}`);
  process.exit(1);
}

// Spawn the actual binary with all arguments
const child = spawn(binaryPath, process.argv.slice(2), {
  stdio: 'inherit',
  shell: false
});

child.on('error', (err) => {
  console.error('Failed to start tree-sorter-ts:', err.message);
  process.exit(1);
});

child.on('exit', (code) => {
  process.exit(code || 0);
});