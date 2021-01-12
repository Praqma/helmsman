package app

import (
	"fmt"
	"strings"
)

const (
	preInstall  = "preInstall"
	postInstall = "postInstall"
	preUpgrade  = "preUpgrade"
	postUpgrade = "postUpgrade"
	preDelete   = "preDelete"
	postDelete  = "postDelete"
	test        = "test"
)

// TODO: Create different types for Command and Manifest hooks
// with methods for getting their commands for the plan
type hookCmd struct {
	Command
	Type string
}

func (h *hookCmd) getAnnotationKey() (string, error) {
	if h.Type == "" {
		return "", fmt.Errorf("no type specified")
	}
	return "helmsman/" + h.Type, nil
}

// validateHooks validates that hook files exist and are of correct type
func validateHooks(hooks map[string]interface{}) error {
	validFiles := []string{".yaml", ".yml", ".json"}
	for key, value := range hooks {
		switch key {
		case preInstall, postInstall, preUpgrade, postUpgrade, preDelete, postDelete:
			hook := value.(string)
			if !isOfType(hook, validFiles) && ToolExists(strings.Fields(hook)[0]) {
				return nil
			}
			if err := isValidFile(hook, validFiles); err != nil {
				return fmt.Errorf("invalid hook manifest: %w", err)
			}
		case "successCondition", "successTimeout", "deleteOnSuccess":
			continue
		default:
			return fmt.Errorf("%s is an Invalid hook type", key)
		}
	}
	return nil
}
