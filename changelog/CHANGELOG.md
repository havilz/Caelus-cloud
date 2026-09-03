# Changelog

Semua catatan perubahan pada proyek Caelus Cloud didokumentasikan di sini secara berkala.

Format penulisan mengacu pada standar formal dengan pencatatan stempel tanggal dan waktu `[YYYY-MM-DD HH:mm:ss]`.

---

## [Unreleased]

### [2026-09-03 11:55:00] - Frontend UI Hardening & Layout Refinement

- **Perbaikan Modal Detail Volume (`frontend/src/app/(dashboard)/infrastructure/volumes/page.tsx`)**:
  - Memperbaiki tata letak modal dialog detail volume dengan struktur card yang responsif (`maxWidth="lg"`).
  - Menambahkan pembungkusan teks (`break-all`, `select-all`) untuk identifier panjang seperti UUID, Docker Volume Hash, dan Mount Path direktori agar tidak merusak margin horizontal layout.
  - Memperbaiki hierarki tipografi dan kontras warna status container dan performance IOPS.
- **Penyelarasan Kartu Ringkasan Storage Pool (`frontend/src/app/(dashboard)/infrastructure/volumes/page.tsx`)**:
  - Merestrukturisasi grid kartu ringkasan telemetri storage menjadi layout yang kompak dan seimbang (`grid-cols-1 sm:grid-cols-2 lg:grid-cols-4`).
  - Menyelaraskan padding, container icon, dan dimensi kartu agar tidak merenggang secara berlebihan pada resolusi layar lebar.
- **Pembersihan Badge Phase pada Antarmuka Pengguna (`frontend/src/app/(dashboard)/infrastructure/containers/page.tsx`, `iac/page.tsx`)**:
  - Menghapus seluruh tampilan label internal seperti `Phase 6.2 Active` dan `Phase 6.3 Active` pada header halaman Containers dan Declarative IaC untuk menjaga standar antarmuka produksi enterprise.
- **Pembaruan Skema Migrasi Deployments (`backend/migrations/000008_create_iac_and_deployments.up.sql`)**:
  - Menambahkan kolom `network_name VARCHAR(255)` dan `command TEXT` pada tabel `deployments` agar sinkronisasi auto-discovery container berjalan konsisten tanpa error query database.
