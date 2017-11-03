package main

import (
	"log"
	"os"
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
func decide(r release, s *state) {

	// check for deletion
	if !r.Enabled {

		inspectDeleteScenario(s.Namespaces[r.Env], r)

	} else { // check for install/upgrade/rollback
		if helmReleaseExists(s.Namespaces[r.Env], r.Name, "deployed") {

			inspectUpgradeScenario(s.Namespaces[r.Env], r)

		} else if helmReleaseExists(s.Namespaces[r.Env], r.Name, "deleted") {

			inspectRollbackScenario(s.Namespaces[r.Env], r)

		} else {
			if !helmReleaseExists(s.Namespaces[r.Env], r.Name, "all") {

				installRelease(s.Namespaces[r.Env], r)

			} else {

				log.Fatal("ERROR: it seems that release [ " + r.Name + " ] exists in the current k8s context. Please double check!")
				os.Exit(1)

			}
		}

	}

}

// testRelease creates a Helm command to test a particular release.
func testRelease(releaseName string) {

	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "helm test " + releaseName},
		Description: "running tests for release [ " + releaseName + " ]",
	}
	outcome.addCommand(cmd)
	logDecision("DECISION: release [ " + releaseName + " ] is required to be tested when installed/upgraded/rolledback. Got it!")

}

// installRelease creates a Helm command to install a particular release in a particular namespace.
func installRelease(namespace string, r release) {

	releaseName := r.Name
	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "helm install " + r.Chart + " -n " + releaseName + " --namespace " + namespace + " -f " + r.ValuesFile},
		Description: "installing release [ " + releaseName + " ] in namespace [[ " + namespace + " ]]",
	}
	outcome.addCommand(cmd)
	logDecision("DECISION: release [ " + releaseName + " ] is not present in the current k8s context. Will install it in namespace [[ " +
		namespace + " ]]")

	if r.Test {
		testRelease(releaseName)
	}
}

// inspectRollbackScenario evaluates if a rollback action needs to be taken for a given release.
// if the release is already deleted but from a different namespace than the one specified in input,
// it purge deletes it and create it in the spcified namespace.
func inspectRollbackScenario(namespace string, r release) {

	releaseName := r.Name
	if getReleaseNamespace(r.Name) == namespace {

		cmd := command{
			Cmd:         "bash",
			Args:        []string{"-c", "helm rollback " + releaseName},
			Description: "rolling back release [ " + releaseName + " ]",
		}
		outcome.addCommand(cmd)
		logDecision("DECISION: release [ " + releaseName + " ] is currently deleted and is desired to be rolledback to " +
			"namespace [[ " + namespace + " ]] . No problem!")

		if r.Test {
			testRelease(releaseName)
		}

	} else {

		reInstallRelease(namespace, r)
		logDecision("DECISION: release [ " + releaseName + " ] is deleted BUT from namespace [[ " + getReleaseNamespace(releaseName) +
			" ]]. Will purge delete it from there and install it in namespace [[ " + namespace + " ]]")

	}
}

// inspectDeleteScenario evaluates if a delete action needs to be taken for a given release.
// If the purge flage is set to true for the release in question, then it will be permenantly removed.
// If the release is not deployed, it will be skipped.
func inspectDeleteScenario(namespace string, r release) {

	releaseName := r.Name
	//if it exists in helm list , add command to delete it, else log that it is skipped
	if helmReleaseExists(namespace, releaseName, "deployed") {
		purge := ""
		purgeDesc := ""
		if r.Purge {
			purge = "--purge"
			purgeDesc = "and purged!"
		}
		// delete it
		cmd := command{
			Cmd:         "bash",
			Args:        []string{"-c", "helm delete " + purge + " " + releaseName},
			Description: "deleting release [ " + releaseName + " ]",
		}
		outcome.addCommand(cmd)
		logDecision("DECISION: release [ " + releaseName + " ] is desired to be deleted " + purgeDesc + ". Planing this for you!")

	} else {
		logDecision("DECISION: release [ " + releaseName + " ] is set to be disabled but is not yet deployed. Skipping.")
	}
}

// inspectUpgradeScenario evaluates if a release should be upgraded.
// - If the relase is already in the same namespace specified in the input,
// it will be upgraded using the values file specified in the release info.
// - If the relase is already in the same namespace specified in the input but is using a different chart,
// it will be purge deleted and installed in the same namespace using the new chart.
// - If the release is NOT in the same namespace specified in the input,
// it will be purge deleted and installed in the new namespace.
func inspectUpgradeScenario(namespace string, r release) {

	releaseName := r.Name
	if getReleaseNamespace(releaseName) == namespace {
		if extractChartName(r.Chart) == getReleaseChartName(releaseName) && r.Version != getReleaseChartVersion(releaseName) {
			// upgrade
			cmd := command{
				Cmd:         "bash",
				Args:        []string{"-c", "helm upgrade " + releaseName + " " + r.Chart + " -f " + r.ValuesFile},
				Description: "upgrading release [ " + releaseName + " ]",
			}
			outcome.addCommand(cmd)
			logDecision("DECISION: release [ " + releaseName + " ] is desired to be upgraded. Planing this for you!")
			if r.Test {
				testRelease(releaseName)
			}

		} else if extractChartName(r.Chart) != getReleaseChartName(releaseName) {
			reInstallRelease(namespace, r)
			logDecision("DECISION: release [ " + releaseName + " ] is desired to use a new Chart [ " + r.Chart +
				" ]. I am planning a purge delete of the current release and will install it with the new chart in namespace [[ " +
				namespace + " ]]")

		} else {
			logDecision("DECISION: release [ " + releaseName + " ] is desired to be enabled and is currently enabled." +
				"Nothing for me to do. Horraayyy!")
		}
	} else {
		reInstallRelease(namespace, r)
		logDecision("DECISION: release [ " + releaseName + " ] is desired to be enabled in a new namespace [[ " + namespace +
			" ]]. I am planning a purge delete of the current release from namespace [[ " + getReleaseNamespace(releaseName) + " ]] " +
			"and will install it for you in namespace [[ " + namespace + " ]]")
	}
}

// reInstallRelease purge deletes a release and reinstall it.
// This is used when moving a release to another namespace or when changing the chart used for it.
func reInstallRelease(namespace string, r release) {

	releaseName := r.Name
	delCmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "helm delete --purge " + releaseName},
		Description: "deleting release [ " + releaseName + " ]",
	}
	outcome.addCommand(delCmd)

	installCmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "helm install " + r.Chart + " -n " + releaseName + " --namespace " + namespace + " -f " + r.ValuesFile},
		Description: "installing release [ " + releaseName + " ] in namespace [[ " + namespace + " ]]",
	}
	outcome.addCommand(installCmd)

	if r.Test {
		testRelease(releaseName)
	}
}

// logDecision adds the decisions made to the plan.
// Depending on the debug flag being set or not, it will either log the the decision to output or not.
func logDecision(decision string) {

	if debug {
		log.Println(decision)
	}
	outcome.addDecision(decision)

}

// extractChartName extracts the Helm chart name from full chart name.
// example: it extracts "chartY" from "repoX/chartY"
func extractChartName(releaseChart string) string {

	return strings.TrimSpace(strings.Split(releaseChart, "/")[1])

}
