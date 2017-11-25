package main

import (
	"log"
	"strings"
)

// helmRealseExists checks if a Helm release is/was deployed in a k8s cluster.
// The search criteria is:
//
// -releaseName: the name of the release to look for. Helm releases have unique names within a k8s cluster.
// -scope: defines where to search for the release. Options are: [deleted, deployed, all, failed]
// -namespace: search in that namespace (only applicable if searching for currently deployed releases)
func helmReleaseExists(namespace string, releaseName string, scope string) bool {

	var options string
	if scope == "all" {
		options = "--all -q"
	} else if scope == "deleted" {
		options = "--deleted -q"
	} else if scope == "deployed" {
		options = "--deployed -q --namespace " + namespace
	} else if scope == "failed" {
		options = "--failed -q"
	} else {
		options = "--all -q"
		log.Println("INFO: scope " + scope + " is not valid, using [ all ] instead!")
	}

	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "helm list " + options},
		Description: "listing the existing releases in namespace [ " + namespace + " ] with status [ " + scope + " ]",
	}

	if exitCode, result := cmd.exec(debug); exitCode == 0 {
		return strings.Contains(result, releaseName+"\n")
	}

	log.Fatal("ERROR: something went wrong while checking helm release.")

	return false
}

// getReleaseNamespace returns the namespace in which a release is deployed.
// throws an error and exits the program if the release does not exist.
func getReleaseNamespace(releaseName string) string {

	if result := getReleaseStatus(releaseName); result != "" {
		if strings.Contains(result, "NAMESPACE:") {
			s := strings.Split(result, "\nNAMESPACE: ")[1]
			return strings.Split(s, "\n")[0]
		}
	} else {
		log.Fatal("ERROR: seems release [ " + releaseName + " ] does not exist.")
	}
	return ""
}

// getReleaseChart returns the Helm chart which is used by a deployed release.
// throws an error and exits the program if the release does not exist.
func getReleaseChart(releaseName string) string {
	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "helm list " + releaseName},
		Description: "inspecting the chart used for release:  " + releaseName,
	}
	exitCode, result := cmd.exec(debug)

	if exitCode == 0 {
		line := strings.Split(result, "\n")[1]
		return strings.Fields(line)[8] // 8 is the position of chart details in helm ls output
	}
	log.Fatal("ERROR: seems release [ " + releaseName + " ] does not exist.")

	return ""
}

// getReleaseRevision returns the revision number for a release (if it exists)
func getReleaseRevision(releaseName string, state string) string {
	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "helm list " + releaseName + " --" + state},
		Description: "inspecting the release revision for :  " + releaseName,
	}
	exitCode, result := cmd.exec(debug)

	if exitCode == 0 {
		line := strings.Split(result, "\n")[1]
		return strings.Fields(line)[1] // 1 is the position of revision number in helm ls output
	}
	log.Fatal("ERROR: seems release [ " + releaseName + " ] does not exist.")

	return ""
}

// getReleaseChartName extracts and returns the Helm chart name from the chart info retrieved by getReleaseChart().
// example: getReleaseChart() returns "stable/jenkins-0.9.0" and this functions will extract "stable/jenkins" from it.
func getReleaseChartName(releaseName string) string {
	return strings.TrimSpace(strings.Split(getReleaseChart(releaseName), "-")[0])
}

// getReleaseChartVersion extracts and returns the Helm chart version from the chart info retrieved by getReleaseChart().
// example: getReleaseChart() returns "stable/jenkins-0.9.0" and this functions will extract "0.9.0" from it.
func getReleaseChartVersion(releaseName string) string {
	return strings.TrimSpace(strings.Split(getReleaseChart(releaseName), "-")[1])
}

// getReleaseStatus returns the output of Helm status command for a release.
// if the release does not exist, it returns an empty string without breaking the program execution.
func getReleaseStatus(releaseName string) string {
	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "helm status " + releaseName},
		Description: "inspecting the status of release:  " + releaseName,
	}

	if exitCode, result := cmd.exec(debug); exitCode == 0 {
		return result
	}

	log.Fatal("ERROR: something went wrong while checking release status.")

	return ""
}
