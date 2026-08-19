# Task List & Milestone Pengerjaan Caelus Cloud

Dokumen ini berfungsi sebagai pelacak tugas (*task tracking*) modular yang dikelompokkan berdasarkan lapisan arsitektur (*Architecture Layer Slicing*) untuk memonitor progres implementasi proyek Caelus Cloud dari Phase 1 hingga Phase 6.

Status Legend:
- [ ] Belum Dikerjakan (Pending)
- [-] Sedang Dikerjakan (In Progress)
- [x] Selesai (Completed)

Kategori Layer Slicing:
- [INFRA] : Infrastruktur, Docker, Container, Konfigurasi Environment, Observability Engine.
- [DATABASE] : Skema Basis Data Relasional, Migrasi SQL, Connection Pool, Repository Layer.
- [BACKEND] : Core API Service (Go + Chi), Business Logic / Use Cases, Middleware, Security / Auth.
- [FRONTEND] : User Interface (Next.js, Tailwind CSS, Shadcn/ui, Zustand), State Management.
- [AGENT] : Host Agent Binary (Go) di VPS pengguna.
- [WORKER] : Background Task Engine, Redis Queue, Job Scheduler, Automation Worker.
- [SECURITY] : Sentinel Security Subsystem (Scanners, Finding Normalizer, Risk Engine).

---

## Phase 1: Control Plane & Core MVP

### 1.1 Inisialisasi & Lingkungan Pengembangan
- [x] [INFRA] Setup struktur direktori monorepo / multi-module (backend Go, frontend Next.js, docker-compose).
- [x] [INFRA] Konfigurasi environment (`.env.example`, Docker Compose untuk PostgreSQL dan Redis).
- [x] [BACKEND] Inisialisasi framework HTTP backend (Go + Chi) dengan struktur Clean Architecture.
- [x] [BACKEND] Setup library logging terstruktur (structured logging) dan error handling standar.
- [x] [FRONTEND] Inisialisasi project frontend (Next.js App Router, TypeScript, Tailwind CSS, Shadcn/ui, Zustand).

### 1.2 Basis Data & Data Modeling
- [x] [DATABASE] Perancangan skema relasional tabel identitas dan akses:
  - [x] Tabel `users`
  - [x] Tabel `organizations`
  - [x] Tabel `organization_members`
- [x] [DATABASE] Perancangan skema relasional tabel server dan provider:
  - [x] Tabel `providers`
  - [x] Tabel `credentials`
  - [x] Tabel `servers`
- [x] [DATABASE] Perancangan skema tabel pencatatan aktivitas (`audit_logs`).
- [x] [DATABASE] Pembuatan skrip migrasi SQL (DDL Up & Down).
- [x] [DATABASE] Implementasi database connection pool dan transactional repository interface.

### 1.3 Modul Autentikasi & Otorisasi
- [x] [BACKEND] Implementasi use case registrasi dan login (hashing password dengan Argon2id / bcrypt).
- [x] [BACKEND] Implementasi manajemen token otentikasi JWT (Access Token & Refresh Token).
- [x] [BACKEND] Implementasi middleware autentikasi dan Role-Based Access Control (RBAC).
- [x] [BACKEND] Implementasi interceptor audit logging untuk setiap request terautentikasi.

### 1.4 Layer Abstraksi Provider
- [ ] [BACKEND] Definisi interface Provider Lifecycle (`CreateServer`, `GetServer`, `ListServers`, `RebootServer`, `ShutdownServer`, `ResizeServer`, `DeleteServer`).
- [ ] [BACKEND] Implementasi `MockProvider` untuk simulasi lifecycle VPS dan testing end-to-end tanpa dependensi cloud eksternal.
- [ ] [BACKEND] Use case dan repository manajemen kredensial provider.

### 1.5 Modul Manajemen Server & VPS
- [ ] [BACKEND] REST API endpoint untuk CRUD data server.
- [ ] [BACKEND] REST API endpoint untuk aksi kontrol server (Reboot, Shutdown, Power On).
- [ ] [BACKEND] Validasi input request (request DTO) dan standarisasi response payload.
- [ ] [BACKEND] Unit test dan integration test untuk modul server management.

### 1.6 Frontend Control Panel (Dashboard MVP)
- [ ] [FRONTEND] Implementasi halaman Autentikasi (Login & Register form dengan validasi skema).
- [ ] [FRONTEND] Setup layout Dashboard utama (Sidebar navigasi, Header, Breadcrumbs, Theme Switcher).
- [ ] [FRONTEND] Halaman Overview Dashboard (Kartu metrik agregat, status server aktif/nonaktif).
- [ ] [FRONTEND] Halaman Server Management:
  - [ ] Tabel daftar server dengan status badge, IP address, dan action menu.
  - [ ] Form / Modal penambahan server baru dan pemilihan provider.
  - [ ] Halaman detail server (Spesifikasi CPU/RAM/Disk, Uptime, Tombol aksi Reboot/Shutdown).

---

## Phase 2: Monitoring & Telemetri

### 2.1 Caelus Agent (Binary Host)
- [ ] [AGENT] Inisialisasi arsitektur binary `caelus-agent` (Go).
- [ ] [AGENT] Modul pengumpul metrik sistem lokal (CPU, RAM, Disk usage, Network I/O, Uptime).
- [ ] [AGENT] Modul inspeksi lokal Docker daemon (status container dan resource per container).
- [ ] [AGENT] Klien pengiriman metrik aman ke backend (HTTPS / mTLS payload transport).

### 2.2 Observability & Ingestion Engine
- [ ] [INFRA] Setup service Prometheus dalam Docker Compose untuk penyimpanan time-series.
- [ ] [INFRA] Setup service Grafana Loki dalam Docker Compose untuk agregasi log sistem.
- [ ] [BACKEND] Endpoint API penerimaan payload metrik telemetri dari agent.
- [ ] [BACKEND] Integrasi query adapter ke Prometheus dan Loki.
- [ ] [BACKEND] Implementasi WebSocket / SSE Hub untuk streaming data real-time ke client.
- [ ] [BACKEND] Engine evaluasi threshold alert (misal: CPU > 90% selama 5 menit).

### 2.3 Visualisasi Monitoring & Alerting
- [ ] [FRONTEND] Komponen chart interaktif (Grafik real-time CPU, RAM, Disk, Network throughput).
- [ ] [FRONTEND] Komponen Log Viewer real-time dengan kemampuan filtering berdasarkan level dan keyword.
- [ ] [FRONTEND] Halaman / Drawer Notifikasi Alert (Daftar alert aktif, aksi Acknowledge dan Resolve).

---

## Phase 3: Storage Management

### 3.1 Abstraksi Object Storage
- [ ] [INFRA] Setup MinIO service pada Docker Compose untuk lingkungan pengujian lokal.
- [ ] [BACKEND] Definisi interface Object Storage (`CreateBucket`, `ListBuckets`, `DeleteBucket`, `UploadObject`, `DownloadObject`, `DeleteObject`, `GenerateSignedURL`).
- [ ] [BACKEND] Implementasi adapter MinIO (S3-compatible API).
- [ ] [BACKEND] Implementasi adapter AWS S3 dan Cloudflare R2.

### 3.2 Manajemen Bucket & Objek
- [ ] [BACKEND] REST API untuk pengelolaan bucket (Create, List, Delete).
- [ ] [BACKEND] REST API untuk operasi file/objek dan pembuatan Signed URL.
- [ ] [FRONTEND] Antarmuka File Explorer / Object Browser (Navigasi folder, upload modal, download, delete).
- [ ] [FRONTEND] Modal generator Signed URL dengan pilihan waktu kadaluarsa (*expiration time*).

### 3.3 Pipeline Backup Otomatis
- [ ] [DATABASE] Skema tabel `backup_policies` dan `backup_records`.
- [ ] [BACKEND] Orchestrator snapshot server dan kompresi data backup.
- [ ] [WORKER] Worker pengunggahan file backup ke bucket object storage target.
- [ ] [WORKER] Worker eksekusi retensi backup (Lifecycle policy cleaner).
- [ ] [FRONTEND] Tampilan konfigurasi jadwal backup dan tabel riwayat backup server.

---

## Phase 4: Automation & Event Engine

### 4.1 Distributed Worker & Queue
- [ ] [INFRA] Setup Redis instance untuk Message Broker / Task Queue.
- [ ] [WORKER] Inisialisasi binary `caelus-worker` (Go).
- [ ] [WORKER] Implementasi Task Queue Engine (Task Producer & Consumer dengan retry mechanism).
- [ ] [WORKER] Implementasi Distributed Task Scheduler untuk pekerjaan berbasis jadwal (Cron-like).

### 4.2 Automation Rule Engine
- [ ] [DATABASE] Skema tabel `automation_rules` dan `automation_execution_logs`.
- [ ] [BACKEND] Event Dispatcher terpusat untuk mendistribusikan event sistem ke Rule Engine.
- [ ] [BACKEND] Rule Evaluation Engine (Validasi Trigger -> Evaluasi Condition -> Eksekusi Action).
- [ ] [BACKEND] REST API untuk CRUD aturan otomasi dan riwayat eksekusi.
- [ ] [FRONTEND] Antarmuka Visual Rule Builder (Form Trigger, Condition operator, Action selector).
- [ ] [FRONTEND] Tampilan log audit eksekusi otomasi.

### 4.3 Dispatcher Notifikasi
- [ ] [BACKEND] Adapter notifikasi Webhook eksternal.
- [ ] [BACKEND] Adapter notifikasi Email via SMTP.

---

## Phase 5: Sentinel Security Subsystem

### 5.1 Modular Scanner Workers
- [ ] [SECURITY] Port & Service Exposure Scanner worker.
- [ ] [SECURITY] TLS/SSL Certificate & Cipher Suite Validator worker.
- [ ] [SECURITY] HTTP/HTTPS Security Headers Auditor worker.
- [ ] [SECURITY] Host & Container Configuration Auditor worker.
- [ ] [SECURITY] System Dependency Vulnerability Scanner worker.

### 5.2 Sentinel Core & Risk Engine
- [ ] [DATABASE] Skema tabel `security_findings`, `security_scans`, dan `security_incidents`.
- [ ] [SECURITY] Finding Normalizer pipeline (Standarisasi temuan ke schema `Finding`).
- [ ] [SECURITY] Risk Engine (Perhitungan Security Score 0 - 100 dan kalkulasi severity).
- [ ] [BACKEND] REST API untuk eksekusi scan, pelaporan finding, dan metrik skor keamanan.

### 5.3 Sentinel Security Dashboard
- [ ] [FRONTEND] Dashboard Security Overview (Security score badge, distribusi keparahan: Critical, High, Medium, Low).
- [ ] [FRONTEND] Tabel daftar temuan keamanan (Detail temuan, evidence, rekomendasi perbaikan).
- [ ] [FRONTEND] Tombol trigger security scan manual dan indikator progres pemindaian.
- [ ] [FRONTEND] Tampilan manajemen insiden keamanan dan status remediasi.

---

## Phase 6: Orchestration & Multi-Provider

### 6.1 Multi-Provider Integration
- [ ] [BACKEND] Implementasi adapter provider AWS EC2.
- [ ] [BACKEND] Implementasi adapter provider VPS publik (Hetzner, DigitalOcean, Contabo).
- [ ] [BACKEND] Sinkronisasi status resource berkala antar provider pihak ketiga.
- [ ] [FRONTEND] Antarmuka manajemen multi-provider credentials.

### 6.2 Declarative Infrastructure as Code (IaC)
- [ ] [DATABASE] Skema tabel `iac_configurations`, `iac_plans`, dan `iac_states`.
- [ ] [BACKEND] Parser spesifikasi deklaratif infrastruktur berbasis YAML.
- [ ] [BACKEND] Engine pembanding state (Desired State vs Actual State / Plan Engine).
- [ ] [BACKEND] Engine eksekusi konfigurasi (Apply Engine) dengan mekanisme rollback.
- [ ] [FRONTEND] Editor konfigurasi deklaratif dengan validasi sintaks.
- [ ] [FRONTEND] Visual Diff Viewer untuk perbandingan Plan vs Actual state sebelum Apply.

### 6.3 Deployment & Container Orchestration
- [ ] [BACKEND] Pipeline deployment aplikasi berbasis Docker container.
- [ ] [FRONTEND] Web-based Terminal streaming untuk pemantauan log deployment secara langsung.
