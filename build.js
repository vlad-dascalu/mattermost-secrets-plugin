const { execSync } = require('child_process');
const path = require('path');
const fs = require('fs');

// Create necessary directories
console.log('Creating build directories...');
if (!fs.existsSync('build')) {
  fs.mkdirSync('build');
}
if (!fs.existsSync('build/webapp')) {
  fs.mkdirSync('build/webapp');
}
if (!fs.existsSync('build/server')) {
  fs.mkdirSync('build/server');
}
if (!fs.existsSync('build/server/dist')) {
  fs.mkdirSync('build/server/dist');
}

// Build webapp
console.log('Building webapp...');
try {
  process.chdir('webapp');
  execSync('npm install --legacy-peer-deps', { stdio: 'inherit' });
  execSync('npm run build', { stdio: 'inherit' });
  process.chdir('..');
  
  // Copy files to build directory
  console.log('Copying webapp files to build directory...');
  fs.cpSync('webapp/dist', 'build/webapp/dist', { recursive: true });
} catch (error) {
  console.error('Error building webapp:', error);
  process.exit(1);
}

// Build server
console.log('Building server...');
try {
  process.chdir('server');
  execSync('go mod tidy', { stdio: 'inherit' });
  
  // Build for the current platform
  const platform = process.platform;
  const arch = process.arch === 'x64' ? 'amd64' : process.arch;
  let outputName;
  
  if (platform === 'win32') {
    outputName = 'plugin-windows-amd64.exe';
  } else if (platform === 'darwin') {
    outputName = 'plugin-darwin-amd64';
  } else {
    outputName = 'plugin-linux-amd64';
  }
  
  execSync(`go build -o ../build/server/dist/${outputName}`, { stdio: 'inherit' });
  process.chdir('..');
} catch (error) {
  console.error('Error building server:', error);
  process.exit(1);
}

// Copy plugin.json to build directory
console.log('Copying plugin.json to build directory...');
fs.copyFileSync('plugin.json', 'build/plugin.json');

// Copy assets to build directory if they exist
if (fs.existsSync('assets')) {
  console.log('Copying assets to build directory...');
  if (!fs.existsSync('build/assets')) {
    fs.mkdirSync('build/assets');
  }
  fs.cpSync('assets', 'build/assets', { recursive: true });
}

console.log('Build completed successfully!'); 