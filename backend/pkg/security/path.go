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
// one of the allowed volume root directories, preventing directory traversal and symlink escapes.
func ValidateHostPath(hostPath string, customAllowedRoots ...[]string) (string, error) {
	trimmed := strings.TrimSpace(hostPath)
	if trimmed == "" {
		return "", errors.New("host_path must not be empty")
	}

	if !filepath.IsAbs(trimmed) {
		return "", fmt.Errorf("host_path '%s' must be an absolute path", hostPath)
	}

	cleanPath := filepath.Clean(trimmed)

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
		return "", errors.New("no allowed volume roots configured")
	}

	canonicalPath, err := resolveCanonicalPath(cleanPath)
	if err != nil {
		return "", fmt.Errorf("failed to evaluate canonical path '%s': %w", hostPath, err)
	}

	for _, root := range cleanedAllowedRoots {
		canonicalRoot, err := resolveCanonicalPath(root)
		if err != nil {
			canonicalRoot = root
		}

		if canonicalPath == canonicalRoot {
			return "", fmt.Errorf("host_path '%s' must not be the volume root itself (must be a specific subpath)", hostPath)
		}

		rootPrefix := canonicalRoot
		if !strings.HasSuffix(rootPrefix, string(filepath.Separator)) {
			rootPrefix += string(filepath.Separator)
		}

		if strings.HasPrefix(canonicalPath, rootPrefix) {
			return canonicalPath, nil
		}
	}

	return "", fmt.Errorf("host_path '%s' is outside allowed volume directories (C-2 path allowlist enforcement)", hostPath)
}

// resolveCanonicalPath resolves symlinks and canonicalizes the path, walking ancestors if path does not exist.
func resolveCanonicalPath(path string) (string, error) {
	cleaned := filepath.Clean(path)

	if resolved, err := filepath.EvalSymlinks(cleaned); err == nil {
		return filepath.Clean(resolved), nil
	} else if !os.IsNotExist(err) {
		return "", err
	}

	curr := cleaned
	var missingParts []string
	for {
		parent := filepath.Dir(curr)
		if parent == curr {
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

	return cleaned, nil
}
