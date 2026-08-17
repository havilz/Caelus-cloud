# Caelus Cloud Project

Dokumen ini berisi spesifikasi arsitektur dan perencanaan teknis untuk proyek **Caelus Cloud**.
Caelus Cloud adalah platform manajemen infrastruktur cloud (*Cloud Infrastructure Management Platform*) berupa *control panel* terpusat yang memungkinkan pengguna mengelola, memonitor, mengamankan, dan mengotomatisasi infrastruktur server dari satu dashboard terpadu.

---

## 1. Gambaran Umum

Caelus Cloud tidak mengharuskan pengguna memiliki server dari penyedia tertentu. Platform ini bertindak sebagai *control layer* di atas infrastruktur yang sudah dimiliki oleh pengguna.

```text
                    CAELUS CLOUD
                         │
               ┌─────────┴─────────┐
               │                   │
         Infrastructure        Security
               │                   │
        ┌──────┼──────┐            │
        │      │      │            │
       VPS  Storage Network     Sentinel
        │      │
        │      │
     AWS/VPS   S3
    Provider  Provider
```

---

## 2. Lima Domain Utama

### 2.1. Infrastructure Management
Berfungsi untuk mengelola berbagai sumber daya komputasi dan jaringan:
- VPS / Virtual Machine
- Docker Container
- Volume & Persistent Storage
- Jaringan (Network) & Firewall Rules
- SSH Key Management
- Snapshot & Backup Instance
- Automated Deployment

**Contoh Monitoring Status & Resource:**
```text
VPS: production-api
Status: Running (Port 8080)

CPU Usage     : 34%
RAM Usage     : 1.2 / 2.0 GB
Disk Usage    : 28 / 40 GB
Network I/O   : 2.4 GB
Uptime        : 14 days
```

**Operasi yang Didukung (via API Provider / Agent):**
- Restart / Reboot VPS
- Shutdown VPS
- Resize Spec (CPU, RAM, Disk)
- Attach / Detach Volume

---

### 2.2. Storage Management
Integrasi dengan penyimpanan berbasis Amazon S3 atau penyedia object storage lain yang kompatibel dengan protokol S3 (S3-compatible storage).

**Struktur Hirarki Penyimpanan:**
```text
Storage
│
├── caelus-backups
├── streamvault-assets
├── project-files
└── caelus-logs
```

**Informasi & Metrik Bucket:**
- Nama Bucket & Region
- Total Objek & Struktur Direktori
- Kapasitas Penyimpanan Terpakai
- Bandwidth / Transfer Data
- Timestamp Modifikasi Terakhir

**Fitur & Operasi:**
- Upload, Download, dan Hapus Objek
- Manajemen Siklus Hidup Objek (Lifecycle Management)
- Backup & Restore Terjadwal
- Pembuatan URL Terautentikasi (Signed URLs)

---

### 2.3. Monitoring & Observability
Mengumpulkan dan memvisualisasikan metrik kinerja server pengguna secara real-time.

**Metrik yang Dikumpulkan:**
- Utilisasi CPU, RAM, dan Disk
- Trafik Jaringan (Inbound / Outbound)
- Daftar Proses Berjalan & Status Container
- System Logs & Application Logs
- Uptime & Response Time / Latency

**Visualisasi & Notifikasi:**
```text
CPU Usage (Grafik Real-Time)
100% ┤
 80% ┤       ╭─╮
 60% ┤    ╭──╯ ╰──╮
 40% ┤───╯        ╰──
 20% ┤
  0% └────────────────
       10  11  12  13 (Jam)
```

**Contoh Aturan Notifikasi:**
- CPU Usage > 90% selama 5 menit berturut-turut -> Trigger Alert ke Administrator.

---

### 2.4. Security (Sentinel)
Sentinel adalah subsistem keamanan modular di dalam Caelus Cloud yang bertugas melakukan asesmen keamanan berkala pada infrastruktur pengguna.

```text
CAELUS CLOUD
│
├── Infrastructure
├── Storage
├── Monitoring
├── Automation
│
└── Sentinel Security Subsystem
```

**Cakupan Asesmen Sentinel:**
- Deteksi Port Terbuka (Port Exposure)
- Validasi Konfigurasi TLS/SSL & Sertifikat
- Audit Security Headers (HTTP/HTTPS)
- Service Discovery & Port Scanning
- Pemindaian Kerentanan Dependensi (Dependency Vulnerabilities)
- Pemeriksaan Kesalahan Konfigurasi Sistem & Container (Misconfigurations)
- Deteksi Anomali & Analisis Log Keamanan

**Contoh Laporan Keamanan:**
```text
Security Score: 82 / 100

Temuan Kerentanan:
- CRITICAL : 1
- HIGH     : 3
- MEDIUM   : 8
- LOW      : 14
```

---

### 2.5. Automation Engine
Sistem otomasi berbasis kejadian (*event-driven automation*) untuk alur kerja operasional, pemeliharaan, dan keamanan.

**Contoh Alur Kerja Otomasi:**

*Otomasi Alert Beban Server:*
```text
WHEN  : VPS CPU > 90%
FOR   : 5 Minutes
THEN  : Create High-Priority Alert
AND   : Notify Administrator via Email/Webhook
```

*Otomasi Backup Terjadwal:*
```text
WHEN  : Backup schedule reached (02:00 UTC)
THEN  : Create VPS snapshot
THEN  : Export snapshot & upload backup to S3
THEN  : Verify backup integrity
THEN  : Send success notification
```

*Otomasi Respon Insiden Keamanan:*
```text
WHEN  : Sentinel detects HIGH or CRITICAL severity issue
THEN  : Create Security Incident Record
AND   : Quarantine / Block offending port or IP
AND   : Notify Security Administrator immediately
```

---

## 3. Arsitektur Sistem

```text
                         CAELUS CLOUD
                              │
             ┌────────────────┼────────────────┐
             │                │                │
        Infrastructure    Monitoring        Security
             │                │                │
       ┌─────┼─────┐          │             Sentinel
       │     │     │          │                │
      VPS  Docker Network   Metrics         Scanner
       │                      │                │
       └──────────┬───────────┴────────────────┘
                  │
             Automation
                  │
          ┌───────┼────────┐
          │       │        │
       Backup   Deploy   Alert
          │
          ▼
      Object Storage
          │
          ▼
      S3 / S3-Compatible
```

---

## 4. Target & Fase Pengembangan

| Versi | Fase | Fokus Utama |
| :--- | :--- | :--- |
| **V1** | **MVP** | Menggunakan *mock provider* untuk memvalidasi seluruh fungsi inti dan alur bisnis kontrol panel. |
| **V2** | **Personal Infrastructure** | Integrasi langsung ke infrastruktur riil: VPS (SSH/Agent), Docker daemon, PostgreSQL, S3 storage, dan monitoring metrik. |
| **V3** | **Multi-Provider** | Implementasi layer abstraksi provider pihak ketiga (AWS, Hetzner, Contabo, DigitalOcean, Custom VPS). |
| **V4** | **DevSecOps Platform** | Integrasi mendalam Sentinel dan engine otomasi dengan siklus tertutup (*Closed-loop DevSecOps*). |

**Alur DevSecOps Terintegrasi (V4):**
```text
Deploy ──> Monitor ──> Scan ──> Detect ──> Alert ──> Remediate ──> Audit
```

---

## 5. Rincian Tech Stack & Arsitektur Modul

### 5.1. Frontend (Next.js + TypeScript)
- **Framework**: Next.js (App Router)
- **Bahasa**: TypeScript
- **Styling**: Tailwind CSS
- **Komponen UI**: Shadcn/ui
- **State Management**: Zustand
- **Fitur Utama**: Server & Client State Synchronization, Session Management, Dynamic Dashboard Routing, Real-time Charts, dan Complex Forms.

**Struktur Navigasi Dashboard:**
```text
Dashboard
├── Overview
├── Infrastructure
│   ├── VPS
│   ├── Containers
│   ├── Networks
│   └── Volumes
├── Storage
├── Monitoring
├── Security (Sentinel)
├── Automation
└── Settings
```

---

### 5.2. Backend API & Workers (Go)
Dipilih karena performa tinggi, efisiensi memori, konkurensi native (goroutine), serta kemudahan kompilasi binary tanpa dependensi eksternal.

- **Router / HTTP Framework**: Chi (ringan, idiomatik, dan standar `net/http`)
- **Pemisahan Binary & Komponen**:
  1. `caelus-api`: Menangani HTTP REST API, autentikasi, otorisasi, manajemen resource, dan koneksi WebSocket/SSE.
  2. `caelus-worker`: Menangani background jobs, eksekusi backup, deployment pipeline, scheduled scanning, dan pemrosesan metrik.
  3. `caelus-agent`: Binary ringan yang diinstal pada VPS pengguna untuk pengumpulan metrik internal dan eksekusi perintah lokal.

**Komunikasi Agent ke Control Plane:**
```text
              CAELUS CLOUD
                   │
              HTTPS / mTLS
                   │
                   ▼
               VPS Agent
                   │
        ┌──────────┼──────────┐
        │          │          │
       CPU        RAM       Docker
```

---

### 5.3. Basis Data & Cache

#### PostgreSQL (Primary Database)
Menyimpan data relasional utama:
- `users`, `organizations`, `projects`
- `servers`, `providers`, `credentials`
- `containers`, `volumes`, `storage`
- `alerts`, `incidents`, `scan_results`
- `automation_rules`, `audit_logs`
*(Dapat ditambahkan ekstensi PostGIS di masa mendatang jika membutuhkan data geolokasi datacenter).*

#### Redis (In-Memory Data Store & Queue)
Digunakan untuk operasi asynchronous dan data ephemeral:
- Message Queue / Job Queue untuk Worker
- Caching layer & Session management
- Rate Limiting
- Pub/Sub untuk distribusi notifikasi dan live updates

**Alur Pemrosesan Job Asynchronous:**
```text
User ──(Request Scan)──> Go API ──(Push Job)──> Redis Queue
                                                    │
PostgreSQL <──(Save Result)── Security Worker <─────┘
```

---

### 5.4. Object Storage (S3 Abstraction Layer)
Menggunakan pola adapter agar platform tidak terikat (*vendor lock-in*) pada satu penyedia cloud:

```text
ObjectStorage Interface
         │
    ┌────┼────────────┬─────────────┐
    │    │            │             │
 AWS S3  MinIO  Cloudflare R2  Lainnya (S3-Compatible)
```
- **Development / Local**: Menggunakan MinIO.
- **Production**: Menggunakan AWS S3, Cloudflare R2, atau penyedia S3-compatible lainnya.

---

### 5.5. Security Engine (Sentinel Architecture)
Sentinel dirancang dengan model modular berbasis worker terpisah:

```text
Sentinel Engine
       │
       ├── Scanner Workers (Port, TLS, HTTP, Dependency, Config, Container)
       │
       ├── Result Normalizer (Standard Finding Format)
       │
       ├── Risk Engine (Scoring & Severity Calculator)
       │
       └── Remediation Engine (Automated Patching / Mitigation)
```

**Struktur Data Temuan (Finding Format):**
```text
Finding:
├── id
├── severity       (CRITICAL | HIGH | MEDIUM | LOW | INFO)
├── category       (NETWORK | SSL | CONFIG | DEPENDENCY | CONTAINER)
├── resource_id
├── title
├── description
├── evidence
├── recommendation
└── status         (OPEN | IN_PROGRESS | RESOLVED | IGNORED)
```

---

### 5.6. Autentikasi & Otorisasi
- **Tahap Awal**: JWT / Session auth dengan dukungan OAuth 2.0 / OIDC.
- **Tahap Lanjutan (Caelus Identity)**: Provider identitas mandiri yang mendukung:
  - Email / Password dengan Argon2id hashing
  - Social Login (Google, GitHub)
  - Multi-Factor Authentication (MFA / TOTP)
  - API Keys & Service Account Tokens (RBAC granular)

---

### 5.7. Komunikasi Real-time & Observability
- **API Transport**: REST API untuk transaksi data standar.
- **Live Streams**: WebSocket / SSE (Server-Sent Events) untuk streaming deployment logs, live terminal, dan alert instan.
- **Metrics Engine**: Prometheus untuk agregasi time-series data.
- **Log Management**: Grafana Loki untuk log agregasi terpusat (aplikasi, Docker, sistem).

---

### 5.8. Infrastructure as Code (IaC) & Konfigurasi Deklaratif
Mendukung konfigurasi infrastruktur dalam format deklaratif (YAML):

```yaml
server:
  name: production-api
  cpu: 2
  memory: 4GB

storage:
  type: s3
  bucket: production-backup
  region: us-east-1
```

**Alur Rekonsiliasi State:**
```text
Desired State (YAML) ──> State Comparison ──> Execution Plan ──> Apply & Verify
```

---

## 6. Ringkasan Arsitektur Stack

```text
┌────────────────────────────────────────────────────────┐
│                      CAELUS CLOUD                      │
├─────────────────────────┬──────────────────────────────┤
│ Frontend                │ Next.js, TypeScript, Tailwind│
│                         │ Shadcn/ui, Zustand           │
├─────────────────────────┼──────────────────────────────┤
│ Backend API             │ Go + Chi Router              │
├─────────────────────────┼──────────────────────────────┤
│ Workers & Agent         │ Go (caelus-worker, agent)    │
├─────────────────────────┼──────────────────────────────┤
│ Database                │ PostgreSQL                   │
├─────────────────────────┼──────────────────────────────┤
│ Cache & Message Queue   │ Redis                        │
├─────────────────────────┼──────────────────────────────┤
│ Object Storage          │ S3-compatible API (MinIO/S3) │
├─────────────────────────┼──────────────────────────────┤
│ Monitoring & Metrics    │ Prometheus                   │
├─────────────────────────┼──────────────────────────────┤
│ Log Aggregation         │ Grafana Loki + OpenTelemetry │
├─────────────────────────┼──────────────────────────────┤
│ Security Subsystem      │ Sentinel Modular Scanners    │
├─────────────────────────┼──────────────────────────────┤
│ Real-time Communication │ REST API + WebSocket / SSE   │
├─────────────────────────┼──────────────────────────────┤
│ Container Runtime       │ Docker & Docker Compose      │
└─────────────────────────┴──────────────────────────────┘
```

---

## 7. Tahapan Eksekusi Bertahap (Phased Roadmap)

1. **Phase 1 — Control Plane**
   - Setup autentikasi & organisasi.
   - Manajemen entitas VPS & resource dashboard.
   - Pembuatan layer abstraksi provider (Mock Provider untuk MVP).

2. **Phase 2 — Monitoring & Telemetri**
   - Pengembangan `caelus-agent` ringan.
   - Pipeline pengumpulan metrik dan integrasi Prometheus/Loki.
   - Dashboard monitoring visual interaktif real-time.

3. **Phase 3 — Storage Management**
   - Implementasi adapter S3 (MinIO & AWS S3).
   - Pengelolaan bucket, file browser, dan generator Signed URL.
   - Pipeline backup otomatis ke object storage.

4. **Phase 4 — Automation & Event Engine**
   - Redis Queue & Task Scheduler terdistribusi.
   - Rule builder (Trigger - Condition - Action).
   - Notifikasi multi-channel (Webhook, Email).

5. **Phase 5 — Sentinel Security**
   - Modul scanner (Port, TLS, Header, Container, Dependency).
   - Result normalizer dan risk engine scoring (0 - 100).
   - Alur remediasi dan pelaporan insiden keamanan.

6. **Phase 6 — Orchestration & Multi-Provider**
   - Declarative IaC engine (Plan / Apply).
   - Integrasi langsung API provider publik (AWS, Hetzner, DigitalOcean).
   - Container orchestration dan deployment pipeline.