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
	Name         string                 `yaml:"name"`
	Description  string                 `yaml:"description"`
	Namespace    string                 `yaml:"namespace"`
	Enabled      bool                   `yaml:"enabled"`
	Group        string                 `yaml:"group"`
	Chart        string                 `yaml:"chart"`
	Version      string                 `yaml:"version"`
	ValuesFile   string                 `yaml:"valuesFile"`
	ValuesFiles  []string               `yaml:"valuesFiles"`
	SecretsFile  string                 `yaml:"secretsFile"`
	SecretsFiles []string               `yaml:"secretsFiles"`
	Test         bool                   `yaml:"test"`
	Protected    bool                   `yaml:"protected"`
	Wait         bool                   `yaml:"wait"`
	Priority     int                    `yaml:"priority"`
	Set          map[string]string      `yaml:"set"`
	SetString    map[string]string      `yaml:"setString"`
	SetFile      map[string]string      `yaml:"setFile"`
	HelmFlags    []string               `yaml:"helmFlags"`
	NoHooks      bool                   `yaml:"noHooks"`
	Timeout      int                    `yaml:"timeout"`
	Hooks        map[string]interface{} `yaml:"hooks"`
	MaxHistory   int                    `yaml:"maxHistory"`
}

type chartVersion struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	AppVersion  string `json:"app_version"`
	Description string `json:"description"`
}

func (r *release) key() string {
	return fmt.Sprintf("%s-%s", r.Name, r.Namespace)
}

// isReleaseConsideredToRun checks if a release is being targeted for operations as specified by user cmd flags (--group or --target)
func (r *release) isConsideredToRun(s *state) bool {
	if len(s.TargetMap) > 0 {
		if _, ok := s.TargetMap[r.Name]; ok {
			return true
		}
		return false
	}
	return true
}

// validate validates if a release inside a desired state meets the specifications or not.
// check the full specification @ https://github.com/Praqma/helmsman/docs/desired_state_spec.md
func (r *release) validate(appLabel string, names map[string]map[string]bool, s *state) error {
	if r.Name == "" {
		r.Name = appLabel
	}

	if names[r.Name][r.Namespace] {
		return errors.New("release name must be unique within a given namespace")
	}

	if flags.nsOverride == "" && r.Namespace == "" {
		return errors.New("release targeted namespace can't be empty")
	} else if flags.nsOverride == "" && r.Namespace != "" && r.Namespace != "kube-system" && !s.isNamespaceDefined(r.Namespace) {
		return errors.New("release " + r.Name + " is using namespace [ " + r.Namespace + " ] which is not defined in the Namespaces section of your desired state file." +
			" Release [ " + r.Name + " ] can't be installed in that Namespace until its defined.")
	}
	_, err := os.Stat(r.Chart)
	if r.Chart == "" || os.IsNotExist(err) && !strings.ContainsAny(r.Chart, "/") {
		return errors.New("chart can't be empty and must be of the format: repo/chart")
	}
	if r.Version == "" {
		return errors.New("version can't be empty")
	}

	if r.ValuesFile != "" && len(r.ValuesFiles) > 0 {
		return errors.New("valuesFile and valuesFiles should not be used together")
	} else if r.ValuesFile != "" {
		if err := isValidFile(r.ValuesFile, []string{".yaml", ".yml", ".json"}); err != nil {
			return fmt.Errorf(err.Error())
		}
	} else if len(r.ValuesFiles) > 0 {
		for _, filePath := range r.ValuesFiles {
			if err := isValidFile(filePath, []string{".yaml", ".yml", ".json"}); err != nil {
				return fmt.Errorf(err.Error())
			}
		}
	}

	if r.SecretsFile != "" && len(r.SecretsFiles) > 0 {
		return errors.New("secretsFile and secretsFiles should not be used together")
	} else if r.SecretsFile != "" {
		if err := isValidFile(r.SecretsFile, []string{".yaml", ".yml", ".json"}); err != nil {
			return fmt.Errorf(err.Error())
		}
	} else if len(r.SecretsFiles) > 0 {
		for _, filePath := range r.SecretsFiles {
			if err := isValidFile(filePath, []string{".yaml", ".yml", ".json"}); err != nil {
				return fmt.Errorf(err.Error())
			}
		}
	}

	if r.Priority != 0 && r.Priority > 0 {
		return errors.New("priority can only be 0 or negative value, positive values are not allowed")
	}

	if (len(r.Hooks)) != 0 {
		if ok, errorMsg := validateHooks(r.Hooks); !ok {
			return fmt.Errorf(errorMsg)
		}
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
				return errors.New("env var [ " + v + " ] is not set, but is wanted to be passed for [ " + k + " ] in [[ " + r.Name + " ]]")
			}
		}
	}

	return nil
}

// validateHooks validates that hook files exist and of YAML type
func validateHooks(hooks map[string]interface{}) (bool, string) {
	for key, value := range hooks {
		switch key {
		case "preInstall", "postInstall", "preUpgrade", "postUpgrade", "preDelete", "postDelete":
			if err := isValidFile(value.(string), []string{".yaml", ".yml"}); err != nil {
				return false, err.Error()
			}
		case "successCondition", "successTimeout", "deleteOnSuccess":
			continue
		default:
			return false, key + " is an Invalid hook type."
		}
	}
	return true, ""
}

// validateReleaseCharts validates if the charts defined in a release are valid.
// Valid charts are the ones that can be found in the defined repos.
// This function uses Helm search to verify if the chart can be found or not.
func validateReleaseCharts(s *state) error {
	var fail bool
	var apps map[string]*release
	wg := sync.WaitGroup{}
	sem := make(chan struct{}, resourcePool)
	if len(s.TargetMap) > 0 {
		apps = s.TargetApps
	} else {
		apps = s.Apps
	}
	c := make(chan string, len(apps))
	for app, r := range apps {
		sem <- struct{}{}
		wg.Add(1)
		go func(r *release, app string) {
			defer func() {
				wg.Done()
				<-sem
			}()
			r.validateChart(app, s, c)
		}(r, app)
	}
	wg.Wait()
	close(c)
	for err := range c {
		if err != "" {
			fail = true
			log.Error(err)
		}
	}
	if fail {
		return errors.New("chart validation failed")
	}
	return nil
}

var versionExtractor = regexp.MustCompile(`[\n]version:\s?(.*)`)

// validateChart validates if chart with the same name and version as specified in the DSF exists
func (r *release) validateChart(app string, s *state, c chan string) {

	validateCurrentChart := true
	if !r.isConsideredToRun(s) {
		validateCurrentChart = false
	}
	if validateCurrentChart {
		if isLocalChart(r.Chart) {
			cmd := helmCmd([]string{"inspect", "chart", r.Chart}, "Validating [ "+r.Chart+" ] chart's availability")

			result := cmd.exec()
			if result.code != 0 {
				maybeRepo := filepath.Base(filepath.Dir(r.Chart))
				c <- "Chart [ " + r.Chart + " ] for app [" + app + "] can't be found. Inspection returned error: \"" +
					strings.TrimSpace(result.errors) + "\" -- If this is not a local chart, add the repo [ " + maybeRepo + " ] in your helmRepos stanza."
				return
			}
			matches := versionExtractor.FindStringSubmatch(result.output)
			if len(matches) == 2 {
				version := strings.Trim(matches[1], `'"`)
				if strings.Trim(r.Version, `'"`) != version {
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
			cmd := helmCmd([]string{"search", "repo", r.Chart, "--version", version, "-l"}, "Validating [ "+r.Chart+" ] chart's version [ "+r.Version+" ] availability")

			if result := cmd.exec(); result.code != 0 || strings.Contains(result.output, "No results found") {
				c <- "Chart [ " + r.Chart + " ] with version [ " + r.Version + " ] is specified for " +
					"app [" + app + "] but was not found. If this is not a local chart, define its helm repo in the helmRepo stanza in your DSF."
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
	cmd := helmCmd([]string{"search", "repo", r.Chart, "--version", r.Version, "-o", "json"}, "Getting latest chart's version "+r.Chart+"-"+r.Version+"")

	result := cmd.exec()
	if result.code != 0 {
		return "", "Chart [ " + r.Chart + " ] with version [ " + r.Version + " ] is specified but not found in the helm repositories"
	}

	chartVersions := make([]chartVersion, 0)
	if err := json.Unmarshal([]byte(result.output), &chartVersions); err != nil {
		log.Fatal(fmt.Sprint(err))
	}

	filteredChartVersions := make([]chartVersion, 0)
	for _, chart := range chartVersions {
		if chart.Name == r.Chart {
			filteredChartVersions = append(filteredChartVersions, chart)
		}
	}

	if len(filteredChartVersions) < 1 {
		return "", "Chart [ " + r.Chart + " ] with version [ " + r.Version + " ] is specified but not found in the helm repositories"
	} else if len(filteredChartVersions) > 1 {
		return "", "Multiple versions of chart [ " + r.Chart + " ] with version [ " + r.Version + " ] found in the helm repositories"
	}
	return filteredChartVersions[0].Version, ""
}

// testRelease creates a Helm command to test a particular release.
func (r *release) test(p *plan) {
	cmd := helmCmd(r.getHelmArgsFor("test"), "Running tests for release [ "+r.Name+" ] in namespace [ "+r.Namespace+" ]")
	p.addCommand(cmd, r.Priority, r, []command{}, []command{})
}

// installRelease creates a Helm command to install a particular release in a particular namespace using a particular Tiller.
func (r *release) install(p *plan) {

	before, after := r.checkHooks("install", p)
	cmd := helmCmd(r.getHelmArgsFor("install"), "Install release [ "+r.Name+" ] version [ "+r.Version+" ] in namespace [ "+r.Namespace+" ]")
	p.addCommand(cmd, r.Priority, r, before, after)

	if r.Test {
		r.test(p)
	}
}

// uninstall uninstalls a release
func (r *release) uninstall(p *plan, optionalNamespace ...string) {
	ns := r.Namespace
	if len(optionalNamespace) > 0 {
		ns = optionalNamespace[0]
	}
	priority := r.Priority
	if settings.ReverseDelete {
		priority = priority * -1
	}

	before, after := r.checkHooks("delete", p, ns)

	cmd := helmCmd(r.getHelmArgsFor("uninstall", ns), "Delete release [ "+r.Name+" ] in namespace [ "+ns+" ]")
	p.addCommand(cmd, priority, r, before, after)

}

// diffRelease diffs an existing release with the specified values.yaml
func (r *release) diff() string {
	colorFlag := ""
	diffContextFlag := []string{}
	suppressDiffSecretsFlag := "--suppress-secrets"
	if flags.noColors {
		colorFlag = "--no-color"
	}
	if flags.diffContext != -1 {
		diffContextFlag = []string{"--context", strconv.Itoa(flags.diffContext)}
	}

	cmd := helmCmd(concat([]string{"diff", colorFlag, suppressDiffSecretsFlag}, diffContextFlag, r.getHelmArgsFor("diff")), "Diffing release [ "+r.Name+" ] in namespace [ "+r.Namespace+" ]")

	result := cmd.exec()
	if result.code != 0 {
		log.Fatal(fmt.Sprintf("Command returned with exit code: %d. And error message: %s ", result.code, result.errors))
	} else {
		if (flags.verbose || flags.showDiff) && result.output != "" {
			fmt.Println(result.output)
		}
	}

	return result.output
}

// upgradeRelease upgrades an existing release with the specified values.yaml
func (r *release) upgrade(p *plan) {

	before, after := r.checkHooks("upgrade", p)

	cmd := helmCmd(r.getHelmArgsFor("upgrade"), "Upgrade release [ "+r.Name+" ] to version [ "+r.Version+" ] in namespace [ "+r.Namespace+" ]")

	p.addCommand(cmd, r.Priority, r, before, after)

}

// reInstall uninstalls a release and reinstalls it.
// This is used when moving a release to another namespace or when changing the chart used for it.
// When the release is being moved to another namespace, the optionalOldNamespace is used to provide
// the namespace from which the release is deleted.
func (r *release) reInstall(p *plan, optionalOldNamespace ...string) {
	oldNamespace := ""
	if len(optionalOldNamespace) > 0 {
		oldNamespace = optionalOldNamespace[0]
	}
	r.uninstall(p, oldNamespace)
	r.install(p)
}

// rollbackRelease evaluates if a rollback action needs to be taken for a given release.
// if the release is already deleted but from a different namespace than the one specified in input,
// it purge deletes it and create it in the specified namespace.
func (r *release) rollback(cs *currentState, p *plan) {
	rs, ok := cs.releases[r.key()]
	if !ok {
		return
	}

	if r.Namespace == rs.Namespace {

		cmd := helmCmd(concat([]string{"rollback", r.Name, rs.getRevision()}, r.getWait(), r.getTimeout(), r.getNoHooks(), flags.getDryRunFlags()), "Rolling back release [ "+r.Name+" ] in namespace [ "+r.Namespace+" ]")
		p.addCommand(cmd, r.Priority, r, []command{}, []command{})
		r.upgrade(p) // this is to reflect any changes in values file(s)
		p.addDecision("Release [ "+r.Name+" ] was deleted and is desired to be rolled back to "+
			"namespace [ "+r.Namespace+" ]", r.Priority, create)
	} else {
		r.reInstall(p)
		p.addDecision("Release [ "+r.Name+" ] is deleted BUT from namespace [ "+rs.Namespace+
			" ]. Will purge delete it from there and install it in namespace [ "+r.Namespace+" ]", r.Priority, create)
		p.addDecision("WARNING: rolling back release [ "+r.Name+" ] from [ "+rs.Namespace+" ] to [ "+r.Namespace+
			" ] might not correctly connect to existing volumes. Check https://github.com/Praqma/helmsman/blob/master/docs/how_to/apps/moving_across_namespaces.md"+
			" for details if this release uses PV and PVC.", r.Priority, create)
	}
}

// label applies Helmsman specific labels to Helm's state resources (secrets/configmaps)
func (r *release) label() {
	if r.Enabled {
		storageBackend := settings.StorageBackend

		cmd := kubectl([]string{"label", storageBackend, "-n", r.Namespace, "-l", "owner=helm,name=" + r.Name, "MANAGED-BY=HELMSMAN", "NAMESPACE=" + r.Namespace, "HELMSMAN_CONTEXT=" + curContext, "--overwrite"}, "Applying Helmsman labels to [ "+r.Name+" ] release")

		result := cmd.exec()
		if result.code != 0 {
			log.Fatal(result.errors)
		}
	}
}

// isProtected checks if a release is protected or not.
// A protected is release is either: a) deployed in a protected namespace b) flagged as protected in the desired state file
// Any release in a protected namespace is protected by default regardless of its flag
// returns true if a release is protected, false otherwise
func (r *release) isProtected(cs *currentState, s *state) bool {
	// if the release does not exist in the cluster, it is not protected
	if ok := cs.releaseExists(r, ""); !ok {
		return false
	}
	if s.Namespaces[r.Namespace].Protected || r.Protected {
		return true
	}
	return false
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

	if r.SecretsFile != "" || len(r.SecretsFiles) > 0 {
		if settings.EyamlEnabled {
			if !toolExists("eyaml") {
				log.Fatal("hiera-eyaml is not installed/configured correctly. Aborting!")
			}
		} else {
			if !helmPluginExists("secrets") {
				log.Fatal("helm secrets plugin is not installed/configured correctly. Aborting!")
			}
		}
	}
	if r.SecretsFile != "" {
		if err := decryptSecret(r.SecretsFile); err != nil {
			log.Fatal(err.Error())
		}
		fileList = append(fileList, r.SecretsFile+".dec")
	} else if len(r.SecretsFiles) > 0 {
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

// getSetFileValues returns --set-file params to be used with helm install/upgrade commands
func (r *release) getSetFileValues() []string {
	result := []string{}
	for k, v := range r.SetFile {
		result = append(result, "--set-file", k+"="+strings.Replace(v, ",", "\\,", -1)+"")
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

// getMaxHistory returns the max-history flag for upgrade commands
func (r *release) getMaxHistory() []string {
	if r.MaxHistory != 0 {
		return []string{"--history-max", strconv.Itoa(r.MaxHistory)}
	}
	return []string{}
}

// getHelmFlags returns helm flags
func (r *release) getHelmFlags() []string {
	var flgs []string
	var force string
	if flags.forceUpgrades {
		force = "--force"
	}

	flgs = append(flgs, r.HelmFlags...)
	return concat(r.getNoHooks(), r.getWait(), r.getTimeout(), r.getMaxHistory(), flags.getDryRunFlags(), []string{force}, flgs)
}

// getHelmArgsFor returns helm arguments for a specific helm operation
func (r *release) getHelmArgsFor(action string, optionalNamespaceOverride ...string) []string {
	ns := r.Namespace
	if len(optionalNamespaceOverride) > 0 {
		ns = optionalNamespaceOverride[0]
	}
	switch action {
	case "install", "upgrade":
		return concat([]string{"upgrade", r.Name, r.Chart, "--install", "--version", r.Version, "--namespace", r.Namespace}, r.getValuesFiles(), r.getSetValues(), r.getSetStringValues(), r.getSetFileValues(), r.getHelmFlags())
	case "diff":
		return concat([]string{"upgrade", r.Name, r.Chart, "--version", r.Version, "--namespace", r.Namespace}, r.getValuesFiles(), r.getSetValues(), r.getSetStringValues(), r.getSetFileValues())
	case "uninstall":
		return concat([]string{action, "--namespace", ns, r.Name}, flags.getDryRunFlags())
	default:
		return []string{action, "--namespace", ns, r.Name}
	}
}

func (r *release) checkChartDepUpdate() {
	if flags.updateDeps && isLocalChart(r.Chart) {
		if err := updateChartDep(r.Chart); err != nil {
			log.Fatal("helm dependency update failed: " + err.Error())
		}
	}
}

// overrideNamespace overrides a release defined namespace with a new given one
func (r *release) overrideNamespace(newNs string) {
	log.Info("Overriding namespace for app:  " + r.Name)
	r.Namespace = newNs
}

// inheritHooks passes global hooks config from the state to the release hooks if they are unset
// release hooks override the global ones
func (r *release) inheritHooks(s *state) {
	if len(s.Settings.GlobalHooks) != 0 {
		if len(r.Hooks) == 0 {
			r.Hooks = s.Settings.GlobalHooks
		} else {
			for key := range s.Settings.GlobalHooks {
				if _, ok := r.Hooks[key]; !ok {
					r.Hooks[key] = s.Settings.GlobalHooks[key]
				}
			}
		}
	}
}

// inheritMaxHistory passes global max history from the state to the release if it is unset
func (r *release) inheritMaxHistory(s *state) {
	if s.Settings.GlobalMaxHistory != 0 {
		if r.MaxHistory == 0 {
			r.MaxHistory = s.Settings.GlobalMaxHistory
		}
	}
}

// checkHooks checks if a hook of certain type exists and creates its command
// if success condition for the hook is defined, a "kubectl wait" command is created
// returns two slices of before and after commands
func (r *release) checkHooks(hookType string, p *plan, optionalNamespace ...string) ([]command, []command) {
	ns := r.Namespace
	if len(optionalNamespace) > 0 {
		ns = optionalNamespace[0]
	}
	var beforeCommands []command
	var afterCommands []command
	switch hookType {
	case "install":
		{
			if _, ok := r.Hooks["preInstall"]; ok {
				beforeCommands = append(beforeCommands, kubectl([]string{"apply", "-n", ns, "-f", r.Hooks["preInstall"].(string), flags.getKubeDryRunFlag("apply")}, "Apply pre-install manifest "+r.Hooks["preInstall"].(string)))
				if wait, cmds := r.shouldWaitForHook(r.Hooks["preInstall"].(string), "pre-install", ns); wait {
					beforeCommands = append(beforeCommands, cmds...)
				}
			}
			if _, ok := r.Hooks["postInstall"]; ok {
				afterCommands = append(afterCommands, kubectl([]string{"apply", "-n", ns, "-f", r.Hooks["postInstall"].(string), flags.getKubeDryRunFlag("apply")}, "Apply post-install manifest "+r.Hooks["postInstall"].(string)))
				if wait, cmds := r.shouldWaitForHook(r.Hooks["postInstall"].(string), "post-install", ns); wait {
					afterCommands = append(afterCommands, cmds...)
				}
			}
		}
	case "upgrade":
		{
			if _, ok := r.Hooks["preUpgrade"]; ok {
				beforeCommands = append(beforeCommands, kubectl([]string{"apply", "-n", ns, "-f", r.Hooks["preUpgrade"].(string), flags.getKubeDryRunFlag("apply")}, "Apply pre-upgrade manifest "+r.Hooks["preUpgrade"].(string)))
				if wait, cmds := r.shouldWaitForHook(r.Hooks["preUpgrade"].(string), "pre-upgrade", ns); wait {
					beforeCommands = append(beforeCommands, cmds...)
				}
			}
			if _, ok := r.Hooks["postUpgrade"]; ok {
				afterCommands = append(afterCommands, kubectl([]string{"apply", "-n", ns, "-f", r.Hooks["postUpgrade"].(string), flags.getKubeDryRunFlag("apply")}, "Apply post-upgrade manifest "+r.Hooks["postUpgrade"].(string)))
				if wait, cmds := r.shouldWaitForHook(r.Hooks["postUpgrade"].(string), "post-upgrade", ns); wait {
					afterCommands = append(afterCommands, cmds...)
				}
			}
		}
	case "delete":
		{
			if _, ok := r.Hooks["preDelete"]; ok {
				beforeCommands = append(beforeCommands, kubectl([]string{"apply", "-n", ns, "-f", r.Hooks["preDelete"].(string), flags.getKubeDryRunFlag("apply")}, "Apply pre-delete manifest "+r.Hooks["preDelete"].(string)))
				if wait, cmds := r.shouldWaitForHook(r.Hooks["preDelete"].(string), "pre-delete", ns); wait {
					beforeCommands = append(beforeCommands, cmds...)
				}
			}
			if _, ok := r.Hooks["postDelete"]; ok {
				afterCommands = append(afterCommands, kubectl([]string{"apply", "-n", ns, "-f", r.Hooks["postDelete"].(string), flags.getKubeDryRunFlag("apply")}, "Apply post-delete manifest "+r.Hooks["postDelete"].(string)))
				if wait, cmds := r.shouldWaitForHook(r.Hooks["postDelete"].(string), "post-delete", ns); wait {
					afterCommands = append(afterCommands, cmds...)
				}
			}
		}
	}
	return beforeCommands, afterCommands
}

// shouldWaitForHook checks if there is a success condition to wait for after applying a hook
// returns a boolean and the wait command if applicable
func (r *release) shouldWaitForHook(hookFile string, hookType string, namespace string) (bool, []command) {
	var cmds []command
	if flags.dryRun {
		return false, cmds
	} else if _, ok := r.Hooks["successCondition"]; ok {
		timeoutFlag := ""
		if _, ok := r.Hooks["successTimeout"]; ok {
			timeoutFlag = "--timeout=" + r.Hooks["successTimeout"].(string)
		}
		cmds = append(cmds, kubectl([]string{"wait", "-n", namespace, "-f", hookFile, "--for=condition=" + r.Hooks["successCondition"].(string), timeoutFlag}, "Wait for "+hookType+" : "+hookFile))
		if _, ok := r.Hooks["deleteOnSuccess"]; ok && r.Hooks["deleteOnSuccess"].(bool) {
			cmds = append(cmds, kubectl([]string{"delete", "-n", namespace, "-f", hookFile}, "Delete "+hookType+" : "+hookFile))
		}
		return true, cmds
	}
	return false, cmds
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
	fmt.Println("\tSuccessCondition : ", r.Hooks["successCondition"])
	fmt.Println("\tSuccessTimeout : ", r.Hooks["successTimeout"])
	fmt.Println("\tDeleteOnSuccess : ", r.Hooks["deleteOnSuccess"])
	fmt.Println("\tpreInstall : ", r.Hooks["preInstall"])
	fmt.Println("\tpostInstall : ", r.Hooks["postInstall"])
	fmt.Println("\tpreUpgrade : ", r.Hooks["preUpgrade"])
	fmt.Println("\tpostUpgrade : ", r.Hooks["postUpgrade"])
	fmt.Println("\tpreDelete : ", r.Hooks["preDelete"])
	fmt.Println("\tpostDelete : ", r.Hooks["postDelete"])
	fmt.Println("\tno-hooks : ", r.NoHooks)
	fmt.Println("\ttimeout : ", r.Timeout)
	fmt.Println("\tvalues to override from env:")
	printMap(r.Set, 2)
	fmt.Println("------------------- ")
}
