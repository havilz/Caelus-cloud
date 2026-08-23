# Caelus Cloud - Backend API Service

Layanan Backend RESTful API untuk platform **Caelus Cloud** yang dibangun menggunakan bahasa pemrograman **Go** dengan pendekatan **Clean Architecture**, router **Chi**, dan basis data **PostgreSQL** (Supabase).

---

## 1. Arsitektur & Struktur Direktori

```text
backend/
├── cmd/
│   ├── api/          # Entry point aplikasi HTTP API server (make api)
│   ├── migrate/      # Standalone CLI tool untuk migrasi skema PostgreSQL (make migrate-up)
│   └── worker/       # Background worker, task queue, & cron scheduler (make worker)
├── internal/
│   ├── delivery/     # HTTP router, middleware (auth, rbac, audit, cors, logger), v1 handlers, WS/SSE Hub
│   ├── domain/       # Entitas bisnis murni, interface kontrak, dan error types
│   ├── provider/     # Implementasi multi-cloud drivers (AWS, Hetzner, DO, Contabo, Mock) & Sync Engine
│   ├── iac/          # Parser manifest YAML, Plan engine, dan Apply rollback engine
│   ├── orchestration/# Docker container deployment pipeline & streaming log emitter
│   ├── storage/      # Adapter S3-compatible (MinIO, AWS S3, Cloudflare R2)
│   ├── sentinel/     # Subsistem Sentinel Security (Scanner workers, Normalizer, Risk Engine)
│   ├── automation/   # Central event dispatcher & Rule engine
│   ├── queue/        # Distributed task queue engine (Redis)
│   ├── repository/   # PostgreSQL repository implementations dengan Row Level Security (RLS)
│   └── usecase/      # Interactor logika bisnis seluruh domain
├── pkg/              # Paket pembantu (config, hasher, jwt, encryptor, logger, validator, response)
├── migrations/       # Berkas DDL SQL migrasi (000001 hingga 000008)
└── tests/            # Test suite terpusat (Unit & Integration tests)
```

---

## 2. Prasyarat & Konfigurasi Lingkungan

Pastikan berkas `.env` telah disiapkan pada direktori root atau backend dengan konfigurasi berikut:

```env
# Aplikasi
APP_NAME=caelus-cloud-api
APP_ENV=development
APP_PORT=you_ruunning_port
APP_HOST=0.0.0.0
APP_DEBUG=true
APP_LOG_LEVEL=info
CORS_ORIGINS=*

# Basis Data (PostgreSQL / Supabase)
DB_HOST=aws-0-ap-southeast-1.pooler.supabase.com
DB_PORT=5432
DB_USER=postgres.your_project_ref
DB_PASSWORD=your_database_password
DB_NAME=your_db_name
DB_SSL_MODE=require
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=5m

# Keamanan & JWT
JWT_SECRET=super_secret_jwt_key_at_least_32_characters_long
JWT_ACCESS_EXPIRATION=15m
JWT_REFRESH_EXPIRATION=168h
ENCRYPTION_KEY=super_secret_encryption_key_32_bytes!
```

---

## 3. Cara Menjalankan Aplikasi

Anda dapat menjalankan backend service dengan cepat menggunakan perintah **`Makefile`** dari root direktori proyek, atau menjalankan perintah `go` secara langsung.

### A. Menggunakan Makefile (Direkomendasikan dari Root Monorepo)

| Perintah | Deskripsi |
| :--- | :--- |
| **`make deps-backend`** | Mengunduh dan memverifikasi modul dependensi Go backend |
| **`make infra-up`** | Menjalankan infrastruktur lokal (PostgreSQL, Redis, MinIO) via Docker |
| **`make migrate-up`** | Menjalankan migrasi basis data pending ke versi terbaru (*Up*) |
| **`make api`** | Menjalankan Backend REST API Server pada `http://localhost:8080` |
| **`make worker`** | Menjalankan Asynchronous Background Worker & Task Scheduler |
| **`make test-backend`** | Menjalankan seluruh test suite unit & integrasi backend |
| **`make build-backend`** | Melakukan kompilasi binary produksi `backend/bin/api` |

### B. Menjalankan Langsung via Go CLI (Dari Direktori `backend/`)

#### 1. Menjalankan Migrasi Skema Basis Data
```bash
cd backend
go run cmd/migrate/main.go -direction=up
```

#### 2. Menjalankan Server API (Development Mode)
```bash
cd backend
go run cmd/api/main.go
```

#### 3. Menjalankan Background Worker & Task Scheduler
```bash
cd backend
go run cmd/worker/main.go
```

#### 4. Melakukan Kompilasi dan Menjalankan Binary Produksi
```bash
cd backend
go build -ldflags="-s -w" -o bin/caelus-api cmd/api/main.go
./bin/caelus-api
```

#### 5. Menjalankan Pengujian Otomatis (Test Suite)
Seluruh pengujian unit dan integrasi tersimpan terpusat di direktori `tests/`:
```bash
cd backend
go test -v ./tests/...
```

---

## 4. Dokumentasi Endpoint REST API

### 4.1. Endpoint Pemantauan Kesehatan (Public)

#### `GET /health`
Mengembalikan status kesehatan umum API service.
```bash
curl http://localhost:8080/health
```
**Respon Sukses (200 OK):**
```json
{
  "success": true,
  "message": "Caelus Cloud API is healthy",
  "data": {
    "env": "development",
    "service": "caelus-cloud-api",
    "status": "ok"
  }
}
```

#### `GET /api/v1/health`
Mengembalikan status kesehatan operasional API v1.
```bash
curl http://localhost:8080/api/v1/health
```

---

### 4.2. Modul Autentikasi (Public)

#### `POST /api/v1/auth/register`
Mendaftarkan akun pengguna baru sekaligus membuat workspace/organisasi default.
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "full_name": "Administrator",
    "email": "admin@caelus.cloud",
    "password": "supersecretpassword123",
    "organization_name": "Production Workspace"
  }'
```
**Respon Sukses (201 Created):**
```json
{
  "success": true,
  "message": "Registrasi akun berhasil",
  "data": {
    "user": {
      "id": "11111111-1111-1111-1111-111111111111",
      "email": "admin@caelus.cloud",
      "full_name": "Administrator",
      "is_active": true,
      "created_at": "2026-08-20T00:00:00Z"
    },
    "organization": {
      "id": "22222222-2222-2222-2222-222222222222",
      "name": "Production Workspace",
      "slug": "production-workspace",
      "role": "owner"
    },
    "tokens": {
      "access_token": "eyJhbGciOiJIUzI1Ni...",
      "refresh_token": "eyJhbGciOiJIUzI1Ni...",
      "token_type": "Bearer",
      "expires_in": 900
    }
  }
}
```

#### `POST /api/v1/auth/login`
Melakukan otentikasi kredensial pengguna dan menerbitkan pasangan token JWT.
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@caelus.cloud",
    "password": "supersecretpassword123"
  }'
```

#### `POST /api/v1/auth/refresh`
Memperbarui access token menggunakan refresh token yang masih aktif.
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGciOiJIUzI1Ni..."
  }'
```

---

### 4.3. Modul Provider Cloud (Public)

#### `GET /api/v1/providers`
Mengambil daftar seluruh penyedia cloud yang didukung sistem (beserta UUID `id` masing-masing untuk digunakan saat membuat server).
```bash
curl http://localhost:8080/api/v1/providers
```
**Respon Sukses (200 OK):**
```json
{
  "success": true,
  "message": "Daftar provider yang didukung berhasil diambil",
  "data": [
    {
      "id": "a1b2c3d4-0000-0000-0000-000000000001",
      "name": "Mock Cloud Provider",
      "slug": "mock",
      "is_active": true,
      "created_at": "2026-08-20T00:00:00Z"
    },
    {
      "id": "a1b2c3d4-0000-0000-0000-000000000002",
      "name": "Amazon Web Services",
      "slug": "aws",
      "is_active": true,
      "created_at": "2026-08-20T00:00:00Z"
    }
  ]
}
```

---

### 4.4. Modul Manajemen Server & VPS (Protected)

Semua endpoint di bawah ini membutuhkan header otorisasi JWT Bearer:
`Authorization: Bearer <ACCESS_TOKEN>`

#### `GET /api/v1/servers`
Mengambil daftar seluruh server milik organisasi pengguna dengan paginasi.
```bash
curl -X GET "http://localhost:8080/api/v1/servers?page=1&limit=20" \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```
**Respon Sukses (200 OK):**
```json
{
  "success": true,
  "message": "Daftar server berhasil diambil",
  "data": [
    {
      "id": "33333333-3333-3333-3333-333333333333",
      "organization_id": "22222222-2222-2222-2222-222222222222",
      "provider_id": "44444444-4444-4444-4444-444444444444",
      "external_server_id": "mock-srv-a1b2c3d4",
      "name": "web-production-1",
      "hostname": "web-production-1",
      "ip_address": "198.51.100.45",
      "status": "running",
      "os_type": "ubuntu-22.04",
      "cpu_cores": 2,
      "memory_mb": 4096,
      "disk_gb": 50,
      "region": "ap-southeast-1",
      "created_at": "2026-08-20T00:00:00Z",
      "updated_at": "2026-08-20T00:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 20,
    "total_items": 1,
    "total_pages": 1
  }
}
```

#### `POST /api/v1/servers`
Membuat (provisioning) instance server VPS baru.
```bash
curl -X POST http://localhost:8080/api/v1/servers \
  -H "Authorization: Bearer <ACCESS_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "provider_id": "00000000-0000-0000-0000-000000000001",
    "name": "web-production-1",
    "region": "ap-southeast-1",
    "os_type": "ubuntu-22.04",
    "plan_id": "std-2vcpu-4gb",
    "cpu_cores": 2,
    "memory_mb": 4096,
    "disk_gb": 50
  }'
```

#### `GET /api/v1/servers/{id}`
Mengambil data detail server berdasarkan UUID.
```bash
curl -X GET http://localhost:8080/api/v1/servers/33333333-3333-3333-3333-333333333333 \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```

#### `POST /api/v1/servers/{id}/reboot`
Mengirim instruksi restart/reboot sistem operasi pada instance server.
```bash
curl -X POST http://localhost:8080/api/v1/servers/33333333-3333-3333-3333-333333333333/reboot \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```

#### `POST /api/v1/servers/{id}/shutdown`
Mematikan daya instance server (status berubah menjadi `stopped`).
```bash
curl -X POST http://localhost:8080/api/v1/servers/33333333-3333-3333-3333-333333333333/shutdown \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```

#### `POST /api/v1/servers/{id}/start`
Menyalakan kembali daya instance server yang sedang mati (status berubah menjadi `running`).
```bash
curl -X POST http://localhost:8080/api/v1/servers/33333333-3333-3333-3333-333333333333/start \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```

#### `PATCH /api/v1/servers/{id}/resize`
Mengubah alokasi kapasitas spesifikasi vCPU, RAM, atau Disk server.
```bash
curl -X PATCH http://localhost:8080/api/v1/servers/33333333-3333-3333-3333-333333333333/resize \
  -H "Authorization: Bearer <ACCESS_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "cpu_cores": 4,
    "memory_mb": 8192,
    "disk_gb": 100
  }'
```

#### `DELETE /api/v1/servers/{id}`
Menterminasi instance dari provider cloud dan menghapus rekaman server dari sistem.
```bash
curl -X DELETE http://localhost:8080/api/v1/servers/33333333-3333-3333-3333-333333333333 \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```

---

### 4.5. Modul Kredensial Multi-Provider Cloud (Protected)

Semua endpoint kredensial membutuhkan header otorisasi JWT Bearer:
`Authorization: Bearer <ACCESS_TOKEN>`

#### `GET /api/v1/credentials`
Mengambil daftar seluruh kredensial provider cloud milik organisasi pengguna (secret key di-mask secara aman).
```bash
curl -X GET http://localhost:8080/api/v1/credentials \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```
**Respon Sukses (200 OK):**
```json
{
  "success": true,
  "message": "Daftar kredensial provider berhasil diambil",
  "data": [
    {
      "id": "55555555-5555-5555-5555-555555555555",
      "organization_id": "22222222-2222-2222-2222-222222222222",
      "provider_id": "a1b2c3d4-0000-0000-0000-000000000002",
      "name": "Production AWS Account",
      "metadata": {
        "region": "us-east-1"
      },
      "created_at": "2026-08-23T00:00:00Z",
      "updated_at": "2026-08-23T00:00:00Z",
      "provider": {
        "id": "a1b2c3d4-0000-0000-0000-000000000002",
        "name": "Amazon Web Services",
        "slug": "aws",
        "is_active": true
      }
    }
  ]
}
```

#### `POST /api/v1/credentials`
Menyimpan kredensial provider baru dengan enkripsi otomatis AES-256-GCM pada seluruh field rahasia.
```bash
curl -X POST http://localhost:8080/api/v1/credentials \
  -H "Authorization: Bearer <ACCESS_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "provider_id": "a1b2c3d4-0000-0000-0000-000000000002",
    "name": "Production AWS Account",
    "api_key": "AKIAIOSFODNN7EXAMPLE",
    "api_secret": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
    "metadata": {
      "region": "us-east-1"
    }
  }'
```

#### `GET /api/v1/credentials/{id}`
Mengambil detail informasi dan metadata kredensial provider berdasarkan UUID.
```bash
curl -X GET http://localhost:8080/api/v1/credentials/55555555-5555-5555-5555-555555555555 \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```

#### `PUT /api/v1/credentials/{id}`
Memperbarui nama alias, region metadata, atau mengganti secret key kredensial.
```bash
curl -X PUT http://localhost:8080/api/v1/credentials/55555555-5555-5555-5555-555555555555 \
  -H "Authorization: Bearer <ACCESS_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated AWS Account Name"
  }'
```

#### `DELETE /api/v1/credentials/{id}`
Menghapus rekaman kredensial provider dari database.
```bash
curl -X DELETE http://localhost:8080/api/v1/credentials/55555555-5555-5555-5555-555555555555 \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```

#### `POST /api/v1/credentials/{id}/test`
Menguji validitas dan konektivitas kredensial secara langsung ke API provider cloud terkait.
```bash
curl -X POST http://localhost:8080/api/v1/credentials/55555555-5555-5555-5555-555555555555/test \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```
**Respon Sukses (200 OK):**
```json
{
  "success": true,
  "message": "Koneksi ke cloud provider berhasil diverifikasi",
  "data": {
    "provider": "aws",
    "status": "connected",
    "server_count": 3
  }
}
```

---

### 4.5. Endpoint Declarative Infrastructure as Code (IaC)

#### `POST /api/v1/iac/validate`
Memvalidasi sintaks dan skema semantik manifest YAML deklaratif sebelum disimpan.
```bash
curl -X POST http://localhost:8080/api/v1/iac/validate \
  -H "Authorization: Bearer <ACCESS_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "raw_yaml": "version: \"1.0\"\nresources:\n  servers:\n    - name: api-prod\n      provider: aws\n      region: us-east-1\n      instance_type: t3.micro"
  }'
```

#### `POST /api/v1/iac/configurations`
Membuat atau mendaftarkan konfigurasi manifest IaC baru.
```bash
curl -X POST http://localhost:8080/api/v1/iac/configurations \
  -H "Authorization: Bearer <ACCESS_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Production Infrastructure",
    "raw_yaml": "version: \"1.0\"\nresources:\n  servers:\n    - name: web-node\n      provider: digitalocean\n      region: sgp1\n      instance_type: s-1vcpu-1gb"
  }'
```

#### `GET /api/v1/iac/configurations`
Mengambil seluruh daftar konfigurasi manifest IaC milik organisasi.

#### `POST /api/v1/iac/configurations/{id}/plan`
Mengomputasi rencana perubahan (*Plan*) dengan membandingkan Desired State vs Actual State.
```bash
curl -X POST http://localhost:8080/api/v1/iac/configurations/33333333-3333-3333-3333-333333333333/plan \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```

#### `POST /api/v1/iac/configurations/{id}/apply`
Mengeksekusi rencana IaC yang telah dibuat dengan garansi LIFO rollback jika terjadi kegagalan.
```bash
curl -X POST http://localhost:8080/api/v1/iac/configurations/33333333-3333-3333-3333-333333333333/apply \
  -H "Authorization: Bearer <ACCESS_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "plan_id": "44444444-4444-4444-4444-444444444444"
  }'
```

#### `POST /api/v1/iac/configurations/{id}/rollback`
Mengembalikan infrastruktur ke versi snapshot state sebelumnya.
```bash
curl -X POST http://localhost:8080/api/v1/iac/configurations/33333333-3333-3333-3333-333333333333/rollback \
  -H "Authorization: Bearer <ACCESS_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "target_version": 1
  }'
```

#### `GET /api/v1/iac/configurations/{id}/states`
Mengambil riwayat snapshot state versi terdahulu berserta checksum integritas SHA-256.

---

### 4.6. Endpoint Container Orchestration & Live Stream Logs

#### `POST /api/v1/deployments`
Mendistribusikan dan meluncurkan kontainer Docker baru ke host target.
```bash
curl -X POST http://localhost:8080/api/v1/deployments \
  -H "Authorization: Bearer <ACCESS_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "app_name": "web-nginx",
    "image_tag": "nginx:alpine",
    "container_name": "caelus-web-nginx",
    "port_bindings": [{"host_port": 80, "container_port": 80, "protocol": "tcp"}],
    "environment_variables": {"NODE_ENV": "production"}
  }'
```

#### `GET /api/v1/deployments`
Mengambil daftar kontainer dan deployment aktif.

#### `GET /api/v1/deployments/{id}`
Mengambil detail dan status terkini spesifik deployment kontainer.

#### `GET /api/v1/deployments/{id}/logs`
Mengambil histori log deployment (stdout, stderr, system) dari database.

#### `GET /api/v1/deployments/{id}/logs/stream`
Endpoint live streaming Server-Sent Events (SSE) untuk konsol terminal frontend.
```bash
curl -N http://localhost:8080/api/v1/deployments/55555555-5555-5555-5555-555555555555/logs/stream \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```

#### `POST /api/v1/deployments/{id}/stop`
Menghentikan kontainer yang sedang berjalan.

#### `POST /api/v1/deployments/{id}/rollback`
Mengembalikan deployment ke image tag versi sebelumnya.

---

## 5. Keamanan & Proteksi Kredensial Sensitif (Encrypted at Rest)

Caelus Cloud menerapkan standar keamanan data tingkat perbankan dan enterprise untuk melindungi seluruh informasi sensitif pengguna:

### 5.1. Enkripsi AES-256-GCM (Encrypted at Rest)
* **Kunci API & Secret Key**: Seluruh field kredensial (seperti `api_key`, `api_secret`, dan `ssh_key`) tidak pernah disimpan dalam bentuk teks biasa (*plain text*) di dalam basis data PostgreSQL.
* **Algoritma Militer**: Menggunakan kriptografi autentikasi simetris **AES-256-GCM** (*Galois/Counter Mode*) dengan *nonce* acak 12-byte per transaksi enkripsi untuk mencegah serangan *replay* dan *tampering*.
* **Kunci Master Server**: Enkripsi menggunakan kunci simetris 32-byte (`ENCRYPTION_KEY`) yang dikonfigurasi melalui *environment variable* server backend.

### 5.2. Alur Dekripsi Dinamis di Memori (RAM-Only)
* Data rahasia hanya didekripsi di dalam memori kerja (*RAM*) secara sementara pada saat backend menjalankan operasi ke API provider eksternal (misal: saat *provisioning*, *reboot*, atau rekonsiliasi status VM).
* Setelah pemanggilan API selesai, data hasil dekripsi segera dibersihkan dari alokasi memori.

### 5.3. Perlindungan Terhadap Kebocoran Data (Data Masking)
* Endpoint REST API publik maupun privat **tidak pernah mengembalikan plaintext API Key / Secret Key** pada respons JSON (`json:"-"`).
* Basis data yang diekspor (*database dump*) tetap terlindungi secara aman karena penyerang tidak dapat membaca teks rahasia tanpa master `ENCRYPTION_KEY`.

---

## 6. Background Engine & Sinkronisasi Berkala

1. **Heartbeat Liveness Watchdog**: Memantau detak jantung telemetri agent setiap 15 detik. Mengubah status server menjadi `stopped` jika tidak ada telemetri yang diterima.
2. **Multi-Provider Sync Engine (`provSync.SyncEngine`)**: Melakukan rekonsiliasi otomatis setiap 60 detik antara status instance di cloud provider eksternal (AWS, Hetzner, DigitalOcean, Contabo) dengan basis data lokal Caelus. Memperbarui IP publik dan status daya secara otomatis jika terjadi perubahan dari konsol cloud pihak ketiga.
3. **Docker Deployment Pipeline**: Mengeksekusi pipeline container asynchronous (Pull -> Validate -> Configure -> Start -> Healthcheck) dan menyiarkan log realtime.
4. **Backup Scheduler & Retention Cleaner**: Mengevaluasi kebijakan snapshot dan backup server secara periodik serta membersihkan backup kadaluarsa.


