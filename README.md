# Caelus Cloud

Platform Manajemen Infrastruktur Cloud Terpadu (*Cloud Infrastructure Management & Control Plane Platform*) modern yang memungkinkan pengelolaan, pemantauan, pengamanan, dan otomatisasi server multi-lingkungan dari satu dashboard terpusat.

Caelus Cloud bertindak sebagai *control layer* independen di atas infrastruktur milik pengguna (*Bring Your Own Server / BYOS* maupun penyedia cloud publik), menggabungkan kapabilitas orkestrasi VPS multi-cloud, deklaratif *Infrastructure as Code* (IaC), *container deployment pipeline*, manajemen *object storage*, telemetri performa *real-time*, mesin otomasi *event-driven*, dan subsistem audit keamanan modular (*Sentinel*).

---

## Arsitektur Sistem

```text
                                CAELUS CLOUD CONTROL PLANE
                                            │
        ┌───────────────────┬───────────────┴───────────────┬───────────────────┐
        │                   │                               │                   │
  Infrastructure         Storage                       Monitoring            Security
    Management          Management                    Observability         (Sentinel)
        │                   │                               │                   │
  ┌─────┴─────┐             │                         ┌─────┴─────┐       ┌─────┴─────┐
 Multi-   Container     S3 / MinIO                 Prometheus   Loki    Port     TLS/SSL
 Cloud    Pipeline       Storage                    Metrics     Logs    Headers  Host/Vuln
 (IaC)   (Live Logs)        │                               │                   │
        │                   │                               │                   │
        └───────────────────┴───────────────┬───────────────┴───────────────────┘
                                            │
                                     Automation Engine
                                            │
                               ┌────────────┴────────────┐
                         Task Scheduler            Redis Queue
                         (Cron Routines)         (Workers & DLQ)
```

---

## 6 Domain Utama & Katalog Fitur Lengkap

### 1. Infrastructure Management & Multi-Cloud Compute
- **Multi-Cloud Public Drivers**: Integrasi terpadu ke berbagai penyedia VPS dan compute publik:
  - **Amazon Web Services (AWS EC2)**: Pengelolaan daur hidup instance (Launch, Describe, Start, Stop, Reboot, Terminate, Resize).
  - **DigitalOcean**: Manajemen Droplet v2 API, power operations, dan resizing.
  - **Hetzner Cloud**: Integrasi Hetzner Cloud REST API dan perubahan tipe resource.
  - **Contabo Cloud**: Manajemen instance via Contabo Compute API.
  - **Bring Your Own Server (BYOS)**: Pengelolaan server mandiri dan dedicated host on-premise.
- **Onboarding Otomatis 1 Baris (`/install.sh`)**: Script instalasi shell instan yang mengonfigurasi `caelus-agent` sebagai daemon systemd Linux.
- **Auto-Sync Spesifikasi Perangkat Keras**: Sinkronisasi otomatis CPU cores, RAM, Disk, OS Platform, dan Hostname langsung dari telemetri agent ke database.
- **Periodic Resource Status Sync Engine**: Background worker yang secara berkala (interval 60 detik) melakukan rekonsiliasi status daya dan IP remote instance pihak ketiga.
- **Heartbeat Liveness Watchdog**: Pemantauan detak jantung server setiap 15 detik yang mendeteksi server offline dan memicu event pemulihan.

### 2. Declarative Infrastructure as Code (IaC) & Rollback Engine
- **YAML Manifest Management**: Pendefinisian seluruh konfigurasi server, storage, container, dan aturan firewall dalam format deklaratif terpadu.
- **YAML Parser & Syntax Validator**: Validasi skema semantik dan sintaksis dengan pelaporan baris dan kolom error secara *real-time*.
- **Starter Templates**: Template konfigurasi siap pakai untuk berbagai arsitektur (*Fullstack App*, *Multi-Cloud High Availability*, *Microservices Stack*).
- **Plan Engine (State Comparator)**: Komparasi cerdas antara *Desired State* (YAML) dengan *Actual State* (Database/Cloud) yang menghasilkan visualisasi *diff* terstruktur (+ Create, ~ Update, - Delete, = No-op).
- **Transactional Apply Engine & Automatic Rollback**: Eksekusi perubahan transaksional dengan mekanisme *LIFO rollback stack* jika terjadi kegagalan eksekusi bertahap.
- **State Snapshot Versioning & 1-Click Rollback**: Riwayat versi snapshot state dengan checksum integritas SHA-256 dan kemampuan restore ke versi state terdahulu secara instan.

### 3. Container Orchestration & Real-time Live Log Terminal
- **Docker Deployment Pipeline**: Pipeline asinkron deployment kontainer (Pulling Image -> Schema Validation -> Port & Volume Binding -> Environment Config -> Launch -> Healthcheck).
- **Quick Image Presets**: Peluncuran cepat image populer (Nginx, Redis, PostgreSQL, Node.js) dengan konfigurasi otomatis.
- **Dynamic Configuration**: Dukungan penuh port mapping, volume persistence binding, restart policy, dan custom environment variables.
- **Web-Based Streaming Terminal**: Konsol ANSI gelap berkecepatan tinggi yang mengalirkan log stdout, stderr, dan system event secara langsung melalui *Server-Sent Events (SSE)*.
- **Console Utilities**: Filter log per kategori, pencarian teks interaktif, toggle auto-scroll, dan 1-klik copy log terminal.
- **Deployment Rollback**: Kemampuan membalikkan deployment kontainer ke versi image tag sebelumnya jika terjadi galat pada rilis baru.

### 4. Storage Management & Disaster Recovery
- **S3-Compatible Object Storage**: Abstraksi penyimpanan terpadu yang kompatibel dengan MinIO, AWS S3, dan Cloudflare R2.
- **Interactive Object Explorer**: Penjelajah direktori file, drag-and-drop multipart upload, penghapusan, dan manajemen bucket multi-region.
- **Pre-signed URLs Generator**: Pembuatan tautan unduh/unggah aman terautentikasi dengan masa kedaluwarsa 15 menit hingga 24 jam.
- **Automated Backup Policies**: Penjadwal pencadangan otomatis data server dan snapshot dengan kebijakan retensi penyimpanan (*Retention Policy*).

### 5. Monitoring, Telemetry & Observability
- **Lightweight Telemetry Agent (`caelus-agent`)**: Binary Go tunggal berukuran ringan (< 10MB) dengan konsumsi memori minimal (< 15MB RSS) yang mengumpulkan metrik sistem langsung dari *Linux procfs* (`/proc/stat`, `/proc/meminfo`, `/proc/net/dev`, `statfs`) dan *Docker Unix Socket*.
- **Live WebSocket Streams**: Distribusi metrik utilisasi CPU, RAM, Disk, dan Network Throughput ke antarmuka web per detik tanpa beban reload.
- **Prometheus & Grafana Loki Query Adapters**: Integrasi adapter ke penyimpanan time-series Prometheus dan agregasi log terdistribusi Loki.
- **Alert Threshold Evaluator**: Mesin evaluasi anomali beban server dengan visualisasi grafik performa interaktif.

### 6. Sentinel Security & Event-Driven Automation Engine
- **Modular Scanner Workers**:
  - **Port Scanner**: Audit keterbukaan port publik dan deteksi eksposur database berisiko tinggi (FTP, Telnet, MySQL, Postgres, Redis, MongoDB).
  - **TLS/SSL Scanner**: Validasi masa berlaku sertifikat, kecocokan hostname, dan deteksi protokol usang (TLS 1.0/1.1).
  - **HTTP Security Headers Auditor**: Audit header kepatuhan OWASP (HSTS, CSP, X-Frame-Options, X-Content-Type-Options, Referrer-Policy).
  - **Host Configuration Scanner**: Audit pengerasan keamanan standar CIS Linux dan stabilitas sistem.
  - **Vulnerability Scanner**: Pemindaian kerentanan CVE paket sistem operasi dan dependensi umum.
- **Finding Normalizer & Fingerprinting**: Deduplikasi temuan kerentanan dengan sidik jari unik untuk mencegah spam laporan.
- **Unified Security Score Engine**: Kalkulasi skor mutu keamanan (0 - 100) dan predikat prediktif (Grade A/B/C/D/F) beserta rekomendasi perbaikan 1-klik salin.
- **Distributed Task Queue & Scheduler**: Mesin antrean tugas berbasis Redis dengan dukungan *exponential backoff retry* dan *Dead Letter Queue* (DLQ).
- **Rule Engine**: Pembangun aturan otomasi berbasis *Trigger - Condition - Action* dengan notifikasi multi-saluran via HTTP Webhook (HMAC-SHA256 Signed) dan Email SMTP.

---

## Standar Keamanan Enterprise

1. **Encrypted at Rest (AES-256-GCM)**: Seluruh kredensial sensitif provider (API Key, Secret Key, SSH Key) dienkripsi secara simetris di tingkat basis data dengan algoritma militer AES-256-GCM.
2. **Dynamic In-Memory Decryption**: Data rahasia hanya didekripsi di dalam memori kerja (RAM) secara sementara saat operasi API berlangsung dan segera dibersihkan.
3. **Multi-Tenant Row Level Security (RLS)**: Isolasi ketat data antar organisasi pada tingkat basis data PostgreSQL menggunakan kebijakan PostgreSQL RLS policies.
4. **Data Masking**: API respons publik maupun privat tidak pernah mengembalikan plaintext kredensial rahasia (`json:"-"`).

---

## Matriks Tech Stack

| Komponen | Teknologi | Keterangan |
| :--- | :--- | :--- |
| **Backend API** | Go 1.22+, Chi Router | Clean Architecture, REST API, WebSocket & SSE Hub |
| **IaC & Pipelines** | Go, YAML Parser, Docker Engine | Desired/Actual Diff Engine, LIFO Rollback, Live SSE Streaming |
| **Worker Engine** | Go, Redis 7+ | Distributed Task Queue, Cron Scheduler, Asynchronous Jobs |
| **Host Agent** | Go (`caelus-agent`) | Linux procfs collector, Docker unix domain socket inspector |
| **Frontend Panel** | Next.js 16, TypeScript, React 19 | Tailwind CSS v4, Zustand, Supabase Theme Design System |
| **Database** | PostgreSQL 16 | Row Level Security (RLS) tenant isolation, pgxpool |
| **Cache & Queue** | Redis 7 | Distributed Queue, Delayed Jobs, Pub/Sub, DLQ |
| **Object Storage** | MinIO / AWS S3 / Cloudflare R2 | S3-Compatible Storage API & Presigned URLs |
| **Observability** | Prometheus & Grafana Loki | Time-series metrics & Log stream aggregation |

---

## Dokumentasi Terkait & Panduan Lengkap

- **[Panduan Setup & Cara Penggunaan Fitur (How to Setup)](HOW_TO_SETUP.md)**: Panduan instalasi dari awal hingga akhir, konfigurasi `.env`, inisialisasi basis data, menjalankan layanan via Makefile, dan tutorial langkah-demi-langkah penggunaan setiap fitur.
- **[Struktur Proyek & Cetak Biru Arsitektur](docs/PROJECT_STRUCTURE.md)**: Dokumentasi hierarki direktori monorepo, batas impor Clean Architecture, diagram ERD, dan definisi skema tabel PostgreSQL.
- **[Spesifikasi Ruang Lingkup Proyek](docs/PROJECT.md)**: Dokumen perencanaan arsitektur, 6 fase eksekusi, dan strategi staging lingkungan.
- **[Backend REST API Reference](backend/README.md)**: Dokumentasi endpoint REST API, parameter JSON, dan petunjuk operasional backend.
- **[Frontend Control Panel Reference](frontend/README.md)**: Dokumentasi antarmuka Next.js App Router, token desain tema, dan client service.
- **[Host Agent Daemon Reference](agent/README.md)**: Dokumentasi spesifikasi daemon telemetri `caelus-agent`.
- **[Changelog Proyek](changelog/CHANGELOG.md)**: Catatan histori perubahan dan pembaruan berkala terstempel waktu.
