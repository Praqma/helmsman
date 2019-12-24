package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
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

type chartVersion struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	AppVersion  string `json:"app_version"`
	Description string `json:"description"`
}

// isReleaseConsideredToRun checks if a release is being targeted for operations as specified by user cmd flags (--group or --target)
func (r *release) isConsideredToRun() bool {
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
func (r *release) validate(appLabel string, names map[string]map[string]bool, s state) (bool, string) {
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

// validateReleaseCharts validates if the charts defined in a release are valid.
// Valid charts are the ones that can be found in the defined repos.
// This function uses Helm search to verify if the chart can be found or not.
func validateReleaseCharts(apps map[string]*release) error {
	wg := sync.WaitGroup{}
	c := make(chan string, len(apps))
	for app, r := range apps {
		wg.Add(1)
		go r.validateChart(app, &wg, c)
	}
	wg.Wait()
	if len(c) > 0 {
		err := <-c
		if err != "" {
			return errors.New(err)
		}
	}
	return nil
}

var versionExtractor = regexp.MustCompile(`version:\s?(.*)`)

func (r *release) validateChart(app string, wg *sync.WaitGroup, c chan string) {

	defer wg.Done()
	validateCurrentChart := true
	if !r.isConsideredToRun() {
		validateCurrentChart = false
	}
	if validateCurrentChart {
		if isLocalChart(r.Chart) {
			cmd := command{
				Cmd:         helmBin,
				Args:        []string{"inspect", "chart", r.Chart},
				Description: "Validating [ " + r.Chart + " ] chart's availability",
			}

			var output string
			var exitCode int
			if exitCode, output, _ = cmd.exec(debug, verbose); exitCode != 0 {
				maybeRepo := filepath.Base(filepath.Dir(r.Chart))
				c <- "Chart [ " + r.Chart + " ] for app [" + app + "] can't be found. Did you mean to add a repo [ " + maybeRepo + " ]?"
				return
			}
			matches := versionExtractor.FindStringSubmatch(output)
			if len(matches) == 2 {
				version := matches[1]
				if r.Version != version {
					c <- "Chart [ " + r.Chart + " ] with version [ " + r.Version + " ] is specified for " +
						"app [" + app + "] but the chart found at that path has version [ " + version + " ] which does not match."
					return
				}
			}

		} else {
			version := r.Version
			if len(version) == 0 {
				version = "*"
			}
			cmd := command{
				Cmd:         helmBin,
				Args:        []string{"search", "repo", r.Chart, "--version", version, "-l"},
				Description: "Validating [ " + r.Chart + " ] chart's version [ " + r.Version + " ] availability",
			}

			if exitCode, result, _ := cmd.exec(debug, verbose); exitCode != 0 || strings.Contains(result, "No results found") {
				c <- "Chart [ " + r.Chart + " ] with version [ " + r.Version + " ] is specified for " +
					"app [" + app + "] but was not found"
				return
			}
		}
	}
}

// getChartVersion fetches the lastest chart version matching the semantic versioning constraints.
// If chart is local, returns the given release version
func (r *release) getChartVersion() (string, string) {
	if isLocalChart(r.Chart) {
		return r.Version, ""
	}
	cmd := command{
		Cmd:         helmBin,
		Args:        []string{"search", "repo", r.Chart, "--version", r.Version, "-o", "json"},
		Description: "Getting latest chart's version " + r.Chart + "-" + r.Version + "",
	}

	var (
		exitCode int
		result   string
	)

	if exitCode, result, _ = cmd.exec(debug, verbose); exitCode != 0 {
		return "", "Chart [ " + r.Chart + " ] with version [ " + r.Version + " ] is specified but not found in the helm repositories"
	}

	chartVersions := make([]chartVersion, 0)
	if err := json.Unmarshal([]byte(result), &chartVersions); err != nil {
		log.Fatal(fmt.Sprint(err))
	}

	if len(chartVersions) < 1 {
		return "", "Chart [ " + r.Chart + " ] with version [ " + r.Version + " ] is specified but not found in the helm repositories"
	} else if len(chartVersions) > 1 {
		return "", "Multiple versions of chart [ " + r.Chart + " ] with version [ " + r.Version + " ] found in the helm repositories"
	}
	return chartVersions[0].Version, ""
}

// testRelease creates a Helm command to test a particular release.
func (r *release) test() {
	cmd := command{
		Cmd:         helmBin,
		Args:        []string{"test", "--namespace", r.Namespace, r.Name},
		Description: "Running tests for release [ " + r.Name + " ] in namespace [ " + r.Namespace + " ]",
	}
	outcome.addCommand(cmd, r.Priority, r)
	logDecision("Release [ "+r.Name+" ] in namespace [ "+r.Namespace+" ] is required to be tested during installation", r.Priority, noop)
}

// installRelease creates a Helm command to install a particular release in a particular namespace using a particular Tiller.
func (r *release) install() {
	cmd := command{
		Cmd:         helmBin,
		Args:        concat([]string{"install", r.Name, r.Chart, "--namespace", r.Namespace}, r.getValuesFiles(), []string{"--version", r.Version}, r.getSetValues(), r.getSetStringValues(), r.getWait(), r.getHelmFlags()),
		Description: "Installing release [ " + r.Name + " ] in namespace [ " + r.Namespace + " ]",
	}
	outcome.addCommand(cmd, r.Priority, r)
	logDecision("Release [ "+r.Name+" ] will be installed in [ "+r.Namespace+" ] namespace", r.Priority, create)

	if r.Test {
		r.test()
	}
}

// uninstall deletes a release from a particular Tiller in a k8s cluster
func (r *release) uninstall() {
	priority := r.Priority
	if settings.ReverseDelete {
		priority = priority * -1
	}

	cmd := command{
		Cmd:         helmBin,
		Args:        concat([]string{"uninstall", "--namespace", r.Namespace, r.Name}, getDryRunFlags()),
		Description: "Deleting release [ " + r.Name + " ] in namespace [ " + r.Namespace + " ]",
	}
	outcome.addCommand(cmd, priority, r)
	logDecision(fmt.Sprintf("release [ %s ] is desired to be DELETED.", r.Name), r.Priority, delete)
}

// diffRelease diffs an existing release with the specified values.yaml
func (r *release) diff() string {
	exitCode := 0
	msg := ""
	colorFlag := ""
	diffContextFlag := []string{}
	suppressDiffSecretsFlag := ""
	if noColors {
		colorFlag = "--no-color"
	}
	if diffContext != -1 {
		diffContextFlag = []string{"--context", strconv.Itoa(diffContext)}
	}
	if suppressDiffSecrets {
		suppressDiffSecretsFlag = "--suppress-secrets"
	}

	cmd := command{
		Cmd:         helmBin,
		Args:        concat([]string{"diff", colorFlag}, diffContextFlag, []string{suppressDiffSecretsFlag, "--namespace", r.Namespace, "upgrade", r.Name, r.Chart}, r.getValuesFiles(), []string{"--version", r.Version}, r.getSetValues(), r.getSetStringValues()),
		Description: "Diffing release [ " + r.Name + " ] in namespace [ " + r.Namespace + " ]",
	}

	if exitCode, msg, _ = cmd.exec(debug, verbose); exitCode != 0 {
		log.Fatal(fmt.Sprintf("Command returned with exit code: %d. And error message: %s ", exitCode, msg))
	} else {
		if (verbose || showDiff) && msg != "" {
			fmt.Println(msg)
		}
	}

	return msg
}

// upgradeRelease upgrades an existing release with the specified values.yaml
func (r *release) upgrade() {
	var force string
	if forceUpgrades {
		force = "--force"
	}
	cmd := command{
		Cmd:         helmBin,
		Args:        concat([]string{"upgrade", "--namespace", r.Namespace, r.Name, r.Chart}, r.getValuesFiles(), []string{"--version", r.Version, force}, r.getSetValues(), r.getSetStringValues(), r.getWait(), r.getHelmFlags()),
		Description: "Upgrading release [ " + r.Name + " ] in namespace [ " + r.Namespace + " ]",
	}

	outcome.addCommand(cmd, r.Priority, r)
}

// reInstall purge deletes a release and reinstalls it.
// This is used when moving a release to another namespace or when changing the chart used for it.
func (r *release) reInstall(rs helmRelease) {

	delCmd := command{
		Cmd:         helmBin,
		Args:        concat([]string{"delete", "--purge", r.Name}, getDryRunFlags()),
		Description: "Deleting release [ " + r.Name + " ] in namespace [ " + r.Namespace + " ]",
	}
	outcome.addCommand(delCmd, r.Priority, r)

	installCmd := command{
		Cmd:         helmBin,
		Args:        concat([]string{"install", r.Chart, "--version", r.Version, "-n", r.Name, "--namespace", r.Namespace}, r.getValuesFiles(), r.getSetValues(), r.getSetStringValues(), r.getWait(), r.getHelmFlags()),
		Description: "Installing release [ " + r.Name + " ] in namespace [ " + r.Namespace + " ]",
	}
	outcome.addCommand(installCmd, r.Priority, r)
}

// rollbackRelease evaluates if a rollback action needs to be taken for a given release.
// if the release is already deleted but from a different namespace than the one specified in input,
// it purge deletes it and create it in the specified namespace.
func (r *release) rollback(cs *currentState) {
	rs, ok := (*cs)[fmt.Sprintf("%s-%s", r.Name, r.Namespace)]
	if !ok {
		return
	}

	if r.Namespace == rs.Namespace {

		cmd := command{
			Cmd:         helmBin,
			Args:        concat([]string{"rollback", r.Name, rs.getRevision()}, r.getWait(), r.getTimeout(), r.getNoHooks(), getDryRunFlags()),
			Description: "Rolling back release [ " + r.Name + " ] in namespace [ " + r.Namespace + " ]",
		}
		outcome.addCommand(cmd, r.Priority, r)
		r.upgrade() // this is to reflect any changes in values file(s)
		logDecision("Release [ "+r.Name+" ] was deleted and is desired to be rolled back to "+
			"namespace [ "+r.Namespace+" ]", r.Priority, create)
	} else {
		r.reInstall(rs)
		logDecision("Release [ "+r.Name+" ] is deleted BUT from namespace [ "+rs.Namespace+
			" ]. Will purge delete it from there and install it in namespace [ "+r.Namespace+" ]", r.Priority, create)
		logDecision("WARNING: rolling back release [ "+r.Name+" ] from [ "+rs.Namespace+" ] to [ "+r.Namespace+
			" ] might not correctly connect to existing volumes. Check https://github.com/Praqma/helmsman/blob/master/docs/how_to/apps/moving_across_namespaces.md"+
			" for details if this release uses PV and PVC.", r.Priority, create)

	}
}

// getNoHooks returns the no-hooks flag for install/upgrade commands
func (r *release) getNoHooks() []string {
	if r.NoHooks {
		return []string{"--no-hooks"}
	}
	return []string{}
}

// getTimeout returns the timeout flag for install/upgrade commands
func (r *release) getTimeout() []string {
	if r.Timeout != 0 {
		return []string{"--timeout", strconv.Itoa(r.Timeout) + "s"}
	}
	return []string{}
}

// getValuesFiles return partial install/upgrade release command to substitute the -f flag in Helm.
func (r *release) getValuesFiles() []string {
	var fileList []string

	if r.ValuesFile != "" {
		fileList = append(fileList, r.ValuesFile)
	} else if len(r.ValuesFiles) > 0 {
		fileList = append(fileList, r.ValuesFiles...)
	}

	if r.SecretsFile != "" {
		if !helmPluginExists("secrets") {
			log.Fatal("helm secrets plugin is not installed/configured correctly. Aborting!")
		}
		if err := decryptSecret(r.SecretsFile); err != nil {
			log.Fatal(err.Error())
		}
		fileList = append(fileList, r.SecretsFile+".dec")
	} else if len(r.SecretsFiles) > 0 {
		if !helmPluginExists("secrets") {
			log.Fatal("helm secrets plugin is not installed/configured correctly. Aborting!")
		}
		for i := 0; i < len(r.SecretsFiles); i++ {
			if err := decryptSecret(r.SecretsFiles[i]); err != nil {
				log.Fatal(err.Error())
			}
			// if .dec extension is added before to the secret filename, don't add it again.
			// This happens at upgrade time (where diff and upgrade both call this function)
			if !isOfType(r.SecretsFiles[i], []string{".dec"}) {
				r.SecretsFiles[i] = r.SecretsFiles[i] + ".dec"
			}
		}
		fileList = append(fileList, r.SecretsFiles...)
	}

	fileListArgs := []string{}
	for _, file := range fileList {
		fileListArgs = append(fileListArgs, "-f", file)
	}
	return fileListArgs
}

// getSetValues returns --set params to be used with helm install/upgrade commands
func (r *release) getSetValues() []string {
	result := []string{}
	for k, v := range r.Set {
		result = append(result, "--set", k+"="+strings.Replace(v, ",", "\\,", -1)+"")
	}
	return result
}

// getSetStringValues returns --set-string params to be used with helm install/upgrade commands
func (r *release) getSetStringValues() []string {
	result := []string{}
	for k, v := range r.SetString {
		result = append(result, "--set-string", k+"="+strings.Replace(v, ",", "\\,", -1)+"")
	}
	return result
}

// getWait returns a partial helm command containing the helm wait flag (--wait) if the wait flag for the release was set to true
// Otherwise, retruns an empty string
func (r *release) getWait() []string {
	result := []string{}
	if r.Wait {
		result = append(result, "--wait")
	}
	return result
}

// getDesiredNamespace returns the namespace of a release
func (r *release) getDesiredNamespace() string {
	return r.Namespace
}

// getHelmFlags returns helm flags
func (r *release) getHelmFlags() []string {
	var flags []string

	flags = append(flags, r.HelmFlags...)
	return concat(r.getNoHooks(), r.getTimeout(), getDryRunFlags(), flags)
}

func (r *release) checkChartDepUpdate() {
	if updateDeps && isLocalChart(r.Chart) {
		if ok, err := updateChartDep(r.Chart); !ok {
			log.Fatal("helm dependency update failed: " + err)
		}
	}
}

// overrideNamespace overrides a release defined namespace with a new given one
func (r *release) overrideNamespace(newNs string) {
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
