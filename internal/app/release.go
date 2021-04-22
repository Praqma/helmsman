package app

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
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
	PostRenderer string                 `yaml:"postRenderer"`
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
	disabled     bool
}

func (r *release) key() string {
	return fmt.Sprintf("%s-%s", r.Name, r.Namespace)
}

func (r *release) Disable() {
	r.disabled = true
}

// isReleaseConsideredToRun checks if a release is being targeted for operations as specified by user cmd flags (--group or --target)
func (r *release) isConsideredToRun() bool {
	if r == nil {
		return false
	}
	return !r.disabled
}

// validate validates if a release inside a desired state meets the specifications or not.
// check the full specification @ https://github.com/Praqma/helmsman/blob/master/docs/desired_state_specification.md
func (r *release) validate(appLabel string, seen map[string]map[string]bool, s *state) error {
	if seen[r.Name][r.Namespace] {
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
		if err := isValidFile(r.ValuesFile, validManifestFiles); err != nil {
			return fmt.Errorf("invalid values file: %w", err)
		}
	} else if len(r.ValuesFiles) > 0 {
		for _, filePath := range r.ValuesFiles {
			if err := isValidFile(filePath, validManifestFiles); err != nil {
				return fmt.Errorf("invalid values file: %w", err)
			}
		}
	}

	if r.SecretsFile != "" && len(r.SecretsFiles) > 0 {
		return errors.New("secretsFile and secretsFiles should not be used together")
	} else if r.SecretsFile != "" {
		if err := isValidFile(r.SecretsFile, validManifestFiles); err != nil {
			return fmt.Errorf("invalid secrets file: %w", err)
		}
	} else if len(r.SecretsFiles) > 0 {
		for _, filePath := range r.SecretsFiles {
			if err := isValidFile(filePath, validManifestFiles); err != nil {
				return fmt.Errorf("invalid secrets file: %w", err)
			}
		}
	}

	if r.PostRenderer != "" && !ToolExists(r.PostRenderer) {
		return fmt.Errorf("%s must be executable and available in your PATH", r.PostRenderer)
	}

	if r.Priority != 0 && r.Priority > 0 {
		return errors.New("priority can only be 0 or negative value, positive values are not allowed")
	}

	if (len(r.Hooks)) != 0 {
		if err := validateHooks(r.Hooks); err != nil {
			return err
		}
	}

	if seen[r.Name] == nil {
		seen[r.Name] = make(map[string]bool)
	}
	seen[r.Name][r.Namespace] = true

	return nil
}

// testRelease creates a Helm command to test a particular release.
func (r *release) test(afterCommands *[]hookCmd) {
	if flags.dryRun {
		log.Verbose("Dry-run, skipping tests:  " + r.Name)
		return
	}
	cmd := helmCmd(r.getHelmArgsFor("test"), "Running tests for release [ "+r.Name+" ] in namespace [ "+r.Namespace+" ]")
	*afterCommands = append([]hookCmd{{Command: cmd, Type: test}}, *afterCommands...)
}

// installRelease creates a Helm command to install a particular release in a particular namespace using a particular Tiller.
func (r *release) install(p *plan) {
	before, after := r.checkHooks("install")

	if r.Test {
		r.test(&after)
	}

	cmd := helmCmd(r.getHelmArgsFor("install"), "Install release [ "+r.Name+" ] version [ "+r.Version+" ] in namespace [ "+r.Namespace+" ]")
	p.addCommand(cmd, r.Priority, r, before, after)
}

// uninstall uninstalls a release
func (r *release) uninstall(p *plan, optionalNamespace ...string) {
	ns := r.Namespace
	if len(optionalNamespace) > 0 {
		ns = optionalNamespace[0]
	}
	priority := r.Priority
	if p.ReverseDelete {
		priority *= -1
	}

	before, after := r.checkHooks("delete", ns)

	cmd := helmCmd(r.getHelmArgsFor("uninstall", ns), "Delete release [ "+r.Name+" ] in namespace [ "+ns+" ]")
	p.addCommand(cmd, priority, r, before, after)
}

// diffRelease diffs an existing release with the specified values.yaml
func (r *release) diff() (string, error) {
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

	res, err := cmd.RetryExec(3)
	if err != nil {
		return "", fmt.Errorf("command failed: %w", err)
	}

	return res.output, nil
}

// upgradeRelease upgrades an existing release with the specified values.yaml
func (r *release) upgrade(p *plan) {
	before, after := r.checkHooks("upgrade")

	if r.Test {
		r.test(&after)
	}

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

		cmd := helmCmd(concat([]string{"rollback", r.Name, rs.getRevision()}, r.getWait(), r.getTimeout(), r.getNoHooks(), flags.getRunFlags()), "Rolling back release [ "+r.Name+" ] in namespace [ "+r.Namespace+" ]")
		p.addCommand(cmd, r.Priority, r, []hookCmd{}, []hookCmd{})
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

// mark applies Helmsman specific labels to Helm's state resources (secrets/configmaps)
func (r *release) mark(storageBackend string) {
	r.label(storageBackend, "MANAGED-BY=HELMSMAN", "NAMESPACE="+r.Namespace, "HELMSMAN_CONTEXT="+curContext)
}

// label labels Helm's state resources (secrets/configmaps)
func (r *release) label(storageBackend string, labels ...string) {
	if len(labels) == 0 {
		return
	}
	if r.Enabled {

		args := []string{"label", "--overwrite", storageBackend, "-n", r.Namespace, "-l", "owner=helm,name=" + r.Name}
		args = append(args, labels...)
		cmd := kubectl(args, "Applying Helmsman labels to [ "+r.Name+" ] release")

		if _, err := cmd.Exec(); err != nil {
			log.Fatal(err.Error())
		}
	}
}

// annotate annotates Helm's state resources (secrets/configmaps)
func (r *release) annotate(storageBackend string, annotations ...string) {
	if len(annotations) == 0 {
		return
	}
	if r.Enabled {

		args := []string{"annotate", "--overwrite", storageBackend, "-n", r.Namespace, "-l", "owner=helm,name=" + r.Name}
		args = append(args, annotations...)
		cmd := kubectl(args, "Applying Helmsman annotations to [ "+r.Name+" ] release")

		if _, err := cmd.Exec(); err != nil {
			log.Fatal(err.Error())
		}
	}
}

// isProtected checks if a release is protected or not.
// A protected is release is either: a) deployed in a protected namespace b) flagged as protected in the desired state file
// Any release in a protected namespace is protected by default regardless of its flag
// returns true if a release is protected, false otherwise
func (r *release) isProtected(cs *currentState, n *namespace) bool {
	// if the release does not exist in the cluster, it is not protected
	if ok := cs.releaseExists(r, ""); !ok {
		return false
	}
	if n.Protected || r.Protected {
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

// getSetValues returns --set params to be used with helm install/upgrade commands
func (r *release) getSetValues() []string {
	res := []string{}
	for k, v := range r.Set {
		res = append(res, "--set", k+"="+strings.ReplaceAll(v, ",", "\\,")+"")
	}
	return res
}

// getSetStringValues returns --set-string params to be used with helm install/upgrade commands
func (r *release) getSetStringValues() []string {
	res := []string{}
	for k, v := range r.SetString {
		res = append(res, "--set-string", k+"="+strings.ReplaceAll(v, ",", "\\,")+"")
	}
	return res
}

// getSetFileValues returns --set-file params to be used with helm install/upgrade commands
func (r *release) getSetFileValues() []string {
	res := []string{}
	for k, v := range r.SetFile {
		res = append(res, "--set-file", k+"="+strings.ReplaceAll(v, ",", "\\,")+"")
	}
	return res
}

// getWait returns a partial helm command containing the helm wait flag (--wait) if the wait flag for the release was set to true
// Otherwise, retruns an empty string
func (r *release) getWait() []string {
	res := []string{}
	if r.Wait {
		res = append(res, "--wait")
	}
	return res
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
	return concat(r.getNoHooks(), r.getWait(), r.getTimeout(), r.getMaxHistory(), flags.getRunFlags(), []string{force}, flgs)
}

// getPostRenderer returns the post-renderer Helm flag
func (r *release) getPostRenderer() []string {
	args := []string{}
	if r.PostRenderer != "" {
		args = append(args, "--post-renderer", r.PostRenderer)
	}
	return args
}

// getHelmArgsFor returns helm arguments for a specific helm operation
func (r *release) getHelmArgsFor(action string, optionalNamespaceOverride ...string) []string {
	ns := r.Namespace
	if len(optionalNamespaceOverride) > 0 {
		ns = optionalNamespaceOverride[0]
	}
	switch action {
	case "install", "upgrade":
		return concat([]string{"upgrade", r.Name, r.Chart, "--install", "--version", r.Version, "--namespace", r.Namespace}, r.getValuesFiles(), r.getSetValues(), r.getSetStringValues(), r.getSetFileValues(), r.getHelmFlags(), r.getPostRenderer())
	case "diff":
		return concat([]string{"upgrade", r.Name, r.Chart, "--version", r.Version, "--namespace", r.Namespace}, r.getValuesFiles(), r.getSetValues(), r.getSetStringValues(), r.getSetFileValues(), r.getPostRenderer())
	case "uninstall":
		return concat([]string{action, "--namespace", ns, r.Name}, flags.getRunFlags())
	default:
		return []string{action, "--namespace", ns, r.Name}
	}
}

func (r *release) checkChartDepUpdate() {
	if !r.isConsideredToRun() {
		return
	}
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
func (r *release) checkHooks(action string, optionalNamespace ...string) ([]hookCmd, []hookCmd) {
	ns := r.Namespace
	if len(optionalNamespace) > 0 {
		ns = optionalNamespace[0]
	}
	var beforeCmds, afterCmds []hookCmd

	switch action {
	case "install":
		{
			if _, ok := r.Hooks[preInstall]; ok {
				beforeCmds = append(beforeCmds, r.getHookCommands(preInstall, ns)...)
			}
			if _, ok := r.Hooks[postInstall]; ok {
				afterCmds = append(afterCmds, r.getHookCommands(postInstall, ns)...)
			}
		}
	case "upgrade":
		{
			if _, ok := r.Hooks[preUpgrade]; ok {
				beforeCmds = append(beforeCmds, r.getHookCommands(preUpgrade, ns)...)
			}
			if _, ok := r.Hooks[postUpgrade]; ok {
				afterCmds = append(afterCmds, r.getHookCommands(postUpgrade, ns)...)
			}
		}
	case "delete":
		{
			if _, ok := r.Hooks[preDelete]; ok {
				beforeCmds = append(beforeCmds, r.getHookCommands(preDelete, ns)...)
			}
			if _, ok := r.Hooks[postDelete]; ok {
				afterCmds = append(afterCmds, r.getHookCommands(postDelete, ns)...)
			}
		}
	}
	return beforeCmds, afterCmds
}

func (r *release) getHookCommands(hookType, ns string) []hookCmd {
	var cmds []hookCmd
	if _, ok := r.Hooks[hookType]; ok {
		hook := r.Hooks[hookType].(string)
		if err := isValidFile(hook, validManifestFiles); err == nil {
			cmd := kubectl([]string{"apply", "-n", ns, "-f", hook, flags.getKubeDryRunFlag("apply")}, "Apply "+hook+" manifest "+hookType)
			cmds = append(cmds, hookCmd{Command: cmd, Type: hookType})
			if wait, waitCmds := r.shouldWaitForHook(hook, hookType, ns); wait {
				cmds = append(cmds, waitCmds...)
			}
		} else { // shell hook
			args := strings.Fields(hook)
			cmds = append(cmds, hookCmd{
				Command: Command{
					Cmd:         args[0],
					Args:        args[1:],
					Description: fmt.Sprintf("%s shell hook", hookType),
				},
				Type: hookType,
			})
		}
	}
	return cmds
}

// shouldWaitForHook checks if there is a success condition to wait for after applying a hook
// returns a boolean and the wait command if applicable
func (r *release) shouldWaitForHook(hookFile string, hookType string, namespace string) (bool, []hookCmd) {
	var cmds []hookCmd
	if flags.dryRun {
		return false, cmds
	} else if _, ok := r.Hooks["successCondition"]; ok {
		timeoutFlag := ""
		if _, ok := r.Hooks["successTimeout"]; ok {
			timeoutFlag = "--timeout=" + r.Hooks["successTimeout"].(string)
		}
		cmd := kubectl([]string{"wait", "-n", namespace, "-f", hookFile, "--for=condition=" + r.Hooks["successCondition"].(string), timeoutFlag}, "Wait for "+hookType+" : "+hookFile)
		cmds = append(cmds, hookCmd{Command: cmd})
		if _, ok := r.Hooks["deleteOnSuccess"]; ok && r.Hooks["deleteOnSuccess"].(bool) {
			cmd = kubectl([]string{"delete", "-n", namespace, "-f", hookFile}, "Delete "+hookType+" : "+hookFile)
			cmds = append(cmds, hookCmd{Command: cmd})
		}
		return true, cmds
	}
	return false, cmds
}

// print prints the details of the release
func (r release) print() {
	fmt.Println("")
	fmt.Println("\tname: ", r.Name)
	fmt.Println("\tdescription: ", r.Description)
	fmt.Println("\tnamespace: ", r.Namespace)
	fmt.Println("\tenabled: ", r.Enabled)
	fmt.Println("\tchart: ", r.Chart)
	fmt.Println("\tversion: ", r.Version)
	fmt.Println("\tvaluesFile: ", r.ValuesFile)
	fmt.Println("\tvaluesFiles: ", strings.Join(r.ValuesFiles, ","))
	fmt.Println("\tpostRenderer: ", r.PostRenderer)
	fmt.Println("\ttest: ", r.Test)
	fmt.Println("\tprotected: ", r.Protected)
	fmt.Println("\twait: ", r.Wait)
	fmt.Println("\tpriority: ", r.Priority)
	fmt.Println("\tSuccessCondition: ", r.Hooks["successCondition"])
	fmt.Println("\tSuccessTimeout: ", r.Hooks["successTimeout"])
	fmt.Println("\tDeleteOnSuccess: ", r.Hooks["deleteOnSuccess"])
	fmt.Println("\tpreInstall: ", r.Hooks[preInstall])
	fmt.Println("\tpostInstall: ", r.Hooks[postInstall])
	fmt.Println("\tpreUpgrade: ", r.Hooks[preUpgrade])
	fmt.Println("\tpostUpgrade: ", r.Hooks[postUpgrade])
	fmt.Println("\tpreDelete: ", r.Hooks[preDelete])
	fmt.Println("\tpostDelete: ", r.Hooks[postDelete])
	fmt.Println("\tno-hooks: ", r.NoHooks)
	fmt.Println("\ttimeout: ", r.Timeout)
	fmt.Println("\tvalues to override from env:")
	printMap(r.Set, 2)
	fmt.Println("------------------- ")
}
