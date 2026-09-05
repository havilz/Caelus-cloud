#!/usr/bin/env bash
# ==============================================================================
#  CAELUS CLOUD - Universal Multi-Cloud & Hybrid VPS Control Plane
#  Production Installer & Provisioning Wizard
# ==============================================================================
#  Usage:
#    curl -fsSL https://get.caelus.cloud/install.sh | bash
#    or:
#    bash scripts/install.sh
# ==============================================================================

set -eo pipefail

# ------------------------------------------------------------------------------
# Color Palettes & UI Formats
# ------------------------------------------------------------------------------
C_RESET="\033[0m"
C_BOLD="\033[1m"
C_CYAN="\033[36m"
C_GREEN="\033[32m"
C_YELLOW="\033[33m"
C_RED="\033[31m"
C_BLUE="\033[34m"
C_MAGENTA="\033[35m"
C_DIM="\033[2m"

INSTALL_DIR="${CAELUS_DIR:-$HOME/caelus-cloud}"
REPO_URL="https://github.com/havilz/Caelus-cloud.git"
COMPOSE_URL="https://raw.githubusercontent.com/havilz/Caelus-cloud/master/deploy/docker-compose.prod.yml"

print_banner() {
  clear 2>/dev/null || true
  echo -e "${C_CYAN}${C_BOLD}"
  cat << "EOF"
  ██████╗ █████╗ ███████╗██╗     ██╗   ██╗███████╗
 ██╔════╝██╔══██╗██╔════╝██║     ██║   ██║██╔════╝
 ██║     ███████║█████╗  ██║     ██║   ██║███████╗
 ██║     ██╔══██║██╔══╝  ██║     ██║   ██║╚════██║
 ╚██████╗██║  ██║███████╗███████╗╚██████╔╝███████║
  ╚═════╝╚═╝  ╚═╝╚══════╝╚══════╝ ╚═════╝ ╚══════╝
      Universal Multi-Cloud & Edge Platform
EOF
  echo -e "${C_RESET}"
  echo -e "${C_DIM}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${C_RESET}"
  echo -e "  ${C_BOLD}Caelus Cloud Control Plane — CLI Provisioner${C_RESET}"
  echo -e "  ${C_DIM}Version: v1.0.0-production-ready | Automated Setup Wizard${C_RESET}"
  echo -e "${C_DIM}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${C_RESET}"
  echo ""
}

log_info() {
  echo -e "  ${C_CYAN}[INFO]${C_RESET} $1"
}

log_success() {
  echo -e "  ${C_GREEN}[SUCCESS]${C_RESET} $1"
}

log_warn() {
  echo -e "  ${C_YELLOW}[WARN]${C_RESET} $1"
}

log_error() {
  echo -e "  ${C_RED}[ERROR]${C_RESET} $1"
}

# ------------------------------------------------------------------------------
# 1. Environment & Preflight Checks
# ------------------------------------------------------------------------------
check_system_requirements() {
  log_info "Checking operating system architecture and privileges..."

  # Detect OS
  OS="$(uname -s)"
  ARCH="$(uname -m)"
  case "$OS" in
    Linux)  OS_TYPE="linux" ;;
    Darwin) OS_TYPE="darwin" ;;
    *)      log_error "Operating System $OS is not officially supported."; exit 1 ;;
  esac

  log_success "Detected system: ${C_BOLD}$OS ($ARCH)${C_RESET}"
}

ensure_docker_installed() {
  log_info "Checking availability of Docker Engine & Docker Compose..."

  if ! command -v docker &> /dev/null; then
    log_warn "Docker is not installed on this system."
    echo -ne "  ${C_YELLOW}Would you like this script to install Docker automatically? (Y/n): ${C_RESET}"
    read -r INSTALL_DOCKER_CHOICE
    INSTALL_DOCKER_CHOICE="${INSTALL_DOCKER_CHOICE:-Y}"

    if [[ "$INSTALL_DOCKER_CHOICE" =~ ^[Yy]$ ]]; then
      log_info "Downloading and installing official Docker Engine..."
      curl -fsSL https://get.docker.com -o get-docker.sh
      sh get-docker.sh
      rm -f get-docker.sh

      # Start and enable docker service
      if command -v systemctl &> /dev/null; then
        sudo systemctl enable --now docker || true
      fi
      log_success "Docker Engine installed successfully."
    else
      log_error "Docker is required to run Caelus Cloud. Installation aborted."
      exit 1
    fi
  else
    DOCKER_VER=$(docker --version | awk '{print $3}' | tr -d ',')
    log_success "Docker Engine active: ${C_BOLD}v$DOCKER_VER${C_RESET}"
  fi

  # Check compose
  if ! docker compose version &> /dev/null; then
    log_warn "Docker Compose plugin not found. Installing docker-compose-plugin..."
    if command -v apt-get &> /dev/null; then
      sudo apt-get update -qq && sudo apt-get install -y -qq docker-compose-plugin
    elif command -v yum &> /dev/null; then
      sudo yum install -y -q docker-compose-plugin
    fi
  fi
  log_success "Docker Compose plugin active."
}

# ------------------------------------------------------------------------------
# 2. Key Generator
# ------------------------------------------------------------------------------
generate_crypto_token() {
  if command -v openssl &> /dev/null; then
    openssl rand -hex 32
  else
    # Fallback to /dev/urandom
    head -c 32 /dev/urandom | od -A n -t x1 | tr -d ' \n'
  fi
}

# ------------------------------------------------------------------------------
# 3. Interactive Topology Selection Wizard
# ------------------------------------------------------------------------------
wizard_topology_selection() {
  echo ""
  echo -e "${C_BOLD}Select Caelus Cloud Topology & Installation Architecture:${C_RESET}"
  echo -e "${C_DIM}────────────────────────────────────────────────────────────────────${C_RESET}"
  echo -e "  ${C_CYAN}${C_BOLD}[1] All-in-One Full Stack (Recommended Quickstart)${C_RESET}"
  echo -e "      ${C_DIM}- Runs local Postgres 16, Redis 7, MinIO, Prometheus & Loki via Docker.${C_RESET}"
  echo -e "      ${C_DIM}- Best for standalone/fresh VPS without external database.${C_RESET}"
  echo ""
  echo -e "  ${C_CYAN}${C_BOLD}[2] External Managed Database & Cache (Enterprise / Cloud-Native)${C_RESET}"
  echo -e "      ${C_DIM}- Uses cloud PostgreSQL (Supabase / AWS RDS / Neon) & Redis (Upstash / Aiven).${C_RESET}"
  echo -e "      ${C_DIM}- Saves RAM & Disk by 50%+; local DB containers are disabled.${C_RESET}"
  echo ""
  echo -e "  ${C_CYAN}${C_BOLD}[3] Local Workstation + Cloudflare Tunnel (Hybrid Remote VPS Agent)${C_RESET}"
  echo -e "      ${C_DIM}- Caelus Control Plane on your local workstation.${C_RESET}"
  echo -e "      ${C_DIM}- Connects Cloudflare Tunnel so remote VPS nodes can report telemetry.${C_RESET}"
  echo -e "${C_DIM}────────────────────────────────────────────────────────────────────${C_RESET}"
  echo -ne "  ${C_BOLD}Enter choice [1/2/3] (Default: 1): ${C_RESET}"
  read -r TOPOLOGY_CHOICE < /dev/tty || TOPOLOGY_CHOICE="1"
  TOPOLOGY_CHOICE="${TOPOLOGY_CHOICE:-1}"

  # Default Variable Initialization
  DB_HOST="postgres"
  DB_PORT="5432"
  DB_USER="caelus_user"
  DB_PASSWORD="$(generate_crypto_token | cut -c1-16)"
  DB_NAME="caelus_cloud_db"
  DB_SSL_MODE="disable"

  REDIS_HOST="redis"
  REDIS_PORT="6379"
  REDIS_PASSWORD=""
  REDIS_DB="0"
  REDIS_USE_TLS="false"

  PUBLIC_DOMAIN="localhost:3000"
  PUBLIC_API_URL="http://localhost:8080"
  TUNNEL_ENABLED="false"
  TUNNEL_TOKEN=""

  case "$TOPOLOGY_CHOICE" in
    1)
      log_info "All-in-One Full Stack mode selected. Database & Redis automatically provisioned via Docker."
      ;;

    2)
      echo ""
      echo -e "  ${C_YELLOW}${C_BOLD}-- External Managed PostgreSQL Configuration --${C_RESET}"
      echo -ne "  PostgreSQL Host (e.g. db.xyz.supabase.co / rds.amazonaws.com): "
      read -r DB_HOST < /dev/tty || true
      echo -ne "  PostgreSQL Port [5432]: "
      read -r INPUT_DB_PORT < /dev/tty || true
      DB_PORT="${INPUT_DB_PORT:-5432}"
      echo -ne "  PostgreSQL User [postgres]: "
      read -r INPUT_DB_USER < /dev/tty || true
      DB_USER="${INPUT_DB_USER:-postgres}"
      echo -ne "  PostgreSQL Password: "
      read -s DB_PASSWORD < /dev/tty || true
      echo ""
      echo -ne "  Database Name [postgres]: "
      read -r INPUT_DB_NAME < /dev/tty || true
      DB_NAME="${INPUT_DB_NAME:-postgres}"
      echo -ne "  SSL Mode (require/verify-full/disable) [require]: "
      read -r INPUT_SSL_MODE < /dev/tty || true
      DB_SSL_MODE="${INPUT_SSL_MODE:-require}"

      echo ""
      echo -e "  ${C_YELLOW}${C_BOLD}-- External Managed Redis Configuration --${C_RESET}"
      echo -ne "  Redis Host (e.g. global-redis.upstash.io / aivencloud.com): "
      read -r REDIS_HOST < /dev/tty || true
      echo -ne "  Redis Port [6379]: "
      read -r INPUT_REDIS_PORT < /dev/tty || true
      REDIS_PORT="${INPUT_REDIS_PORT:-6379}"
      echo -ne "  Redis Password: "
      read -s REDIS_PASSWORD < /dev/tty || true
      echo ""
      echo -ne "  Use Redis TLS (true/false) [true]: "
      read -r INPUT_REDIS_TLS < /dev/tty || true
      REDIS_USE_TLS="${INPUT_REDIS_TLS:-true}"
      ;;

    3)
      log_info "Hybrid Workstation + Cloudflare Tunnel mode selected."
      TUNNEL_ENABLED="true"
      echo ""
      echo -ne "  Enter Public Domain for Caelus Dashboard (e.g. caelus.yourdomain.com): "
      read -r PUBLIC_DOMAIN < /dev/tty || true
      PUBLIC_DOMAIN="${PUBLIC_DOMAIN:-localhost:3000}"
      PUBLIC_API_URL="https://${PUBLIC_DOMAIN}/api"

      echo -ne "  Enter Cloudflare Tunnel Token (Optional, press Enter to skip if using Quick Tunnel): "
      read -r TUNNEL_TOKEN < /dev/tty || true
      ;;

    *)
      log_warn "Invalid selection, falling back to All-in-One Full Stack."
      ;;
  esac
}

# ------------------------------------------------------------------------------
# 4. Generate Production .env File
# ------------------------------------------------------------------------------
generate_env_configuration() {
  log_info "Generating cryptographic tokens (AES-256-GCM & JWT Tokens)..."

  JWT_SECRET="$(generate_crypto_token)"
  ENCRYPTION_KEY="$(generate_crypto_token)"
  WEBHOOK_SECRET="$(generate_crypto_token)"

  mkdir -p "$INSTALL_DIR"
  cat << EOF > "$INSTALL_DIR/.env"
# ==============================================================================
# CAELUS CLOUD - PRODUCTION CONFIGURATION (AUTO-GENERATED BY INSTALLER)
# Generated At: $(date -u +"%Y-%m-%dT%H:%M:%SZ")
# ==============================================================================

# ------------------------------------------------------------------------------
# 1. APPLICATION CORE
# ------------------------------------------------------------------------------
APP_ENV=production
APP_PORT=8080
LOG_LEVEL=info
CORS_ALLOWED_ORIGINS=http://${PUBLIC_DOMAIN},https://${PUBLIC_DOMAIN},http://localhost:3000

# ------------------------------------------------------------------------------
# 2. DATABASE CONFIGURATION (PostgreSQL)
# ------------------------------------------------------------------------------
DB_HOST=${DB_HOST}
DB_PORT=${DB_PORT}
DB_USER=${DB_USER}
DB_PASSWORD=${DB_PASSWORD}
DB_NAME=${DB_NAME}
DB_SSL_MODE=${DB_SSL_MODE}
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=5m

# ------------------------------------------------------------------------------
# 3. CACHE & MESSAGE BROKER (Redis)
# ------------------------------------------------------------------------------
REDIS_HOST=${REDIS_HOST}
REDIS_PORT=${REDIS_PORT}
REDIS_PASSWORD=${REDIS_PASSWORD}
REDIS_DB=${REDIS_DB}
REDIS_USE_TLS=${REDIS_USE_TLS}

# ------------------------------------------------------------------------------
# 4. SECURITY & CRYPTOGRAPHY (AES-256 & JWT)
# ------------------------------------------------------------------------------
JWT_SECRET=${JWT_SECRET}
JWT_ACCESS_EXPIRATION=15m
JWT_REFRESH_EXPIRATION=7d
ENCRYPTION_KEY=${ENCRYPTION_KEY}
WEBHOOK_SIGNING_SECRET=${WEBHOOK_SECRET}

# ------------------------------------------------------------------------------
# 5. STORAGE & OBSERVABILITY
# ------------------------------------------------------------------------------
STORAGE_DRIVER=minio
STORAGE_ENDPOINT=http://minio:9000
STORAGE_ACCESS_KEY=minioadmin
STORAGE_SECRET_KEY=minioadmin
STORAGE_BUCKET=caelus-storage
STORAGE_REGION=us-east-1

PROMETHEUS_URL=http://prometheus:9090
LOKI_URL=http://loki:3100
EOF

  log_success "Configuration file created at: ${C_BOLD}$INSTALL_DIR/.env${C_RESET}"
}

# ------------------------------------------------------------------------------
# 5. Launch Services
# ------------------------------------------------------------------------------
launch_caelus_platform() {
  log_info "Setting up and launching Caelus Cloud platform services..."
  cd "$INSTALL_DIR"

  # Copy or create docker-compose.yml if running via standalone curl
  if [ ! -f "$INSTALL_DIR/docker-compose.yml" ]; then
    if [ -f "$(dirname "$0")/../docker-compose.yml" ]; then
      cp "$(dirname "$0")/../docker-compose.yml" "$INSTALL_DIR/docker-compose.yml"
    else
      log_info "Downloading official docker-compose.yml from repository..."
      curl -fsSL "$COMPOSE_URL" -o "$INSTALL_DIR/docker-compose.yml"
    fi
  fi

  # Start services based on topology
  if [ "$TOPOLOGY_CHOICE" == "2" ]; then
    log_info "Starting Caelus Cloud (External Managed Database mode)..."
    docker compose up -d --no-deps api worker frontend prometheus loki
  else
    log_info "Starting full Caelus Cloud stack (All-in-One)..."
    docker compose up -d
  fi

  log_success "All Caelus Cloud service containers started successfully!"
}

# ------------------------------------------------------------------------------
# 6. Final Summary & Next Steps
# ------------------------------------------------------------------------------
print_installation_summary() {
  # Detect Public IP
  PUBLIC_IP="$(curl -s4 https://ifconfig.me 2>/dev/null || echo "127.0.0.1")"

  echo ""
  echo -e "${C_GREEN}${C_BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${C_RESET}"
  echo -e "  ${C_GREEN}${C_BOLD}CAELUS CLOUD INSTALLATION COMPLETED SUCCESSFULLY!${C_RESET}"
  echo -e "${C_GREEN}${C_BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${C_RESET}"
  echo ""
  echo -e "  ${C_BOLD}Dashboard Web URL:${C_RESET}"
  echo -e "  - ${C_CYAN}${C_BOLD}http://${PUBLIC_IP}:3000${C_RESET} ${C_DIM}(or http://${PUBLIC_DOMAIN})${C_RESET}"
  echo ""
  echo -e "  ${C_BOLD}Backend API Endpoint:${C_RESET}"
  echo -e "  - ${C_CYAN}http://${PUBLIC_IP}:8080${C_RESET}"
  echo ""
  echo -e "  ${C_BOLD}Installation Directory:${C_RESET}"
  echo -e "  - ${C_DIM}$INSTALL_DIR${C_RESET}"
  echo ""
  echo -e "  ${C_YELLOW}${C_BOLD}-- Connecting Host / Client VPS Nodes: --${C_RESET}"
  echo -e "  Run this command on any target server to connect it:"
  echo -e "  ${C_CYAN}curl -fsSL http://${PUBLIC_IP}:8080/install.sh | bash -s -- --endpoint=http://${PUBLIC_IP}:8080${C_RESET}"
  echo ""
  echo -e "${C_DIM}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${C_RESET}"
  echo -e "  ${C_DIM}Documentation & Guide: https://github.com/havilz/Caelus-cloud${C_RESET}"
  echo ""
}

# ------------------------------------------------------------------------------
# Main Entrypoint
# ------------------------------------------------------------------------------
main() {
  print_banner
  check_system_requirements
  ensure_docker_installed
  wizard_topology_selection
  generate_env_configuration
  launch_caelus_platform
  print_installation_summary
}

main "$@"
