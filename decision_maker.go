package main

import (
	"log"
	"strings"
)

var outcome plan

// makePlan creates a plan of the actions needed to make the desired state come true.
func makePlan(s *state) *plan {
	outcome = createPlan()
	for _, r := range s.Apps {
		decide(r, s)
	}

	return &outcome
}

// decide makes a decision about what commands (actions) need to be executed
// to make a release section of the desired state come true.
func decide(r *release, s *state) {

	// check for deletion
	if !r.Enabled {
		if !isProtected(r) {
			inspectDeleteScenario(getDesiredNamespace(r), r)
		} else {
			logDecision("DECISION: release "+r.Name+" is PROTECTED. Operations are not allowed on this release until "+
				"you remove its protection.", r.Priority)
		}

	} else { // check for install/upgrade/rollback
		if helmReleaseExists(getDesiredNamespace(r), r.Name, "deployed") {
			if !isProtected(r) {
				inspectUpgradeScenario(getDesiredNamespace(r), r) // upgrade
			} else {
				logDecision("DECISION: release "+r.Name+" is PROTECTED. Operations are not allowed on this release until "+
					"you remove its protection.", r.Priority)
			}

		} else if helmReleaseExists("", r.Name, "deleted") {
			if !isProtected(r) {

				inspectRollbackScenario(getDesiredNamespace(r), r) // rollback

			} else {
				logDecision("DECISION: release "+r.Name+" is PROTECTED. Operations are not allowed on this release until "+
					"you remove its protection.", r.Priority)
			}

		} else if helmReleaseExists("", r.Name, "failed") {

			if !isProtected(r) {

				reInstallRelease(getDesiredNamespace(r), r) // re-install failed release

			} else {
				logDecision("DECISION: release "+r.Name+" is PROTECTED. Operations are not allowed on this release until "+
					"you remove its protection.", r.Priority)
			}

		} else if helmReleaseExists("", r.Name, "all") { // not deployed in the desired namespace but deployed elsewhere

			if !isProtected(r) {

				reInstallRelease(getDesiredNamespace(r), r) // move the release to a new (the desired) namespace
				logDecision("WARNING: moving release [ "+r.Name+" ] from [[ "+getReleaseNamespace(r.Name)+" ]] to [[ "+getDesiredNamespace(r)+
					" ]] might not correctly connect to existing volumes. Check https://github.com/Praqma/helmsman/blob/master/docs/how_to/move_charts_across_namespaces.md"+
					" for details if this release uses PV and PVC.", r.Priority)

			} else {
				logDecision("DECISION: release "+r.Name+" is PROTECTED. Operations are not allowed on this release until "+
					"you remove its protection.", r.Priority)
			}

		} else {

			installRelease(getDesiredNamespace(r), r) // install a new release

		}

	}

}

// testRelease creates a Helm command to test a particular release.
func testRelease(r *release) {

	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "helm test " + r.Name},
		Description: "running tests for release [ " + r.Name + " ]",
	}
	outcome.addCommand(cmd, r.Priority)
	logDecision("DECISION: release [ "+r.Name+" ] is required to be tested when installed/upgraded/rolledback. Got it!", r.Priority)

}

// installRelease creates a Helm command to install a particular release in a particular namespace.
func installRelease(namespace string, r *release) {

	releaseName := r.Name
	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "helm install " + r.Chart + " -n " + releaseName + " --namespace " + namespace + getValuesFile(r) + " --version " + r.Version + getSetValues(r) + getWait(r)},
		Description: "installing release [ " + releaseName + " ] in namespace [[ " + namespace + " ]]",
	}
	outcome.addCommand(cmd, r.Priority)
	logDecision("DECISION: release [ "+releaseName+" ] is not present in the current k8s context. Will install it in namespace [[ "+
		namespace+" ]]", r.Priority)

	if r.Test {
		testRelease(r)
	}
}

// inspectRollbackScenario evaluates if a rollback action needs to be taken for a given release.
// if the release is already deleted but from a different namespace than the one specified in input,
// it purge deletes it and create it in the spcified namespace.
func inspectRollbackScenario(namespace string, r *release) {

	releaseName := r.Name
	if getReleaseNamespace(r.Name) == namespace {

		cmd := command{
			Cmd:         "bash",
			Args:        []string{"-c", "helm rollback " + releaseName + " " + getReleaseRevision(releaseName, "deleted") + getWait(r)},
			Description: "rolling back release [ " + releaseName + " ]",
		}
		outcome.addCommand(cmd, r.Priority)
		logDecision("DECISION: release [ "+releaseName+" ] is currently deleted and is desired to be rolledback to "+
			"namespace [[ "+namespace+" ]] . No problem!", r.Priority)

		// if r.Test {
		// 	testRelease(r)
		// }

	} else {

		reInstallRelease(namespace, r)
		logDecision("DECISION: release [ "+releaseName+" ] is deleted BUT from namespace [[ "+getReleaseNamespace(releaseName)+
			" ]]. Will purge delete it from there and install it in namespace [[ "+namespace+" ]]", r.Priority)
		logDecision("WARNING: rolling back release [ "+releaseName+" ] from [[ "+getReleaseNamespace(releaseName)+" ]] to [[ "+namespace+
			" ]] might not correctly connect to existing volumes. Check https://github.com/Praqma/helmsman/blob/master/docs/how_to/move_charts_across_namespaces.md"+
			" for details if this release uses PV and PVC.", r.Priority)

	}
}

// inspectDeleteScenario evaluates if a delete action needs to be taken for a given release.
// If the purge flage is set to true for the release in question, then it will be permenantly removed.
// If the release is not deployed, it will be skipped.
func inspectDeleteScenario(namespace string, r *release) {

	releaseName := r.Name
	//if it exists in helm list , add command to delete it, else log that it is skipped
	if helmReleaseExists(namespace, releaseName, "deployed") {
		// delete it
		deleteRelease(r)

	} else {
		logDecision("DECISION: release [ "+releaseName+" ] is set to be disabled but is not yet deployed. Skipping.", r.Priority)
	}
}

// deleteRelease deletes a release from a k8s cluster
func deleteRelease(r *release) {
	p := ""
	purgeDesc := ""
	if r.Purge {
		p = "--purge"
		purgeDesc = "and purged!"
	}

	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "helm delete " + p + " " + r.Name},
		Description: "deleting release [ " + r.Name + " ]",
	}
	outcome.addCommand(cmd, r.Priority)
	logDecision("DECISION: release [ "+r.Name+" ] is desired to be deleted "+purgeDesc+". Planing this for you!", r.Priority)
}

// inspectUpgradeScenario evaluates if a release should be upgraded.
// - If the relase is already in the same namespace specified in the input,
// it will be upgraded using the values file specified in the release info.
// - If the relase is already in the same namespace specified in the input but is using a different chart,
// it will be purge deleted and installed in the same namespace using the new chart.
// - If the release is NOT in the same namespace specified in the input,
// it will be purge deleted and installed in the new namespace.
func inspectUpgradeScenario(namespace string, r *release) {

	releaseName := r.Name
	if getReleaseNamespace(releaseName) == namespace {
		if extractChartName(r.Chart) == getReleaseChartName(releaseName) && r.Version != getReleaseChartVersion(releaseName) {
			// upgrade
			upgradeRelease(r)
			logDecision("DECISION: release [ "+releaseName+" ] is desired to be upgraded. Planing this for you!", r.Priority)

		} else if extractChartName(r.Chart) != getReleaseChartName(releaseName) {
			reInstallRelease(namespace, r)
			logDecision("DECISION: release [ "+releaseName+" ] is desired to use a new Chart [ "+r.Chart+
				" ]. I am planning a purge delete of the current release and will install it with the new chart in namespace [[ "+
				namespace+" ]]", r.Priority)

		} else {
			upgradeRelease(r)
			logDecision("DECISION: release [ "+releaseName+" ] is desired to be enabled and is currently enabled."+
				"I will upgrade it in case you changed your values.yaml!", r.Priority)
		}
	} else {
		reInstallRelease(namespace, r)
		logDecision("DECISION: release [ "+releaseName+" ] is desired to be enabled in a new namespace [[ "+namespace+
			" ]]. I am planning a purge delete of the current release from namespace [[ "+getReleaseNamespace(releaseName)+" ]] "+
			"and will install it for you in namespace [[ "+namespace+" ]]", r.Priority)
		logDecision("WARNING: moving release [ "+releaseName+" ] from [[ "+getReleaseNamespace(releaseName)+" ]] to [[ "+namespace+
			" ]] might not correctly connect to existing volumes. Check https://github.com/Praqma/helmsman/blob/master/docs/how_to/move_charts_across_namespaces.md"+
			" for details if this release uses PV and PVC.", r.Priority)
	}
}

// upgradeRelease upgrades an existing release with the specified values.yaml
func upgradeRelease(r *release) {
	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "helm upgrade " + r.Name + " " + r.Chart + getValuesFile(r) + " --version " + r.Version + " --force " + getSetValues(r) + getWait(r)},
		Description: "upgrading release [ " + r.Name + " ]",
	}

	outcome.addCommand(cmd, r.Priority)

	// if r.Test {
	// 	testRelease(r)
	// }
}

// reInstallRelease purge deletes a release and reinstalls it.
// This is used when moving a release to another namespace or when changing the chart used for it.
func reInstallRelease(namespace string, r *release) {

	releaseName := r.Name
	delCmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "helm delete --purge " + releaseName},
		Description: "deleting release [ " + releaseName + " ]",
	}
	outcome.addCommand(delCmd, r.Priority)

	installCmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "helm install " + r.Chart + " --version " + r.Version + " -n " + releaseName + " --namespace " + namespace + getValuesFile(r) + getSetValues(r) + getWait(r)},
		Description: "installing release [ " + releaseName + " ] in namespace [[ " + namespace + " ]]",
	}
	outcome.addCommand(installCmd, r.Priority)
	logDecision("DECISION: release [ "+releaseName+" ] will be deleted from namespace [[ "+getReleaseNamespace(releaseName)+" ]] and reinstalled in [[ "+namespace+"]].", r.Priority)

	// if r.Test {
	// 	testRelease(releaseName)
	// }
}

// logDecision adds the decisions made to the plan.
// Depending on the debug flag being set or not, it will either log the the decision to output or not.
func logDecision(decision string, priority int) {

	if debug {
		log.Println(decision)
	}
	outcome.addDecision(decision, priority)

}

// extractChartName extracts the Helm chart name from full chart name in the desired state.
// example: it extracts "chartY" from "repoX/chartY"
func extractChartName(releaseChart string) string {

	return strings.TrimSpace(strings.Split(releaseChart, "/")[1])

}

// getValuesFile return partial install/upgrade release command to substitute the -f flag in Helm.
func getValuesFile(r *release) string {
	if r.ValuesFile != "" {
		return " -f " + r.ValuesFile
	}
	return ""
}

// getSetValues returns --set params to be used with helm install/upgrade commands
func getSetValues(r *release) string {
	result := ""
	for k, v := range r.Set {
		_, value := envVarExists(v)
		result = result + " --set " + k + "=\"" + strings.Replace(value, ",", "\\,", -1) + "\""
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
func getCurrentNamespaceProtection(r *release) bool {

	return s.Namespaces[getReleaseNamespace(r.Name)].Protected
}

// isProtected checks if a release is protected or not.
// A protected is release is either: a) deployed in a protected namespace b) flagged as protected in the desired state file
// Any release in a protected namespace is protected by default regardless of its flag
// returns true if a release is protected, false otherwise
func isProtected(r *release) bool {

	// if the release does not exist in the cluster, it is not protected
	if !helmReleaseExists("", r.Name, "all") {
		return false
	}

	if getCurrentNamespaceProtection(r) {
		return true
	}

	if r.Protected {
		return true
	}

	return false

}
