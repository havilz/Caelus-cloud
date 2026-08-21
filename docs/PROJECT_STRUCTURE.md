# Struktur Proyek & Arsitektur Sistem Caelus Cloud

Dokumen ini mendefinisikan struktur direktori lengkap, arsitektur backend, arsitektur frontend, skema basis data, dan relasi antar entitas untuk sistem Caelus Cloud.

---

## 1. Struktur Root Monorepo

```text
caelus_cloud/
├── .gitignore                          # Aturan pengecualian Git
├── .env.example                        # Template konfigurasi environment root
├── docker-compose.yml                  # Orkestrasi lokal (Postgres, Redis, MinIO, Prometheus, Loki, API, Worker)
│
├── frontend/                           # Frontend Dashboard Panel (Next.js + TypeScript)
├── backend/                            # Core Backend Services & Workers (Go)
├── agent/                              # Host Agent Daemon (Go)
│
├── deploy/                             # Konfigurasi deployment & containerization
│   ├── docker/                         # Dockerfile per service (API, Worker, Agent, UI)
│   └── observability/                  # Konfigurasi Prometheus, Loki, Grafana
│
├── changelog/                          # Catatan histori perubahan terstempel waktu (CHANGELOG.md)
├── graphify-out/                       # Living Knowledge Graph & Topology Metadata (graph.json, tree)
├── .agents/                            # Konfigurasi rules & workflows AI Agent
└── docs/                               # Dokumentasi Teknis Internal Proyek
    ├── PROJECT.md                      # Spesifikasi ruang lingkup & 5 domain
    ├── RULES.md                        # Standar kode & alur tata kelola
    ├── TASK.md                         # Task list & milestone tracking
    └── PROJECT_STRUCTURE.md            # Cetak biru arsitektur & struktur folder
```

---

## 2. Struktur Frontend (`frontend`)

Frontend bertindak sebagai klien visual dan panel kontrol murni (*Client-Side UI / API Consumer*).

```text
frontend/
├── public/                             # Aset statis (Logo, favicon, icons)
│
├── src/
│   ├── app/                            # Routing & Layout Panel (Next.js App Router)
│   │   ├── (auth)/                     # Rute Autentikasi Publik
│   │   │   ├── login/page.tsx
│   │   │   ├── register/page.tsx
│   │   │   └── layout.tsx              # Layout khusus halaman Auth (Centered card)
│   │   │
│   │   ├── (dashboard)/                # Rute Dashboard Terproteksi
│   │   │   ├── overview/page.tsx       # Tampilan ringkasan metrik & server
│   │   │   ├── infrastructure/
│   │   │   │   ├── vps/
│   │   │   │   │   ├── page.tsx        # Tabel daftar VPS
│   │   │   │   │   └── [id]/page.tsx   # Detail server & kontrol tombol (Reboot, Power, Resize)
│   │   │   │   ├── containers/page.tsx # Status & metrik Docker container lokal
│   │   │   │   ├── networks/page.tsx   # Pengaturan firewall & network
│   │   │   │   └── volumes/page.tsx    # Pengaturan storage volume
│   │   │   ├── storage/page.tsx        # S3 Object Storage bucket manager & file explorer
│   │   │   ├── monitoring/page.tsx     # Grafik utilisasi real-time & live logs viewer
│   │   │   ├── security/page.tsx       # Sentinel security score, scanner, & vulnerability list
│   │   │   ├── automation/page.tsx     # Antarmuka rule builder & scheduler
│   │   │   ├── settings/page.tsx       # Pengaturan organisasi, provider credential, & API keys
│   │   │   └── layout.tsx              # Master Layout Panel (Sidebar, Header, Breadcrumbs)
│   │   │
│   │   ├── globals.css                 # Konfigurasi Tailwind CSS & CSS Variables
│   │   └── layout.tsx                  # Root HTML Layout (Theme Provider, Query Client, Toaster)
│   │
│   ├── core/                           # Inti Token Desain, Tema, & Konstanta Global
│   │   └── theme/                      # Modul Sistem Desain Terpusat (Supabase Theme)
│   │       ├── app_colors.ts           # Token warna Supabase Deep Black & Emerald Green
│   │       ├── app_text.ts             # Hierarki & ukuran tipografi terstandarisasi
│   │       ├── app_containers.ts       # Batasan container, card padding, & modal scroll constraints
│   │       ├── app_theme.ts            # Konfigurasi mode tema (Dark / Light)
│   │       └── index.ts                # Barrel export (@/core/theme)
│   │
│   ├── components/                     # Komponen UI Reusable
│   │   ├── ui/                         # Primitif Design System (Button, Input, Card, Table, Dialog, Badge)
│   │   ├── layout/                     # Komponen layout (Sidebar, Header, Breadcrumbs)
│   │   └── server/                     # Komponen modul server (ServerStatusBadge, CreateServerModal, ResizeServerModal)
│   │
│   ├── features/                       # Modul Tampilan Per Fitur Dashboard (UI Components & Logic)
│   │   ├── auth/                       # LoginForm, RegisterForm, useAuth hook
│   │   ├── infrastructure/             # ServerListTable, ServerDetailCard, ActionButtons, AddServerModal
│   │   ├── storage/                    # BucketTable, FileExplorer, UploadModal, SignedUrlGenerator
│   │   ├── monitoring/                 # RealtimeCharts (CPU/RAM/Disk), LogsStreamViewer, AlertBadge
│   │   ├── security/                   # SecurityScoreWidget, FindingTable, ScanTriggerButton
│   │   └── automation/                 # RuleBuilderForm, TriggerPicker, ActionForm
│   │
│   ├── services/                       # HTTP API Client (Jembatan komunikasi ke Backend Go)
│   │   ├── api.ts                      # Axios instance terpusat (Base URL backend Go, Auth Bearer Interceptor)
│   │   ├── auth.service.ts             # Panggilan ke POST /api/v1/auth/login, register, refresh
│   │   ├── server.service.ts           # Panggilan ke GET /api/v1/servers, POST /api/v1/servers/:id/reboot
│   │   ├── provider.service.ts         # Panggilan ke GET /api/v1/providers
│   │   ├── storage.service.ts          # Panggilan ke GET /api/v1/storage/buckets, objects
│   │   ├── monitoring.service.ts       # Panggilan ke endpoint metrik & WebSocket stream
│   │   └── security.service.ts         # Panggilan ke GET /api/v1/sentinel/findings, scan
│   │
│   ├── stores/                         # State Management Browser (Zustand)
│   │   ├── useAuthStore.ts             # Menyimpan data session user & token login aktif
│   │   └── useServerStore.ts           # State cache daftar server & filter aktif
│   │
│   ├── types/                          # Definisi Tipe TypeScript (DTO response dari API Go)
│   │   ├── api.ts                      # Standard ApiResponse<T>, ErrorResponse
│   │   ├── auth.ts                     # UserProfile, LoginCredentials, AuthTokens
│   │   ├── server.ts                   # Server, Provider, CreateServerPayload, ResizeServerPayload
│   │   ├── storage.ts                  # BucketItem, StorageObjectItem
│   │   ├── monitoring.ts               # MetricSeries, LogMessage
│   │   └── security.ts                 # SentinelFinding, SecurityReport
│   │
│   └── lib/                            # Helper & Utilities
│       ├── utils.ts                    # Utility `cn` (clsx + tailwind-merge)
│       └── formatters.ts               # Formatter byte size, durasi uptime, tanggal
│
├── .env.example                        # Template environment variabel frontend (NEXT_PUBLIC_API_URL)
├── .env.local                          # Konfigurasi environment aktif lokal (git-ignored)
├── package.json
└── tsconfig.json
```

---

## 3. Struktur Backend (`backend/` & `agent/`)

Backend dibangun menggunakan bahasa **Go** dengan mematuhi prinsip **Clean Architecture** secara ketat.

### 3.1 Aturan Dependensi (The Dependency Rule)

Sesuai aturan di [docs/RULES.md](file:///c:/project/caelus_cloud/docs/RULES.md), seluruh dependensi kode **wajib selalu mengarah ke dalam**:

```text
+---------------------------------------------------------------------------------+
|                                 FRAMEWORKS & DRIVERS                            |
|       (cmd/api, cmd/worker, Chi Router, PostgreSQL Driver, AWS SDK, Redis)      |
|                                                                                 |
|        +---------------------------------------------------------------+        |
|        |                       INTERFACE ADAPTERS                      |        |
|        |   (delivery/http/v1, delivery/ws, repository/postgres,        |        |
|        |    provider/aws, storage/s3, delivery/worker)                 |        |
|        |                                                               |        |
|        |        +---------------------------------------------+        |        |
|        |        |             APPLICATION BUSINESS RULES      |        |        |
|        |        |                 (internal/usecase/*)        |        |        |
|        |        |                                             |        |        |
|        |        |        +---------------------------+        |        |        |
|        |        |        |   ENTERPRISE DOMAIN LOGIC |        |        |        |
|        |        |        |     (internal/domain/*)   |        |        |        |
|        |        |        +---------------------------+        |        |        |
|        |        +---------------------------------------------+        |        |
|        +---------------------------------------------------------------+        |
+---------------------------------------------------------------------------------+
```

#### Batasan Impor Paket (Package Import Boundaries):
1. **Layer 1 - `internal/domain` (Inti Paling Dalam)**:
   - **Bebas dari Dependensi Eksternal**: Berisi entitas murni (*structs*), tipe data domain, konstanta, dan kontrak antarmuka (*repository interfaces*, *driver interfaces*, *adapter interfaces*).
   - **Aturan**: DILARANG mengimpor paket dari `usecase`, `repository`, `delivery`, `database/sql`, `pgx`, framework HTTP `chi`, maupun SDK eksternal pihak ketiga.

2. **Layer 2 - `internal/usecase` (Logika Aplikasi)**:
   - **Hanya Bergantung pada Domain**: Berisi orkestrasi alur kerja aplikasi (*interactors*). Berinteraksi dengan basis data dan layanan eksternal HANYA melalui interface yang dideklarasikan di `internal/domain`.
   - **Aturan**: DILARANG mengimpor paket dari `repository/postgres`, `delivery/http`, framework HTTP, atau driver database.

3. **Layer 3 - `internal/repository`, `internal/provider`, `internal/storage`, `internal/delivery` (Adapter)**:
   - **Implementasi Kontrak Interface**:
     - `repository/postgres/` mengimplementasikan interface domain (`UserRepository`, `ServerRepository`, dll.) menggunakan SQL/pgx.
     - `provider/*` dan `storage/*` mengimplementasikan driver interface domain.
     - `delivery/http/` mengonversi request HTTP menjadi pemanggilan use case dan mengembalikan format JSON standar.
   - **Aturan**: Bergantung pada `domain` dan `usecase`. Tidak boleh membocorkan detail database/framework ke layer usecase.

4. **Layer 4 - `cmd/api` & `cmd/worker` (Composition Root / Entry Points)**:
   - Bertanggung jawab melakukan inisialisasi koneksi database, instansiasi repository, injeksi dependensi (*Dependency Injection*) ke usecase, dan menghubungkan usecase ke HTTP handler / worker sebelum server dijalankan.


```text
backend/
├── cmd/                                # Application Entry Points
│   ├── api/                            # HTTP REST API Server
│   │   └── main.go
│   └── worker/                         # Background Jobs, Queue Consumer, & Task Scheduler
│       └── main.go
│
├── internal/                           # Private Application & Business Logic
│   ├── domain/                         # Layer 1: Domain Entities & Interfaces (Pure Go, No External Dependencies)
│   │   ├── user.go                     # Entity User, Organization, Role, & Repository Interface
│   │   ├── server.go                   # Entity Server, ServerStatus, & ServerRepository Interface
│   │   ├── provider.go                 # Entity Provider, Credential, & ProviderDriver Interface
│   │   ├── storage.go                  # Entity Bucket, StorageObject, & StorageAdapter Interface
│   │   ├── metric.go                   # Entity MetricPoint, LogEntry, & MetricRepository Interface
│   │   ├── security.go                 # Entity Finding, ScanSession, Severity, & SecurityRepository Interface
│   │   ├── automation.go               # Entity AutomationRule, ExecutionLog, & AutomationRepository Interface
│   │   ├── audit.go                    # Entity AuditLog & AuditRepository Interface
│   │   └── errors.go                   # Standar Domain Error Types (ErrNotFound, ErrUnauthorized, dll.)
│   │
│   ├── usecase/                        # Layer 2: Application Business Rules (Use Cases / Interactors)
│   │   ├── auth/                       # Usecase: Register, Login, RefreshToken, VerifySession
│   │   ├── server/                     # Usecase: CreateServer, ListServers, GetServerDetail, RebootServer, ShutdownServer
│   │   ├── provider/                   # Usecase: RegisterProvider, ValidateCredentials
│   │   ├── storage/                    # Usecase: ListBuckets, UploadObject, GenerateSignedURL
│   │   ├── monitoring/                 # Usecase: IngestMetrics, QueryMetricsHistory, EvaluateAlerts
│   │   ├── security/                   # Usecase: TriggerScan, ProcessFindings, CalculateRiskScore
│   │   └── automation/                 # Usecase: RegisterRule, EvaluateTrigger, DispatchAction
│   │
│   ├── repository/                     # Layer 3: Interface Adapters - Database Repositories (PostgreSQL)
│   │   ├── postgres/
│   │   │   ├── client.go               # Database connection pool (pgx / database/sql)
│   │   │   ├── tx_manager.go           # Transaction manager implementation
│   │   │   ├── user_repo.go            # Implementasi UserRepository
│   │   │   ├── server_repo.go          # Implementasi ServerRepository
│   │   │   ├── provider_repo.go        # Implementasi ProviderRepository
│   │   │   ├── storage_repo.go         # Implementasi StorageRepository
│   │   │   ├── metric_repo.go          # Implementasi MetricRepository
│   │   │   ├── security_repo.go        # Implementasi SecurityRepository
│   │   │   ├── automation_repo.go      # Implementasi AutomationRepository
│   │   │   └── audit_repo.go           # Implementasi AuditRepository
│   │
│   ├── provider/                       # Layer 3: Infrastructure Provider Drivers (Abstraksi Multi-Provider)
│   │   ├── provider.go                 # Factory & Registry Provider Driver
│   │   ├── mock/                       # Mock Driver untuk simulasi lokal (MVP V1)
│   │   │   └── mock_driver.go
│   │   ├── aws/                        # AWS EC2 Driver
│   │   │   └── aws_driver.go
│   │   └── hetzner/                    # Hetzner Cloud Driver
│   │       └── hetzner_driver.go
│   │
│   ├── storage/                        # Layer 3: Object Storage Adapters (S3 Compatible)
│   │   ├── s3_adapter.go               # AWS S3 / MinIO Adapter implementation
│   │   └── r2_adapter.go               # Cloudflare R2 Adapter implementation
│   │
│   ├── delivery/                       # Layer 4: Primary Adapters (HTTP Transport, WebSocket, Queue Handlers)
│   │   ├── http/
│   │   │   ├── router.go               # Inisialisasi Chi Router & Route Mapping
│   │   │   ├── middleware/             # HTTP Middlewares
│   │   │   │   ├── auth_middleware.go  # Validasi JWT Access Token
│   │   │   │   ├── rbac_middleware.go  # Validasi Role & Permission
│   │   │   │   ├── audit_middleware.go # Pencatatan otomatis ke audit_logs
│   │   │   │   ├── rate_limit.go       # Redis Rate Limiter
│   │   │   │   └── logger.go           # HTTP Request/Response logger
│   │   │   ├── v1/                     # REST API v1 Handlers & Request/Response DTO
│   │   │   │   ├── auth_handler.go
│   │   │   │   ├── server_handler.go
│   │   │   │   ├── provider_handler.go
│   │   │   │   ├── storage_handler.go
│   │   │   │   ├── monitoring_handler.go
│   │   │   │   ├── security_handler.go
│   │   │   │   └── automation_handler.go
│   │   │   └── response/               # Standard JSON Response Formatter (Success, Error, Pagination)
│   │   │       └── response.go
│   │   │
│   │   ├── ws/                         # Real-time WebSockets / SSE Hub
│   │   │   ├── hub.go                  # Connection manager & client broadcast
│   │   │   └── client.go               # WebSocket client connection handler
│   │   │
│   │   └── worker/                     # Asynchronous Job Consumers (Redis Queue)
│   │       ├── queue.go                # Redis queue client
│   │       ├── backup_consumer.go      # Consumer tugas backup berkala
│   │       ├── scan_consumer.go        # Consumer pemindaian security scanner
│   │       └── notification_consumer.go# Consumer pengiriman email / webhook
│   │
│   └── security/                       # Sentinel Security Scanner Modules
│       ├── scanner.go                  # Scanner engine interface
│       ├── port_scanner.go             # TCP/UDP Port exposure scanner
│       ├── tls_scanner.go              # SSL/TLS Certificate & cipher validator
│       ├── header_scanner.go           # HTTP Security headers scanner
│       └── risk_engine.go              # Kalkulator skor keamanan & normalizer
│
├── pkg/                                # Reusable Public Utilities (Cross-Cutting Concerns)
│   ├── config/                         # Environment & YAML Configuration Loader
│   ├── logger/                         # Structured Logger (Zap / Zerolog)
│   ├── hasher/                         # Password hashing (Argon2id / bcrypt)
│   ├── jwt/                            # JWT Token Generator & Validator
│   ├── validator/                      # Request payload validator
│   └── pagination/                     # Helper pagination query & metadata
│
├── migrations/                         # Skrip Migrasi SQL Basis Data (DDL)
│   ├── 000001_init_schema.up.sql
│   ├── 000001_init_schema.down.sql
│   ├── 000002_enable_rls.up.sql
│   ├── 000002_enable_rls.down.sql
│   ├── 000003_fix_function_search_path.up.sql
│   └── 000003_fix_function_search_path.down.sql
│
├── tests/                              # Suite Pengujian Terpusat (Unit & Integration Tests)
│   ├── auth_usecase_test.go            # Pengujian logika usecase autentikasi & registrasi
│   ├── hasher_test.go                  # Pengujian algoritma hashing password Argon2id
│   └── router_test.go                  # Pengujian rute & middleware router HTTP
│
├── go.mod
├── go.sum
└── config.yaml.example

# Direktori Agent Terpisah (Host Agent Daemon)
agent/
├── cmd/
│   └── main.go                         # Entry point daemon caelus-agent (lifecycle, scheduler, graceful shutdown)
├── internal/
│   ├── collector/                      # Pengumpul metrik host lokal
│   │   ├── collector.go                # Interface MetricCollector
│   │   ├── system_linux.go             # Pembacaan procfs (/proc/stat, /proc/meminfo, statfs, /proc/net/dev)
│   │   └── system_fallback.go          # Fallback metrik sistem non-Linux
│   ├── config/                         # Manajemen konfigurasi environment
│   │   └── config.go
│   ├── docker/                         # Inspektor Docker daemon via Unix socket (/var/run/docker.sock)
│   │   └── inspector.go
│   └── transport/                      # Pengiriman data telemetri ke Caelus API
│       ├── client.go                   # HTTP/HTTPS transport client dengan Bearer auth & retry
│       └── payload.go                  # Definisi struktur data payload metrik
├── pkg/
│   └── logger/                         # Structured logger (log/slog)
│       └── logger.go
├── tests/                              # Suite pengujian unit & integrasi terpusat
│   ├── collector_test.go
│   ├── config_test.go
│   ├── docker_test.go
│   └── transport_test.go
├── config.yaml.example                 # Template konfigurasi agent
├── README.md                           # Dokumentasi operasional agent
├── go.mod
└── go.sum
```

---

## 4. Skema Basis Data & Relasi Entitas (PostgreSQL)

### 4.1 Diagram Relasi Entitas (Entity-Relationship Diagram)

```text
  +------------------+         1:N         +------------------------+
  |      users       | <------------------ |  organization_members  |
  +------------------+                     +------------------------+
           | 1:N                                       | N:1
           v                                           v
  +------------------+         1:N         +------------------------+
  |    audit_logs    | <------------------ |     organizations      |
  +------------------+                     +------------------------+
                                                       |
                         +-----------------------------+-----------------------------+
                         | 1:N                         | 1:N                         | 1:N
                         v                             v                             v
               +-------------------+         +-------------------+         +-------------------+
               |     providers     |         |  storage_buckets  |         | automation_rules  |
               +-------------------+         +-------------------+         +-------------------+
                         | 1:N                         |                             | 1:N
                         v                             |                             v
               +-------------------+                   |                   +-------------------+
               |      servers      |                   |                   |  auto_executions  |
               +-------------------+                   |                   +-------------------+
                 | 1:N          | 1:N                  |
                 v              v                      |
      +------------------+  +------------------+       |
      |  server_metrics  |  |  security_scans  |       |
      +------------------+  +------------------+       |
                                      | 1:N            |
                                      v                v
                            +--------------------+ +-------------------+
                            | security_findings  | |  backup_records   |
                            +--------------------+ +-------------------+
```

---

### 4.2 Definisi Tabel & Relasi Kolom

#### 1. Tabel `users`
Menyimpan data akun pengguna utama.
- `id` (UUID, Primary Key)
- `email` (VARCHAR(255), Unique, Not Null)
- `password_hash` (VARCHAR(255), Not Null)
- `full_name` (VARCHAR(100), Not Null)
- `is_active` (BOOLEAN, Default: true)
- `created_at` (TIMESTAMPTZ, Default: NOW())
- `updated_at` (TIMESTAMPTZ, Default: NOW())

#### 2. Tabel `organizations`
Mendukung multi-tenancy dan pengelolaan bersama dalam satu workspace.
- `id` (UUID, Primary Key)
- `name` (VARCHAR(100), Not Null)
- `slug` (VARCHAR(100), Unique, Not Null)
- `created_at` (TIMESTAMPTZ, Default: NOW())
- `updated_at` (TIMESTAMPTZ, Default: NOW())

#### 3. Tabel `organization_members`
Menghubungkan user ke organisasi dengan hak akses berbasis peran (RBAC).
- `id` (UUID, Primary Key)
- `org_id` (UUID, Foreign Key -> `organizations.id` ON DELETE CASCADE)
- `user_id` (UUID, Foreign Key -> `users.id` ON DELETE CASCADE)
- `role` (VARCHAR(20), Not Null: `owner`, `admin`, `member`, `viewer`)
- `created_at` (TIMESTAMPTZ, Default: NOW())
- *Index*: UNIQUE(`org_id`, `user_id`)

#### 4. Tabel `providers`
Menyimpan konfigurasi integrasi penyedia cloud (Mock, AWS, Hetzner, dll.).
- `id` (UUID, Primary Key)
- `org_id` (UUID, Foreign Key -> `organizations.id` ON DELETE CASCADE)
- `name` (VARCHAR(100), Not Null)
- `provider_type` (VARCHAR(50), Not Null: `mock`, `aws`, `hetzner`, `digitalocean`, `custom_vps`)
- `credentials_encrypted` (TEXT, Not Null) - Menyimpan encrypted API Key / Secret / SSH Key
- `is_active` (BOOLEAN, Default: true)
- `created_at` (TIMESTAMPTZ, Default: NOW())
- `updated_at` (TIMESTAMPTZ, Default: NOW())

#### 5. Tabel `servers`
Menyimpan data instance server / VPS yang dikelola.
- `id` (UUID, Primary Key)
- `org_id` (UUID, Foreign Key -> `organizations.id` ON DELETE CASCADE)
- `provider_id` (UUID, Foreign Key -> `providers.id` ON DELETE SET NULL)
- `name` (VARCHAR(100), Not Null)
- `ip_address` (VARCHAR(45), Not Null)
- `status` (VARCHAR(30), Not Null: `running`, `stopped`, `rebooting`, `provisioning`, `error`)
- `os_info` (VARCHAR(100))
- `cpu_cores` (INT, Default: 1)
- `ram_mb` (INT, Default: 1024)
- `disk_gb` (INT, Default: 20)
- `agent_installed` (BOOLEAN, Default: false)
- `last_ping_at` (TIMESTAMPTZ)
- `created_at` (TIMESTAMPTZ, Default: NOW())
- `updated_at` (TIMESTAMPTZ, Default: NOW())

#### 6. Tabel `server_metrics`
Menyimpan data histori telemetri performa server (dioptimalkan dengan time-series index).
- `id` (BIGSERIAL, Primary Key)
- `server_id` (UUID, Foreign Key -> `servers.id` ON DELETE CASCADE)
- `cpu_usage_pct` (NUMERIC(5,2), Not Null)
- `ram_usage_mb` (INT, Not Null)
- `disk_usage_gb` (NUMERIC(6,2), Not Null)
- `network_in_kb` (BIGINT, Default: 0)
- `network_out_kb` (BIGINT, Default: 0)
- `recorded_at` (TIMESTAMPTZ, Not Null, Index)

#### 7. Tabel `storage_buckets`
Mencatat metadata bucket object storage yang terhubung.
- `id` (UUID, Primary Key)
- `org_id` (UUID, Foreign Key -> `organizations.id` ON DELETE CASCADE)
- `provider_id` (UUID, Foreign Key -> `providers.id` ON DELETE SET NULL)
- `bucket_name` (VARCHAR(100), Not Null)
- `region` (VARCHAR(50), Not Null)
- `endpoint` (VARCHAR(255))
- `created_at` (TIMESTAMPTZ, Default: NOW())

#### 8. Tabel `backup_records`
Mencatat histori dan status backup data atau snapshot server.
- `id` (UUID, Primary Key)
- `server_id` (UUID, Foreign Key -> `servers.id` ON DELETE CASCADE)
- `bucket_id` (UUID, Foreign Key -> `storage_buckets.id` ON DELETE SET NULL)
- `file_path` (VARCHAR(255), Not Null)
- `size_bytes` (BIGINT, Default: 0)
- `status` (VARCHAR(30), Not Null: `pending`, `in_progress`, `completed`, `failed`)
- `completed_at` (TIMESTAMPTZ)
- `created_at` (TIMESTAMPTZ, Default: NOW())

#### 9. Tabel `security_scans`
Mencatat sesi eksekusi pemindaian keamanan oleh Sentinel.
- `id` (UUID, Primary Key)
- `org_id` (UUID, Foreign Key -> `organizations.id` ON DELETE CASCADE)
- `server_id` (UUID, Foreign Key -> `servers.id` ON DELETE CASCADE)
- `scan_type` (VARCHAR(50), Not Null: `full`, `port`, `tls`, `headers`, `dependencies`)
- `status` (VARCHAR(30), Not Null: `running`, `completed`, `failed`)
- `score` (INT, Default: 0) - Nilai keamanan 0 hingga 100
- `executed_at` (TIMESTAMPTZ, Default: NOW())

#### 10. Tabel `security_findings`
Menyimpan temuan detail kerentanan dari setiap sesi scan.
- `id` (UUID, Primary Key)
- `scan_id` (UUID, Foreign Key -> `security_scans.id` ON DELETE CASCADE)
- `server_id` (UUID, Foreign Key -> `servers.id` ON DELETE CASCADE)
- `severity` (VARCHAR(20), Not Null: `critical`, `high`, `medium`, `low`, `info`)
- `category` (VARCHAR(50), Not Null: `network`, `tls`, `headers`, `config`, `dependency`)
- `title` (VARCHAR(255), Not Null)
- `description` (TEXT)
- `evidence` (TEXT)
- `recommendation` (TEXT)
- `status` (VARCHAR(30), Default: `open`, Not Null: `open`, `in_progress`, `resolved`, `ignored`)
- `created_at` (TIMESTAMPTZ, Default: NOW())

#### 11. Tabel `automation_rules`
Menyimpan konfigurasi otomasi event-driven.
- `id` (UUID, Primary Key)
- `org_id` (UUID, Foreign Key -> `organizations.id` ON DELETE CASCADE)
- `name` (VARCHAR(100), Not Null)
- `trigger_type` (VARCHAR(50), Not Null: `metric_threshold`, `schedule`, `security_incident`)
- `condition_json` (JSONB, Not Null) - Parameter evaluasi rule
- `action_json` (JSONB, Not Null) - Aksi eksekusi (Alert, Snapshot, Script)
- `is_active` (BOOLEAN, Default: true)
- `created_at` (TIMESTAMPTZ, Default: NOW())
- `updated_at` (TIMESTAMPTZ, Default: NOW())

#### 12. Tabel `automation_executions`
Mencatat riwayat eksekusi dari rule otomasi.
- `id` (UUID, Primary Key)
- `rule_id` (UUID, Foreign Key -> `automation_rules.id` ON DELETE CASCADE)
- `status` (VARCHAR(30), Not Null: `success`, `failed`, `running`)
- `execution_log` (TEXT)
- `executed_at` (TIMESTAMPTZ, Default: NOW())

#### 13. Tabel `audit_logs`
Mencatat seluruh aktivitas perubahan resource untuk kepatuhan dan audit keamanan.
- `id` (BIGSERIAL, Primary Key)
- `org_id` (UUID, Foreign Key -> `organizations.id` ON DELETE CASCADE)
- `user_id` (UUID, Foreign Key -> `users.id` ON DELETE SET NULL)
- `action` (VARCHAR(100), Not Null) - Contoh: `server.reboot`, `auth.login`, `rule.create`
- `resource_type` (VARCHAR(50), Not Null)
- `resource_id` (VARCHAR(100))
- `ip_address` (VARCHAR(45))
- `user_agent` (TEXT)
- `payload_json` (JSONB)
- `created_at` (TIMESTAMPTZ, Default: NOW(), Index)