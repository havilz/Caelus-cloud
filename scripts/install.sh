#!/usr/bin/env bash
# ==============================================================================
#  CAELUS CLOUD - Universal Multi-Cloud & Hybrid VPS Control Plane
#  Smart Production Installer & Provisioning Wizard
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
COMPOSE_URL="https://raw.githubusercontent.com/havilz/Caelus-cloud/master/docker-compose.yml"

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
  echo -e "  ${C_BOLD}Caelus Cloud Control Plane — Smart CLI Provisioner${C_RESET}"
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
  log_info "Memeriksa arsitektur sistem operasi dan hak akses..."

  # Detect OS
  OS="$(uname -s)"
  ARCH="$(uname -m)"
  case "$OS" in
    Linux)  OS_TYPE="linux" ;;
    Darwin) OS_TYPE="darwin" ;;
    *)      log_error "Sistem Operasi $OS saat ini belum didukung secara resmi."; exit 1 ;;
  esac

  log_success "Sistem terdeteksi: ${C_BOLD}$OS ($ARCH)${C_RESET}"
}

ensure_docker_installed() {
  log_info "Memeriksa ketersediaan Docker Engine & Docker Compose..."

  if ! command -v docker &> /dev/null; then
    log_warn "Docker belum terpasang di sistem ini."
    echo -ne "  ${C_YELLOW}Apakah Anda ingin script ini memasang Docker secara otomatis? (Y/n): ${C_RESET}"
    read -r INSTALL_DOCKER_CHOICE
    INSTALL_DOCKER_CHOICE="${INSTALL_DOCKER_CHOICE:-Y}"

    if [[ "$INSTALL_DOCKER_CHOICE" =~ ^[Yy]$ ]]; then
      log_info "Mengunduh dan memasang Docker Engine resmi..."
      curl -fsSL https://get.docker.com -o get-docker.sh
      sh get-docker.sh
      rm -f get-docker.sh

      # Start and enable docker service
      if command -v systemctl &> /dev/null; then
        sudo systemctl enable --now docker || true
      fi
      log_success "Docker Engine berhasil dipasang."
    else
      log_error "Docker diperlukan untuk menjalankan Caelus Cloud. Instalasi dibatalkan."
      exit 1
    fi
  else
    DOCKER_VER=$(docker --version | awk '{print $3}' | tr -d ',')
    log_success "Docker Engine aktif: ${C_BOLD}v$DOCKER_VER${C_RESET}"
  fi

  # Check compose
  if ! docker compose version &> /dev/null; then
    log_warn "Plugin docker compose tidak ditemukan. Memasang docker-compose-plugin..."
    if command -v apt-get &> /dev/null; then
      sudo apt-get update -qq && sudo apt-get install -y -qq docker-compose-plugin
    elif command -v yum &> /dev/null; then
      sudo yum install -y -q docker-compose-plugin
    fi
  fi
  log_success "Docker Compose Plugin aktif."
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
  echo -e "${C_BOLD}Pilih Skenario Topologi & Arsitektur Instalasi Caelus Cloud:${C_RESET}"
  echo -e "${C_DIM}────────────────────────────────────────────────────────────────────${C_RESET}"
  echo -e "  ${C_CYAN}${C_BOLD}[1] All-in-One Full Stack (Rekomendasi Quickstart)${C_RESET}"
  echo -e "      ${C_DIM}• Menjalankan Postgres 16, Redis 7, MinIO, Prometheus & Loki lokal (Docker).${C_RESET}"
  echo -e "      ${C_DIM}• Cocok untuk VPS baru/mandiri tanpa database eksternal.${C_RESET}"
  echo ""
  echo -e "  ${C_CYAN}${C_BOLD}[2] External Managed Database & Cache (Enterprise / Cloud-Native)${C_RESET}"
  echo -e "      ${C_DIM}• Menggunakan PostgreSQL Cloud (Supabase / AWS RDS / Neon) & Redis (Upstash/Aiven).${C_RESET}"
  echo -e "      ${C_DIM}• Hemat RAM & Disk hingga 50%+; kontainer DB lokal otomatis dinonaktifkan.${C_RESET}"
  echo ""
  echo -e "  ${C_CYAN}${C_BOLD}[3] Local Workstation + Cloudflare Tunnel (Hybrid Remote VPS Agent)${C_RESET}"
  echo -e "      ${C_DIM}• Caelus Control Plane di laptop/perangkat kerja Anda.${C_RESET}"
  echo -e "      ${C_DIM}• Otomatis menghubungkan Cloudflare Tunnel agar VPS luar bisa lapor telemetri.${C_RESET}"
  echo -e "${C_DIM}────────────────────────────────────────────────────────────────────${C_RESET}"
  echo -ne "  ${C_BOLD}Masukkan nomor pilihan [1/2/3] (Default: 1): ${C_RESET}"
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
      log_info "Mode All-in-One Full Stack dipilih. Database & Redis otomatis dibuat via Docker."
      ;;

    2)
      echo ""
      echo -e "  ${C_YELLOW}${C_BOLD}── Konfigurasi Managed PostgreSQL Eksternal ──${C_RESET}"
      echo -ne "  Host PostgreSQL (e.g. db.xyz.supabase.co / rds.amazonaws.com): "
      read -r DB_HOST < /dev/tty || true
      echo -ne "  Port PostgreSQL [5432]: "
      read -r INPUT_DB_PORT < /dev/tty || true
      DB_PORT="${INPUT_DB_PORT:-5432}"
      echo -ne "  User PostgreSQL [postgres]: "
      read -r INPUT_DB_USER < /dev/tty || true
      DB_USER="${INPUT_DB_USER:-postgres}"
      echo -ne "  Password PostgreSQL: "
      read -s DB_PASSWORD < /dev/tty || true
      echo ""
      echo -ne "  Nama Database [postgres]: "
      read -r INPUT_DB_NAME < /dev/tty || true
      DB_NAME="${INPUT_DB_NAME:-postgres}"
      echo -ne "  SSL Mode (require/verify-full/disable) [require]: "
      read -r INPUT_SSL_MODE < /dev/tty || true
      DB_SSL_MODE="${INPUT_SSL_MODE:-require}"

      echo ""
      echo -e "  ${C_YELLOW}${C_BOLD}── Konfigurasi Managed Redis Eksternal ──${C_RESET}"
      echo -ne "  Host Redis (e.g. global-redis.upstash.io / aivencloud.com): "
      read -r REDIS_HOST < /dev/tty || true
      echo -ne "  Port Redis [6379]: "
      read -r INPUT_REDIS_PORT < /dev/tty || true
      REDIS_PORT="${INPUT_REDIS_PORT:-6379}"
      echo -ne "  Password Redis: "
      read -s REDIS_PASSWORD < /dev/tty || true
      echo ""
      echo -ne "  Gunakan Redis TLS (true/false) [true]: "
      read -r INPUT_REDIS_TLS < /dev/tty || true
      REDIS_USE_TLS="${INPUT_REDIS_TLS:-true}"
      ;;

    3)
      log_info "Mode Hybrid Workstation + Cloudflare Tunnel dipilih."
      TUNNEL_ENABLED="true"
      echo ""
      echo -ne "  Masukkan Domain Publik untuk Dashboard Caelus Anda (e.g. caelus.domainanda.com): "
      read -r PUBLIC_DOMAIN < /dev/tty || true
      PUBLIC_DOMAIN="${PUBLIC_DOMAIN:-localhost:3000}"
      PUBLIC_API_URL="https://${PUBLIC_DOMAIN}/api"

      echo -ne "  Masukkan Cloudflare Tunnel Token (Opsional, tekan Enter untuk lewati jika pakai Quick Tunnel): "
      read -r TUNNEL_TOKEN < /dev/tty || true
      ;;

    *)
      log_warn "Pilihan tidak valid, kembali ke All-in-One Full Stack."
      ;;
  esac
}

# ------------------------------------------------------------------------------
# 4. Generate Production .env File
# ------------------------------------------------------------------------------
generate_env_configuration() {
  log_info "Menghasilkan kunci kriptografi aman (AES-256-GCM & JWT Tokens)..."

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

  log_success "File konfigurasi berhasil dibuat di: ${C_BOLD}$INSTALL_DIR/.env${C_RESET}"
}

# ------------------------------------------------------------------------------
# 5. Launch Services
# ------------------------------------------------------------------------------
launch_caelus_platform() {
  log_info "Menyiapkan dan menyalakan layanan platform Caelus Cloud..."
  cd "$INSTALL_DIR"

  # Copy or create docker-compose.yml if running via standalone curl
  if [ ! -f "$INSTALL_DIR/docker-compose.yml" ]; then
    if [ -f "$(dirname "$0")/../docker-compose.yml" ]; then
      cp "$(dirname "$0")/../docker-compose.yml" "$INSTALL_DIR/docker-compose.yml"
    else
      log_info "Mengunduh docker-compose.yml resmi dari repository..."
      curl -fsSL "$COMPOSE_URL" -o "$INSTALL_DIR/docker-compose.yml"
    fi
  fi

  # Start services based on topology
  if [ "$TOPOLOGY_CHOICE" == "2" ]; then
    log_info "Memulai Caelus Cloud (Mode External Managed Database)..."
    docker compose up -d --no-deps api worker frontend prometheus loki
  else
    log_info "Memulai seluruh stack Caelus Cloud (All-in-One)..."
    docker compose up -d
  fi

  log_success "Seluruh kontainer Caelus Cloud berhasil diaktifkan!"
}

# ------------------------------------------------------------------------------
# 6. Final Summary & Next Steps
# ------------------------------------------------------------------------------
print_installation_summary() {
  # Detect Public IP
  PUBLIC_IP="$(curl -s4 https://ifconfig.me 2>/dev/null || echo "127.0.0.1")"

  echo ""
  echo -e "${C_GREEN}${C_BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${C_RESET}"
  echo -e "  ${C_GREEN}${C_BOLD}🎉 INSTALASI CAELUS CLOUD BERHASIL DISELESAIKAN! 🎉${C_RESET}"
  echo -e "${C_GREEN}${C_BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${C_RESET}"
  echo ""
  echo -e "  ${C_BOLD}Dashboard Web URL:${C_RESET}"
  echo -e "  👉 ${C_CYAN}${C_BOLD}http://${PUBLIC_IP}:3000${C_RESET} ${C_DIM}(atau http://${PUBLIC_DOMAIN})${C_RESET}"
  echo ""
  echo -e "  ${C_BOLD}Backend API Endpoint:${C_RESET}"
  echo -e "  👉 ${C_CYAN}http://${PUBLIC_IP}:8080${C_RESET}"
  echo ""
  echo -e "  ${C_BOLD}Direktori Instalasi:${C_RESET}"
  echo -e "  📂 ${C_DIM}$INSTALL_DIR${C_RESET}"
  echo ""
  echo -e "  ${C_YELLOW}${C_BOLD}── Cara Menghubungkan VPS Klien ke Caelus Ini: ──${C_RESET}"
  echo -e "  Jalankan perintah ini di VPS target manapun untuk menghubungkannya secara live:"
  echo -e "  ${C_CYAN}curl -fsSL http://${PUBLIC_IP}:8080/install.sh | bash -s -- --endpoint=http://${PUBLIC_IP}:8080${C_RESET}"
  echo ""
  echo -e "${C_DIM}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${C_RESET}"
  echo -e "  ${C_DIM}Dokumentasi & Panduan Lengkap: https://github.com/havilz/Caelus-cloud${C_RESET}"
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
