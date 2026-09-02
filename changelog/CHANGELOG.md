# Changelog

Semua catatan perubahan pada proyek Caelus Cloud didokumentasikan di sini secara berkala.

Format penulisan mengacu pada standar formal dengan pencatatan stempel tanggal dan waktu `[YYYY-MM-DD HH:mm:ss]`.

---

## [Unreleased]

### [2026-09-02 20:12:00] - Security Audit Remediation Phase 7.4: Medium & Low Priority Hardening

- **Otomasi Pipeline CI/CD GitHub Actions (`.github/workflows/docker-publish.yml`)**:
  - Mengimplementasikan alur CI/CD otomatis untuk menjalankan unit test backend & agent pada setiap push ke branch `main` atau git tag versi (`v*`).
  - Mengompilasi biner statis `caelus-agent-linux` secara otomatis.
  - Membangun dan mengunggah (push) kelima Docker image Caelus Cloud ke GitHub Container Registry (`ghcr.io/havilz/caelus-api`, `caelus-worker`, `caelus-migrate`, `caelus-agent`, `caelus-frontend`) tanpa perlu build/push manual.
- **Sanitasi Argumen Docker CLI — M-1 (`usecase/orchestration/deployment_usecase.go`)**:
  - Mengimplementasikan fungsi validasi regex ketat untuk `AppName`, `ContainerName`, `ImageTag`, `NetworkName`, dan `RestartPolicy` untuk mencegah command & flag injection pada Docker CLI execution.
- **Deteksi Kerentanan & Audit Sentinel — M-2 (`sentinel/scanner/port_scanner.go`, `sentinel/scanner/vuln_scanner.go`)**:
  - Menambahkan aturan deteksi port berisiko tinggi pada PortScanner: Memcached (11211), Docker API unencrypted (2375), Elasticsearch (9200), dan VNC (5900).
  - Menambahkan aturan audit kerentanan CVE pada VulnScanner: Sudo Privilege Escalation (CVE-2021-3156 / Baron Samedit) dan Docker Socket Mount Security Audit.
- **Penyempurnaan Keamanan Transport Agent — M-3 (`agent/cmd/main.go`)**:
  - Menegaskan default `TLSSkipVerify = false` dan menambahkan log peringatan eksplisit (`SECURITY WARNING`) jika opsi verifikasi TLS dinonaktifkan pada daemon `caelus-agent`.
- **Entropi & Validasi Enkripsi AES-256-GCM — L-2 (`pkg/encryptor/encryptor.go`)**:
  - Mengimplementasikan helper `GenerateRandomKey` dan `ValidateKeyEntropy` untuk menjamin penggunaan kunci 32-byte (256-bit) berentropi tinggi.

### [2026-09-02 20:01:00] - Security Audit Remediation Phase 7.3: Rate Limiting & IaC Transparency (H-1 & H-2)

- **Rate Limiting Autentikasi & Pencatatan Log Gagal — H-1 (`middleware/rate_limit.go`, `v1/auth_handler.go`, `router.go`)**:
  - Mengimplementasikan middleware `AuthRateLimiter` dengan pembatasan berbasis jendela sliding (maksimum 5 percobaan per menit per kombinasi IP + Email).
  - Memasang `AuthRateLimiter` pada endpoint `POST /api/v1/auth/login` dan `POST /api/v1/auth/register`.
  - Mengintegrasikan pencatatan otomatis ke tabel `audit_logs` untuk seluruh percobaan login/registrasi gagal (`auth.login_failed`, `auth.register_failed`) maupun berhasil, merekam IP client, UserAgent, attempted email, dan status code.
- **Transparansi Mode Provider pada IaC Engine — H-2 (`iac/planner/plan_engine.go`)**:
  - Memperbarui `diffServers` pada IaC Planner untuk secara eksplisit melabeli mode eksekusi server: `BYOS / Local Host Agent` (custom driver) vs `Cloud Provider Driver (<provider>)`.
  - Menyematkan metadata `provider_mode` pada payload perubahan IaC untuk kejelasan visibilitas pada antarmuka dashboard dan audit plan.

### [2026-09-02 19:37:00] - Security Audit Remediation Phase 7.2: True Database Multi-Tenant Row Level Security (C-3)

- **Penegakan RLS Berbasis Session Context — C-3 (`backend/migrations/000015_enforce_strict_rls_policies.up.sql`, `middleware/org_context.go`, `router.go`)**:
  - Membuat migrasi `000015` yang mendefinisikan `CREATE POLICY ... AS RESTRICTIVE` pada 15 tabel multi-tenant: `servers`, `credentials`, `providers`, `organization_members`, `audit_logs`, `networks`, `volumes`, `deployments`, `deployment_logs`, `security_scans`, `security_findings`, `automation_rules`, `backup_policies`, `backup_records`, `iac_configurations`.
  - Setiap policy menggunakan `current_setting('app.current_org_id', true)` — mengembalikan string kosong jika variabel belum di-set, sehingga DENY-ALL aktif secara default.
  - Menerapkan `ALTER TABLE ... FORCE ROW LEVEL SECURITY` pada seluruh tabel di atas agar RLS berlaku bahkan untuk table owner PostgreSQL.
  - Mengimplementasikan middleware `InjectOrgContext` yang mengambil koneksi dari pgx pool dan mengeksekusi `SET LOCAL app.current_org_id = '<uuid>'` sebelum setiap request diproses.
  - Memasang `InjectOrgContext` pada grup rute terproteksi JWT di `router.go`, urutan: `Authenticate` → `InjectOrgContext` → `AuditLogInterceptor`.

### [2026-09-02 19:25:00] - Security Audit Remediation Phase 7.1: Telemetry Auth & Container Escape Mitigation

- **Autentikasi Telemetri Agen — C-1 (`backend/migrations/000014_add_agent_secret_hash.up.sql`, `domain/server.go`, `postgres/server_repository.go`, `middleware/agent_auth.go`, `router.go`)**:
  - Menambahkan kolom `agent_secret_hash` (Argon2id) dan `agent_secret_prefix` pada tabel `servers` via migrasi `000014`.
  - Mengimplementasikan method `SetAgentSecret` dan `GetByIDWithSecret` pada `ServerRepository` untuk menyimpan dan memverifikasi hash secret agen.
  - Membuat middleware `RequireAgentAuth` yang memvalidasi header `Authorization: Bearer` agen terhadap hash Argon2id yang tersimpan di database sebelum request telemetri diproses.
  - Memasang `RequireAgentAuth` pada endpoint `POST /api/v1/telemetry/report` sebagai lapisan autentikasi mandiri.
  - Memindahkan endpoint stream SSE `GET /api/v1/telemetry/stream/{server_id}` ke dalam grup rute terproteksi JWT — akses tidak lagi bersifat publik.
- **Pencegahan Container Escape — C-2 (`usecase/orchestration/deployment_usecase.go`, `orchestration/pipeline/docker_pipeline.go`)**:
  - Mengimplementasikan fungsi `validateHostPath` yang memblokir bind-mount path berbahaya: root (`/`), socket Docker (`/var/run/docker.sock`), dan direktori sistem (`/etc`, `/root`, `/sys`, `/proc`, `/dev`, `/bin`, `/usr`, `/var/run`, `/var/lib/docker`).
  - Validasi dieksekusi sebelum deployment record dibuat di database pada lapisan use case.
  - Menambahkan guard layer kedua pada `docker_pipeline.go` sebagai defense-in-depth sebelum argumen `-v` diteruskan ke Docker CLI.
- **Konsolidasi CORS — L-1 (`ws/handler.go`)**:
  - Menghapus header `Access-Control-Allow-Origin: *` yang ditulis manual di handler SSE — kebijakan CORS kini dikelola sepenuhnya oleh middleware global `cors.go`.
