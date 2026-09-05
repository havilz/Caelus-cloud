# Changelog

Semua catatan perubahan pada proyek Caelus Cloud didokumentasikan di sini secara berkala.

Format penulisan mengacu pada standar formal dengan pencatatan stempel tanggal dan waktu `[YYYY-MM-DD HH:mm:ss]`.

---

## [Unreleased]

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

