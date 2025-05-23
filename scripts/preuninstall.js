#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

const binDir = path.join(__dirname, '..', 'bin');
const binaries = ['tree-sorter-ts', 'tree-sorter-ts.exe'];

console.log('Cleaning up tree-sorter-ts binaries...');

binaries.forEach(binary => {
  const binaryPath = path.join(binDir, binary);
  if (fs.existsSync(binaryPath)) {
    try {
      fs.unlinkSync(binaryPath);
      console.log(`Removed ${binary}`);
    } catch (e) {
      console.error(`Failed to remove ${binary}:`, e.message);
    }
  }
});