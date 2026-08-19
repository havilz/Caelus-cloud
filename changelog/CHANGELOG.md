# Changelog

Semua catatan perubahan pada proyek Caelus Cloud didokumentasikan di sini secara berkala.

Format penulisan mengacu pada standar formal dengan pencatatan stempel tanggal dan waktu `[YYYY-MM-DD HH:mm:ss]`.

---

## [Unreleased]

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







