# Caelus Cloud - Host Agent Daemon (`caelus-agent`)

Daemon pengumpul metrik telemetri sistem host dan status Docker container yang berjalan secara mandiri dan ringan pada VPS / VM pengguna.

---

## 1. Fitur Utama
- **Metrik Sistem Lokal**: Utilisasi CPU (delta ticks), alokasi RAM (`/proc/meminfo`), kapasitas dan ruang disk (`syscall.Statfs`), throughput jaringan (`/proc/net/dev`), load average, serta uptime sistem.
- **Inspeksi Docker Native**: Pemantauan status container dan penggunaan resource via Unix Socket (`/var/run/docker.sock`) tanpa SDK eksternal berat.
- **Transport Aman**: Pengiriman berkala ke Caelus API Control Plane via HTTPS dengan autentikasi header dan *exponential backoff retry*.
- **Efisiensi Tinggi**: Binary tunggal independen berukuran ringan (< 10MB) dengan konsumsi memori minimal (< 15MB RSS).

---

## 2. Konfigurasi Lingkungan

Konfigurasi diatur melalui variabel lingkungan (*environment variables*):

| Variabel | Wajib | Default | Keterangan |
| :--- | :---: | :--- | :--- |
| `SERVER_ID` | Ya | - | UUID server yang terdaftar di Caelus Cloud |
| `AGENT_SECRET` | Ya | - | Kunci otentikasi rahasia agent |
| `API_ENDPOINT` | Tidak | `http://localhost:8080` | URL basis endpoint Caelus API |
| `COLLECTION_INTERVAL_SEC` | Tidak | `15` | Interval siklus telemetri (detik) |
| `DOCKER_SOCKET_PATH` | Tidak | `/var/run/docker.sock` | Path Unix domain socket Docker daemon |
| `TLS_SKIP_VERIFY` | Tidak | `false` | Abaikan validasi TLS (hanya development) |
| `LOG_LEVEL` | Tidak | `info` | Level log (`debug`, `info`, `warn`, `error`) |

---

## 3. Kompilasi & Menjalankan Agent

### A. Menjalankan Mode Development
```bash
export SERVER_ID="11111111-2222-3333-4444-555555555555"
export AGENT_SECRET="your_agent_secret_key"
go run ./cmd
```

### B. Melakukan Kompilasi Binary Produksi
```bash
go build -ldflags="-s -w" -o bin/caelus-agent ./cmd
./bin/caelus-agent
```

### C. Menjalankan Suite Pengujian Otomatis
```bash
go test -v ./tests/...
```
