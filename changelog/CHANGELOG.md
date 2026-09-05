# Changelog

Semua catatan perubahan pada proyek Caelus Cloud didokumentasikan di sini secara berkala.

Format penulisan mengacu pada standar formal dengan pencatatan stempel tanggal dan waktu `[YYYY-MM-DD HH:mm:ss]`.

---

## [Unreleased]

### [2026-09-05 19:50:00] - Architecture Realignment & Test Suite Verification: Task 9.4 Resolution (Audit C-3)

- **Penyelarasan Arsitektur Self-Hosted & Penghapusan Task 9.4 (`docs/TASK.md`)**:
  - Menyelaraskan dokumen perencanaan dengan model deployment aktual Caelus Cloud (Single-Tenant / Self-Hosted Private Instance).
  - Menghapus Task 9.4 (Uji Integrasi Otomatis Isolasi Multi-Tenant) karena kebutuhan proteksi antar-organisasi eksternal telah usang (obsolete) pada instance mandiri yang tidak berbagi database publik.
- **Eksekusi dan Validasi Pengujian Otomatis Menyeluruh (`backend/tests/...`, `agent/tests/...`)**:
  - Menjalankan seluruh test suite otomatis dan pengujian integrasi backend (30 test suites, 100% PASS dalam 5.68s) yang memvalidasi stabilitas seluruh modul inti (autentikasi JWT, intra-org RBAC, sanitasi IP klien, proteksi path volume allowlist, lifecycle server, storage, backup, dan security scanners).
  - Memverifikasi keberhasilan build binary backend (`go build ./...`) dan binary agent (`go build ./...`) tanpa error kompilasi.

### [2026-09-05 19:15:00] - Security Hardening: Container Escape Mitigation via Path Allowlist (Audit C-2)

- **Package Validasi Path Keamanan (`backend/pkg/security/path.go`)**:
  - Mengimplementasikan `ValidateHostPath` dengan pendekatan *strict allowlist* (hanya mengizinkan subpath volume di bawah `/var/lib/caelus/volumes`, `/opt/caelus/volumes`, atau direktori yang dikonfigurasi via `ALLOWED_VOLUME_ROOTS`).
  - Menerapkan resolusi canonical path dengan `filepath.Clean` dan evaluasi traversal symlink via `filepath.EvalSymlinks` yang mengecek hierarki ancestor secara mendalam untuk menggagalkan eksploitasi *symlink escape*.
  - Melarang mounting root directory itu sendiri secara langsung untuk menjamin isolasi volume antar-aplikasi.
- **Konfigurasi Allowed Volume Roots (`backend/pkg/config/config.go`, `.env.example`, `docker-compose.yml`, `cmd/api/main.go`)**:
  - Menambahkan field `AllowedVolumeRoots` pada `AppConfig` dan inisialisasi allowlist saat server API bootstrap.
  - Mendokumentasikan variabel `ALLOWED_VOLUME_ROOTS` pada `.env.example` dan meneruskannya ke kontainer `api` di `docker-compose.yml`.
- **Harmonisasi di Layer Orchestration & IaC (`deployment_usecase.go`, `docker_pipeline.go`, `apply_engine.go`)**:
  - Mengganti denylist hardcoded lama pada deployment usecase dan docker pipeline dengan delegasi ke `security.ValidateHostPath`.
  - Mengamankan parsing volume bind mount pada engine eksekusi template IaC.
- **Suite Pengujian Keamanan Komprehensif (`backend/tests/path_validator_test.go`)**:
  - Menambahkan pengujian menyeluruh: penolakan direktori sensitif sistem (`/`, `/etc`, `/root`, `/home`, `/tmp`, `/opt`, `/var/run/docker.sock`), penolakan path traversal (`../`), penolakan relative path, serta simulasi serangan symlink escape nyata pada filesystem (`t.TempDir()`).
  - Memverifikasi penolakan binding host path tidak aman saat pembuatan deployment melalui `CreateDeployment`.

### [2026-09-05 18:55:00] - Security Hardening: IP Sanitization & Anti-Spoofing Rate Limiter (Audit H-1)

- **Konfigurasi Trusted Proxies (`backend/pkg/config/config.go`, `.env.example`, `docker-compose.yml`)**:
  - Menambahkan field `TrustedProxies` pada `AppConfig` yang membaca `TRUSTED_PROXIES` dari environment (default `127.0.0.1,::1`).
  - Meneruskan `TRUSTED_PROXIES` pada deklarasi service container `api` di `docker-compose.yml`.
- **Middleware Sanitasi & Validasi IP Klien (`backend/internal/delivery/http/middleware/client_ip.go`, `router.go`)**:
  - Mengganti middleware `chimiddleware.RealIP` bawaan Chi dengan `ClientIPMiddleware` kustom berbasis `IPValidator`.
  - Mengimplementasikan pre-parsing alamat IP tunggal dan CIDR subnet block (`net.IPNet`) untuk evaluasi reverse proxy yang efisien dan aman.
  - Menerapkan penelusuran rantai proxy *right-to-left* pada header `X-Forwarded-For` untuk mengidentifikasi alamat IP klien pertama yang tidak terpercaya, serta fallback langsung ke `RemoteAddr` jika koneksi berasal dari pengirim langsung tanpa proxy terdaftar.
- **Harmonisasi Komponen Autentikasi, Audit Log, dan Rate Limiter (`rate_limit.go`, `audit.go`, `auth_handler.go`)**:
  - Menyelaraskan seluruh pencatatan audit dan penentuan kunci rate limiting login/registrasi agar selalu menggunakan `ExtractClientIP(r)` yang telah tervalidasi.
  - Menghapus fungsi ekstraksi IP duplikat dan metode verifikasi longgar `isTrustedProxy` lama.
- **Suite Pengujian Keamanan (`backend/tests/client_ip_test.go`)**:
  - Menambahkan pengujian integrasi dan unit test mitigasi bypass rate limiting: serangan 6 request berturut-turut dengan variasi header `X-Forwarded-For` palsu berhasil diblokir dengan status HTTP 429 Too Many Requests.
  - Memverifikasi isolasi rate limit yang adil bagi klien sah yang berada di balik reverse proxy yang sama.

### [2026-09-05 09:45:00] - Security Hardening: Intra-Organization Role Enforcement (Audit M-4)

- **Middleware RBAC & Proteksi Rute Backend (`backend/internal/delivery/http/router.go`, `middleware/rbac.go`)**:
  - Mengintegrasikan `RequireOrganizationRole` pada seluruh endpoint kredensial cloud provider (`/credentials`) sehingga hanya pengguna dengan role `admin` atau `owner` yang berhak melihat dan mengelola kredensial.
  - Membatasi operasi mutasi dan kontrol destruktif server (`CreateServer`, `DeleteServer`, `RebootServer`, `ShutdownServer`, `ResizeServer`) dengan izin minimal role `admin`.
  - Mengamankan manajemen keanggotaan dan API keys/webhooks pada `/settings` (khusus perubahan role `UpdateMemberRole` diwajibkan role `owner`).
  - Menambahkan helper `RequireAdmin` dan `RequireOwner` serta integrasi context organisasi dari `GetOrgIDFromContext`.
- **Penyesuaian Antarmuka Pengguna Frontend (`frontend/src/hooks/useRoleGuard.ts`, `providers/page.tsx`, `vps/page.tsx`, `MembersTab.tsx`)**:
  - Mengimplementasikan custom React hook `useRoleGuard` untuk mengevaluasi hak akses pengguna secara reaktif berbasis peran organisasi.
  - Menyembunyikan form dan tombol penambahan kredensial serta menampilkan banner peringatan akses terbatas pada halaman Cloud Providers jika pengguna berstatus non-admin.
  - Membatasi tombol terminasi dan kontrol daya server pada halaman VPS Management dan VPS Detail untuk role `member` dan `viewer`.
  - Membatasi aksi undang anggota tim dan pembatalan undangan untuk role non-admin, serta mengunci dropdown perubahan role khusus untuk role `owner`.

