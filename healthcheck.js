// Enhanced healthcheck script for Railway deployment
const http = require('http');
const fs = require('fs');

// Check if dist directory exists
if (!fs.existsSync('./dist')) {
  console.error('ERROR: dist directory not found');
  process.exit(1);
}

// Check if vite is available
const { execSync } = require('child_process');
try {
  const viteVersion = execSync('vite --version').toString().trim();
  console.log(`Vite version: ${viteVersion}`);
} catch (error) {
  console.log('Vite not found, but continuing with healthcheck...');
}

// Perform HTTP check
const options = {
  host: 'localhost',
  port: 4173,
  path: '/',
  timeout: 5000
};

console.log(`Checking health at http://${options.host}:${options.port}${options.path}`);

const request = http.request(options, (res) => {
  console.log(`STATUS: ${res.statusCode}`);
  let data = '';
  
  res.on('data', (chunk) => {
    data += chunk;
  });
  
  res.on('end', () => {
    if (res.statusCode === 200) {
      console.log('Healthcheck successful');
      process.exit(0);
    } else {
      console.log(`Healthcheck failed with status: ${res.statusCode}`);
      console.log(`Response body: ${data.substring(0, 200)}...`);
      process.exit(1);
    }
  });
});

request.on('error', (err) => {
  console.error(`ERROR: ${err.message}`);
  process.exit(1);
});

// Set timeout for the entire healthcheck
request.setTimeout(options.timeout, () => {
  console.error('Healthcheck timed out');
  request.destroy();
  process.exit(1);
});

request.end();