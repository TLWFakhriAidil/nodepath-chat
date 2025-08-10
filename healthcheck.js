// Enhanced healthcheck script for Railway deployment
const http = require('http');
const fs = require('fs');
const { execSync } = require('child_process');

// Set a timeout for the entire healthcheck process
const healthcheckTimeout = setTimeout(() => {
  console.error('Healthcheck timed out after 15 seconds');
  process.exit(1);
}, 15000);

// Check if dist directory exists
if (!fs.existsSync('./dist')) {
  console.error('ERROR: dist directory not found');
  clearTimeout(healthcheckTimeout);
  process.exit(1);
}

// Check if vite is available
try {
  const viteVersion = execSync('vite --version').toString().trim();
  console.log(`Vite version: ${viteVersion}`);
} catch (error) {
  console.log('Vite not found, but continuing with healthcheck...');
}

// Check if PHP is available
try {
  const phpVersion = execSync('php -v').toString().trim().split('\n')[0];
  console.log(`PHP version: ${phpVersion}`);
} catch (error) {
  console.error('Error: PHP is not available', error.message);
  clearTimeout(healthcheckTimeout);
  process.exit(1);
}

// Function to make an HTTP request
function makeRequest(options) {
  return new Promise((resolve, reject) => {
    console.log(`Checking health at http://${options.host}:${options.port}${options.path}`);
    
    const req = http.request(options, (res) => {
      console.log(`STATUS for ${options.path}: ${res.statusCode}`);
      let data = '';
      
      res.on('data', (chunk) => {
        data += chunk;
      });
      
      res.on('end', () => {
        resolve({ statusCode: res.statusCode, data });
      });
    });
    
    req.on('error', (error) => {
      reject(error);
    });
    
    req.setTimeout(options.timeout, () => {
      req.destroy();
      reject(new Error(`Request to ${options.path} timed out`));
    });
    
    req.end();
  });
}

// Run all checks in sequence
makeRequest({
  host: 'localhost',
  port: 4173,
  path: '/',
  timeout: 5000
})
.then(mainResult => {
  if (mainResult.statusCode !== 200) {
    console.error(`Healthcheck failed: Main app returned status code ${mainResult.statusCode}`);
    console.error(`Response body: ${mainResult.data.substring(0, 200)}...`);
    throw new Error('Main app healthcheck failed');
  }
  
  console.log('Main app healthcheck passed');
  
  // Now check the PHP test endpoint
  return makeRequest({
    host: 'localhost',
    port: 4173,
    path: '/test-php.php',
    timeout: 5000
  });
})
.then(phpResult => {
  if (phpResult.statusCode !== 200) {
    console.log(`PHP endpoint returned status code ${phpResult.statusCode}`);
    console.log(`Response body: ${phpResult.data.substring(0, 200)}...`);
    console.log('PHP endpoint check failed, but continuing...');
  } else {
    console.log('PHP endpoint healthcheck passed');
  }
  
  console.log('All required healthchecks passed');
  clearTimeout(healthcheckTimeout);
  process.exit(0);
})
.catch(error => {
  console.error('Healthcheck failed:', error.message);
  clearTimeout(healthcheckTimeout);
  process.exit(1);
});