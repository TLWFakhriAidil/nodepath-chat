#!/bin/bash
# SSH Key Setup Script for Railway SSH Tunnel
# This script helps generate and configure SSH keys for the tunnel service

set -e

echo "=== SSH Key Setup for Railway SSH Tunnel ==="
echo "This script will help you set up SSH keys for secure database connection"
echo ""

# Configuration
KEY_NAME="railway-tunnel-key"
KEY_DIR="./ssh-keys"
KEY_PATH="$KEY_DIR/$KEY_NAME"
PUB_KEY_PATH="$KEY_PATH.pub"

# Create SSH keys directory
echo "Creating SSH keys directory..."
mkdir -p "$KEY_DIR"

# Check if key already exists
if [ -f "$KEY_PATH" ]; then
    echo "SSH key already exists at $KEY_PATH"
    read -p "Do you want to overwrite it? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Using existing key..."
    else
        echo "Removing existing key..."
        rm -f "$KEY_PATH" "$PUB_KEY_PATH"
    fi
fi

# Generate SSH key if it doesn't exist
if [ ! -f "$KEY_PATH" ]; then
    echo "Generating new SSH key pair..."
    ssh-keygen -t rsa -b 4096 -f "$KEY_PATH" -N "" -C "railway-tunnel@nodepath-chat"
    echo "SSH key pair generated successfully!"
fi

# Set proper permissions
echo "Setting proper permissions..."
chmod 600 "$KEY_PATH"
chmod 644 "$PUB_KEY_PATH"

# Display public key
echo ""
echo "=== PUBLIC KEY (copy this to your SSH server) ==="
echo "File: $PUB_KEY_PATH"
echo ""
cat "$PUB_KEY_PATH"
echo ""
echo "=== END PUBLIC KEY ==="
echo ""

# Display private key for Railway
echo "=== PRIVATE KEY (add this to Railway secrets as SSH_PRIVATE_KEY) ==="
echo "File: $KEY_PATH"
echo ""
echo "Copy the entire content below (including BEGIN/END lines):"
echo ""
cat "$KEY_PATH"
echo ""
echo "=== END PRIVATE KEY ==="
echo ""

# Instructions
echo "=== SETUP INSTRUCTIONS ==="
echo ""
echo "1. COPY PUBLIC KEY TO YOUR SSH SERVER:"
echo "   Run this command on your SSH server (replace user@server):"
echo "   echo '$(cat "$PUB_KEY_PATH")' >> ~/.ssh/authorized_keys"
echo ""
echo "2. TEST SSH CONNECTION:"
echo "   ssh -i $KEY_PATH user@your-server.com"
echo ""
echo "3. ADD PRIVATE KEY TO RAILWAY:"
echo "   - Go to Railway dashboard > Your Project > Variables"
echo "   - Add new variable: SSH_PRIVATE_KEY"
echo "   - Paste the entire private key content (including BEGIN/END lines)"
echo ""
echo "4. UPDATE RAILWAY ENVIRONMENT VARIABLES:"
echo "   TUNNEL_HOST=your-server.com"
echo "   TUNNEL_USER=your-username"
echo "   TUNNEL_PORT=22"
echo ""
echo "5. TEST MYSQL CONNECTION FROM YOUR SSH SERVER:"
echo "   mysql -h 159.89.198.71 -P 3306 -u admin_aqil -p admin_railway"
echo ""

# Create a test script
echo "Creating test script..."
cat > "$KEY_DIR/test-connection.sh" << 'EOF'
#!/bin/bash
# Test script for SSH tunnel connection

set -e

echo "Testing SSH connection..."
SSH_KEY="./railway-tunnel-key"
SSH_USER="${TUNNEL_USER:-your-username}"
SSH_HOST="${TUNNEL_HOST:-your-server.com}"
SSH_PORT="${TUNNEL_PORT:-22}"

if [ ! -f "$SSH_KEY" ]; then
    echo "Error: SSH key not found at $SSH_KEY"
    echo "Run ssh-key-setup.sh first"
    exit 1
fi

echo "Testing SSH connection to $SSH_USER@$SSH_HOST:$SSH_PORT..."
ssh -i "$SSH_KEY" -o StrictHostKeyChecking=no -o ConnectTimeout=10 \
    "$SSH_USER@$SSH_HOST" "echo 'SSH connection successful!'"

echo "Testing MySQL connectivity from SSH server..."
ssh -i "$SSH_KEY" -o StrictHostKeyChecking=no \
    "$SSH_USER@$SSH_HOST" \
    "timeout 10 bash -c '</dev/tcp/159.89.198.71/3306' && echo 'MySQL port is reachable' || echo 'MySQL port is not reachable'"

echo "Testing SSH tunnel locally..."
echo "Starting local tunnel (press Ctrl+C to stop)..."
ssh -i "$SSH_KEY" -o StrictHostKeyChecking=no \
    -L 3307:159.89.198.71:3306 \
    "$SSH_USER@$SSH_HOST" \
    "echo 'Tunnel established. Test with: mysql -h 127.0.0.1 -P 3307 -u admin_aqil -p admin_railway'"
EOF

chmod +x "$KEY_DIR/test-connection.sh"

echo "=== NEXT STEPS ==="
echo ""
echo "1. Copy the public key to your SSH server"
echo "2. Test the connection: cd $KEY_DIR && ./test-connection.sh"
echo "3. Add the private key to Railway secrets"
echo "4. Deploy the SSH tunnel service"
echo ""
echo "Files created:"
echo "  - $KEY_PATH (private key)"
echo "  - $PUB_KEY_PATH (public key)"
echo "  - $KEY_DIR/test-connection.sh (test script)"
echo ""
echo "Setup complete! 🚀"