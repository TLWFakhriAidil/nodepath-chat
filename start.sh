#!/bin/bash

# Display environment information
echo "Node.js version: $(node -v)"
echo "npm version: $(npm -v)"

# Check if dist directory exists
if [ ! -d "./dist" ]; then
  echo "Error: dist directory not found. The build may have failed."
  echo "Creating a simple static page as fallback..."
  mkdir -p dist
  cp ./public/index.html ./dist/index.html
fi

# Check if vite is installed globally, if not install it
if ! command -v vite &> /dev/null; then
  echo "Vite not found, installing globally..."
  npm install -g vite
  
  # Double-check if vite was installed successfully
  if ! command -v vite &> /dev/null; then
    echo "Failed to install vite globally. Using node to serve static files..."
    # Create a simple static file server as fallback
    cat > server.js << 'EOF'
const http = require('http');
const fs = require('fs');
const path = require('path');

const PORT = 4173;
const DIST_DIR = './dist';

const MIME_TYPES = {
  '.html': 'text/html',
  '.js': 'text/javascript',
  '.css': 'text/css',
  '.json': 'application/json',
  '.png': 'image/png',
  '.jpg': 'image/jpeg',
  '.gif': 'image/gif',
  '.svg': 'image/svg+xml',
  '.ico': 'image/x-icon',
};

const server = http.createServer((req, res) => {
  console.log(`${new Date().toISOString()} - ${req.method} ${req.url}`);
  
  // Handle PHP files separately
  if (req.url.endsWith('.php')) {
    if (req.url === '/mysql-api.php' || req.url === '/test-php.php' || req.url === '/debug-api.php' || req.url === '/db-test.php' || req.url === '/request-logger.php') {
      // Collect request body for POST requests
      if (req.method === 'POST') {
        let body = '';
        req.on('data', chunk => {
          body += chunk.toString();
        });
        
        req.on('end', () => {
          // Set up environment variables to simulate PHP $_SERVER and pass the request body
          const env = {
            ...process.env,
            REQUEST_METHOD: req.method,
            CONTENT_TYPE: req.headers['content-type'] || 'application/json',
            HTTP_ACCEPT: req.headers['accept'] || '*/*',
            QUERY_STRING: req.url.includes('?') ? req.url.split('?')[1] : '',
            // Add headers as environment variables
            ...Object.keys(req.headers).reduce((acc, header) => {
              const headerEnvName = 'HTTP_' + header.toUpperCase().replace(/-/g, '_');
              acc[headerEnvName] = req.headers[header];
              return acc;
            }, {})
          };
          
          // Create a temporary file with the request body
          const tempInputFile = `./temp_input_${Date.now()}.json`;
          try {
            fs.writeFileSync(tempInputFile, body);
            
            // Execute PHP with the request body piped in
            const { spawn } = require('child_process');
            const php = spawn('php', ['./mysql-api.php'], { env });
            
            // Pipe the request body to PHP's stdin
            const inputStream = fs.createReadStream(tempInputFile);
            inputStream.pipe(php.stdin);
            
            let stdout = '';
            let stderr = '';
            
            php.stdout.on('data', (data) => {
              stdout += data.toString();
            });
            
            php.stderr.on('data', (data) => {
              stderr += data.toString();
              console.error(`PHP stderr: ${data}`);
            });
            
            php.on('close', (code) => {
              // Clean up the temporary file
              try { fs.unlinkSync(tempInputFile); } catch (e) { /* ignore */ }
              
              if (code !== 0) {
                console.error(`PHP process exited with code ${code}`);
                res.writeHead(500, { 'Content-Type': 'application/json' });
                res.end(JSON.stringify({ 
                  success: false, 
                  error: 'PHP execution failed', 
                  details: stderr,
                  code: code
                }));
                return;
              }
              
              console.log(`PHP output: ${stdout}`);
              
              // Set CORS headers
              res.writeHead(200, { 
                'Content-Type': 'application/json',
                'Access-Control-Allow-Origin': '*',
                'Access-Control-Allow-Methods': 'POST, OPTIONS',
                'Access-Control-Allow-Headers': 'Content-Type, Accept'
              });
              res.end(stdout || JSON.stringify({ success: true, data: [] }));
            });
          } catch (error) {
            console.error(`Error handling PHP request: ${error}`);
            try { fs.unlinkSync(tempInputFile); } catch (e) { /* ignore */ }
            res.writeHead(500, { 'Content-Type': 'application/json' });
            res.end(JSON.stringify({ success: false, error: 'Internal server error' }));
          }
        });
        return;
      } else if (req.method === 'GET') {
        // Handle GET requests for PHP files
        const { spawn } = require('child_process');
        
        // Create environment variables for PHP
        const env = {
          ...process.env,
          REQUEST_METHOD: req.method,
          QUERY_STRING: req.url.includes('?') ? req.url.split('?')[1] : '',
          // Add headers as environment variables
          ...Object.keys(req.headers).reduce((acc, header) => {
            const headerEnvName = 'HTTP_' + header.toUpperCase().replace(/-/g, '_');
            acc[headerEnvName] = req.headers[header];
            return acc;
          }, {})
        };
        
        // Execute the PHP file based on the URL
        let phpFile = './mysql-api.php';
        if (req.url === '/test-php.php') phpFile = './test-php.php';
        if (req.url === '/debug-api.php') phpFile = './debug-api.php';
        if (req.url === '/db-test.php') phpFile = './db-test.php';
        if (req.url === '/request-logger.php') phpFile = './request-logger.php';
        const php = spawn('php', [phpFile], { env });
        
        let stdout = '';
        let stderr = '';
        
        php.stdout.on('data', (data) => {
          stdout += data.toString();
        });
        
        php.stderr.on('data', (data) => {
          stderr += data.toString();
          console.error(`PHP stderr: ${data}`);
        });
        
        php.on('close', (code) => {
          if (code !== 0) {
            console.error(`PHP process exited with code ${code}`);
            res.writeHead(500, { 'Content-Type': 'application/json' });
            res.end(JSON.stringify({ 
              success: false, 
              error: 'PHP execution failed', 
              details: stderr,
              code: code
            }));
            return;
          }
          
          console.log(`PHP output: ${stdout}`);
          
          // Set CORS headers
          res.writeHead(200, { 
            'Content-Type': 'application/json',
            'Access-Control-Allow-Origin': '*',
            'Access-Control-Allow-Methods': 'GET, POST, OPTIONS',
            'Access-Control-Allow-Headers': 'Content-Type, Accept'
          });
          res.end(stdout || JSON.stringify({ success: true, data: [] }));
        });
        return;
      } else if (req.method === 'OPTIONS') {
        // Handle CORS preflight requests
        res.writeHead(200, {
          'Access-Control-Allow-Origin': '*',
          'Access-Control-Allow-Methods': 'GET, POST, OPTIONS',
          'Access-Control-Allow-Headers': 'Content-Type, Accept',
          'Content-Type': 'application/json'
        });
        res.end();
        return;
      } else {
        // Method not allowed
        res.writeHead(405, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify({ success: false, error: 'Method not allowed' }));
        return;
      }
    }
  }

  // Normalize the URL
  let filePath = path.join(DIST_DIR, req.url);
  if (filePath === path.join(DIST_DIR, '/')) {
    filePath = path.join(DIST_DIR, '/index.html');
  }

  // Get the file extension
  const extname = path.extname(filePath);
  const contentType = MIME_TYPES[extname] || 'application/octet-stream';

  // Read the file
  fs.readFile(filePath, (error, content) => {
    if (error) {
      if (error.code === 'ENOENT') {
        // File not found, try serving index.html
        fs.readFile(path.join(DIST_DIR, '/index.html'), (err, indexContent) => {
          if (err) {
            res.writeHead(404);
            res.end('404 Not Found');
          } else {
            res.writeHead(200, { 'Content-Type': 'text/html' });
            res.end(indexContent, 'utf-8');
          }
        });
      } else {
        // Server error
        res.writeHead(500);
        res.end(`Server Error: ${error.code}`);
      }
    } else {
      // Success
      res.writeHead(200, { 'Content-Type': contentType });
      res.end(content, 'utf-8');
    }
  });
});

server.listen(PORT, '0.0.0.0', () => {
  console.log(`Server running at http://0.0.0.0:${PORT}/`);
});
EOF
    # Start the fallback server
    exec node server.js
  fi
fi

# Try to get vite version
VITE_VERSION=$(vite --version 2>/dev/null || echo "not found")
echo "Vite version: $VITE_VERSION"

# Check PHP version and modules
PHP_VERSION=$(php -v 2>/dev/null | head -n 1 || echo "PHP not found")
echo "PHP version: $PHP_VERSION"

# List PHP modules
echo "PHP modules:"
php -m 2>/dev/null || echo "Could not list PHP modules"

# Check database environment variables
echo "Database environment variables:"
echo "DB_HOST: ${DB_HOST:-not set}"
echo "DB_NAME: ${DB_NAME:-not set}"
echo "DB_USER: ${DB_USER:-not set}"
echo "DB_PASSWORD: ${DB_PASSWORD:-masked}"
echo "DB_PORT: ${DB_PORT:-not set}"

# If environment variables are not set, use defaults
if [ -z "$DB_HOST" ]; then
  export DB_HOST="159.89.198.71"
  echo "Using default DB_HOST: $DB_HOST"
fi

if [ -z "$DB_NAME" ]; then
  export DB_NAME="admin_railway"
  echo "Using default DB_NAME: $DB_NAME"
fi

if [ -z "$DB_USER" ]; then
  export DB_USER="admin_aqil"
  echo "Using default DB_USER: $DB_USER"
fi

if [ -z "$DB_PASSWORD" ]; then
  export DB_PASSWORD="admin_aqil"
  echo "Using default DB_PASSWORD: [masked]"
fi

if [ -z "$DB_PORT" ]; then
  export DB_PORT="3306"
  echo "Using default DB_PORT: $DB_PORT"
fi

# Start the application
echo "Starting application..."
exec vite preview --host 0.0.0.0 --port 4173