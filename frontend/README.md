# Caelus Cloud - Frontend Control Panel

Antarmuka web modern (*Control Plane Panel*) untuk platform **Caelus Cloud**, dibangun menggunakan **Next.js 16 (App Router)**, **React 19**, **TypeScript**, dan **Tailwind CSS v4** dengan modul sistem desain terpusat (*Supabase Dark Theme Tokens*).

---

## 1. Arsitektur & Struktur Direktori

```text
frontend/
├── src/
│   ├── app/                            # Next.js App Router (Rute & Layouts)
│   │   ├── (auth)/                     # Rute Autentikasi Publik (Login, Register)
│   │   │   ├── login/page.tsx
│   │   │   └── register/page.tsx
│   │   │
│   │   ├── (dashboard)/                # Rute Terproteksi Master Control Plane
│   │   │   ├── overview/page.tsx       # Tampilan Ringkasan Metrik, Server & Postur Keamanan
│   │   │   ├── infrastructure/
│   │   │   │   ├── vps/page.tsx        # Manajemen VPS / Server (BYOS & Cloud)
│   │   │   │   ├── providers/page.tsx  # Multi-Cloud Providers & Encrypted Credential Hub
│   │   │   │   ├── iac/page.tsx        # Declarative IaC Editor, Plan Diff Viewer & Rollback
│   │   │   │   └── containers/page.tsx # Container Orchestration & Live Web Streaming Terminal
│   │   │   ├── storage/page.tsx        # S3 Object Storage Explorer, Upload & Presigned URLs
│   │   │   ├── monitoring/page.tsx     # Metrik Telemetri Host Real-time & Live System Logs
│   │   │   ├── security/page.tsx       # Sentinel Security Posture, Scanners & Finding Fixes
│   │   │   ├── automation/page.tsx     # Rule Engine Builder & Multi-Channel Incident Notifier
│   │   │   └── settings/page.tsx       # Pengaturan Organisasi & Profil Akun
│   │   │
│   │   ├── globals.css                 # Konfigurasi Tailwind CSS v4 & Theme Variables
│   │   └── layout.tsx                  # Root Layout (Theme Provider, Query Client, Toaster)
│   │
│   ├── core/                           # Token Desain & Konstanta Tema Terpusat
│   │   └── theme/                      # Supabase Dark Theme System
│   │       ├── app_colors.ts           # Token Warna (#121212 Deep Black, Emerald Green, Slate)
│   │       ├── app_text.ts             # Skala Tipografi Terstandarisasi
│   │       ├── app_containers.ts       # Batasan Card, Padding, & Modal Scroll
│   │       └── index.ts                # Barrel Export (@/core/theme)
│   │
│   ├── components/                     # Komponen UI Reusable
│   │   ├── layout/                     # Sidebar, Header, Breadcrumbs, UserDropdown
│   │   ├── server/                     # ConnectAgentModal, CreateServerModal, ServerStatusBadge
│   │   └── ui/                         # Primitif UI (Button, Input, Card, Table, Dialog, Badge)
│   │
│   ├── services/                       # HTTP API Clients (Axios Interceptors & Bearer Auth)
│   │   ├── api.ts                      # Base Axios Client dengan Refresh Token Interceptor
│   │   ├── auth.service.ts             # Login, Register, Session Validation
│   │   ├── server.service.ts           # Server CRUD, Power Actions, Resize
│   │   ├── credential.service.ts       # Multi-Cloud Credential CRUD & Test Connection
│   │   ├── iac.service.ts              # IaC Manifests, Syntax Validation, Plan, Apply, Rollback
│   │   ├── deployment.service.ts       # Container Deployments, Stop, Rollback, SSE Logs
│   │   ├── storage.service.ts          # Bucket Management & S3 Object Actions
│   │   ├── monitoring.service.ts       # Telemetri Real-time & Alert Rules
│   │   └── security.service.ts         # Sentinel Scans, Score & Findings Remediation
│   │
│   ├── stores/                         # State Management (Zustand)
│   │   ├── useAuthStore.ts             # Sesi User, Active Organization, & Token State
│   │   └── useServerStore.ts           # Cache Data Server & Filter State
│   │
│   └── types/                          # Definisi Tipe TypeScript Terstruktur
│       ├── api.ts                      # ApiResponse<T>, ErrorResponse, PaginatedResponse
│       ├── auth.ts                     # UserProfile, LoginCredentials, OrgInfo
│       ├── server.ts                   # Server, Provider, HardwareSpec
│       ├── credential.ts               # CloudCredential, ProviderType, TestResult
│       ├── iac.ts                      # IaCConfiguration, IaCPlan, IaCChange, IaCState
│       ├── deployment.ts               # Deployment, DeploymentLog, PortBinding, VolumeBinding
│       ├── storage.ts                  # BucketItem, StorageObjectItem
│       ├── monitoring.ts               # MetricSeries, TelemetrySnapshot, AlertRule
│       └── security.ts                 # SentinelFinding, SecurityScan, SecurityPosture
│
├── .env.example                        # Template Variabel Environment Frontend
├── .env.local                          # Environment Lokal Aktif (Git-ignored)
├── package.json
└── tsconfig.json
```

---

## 2. Fitur Utama Panel

1. **Declarative IaC Hub (`/infrastructure/iac`)**:
   - YAML Manifest Code Editor dengan deteksi kesalahan sintaks dan baris error secara *real-time*.
   - *Starter Templates* siap pakai (Fullstack App, Multi-Cloud High Availability, Microservices Stack).
   - **Visual Diff Viewer**: Komparasi grafis sebelum Apply dengan badge warna (+ Create, ~ Update, - Delete, = No-op) dan rincian sebelum vs sesudah.
   - **State History & Rollback Drawer**: Riwayat versi snapshot state dengan checksum integritas SHA-256 dan tombol 1-klik Rollback.

2. **Container Orchestration & Live Console (`/infrastructure/containers`)**:
   - Manajemen siklus hidup kontainer Docker dengan status dinamis (`RUNNING`, `PULLING`, `DEPLOYING`, `STOPPED`).
   - Modal peluncuran kontainer dengan *quick image presets* (Nginx, Redis, PostgreSQL, Node.js), alokasi port, volume mount, dan variabel environment dinamis.
   - **Web-based Streaming Terminal**: Konsol ANSI gelap berkecepatan tinggi yang terhubung via Server-Sent Events (SSE) dengan auto-scroll, log search, filter kategori log (`stdout`, `stderr`, `system`), dan 1-klik copy.

3. **Multi-Cloud Providers Hub (`/infrastructure/providers`)**:
   - Katalog terpadu provider: AWS EC2, Hetzner Cloud, DigitalOcean, Contabo, Custom Host, dan Mock Provider.
   - Tabel kredensial aktif dengan indikator enkripsi tingkat militer (AES-256-GCM Encrypted at Rest) dan uji konektivitas API langsung (*Live Test*).

4. **Sentinel Security Hub (`/security`)**:
   - Visualisasi skor postur keamanan terpadu (0 - 100) dan predikat Grade A/B/C/D/F.
   - Peluncuran pemindaian instan (Port Scanner, TLS/SSL, HTTP Security Headers, Host CIS Config, Vulnerability).
   - Tabel temuan kerentanan dengan filter tingkat keparahan (*Critical*, *High*, *Medium*, *Low*) dan modal rekomendasi tindakan remediasi 1-klik salin.

---

## 3. Konfigurasi Lingkungan (`.env.local`)

Buat berkas `.env.local` pada direktori `frontend/` atau gunakan konfigurasi default:

```env
# Caelus Backend REST API URL
NEXT_PUBLIC_API_URL=http://localhost:8080

# Caelus Backend WebSocket Hub URL
NEXT_PUBLIC_WS_URL=ws://localhost:8080/api/v1/ws
```

---

## 4. Panduan Menjalankan Frontend

Anda dapat menjalankan panel kontrol frontend dengan cepat menggunakan perintah **`Makefile`** dari root direktori proyek, atau menggunakan package manager (`pnpm` / `npm`) secara langsung.

### A. Menggunakan Makefile (Direkomendasikan dari Root Monorepo)

| Perintah | Deskripsi |
| :--- | :--- |
| **`make deps-frontend`** | Menginstal seluruh paket dependensi frontend via pnpm |
| **`make frontend`** | Menjalankan Next.js Dev Server pada `http://localhost:3000` (dengan Hot Reload) |
| **`make lint`** | Menjalankan audit linting ESLint dan TypeScript checker |
| **`make build`** | Melakukan kompilasi produksi seluruh monorepo termasuk frontend |

### B. Menjalankan Langsung via Package Manager (Dari Direktori `frontend/`)

#### 1. Menginstal Dependensi
```bash
cd frontend
pnpm install
# atau: npm install
```

#### 2. Menjalankan Server Development
```bash
cd frontend
pnpm dev
# atau: npm run dev
```
Buka peramban di [http://localhost:3000](http://localhost:3000).

#### 3. Memeriksa Kualitas Kode & Linting
```bash
cd frontend
pnpm lint
# atau: npm run lint
```

#### 4. Melakukan Kompilasi & Build Produksi
```bash
cd frontend
pnpm build
# atau: npm run build
```
Kompilasi akan memvalidasi seluruh tipe data TypeScript dan menghasilkan bundle produksi yang teroptimasi tanpa error.

