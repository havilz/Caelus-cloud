# Caelus Cloud - Backend API Service

Layanan Backend RESTful API untuk platform **Caelus Cloud** yang dibangun menggunakan bahasa pemrograman **Go** dengan pendekatan **Clean Architecture**, router **Chi**, dan basis data **PostgreSQL** (Supabase).

---

## 1. Arsitektur & Struktur Direktori

```text
backend/
├── cmd/
│   ├── api/          # Entry point aplikasi HTTP API server
│   └── migrate/      # Standalone CLI tool untuk migrasi skema database
├── internal/
│   ├── delivery/
│   │   └── http/     # HTTP router, middleware (auth, rbac, audit, cors, logger), dan v1 handlers
│   ├── domain/       # Entitas bisnis murni, interface kontrak, dan error types
│   ├── provider/     # Implementasi driver cloud (MockDriver) dan Driver Factory
│   ├── repository/   # PostgreSQL repository implementations (User, Org, Server, Provider, Credential, Audit)
│   └── usecase/      # Interactor logika bisnis (Auth, Server, Credential)
├── pkg/              # Paket pembantu (config, hasher, jwt, encryptor, logger, validator, response)
├── migrations/       # Berkas DDL SQL migrasi (000001, 000002, 000003)
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

### A. Menjalankan Migrasi Skema Basis Data
```powershell
cd backend
go run ./cmd/migrate -direction=up
```

### B. Menjalankan Server API (Development Mode)
```powershell
cd backend
go run ./cmd/api
```

### C. Melakukan Kompilasi dan Menjalankan Binary
```powershell
cd backend
go build -o bin/api.exe ./cmd/api
.\bin\api.exe
```

### D. Menjalankan Pengujian Otomatis (Test Suite)
Seluruh pengujian unit dan integrasi tersimpan terpusat di direktori `tests/`:
```powershell
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
