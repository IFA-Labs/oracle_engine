#!/bin/bash
# ==============================================================================
# Oracle Engine - Server Setup Script
# Run this on a fresh Ubuntu/Debian server to prepare for deployment
# ==============================================================================

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   log_error "This script must be run as root"
   exit 1
fi

log_info "Starting Oracle Engine server setup..."

# ==============================================================================
# System Updates
# ==============================================================================
log_info "Updating system packages..."
apt-get update -y
apt-get upgrade -y

# ==============================================================================
# Install Docker
# ==============================================================================
log_info "Installing Docker..."
if ! command -v docker &> /dev/null; then
    curl -fsSL https://get.docker.com -o get-docker.sh
    sh get-docker.sh
    rm get-docker.sh
    
    # Enable and start Docker
    systemctl enable docker
    systemctl start docker
else
    log_info "Docker already installed"
fi

# ==============================================================================
# Install Docker Compose
# ==============================================================================
log_info "Installing Docker Compose..."
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    apt-get install -y docker-compose-plugin
else
    log_info "Docker Compose already installed"
fi

# ==============================================================================
# Create application user
# ==============================================================================
APP_USER="oracle"
log_info "Creating application user: ${APP_USER}..."
if ! id "$APP_USER" &>/dev/null; then
    useradd -m -s /bin/bash -G docker "$APP_USER"
    log_info "User ${APP_USER} created and added to docker group"
else
    log_info "User ${APP_USER} already exists"
    usermod -aG docker "$APP_USER"
fi

# ==============================================================================
# Create application directories
# ==============================================================================
APP_DIR="/var/www/oracle_engine"
log_info "Creating application directory: ${APP_DIR}..."
mkdir -p "$APP_DIR"
mkdir -p "$APP_DIR/logs"
mkdir -p "$APP_DIR/backups"
chown -R "$APP_USER:$APP_USER" "$APP_DIR"

# ==============================================================================
# Install additional tools
# ==============================================================================
log_info "Installing additional tools..."
apt-get install -y \
    curl \
    wget \
    git \
    htop \
    vim \
    unzip \
    jq \
    fail2ban \
    ufw

# ==============================================================================
# Configure Firewall
# ==============================================================================
log_info "Configuring firewall..."
ufw default deny incoming
ufw default allow outgoing
ufw allow ssh
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable

# ==============================================================================
# Configure Fail2ban
# ==============================================================================
log_info "Configuring Fail2ban..."
systemctl enable fail2ban
systemctl start fail2ban

# ==============================================================================
# Install Certbot for SSL
# ==============================================================================
log_info "Installing Certbot..."
apt-get install -y certbot

# ==============================================================================
# Create systemd service for the application
# ==============================================================================
log_info "Creating systemd service..."
cat > /etc/systemd/system/oracle-engine.service << 'EOF'
[Unit]
Description=Oracle Engine Docker Compose Application
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
User=oracle
Group=oracle
WorkingDirectory=/var/www/oracle_engine
ExecStart=/usr/bin/docker compose -f docker-compose.prod.yml up -d --remove-orphans
ExecStop=/usr/bin/docker compose -f docker-compose.prod.yml down
ExecReload=/usr/bin/docker compose -f docker-compose.prod.yml up -d --remove-orphans
TimeoutStartSec=300

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable oracle-engine

# ==============================================================================
# Create log rotation
# ==============================================================================
log_info "Configuring log rotation..."
cat > /etc/logrotate.d/oracle-engine << 'EOF'
/var/www/oracle_engine/logs/*.log {
    daily
    missingok
    rotate 14
    compress
    delaycompress
    notifempty
    create 0640 oracle oracle
    sharedscripts
}
EOF

# ==============================================================================
# Setup completed
# ==============================================================================
log_info "=========================================="
log_info "Server setup completed successfully!"
log_info "=========================================="
log_info ""
log_info "Next steps:"
log_info "1. Copy your application files to: ${APP_DIR}"
log_info "2. Create .env file with your secrets"
log_info "3. Obtain SSL certificate: certbot certonly --standalone -d api.ifalabs.com"
log_info "4. Start the application: systemctl start oracle-engine"
log_info ""
log_info "Useful commands:"
log_info "  - View logs: docker compose -f docker-compose.prod.yml logs -f"
log_info "  - Restart: systemctl restart oracle-engine"
log_info "  - Status: systemctl status oracle-engine"
