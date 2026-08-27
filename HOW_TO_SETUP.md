# Panduan Setup Lengkap, Skenario Nyata & Penggunaan Fitur Caelus Cloud

Dokumen ini memuat panduan komprehensif mulai dari instalasi prasyarat, konfigurasi lingkungan, inisialisasi basis data, menjalankan seluruh layanan, hingga **skenario arsitektur nyata multi-staging (Development & Production Parity)** yang dapat dikendalikan dari satu dasbor terpusat.

---

## Daftar Isi

0. [Instalasi Instan 1-Baris Perintah (Smart CLI Provisioner)](#0-instalasi-instan-1-baris-perintah-smart-cli-provisioner)
1. [Prasyarat Sistem & Perangkat Lunak](#1-prasyarat-sistem--perangkat-lunak)
2. [Konfigurasi Lingkungan (Environment Variables)](#2-konfigurasi-lingkungan-environment-variables)
3. [Inisialisasi & Migrasi Basis Data](#3-inisialisasi--migrasi-basis-data)
4. [Menjalankan Seluruh Layanan (Development & Production)](#4-menjalankan-seluruh-layanan-development--production)
5. [Panduan Penggunaan Modul Inti Caelus Cloud](#5-panduan-penggunaan-modul-inti-caelus-cloud)
   - [5.1. Registrasi Server & Onboarding Agent (BYOS & Cloud Provider)](#51-registrasi-server--onboarding-agent-byos--cloud-provider)
   - [5.2. Persistent Block Volumes & Telemetri Harddisk Fisik](#52-persistent-block-volumes--telemetri-harddisk-fisik)
   - [5.3. Virtual Networks (VPC Subnet Isolation) & Cloud Firewall](#53-virtual-networks-vpc-subnet-isolation--cloud-firewall)
   - [5.4. Container Orchestration & Web-Based Streaming Terminal](#54-container-orchestration--web-based-streaming-terminal)
   - [5.5. Object Storage Explorer (MinIO S3 & Multi-Cloud Provider)](#55-object-storage-explorer-minio-s3--multi-cloud-provider)
   - [5.6. Declarative Infrastructure as Code (IaC) & Rollback Engine](#56-declarative-infrastructure-as-code-iac--rollback-engine)
   - [5.7. Sentinel Security Hub & Remediasi Kerentanan](#57-sentinel-security-hub--remediasi-kerentanan)
   - [5.8. Monitoring Telemetri Performa Host & Alert Engine](#58-monitoring-telemetri-performa-host--alert-engine)
   - [5.9. Mesin Otomasi & Notifikasi Multi-Channel](#59-mesin-otomasi--notifikasi-multi-channel)
6. [Skenario Arsitektur Nyata: 1 Platform Mengendalikan Multi-Staging](#6-skenario-arsitektur-nyata-1-platform-mengendalikan-multi-staging)
   - [6.1. Skenario 1: Mengontrol Dev Machine & Production VPS dalam 1 Dasbor](#61-skenario-1-mengontrol-dev-machine--production-vps-dalam-1-dasbor)
   - [6.2. Skenario 2: Komunikasi Antar-Kontainer & Isolasi VPC (Frontend + Backend + DB + Redis)](#62-skenario-2-komunikasi-antar-kontainer--isolasi-vpc-frontend--backend--db--redis)
   - [6.3. Skenario 3: Kapan Pakai Block Volume vs Kapan Pakai S3 Bucket](#63-skenario-3-kapan-pakai-block-volume-vs-kapan-pakai-s3-bucket)
   - [6.4. Skenario 4: Menghubungkan Backend ke Bucket Provider (AWS S3 / Cloudflare R2)](#64-skenario-4-menghubungkan-backend-ke-bucket-provider-aws-s3--cloudflare-r2)
   - [6.5. Skenario 5: Testing Nyata di Perangkat Lain via Service Tunneling (Cloudflare Tunnel)](#65-skenario-5-testing-nyata-di-perangkat-lain-via-service-tunneling-cloudflare-tunnel)
   - [6.6. Skenario 6: Kedaulatan Penuh (Apa yang Terjadi Jika Caelus / Agent Di-uninstall)](#66-skenario-6-kedaulatan-penuh-apa-yang-terjadi-jika-caelus--agent-di-uninstall)
   - [6.7. Skenario 7: Otomatisasi Dynamic Endpoint Generator pada Agent Command](#67-skenario-7-otomatisasi-dynamic-endpoint-generator-pada-agent-command)
7. [Troubleshooting & Solusi Kendala Umum](#7-troubleshooting--solusi-kendala-umum)

---

## 0. Instalasi Instan 1-Baris Perintah (Smart CLI Provisioner)

Untuk pengguna atau enterprise yang ingin langsung memasang Caelus Cloud di server tanpa perlu *clone repo* atau konfigurasi manual, cukup jalankan perintah berikut di terminal:

```bash
curl -fsSL https://raw.githubusercontent.com/havilz/Caelus-cloud/master/scripts/install.sh | bash
# atau secara lokal dari repository yang sudah di-clone:
bash scripts/install.sh
```

### Fitur Wizard CLI Pintar:
1. **Otomatis Pasang Docker**: Jika Docker belum ada, script akan memasang Docker & Compose resmi secara otomatis.
2. **Pilihan Topologi Fleksibel**:
   - **All-in-One Full Stack**: Menjalankan Postgres, Redis, MinIO, Prometheus, Loki, API, dan UI lokal via Docker.
   - **External Managed Database (Cloud-Native)**: Menghubungkan ke PostgreSQL eksternal (AWS RDS, Supabase, Neon) & Redis (Upstash, Aiven) tanpa menyalakan kontainer DB lokal (menghemat RAM 50%+).
   - **Hybrid Workstation + Cloudflare Tunnel**: Menjalankan Control Plane di perangkat lokal dan membuka jalur Tunnel aman untuk menerima telemetri VPS jarak jauh.
3. **Auto-Generate Kunci Kriptografi**: Otomatis menghasilkan `JWT_SECRET` dan `ENCRYPTION_KEY` AES-256 berkekuatan tinggi.

---

## 1. Prasyarat Sistem & Perangkat Lunak

Sebelum memulai, pastikan perangkat Anda telah terpasang perangkat lunak berikut:

| Perangkat Lunak | Versi Minimal | Keterangan |
| :--- | :--- | :--- |
| **Go** | 1.22 atau lebih baru | Diperlukan untuk Backend API, Worker, dan Agent ([go.dev](https://go.dev)) |
| **Node.js** | 20.x atau lebih baru | Runtime JavaScript untuk frontend Next.js ([nodejs.org](https://nodejs.org)) |
| **pnpm / npm** | Node v20+ npm | Package manager frontend |
| **Docker & Compose** | Docker v24+, Compose v2+ | Untuk container runtime dan database lokal (Postgres, Redis, MinIO) |
| **GNU Make** | 4.x atau lebih baru | Untuk menjalankan perintah cepat monorepo |

---

## 2. Konfigurasi Lingkungan (Environment Variables)

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

# 2. Database PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_USER=caelus_user
DB_PASSWORD=caelus_password
DB_NAME=caelus_cloud_db
DB_SSLMODE=disable

# 3. Redis Cache & Queue
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# 4. Storage S3 / MinIO
STORAGE_PROVIDER=minio
STORAGE_ENDPOINT=http://localhost:9000
STORAGE_ACCESS_KEY=minioadmin
STORAGE_SECRET_KEY=minioadmin
STORAGE_BUCKET_NAME=caelus-system-backups
STORAGE_USE_SSL=false

# 5. Keamanan & Kriptografi
JWT_SECRET=super_secret_jwt_key_min_32_characters_long_12345
ENCRYPTION_KEY=12345678901234567890123456789012 # Tepat 32 bytes untuk AES-256-GCM
```

---

## 3. Inisialisasi & Migrasi Basis Data

Jalankan database container dan migrasi skema database:

```bash
# 1. Jalankan layanan infrastruktur di background
docker compose up -d postgres redis minio prometheus loki

# 2. Eksekusi seluruh file migrasi SQL
make migrate-up
```

---

## 4. Menjalankan Seluruh Layanan (Development & Production)

### Mode Docker Compose Terpadu (Direkomendasikan)
Jalankan seluruh ekosistem Caelus Cloud dalam 1 perintah:

```bash
docker compose up -d --build
```

Setelah aktif, akses layanan melalui browser:
* **Frontend Web Dashboard**: `http://localhost:3000`
* **Backend REST API & Swagger**: `http://localhost:8080`
* **MinIO Storage Console**: `http://localhost:9001` (User: `minioadmin`, Pass: `minioadmin`)

---

## 5. Panduan Penggunaan Modul Inti Caelus Cloud

### 5.1. Registrasi Server & Onboarding Agent (BYOS & Cloud Provider)
* **BYOS (Bring Your Own Server - Seperti VPS Debian)**:
  1. Buka `/infrastructure/vps` ➔ Klik **Add Server**.
  2. Masukkan nama server ➔ Dapatkan perintah 1-baris instalasi shell:
     ```bash
     curl -sSL http://<YOUR_API_HOST>:8080/install.sh | sudo bash -s -- --server-id="..." --secret="..." --api="..."
     ```
  3. Jalankan di terminal SSH VPS Anda. Dalam 3 detik, server akan terdeteksi aktif dan telemetri CPU/RAM/Disk otomatis tersinkronisasi.
* **Cloud Provider Provisioning (AWS / DigitalOcean / Hetzner)**:
  1. Daftarkan API Key di `/infrastructure/providers`.
  2. Di modal Add Server, pilih tab **Cloud Provider** ➔ Pilih Region & Ukuran VM ➔ Klik **Provision Server**. Caelus otomatis membuat VM di datacenter provider.

### 5.2. Persistent Block Volumes & Telemetri Harddisk Fisik
* Buka menu **Infrastructure -> Volumes** (`/infrastructure/volumes`).
* **Storage Pool Telemetry**: Dasbor otomatis membaca kapasitas SSD fisik server (`df -B1`) dan menampilkan total disk, terpakai, dan sisa ruang bebas.
* **Anti Over-Provisioning**: Form alokasi kapasitas terkunci dengan batas maksimal ruang disk nyata (`max = free_gb`).
* **Docker Named Volume Fisik**: Saat dibuat, backend mengeksekusi `docker volume create caelus-<nama>` secara fisik di harddisk host.

### 5.3. Virtual Networks (VPC Subnet Isolation) & Cloud Firewall
* Buka menu **Infrastructure -> Networks** (`/infrastructure/networks`).
* **Virtual Private Cloud (VPC)**: Buat subnet privat (misal: `production-vpc` dengan CIDR `10.20.0.0/16`).
* **Cloud Firewall (Security Groups)**: Buat aturan `ALLOW` atau `DENY` berdasarkan port (misal: `DENY TCP 8080` dari `0.0.0.0/0`).

### 5.4. Container Orchestration & Web-Based Streaming Terminal
* Buka menu **Infrastructure -> Containers** (`/infrastructure/containers`).
* **Target Node & VPC Binding**: Pilih server target, nama container, port mapping, environment variables, dan VPC Subnet tujuan.
* **Local Image Fallback**: Otomatis mendeteksi image lokal yang sudah di-build di host tanpa memerlukan push ke Docker Hub.
* **Live SSE Streaming Terminal**: Tinjau log deployment secara real-time via Server-Sent Events dengan toolbar 2-tier yang bersih.

### 5.5. Object Storage Explorer (MinIO S3 & Multi-Cloud Provider)
* Buka menu **Storage** (`/storage`).
* **Multi-Cloud S3 Explorer**: Kelola bucket lokal (MinIO) maupun bucket cloud eksternal (AWS S3 / Cloudflare R2 / DigitalOcean Spaces).
* **File Management**: Upload file multipart, unduh, hapus, dan buat tautan unduhan aman sementara (*Pre-Signed URL*).
* **Backup Management (`/storage/backups`)**: Jadwalkan snapshot backup database otomatis dengan retensi rotasi berkala.

---

## 6. Skenario Arsitektur Nyata: 1 Platform Mengendalikan Multi-Staging

Berikut adalah cetak biru skenario nyata bagaimana Caelus Cloud digunakan sebagai **Satu Pusat Kendali (Single Pane of Glass)** untuk tahap **Development hingga Production**:

```text
                                [ CAELUS CLOUD CONTROL PLANE ]
                                (1 Dasbor untuk Semua Staging)
                                               │
             ┌─────────────────────────────────┴─────────────────────────────────┐
             ▼                                                                   ▼
    [ NODE 1: KOMPUTER LOKAL / DEV ]                                   [ NODE 2: VPS PRODUCTION DEBIAN ]
     ├── Caelus Agent (Local Node)                                      ├── Caelus Agent (Remote Node)
     ├── VPC: "development-staging-vpc"                                 ├── VPC: "production-staging-vpc"
     ├── Bucket: "dev-media-bucket" (MinIO Local)                       ├── Bucket: "prod-media-bucket" (MinIO / AWS S3)
     └── Containers: Dev Frontend + Backend + Postgres                  └── Containers: Prod Frontend + Backend + Postgres
```

---

### 6.1. Skenario 1: Mengontrol Dev Machine & Production VPS dalam 1 Dasbor

Anda tidak perlu menggunakan banyak tool terpisah untuk laptop lokal dan server VPS produksi:

1. **Daftarkan Komputer Lokal**: Daftarkan mesin lokal Anda sebagai Node 1 (`dev-laptop`).
2. **Daftarkan Server VPS**: Pasang Agent di server VPS Debian Anda sebagai Node 2 (`prod-vps`).
3. **Pemisahan Jaringan (VPC Staging)**:
   - Buat Network 1: **`dev-staging-vpc`** (CIDR `10.10.0.0/16`) ditugaskan ke Node 1.
   - Buat Network 2: **`prod-staging-vpc`** (CIDR `10.20.0.0/16`) ditugaskan ke Node 2.
4. **Hasilnya**: Anda bisa menguji coba fitur baru di Node 1, lalu men-deploy container produksi ke Node 2 cukup dengan mengganti pilihan dropdown **Target Server** di Caelus!

---

### 6.2. Skenario 2: Komunikasi Antar-Kontainer & Isolasi VPC (Frontend + Backend + DB + Redis)

Di dalam satu network/VPC yang sama (misal `production-staging-vpc`):

```text
[ Browser Pengguna Luar ] ──> (Port 3000) ──> [ Container: Frontend ]
                                                     │ (Panggilan API)
                                                     ▼
                                            [ Container: Backend API ]
                                             ├──> memanggil "caelus-postgres:5432"
                                             └──> memanggil "caelus-redis:6379"
```

* **Komunikasi Internal Menggunakan Nama Kontainer**:
  Backend memanggil database cukup dengan `DB_HOST=caelus-postgres` dan `REDIS_HOST=caelus-redis` berkat fitur **Internal DNS VPC**.
* **Keamanan Maksimal**:
  Port Database (`5432`) dan Redis (`6379`) **tidak perlu diekspos ke IP publik**. Hanya port Frontend (`3000`) yang dibuka untuk pengguna luar.
* **Kinerja Firewall (Security Groups)**:
  Jika aturan firewall disetel **`DENY TCP 8080 (0.0.0.0/0)`**, maka pengguna publik dari internet tidak bisa mengakses port 8080 secara langsung, namun Frontend di dalam VPC yang sama **tetap bisa memanggil Backend API di port 8080 secara lancar**.

---

### 6.3. Skenario 3: Kapan Pakai Block Volume vs Kapan Pakai S3 Bucket

| Kebutuhan Data | Gunakan Layanan Ini | Alasan Teknis |
| :--- | :--- | :--- |
| **Data Tabel PostgreSQL, MySQL, Redis** | **Block Volume (`/infrastructure/volumes`)** | Membutuhkan partisi disk fisik dengan IOPS tinggi (NVMe/SSD) dan latensi rendah 0ms agar transaksi database cepat. |
| **Foto Profil, Video Produk, Dokumen PDF, File Backup** | **Object Storage (`/storage`)** | File mentah (*unstructured*) yang diunggah pengguna. Disimpan di S3/MinIO dan diakses via link URL publik/privat tanpa membebani database. |

---

### 6.4. Skenario 4: Menghubungkan Backend ke Bucket Provider (AWS S3 / Cloudflare R2)

Sistem Caelus Cloud menggunakan standar industri **S3-Compatible Protocol**. Artinya, Anda **TIDAK PERLU MENGUBAH KODE BACKEND ANDA** saat berpindah dari MinIO lokal ke AWS S3:

#### Cukup Ganti 4 Baris Variabel Environment (`.env`):

* **Saat di Lingkungan Lokal (MinIO)**:
  ```bash
  STORAGE_ENDPOINT=http://caelus-minio:9000
  STORAGE_ACCESS_KEY=minioadmin
  STORAGE_SECRET_KEY=minioadmin
  STORAGE_BUCKET_NAME=assets-development
  ```

* **Saat di Lingkungan Produksi (AWS S3)**:
  ```bash
  STORAGE_ENDPOINT=https://s3.ap-southeast-1.amazonaws.com
  STORAGE_ACCESS_KEY=AKIAIOSFODNN7EXAMPLE
  STORAGE_SECRET_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
  STORAGE_BUCKET_NAME=assets-production-live
  STORAGE_REGION=ap-southeast-1
  ```

---

### 6.5. Skenario 5: Testing Nyata di Perangkat Lain via Service Tunneling (Cloudflare Tunnel)

Jika Caelus atau aplikasi Anda berjalan di laptop lokal dan Anda ingin mengujinya langsung dari HP/Tablet tanpa ribet setting router WiFi:

1. Deploy kontainer tunneling langsung dari dasbor Caelus (`/infrastructure/containers`):
   - **Image**: `cloudflare/cloudflared:latest`
   - **Command**: `tunnel --url http://caelus-frontend:3000`
   - **VPC Network**: `caelus-network`
2. Di terminal log Caelus akan muncul tautan publik resmi ber-HTTPS:
   ```text
   [stdout] INF https://caelus-dev-preview.trycloudflare.com
   ```
3. Buka link tersebut di HP (menggunakan kuota seluler 4G/5G) ➔ Dashboard atau aplikasi web Anda langsung bisa diuji secara live dari perangkat mana pun di dunia!

---

### 6.6. Skenario 6: Kedaulatan Penuh (Apa yang Terjadi Jika Caelus / Agent Di-uninstall)

Caelus Cloud menganut prinsip **"Zero Vendor Lock-in & Full Data Sovereignty"**:

1. **Semua Aset Tertanam Fisik di OS Linux VPS**:
   - Kontainer berjalan di bawah Docker Daemon sistem operasi Linux VPS.
   - Jaringan terdaftar di Linux Bridge Network.
   - Volume tersimpan di filesystem SSD Linux (`/var/lib/docker/volumes`).
2. **Jika Caelus Dihapus**:
   - Kontainer website, database, dan API di VPS Anda **tetap berjalan (RUNNING) 100% dan tetap online melayani pengunjung**.
3. **Jika Caelus Agent di VPS Di-uninstall**:
   - Docker Engine di VPS **tidak akan mematikan atau menghapus kontainer Anda**. Seluruh file `.env` dan kredensial AWS yang sudah tertanam di memori kontainer tetap aktif dan berjalan normal!

---

### 6.7. Skenario 7: Otomatisasi Dynamic Endpoint Generator pada Agent Command

Perintah instalasi Caelus Agent (`curl -sSL ... | sudo bash`) dirancang adaptif secara dinamis mengikuti hostname tempat dashboard dibuka:

* **Jika dibuka di `localhost:3000`** ➔ Perintah otomatis mengarah ke `http://localhost:8080`.
* **Jika dibuka di domain publik `https://caelus.cloud`** ➔ Perintah otomatis mengarah ke `https://caelus.cloud`.
* **Jika dibuka via IP Publik VPS `http://103.180.20.5:3000`** ➔ Perintah otomatis mengarah ke `http://103.180.20.5:8080`.

---

## 7. Troubleshooting & Solusi Kendala Lengkap (Production & Edge Cases)

### 1. Database Connection Timeout atau Gagal Tersambung
> **Kondisi Terjadinya Kendala:**
> Saat pertama kali menginstal Caelus di VPS, atau ketika memilih topologi External Managed Database (Supabase/Neon/AWS RDS) dengan kredensial atau firewall pooler yang belum diizinkan.
* **Penyebab**: Variabel host, port, user, atau password database di `/opt/caelus/.env` salah, atau koneksi SSLMode tidak sesuai.
* **Solusi**:
  * Periksa konfigurasi di file konfigurasi server: `nano /opt/caelus/.env`.
  * Pastikan kontainer migrasi database berjalan sukses: `docker logs caelus-migrate`.
  * Restart layanan Caelus di server: `cd /opt/caelus && docker compose restart api`.

---

### 2. Live WebSocket / SSE Streaming Terminal Tidak Terhubung
> **Kondisi Terjadinya Kendala:**
> Saat membuka fitur Web Terminal / Live Logs melalui Reverse Proxy (Nginx, Traefik, Cloudflare Proxy, Caddy) tanpa konfigurasi WebSocket upgrade.
* **Penyebab**: Proxy perantara memotong koneksi HTTP connection upgrade sehingga protokol WebSocket terputus.
* **Solusi**:
  * Pastikan backend API aktif dan dapat diakses pada port yang ditentukan di `/opt/caelus/.env` (variabel `APP_PORT`, default: `8080`).
  * Pastikan firewall atau proxy reverse menyertakan header `Upgrade: websocket` dan `Connection: Upgrade`.

---

### 3. Agent Host Tidak Mengirimkan Telemetri (Umum)
> **Kondisi Terjadinya Kendala:**
> Setelah perintah onboarding dijalankan di VPS, namun status server di dashboard Caelus tetap berstatus Merah / Offline.
* **Penyebab**: Firewall host lokal atau security group cloud memblokir lalu lintas keluar (*outbound traffic*), atau service agent mengalami crash.
* **Solusi**:
  * Periksa status service agent di VPS: `systemctl status caelus-agent`.
  * Cek log detail error telemetri: `journalctl -u caelus-agent -n 20 --no-pager`.
  * Pastikan firewall VPS mengizinkan koneksi keluar port 8080 / 443 ke Caelus API Control Plane.

---

### 4. `curl: (6) Could not resolve host` pada Tailscale MagicDNS di Container VPS / LXC
> **Kondisi Terjadinya Kendala:**
> Saat Control Plane Caelus dijalankan di laptop dengan Tailscale HTTPS MagicDNS (`https://laptop.tailnet.ts.net`), dan perintah onboarding dijalankan pada VPS bertipe **Container / LXC / OpenVZ** (seperti VPS murah, VPS NAT, atau Proxmox unprivileged container).
* **Penyebab**: Container LXC di beberapa penyedia hosting tidak menyuntikkan DNS Tailscale ke dalam file `/etc/resolv.conf` container tersebut.
* **Solusi**:
  Gunakan **Alamat IP Numerik Tailscale** (misal `http://100.115.162.110:8080`) yang tersedia langsung pada tombol preset cepat di modal tambah server Caelus Dashboard, bukan menggunakan nama domain `.ts.net`.

---

### 5. `503 Service Unavailable: no backend` atau Ketiadaan Device `/dev/net/tun` pada LXC Container
> **Kondisi Terjadinya Kendala:**
> Saat mencoba menjalankan `tailscaled` atau `tailscale up` di dalam VPS container (LXC/Docker container di VPS) di mana provider membatasi hak akses kernel virtual device TUN/TAP demi keamanan multi-tenant.
* **Penyebab**: VPS berjenis LXC Unprivileged tidak diberikan izin oleh kernel hypervisor induk untuk membuat virtual network adapter `tailscale0`.
* **Solusi (Userspace Networking & SOCKS5 Outbound Proxy)**:
  Jalankan daemon `tailscaled` dalam mode userspace networking dengan proxy lokal port `1055`:
  ```bash
  # 1. Bersihkan proses lama & socket yang menggantung
  pkill -9 tailscaled 2>/dev/null
  rm -f /var/run/tailscale/tailscaled.sock

  # 2. Jalankan tailscaled mode userspace dengan proxy SOCKS5
  tailscaled --tun=userspace-networking --socks5-server=localhost:1055 --state=/var/lib/tailscale/tailscaled.state > /var/log/tailscaled.log 2>&1 &
  sleep 2

  # 3. Hubungkan menggunakan Auth Key Tailscale
  tailscale up --authkey="tskey-auth-..." --force-reauth
  ```

---

### 6. Menjalankan Instalasi Agent Melalui Jalur Userspace SOCKS5 Proxy
> **Kondisi Terjadinya Kendala:**
> Pada VPS container yang menggunakan Tailscale mode userspace di atas (karena tidak ada interface kernel `tailscale0`), sehingga perintah `curl` dan service `caelus-agent` perlu diarahkan melewati proxy lokal.
* **Penyebab**: Tanpa proxy, perintah `curl` mencoba menghubungi IP Tailscale `100.x.y.z` melalui internet publik biasa dan berakhir timeout.
* **Solusi**:
  Jalankan perintah instalasi dengan menyertakan awalan `ALL_PROXY`:
  ```bash
  ALL_PROXY=socks5://localhost:1055 curl -sSL http://<IP_TAILSCALE>:8080/install.sh | ALL_PROXY=socks5://localhost:1055 bash -s -- --server-id="<SERVER_UUID>" --secret="<AGENT_SECRET>" --api="http://<IP_TAILSCALE>:8080"
  ```
  *(Script installer Caelus otomatis menyimpan variabel `ALL_PROXY` ke dalam `/opt/caelus/agent.env` sehingga service Systemd tetap terhubung permanen).*

---

### 7. `curl: (23) client returned ERROR on write` saat Memperbarui Binary Agent
> **Kondisi Terjadinya Kendala:**
> Saat melakukan update agent di server produksi yang sudah aktif berjalan, di mana file binary lama sedang dieksekusi oleh memori proses Linux.
* **Penyebab**: File `/opt/caelus/caelus-agent` sedang aktif berjalan di memori dan dikunci oleh kernel Linux (*Text file busy* / Write Lock).
* **Solusi**:
  Hentikan service terlebih dahulu atau gunakan pemindahan file atomik (*atomic file replacement* via `/tmp`):
  ```bash
  systemctl stop caelus-agent
  ALL_PROXY=socks5://localhost:1055 curl -sSL http://<IP_TAILSCALE>:8080/agent-bin -o /tmp/caelus-agent
  mv -f /tmp/caelus-agent /opt/caelus/caelus-agent
  chmod +x /opt/caelus/caelus-agent
  systemctl start caelus-agent
  ```

---

### 8. `'xterm-kitty': unknown terminal type` saat Mengetik `clear` atau `nano` di VPS
> **Kondisi Terjadinya Kendala:**
> Saat developer melakukan SSH ke remote VPS minimal (Debian/Alpine) menggunakan terminal modern di laptop (seperti Kitty, Alacritty, Ghostty, WezTerm).
* **Penyebab**: Terminal modern mengirimkan identitas `$TERM` (misal `xterm-kitty`) yang belum terdaftar di database `/usr/share/terminfo` bawaan OS VPS minimal.
* **Solusi**:
  Jalankan 1 baris perintah ini di terminal VPS untuk menyamakan terminfo standar Linux:
  ```bash
  echo 'export TERM=xterm-256color' >> ~/.bashrc && export TERM=xterm-256color
  ```

---

### 9. Cloudflare WARP Default Route Hijacking pada Single-NIC VPS (SSH Terputus)
> **Kondisi Terjadinya Kendala:**
> Saat mencoba menghubungkan VPS remote menggunakan Cloudflare WARP Client (`warp-cli connect`) pada VPS yang hanya memiliki 1 kartu jaringan (Single-NIC) dengan port SSH non-standar (misal port NAT `10082`).
* **Penyebab**: WARP Client membajak rute gateway default `0.0.0.0/0` ke interface `CloudflareWARP`, sehingga paket balasan koneksi SSH terkirim ke Cloudflare alih-alih ke router gateway provider host, menyebabkan SSH langsung terputus seketika.
* **Solusi**:
  Hindari penggunaan Cloudflare WARP Client pada VPS single-NIC. Gunakan **Tailscale Mesh VPN** yang hanya merutekan subnet privat CGNAT `100.64.0.0/10` tanpa membajak default gateway internet utama VPS.
