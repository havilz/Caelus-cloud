# Panduan Setup Lengkap & Cara Penggunaan Fitur Caelus Cloud

Dokumen ini memuat panduan komprehensif mulai dari instalasi prasyarat, konfigurasi lingkungan, inisialisasi basis data, menjalankan seluruh layanan, hingga panduan langkah demi langkah cara menggunakan setiap fitur pada platform Caelus Cloud.

---

## Daftar Isi

1. [Prasyarat Sistem & Perangkat Lunak](#1-prasyarat-sistem--perangkat-lunak)
2. [Konfigurasi Lingkungan (Environment Variables)](#2-konfigurasi-lingkungan-environment-variables)
3. [Inisialisasi & Migrasi Basis Data](#3-inisialisasi--migrasi-basis-data)
4. [Menjalankan Seluruh Layanan (Development & Production)](#4-menjalankan-seluruh-layanan-development--production)
5. [Panduan Langkah demi Langkah Penggunaan Setiap Fitur](#5-panduan-langkah-demi-langkah-penggunaan-setiap-fitur)
   - [5.1. Autentikasi & Ruang Kerja Organisasi (Workspace)](#51-autentikasi--ruang-kerja-organisasi-workspace)
   - [5.2. Manajemen Cloud Provider & Kredensial Terenkripsi](#52-manajemen-cloud-provider--kredensial-terenkripsi)
   - [5.3. Registrasi Server & Onboarding Agent (BYOS)](#53-registrasi-server--onboarding-agent-byos)
   - [5.4. Declarative Infrastructure as Code (IaC) & Rollback Engine](#54-declarative-infrastructure-as-code-iac--rollback-engine)
   - [5.5. Container Orchestration & Web-Based Streaming Terminal](#55-container-orchestration--web-based-streaming-terminal)
   - [5.6. Object Storage Explorer & Backup Otomatis](#56-object-storage-explorer--backup-otomatis)
   - [5.7. Monitoring Performa Host & Telemetri Real-Time](#57-monitoring-performa-host--telemetri-real-time)
   - [5.8. Sentinel Security Hub & Remediasi Kerentanan](#58-sentinel-security-hub--remediasi-kerentanan)
   - [5.9. Mesin Otomasi & Notifikasi Multi-Channel](#59-mesin-otomasi--notifikasi-multi-channel)
6. [Troubleshooting & Pemecahan Masalah](#6-troubleshooting--pemecahan-masalah)

---

## 1. Prasyarat Sistem & Perangkat Lunak

Sebelum memulai, pastikan perangkat Anda telah terpasang perangkat lunak berikut:

| Perangkat Lunak | Versi Minimal | Keterangan |
| :--- | :--- | :--- |
| **Go** | 1.22 atau lebih baru | Diperlukan untuk Backend API, Worker, dan Agent ([go.dev](https://go.dev)) |
| **Node.js** | 20.x atau lebih baru | Runtime JavaScript untuk frontend Next.js ([nodejs.org](https://nodejs.org)) |
| **pnpm** | 9.x atau lebih baru | Package manager frontend (`npm install -g pnpm`) |
| **Docker & Compose** | Docker v24+, Compose v2+ | Untuk infrastruktur lokal (Postgres, Redis, MinIO) |
| **GNU Make** | 4.x atau lebih baru | Untuk menjalankan tugas otomatisasi monorepo |

---

## 2. Konfigurasi Lingkungan (Environment Variables)

### 2.1. Konfigurasi Backend & Root Monorepo (`.env`)
Salin file `.env.example` menjadi `.env` pada root direktori:

```bash
cp .env.example .env
```

Sesuaikan parameter kredensial pada berkas `.env`:

```env
# 1. Server Application
APP_ENV=development
APP_NAME=caelus-cloud-api
APP_HOST=0.0.0.0
APP_PORT=8080
APP_DEBUG=true
APP_LOG_LEVEL=info
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://127.0.0.1:3000

# 2. Database PostgreSQL (Supabase / Local Postgres)
DB_HOST=aws-0-ap-southeast-1.pooler.supabase.com
DB_PORT=5432
DB_USER=postgres.your_project_ref
DB_PASSWORD=your_database_password
DB_NAME=postgres
DB_SSLMODE=require
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=5m

# 3. Redis Cache & Queue (Upstash / Local Redis)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# 4. Security & Cryptography (Wajib 32-byte string untuk ENCRYPTION_KEY)
JWT_SECRET=super_secret_jwt_key_at_least_32_characters_long
JWT_ACCESS_EXPIRATION=15m
JWT_REFRESH_EXPIRATION=7d
ENCRYPTION_KEY=2ad69b09d78f42228c74e54bd39b8c2e17f136f34bb94b67a3e3612549781d94

# 5. Object Storage (S3 / MinIO / Cloudflare R2)
STORAGE_DRIVER=s3
STORAGE_ENDPOINT=http://localhost:9000
STORAGE_ACCESS_KEY=minioadmin
STORAGE_SECRET_KEY=minioadmin
STORAGE_BUCKET=caelus-storage
STORAGE_REGION=us-east-1
STORAGE_USE_SSL=false
```

### 2.2. Konfigurasi Frontend (`frontend/.env.local`)
Buat file `frontend/.env.local` untuk mengarahkan Next.js ke backend API:

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_WS_URL=ws://localhost:8080/api/v1/ws
```

---

## 3. Inisialisasi & Migrasi Basis Data

### 3.1. Menjalankan Kontainer Infrastruktur Lokal (Opsional jika menggunakan Local Docker)
Jika Anda menggunakan PostgreSQL, Redis, dan MinIO lokal:
```bash
make infra-up
```

### 3.2. Menjalankan Migrasi Skema Basis Data
Eksekusi migrasi DDL SQL untuk membuat seluruh tabel, indeks performa, trigger, dan kebijakan Row Level Security (RLS):
```bash
make migrate-up
```

Perintah di atas akan mengeksekusi migrasi `000001` hingga `000008` secara sekuensial.

---

## 4. Menjalankan Seluruh Layanan (Development & Production)

Semua layanan dapat dikontrol melalui target `Makefile` terpusat dari root monorepo:

### 4.1. Instalasi Dependensi
```bash
make deps
```
Perintah ini akan otomatis mengunduh Go modules backend, Go modules agent, dan pnpm packages frontend.

### 4.2. Menjalankan Layanan Secara Terpisah

1. **Terminal 1 - Backend API Server (Port 8080)**:
   ```bash
   make api
   ```
2. **Terminal 2 - Background Worker & Task Scheduler**:
   ```bash
   make worker
   ```
3. **Terminal 3 - Frontend Web Panel (Port 3000)**:
   ```bash
   make frontend
   ```
4. **Terminal 4 - Host Agent Daemon (Opsional di mesin lokal)**:
   ```bash
   make agent
   ```

Akses antarmuka dashboard pada peramban: **`http://localhost:3000`**

### 4.3. Menjalankan Suite Pengujian & Validasi Kualitas
```bash
# Uji coba backend test suite
make test-backend

# Uji coba agent test suite
make test-agent

# Audit linting frontend TypeScript/ESLint
make lint

# Kompilasi seluruh binary produksi
make build
```

---

## 5. Panduan Langkah demi Langkah Penggunaan Setiap Fitur

### 5.1. Autentikasi & Ruang Kerja Organisasi (Workspace)

1. Buka halaman registrasi di `http://localhost:3000/register`.
2. Masukkan **Nama Lengkap**, **Email**, **Password**, dan **Nama Organisasi / Workspace**.
3. Sistem akan membuat akun, mengonfigurasi organisasi default dengan peran `owner`, dan menerbitkan token JWT.
4. Pada sesi berikutnya, gunakan `http://localhost:3000/login` untuk masuk ke panel.
5. Anda dapat mengelola profil dan melihat informasi tenant pada menu `Settings` (`/settings`).

---

### 5.2. Manajemen Cloud Provider & Kredensial Terenkripsi

Fitur ini memungkinkan Anda menghubungkan akun cloud publik untuk provisioning dan monitoring terpadu.

1. Navigasikan ke menu **Infrastructure -> Cloud Providers** (`/infrastructure/providers`).
2. Klik tombol **Connect Provider**.
3. Pilih penyedia cloud yang didukung:
   - **Amazon Web Services (AWS)**: Masukkan `Access Key ID`, `Secret Access Key`, dan default `Region` (misal: `us-east-1` atau `ap-southeast-1`).
   - **DigitalOcean**: Masukkan `Personal Access Token`.
   - **Hetzner Cloud**: Masukkan `API Token`.
   - **Contabo**: Masukkan `Client ID`, `Client Secret`, `API User`, dan `API Password`.
   - **Custom Host / BYOS**: Untuk server mandiri non-cloud publik.
4. Beri nama alias untuk kredensial tersebut (contoh: `Production AWS Account`).
5. Klik **Save Credential**. Kredensial akan langsung dienkripsi menggunakan **AES-256-GCM** sebelum disimpan ke database.
6. Klik tombol **Test Connection** pada baris kredensial untuk memverifikasi keabsahan API Key secara langsung ke server cloud provider.

---

### 5.3. Registrasi Server & Onboarding Agent (BYOS)

Ada dua metode menambahkan server ke Caelus Cloud:

#### Metode A: Bring Your Own Server (BYOS) via 1-Line Script (Direkomendasikan)
1. Buka menu **Infrastructure -> Servers** (`/infrastructure/vps`).
2. Klik tombol **Add Server**.
3. Pilih tab **Bring Your Own Server (BYOS)** dan masukkan nama server (misal: `api-gateway-prod`).
4. Klik **Create & Get Installer Script**.
5. Sistem akan menampilkan perintah instalasi shell 1 baris yang telah disematkan `SERVER_ID` dan `AGENT_SECRET`:
   ```bash
   curl -sSL http://<YOUR_CAELUS_IP>:8080/install.sh | sudo SERVER_ID="..." AGENT_SECRET="..." bash
   ```
6. Buka terminal SSH VPS Linux Anda, tempel perintah tersebut, dan jalankan dengan hak akses root/sudo.
7. Skrip akan otomatis mengunduh binary `caelus-agent`, mengonfigurasi daemon systemd, dan menjalankan service.
8. Dalam beberapa detik, server akan otomatis terdeteksi `running` dan spesifikasi hardware (CPU, RAM, Disk, OS, Hostname) akan tersinkronisasi otomatis ke dashboard.

#### Metode B: Cloud Provider Provisioning
1. Pilih tab **Cloud Provider** pada modal pembuatan server.
2. Pilih kredensial cloud provider yang sudah terhubung (AWS/DO/Hetzner/Contabo).
3. Pilih region, OS image, dan ukuran instance.
4. Klik **Provision Server**. Backend akan memanggil API cloud provider untuk membuat VM baru.

---

### 5.4. Declarative Infrastructure as Code (IaC) & Rollback Engine

Fitur ini memungkinkan pendefinisian seluruh server, storage, container, dan firewall rules dalam satu manifest YAML deklaratif.

1. Buka menu **Infrastructure -> Declarative IaC** (`/infrastructure/iac`).
2. Tulis atau pilih **Starter Template** (misal: `Fullstack Application`, `Multi-Cloud HA`, `Microservices Stack`).
3. Contoh format manifest YAML:
   ```yaml
   version: "1.0"
   resources:
     servers:
       - name: web-production
         provider: digitalocean
         region: sgp1
         instance_type: s-2vcpu-4gb

     storages:
       - name: static-media
         provider: r2
         region: auto

     containers:
       - name: redis-cache
         image: redis:alpine
         ports:
           - "6379:6379"
   ```
4. Editor memiliki fitur validasi sintaks realtime. Jika ada kesalahan baris/kolom, indikator error akan muncul secara otomatis.
5. Klik **Validate & Generate Plan**.
6. **Visual Diff Viewer** akan membandingkan *Desired State* di YAML vs *Actual State* saat ini:
   - `+ Green Badge`: Sumber daya yang akan dibuat (*Create*).
   - `~ Blue Badge`: Sumber daya yang mengalami perubahan spesifikasi (*Update*).
   - `- Red Badge`: Sumber daya yang dihapus dari YAML (*Delete*).
   - `= Gray Badge`: Sumber daya yang sudah sesuai (*No-op*).
7. Jika rencana sudah sesuai, klik tombol **Apply Infrastructure Changes**.
8. **Mekanisme Rollback**:
   - Jika eksekusi gagal di tengah jalan, sistem melakukan *LIFO rollback* otomatis untuk menjaga konsistensi state.
   - Anda juga dapat membuka panel **State History** di sisi kanan dan mengklik **Rollback** ke versi snapshot sebelumnya dengan 1 klik.

---

### 5.5. Container Orchestration & Web-Based Streaming Terminal

Fitur ini memungkinkan peluncuran dan pemantauan aplikasi berbasis Docker container secara langsung dari web tanpa perlu login SSH ke server.

1. Buka menu **Infrastructure -> Containers** (`/infrastructure/containers`).
2. Klik tombol **Deploy New Container**.
3. Pilih server target dan gunakan **Quick Image Presets** (misal: Nginx, Redis, PostgreSQL, Node.js) atau masukkan nama image Docker kustom.
4. Tentukan konfigurasi:
   - **Container Name**: Nama unik container (contoh: `production-web-app`).
   - **Port Bindings**: Port Host ke Port Container (contoh: `80:80` atau `3000:3000`).
   - **Volume Mounts**: Direktori host ke path container (contoh: `/var/data:/data`).
   - **Environment Variables**: Key-value konfigurasi dinamis aplikasi.
   - **Restart Policy**: `unless-stopped` atau `always`.
5. Klik **Launch Container**.
6. **Web-Based Streaming Terminal**:
   - Klik baris deployment untuk membuka konsol live terminal.
   - Log deployment (Pulling layers, Port binding, Healthcheck) akan mengalir secara real-time via **Server-Sent Events (SSE)** dengan ANSI color formatting.
   - Anda dapat mencari log teks, memfilter stream (`All`, `stdout`, `stderr`, `system`), mengaktifkan `Auto-scroll`, dan menyalin log ke clipboard.
7. Anda dapat menghentikan container (*Stop*) atau mengembalikan ke versi image sebelumnya (*Rollback Deployment*) dari tabel aksi.

---

### 5.6. Object Storage Explorer & Backup Otomatis

1. Buka menu **Storage** (`/storage`).
2. **Object Explorer**:
   - Buat bucket baru dengan memilih region dan provider storage (MinIO / S3 / R2).
   - Masuk ke dalam bucket untuk mengunggah file via drag-and-drop multipart upload.
   - Buat tautan unduhan aman berbatas waktu (*Pre-signed URLs*) dengan masa kedaluwarsa 15 menit hingga 24 jam.
3. **Backup Management** (`/storage/backups`):
   - Buat kebijakan backup berkala (*Backup Policy*) dengan interval harian/mingguan.
   - Tentukan retensi penyimpanan (misal: simpan 7 backup terakhir).
   - Jalankan backup manual instan dengan tombol **Run Backup Now**.

---

### 5.7. Monitoring Performa Host & Telemetri Real-Time

1. Buka menu **Monitoring** (`/monitoring`) atau klik detail spesifik server di `/infrastructure/vps/[id]`.
2. Dashboard menampilkan metrik performa interaktif real-time:
   - Utilisasi CPU (Grafik deret waktu per detik).
   - Utilisasi Memori RAM & Swap.
   - Kapasitas & Utilisasi Disk Storage.
   - Network I/O (Throughput Rx / Tx KB/s).
   - Daftar container Docker aktif beserta konsumsi memori dan CPU masing-masing.
3. Data diperbarui secara instan via WebSocket Hub tanpa perlu melakukan reload halaman.
4. **Alert Thresholds**: Konfigurasi batas peringatan (misal: CPU > 85% selama 3 menit berturut-turut) untuk memicu peringatan otomatis.

---

### 5.8. Sentinel Security Hub & Remediasi Kerentanan

1. Buka menu **Security** (`/security`).
2. Tinjau kartu **Security Score** (0 - 100) dan predikat huruf mutu keamanan (*Grade A/B/C/D/F*).
3. Klik tombol **Launch Security Scan** dan pilih target server:
   - **Full Audit**: Menjalankan seluruh pemindai sekaligus.
   - **Port Scanner**: Memeriksa port publik terbuka dan layanan berisiko tinggi (FTP, Telnet, Database).
   - **TLS/SSL Scanner**: Memeriksa sertifikat kadaluarsa, enkripsi cipher, dan protokol TLS usang.
   - **HTTP Headers Auditor**: Memeriksa kepatuhan OWASP security headers (HSTS, CSP, X-Frame-Options).
   - **Host Config Scanner**: Audit pengerasan keamanan standar CIS Linux.
   - **Vulnerability Scanner**: Memeriksa CVE paket OS yang belum diperbarui.
4. **Tindakan Remediasi**:
   - Klik salah satu temuan kerentanan untuk membuka modal rincian bukti teknis.
   - Sistem menyediakan rekomendasi perbaikan dan perintah terminal remediasi 1-klik salin (misal: perintah `ufw deny 3306` atau perbaikan konfigurasi Nginx).

---

### 5.9. Mesin Otomasi & Notifikasi Multi-Channel

1. Buka menu **Automation** (`/automation`).
2. Klik tombol **Create New Rule**.
3. Susun aturan menggunakan alur *Trigger - Condition - Action*:
   - **Trigger**: Deteksi metrik (misal: `cpu_usage > 90%`), jadwal waktu tertentu (*Cron*), atau insiden keamanan Sentinel.
   - **Condition**: Kondisi logika multi-operator (`>`, `<`, `==`, `!=`, `in`, `contains`).
   - **Action**: Kirim Webhook eksternal, Kirim Email Notifikasi, Picu Backup Instan, atau Eksekusi Restart Server.
4. Konfigurasikan Webhook Destination dengan tanda tangan HMAC-SHA256 untuk integrasi aman ke Discord, Slack, atau server webhook internal.
5. Pantau histori eksekusi otomasi dan waktu respons pada sub-menu **Automation Execution Logs** (`/automation/logs`).

---

## 6. Troubleshooting & Pemecahan Masalah

### 1. Database Connection Timeout atau Gagal Tersambung
- Periksa kembali variabel `DB_HOST`, `DB_PORT`, `DB_USER`, dan `DB_PASSWORD` di `.env`.
- Jika menggunakan Supabase Pooler pada mode Transaction (port 6543), pastikan `DB_PORT` dan parameter SSL mode diatur ke `DB_SSLMODE=require`.
- Jalankan `go run cmd/migrate/main.go -direction=up` untuk menguji koneksi basis data.

### 2. Live WebSocket / SSE Streaming Terminal Tidak Terhubung
- Pastikan backend API aktif di port `8080`.
- Periksa file `frontend/.env.local`, pastikan `NEXT_PUBLIC_API_URL=http://localhost:8080` dan `NEXT_PUBLIC_WS_URL=ws://localhost:8080/api/v1/ws`.
- Pastikan tidak ada proxy/firewall lokal yang memblokir header `Upgrade: websocket` atau `text/event-stream`.

### 3. Agent Host Tidak Mengirimkan Metrik Telemetri
- Periksa log status daemon di server target: `sudo systemctl status caelus-agent`.
- Pastikan `SERVER_ID` dan `AGENT_SECRET` pada file environment agent sesuai dengan server yang terdaftar di database.
- Pastikan firewall server mengizinkan koneksi outbound port 8080/443 ke endpoint Caelus API Control Plane.
