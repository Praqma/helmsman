package main

import (
	"flag"
	"log"
	"os"
)

var s state
var debug bool

func main() {

	file := flag.String("f", "", "desired state file name")
	apply := flag.Bool("apply", false, "apply the plan directly")
	flag.BoolVar(&debug, "debug", false, "show the execution logs")

	flag.Parse()

	// 1) init -- validate helm exists, kubeconfig is configured,
	//add helm repos if not added, create k8s namespaces if not there ...
	// read TOML file ...
	// validate the desired state info are correct ...

	// 3) make a plan -- and prepare a list of helm operations to perfrom

	fromTOML(*file, &s)
	// validate the desired state content
	s.validate() // syntax validation

	// set the kubecontext
	if !setKubeContext(s.Settings["kubeContext"]) {
		os.Exit(1)
	}

	// add repos -- fails if they are not valid
	if !addHelmRepos(s.HelmRepos) {
		os.Exit(1)
	}

	// validate charts-versions exist in supllied repos
	if !validateReleaseCharts(s.Apps) {
		os.Exit(1)
	}

	// add/validate namespaces
	addNamespaces(s.Namespaces)

	p := makePlan(&s)

	if !*apply {
		p.printPlan()
	} else {
		p.execPlan()
	}

	// 4) if planning is succcessful, execute the plan

	// 5) if the plan execution is successful, update the desired state and push it back to git repo (design to be validated)
	// toTOML("newState.toml", &s)

}

func addNamespaces(namespaces map[string]string) {
	for _, namespace := range namespaces {
		cmd := command{
			Cmd:         "bash",
			Args:        []string{"-c", "kubectl create namespace " + namespace},
			Description: "creating namespace  " + namespace,
		}

		exitCode, _ := cmd.exec(debug)

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
			Description: "validating chart " + r.Chart + "-" + r.Version + " is available in the used repos.",
		}

		exitCode, _ := cmd.exec(debug)

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

		exitCode, _ := cmd.exec(debug)

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
		Description: "setting kubectl context to [ " + context + " ]",
	}

	exitCode, result := cmd.exec(debug)

	if exitCode != 0 {
		log.Fatal("ERROR: there has been a problem with setting up KubeContext: " + result)
		return false
	}

	return true
}
