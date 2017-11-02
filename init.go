package main

import (
	"log"
	"os"
)

func init() {

	// check helm exists
	if !helmExists() {
		log.Fatal("ERROR: helm is not installed/configured correctly. Aborting!")
		os.Exit(1)
	}

	// fromTOML(*file, &s)
	// validate the desired state content
	// s.validate() // syntax validation

	// // set the kubecontext
	// if !setKubeContext(s.Settings["kubeContext"]) {
	// 	os.Exit(1)
	// }

	// // add repos -- fails if they are not valid
	// if !addHelmRepos(s.HelmRepos) {
	// 	os.Exit(1)
	// }

	// // validate charts-versions exist in supllied repos
	// if !validateReleaseCharts(s.Apps) {
	// 	os.Exit(1)
	// }

	// // add/validate namespaces
	// if !addNamespaces(s.Namespaces) {
	// 	os.Exit(1)
	// }

	//s.print()
}

func helmExists() bool {
	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "helm "},
		Description: "testing the helm command",
	}

	exitCode, _ := cmd.exec()
	// if strings.Contains(result, "helm: command not found") ||
	// 	strings.Contains(result, "Error") {
	// 	return false
	// }
	if exitCode != 0 {
		return false
	}

	return true
}

func addNamespaces(namespaces map[string]string) {
	for _, namespace := range namespaces {
		cmd := command{
			Cmd:         "bash",
			Args:        []string{"-c", "kubectl create namespace " + namespace},
			Description: "creating namespace  " + namespace,
		}

		exitCode, _ := cmd.exec()

		// if !strings.Contains(result, "created") && !strings.Contains(result, "AlreadyExists") {
		if exitCode != 0 {
			log.Println("WARN: I could not create namespace [" +
				namespace + " ]. It already exists. I am skipping this.")
		}
	}
}

func validateReleaseCharts(apps map[string]release) bool {

	for app, r := range apps {
		cmd := command{
			Cmd:         "bash",
			Args:        []string{"-c", "helm search " + r.Chart + " --version " + r.Version},
			Description: "searching chart " + r.Chart + "-" + r.Version,
		}

		exitCode, _ := cmd.exec()

		// if strings.Contains(result, "No results found") {
		if exitCode != 0 {
			log.Fatal("ERROR: chart "+r.Chart+"-"+r.Version+" is specified for ",
				"app ["+app+"] but is not found in the provided repos.")
			return false
		}
	}
	return true
}

func addHelmRepos(repos map[string]string) bool {

	for repoName, url := range repos {
		cmd := command{
			Cmd:         "bash",
			Args:        []string{"-c", "helm repo add " + repoName + " " + url},
			Description: "adding repo " + repoName,
		}

		exitCode, _ := cmd.exec()

		// if strings.Contains(result, "Error") {
		if exitCode != 0 {
			log.Fatal("ERROR: there has been a problem while adding repo [" +
				repoName + "].")
			return false
		}

	}

	return true
}

func setKubeContext(context string) bool {
	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "kubectl config use-context " + context},
		Description: "setting kubectl context",
	}

	exitCode, result := cmd.exec()

	// if strings.Contains(result, "kubectl: command not found") ||
	// 	strings.Contains(result, "error: no context exists with the name") {
	if exitCode != 0 {
		log.Fatal("ERROR: there has been a problem with setting up KubeContext: " + result)
		return false
	}

	return true
}
