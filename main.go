package main

import "os"

var s state
var debug bool
var file string
var apply bool
var help bool

func main() {

	// set the kubecontext to be used Or create it if it does not exist
	if !setKubeContext(s.Settings["kubeContext"]) {
		if !createContext() {
			os.Exit(1)
		}
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

	if !apply {
		p.printPlan()
	} else {
		p.execPlan()
	}

}
