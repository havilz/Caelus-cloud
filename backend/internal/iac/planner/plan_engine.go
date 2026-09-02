package planner

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
)

// Engine bertanggung jawab untuk membandingkan Desired State (manifest YAML) dengan Actual State (snapshot database/provider).
type Engine struct{}

// NewEngine membuat instance baru IaC Plan Engine.
func NewEngine() *Engine {
	return &Engine{}
}

// GeneratePlan menghasilkan domain.IaCPlan terstruktur berisi daftar diff dan ringkasan aksi.
func (e *Engine) GeneratePlan(configID uuid.UUID, targetVersion int, desired *domain.DeclarativeManifest, currentState *domain.IaCState) (*domain.IaCPlan, error) {
	var changes []domain.IaCChange
	summary := domain.IaCSummary{}

	var currentManifest domain.DeclarativeManifest
	if currentState != nil && len(currentState.StateData) > 0 {
		dataBytes, err := json.Marshal(currentState.StateData)
		if err == nil {
			_ = json.Unmarshal(dataBytes, &currentManifest)
		}
	}

	// 1. Evaluate Servers
	serverChanges := e.diffServers(desired.Servers, currentManifest.Servers)
	changes = append(changes, serverChanges...)

	// 2. Evaluate Storages
	storageChanges := e.diffStorages(desired.Storages, currentManifest.Storages)
	changes = append(changes, storageChanges...)

	// 3. Evaluate Containers
	containerChanges := e.diffContainers(desired.Containers, currentManifest.Containers)
	changes = append(changes, containerChanges...)

	// 4. Evaluate Rules
	ruleChanges := e.diffRules(desired.Rules, currentManifest.Rules)
	changes = append(changes, ruleChanges...)

	// Compute summary counts
	for _, c := range changes {
		summary.Total++
		switch c.Action {
		case domain.ActionCreate:
			summary.Create++
		case domain.ActionUpdate:
			summary.Update++
		case domain.ActionDelete:
			summary.Delete++
		case domain.ActionNoOp:
			summary.NoOp++
		}
	}

	plan := &domain.IaCPlan{
		ID:              uuid.New(),
		ConfigurationID: configID,
		TargetVersion:   targetVersion,
		Changes:         changes,
		Summary:         summary,
		Status:          domain.IaCStatusPlanned,
	}

	return plan, nil
}

func (e *Engine) diffServers(desired []domain.ServerSpec, current []domain.ServerSpec) []domain.IaCChange {
	var changes []domain.IaCChange
	currentMap := make(map[string]domain.ServerSpec)
	for _, s := range current {
		currentMap[s.Name] = s
	}

	desiredMap := make(map[string]bool)
	for _, d := range desired {
		desiredMap[d.Name] = true
		curr, exists := currentMap[d.Name]

		// Format transparansi status provider (H-2: BYOS / Local Host vs Cloud Provider Driver)
		providerMode := "Cloud Provider Driver"
		pLower := strings.ToLower(d.Provider)
		if pLower == "" || pLower == "custom" || pLower == "custom_vps" || pLower == "byos" || pLower == "local" {
			providerMode = "BYOS / Local Host Agent"
		} else {
			providerMode = fmt.Sprintf("Cloud Provider Driver (%s)", d.Provider)
		}

		afterMap := structToMap(d)
		afterMap["provider_mode"] = providerMode

		if !exists {
			changes = append(changes, domain.IaCChange{
				ResourceType: domain.ResourceTypeServer,
				ResourceName: d.Name,
				Action:       domain.ActionCreate,
				After:        afterMap,
				Reason:       fmt.Sprintf("Server '%s' will be provisioned using %s", d.Name, providerMode),
			})
		} else {
			currMap := structToMap(curr)
			currMap["provider_mode"] = providerMode
			changedFields := findChangedFields(currMap, afterMap)
			if len(changedFields) > 0 {
				changes = append(changes, domain.IaCChange{
					ResourceType:  domain.ResourceTypeServer,
					ResourceName:  d.Name,
					Action:        domain.ActionUpdate,
					Before:        currMap,
					After:         afterMap,
					ChangedFields: changedFields,
					Reason:        fmt.Sprintf("Server '%s' (%s) configuration will be updated: %v", d.Name, providerMode, changedFields),
				})
			} else {
				changes = append(changes, domain.IaCChange{
					ResourceType: domain.ResourceTypeServer,
					ResourceName: d.Name,
					Action:       domain.ActionNoOp,
					Before:       currMap,
					After:        afterMap,
					Reason:       fmt.Sprintf("[%s] No changes detected", providerMode),
				})
			}
		}
	}

	for _, c := range current {
		if !desiredMap[c.Name] {
			cMap := structToMap(c)
			changes = append(changes, domain.IaCChange{
				ResourceType: domain.ResourceTypeServer,
				ResourceName: c.Name,
				Action:       domain.ActionDelete,
				Before:       cMap,
				Reason:       fmt.Sprintf("Server '%s' is removed from desired manifest and will be terminated", c.Name),
			})
		}
	}

	return changes
}

func (e *Engine) diffStorages(desired []domain.StorageSpec, current []domain.StorageSpec) []domain.IaCChange {
	var changes []domain.IaCChange
	currentMap := make(map[string]domain.StorageSpec)
	for _, s := range current {
		currentMap[s.Name] = s
	}

	desiredMap := make(map[string]bool)
	for _, d := range desired {
		desiredMap[d.Name] = true
		curr, exists := currentMap[d.Name]
		if !exists {
			changes = append(changes, domain.IaCChange{
				ResourceType: domain.ResourceTypeStorage,
				ResourceName: d.Name,
				Action:       domain.ActionCreate,
				After:        structToMap(d),
				Reason:       fmt.Sprintf("Object storage bucket '%s' (%s) will be created", d.Name, d.Type),
			})
		} else {
			changedFields := findChangedFields(structToMap(curr), structToMap(d))
			if len(changedFields) > 0 {
				changes = append(changes, domain.IaCChange{
					ResourceType:  domain.ResourceTypeStorage,
					ResourceName:  d.Name,
					Action:        domain.ActionUpdate,
					Before:        structToMap(curr),
					After:         structToMap(d),
					ChangedFields: changedFields,
					Reason:        fmt.Sprintf("Storage bucket '%s' settings will be updated", d.Name),
				})
			} else {
				changes = append(changes, domain.IaCChange{
					ResourceType: domain.ResourceTypeStorage,
					ResourceName: d.Name,
					Action:       domain.ActionNoOp,
					Before:       structToMap(curr),
					After:        structToMap(d),
					Reason:       "No changes detected",
				})
			}
		}
	}

	for _, c := range current {
		if !desiredMap[c.Name] {
			changes = append(changes, domain.IaCChange{
				ResourceType: domain.ResourceTypeStorage,
				ResourceName: c.Name,
				Action:       domain.ActionDelete,
				Before:       structToMap(c),
				Reason:       fmt.Sprintf("Storage bucket '%s' will be removed", c.Name),
			})
		}
	}

	return changes
}

func (e *Engine) diffContainers(desired []domain.ContainerSpec, current []domain.ContainerSpec) []domain.IaCChange {
	var changes []domain.IaCChange
	currentMap := make(map[string]domain.ContainerSpec)
	for _, c := range current {
		currentMap[c.Name] = c
	}

	desiredMap := make(map[string]bool)
	for _, d := range desired {
		desiredMap[d.Name] = true
		curr, exists := currentMap[d.Name]
		if !exists {
			changes = append(changes, domain.IaCChange{
				ResourceType: domain.ResourceTypeContainer,
				ResourceName: d.Name,
				Action:       domain.ActionCreate,
				After:        structToMap(d),
				Reason:       fmt.Sprintf("Container '%s' with image '%s' will be deployed", d.Name, d.Image),
			})
		} else {
			changedFields := findChangedFields(structToMap(curr), structToMap(d))
			if len(changedFields) > 0 {
				changes = append(changes, domain.IaCChange{
					ResourceType:  domain.ResourceTypeContainer,
					ResourceName:  d.Name,
					Action:        domain.ActionUpdate,
					Before:        structToMap(curr),
					After:         structToMap(d),
					ChangedFields: changedFields,
					Reason:        fmt.Sprintf("Container '%s' will be redeployed with updated specs: %v", d.Name, changedFields),
				})
			} else {
				changes = append(changes, domain.IaCChange{
					ResourceType: domain.ResourceTypeContainer,
					ResourceName: d.Name,
					Action:       domain.ActionNoOp,
					Before:       structToMap(curr),
					After:        structToMap(d),
					Reason:       "No changes detected",
				})
			}
		}
	}

	for _, c := range current {
		if !desiredMap[c.Name] {
			changes = append(changes, domain.IaCChange{
				ResourceType: domain.ResourceTypeContainer,
				ResourceName: c.Name,
				Action:       domain.ActionDelete,
				Before:       structToMap(c),
				Reason:       fmt.Sprintf("Container '%s' will be stopped and removed", c.Name),
			})
		}
	}

	return changes
}

func (e *Engine) diffRules(desired []domain.RuleSpec, current []domain.RuleSpec) []domain.IaCChange {
	var changes []domain.IaCChange
	currentMap := make(map[string]domain.RuleSpec)
	for _, r := range current {
		currentMap[r.Name] = r
	}

	desiredMap := make(map[string]bool)
	for _, d := range desired {
		desiredMap[d.Name] = true
		curr, exists := currentMap[d.Name]
		if !exists {
			changes = append(changes, domain.IaCChange{
				ResourceType: domain.ResourceTypeRule,
				ResourceName: d.Name,
				Action:       domain.ActionCreate,
				After:        structToMap(d),
				Reason:       fmt.Sprintf("Automation rule '%s' will be created", d.Name),
			})
		} else {
			changedFields := findChangedFields(structToMap(curr), structToMap(d))
			if len(changedFields) > 0 {
				changes = append(changes, domain.IaCChange{
					ResourceType:  domain.ResourceTypeRule,
					ResourceName:  d.Name,
					Action:        domain.ActionUpdate,
					Before:        structToMap(curr),
					After:         structToMap(d),
					ChangedFields: changedFields,
					Reason:        fmt.Sprintf("Automation rule '%s' will be updated", d.Name),
				})
			} else {
				changes = append(changes, domain.IaCChange{
					ResourceType: domain.ResourceTypeRule,
					ResourceName: d.Name,
					Action:       domain.ActionNoOp,
					Before:       structToMap(curr),
					After:        structToMap(d),
					Reason:       "No changes detected",
				})
			}
		}
	}

	for _, c := range current {
		if !desiredMap[c.Name] {
			changes = append(changes, domain.IaCChange{
				ResourceType: domain.ResourceTypeRule,
				ResourceName: c.Name,
				Action:       domain.ActionDelete,
				Before:       structToMap(c),
				Reason:       fmt.Sprintf("Automation rule '%s' will be deleted", c.Name),
			})
		}
	}

	return changes
}

func structToMap(val interface{}) map[string]interface{} {
	b, err := json.Marshal(val)
	if err != nil {
		return make(map[string]interface{})
	}
	var res map[string]interface{}
	_ = json.Unmarshal(b, &res)
	return res
}

func findChangedFields(before, after map[string]interface{}) []string {
	var changed []string
	for k, vAfter := range after {
		vBefore, exists := before[k]
		if !exists || !reflect.DeepEqual(vBefore, vAfter) {
			changed = append(changed, k)
		}
	}
	for k := range before {
		if _, exists := after[k]; !exists {
			changed = append(changed, k)
		}
	}
	return changed
}
