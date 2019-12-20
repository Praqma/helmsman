package app

import (
	"fmt"
	"os"
	"strings"
)

// release type representing Helm releases which are described in the desired state
type release struct {
	Name         string            `yaml:"name"`
	Description  string            `yaml:"description"`
	Namespace    string            `yaml:"namespace"`
	Enabled      bool              `yaml:"enabled"`
	Group        string            `yaml:"group"`
	Chart        string            `yaml:"chart"`
	Version      string            `yaml:"version"`
	ValuesFile   string            `yaml:"valuesFile"`
	ValuesFiles  []string          `yaml:"valuesFiles"`
	SecretsFile  string            `yaml:"secretsFile"`
	SecretsFiles []string          `yaml:"secretsFiles"`
	Purge        bool              `yaml:"purge"`
	Test         bool              `yaml:"test"`
	Protected    bool              `yaml:"protected"`
	Wait         bool              `yaml:"wait"`
	Priority     int               `yaml:"priority"`
	Set          map[string]string `yaml:"set"`
	SetString    map[string]string `yaml:"setString"`
	HelmFlags    []string          `yaml:"helmFlags"`
	NoHooks      bool              `yaml:"noHooks"`
	Timeout      int               `yaml:"timeout"`
}

// isReleaseConsideredToRun checks if a release is being targeted for operations as specified by user cmd flags (--group or --target)
func (r *release) isReleaseConsideredToRun() bool {
	if len(targetMap) > 0 {
		if _, ok := targetMap[r.Name]; ok {
			return true
		}
		return false
	}
	if len(groupMap) > 0 {
		if _, ok := groupMap[r.Group]; ok {
			return true
		}
		return false
	}
	return true
}

// validateRelease validates if a release inside a desired state meets the specifications or not.
// check the full specification @ https://github.com/Praqma/helmsman/docs/desired_state_spec.md
func validateRelease(appLabel string, r *release, names map[string]map[string]bool, s state) (bool, string) {
	if r.Name == "" {
		r.Name = appLabel
	}

	if names[r.Name][r.Namespace] {
		return false, "release name must be unique within a given Tiller."
	}

	if nsOverride == "" && r.Namespace == "" {
		return false, "release targeted namespace can't be empty."
	} else if nsOverride == "" && r.Namespace != "" && r.Namespace != "kube-system" && !checkNamespaceDefined(r.Namespace, s) {
		return false, "release " + r.Name + " is using namespace [ " + r.Namespace + " ] which is not defined in the Namespaces section of your desired state file." +
			" Release [ " + r.Name + " ] can't be installed in that Namespace until its defined."
	}
	_, err := os.Stat(r.Chart)
	if r.Chart == "" || os.IsNotExist(err) && !strings.ContainsAny(r.Chart, "/") {
		return false, "chart can't be empty and must be of the format: repo/chart."
	}
	if r.Version == "" {
		return false, "version can't be empty."
	}

	_, err = os.Stat(r.ValuesFile)
	if r.ValuesFile != "" && (!isOfType(r.ValuesFile, []string{".yaml", ".yml", ".json"}) || err != nil) {
		return false, fmt.Sprintf("valuesFile must be a valid relative (from dsf file) file path for a yaml file, or can be left empty (provided path resolved to %q).", r.ValuesFile)
	} else if r.ValuesFile != "" && len(r.ValuesFiles) > 0 {
		return false, "valuesFile and valuesFiles should not be used together."
	} else if len(r.ValuesFiles) > 0 {
		for i, filePath := range r.ValuesFiles {
			if _, pathErr := os.Stat(filePath); !isOfType(filePath, []string{".yaml", ".yml", ".json"}) || pathErr != nil {
				return false, fmt.Sprintf("valuesFiles must be valid relative (from dsf file) file paths for a yaml file; path at index %d provided path resolved to %q.", i, filePath)
			}
		}
	}

	_, err = os.Stat(r.SecretsFile)
	if r.SecretsFile != "" && (!isOfType(r.SecretsFile, []string{".yaml", ".yml", ".json"}) || err != nil) {
		return false, fmt.Sprintf("secretsFile must be a valid relative (from dsf file) file path for a yaml file, or can be left empty (provided path resolved to %q).", r.SecretsFile)
	} else if r.SecretsFile != "" && len(r.SecretsFiles) > 0 {
		return false, "secretsFile and secretsFiles should not be used together."
	} else if len(r.SecretsFiles) > 0 {
		for _, filePath := range r.SecretsFiles {
			if i, pathErr := os.Stat(filePath); !isOfType(filePath, []string{".yaml", ".yml", ".json"}) || pathErr != nil {
				return false, fmt.Sprintf("secretsFiles must be valid relative (from dsf file) file paths for a yaml file; path at index %d provided path resolved to %q.", i, filePath)
			}
		}
	}

	if r.Priority != 0 && r.Priority > 0 {
		return false, "priority can only be 0 or negative value, positive values are not allowed."
	}

	if names[r.Name] == nil {
		names[r.Name] = make(map[string]bool)
	}
	names[r.Name][r.Namespace] = true

	// add $$ escaping for $ strings
	os.Setenv("HELMSMAN_DOLLAR", "$")
	for k, v := range r.Set {
		if strings.Contains(v, "$") {
			if os.ExpandEnv(strings.Replace(v, "$$", "${HELMSMAN_DOLLAR}", -1)) == "" {
				return false, "env var [ " + v + " ] is not set, but is wanted to be passed for [ " + k + " ] in [[ " + r.Name + " ]]"
			}
		}
	}

	return true, ""
}

// overrideNamespace overrides a release defined namespace with a new given one
func overrideNamespace(r *release, newNs string) {
	log.Info("Overriding namespace for app:  " + r.Name)
	r.Namespace = newNs
}

// print prints the details of the release
func (r release) print() {
	fmt.Println("")
	fmt.Println("\tname : ", r.Name)
	fmt.Println("\tdescription : ", r.Description)
	fmt.Println("\tnamespace : ", r.Namespace)
	fmt.Println("\tenabled : ", r.Enabled)
	fmt.Println("\tchart : ", r.Chart)
	fmt.Println("\tversion : ", r.Version)
	fmt.Println("\tvaluesFile : ", r.ValuesFile)
	fmt.Println("\tvaluesFiles : ", strings.Join(r.ValuesFiles, ","))
	fmt.Println("\tpurge : ", r.Purge)
	fmt.Println("\ttest : ", r.Test)
	fmt.Println("\tprotected : ", r.Protected)
	fmt.Println("\twait : ", r.Wait)
	fmt.Println("\tpriority : ", r.Priority)
	fmt.Println("\tno-hooks : ", r.NoHooks)
	fmt.Println("\ttimeout : ", r.Timeout)
	fmt.Println("\tvalues to override from env:")
	printMap(r.Set, 2)
	fmt.Println("------------------- ")
}
