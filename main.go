package main

import (
	"flag"
	"os"
)

var s state

func main() {

	file := flag.String("f", "", "desired state file name")
	apply := flag.Bool("apply", false, "apply the plan directly")

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
