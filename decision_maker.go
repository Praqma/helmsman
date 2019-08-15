package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var outcome plan
var releases string
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
	// check for presence in defined targets
	if len(targetMap) > 0 {
		if _, ok := targetMap[r.Name]; !ok {
			logDecision("DECISION: release [ "+r.Name+" ] is ignored by target flag. Skipping.", r.Priority, noop)
			return
		}
	}

	if destroy {
		if ok, rs := helmReleaseExists(r, "DEPLOYED"); ok {
			deleteRelease(r, rs)
		}
		if ok, rs := helmReleaseExists(r, "FAILED"); ok {
			deleteRelease(r, rs)
		}
		return
	}

	// check for deletion
	if !r.Enabled {
		if ok, rs := helmReleaseExists(r, ""); ok {
			if !isProtected(r, rs) {

				// delete it
				deleteRelease(r, rs)

			} else {
				logDecision("DECISION: release [ "+r.Name+" ] in namespace [ "+r.Namespace+" ] is PROTECTED. Operations are not allowed on this release until "+
					"you remove its protection.", r.Priority, noop)
			}
		} else {
			logDecision("DECISION: release [ "+r.Name+" ] is set to be disabled but is not yet deployed. Skipping.", r.Priority, noop)
		}

	} else { // check for install/upgrade/rollback
		if ok, rs := helmReleaseExists(r, "deployed"); ok {
			if !isProtected(r, rs) {
				inspectUpgradeScenario(r, rs) // upgrade or move

			} else {
				logDecision("DECISION: release [ "+r.Name+" ] in namespace [ "+r.Namespace+" ] is PROTECTED. Operations are not allowed on this release until "+
					"you remove its protection.", r.Priority, noop)
			}

		} else if ok, rs := helmReleaseExists(r, "deleted"); ok {
			if !isProtected(r, rs) {

				rollbackRelease(r, rs) // rollback

			} else {
				logDecision("DECISION: release [ "+r.Name+" ] in namespace [ "+r.Namespace+" ] is PROTECTED. Operations are not allowed on this release until "+
					"you remove its protection.", r.Priority, noop)
			}

		} else if ok, rs := helmReleaseExists(r, "failed"); ok {

			if !isProtected(r, rs) {

				logDecision("DECISION: release [ "+r.Name+" ] in namespace [ "+r.Namespace+" ] is in FAILED state. I will upgrade it for you. Hope it gets fixed!", r.Priority, change)
				upgradeRelease(r)

			} else {
				logDecision("DECISION: release [ "+r.Name+" ] in namespace [ "+r.Namespace+" ] is PROTECTED. Operations are not allowed on this release until "+
					"you remove its protection.", r.Priority, noop)
			}
		} else {

			installRelease(r) // install a new release

		}

	}

}

// helmCommandFromConfig returns the command used to invoke helm. If configured to
// operate without a tiller it will return `helm tiller run NAMESPACE -- helm`
// where NAMESPACE is the namespace that the release is configured to use.
// If not configured to run without a tiller will just return `helm`.
func helmCommand(namespace string) string {
	if settings.Tillerless {
		return "helm tiller run " + namespace + " -- helm"
	}

	return "helm"
}

// helmCommandFromConfig calls helmCommand returning the correct way to invoke
// helm.
func helmCommandFromConfig(r *release) string {
	return helmCommand(getDesiredTillerNamespace(r))
}

// testRelease creates a Helm command to test a particular release.
func testRelease(r *release) {
	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", helmCommandFromConfig(r) + " test " + r.Name + getDesiredTillerNamespaceFlag(r) + getTLSFlags(r)},
		Description: "running tests for release [ " + r.Name + " ]",
	}
	outcome.addCommand(cmd, r.Priority, r)
	logDecision("DECISION: release [ "+r.Name+" ] in namespace [ "+r.Namespace+" ] is required to be tested when installed. Got it!", r.Priority, noop)

}

// installRelease creates a Helm command to install a particular release in a particular namespace using a particular Tiller.
func installRelease(r *release) {
	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", helmCommandFromConfig(r) + " install " + r.Chart + " -n " + r.Name + " --namespace " + r.Namespace + getValuesFiles(r) + " --version " + strconv.Quote(r.Version) + getSetValues(r) + getSetStringValues(r) + getWait(r) + getDesiredTillerNamespaceFlag(r) + getTLSFlags(r) + getHelmFlags(r)},
		Description: "installing release [ " + r.Name + " ] in namespace [[ " + r.Namespace + " ]] using Tiller in [ " + getDesiredTillerNamespace(r) + " ]",
	}
	outcome.addCommand(cmd, r.Priority, r)
	logDecision("DECISION: release [ "+r.Name+" ] is not installed. Will install it in namespace [[ "+
		r.Namespace+" ]] using Tiller in [ "+getDesiredTillerNamespace(r)+" ]", r.Priority, create)

	if r.Test {
		testRelease(r)
	}
}

// rollbackRelease evaluates if a rollback action needs to be taken for a given release.
// if the release is already deleted but from a different namespace than the one specified in input,
// it purge deletes it and create it in the specified namespace.
func rollbackRelease(r *release, rs releaseState) {

	if r.Namespace == rs.Namespace {

		cmd := command{
			Cmd:         "bash",
			Args:        []string{"-c", helmCommandFromConfig(r) + " rollback " + r.Name + " " + getReleaseRevision(rs) + getWait(r) + getDesiredTillerNamespaceFlag(r) + getTLSFlags(r) + getTimeout(r) + getNoHooks(r) + getDryRunFlags()},
			Description: "rolling back release [ " + r.Name + " ] using Tiller in [ " + getDesiredTillerNamespace(r) + " ]",
		}
		outcome.addCommand(cmd, r.Priority, r)
		upgradeRelease(r) // this is to reflect any changes in values file(s)
		logDecision("DECISION: release [ "+r.Name+" ] is currently deleted and is desired to be rolledback to "+
			"namespace [[ "+r.Namespace+" ]] . It will also be upgraded in case values have changed.", r.Priority, change)

	} else {

		reInstallRelease(r, rs)
		logDecision("DECISION: release [ "+r.Name+" ] is deleted BUT from namespace [[ "+rs.Namespace+
			" ]]. Will purge delete it from there and install it in namespace [[ "+r.Namespace+" ]]", r.Priority, change)
		logDecision("WARNING: rolling back release [ "+r.Name+" ] from [[ "+rs.Namespace+" ]] to [[ "+r.Namespace+
			" ]] might not correctly connect to existing volumes. Check https://github.com/Praqma/helmsman/blob/master/docs/how_to/move_charts_across_namespaces.md"+
			" for details if this release uses PV and PVC.", r.Priority, change)

	}
}

// deleteRelease deletes a release from a particular Tiller in a k8s cluster
func deleteRelease(r *release, rs releaseState) {
	p := ""
	purgeDesc := ""
	if r.Purge {
		p = "--purge"
		purgeDesc = "and purged!"
	}

	priority := r.Priority
	if settings.ReverseDelete == true {
		priority = priority * -1
	}

	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", helmCommandFromConfig(r) + " delete " + p + " " + r.Name + getCurrentTillerNamespaceFlag(rs) + getTLSFlags(r) + getDryRunFlags()},
		Description: "deleting release [ " + r.Name + " ] from namespace [[ " + r.Namespace + " ]] using Tiller in [ " + getDesiredTillerNamespace(r) + " ]",
	}
	outcome.addCommand(cmd, priority, r)
	logDecision("DECISION: release [ "+r.Name+" ] is desired to be deleted "+purgeDesc+". Planning this for you!", priority, delete)
}

// inspectUpgradeScenario evaluates if a release should be upgraded.
// - If the release is already in the same namespace specified in the input,
// it will be upgraded using the values file specified in the release info.
// - If the release is already in the same namespace specified in the input but is using a different chart,
// it will be purge deleted and installed in the same namespace using the new chart.
// - If the release is NOT in the same namespace specified in the input,
// it will be purge deleted and installed in the new namespace.
func inspectUpgradeScenario(r *release, rs releaseState) {

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
			logDecision("DECISION: release [ "+r.Name+" ] is desired to be upgraded. Planning this for you!", r.Priority, change)

		} else if extractChartName(r.Chart) != getReleaseChartName(rs) {
			reInstallRelease(r, rs)
			logDecision("DECISION: release [ "+r.Name+" ] is desired to use a new Chart [ "+r.Chart+
				" ]. I am planning a purge delete of the current release and will install it with the new chart in namespace [[ "+
				r.Namespace+" ]]", r.Priority, change)

		} else {
			if diff := diffRelease(r); diff != "" {
				upgradeRelease(r)
				logDecision("DECISION: release [ "+r.Name+" ] is currently enabled and have some changed parameters. "+
					"I will upgrade it!", r.Priority, change)
			} else {
				logDecision("DECISION: release [ "+r.Name+" ] is desired to be enabled and is currently enabled. "+
					"Nothing to do here!", r.Priority, noop)
			}
		}
	} else {
		reInstallRelease(r, rs)
		logDecision("DECISION: release [ "+r.Name+" ] is desired to be enabled in a new namespace [[ "+r.Namespace+
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
	diffContextFlag := ""
	suppressDiffSecretsFlag := ""
	if noColors {
		colorFlag = "--no-color "
	}
	if diffContext != -1 {
		diffContextFlag = "--context " + strconv.Itoa(diffContext) + " "
	}
	if suppressDiffSecrets {
		suppressDiffSecretsFlag = "--suppress-secrets "
	}

	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", helmCommandFromConfig(r) + " diff " + colorFlag + diffContextFlag + suppressDiffSecretsFlag + "upgrade " + r.Name + " " + r.Chart + getValuesFiles(r) + " --version " + strconv.Quote(r.Version) + " " + getSetValues(r) + getSetStringValues(r) + getDesiredTillerNamespaceFlag(r) + getTLSFlags(r)},
		Description: "diffing release [ " + r.Name + " ] using Tiller in [ " + getDesiredTillerNamespace(r) + " ]",
	}

	if exitCode, msg = cmd.exec(debug, verbose); exitCode != 0 {
		logError("Command returned with exit code: " + string(exitCode) + ". And error message: " + msg)
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
		force = " --force "
	}
	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", helmCommandFromConfig(r) + " upgrade " + r.Name + " " + r.Chart + getValuesFiles(r) + " --version " + strconv.Quote(r.Version) + force + getSetValues(r) + getSetStringValues(r) + getWait(r) + getDesiredTillerNamespaceFlag(r) + getTLSFlags(r) + getHelmFlags(r)},
		Description: "upgrading release [ " + r.Name + " ] using Tiller in [ " + getDesiredTillerNamespace(r) + " ]",
	}

	outcome.addCommand(cmd, r.Priority, r)
}

// reInstallRelease purge deletes a release and reinstalls it.
// This is used when moving a release to another namespace or when changing the chart used for it.
func reInstallRelease(r *release, rs releaseState) {

	delCmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", helmCommandFromConfig(r) + " delete --purge " + r.Name + getCurrentTillerNamespaceFlag(rs) + getTLSFlags(r) + getDryRunFlags()},
		Description: "deleting release [ " + r.Name + " ] from namespace [[ " + r.Namespace + " ]] using Tiller in [ " + getDesiredTillerNamespace(r) + " ]",
	}
	outcome.addCommand(delCmd, r.Priority, r)

	installCmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", helmCommandFromConfig(r) + " install " + r.Chart + " --version " + r.Version + " -n " + r.Name + " --namespace " + r.Namespace + getValuesFiles(r) + getSetValues(r) + getSetStringValues(r) + getWait(r) + getDesiredTillerNamespaceFlag(r) + getTLSFlags(r) + getHelmFlags(r)},
		Description: "installing release [ " + r.Name + " ] in namespace [[ " + r.Namespace + " ]] using Tiller in [ " + getDesiredTillerNamespace(r) + " ]",
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
func getNoHooks(r *release) string {
	if r.NoHooks {
		return " --no-hooks "
	}
	return ""
}

// getTimeout returns the timeout flag for install/upgrade commands
func getTimeout(r *release) string {
	if r.Timeout != 0 {
		return " --timeout " + strconv.Itoa(r.Timeout)
	}
	return ""
}

// getValuesFiles return partial install/upgrade release command to substitute the -f flag in Helm.
func getValuesFiles(r *release) string {
	var fileList []string

	if r.ValuesFile != "" {
		fileList = append(fileList, r.ValuesFile)
	} else if len(r.ValuesFiles) > 0 {
		fileList = append(fileList, r.ValuesFiles...)
	}

	if r.SecretsFile != "" {
		if !helmPluginExists("secrets") {
			logError("ERROR: helm secrets plugin is not installed/configured correctly. Aborting!")
		}
		if ok := decryptSecret(r.SecretsFile); !ok {
			logError("Failed to decrypt secret file" + r.SecretsFile)
		}
		fileList = append(fileList, r.SecretsFile+".dec")
	} else if len(r.SecretsFiles) > 0 {
		if !helmPluginExists("secrets") {
			logError("ERROR: helm secrets plugin is not installed/configured correctly. Aborting!")
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

	if len(fileList) > 0 {
		return " -f " + strings.Join(fileList, " -f ")
	}
	return ""
}

// getSetValues returns --set params to be used with helm install/upgrade commands
func getSetValues(r *release) string {
	result := ""
	for k, v := range r.Set {
		result = result + " --set " + k + "=\"" + strings.Replace(v, ",", "\\,", -1) + "\""
	}
	return result
}

// getSetStringValues returns --set-string params to be used with helm install/upgrade commands
func getSetStringValues(r *release) string {
	result := ""
	for k, v := range r.SetString {
		result = result + " --set-string " + k + "=\"" + strings.Replace(v, ",", "\\,", -1) + "\""
	}
	return result
}

// getWait returns a partial helm command containing the helm wait flag (--wait) if the wait flag for the release was set to true
// Otherwise, retruns an empty string
func getWait(r *release) string {
	result := ""
	if r.Wait {
		result = " --wait"
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
func isProtected(r *release, rs releaseState) bool {

	// if the release does not exist in the cluster, it is not protected
	if ok, _ := helmReleaseExists(r, ""); !ok {
		return false
	}

	if getCurrentNamespaceProtection(rs) {
		return true
	}

	if r.Protected {
		return true
	}

	return false

}

// getDesiredTillerNamespaceFlag returns a tiller-namespace flag with which a release is desired to be maintained
func getDesiredTillerNamespaceFlag(r *release) string {
	return " --tiller-namespace " + getDesiredTillerNamespace(r)
}

// getDesiredTillerNamespace returns the Tiller namespace with which a release should be managed
func getDesiredTillerNamespace(r *release) string {
	if r.TillerNamespace != "" {
		return r.TillerNamespace
	} else if ns, ok := s.Namespaces[r.Namespace]; ok && (s.Settings.Tillerless || (ns.InstallTiller || ns.UseTiller)) {
		return r.Namespace
	}

	return "kube-system"
}

// getCurrentTillerNamespaceFlag returns the tiller-namespace with which a release is currently maintained
func getCurrentTillerNamespaceFlag(rs releaseState) string {
	if rs.TillerNamespace != "" {
		return " --tiller-namespace " + rs.TillerNamespace
	}
	return ""
}

// getTLSFlags returns TLS flags with which a release is maintained
// If the release where the namespace is to be deployed has Tiller deployed, the TLS flags will use certs/keys for that namespace (if any)
// otherwise, it will be the certs/keys for the kube-system namespace.
func getTLSFlags(r *release) string {
	tls := ""
	ns := s.Namespaces[r.TillerNamespace]
	if r.TillerNamespace != "" {
		if tillerTLSEnabled(ns) {

			tls = " --tls --tls-ca-cert " + r.TillerNamespace + "-ca.cert --tls-cert " + r.TillerNamespace + "-client.cert --tls-key " + r.TillerNamespace + "-client.key "
		}
	} else if s.Namespaces[r.Namespace].InstallTiller {
		ns := s.Namespaces[r.Namespace]
		if tillerTLSEnabled(ns) {

			tls = " --tls --tls-ca-cert " + r.Namespace + "-ca.cert --tls-cert " + r.Namespace + "-client.cert --tls-key " + r.Namespace + "-client.key "
		}
	} else {
		ns := s.Namespaces["kube-system"]
		if tillerTLSEnabled(ns) {

			tls = " --tls --tls-ca-cert kube-system-ca.cert --tls-cert kube-system-client.cert --tls-key kube-system-client.key "
		}
	}

	return tls
}

// getDryRunFlags returns dry-run flag
func getDryRunFlags() string {
	if dryRun {
		return " --dry-run --debug "
	}
	return ""
}

// getHelmFlags returns helm flags
func getHelmFlags(r *release) string {
	var flags string

	for _, flag := range r.HelmFlags {
		flags = flags + " " + flag
	}
	return getNoHooks(r) + getTimeout(r) + getDryRunFlags() + flags
}

func checkChartDepUpdate(r *release) {
	if updateDeps && isLocalChart(r.Chart) {
		if ok, err := updateChartDep(r.Chart); !ok {
			logError("ERROR: helm dependency update failed: " + err)
		}
	}
}
