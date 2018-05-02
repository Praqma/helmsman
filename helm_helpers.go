package main

import (
	"log"
	"strconv"
	"strings"
	"time"
)

var currentState map[string]releaseState

// releaseState represents the current state of a release
type releaseState struct {
	Revision        int
	Updated         time.Time
	Status          string
	Chart           string
	Namespace       string
	TillerNamespace string
}

// getAllReleases fetches a list of all releases in a k8s cluster
func getAllReleases() string {
	result := getTillerReleases("kube-system")
	for ns, v := range s.Namespaces {
		if v.InstallTiller && ns != "kube-system" {
			result = result + getTillerReleases(ns)
		}
	}

	return result
}

// getTillerReleases gets releases deployed with a given Tiller (in agiven namespace)
func getTillerReleases(tillerNS string) string {
	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "helm list --all --tiller-namespace " + tillerNS + getNSTLSFlags(tillerNS)},
		Description: "listing all existing releases in namespace [ " + tillerNS + " ]...",
	}

	exitCode, result := cmd.exec(debug, verbose)
	if exitCode != 0 {
		log.Fatal("ERROR: failed to list all releases in namespace [ " + tillerNS + " ]: " + result)
	}

	// appending tiller-namespace to each release found
	lines := strings.Split(result, "\n")
	for i, l := range lines {
		if l != "" && !strings.HasPrefix(l, "NAME") && !strings.HasSuffix(l, "NAMESPACE") {
			lines[i] = strings.TrimSuffix(l, "\n") + " " + tillerNS
		}
	}
	return strings.Join(lines, "\n")
}

// buildState builds the currentState map contianing information about all releases existing in a k8s cluster
func buildState() {
	log.Println("INFO: mapping the current helm state ...")
	currentState = make(map[string]releaseState)
	lines := strings.Split(getAllReleases(), "\n")

	for i := 0; i < len(lines); i++ {
		if lines[i] == "" || (strings.HasPrefix(lines[i], "NAME") && strings.HasSuffix(lines[i], "NAMESPACE")) {
			continue
		}
		r, _ := strconv.Atoi(strings.Fields(lines[i])[1])
		t := strings.Fields(lines[i])[2] + " " + strings.Fields(lines[i])[3] + " " + strings.Fields(lines[i])[4] + " " +
			strings.Fields(lines[i])[5] + " " + strings.Fields(lines[i])[6]
		time, err := time.Parse("Mon Jan _2 15:04:05 2006", t)
		if err != nil {
			log.Fatal("ERROR: while converting release time: " + err.Error())
		}

		currentState[strings.Fields(lines[i])[0]] = releaseState{
			Revision:        r,
			Updated:         time,
			Status:          strings.Fields(lines[i])[7],
			Chart:           strings.Fields(lines[i])[8],
			Namespace:       strings.Fields(lines[i])[9],
			TillerNamespace: strings.Fields(lines[i])[10],
		}
	}
}

// Deprecated: listReleases lists releases in a given namespace and with a given status
func listReleases(namespace string, scope string) string {
	var options string
	if scope == "all" {
		options = "--all -q"
	} else if scope == "deleted" {
		options = "--deleted -q"
	} else if scope == "deployed" && namespace != "" {
		options = "--deployed -q --namespace " + namespace
	} else if scope == "deployed" && namespace == "" {
		options = "--deployed -q "
	} else if scope == "failed" {
		options = "--failed -q"
	} else {
		options = "--all -q"
		log.Println("INFO: scope " + scope + " is not valid, using [ all ] instead!")
	}

	ns := namespace
	tls := ""
	if namespace == "" {
		ns = "all"
		tls = getNSTLSFlags("kube-system")
	} else {
		tls = getNSTLSFlags(namespace)
	}

	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "helm list " + options + tls},
		Description: "listing the existing releases in namespace [ " + ns + " ] with status [ " + scope + " ]",
	}

	exitCode, result := cmd.exec(debug, verbose)
	if exitCode != 0 {
		log.Fatal("ERROR: failed to list " + scope + " releases in " + ns + " namespace(s): " + result)
	}

	return result
}

// helmRealseExists checks if a Helm release is/was deployed in a k8s cluster.
// The search criteria is:
//
// -releaseName: the name of the release to look for. Helm releases have unique names within a k8s cluster.
// -scope: defines where to search for the release. Options are: [deleted, deployed, all, failed]
// -namespace: search in that namespace (only applicable if searching for currently deployed releases)
func helmReleaseExists(namespace string, releaseName string, status string) bool {
	v, ok := currentState[releaseName]
	if !ok {
		return false
	}

	if namespace != "" && status != "" {
		if v.Namespace == namespace && v.Status == strings.ToUpper(status) {
			return true
		}
		return false
	} else if namespace != "" {
		if v.Namespace == namespace {
			return true
		}
		return false
	} else if status != "" {
		if v.Status == strings.ToUpper(status) {
			return true
		}
		return false
	}
	return true
}

// getReleaseNamespace returns the namespace in which a release is deployed.
// throws an error and exits the program if the release does not exist.
func getReleaseNamespace(releaseName string) string {

	v, ok := currentState[releaseName]
	if !ok {
		log.Fatal("ERROR: seems release [ " + releaseName + " ] does not exist.")
	}
	return v.Namespace
}

// getReleaseChart returns the Helm chart which is used by a deployed release.
// throws an error and exits the program if the release does not exist.
func getReleaseChart(releaseName string) string {

	v, ok := currentState[releaseName]
	if !ok {
		log.Fatal("ERROR: seems release [ " + releaseName + " ] does not exist.")
	}
	return v.Chart
}

// getReleaseRevision returns the revision number for a release (if it exists)
func getReleaseRevision(releaseName string, state string) string {

	v, ok := currentState[releaseName]
	if !ok {
		log.Fatal("ERROR: seems release [ " + releaseName + " ] does not exist.")
	}
	return strconv.Itoa(v.Revision)
}

// getReleaseChartName extracts and returns the Helm chart name from the chart info retrieved by getReleaseChart().
// example: getReleaseChart() returns "jenkins-0.9.0" and this functions will extract "jenkins" from it.
func getReleaseChartName(releaseName string) string {
	chart := getReleaseChart(releaseName)
	runes := []rune(chart)
	return string(runes[0:strings.LastIndexByte(chart, '-')])
}

// getReleaseChartVersion extracts and returns the Helm chart version from the chart info retrieved by getReleaseChart().
// example: getReleaseChart() returns "jenkins-0.9.0" and this functions will extract "0.9.0" from it.
func getReleaseChartVersion(releaseName string) string {
	chart := getReleaseChart(releaseName)
	runes := []rune(chart)
	return string(runes[strings.LastIndexByte(chart, '-')+1 : len(chart)])
}

// getReleaseStatus returns the output of Helm status command for a release.
// if the release does not exist, it returns an empty string without breaking the program execution.
func getReleaseStatus(releaseName string) string {
	cmd := command{
		Cmd:         "bash",
		Args:        []string{"-c", "helm status " + releaseName + getNSTLSFlags("kube-system")},
		Description: "inspecting the status of release:  " + releaseName,
	}

	exitCode, result := cmd.exec(debug, verbose)
	if exitCode == 0 {
		return result
	}

	log.Fatal("ERROR: while checking release [ " + releaseName + " ] status: " + result)

	return ""
}

// getNSTLSFlags returns TLS flags for a given namespace if it's deployed with TLS
func getNSTLSFlags(ns string) string {
	tls := ""
	if tillerTLSEnabled(ns) {

		tls = " --tls --tls-ca-cert " + ns + "-ca.cert --tls-cert " + ns + "-client.cert --tls-key " + ns + "-client.key "
	}
	return tls
}
