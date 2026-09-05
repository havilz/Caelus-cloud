# Changelog

Semua catatan perubahan pada proyek Caelus Cloud didokumentasikan di sini secara berkala.

Format penulisan mengacu pada standar formal dengan pencatatan stempel tanggal dan waktu `[YYYY-MM-DD HH:mm:ss]`.

---

## [Unreleased]

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

