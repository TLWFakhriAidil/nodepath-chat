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
    if (req.url === '/mysql-api.php') {
      // Serve the PHP file content directly
      try {
        const phpContent = fs.readFileSync('./mysql-api.php', 'utf8');
        res.writeHead(200, { 'Content-Type': 'text/plain' });
        res.end(phpContent);
      } catch (error) {
        res.writeHead(500);
        res.end('Internal Server Error');
      }
      return;
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

# Start the application
echo "Starting application..."
exec vite preview --host 0.0.0.0 --port 4173