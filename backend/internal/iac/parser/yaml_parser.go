package parser

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"gopkg.in/yaml.v3"
)

var NameRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func ComputeYAMLHash(rawYAML string) string {
	hasher := sha256.New()
	hasher.Write([]byte(strings.TrimSpace(rawYAML)))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (p *Parser) Parse(rawYAML string) (*domain.DeclarativeManifest, []domain.IaCValidationError) {
	var validationErrors []domain.IaCValidationError

	if strings.TrimSpace(rawYAML) == "" {
		return nil, []domain.IaCValidationError{
			{
				Line:    1,
				Column:  1,
				Field:   "document",
				Message: "YAML configuration cannot be empty",
			},
		}
	}

	var manifest domain.DeclarativeManifest
	decoder := yaml.NewDecoder(strings.NewReader(rawYAML))
	if err := decoder.Decode(&manifest); err != nil {
		validationErrors = append(validationErrors, domain.IaCValidationError{
			Line:    extractYAMLErrorLine(err.Error()),
			Column:  1,
			Field:   "yaml_syntax",
			Message: fmt.Sprintf("Syntax error: %v", err),
		})
		return nil, validationErrors
	}

	errs := p.ValidateManifest(&manifest)
	if len(errs) > 0 {
		return nil, errs
	}

	return &manifest, nil
}

func (p *Parser) ValidateManifest(manifest *domain.DeclarativeManifest) []domain.IaCValidationError {
	var errs []domain.IaCValidationError

	if manifest.Version == "" {
		manifest.Version = "v1"
	}

	serverNames := make(map[string]bool)
	for i, s := range manifest.Servers {
		if strings.TrimSpace(s.Name) == "" {
			errs = append(errs, domain.IaCValidationError{
				Line:    i + 1,
				Field:   fmt.Sprintf("servers[%d].name", i),
				Message: "Server name is required",
			})
		} else {
			if !NameRegex.MatchString(s.Name) {
				errs = append(errs, domain.IaCValidationError{
					Line:    i + 1,
					Field:   fmt.Sprintf("servers[%d].name", i),
					Message: fmt.Sprintf("Invalid server name '%s': must be lowercase alphanumeric and hyphens", s.Name),
				})
			}
			if serverNames[s.Name] {
				errs = append(errs, domain.IaCValidationError{
					Line:    i + 1,
					Field:   fmt.Sprintf("servers[%d].name", i),
					Message: fmt.Sprintf("Duplicate server name '%s'", s.Name),
				})
			}
			serverNames[s.Name] = true
		}

		if s.Provider == "" {
			errs = append(errs, domain.IaCValidationError{
				Line:    i + 1,
				Field:   fmt.Sprintf("servers[%d].provider", i),
				Message: fmt.Sprintf("Provider is required for server '%s'", s.Name),
			})
		}
	}

	storageNames := make(map[string]bool)
	for i, st := range manifest.Storages {
		if strings.TrimSpace(st.Name) == "" {
			errs = append(errs, domain.IaCValidationError{
				Line:    i + 1,
				Field:   fmt.Sprintf("storages[%d].name", i),
				Message: "Storage name is required",
			})
		} else {
			if !NameRegex.MatchString(st.Name) {
				errs = append(errs, domain.IaCValidationError{
					Line:    i + 1,
					Field:   fmt.Sprintf("storages[%d].name", i),
					Message: fmt.Sprintf("Invalid storage name '%s': must be lowercase alphanumeric and hyphens", st.Name),
				})
			}
			if storageNames[st.Name] {
				errs = append(errs, domain.IaCValidationError{
					Line:    i + 1,
					Field:   fmt.Sprintf("storages[%d].name", i),
					Message: fmt.Sprintf("Duplicate storage name '%s'", st.Name),
				})
			}
			storageNames[st.Name] = true
		}

		if st.Type == "" {
			errs = append(errs, domain.IaCValidationError{
				Line:    i + 1,
				Field:   fmt.Sprintf("storages[%d].type", i),
				Message: fmt.Sprintf("Storage type is required for '%s' (e.g. s3, r2, local)", st.Name),
			})
		}
	}

	containerNames := make(map[string]bool)
	for i, c := range manifest.Containers {
		if strings.TrimSpace(c.Name) == "" {
			errs = append(errs, domain.IaCValidationError{
				Line:    i + 1,
				Field:   fmt.Sprintf("containers[%d].name", i),
				Message: "Container name is required",
			})
		} else {
			if containerNames[c.Name] {
				errs = append(errs, domain.IaCValidationError{
					Line:    i + 1,
					Field:   fmt.Sprintf("containers[%d].name", i),
					Message: fmt.Sprintf("Duplicate container name '%s'", c.Name),
				})
			}
			containerNames[c.Name] = true
		}

		if strings.TrimSpace(c.Image) == "" {
			errs = append(errs, domain.IaCValidationError{
				Line:    i + 1,
				Field:   fmt.Sprintf("containers[%d].image", i),
				Message: fmt.Sprintf("Container image is required for '%s'", c.Name),
			})
		}
	}

	return errs
}

func extractYAMLErrorLine(errMsg string) int {
	re := regexp.MustCompile(`line (\d+)`)
	matches := re.FindStringSubmatch(errMsg)
	if len(matches) > 1 {
		var line int
		_, _ = fmt.Sscanf(matches[1], "%d", &line)
		if line > 0 {
			return line
		}
	}
	return 1
}
