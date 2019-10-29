package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var outcome plan
var settings config

// makePlan creates a plan of the actions needed to make the desired state come true.
func makePlan(s *state) *plan {
	settings = s.Settings
	outcome = createPlan()
	buildState()

	for _, r := range s.Apps {
		checkChartDepUpdate(r)
		decide(r, s)
	}

	return &outcome
}

// decide makes a decision about what commands (actions) need to be executed
// to make a release section of the desired state come true.
func decide(r *release, s *state) {
	// check for presence in defined targets or groups
	if !r.isReleaseConsideredToRun() {
		logDecision("release [ "+r.Name+" ] is ignored due to passed constraints. Skipping.", r.Priority, ignored)
		return
	}

	if destroy {
		if ok := isReleaseExisting(r, ""); ok {
			deleteRelease(r)
			return
		}
	}

	if !r.Enabled {
		if ok := isReleaseExisting(r, ""); ok {

			if isProtected(r) {

				logDecision("release [ "+r.Name+" ] in namespace [ "+r.Namespace+" ] is PROTECTED. Operations are not allowed on this release until "+
					"protection is removed.", r.Priority, noop)
				return
			}
			deleteRelease(r)
			return
		}
		logDecision("release [ "+r.Name+" ] is disabled. Skipping.", r.Priority, noop)
		return

	} else {
		if ok := isReleaseExisting(r, "deployed"); ok {
			if !isProtected(r) {
				inspectUpgradeScenario(r) // upgrade or move

			} else {
				logDecision("release [ "+r.Name+" ] in namespace [ "+r.Namespace+" ] is PROTECTED. Operations are not allowed on this release until "+
					"you remove its protection.", r.Priority, noop)
			}

		} else if ok := isReleaseExisting(r, "deleted"); ok {
			if !isProtected(r) {

				rollbackRelease(r) // rollback

			} else {
				logDecision("release [ "+r.Name+" ] in namespace [ "+r.Namespace+" ] is PROTECTED. Operations are not allowed on this release until "+
					"you remove its protection.", r.Priority, noop)
			}

		} else if ok := isReleaseExisting(r, "failed"); ok {

			if !isProtected(r) {

				logDecision("release [ "+r.Name+" ] in namespace [ "+r.Namespace+" ] is in FAILED state. Upgrade is scheduled!", r.Priority, change)
				upgradeRelease(r)

			} else {
				logDecision("release [ "+r.Name+" ] in namespace [ "+r.Namespace+" ] is PROTECTED. Operations are not allowed on this release until "+
					"you remove its protection.", r.Priority, noop)
			}
		} else {

			installRelease(r)

		}

	}

}

// testRelease creates a Helm command to test a particular release.
func testRelease(r *release) {
	cmd := command{
		Cmd:         helmBin,
		Args:        []string{"test", "--namespace", r.Namespace, r.Name},
		Description: "running tests for release [ "+r.Name+" ] in namespace [[ "+r.Namespace+" ]]",
	}
	outcome.addCommand(cmd, r.Priority, r)
	logDecision("release [ "+r.Name+" ] in namespace [ "+r.Namespace+" ] is required to be tested when installed. Got it!",  r.Priority, noop)
}

// installRelease creates a Helm command to install a particular release in a particular namespace using a particular Tiller.
func installRelease(r *release) {
	cmd := command{
		Cmd:         helmBin,
		Args:        concat([]string{"install", r.Name, r.Chart, "--namespace", r.Namespace}, getValuesFiles(r), []string{"--version", r.Version}, getSetValues(r), getSetStringValues(r), getWait(r), getHelmFlags(r)),
		Description: "installing release [ "+r.Name+" ] in namespace [[ "+r.Namespace+" ]]",
	}
	outcome.addCommand(cmd, r.Priority, r)
	logDecision("release [ "+r.Name+" ] is not installed. Will install it in namespace [[ "+r.Namespace+" ]]", r.Priority, create)

	if r.Test {
		testRelease(r)
	}
}

// rollbackRelease evaluates if a rollback action needs to be taken for a given release.
// if the release is already deleted but from a different namespace than the one specified in input,
// it purge deletes it and create it in the specified namespace.
func rollbackRelease(r *release) {
	rs, ok := currentState[fmt.Sprintf("%s-%s", r.Name, r.Namespace)]
	if !ok {
		return
	}

	if r.Namespace == rs.Namespace {

		cmd := command{
			Cmd:         helmBin,
			Args:        concat([]string{"rollback", r.Name, getReleaseRevision(rs)}, getWait(r), getTimeout(r), getNoHooks(r), getDryRunFlags()),
			Description: "rolling back release [ "+r.Name+" ] in namespace [[ "+r.Namespace+" ]]",
		}
		outcome.addCommand(cmd, r.Priority, r)
		upgradeRelease(r) // this is to reflect any changes in values file(s)
		logDecision("release [ "+r.Name+" ] is currently deleted and is desired to be rolledback to "+
			"namespace [[ "+r.Namespace+" ]] . It will also be upgraded in case values have changed.", r.Priority, create)
	} else {
		reInstallRelease(r, rs)
		logDecision("release [ "+r.Name+" ] is deleted BUT from namespace [[ "+rs.Namespace+
			" ]]. Will purge delete it from there and install it in namespace [[ "+r.Namespace+" ]]", r.Priority, create)
		logDecision("WARNING: rolling back release [ "+r.Name+" ] from [[ "+rs.Namespace+" ]] to [[ "+r.Namespace+
			" ]] might not correctly connect to existing volumes. Check https://github.com/Praqma/helmsman/blob/master/docs/how_to/apps/moving_across_namespaces.md"+
			" for details if this release uses PV and PVC.", r.Priority, create)

	}
}

// deleteRelease deletes a release from a particular Tiller in a k8s cluster
func deleteRelease(r *release) {
	priority := r.Priority
	if settings.ReverseDelete == true {
		priority = priority * -1
	}

	cmd := command{
		Cmd:         helmBin,
		Args:        concat([]string{"delete", "--namespace", r.Namespace, r.Name}, getDryRunFlags()),
		Description: "deleting release [ "+r.Name+" ] in namespace [[ "+r.Namespace+" ]]",
	}
	outcome.addCommand(cmd, priority, r)
	logDecision(fmt.Sprintf("release [ %s ] is desired to be DELETED.", r.Name), r.Priority, delete)
}

// inspectUpgradeScenario evaluates if a release should be upgraded.
// - If the release is already in the same namespace specified in the input,
// it will be upgraded using the values file specified in the release info.
// - If the release is already in the same namespace specified in the input but is using a different chart,
// it will be purge deleted and installed in the same namespace using the new chart.
// - If the release is NOT in the same namespace specified in the input,
// it will be purge deleted and installed in the new namespace.
func inspectUpgradeScenario(r *release) {

	rs, ok := currentState[fmt.Sprintf("%s-%s", r.Name, r.Namespace)]
	if !ok {
		return
	}

	if r.Namespace == rs.Namespace {

		version, msg := getChartVersion(r)
		if msg != "" {
			logError(msg)
			return
		}
		r.Version = version

		if extractChartName(r.Chart) == getReleaseChartName(rs) && r.Version != getReleaseChartVersion(rs) {
			// upgrade
			diffRelease(r)
			upgradeRelease(r)
			logDecision("release [ "+r.Name+" ] will be upgraded.", r.Priority, change)

		} else if extractChartName(r.Chart) != getReleaseChartName(rs) {
			reInstallRelease(r, rs)
			logDecision("release [ "+r.Name+" ] is desired to use a new Chart [ "+r.Chart+
				" ]. Delete of the current release will be planned and new chart will be installed in namespace [[ "+
				r.Namespace+" ]]", r.Priority, change)
		} else {
			if diff := diffRelease(r); diff != "" {
				upgradeRelease(r)
				logDecision("release [ "+r.Name+" ] is installed and has some changes to apply. "+
					"I will upgrade it!", r.Priority, change)
			} else {
				logDecision("release [ "+r.Name+" ] is enabled and has no changes to apply.", r.Priority, noop)
			}
		}
	} else {
		reInstallRelease(r, rs)
		logDecision("release [ "+r.Name+" ] is desired to be enabled in a new namespace [[ "+r.Namespace+
			" ]]. I am planning a purge delete of the current release from namespace [[ "+rs.Namespace+" ]] "+
			"and will install it for you in namespace [[ "+r.Namespace+" ]]", r.Priority, change)
		logDecision("WARNING: moving release [ "+r.Name+" ] from [[ "+rs.Namespace+" ]] to [[ "+r.Namespace+
			" ]] might not correctly connect to existing volumes. Check https://github.com/Praqma/helmsman/blob/master/docs/how_to/move_charts_across_namespaces.md"+
			" for details if this release uses PV and PVC.", r.Priority, change)
	}
}

// diffRelease diffs an existing release with the specified values.yaml
func diffRelease(r *release) string {
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
		Args:        concat([]string{"diff", colorFlag}, diffContextFlag, []string{suppressDiffSecretsFlag, "--namespace", r.Namespace, "upgrade", r.Name, r.Chart}, getValuesFiles(r), []string{"--version", r.Version}, getSetValues(r), getSetStringValues(r)),
		Description: "diffing release [ "+r.Name+" ] in namespace [[ "+r.Namespace+" ]]",
	}

	if exitCode, msg, _ = cmd.exec(debug, verbose); exitCode != 0 {
		logError(fmt.Sprintf("Command returned with exit code: %d. And error message: %s ", exitCode, msg))
	} else {
		if (verbose || showDiff) && msg != "" {
			fmt.Println(msg)
		}
	}

	return msg
}

// upgradeRelease upgrades an existing release with the specified values.yaml
func upgradeRelease(r *release) {
	var force string
	if forceUpgrades {
		force = "--force"
	}
	cmd := command{
		Cmd:         helmBin,
		Args:        concat([]string{"upgrade", "--namespace", r.Namespace, r.Name, r.Chart}, getValuesFiles(r), []string{"--version", r.Version, force}, getSetValues(r), getSetStringValues(r), getWait(r), getHelmFlags(r)),
		Description: "upgrading release [ "+r.Name+" ] in namespace [[ "+r.Namespace+" ]]",
	}

	outcome.addCommand(cmd, r.Priority, r)
}

// reInstallRelease purge deletes a release and reinstalls it.
// This is used when moving a release to another namespace or when changing the chart used for it.
func reInstallRelease(r *release, rs releaseState) {

	delCmd := command{
		Cmd:         helmBin,
		Args:        concat([]string{"delete", "--purge", r.Name}, getDryRunFlags()),
		Description: "deleting release [ "+r.Name+" ] in namespace [[ "+r.Namespace+" ]]",
	}
	outcome.addCommand(delCmd, r.Priority, r)

	installCmd := command{
		Cmd:         helmBin,
		Args:        concat([]string{"install", r.Chart, "--version", r.Version, "-n", r.Name, "--namespace", r.Namespace}, getValuesFiles(r), getSetValues(r), getSetStringValues(r), getWait(r), getHelmFlags(r)),
		Description: "installing release [ "+r.Name+" ] in namespace [[ "+r.Namespace+" ]]",
	}
	outcome.addCommand(installCmd, r.Priority, r)
}

// logDecision adds the decisions made to the plan.
// Depending on the debug flag being set or not, it will either log the the decision to output or not.
func logDecision(decision string, priority int, decisionType decisionType) {

	outcome.addDecision(decision, priority, decisionType)

}

// extractChartName extracts the Helm chart name from full chart name in the desired state.
// example: it extracts "chartY" from "repoX/chartY" and "chartZ" from "c:\charts\chartZ"
func extractChartName(releaseChart string) string {

	m := chartNameExtractor.FindStringSubmatch(releaseChart)
	if len(m) == 2 {
		return m[1]
	}

	return ""
}

var chartNameExtractor = regexp.MustCompile(`[\\/]([^\\/]+)$`)

// getNoHooks returns the no-hooks flag for install/upgrade commands
func getNoHooks(r *release) []string {
	if r.NoHooks {
		return []string{"--no-hooks"}
	}
	return []string{}
}

// getTimeout returns the timeout flag for install/upgrade commands
func getTimeout(r *release) []string {
	if r.Timeout != 0 {
		return []string{"--timeout", strconv.Itoa(r.Timeout) + "s"}
	}
	return []string{}
}

// getValuesFiles return partial install/upgrade release command to substitute the -f flag in Helm.
func getValuesFiles(r *release) []string {
	var fileList []string

	if r.ValuesFile != "" {
		fileList = append(fileList, r.ValuesFile)
	} else if len(r.ValuesFiles) > 0 {
		fileList = append(fileList, r.ValuesFiles...)
	}

	if r.SecretsFile != "" {
		if !helmPluginExists("secrets") {
			logError("helm secrets plugin is not installed/configured correctly. Aborting!")
		}
		if ok := decryptSecret(r.SecretsFile); !ok {
			logError("Failed to decrypt secret file" + r.SecretsFile)
		}
		fileList = append(fileList, r.SecretsFile+".dec")
	} else if len(r.SecretsFiles) > 0 {
		if !helmPluginExists("secrets") {
			logError("helm secrets plugin is not installed/configured correctly. Aborting!")
		}
		for i := 0; i < len(r.SecretsFiles); i++ {
			if ok := decryptSecret(r.SecretsFiles[i]); !ok {
				logError("Failed to decrypt secret file" + r.SecretsFiles[i])
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
func getSetValues(r *release) []string {
	result := []string{}
	for k, v := range r.Set {
		result = append(result, "--set", k+"="+strings.Replace(v, ",", "\\,", -1)+"")
	}
	return result
}

// getSetStringValues returns --set-string params to be used with helm install/upgrade commands
func getSetStringValues(r *release) []string {
	result := []string{}
	for k, v := range r.SetString {
		result = append(result, "--set-string", k+"="+strings.Replace(v, ",", "\\,", -1)+"")
	}
	return result
}

// getWait returns a partial helm command containing the helm wait flag (--wait) if the wait flag for the release was set to true
// Otherwise, retruns an empty string
func getWait(r *release) []string {
	result := []string{}
	if r.Wait {
		result = append(result, "--wait")
	}
	return result
}

// getDesiredNamespace returns the namespace of a release
func getDesiredNamespace(r *release) string {

	return r.Namespace
}

// getCurrentNamespaceProtection returns the protection state for the namespace where a release is currently installed.
// It returns true if a namespace is defined as protected in the desired state file, false otherwise.
func getCurrentNamespaceProtection(rs releaseState) bool {

	return s.Namespaces[rs.Namespace].Protected
}

// isProtected checks if a release is protected or not.
// A protected is release is either: a) deployed in a protected namespace b) flagged as protected in the desired state file
// Any release in a protected namespace is protected by default regardless of its flag
// returns true if a release is protected, false otherwise
func isProtected(r *release) bool {

	// if the release does not exist in the cluster, it is not protected
	if ok := isReleaseExisting(r, ""); !ok {
		return false
	}

	if s.Namespaces[r.Namespace].Protected || r.Protected {
		return true
	}

	return false

}

// getDryRunFlags returns dry-run flag
func getDryRunFlags() []string {
	if dryRun {
		return []string{"--dry-run", "--debug"}
	}
	return []string{}
}

// getHelmFlags returns helm flags
func getHelmFlags(r *release) []string {
	var flags []string

	for _, flag := range r.HelmFlags {
		flags = append(flags, flag)
	}
	return concat(getNoHooks(r), getTimeout(r), getDryRunFlags(), flags)
}

func checkChartDepUpdate(r *release) {
	if updateDeps && isLocalChart(r.Chart) {
		if ok, err := updateChartDep(r.Chart); !ok {
			logError("helm dependency update failed: " + err)
		}
	}
}
