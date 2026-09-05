package security

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	defaultAllowedRootsMu sync.RWMutex
	defaultAllowedRoots   = []string{"/var/lib/caelus/volumes", "/opt/caelus/volumes"}
)

// SetAllowedVolumeRoots sets the global allowed volume root directories.
func SetAllowedVolumeRoots(roots []string) {
	defaultAllowedRootsMu.Lock()
	defer defaultAllowedRootsMu.Unlock()

	var cleanedRoots []string
	for _, r := range roots {
		trimmed := strings.TrimSpace(r)
		if trimmed != "" {
			cleanedRoots = append(cleanedRoots, filepath.Clean(trimmed))
		}
	}
	if len(cleanedRoots) > 0 {
		defaultAllowedRoots = cleanedRoots
	}
}

// GetAllowedVolumeRoots returns a copy of the active allowed volume root directories.
func GetAllowedVolumeRoots() []string {
	defaultAllowedRootsMu.RLock()
	defer defaultAllowedRootsMu.RUnlock()

	copied := make([]string, len(defaultAllowedRoots))
	copy(copied, defaultAllowedRoots)
	return copied
}

// ValidateHostPath validates that hostPath is an absolute, canonical path strictly residing within
// one of the allowed volume root directories (strict allowlist approach). It resolves symlinks,
// prevents directory traversal (../), and disallows mounting root directories directly (Audit C-2).
func ValidateHostPath(hostPath string, customAllowedRoots ...[]string) (string, error) {
	trimmed := strings.TrimSpace(hostPath)
	if trimmed == "" {
		return "", errors.New("host_path tidak boleh kosong")
	}

	if !filepath.IsAbs(trimmed) {
		return "", fmt.Errorf("host_path '%s' harus berupa path absolut", hostPath)
	}

	// Lexically clean the requested path
	cleanPath := filepath.Clean(trimmed)

	// Determine which allowed root paths to check against
	allowedRoots := GetAllowedVolumeRoots()
	if len(customAllowedRoots) > 0 && len(customAllowedRoots[0]) > 0 {
		allowedRoots = customAllowedRoots[0]
	}

	cleanedAllowedRoots := make([]string, 0, len(allowedRoots))
	for _, root := range allowedRoots {
		cleanedRoot := filepath.Clean(strings.TrimSpace(root))
		if cleanedRoot != "" && cleanedRoot != "/" {
			cleanedAllowedRoots = append(cleanedAllowedRoots, cleanedRoot)
		}
	}

	if len(cleanedAllowedRoots) == 0 {
		return "", errors.New("tidak ada allowed volume roots yang terkonfigurasi")
	}

	// Resolve canonical path and inspect any symlinks
	canonicalPath, err := resolveCanonicalPath(cleanPath)
	if err != nil {
		return "", fmt.Errorf("gagal mengevaluasi canonical path '%s': %w", hostPath, err)
	}

	// Verify that canonicalPath strictly resides inside at least one allowed root
	for _, root := range cleanedAllowedRoots {
		canonicalRoot, err := resolveCanonicalPath(root)
		if err != nil {
			canonicalRoot = root
		}

		// Disallow mounting the root volume itself (must be a specific subpath)
		if canonicalPath == canonicalRoot {
			return "", fmt.Errorf("host_path '%s' tidak boleh merupakan root volume itu sendiri (harus subpath spesifik)", hostPath)
		}

		rootPrefix := canonicalRoot
		if !strings.HasSuffix(rootPrefix, string(filepath.Separator)) {
			rootPrefix += string(filepath.Separator)
		}

		if strings.HasPrefix(canonicalPath, rootPrefix) {
			return canonicalPath, nil
		}
	}

	return "", fmt.Errorf("host_path '%s' berada di luar direktori volume yang diizinkan (C-2 path allowlist enforcement)", hostPath)
}

// resolveCanonicalPath resolves symlinks and canonicalizes the path. If the full path does not exist on disk,
// it walks up to the deepest existing ancestor directory, resolves symlinks for that ancestor, and reconstructs
// the canonical path.
func resolveCanonicalPath(path string) (string, error) {
	cleaned := filepath.Clean(path)

	// If the path exists on disk, resolve symlinks directly
	if resolved, err := filepath.EvalSymlinks(cleaned); err == nil {
		return filepath.Clean(resolved), nil
	} else if !os.IsNotExist(err) {
		return "", err
	}

	// Path does not exist on disk yet: walk up ancestors to find the deepest existing directory
	curr := cleaned
	var missingParts []string
	for {
		parent := filepath.Dir(curr)
		if parent == curr { // Reached filesystem root "/"
			break
		}
		missingParts = append([]string{filepath.Base(curr)}, missingParts...)
		curr = parent

		if resolvedParent, err := filepath.EvalSymlinks(curr); err == nil {
			canonical := resolvedParent
			for _, part := range missingParts {
				canonical = filepath.Join(canonical, part)
			}
			return filepath.Clean(canonical), nil
		} else if !os.IsNotExist(err) {
			return "", err
		}
	}

	// If no existing ancestor was found (e.g. mock or test environment), return cleaned lexical path
	return cleaned, nil
}
