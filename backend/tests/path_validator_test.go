package tests

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/internal/usecase/orchestration"
	"github.com/havilz/caelus-cloud/backend/pkg/security"
)

func TestValidateHostPath_ValidPaths(t *testing.T) {
	validCases := []string{
		"/var/lib/caelus/volumes/app1",
		"/var/lib/caelus/volumes/tenant_abc/database/data",
		"/opt/caelus/volumes/my-app",
		"/opt/caelus/volumes/redis/dump",
	}

	for _, p := range validCases {
		canonical, err := security.ValidateHostPath(p)
		if err != nil {
			t.Errorf("expected path '%s' to be valid, got error: %v", p, err)
		}
		if canonical == "" {
			t.Errorf("expected non-empty canonical path for '%s'", p)
		}
	}
}

func TestValidateHostPath_RootItself_Rejected(t *testing.T) {
	rootCases := []string{
		"/var/lib/caelus/volumes",
		"/var/lib/caelus/volumes/",
		"/opt/caelus/volumes",
		"/opt/caelus/volumes/",
	}

	for _, p := range rootCases {
		_, err := security.ValidateHostPath(p)
		if err == nil {
			t.Errorf("expected mounting root volume itself '%s' to be rejected, but it was allowed", p)
		}
	}
}

func TestValidateHostPath_SensitiveSystemDirs_Rejected(t *testing.T) {
	sensitiveCases := []string{
		"/",
		"/etc",
		"/etc/shadow",
		"/etc/passwd",
		"/root",
		"/root/.ssh/id_rsa",
		"/home",
		"/home/user",
		"/home/user/.bashrc",
		"/opt",
		"/tmp",
		"/tmp/evil.sock",
		"/srv",
		"/mnt",
		"/media",
		"/bin",
		"/sbin",
		"/usr",
		"/usr/bin",
		"/lib",
		"/lib64",
		"/boot",
		"/sys",
		"/proc",
		"/dev",
		"/run",
		"/var/run",
		"/var/run/docker.sock",
		"/var/lib/docker",
	}

	for _, p := range sensitiveCases {
		_, err := security.ValidateHostPath(p)
		if err == nil {
			t.Errorf("expected sensitive path '%s' to be rejected (C-2 escape), but it was allowed", p)
		}
	}
}

func TestValidateHostPath_PathTraversal_Rejected(t *testing.T) {
	traversalCases := []string{
		"/var/lib/caelus/volumes/../../etc/shadow",
		"/var/lib/caelus/volumes/../volumes_secret",
		"/var/lib/caelus/volumes/app/../../../root",
		"/opt/caelus/volumes/../../var/run/docker.sock",
	}

	for _, p := range traversalCases {
		_, err := security.ValidateHostPath(p)
		if err == nil {
			t.Errorf("expected path traversal '%s' to be rejected, but it was allowed", p)
		}
	}
}

func TestValidateHostPath_RelativePath_Rejected(t *testing.T) {
	relativeCases := []string{
		"volumes/app1",
		"./data",
		"../escape",
		"relative/path",
		"",
		"   ",
	}

	for _, p := range relativeCases {
		_, err := security.ValidateHostPath(p)
		if err == nil {
			t.Errorf("expected relative/empty path '%s' to be rejected, but it was allowed", p)
		}
	}
}

func TestValidateHostPath_SymlinkEscape_Rejected(t *testing.T) {
	// Setup real filesystem structure in temp directory
	tempRoot := t.TempDir()

	allowedRoot := filepath.Join(tempRoot, "allowed_volumes")
	forbiddenRoot := filepath.Join(tempRoot, "host_forbidden")

	if err := os.MkdirAll(allowedRoot, 0755); err != nil {
		t.Fatalf("failed creating allowed root: %v", err)
	}
	if err := os.MkdirAll(forbiddenRoot, 0755); err != nil {
		t.Fatalf("failed creating forbidden root: %v", err)
	}

	// Create a secret file in forbidden directory
	secretFile := filepath.Join(forbiddenRoot, "sensitive.key")
	if err := os.WriteFile(secretFile, []byte("super_secret"), 0600); err != nil {
		t.Fatalf("failed creating secret file: %v", err)
	}

	// Create symlink inside allowedRoot pointing to forbiddenRoot
	symlinkPath := filepath.Join(allowedRoot, "escape_link")
	if err := os.Symlink(forbiddenRoot, symlinkPath); err != nil {
		t.Fatalf("failed creating symlink: %v", err)
	}

	// 1. Attempting to mount the symlink directly: allowedRoot/escape_link
	_, err := security.ValidateHostPath(symlinkPath, []string{allowedRoot})
	if err == nil {
		t.Errorf("expected symlink pointing outside allowed root to be rejected, but it was allowed!")
	}

	// 2. Attempting to mount a file inside the symlink: allowedRoot/escape_link/sensitive.key
	targetFileViaSymlink := filepath.Join(symlinkPath, "sensitive.key")
	_, err = security.ValidateHostPath(targetFileViaSymlink, []string{allowedRoot})
	if err == nil {
		t.Errorf("expected file access via symlink escape to be rejected, but it was allowed!")
	}

	// 3. Legitimate volume directory created inside allowedRoot: must succeed
	legitVolume := filepath.Join(allowedRoot, "app_data")
	if err := os.MkdirAll(legitVolume, 0755); err != nil {
		t.Fatalf("failed creating legit volume: %v", err)
	}

	canonical, err := security.ValidateHostPath(legitVolume, []string{allowedRoot})
	if err != nil {
		t.Errorf("expected legitimate subpath to be allowed, got error: %v", err)
	}
	if canonical == "" {
		t.Error("expected non-empty canonical path for legitimate subpath")
	}
}

func TestValidateHostPath_CustomAllowedRoots(t *testing.T) {
	customRoots := []string{"/mnt/nvme-pool/volumes", "/storage/nfs/volumes"}

	// Valid subpaths within custom roots
	validCustom := []string{
		"/mnt/nvme-pool/volumes/app1-cache",
		"/storage/nfs/volumes/shared-media",
	}

	for _, p := range validCustom {
		_, err := security.ValidateHostPath(p, customRoots)
		if err != nil {
			t.Errorf("expected custom allowed path '%s' to be valid, got: %v", p, err)
		}
	}

	// Default roots should now be rejected when custom roots are explicitly passed
	_, err := security.ValidateHostPath("/var/lib/caelus/volumes/app1", customRoots)
	if err == nil {
		t.Error("expected default path to be rejected when explicit custom roots are provided")
	}
}

func TestDeploymentUsecase_VolumeBindingValidation(t *testing.T) {
	ctx := context.Background()
	repo := NewMockDeploymentRepo()
	uc := orchestration.NewUseCase(repo, nil)

	orgID := uuid.New()

	// 1. Unsafe host path: /etc/shadow
	unsafeReq := domain.DeploymentRequest{
		AppName:       "unsafe-app",
		ImageTag:      "nginx:latest",
		ContainerName: "unsafe-container",
		VolumeBindings: []domain.VolumeBinding{
			{HostPath: "/etc/shadow", ContainerPath: "/etc/shadow", Mode: "ro"},
		},
	}

	_, err := uc.CreateDeployment(ctx, orgID, unsafeReq)
	if err == nil {
		t.Fatal("expected CreateDeployment to fail with unsafe volume binding /etc/shadow, but it succeeded!")
	}

	// 2. Safe host path within allowed volume roots: /var/lib/caelus/volumes/myapp/data
	safeReq := domain.DeploymentRequest{
		AppName:       "safe-app",
		ImageTag:      "nginx:latest",
		ContainerName: "safe-container",
		VolumeBindings: []domain.VolumeBinding{
			{HostPath: "/var/lib/caelus/volumes/myapp/data", ContainerPath: "/var/data", Mode: "rw"},
		},
	}

	dep, err := uc.CreateDeployment(ctx, orgID, safeReq)
	if err != nil {
		t.Fatalf("expected CreateDeployment to succeed with safe volume binding, got error: %v", err)
	}

	if dep == nil || len(dep.VolumeBindings) != 1 {
		t.Fatal("expected 1 volume binding on created deployment")
	}

	if dep.VolumeBindings[0].HostPath != "/var/lib/caelus/volumes/myapp/data" {
		t.Errorf("expected canonical host path, got: %s", dep.VolumeBindings[0].HostPath)
	}
}
