package main

import (
	"log"
	"os"
	"strings"
)

var outcome plan

func makePlan(s *state) *plan {

	outcome = createPlan()
	for _, r := range s.Apps {
		decide(r, s)
	}
	// outcome.printPlan()
	// outcome.printPlanCmds()
	return &outcome
}

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

}

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

	} else {

		reInstallRelease(namespace, r)
		logDecision("DECISION: release [ " + releaseName + " ] is deleted BUT from namespace [[ " + getReleaseNamespace(releaseName) +
			" ]]. Will purge delete it from there and install it in namespace [[ " + namespace + " ]]")

	}
}

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

		} else if extractChartName(r.Chart) != getReleaseChartName(releaseName) {
			// TODO: check new chart is valid/exists
			reInstallRelease(namespace, r)
			logDecision("DECISION: release [ " + releaseName + " ] is desired to use a new Chart [ " + r.Chart +
				" ]. I am planning a purge delete of the current release and will install it with the new chart in namespace [[ " +
				namespace + " ]]")

		} else {
			logDecision("DECISION: release [ " + releaseName + " ] is desired to be enabled and is currently enabled." +
				"Nothing for me to do. Horraayyy!")
		}
	} else {
		// TODO: validate new chart exists
		reInstallRelease(namespace, r)
		logDecision("DECISION: release [ " + releaseName + " ] is desired to be enabled in a new namespace [[ " + namespace +
			" ]]. I am planning a purge delete of the current release from namespace [[ " + getReleaseNamespace(releaseName) + " ]] " +
			"and will install it for you in namespace [[ " + namespace + " ]]")
	}
}

// purge delete and install
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
}

func logDecision(decision string) {

	log.Println(decision)
	outcome.addDecision(decision)

}

func extractChartName(releaseChart string) string {

	return strings.TrimSpace(strings.Split(releaseChart, "/")[1])

}
