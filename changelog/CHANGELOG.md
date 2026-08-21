# Changelog

Semua catatan perubahan pada proyek Caelus Cloud didokumentasikan di sini secara berkala.

Format penulisan mengacu pada standar formal dengan pencatatan stempel tanggal dan waktu `[YYYY-MM-DD HH:mm:ss]`.

---

## [Unreleased]

### [2026-08-21 19:25:00] - Sentinel Security Subsystem, Scanner Workers, Risk Engine & Security Hub (Phase 5)
- Eksekusi migrasi database SQL `000007_create_sentinel_security.up.sql` pada live PostgreSQL instance: membuat tabel `security_scans`, `security_findings`, dan `security_incidents` dengan Row Level Security (RLS) tenant isolation dan indeks performa multi-kolom.
- Pembuatan model domain `SecurityScan`, `SecurityFinding`, `SecurityIncident`, `SecurityPostureOverview`, `NormalizedFinding`, dan kontrak repositori `SecurityRepository` pada `backend/internal/domain/security.go`.
- Implementasi worker pemindai modular di `backend/internal/sentinel/scanner/`:
  - `port_scanner.go`: Pemindaian keterbukaan port TCP dan layanan berisiko tinggi (FTP, Telnet, MySQL, Postgres, Redis, MongoDB).
  - `tls_scanner.go`: Validasi masa berlaku sertifikat TLS/SSL, kesesuaian hostname, dan deteksi versi protokol lawas (TLS 1.0/1.1).
  - `headers_scanner.go`: Auditor header keamanan HTTP standar OWASP (HSTS, CSP, X-Frame-Options, X-Content-Type-Options, Referrer-Policy).
  - `host_config_scanner.go`: Audit hardening CIS host Linux, stabilitas memori/disk, dan deteksi stale uptime > 180 hari.
  - `vuln_scanner.go`: Auditor kerentanan CVE paket sistem operasi dan dependensi umum.
- Implementasi pipeline normalisasi temuan dengan sidik jari deduplikasi unik (fingerprinting) di `backend/internal/sentinel/normalizer.go`.
- Implementasi Risk Scoring Engine di `backend/internal/sentinel/risk_engine.go` yang menghitung skor postur keamanan (0 - 100) dan predikat huruf (Grade A/B/C/D/F).
- Implementasi Sentinel Orchestrator di `backend/internal/sentinel/orchestrator.go` yang mengoordinasikan pemindaian paralel konkuren dan memicu event sistem ke Rule Engine.
- Implementasi repositori PostgreSQL di `backend/internal/repository/postgres/security_repository.go` dan usecase di `backend/internal/usecase/security/security_usecase.go`.
- Pendaftaran endpoint REST API Sentinel di `backend/internal/delivery/http/v1/security_handler.go` (`/api/v1/security/overview`, `/api/v1/security/scans`, `/api/v1/security/findings`, `/api/v1/security/incidents`) dan integrasi composition root di `backend/cmd/api/main.go`.
- Pembuatan antarmuka TypeScript di `frontend/src/types/security.ts` dan client API service di `frontend/src/services/security.service.ts`.
- Pembuatan komponen visual `SecurityScoreBadge.tsx` dan modal inspeksi detail `FindingDetailModal.tsx` dengan fitur 1-click remediation command copy.
- Pembuatan halaman dasbor utama Sentinel Security Hub di `frontend/src/app/(dashboard)/security/page.tsx` dengan peluncur scan cepat, filter temuan interaktif, dan riwayat pemindaian.
- Pengujian unit backend (`security_scanner_test.go`, `security_risk_engine_test.go`) lulus 100% dan verifikasi TypeScript compiler Next.js berhasil dengan 0 error.

### [2026-08-21 19:05:00] - BYOS Server Onboarding, Auto-Sync Spesifikasi & Heartbeat Liveness Watchdog
- Implementasi flow onboarding Bring Your Own Server (BYOS) pada antarmuka pendaftaran server di `frontend/src/components/server/CreateServerModal.tsx` dengan penyederhanaan form (hanya nama server) dan tab pemilihan mode registrasi.
- Pembuatan komponen modal `frontend/src/components/server/ConnectAgentModal.tsx` untuk menampilkan instruksi instalasi 1-line script dan kredensial ingestion (`SERVER_ID` & `AGENT_SECRET`) dengan fitur 1-click salin.
- Penambahan tombol akses cepat petunjuk instalasi agent berikon terminal pada tabel manajemen server di `frontend/src/app/(dashboard)/infrastructure/vps/page.tsx`.
- Implementasi driver provider kustom di `backend/internal/provider/custom/custom_driver.go` dan registrasi slug `custom` pada `backend/internal/provider/factory.go`.
- Penambahan endpoint publik `GET /install.sh` di `backend/internal/delivery/http/router.go` untuk menyajikan script instalasi shell otomatis yang mengonfigurasi systemd unit `caelus-agent.service` di OS Linux.
- Pembaruan mekanisme ingestion telemetri pada `backend/internal/usecase/monitoring/monitoring_usecase.go` untuk menyinkronkan spesifikasi hardware asli (CPU Cores, RAM, Disk, OS, Hostname) dari `caelus-agent` ke database secara otomatis.
- Implementasi background worker `HeartbeatWatchdog` di `backend/internal/usecase/monitoring/watchdog.go` yang memantau liveness detak jantung seluruh server aktif, menandai server menjadi `stopped` jika tidak ada telemetri selama 15 detik, menyiarkan status via WebSocket, dan memicu event ke Rule Engine.
- Penambahan method `ListAllRunning` pada `domain.ServerRepository` dan implementasi PostgreSQL di `backend/internal/repository/postgres/server_repository.go`.
- Pengujian unit backend (`go test -v ./tests/...`) lulus 100% dan pengujian kompilasi Next.js production build (`pnpm build`) berhasil tanpa error.
- Pembaruan Knowledge Graph Graphify monorepo menjadi 1.272 nodes, 3.471 edges, dan 76 communities.

### [2026-08-21 01:40:00] - Mesin Otomasi, Antrean Tugas Terdistribusi & Notifikasi (Phase 4: Automation & Event Engine)
- Eksekusi migrasi database SQL `000006_create_automation_rules_and_logs.up.sql` pada live PostgreSQL instance: membuat tabel `automation_rules` dan `automation_execution_logs` dengan Row Level Security (RLS) tenant isolation dan indeks performa multi-kolom.
- Instalasi dependensi Redis driver `github.com/redis/go-redis/v9` pada backend Go.
- Implementasi arsitektur Task Queue Engine di `backend/internal/queue/engine.go` dengan dukungan penanganan payload asinkron, retry eksponensial (exponential backoff), dan Dead Letter Queue (DLQ).
- Implementasi Redis Task Queue Engine di `backend/internal/queue/redis/queue.go` berbasis Redis List (LPUSH/BRPOP) dan Sorted Sets (ZADD/ZRANGEBYSCORE) untuk delayed tasks.
- Implementasi thread-safe in-memory Mock Queue Engine di `backend/internal/queue/mock/queue.go` untuk pengujian unit terisolasi.
- Implementasi Distributed Task Scheduler di `backend/internal/queue/scheduler.go` untuk mengeksekusi pekerjaan berkala (cron-like routines) dan memasukkannya ke TaskQueue.
- Implementasi domain model `AutomationRule`, `RuleCondition`, `RuleAction`, `RuleExecutionLog`, `SystemEvent`, dan `AutomationRepository` di `backend/internal/domain/automation.go`.
- Implementasi repositori PostgreSQL di `backend/internal/repository/postgres/automation_repository.go` dengan query terpaginasi dan filter eksekusi.
- Implementasi Central Event Dispatcher di `backend/internal/automation/dispatcher.go` untuk menyebarkan kejadian sistem (`metric.*`, `server.*`, `backup.*`) ke berbagai subscriber secara non-blocking fan-out.
- Implementasi Mesin Evaluasi Aturan (Rule Engine) di `backend/internal/automation/engine.go` dengan evaluasi klausa kondisi multi-operator (`>`, `>=`, `<`, `<=`, `==`, `!=`, `in`, `contains`), pencegahan flapping dengan jendela cooldown, eksekusi aksi berantai, dan pencatatan audit log otomatis.
- Implementasi Notification Adapters: HTTP Webhook dengan HMAC-SHA256 signature di `backend/internal/notification/webhook/adapter.go`, SMTP Email dengan HTML template responsif di `backend/internal/notification/email/adapter.go`, dan `UnifiedDispatcher` di `backend/internal/notification/dispatcher.go`.
- Implementasi usecase otomasi di `backend/internal/usecase/automation/automation_usecase.go` dan REST HTTP handler di `backend/internal/delivery/http/v1/automation_handler.go`.
- Inisialisasi binary daemon terdistribusi `caelus-worker` di `backend/cmd/worker/main.go` yang menangani konsumsi antrean Redis, registrasi handler eksekusi aksi, pengiriman notifikasi, dan graceful shutdown listener.
- Pendaftaran rute API otomasi pada router Chi (`/api/v1/automation/rules`, `/api/v1/automation/rules/{id}`, `/api/v1/automation/rules/{id}/test`, `/api/v1/automation/logs`) dan integrasi composition root di `backend/cmd/api/main.go`.
- Pembuatan model tipe data TypeScript di `frontend/src/types/automation.ts` dan API service di `frontend/src/services/automation.service.ts`.
- Pembuatan antarmuka modal Visual Rule Builder di `frontend/src/components/automation/CreateRuleModal.tsx` dengan konfigurasi pemicu, klausul kondisi dinamis, dan aksi berantai.
- Pembuatan halaman dasbor utama Rule Manager di `frontend/src/app/(dashboard)/automation/page.tsx` dengan kartu metrik ringkasan, toggle on/off aturan live, uji coba eksekusi manual (*Test Run*), dan penghapusan aturan.
- Pembuatan halaman jejak audit log eksekusi di `frontend/src/app/(dashboard)/automation/logs/page.tsx` dengan filter status, lencana status hasil eksekusi, dan drawer inspeksi JSON payload evaluasi.
- Pengujian unit dan integrasi backend (`queue_engine_test.go`, `automation_engine_test.go`, `notification_dispatcher_test.go`, `automation_http_test.go`) dengan hasil seluruh pengujian lulus 100% (`PASS`).
- Pengujian kompilasi Next.js 16 App Router production build (`pnpm build`) dengan 13 rute berstatus sukses (`PASS`) tanpa error.
- Pembaruan Knowledge Graph Graphify monorepo menjadi 1.245 nodes, 3.393 edges, dan 77 communities.

### [2026-08-21 00:03:00] - Manajemen Bucket, File Explorer & Pipeline Backup Otomatis (Phase 3.2 & Phase 3.3)
- Eksekusi migrasi database SQL `000005_create_storage_and_backups.up.sql` pada live PostgreSQL instance: membuat tabel `buckets`, `backup_policies`, dan `backup_records` dengan Row Level Security (RLS) tenant isolation dan indeks performa tinggi.
- Pembuatan kontrak repository `BucketRepository` di `backend/internal/domain/storage.go` dan model domain `BackupPolicy`, `BackupRecord`, `BackupStatus`, `BackupRepository` di `backend/internal/domain/backup.go`.
- Implementasi repositori PostgreSQL `backend/internal/repository/postgres/bucket_repository.go` dan `backend/internal/repository/postgres/backup_repository.go`.
- Implementasi usecase manajemen siklus hidup bucket dan objek di `backend/internal/usecase/storage/storage_usecase.go`.
- Implementasi usecase orchestrator snapshot server, kompresi stream, upload storage, dan lifecycle retention cleaner di `backend/internal/usecase/backup/backup_usecase.go`.
- Implementasi background worker scheduler di `backend/internal/usecase/backup/backup_scheduler.go` untuk evaluasi cron rutin berkala (60 detik) dan pembersih arsip kedaluwarsa.
- Implementasi HTTP REST handlers `StorageHandler` di `backend/internal/delivery/http/v1/storage_handler.go` dan `BackupHandler` di `backend/internal/delivery/http/v1/backup_handler.go`.
- Registrasi endpoint storage dan backup pada router Chi serta integrasi composition root di `backend/cmd/api/main.go`.
- Pembuatan model tipe data TypeScript di `frontend/src/types/storage.ts` dan `frontend/src/types/backup.ts`.
- Implementasi API client services di `frontend/src/services/storage.service.ts` dan `frontend/src/services/backup.service.ts`.
- Pembuatan komponen interaktif: modal pembuatan bucket `CreateBucketModal.tsx`, modal upload berkas drag-and-drop `UploadObjectModal.tsx`, modal generator presigned URL `GenerateSignedUrlModal.tsx`, dan modal pembuatan jadwal backup `CreateBackupPolicyModal.tsx`.
- Pembuatan halaman dashboard utama Object Storage di `frontend/src/app/(dashboard)/storage/page.tsx`.
- Pembuatan antarmuka File Explorer & Object Browser interaktif di `frontend/src/app/(dashboard)/storage/[bucket]/page.tsx` dengan navigasi breadcrumb, folder virtual, unduh langsung, pembuatan signed URL, dan penghapusan file.
- Pembuatan halaman pusat Disaster Recovery & Backup di `frontend/src/app/(dashboard)/storage/backups/page.tsx` dengan manajemen kebijakan, eksekusi snapshot on-demand (*Run Now*), dan tabel arsip backup berstatus.
- Pengujian unit dan integrasi backend (`storage_usecase_test.go`, `backup_usecase_test.go`, `storage_http_test.go`) dengan hasil 40/40 test lulus 100% (`PASS`).
- Verifikasi kompilasi Next.js 16 App Router production build (`pnpm build`) berstatus sukses (`PASS`) tanpa error.
- Pembaruan Knowledge Graph Graphify monorepo menjadi 1.064 nodes, 2.851 edges, dan 62 communities.

### [2026-08-20 23:40:00] - Abstraksi Object Storage Multi-Provider (Phase 3.1)
- Penyempurnaan konfigurasi service MinIO pada `docker-compose.yml` dengan penambahan healthcheck endpoint otomatis (`/minio/health/live`), binding port 9000 & 9001, serta volume persisten `minio_data`.
- Pembuatan model entitas domain dan kontrak interface `ObjectStorageAdapter` dan `StorageFactory` di `backend/internal/domain/storage.go` (`CreateBucket`, `ListBuckets`, `DeleteBucket`, `BucketExists`, `ListObjects`, `UploadObject`, `DownloadObject`, `DeleteObject`, `DeleteObjects`, `GetObjectMetadata`, `GenerateSignedURL`).
- Implementasi adapter inti S3-compatible universal berbasis AWS SDK Go v2 (`github.com/aws/aws-sdk-go-v2/service/s3`) di `backend/internal/storage/s3/adapter.go` dengan dukungan path-style addressing, multipart uploads, MIME auto-detection, dan presigned URL generator.
- Implementasi adapter MinIO di `backend/internal/storage/minio/adapter.go`.
- Implementasi adapter Cloudflare R2 di `backend/internal/storage/r2/adapter.go`.
- Implementasi mock in-memory thread-safe adapter di `backend/internal/storage/mock/adapter.go` untuk pengujian unit terisolasi.
- Implementasi `StorageFactory` di `backend/internal/storage/factory.go` untuk registrasi dinamis dan resolusi adapter multi-provider.
- Pembuatan suite pengujian unit terpusat di `backend/tests/storage_adapter_test.go` yang menguji siklus hidup bucket, upload/download objek, direktori virtual/delimiter, metadata, batch delete, presigned URL, dan inisialisasi provider dengan hasil 100% kelulusan (`PASS`).
- Pembaruan Knowledge Graph Graphify monorepo menjadi 896 nodes, 2.298 edges, dan 52 communities.

### [2026-08-20 18:47:00] - Visualisasi Monitoring, Telemetri Real-Time & Alerting (Phase 2.3)
- Pembuatan model tipe data TypeScript telemetri metrik, insiden alert, aturan ambang batas, dan entri log di `frontend/src/types/monitoring.ts`.
- Implementasi API client service `monitoringService` di `frontend/src/services/monitoring.service.ts` untuk live metrics, query deret waktu history (1h, 6h, 24h, 7d), list alerts berpaginasi, aksi acknowledge/resolve, dan manajemen aturan threshold.
- Implementasi hook React WebSocket `useRealtimeTelemetry` di `frontend/src/hooks/useRealtimeTelemetry.ts` untuk streaming otomatis event `metrics.updated` dan `alert.created` per server/organisasi.
- Implementasi komponen grafik deret waktu interaktif SVG `MetricTimeSeriesChart` di `frontend/src/components/monitoring/MetricTimeSeriesChart.tsx` dengan gradient fill, neon glow, hover crosshair, floating tooltip dinamis, selektor durasi, dan panel statistik (Current, Avg, Min, Max Peak).
- Implementasi komponen `LogViewer` di `frontend/src/components/monitoring/LogViewer.tsx` dengan tampilan dark monospace terminal, filter level log instan (ALL, INFO, WARN, ERROR, DEBUG), live keyword search, auto-scroll toggle, dan fungsi copy/export file log.
- Implementasi komponen slide-over `AlertDrawer` di `frontend/src/components/monitoring/AlertDrawer.tsx` dengan tab status filter (Aktif, Ditandai, Selesai, Semua), badge severity, perbandingan nilai terukur vs threshold, dan aksi Acknowledge / Resolve.
- Implementasi modal `CreateAlertRuleModal` di `frontend/src/components/monitoring/CreateAlertRuleModal.tsx` untuk konfigurasi ambang batas alert custom.
- Integrasi tombol Notification Bell dengan badge hitungan alert aktif di `frontend/src/components/layout/Header.tsx`.
- Pembaruan antarmuka detail server VPS `frontend/src/app/(dashboard)/infrastructure/vps/[id]/page.tsx` dengan 4 tab komprehensif: Overview & Hardware, Live Telemetry & Metrics Charts, Console Logs, dan Alerts & Thresholds.
- Pembuatan halaman pusat pemantauan global `frontend/src/app/(dashboard)/monitoring/page.tsx` dengan kartu metrik agregat seluruh fleet server, tabel pusat alert organisasi, dan manajer aturan evaluasi threshold.
- Refaktorisasi dan standardisasi styling halaman `monitoring/page.tsx` dan `infrastructure/vps/[id]/page.tsx` menggunakan barrel theme tokens (`AppContainers`, `AppText`, `AppColors`) dari `@/core/theme`.
- Verifikasi kompilasi production build Next.js 16 App Router (`pnpm build`) berstatus sukses (`PASS`) tanpa error linting/tipe.
- Pembaruan Knowledge Graph Graphify monorepo menjadi 835 nodes, 2.154 edges, dan 49 communities.

### [2026-08-20 18:30:00] - Observability & Ingestion Engine Backend (Phase 2.2)
- Setup service time-series Prometheus v2.50.0 (port 9090) dan log aggregator Grafana Loki v2.9.4 (port 3100) pada konfigurasi `docker-compose.yml` beserta file konfigurasi `deploy/observability/prometheus/prometheus.yml` dan `deploy/observability/loki/loki-config.yml`.
- Pembuatan skema DDL migrasi SQL `backend/migrations/000004_create_metrics_and_alerts.up.sql` dan `down.sql` untuk tabel `server_metrics`, `alert_rules`, dan `alerts` dengan indexing performa tinggi dan Row Level Security (RLS).
- Implementasi domain model dan repository contract di `backend/internal/domain/` (`metric.go`, `alert.go`, `observability.go`).
- Implementasi repositori PostgreSQL di `backend/internal/repository/postgres/` (`metric_repository.go` dan `alert_repository.go`) dengan query agregasi dan riwayat time-series.
- Implementasi adapter query observabilitas di `backend/internal/observability/` untuk Prometheus (`client.go`) dan Grafana Loki (`client.go`).
- Implementasi broker pesan real-time terpusat WebSocket dan Server-Sent Events (SSE) di `backend/internal/delivery/ws/` (`hub.go` dan `handler.go`) dengan mekanisme channel fan-out per topic server dan organisasi.
- Implementasi engine evaluasi ambang batas alert di `backend/internal/usecase/monitoring/alert_evaluator.go` dengan pencegahan duplikasi insiden aktif dan siaran otomatis ke WebSocket Hub.
- Implementasi logika bisnis terpadu di `backend/internal/usecase/monitoring/monitoring_usecase.go` untuk ingestion telemetri, live metrics, history time-series, dan manajemen siklus hidup alert (Acknowledge, Resolve, Rules).
- Implementasi HTTP delivery handler di `backend/internal/delivery/http/v1/` (`telemetry_handler.go` dan `alert_handler.go`) beserta registrasi rute pada Chi router `router.go` dan composition root `main.go`.
- Pembuatan suite pengujian unit terpusat di `backend/tests/` (`alert_evaluator_test.go`, `ws_hub_test.go`, `monitoring_usecase_test.go`, `telemetry_handler_test.go`) dengan tingkat kelulusan 100% (`PASS`).
- Pembaruan Knowledge Graph Graphify monorepo menjadi 805 nodes, 2.035 edges, dan 40 communities.

### [2026-08-20 17:44:00] - Inisialisasi Caelus Agent Daemon & Monitoring Telemetri Host (Phase 2.1)
- Inisialisasi modul Go mandiri `github.com/havilz/caelus-cloud/agent` untuk binary daemon `caelus-agent`.
- Implementasi modul pengumpul metrik sistem lokal di `agent/internal/collector/` (pembacaan CPU delta ticks `/proc/stat`, memori `/proc/meminfo`, kapasitas disk `syscall.Statfs`, throughput jaringan `/proc/net/dev`, load average, dan uptime sistem).
- Implementasi inspektor Docker daemon native di `agent/internal/docker/` melalui komunikasi Unix domain socket (`/var/run/docker.sock`) tanpa dependensi SDK berat, dengan kalkulasi persentase CPU dan alokasi memori per container.
- Implementasi klien transport HTTPS aman di `agent/internal/transport/` dengan autentikasi header (`Authorization: Bearer` dan `X-Server-ID`), penanganan status kode server, dan mekanisme *exponential backoff retry*.
- Implementasi parser konfigurasi di `agent/internal/config/` serta structured logger berbasis `log/slog` di `agent/pkg/logger/`.
- Implementasi entry point daemon di `agent/cmd/main.go` dengan scheduler periodik (ticker) dan *graceful shutdown* berbasis penanganan sinyal OS (`SIGINT`/`SIGTERM`).
- Pembuatan suite unit test terpusat di `agent/tests/` (`config_test.go`, `collector_test.go`, `docker_test.go`, `transport_test.go`) dengan hasil kelulusan 100% (`PASS`).
- Pembaruan Knowledge Graph Graphify monorepo menjadi 663 nodes dan 1.620 edges di 35 communities.

### [2026-08-17 19:26:29] - Inisialisasi Arsitektur & Dokumentasi Inti
- Dokumentasi arsitektur sistem, spesifikasi tech stack 5 domain, dan alur roadmap di `docs/PROJECT.md`.
- Standar penulisan kode enterprise, aturan komentar fungsi, dan kebijakan anti-emoji di `docs/RULES.md`.
- Rincian breakdown tugas berfase dan checklist pelacakan di `docs/TASK.md`.
- Blueprint monorepo, arsitektur backend Go Clean Architecture, panel Next.js, dan ERD PostgreSQL 13 tabel di `docs/PROJECT_STRUCTURE.md`.
- Inisialisasi struktur direktori modular untuk `frontend/src`, `backend`, `agent`, dan `deploy`.
- Standarisasi penamaan direktori dan package frontend menjadi `frontend`.
- Dokumentasi strategi lingkungan development (Supabase free-tier, Upstash Redis, MockProvider) dan production di `docs/PROJECT.md`.

### [2026-08-17 19:26:29] - Konfigurasi Lingkungan & Deployment
- Pembuatan template variabel lingkungan `.env.example` di root dan `frontend/.env.example`.
- Pembuatan file orkestrasi container `docker-compose.yml` (PostgreSQL 16, Redis 7, MinIO, API, Worker, Frontend).
- Pembuatan file environment lokal `.env` dan `frontend/.env.local` dengan generate kunci kriptografi acak 256-bit untuk `JWT_SECRET` dan `ENCRYPTION_KEY`.

### [2026-08-17 19:26:29] - Inisialisasi Framework Backend HTTP
- Inisialisasi modul Go `github.com/havilz/caelus-cloud/backend` dengan router Chi v5, CORS middleware, dan structured logger (`log/slog`).
- Implementasi domain errors standar di `internal/domain/errors.go` dan response formatter terstandarisasi (`Success`, `Error`, `Paginated`).
- Implementasi endpoint health check (`/health` dan `/api/v1/health`) dengan graceful shutdown dan unit test otomatis (`PASS`).
- Standarisasi komentar fungsi teknis pada seluruh package backend sesuai aturan `docs/RULES.md`.

### [2026-08-17 19:26:29] - Inisialisasi Framework Frontend
- Instalasi dependensi inti frontend (`zustand`, `lucide-react`, `clsx`, `tailwind-merge`, `class-variance-authority`, `axios`).
- Pembuatan utilitas penggabung class Tailwind di `frontend/src/lib/utils.ts`.
- Verifikasi kompilasi production build Next.js 16 App Router (`pnpm build`) berstatus sukses (`PASS`).

### [2026-08-17 19:26:29] - Pemodelan Basis Data & Migrasi Supabase
- Integrasi connection pool driver PostgreSQL `pgx/v5` di `backend/internal/repository/postgres/client.go` dan verifikasi koneksi live database Supabase (Status: Connected).
- Perancangan dan pembuatan skrip DDL migrasi SQL (`000001_init_schema.up.sql` & `down.sql`) untuk 7 tabel inti (`users`, `organizations`, `organization_members`, `providers`, `credentials`, `servers`, `audit_logs`) beserta index, trigger `updated_at`, dan seed provider.
- Implementasi modul migrator basis data terintegrasi di `backend/internal/repository/postgres/migrator.go`.
- Implementasi CLI migrasi mandiri di `backend/cmd/migrate/main.go`.
- Implementasi entitas domain model dan interface repository Clean Architecture (`User`, `Organization`, `Provider`, `Server`, `AuditLog`) di `backend/internal/domain/`.

### [2026-08-17 19:26:29] - Pengerasan Keamanan Basis Data (Security Hardening)
- Pembuatan dan eksekusi skrip migrasi `000002_enable_rls.up.sql` untuk mengaktifkan Row Level Security (RLS) pada seluruh tabel publik guna mengamankan akses PostgREST Supabase.
- Pembuatan dan eksekusi skrip migrasi `000003_fix_function_search_path.up.sql` untuk menetapkan `SECURITY DEFINER SET search_path = ''` pada fungsi `update_updated_at_column` guna menyelesaikan security linter warning Supabase.

### [2026-08-17 20:10:34] - Modul Autentikasi: Hashing Argon2id & Use Case Auth
- Implementasi modul enkripsi password standar industri Argon2id di `backend/pkg/hasher/hasher.go` (64MB memori, 3 iterasi, 2 thread paralel) dengan verifikasi hash format crypt standar dan perbandingan waktu-konstan (`subtle.ConstantTimeCompare`).
- Implementasi repository PostgreSQL untuk User (`backend/internal/repository/postgres/user_repository.go`) dan Organization (`backend/internal/repository/postgres/organization_repository.go`).
- Implementasi use case registrasi (`Register`) dan login (`Login`) di `backend/internal/usecase/auth/auth_usecase.go` yang mencakup validasi email/password, pembuatan hash Argon2id, inisialisasi entitas User, pembuatan organisasi default, dan penetapan role Owner.

### [2026-08-17 20:16:52] - Restrukturisasi Modul Pengujian (Centralized Testing)
- Pembuatan direktori pengujian terpusat `backend/tests/` untuk memisahkan seluruh file test dari folder kode implementasi internal.
- Pemindahan seluruh file pengujian (`hasher_test.go`, `auth_usecase_test.go`, `router_test.go`) ke dalam `backend/tests/` dan pembersihan file test dari modul `pkg/` serta `internal/`.
- Pembaruan dokumentasi cetak biru struktur arsitektur pada `docs/PROJECT_STRUCTURE.md`.

### [2026-08-17 20:19:27] - Modul Autentikasi: Manajemen Token JWT (Access & Refresh Token)
- Implementasi manajer token JWT di `backend/pkg/jwt/jwt.go` menggunakan `golang-jwt/jwt/v5` dengan algoritma HMAC-SHA256, penanganan masa berlaku (Access Token 15 menit, Refresh Token 7 hari), serta ekstraksi klaim khusus (`UserID`, `Email`, `OrganizationID`, `TokenType`).
- Integrasi penerbitan `TokenPair` pada usecase registrasi (`Register`) dan login (`Login`), serta penambahan alur perpanjangan sesi (`RefreshToken`) pada `backend/internal/usecase/auth/auth_usecase.go`.
- Pembuatan unit test komprehensif di `backend/tests/jwt_test.go` dan pembaruan `backend/tests/auth_usecase_test.go` dengan hasil pengujian lulus 100% (`PASS`).

### [2026-08-17 20:49:19] - Refaktorisasi Kompleksitas Kognitif Kode (SonarQube Compliance)
- Dekomposisi fungsi `Register` dan `Login` pada `backend/internal/usecase/auth/auth_usecase.go` menjadi metode-metode privat terfokus (`validateRegisterInput`, `ensureEmailUnique`, `createUserEntity`, `createOrganizationWithMember`, `authenticateUser`, `resolveActiveOrganizationID`) guna menurunkan skor Cognitive Complexity (aturan go:S3776).
- Ekstraksi pembentukan token pada `backend/pkg/jwt/jwt.go` ke dalam fungsi pembantu `generateToken` dan `validateTokenWithType` untuk menjaga kompleksitas kode tetap rendah dan mudah dirawat.

### [2026-08-18 15:16:25] - Refaktorisasi Suite Pengujian Use Case Auth (SonarQube Compliance)
- Membedakan implementasi logika `Create` dan `Update` pada `mockOrgRepo` dan `mockUserRepo` di `backend/tests/auth_usecase_test.go` untuk mengatasi issue duplikasi fungsi identik.
- Memecah pengujian monolitik `TestRegisterUsecase` dan `TestLoginUsecase` menjadi fungsi pengujian individual terisolasi (`TestRegister_Success`, `TestRegister_DuplicateEmail`, `TestRegister_ShortPassword`, `TestLogin_Success`, `TestLogin_RefreshToken_Success`, `TestLogin_WrongPassword`, `TestLogin_UserNotFound`, `TestLogin_InactiveUser`) sehingga Cognitive Complexity turun drastis ke level minimum (<2).

### [2026-08-18 16:01:53] - Modul Autentikasi: Middleware Autentikasi & RBAC
- Implementasi HTTP Auth Middleware di `backend/internal/delivery/http/middleware/auth.go` untuk validasi header `Authorization: Bearer <token>` dan injeksi `UserClaims`, `UserID`, dan `OrganizationID` ke dalam request context.
- Implementasi HTTP RBAC Middleware di `backend/internal/delivery/http/middleware/rbac.go` dengan validasi keanggotaan organisasi target dan hierarki peran pengguna (`owner` > `admin` > `member` > `viewer`).
- Pembuatan pengujian otomatis unit test di `backend/tests/middleware_test.go` yang menguji skenario token valid/invalid/missing dan otorisasi peran yang diizinkan maupun ditolak dengan hasil kelulusan 100% (`PASS`).

### [2026-08-18 17:22:57] - Modul Autentikasi: Interceptor Audit Logging
- Implementasi PostgreSQL AuditRepository di `backend/internal/repository/postgres/audit_repository.go` untuk persistensi dan paginasi data audit log berdasarkan organisasi.
- Implementasi HTTP AuditLogInterceptor di `backend/internal/delivery/http/middleware/audit.go` yang secara otomatis mencegat request mutasi data (`POST`, `PUT`, `PATCH`, `DELETE`) dan mencatat `UserID`, `OrganizationID`, alamat IP (`X-Forwarded-For`/`RemoteAddr`), `User-Agent`, endpoint aksi, status code, dan metadata resource terkait.
- Pembuatan suite unit test di `backend/tests/audit_test.go` untuk repository dan interceptor dengan hasil kelulusan 100% (`PASS`).

### [2026-08-18 20:08:53] - Layer Abstraksi Provider: Definisi Kontrak Lifecycle Driver
- Pendefinisian interface `ProviderDriver` pada `backend/internal/domain/provider.go` yang mencakup seluruh lifecycle operasi VPS (`CreateServer`, `GetServer`, `ListServers`, `RebootServer`, `ShutdownServer`, `StartServer`, `ResizeServer`, `DeleteServer`).
- Penambahan struktur Data Transfer Objects (DTO) terstandarisasi untuk provider cloud: `CreateServerRequest`, `ResizeServerRequest`, dan `ProviderServer`.

### [2026-08-18 20:12:21] - Layer Abstraksi Provider: Implementasi MockProvider & Manajemen Kredensial
- Implementasi modul kriptografi AES-256-GCM di `backend/pkg/encryptor/encryptor.go` untuk enkripsi dan dekripsi field data sensitif (API Key, API Secret, SSH Key).
- Implementasi driver `MockProvider` di `backend/internal/provider/mock/mock_driver.go` untuk simulasi lokal seluruh siklus hidup VPS (provisioning, restart, stop, start, resize, delete) dengan alokasi IP otomatis.
- Implementasi repository PostgreSQL untuk Provider (`backend/internal/repository/postgres/provider_repository.go`) dan Credential (`backend/internal/repository/postgres/credential_repository.go`).
- Implementasi use case manajemen kredensial provider di `backend/internal/usecase/provider/credential_usecase.go` yang memverifikasi kepemilikan organisasi dan mengamankan kredensial cloud.
- Pembuatan pengujian otomatis unit test di `backend/tests/mock_provider_test.go` dan `backend/tests/credential_usecase_test.go` dengan hasil kelulusan 100% (`PASS`).

### [2026-08-18 20:14:13] - Refaktorisasi Driver MockProvider & Test Suite (SonarQube Compliance)
- Ekstraksi logika transisi status pada `backend/internal/provider/mock/mock_driver.go` ke dalam fungsi pembantu `updateServerStatus` untuk mengatasi issue implementasi metode identik pada `RebootServer` dan `StartServer`.
- Dekomposisi suite pengujian `backend/tests/mock_provider_test.go` menjadi fungsi-fungsi pengujian individual (`TestMockDriver_CreateAndGetServer`, `TestMockDriver_ListServers`, `TestMockDriver_PowerControls`, `TestMockDriver_ResizeServer`, `TestMockDriver_DeleteServer`) guna menurunkan Cognitive Complexity ke level minimum ($\le 2$).

### [2026-08-19 22:37:53] - Modul Manajemen Server & VPS: REST API & Lifecycle Controls
- Implementasi PostgreSQL ServerRepository di `backend/internal/repository/postgres/server_repository.go` untuk operasi CRUD, paginasi per organisasi, dan pembaruan status server.
- Implementasi Provider Factory di `backend/internal/provider/factory.go` untuk registrasi dan pemanggilan dinamis instance `ProviderDriver`.
- Implementasi use case server di `backend/internal/usecase/server/server_usecase.go` yang mengorkestrasi provisioning, reboot, shutdown, start, resize spesifikasi, dan penghapusan server dengan integrasi driver provider.
- Implementasi HTTP AuthHandler (`backend/internal/delivery/http/v1/auth_handler.go`) dan ServerHandler (`backend/internal/delivery/http/v1/server_handler.go`) untuk endpoint REST API (`/api/v1/auth/*` dan `/api/v1/servers/*`).
- Pembaruan HTTP Router dan injeksi dependensi pada `backend/cmd/api/main.go`.
- Pembuatan pengujian otomatis end-to-end dan integrasi HTTP di `backend/tests/server_test.go` dengan hasil kelulusan 100% (`PASS`).

### [2026-08-20 00:11:24] - Frontend Control Panel: Dashboard MVP & Server Management
- Desain antarmuka modern bertema gelap (*dark mode design system*), token CSS HSL, glassmorphism, dan komponen UI reusable (`Button`, `Card`, `Badge`, `Input`, `Dialog`, `ServerStatusBadge`).
- Implementasi klien API Axios terpusat dengan request interceptor injeksi token JWT Bearer dan penanganan otomatis sesi kedaluwarsa.
- Implementasi state management menggunakan Zustand (`useAuthStore`, `useServerStore`, `useThemeStore`).
- Implementasi halaman Autentikasi (`/login` dan `/register`) dengan validasi form, visibilitas sandi, dan auto-redirect.
- Implementasi Shell Layout Dashboard (`Sidebar`, `Header` dengan breadcrumbs dinamis, status badge sistem, profil dropdown, dan theme switcher).
- Implementasi Halaman Overview (`/overview`) dengan kartu metrik agregat total server, status running/stopped, total alokasi vCPU/RAM/Disk, dan ringkasan Sentinel security score.
- Implementasi Halaman Manajemen VPS (`/infrastructure/vps`) lengkap dengan tabel server, salin IP 1-klik, filter status & pencarian, modal deploy server baru, modal resize spesifikasi, serta tombol aksi cepat reboot/shutdown/start/terminate.
- Implementasi Halaman Detail Server (`/infrastructure/vps/[id]`) yang menampilkan rincian komputasi, jaringan, dan grafik utilisasi telemetri.
- Seluruh rute terverifikasi dan berhasil dikompilasi melalui `pnpm run build` dengan status 100% lulus.

### [2026-08-20 00:41:50] - Frontend Design System: Refaktorisasi Tema Supabase Green & Black & Centralized Tokens
- Pembuatan modul token desain terpusat di `frontend/src/core/theme/` (`app_colors.ts`, `app_text.ts`, `app_containers.ts`, `app_theme.ts`, dan barrel export `index.ts`) guna menghilangkan hardcoded styling.
- Penerapan palet tema minimalis Supabase Deep Black (`#0f0f0f` / `#171717`) dan Emerald Green (`#3ECF8E` / `emerald-500`) pada seluruh komponen antarmuka, kartu, tombol, badge, dan input.
- Perbaikan mekanisme fungsionalitas Theme Switcher (`ThemeToggle.tsx` dan `useThemeStore.ts`) dengan reaktivitas penuh pada elemen HTML dan variabel CSS.
- Perbaikan masalah layout modal dialog pada `dialog.tsx` dengan menambahkan flex container `max-h-[90vh]` dan scroll container internal `overflow-y-auto` agar form tidak terpotong pada layar beresolusi rendah.
- Pembersihan seluruh fallback URL hardcode pada `services/api.ts` dan penyesuaian `.env.local` / `.env.example`.
- Verifikasi kompilasi production build Next.js 16 App Router (`pnpm run build`) berstatus 100% lulus tanpa error dan bebas dari peringatan SonarQube.
