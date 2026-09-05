# Changelog

All notable changes to the Caelus Cloud project are documented here.

Entries follow formal conventions with explicit timestamps `[YYYY-MM-DD HH:mm:ss]`.

---

## [Unreleased]

### [2026-09-05 19:15:00] - Security Hardening: Container Escape Mitigation via Path Allowlist (Audit C-2)

- **Security Path Validation Package (`backend/pkg/security/path.go`)**:
  - Implemented `ValidateHostPath` with a strict allowlist approach (only allowing volume subpaths under `/var/lib/caelus/volumes`, `/opt/caelus/volumes`, or directories specified via `ALLOWED_VOLUME_ROOTS`).
  - Implemented canonical path resolution via `filepath.Clean` and deep ancestor symlink traversal evaluation via `filepath.EvalSymlinks` to prevent symlink escape exploits.
  - Disallowed direct mounting of volume root directories to enforce volume isolation between applications.
- **Allowed Volume Roots Configuration (`backend/pkg/config/config.go`, `.env.example`, `docker-compose.yml`, `cmd/api/main.go`)**:
  - Added `AllowedVolumeRoots` to `AppConfig` and initialized allowlists during API server bootstrap.
  - Documented `ALLOWED_VOLUME_ROOTS` in `.env.example` and forwarded it to the `api` container in `docker-compose.yml`.
- **Harmonization in Orchestration & IaC Layer (`deployment_usecase.go`, `docker_pipeline.go`, `apply_engine.go`)**:
  - Replaced legacy hardcoded denylists in deployment usecases and Docker pipeline with delegation to `security.ValidateHostPath`.
  - Secured volume bind mount parsing in IaC template execution engine.
- **Comprehensive Security Test Suite (`backend/tests/path_validator_test.go`)**:
  - Added thorough tests: rejection of sensitive system directories (`/`, `/etc`, `/root`, `/home`, `/tmp`, `/opt`, `/var/run/docker.sock`), path traversal (`../`), relative paths, and real filesystem symlink escape simulations (`t.TempDir()`).
  - Verified rejection of unsafe host path bindings during container creation via `CreateDeployment`.
