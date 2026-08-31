# Changelog

Semua catatan perubahan pada proyek Caelus Cloud didokumentasikan di sini secara berkala.

Format penulisan mengacu pada standar formal dengan pencatatan stempel tanggal dan waktu `[YYYY-MM-DD HH:mm:ss]`.

---

## [Unreleased]

### [2026-08-31 16:55:00] - Settings Module, Container Lifecycle Orchestration, & Infrastructure Resiliency

- **Modul Settings & Manajemen Organisasi / Pengguna / Kredensial Global (`backend/migrations/000013_create_settings_and_integrations.up.sql`, `usecase/settings/`, `delivery/http/v1/settings_handler.go`, `frontend/src/app/(dashboard)/settings/`)**:
  - Mengimplementasikan tab Profil & Keamanan (Argon2id password hashing, avatar, indikator 2FA).
  - Mengimplementasikan Manajemen Organisasi & Tim RBAC (`Owner`, `Admin`, `Member`, `Viewer`) dengan sistem undangan email (`organization_invitations`).
  - Mengimplementasikan Developer Personal Access Tokens (`api_keys`) dengan format prefix `caelus_pat_` dan hash SHA-256 aman.
  - Mengimplementasikan Outgoing Webhook Dispatcher (`webhooks`) dengan verifikasi signature HMAC-SHA256 (`X-Caelus-Signature`) dan tombol Test Ping real-time.
  - Mengimplementasikan Audit Logs platform terintegrasi dengan filter pencarian dan modal detail payload JSON.
- **Container Lifecycle Orchestration & Auto-Discovery Resiliency (`agent/internal/docker/inspector.go`, `agent/cmd/main.go`, `monitoring_usecase.go`)**:
  - Menambahkan dukungan eksekusi aksi remote `START_CONTAINER`, `STOP_CONTAINER`, dan `RESTART_CONTAINER` pada `caelus-agent` melalui Docker REST API Unix Socket.
  - Memperbaiki sinkronisasi volume persistent otomatis (`AttachedContainerName`) agar container yang me-mount volume terdeteksi dan tercatat secara real-time pada dashboard.
  - Mengoptimalkan migrasi database PostgreSQL dengan klausa `DROP TRIGGER IF EXISTS` sebelum pembuatan trigger untuk menjamin 100% idempotensi skrip migrasi.

