package main

import (
	"log"
	"os"
	"strings"
)

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

	if exitCode, result := cmd.exec(); exitCode == 0 {
		// match, _ := regexp.MatchString(releaseName, result)
		return strings.Contains(result, releaseName+"\n")
	}

	return false
}

func getReleaseNamespace(releaseName string) string {

	if result := getReleaseStatus(releaseName); result != "" {
		if strings.Contains(result, "NAMESPACE:") {
			s := strings.Split(result, "\nNAMESPACE: ")[1]
			return strings.Split(s, "\n")[0]
		}
	} else {
		log.Fatal("ERROR: seems release [ " + releaseName + " ] does not exist.")
		os.Exit(1)
	}
	return ""
}

func getReleaseChart(releaseName string) string {
	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "helm list " + releaseName},
		Description: "inspecting the chart used for release:  " + releaseName,
	}
	exitCode, result := cmd.exec()

	if exitCode == 0 {
		line := strings.Split(result, "\n")[1]
		return strings.Fields(line)[4] // 4 is the position of chart details in helm ls output
	}
	log.Fatal("ERROR: seems release [ " + releaseName + " ] does not exist.")
	os.Exit(1)

	return ""
}

func getReleaseChartName(releaseName string) string {
	return strings.TrimSpace(strings.Split(getReleaseChart(releaseName), "-")[0])
}

func getReleaseChartVersion(releaseName string) string {
	return strings.TrimSpace(strings.Split(getReleaseChart(releaseName), "-")[1])
}

func getReleaseStatus(releaseName string) string {
	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "helm status " + releaseName},
		Description: "inspecting the status of release:  " + releaseName,
	}

	if exitCode, result := cmd.exec(); exitCode == 0 {
		return result
	}
	return ""
}
