package app

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// helmRelease represents the current state of a release
type helmRelease struct {
	Name            string   `json:"Name"`
	Namespace       string   `json:"Namespace"`
	Revision        int      `json:"Revision,string"`
	Updated         HelmTime `json:"Updated"`
	Status          string   `json:"Status"`
	Chart           string   `json:"Chart"`
	AppVersion      string   `json:"AppVersion,omitempty"`
	HelmsmanContext string
}

// getHelmReleases fetches a list of all releases in a k8s cluster
func getHelmReleases() []helmRelease {
	var allReleases []helmRelease
	cmd := command{
		Cmd:         helmBin,
		Args:        []string{"list", "--all", "--max", "0", "--output", "json", "--all-namespaces"},
		Description: "Listing all existing releases...",
	}
	exitCode, result, _ := cmd.exec(debug, verbose)
	if exitCode != 0 {
		log.Fatal("Failed to list all releases: " + result)
	}
	if err := json.Unmarshal([]byte(result), &allReleases); err != nil {
		log.Fatal(fmt.Sprintf("failed to unmarshal Helm CLI output: %s", err))
	}
	return allReleases
}

// uninstall creates the helm command to uninstall an untracked release
func (r *helmRelease) uninstall() {
	cmd := command{
		Cmd:         helmBin,
		Args:        concat([]string{"uninstall", r.Name, "--namespace", r.Namespace}, getDryRunFlags()),
		Description: "Deleting untracked release [ " + r.Name + " ] in namespace [ " + r.Namespace + " ]",
	}

	outcome.addCommand(cmd, -800, nil)
}

// getRevision returns the revision number for an existing helm release
func (rs *helmRelease) getRevision() string {
	return strconv.Itoa(rs.Revision)
}

// getChartName extracts and returns the Helm chart name from the chart info in a release state.
// example: chart in release state is "jenkins-0.9.0" and this function will extract "jenkins" from it.
func (rs *helmRelease) getChartName() string {

	chart := rs.Chart
	runes := []rune(chart)
	return string(runes[0:strings.LastIndexByte(chart[0:strings.IndexByte(chart, '.')], '-')])
}

// getChartVersion extracts and returns the Helm chart version from the chart info in a release state.
// example: chart in release state is returns "jenkins-0.9.0" and this functions will extract "0.9.0" from it.
// It should also handle semver-valid pre-release/meta information, example: in: jenkins-0.9.0-1, out: 0.9.0-1
// in the event of an error, an empty string is returned.
func (rs *helmRelease) getChartVersion() string {
	chart := rs.Chart
	re := regexp.MustCompile(`-(v?[0-9]+\.[0-9]+\.[0-9]+.*)`)
	matches := re.FindStringSubmatch(chart)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// getCurrentNamespaceProtection returns the protection state for the namespace where a release is currently installed.
// It returns true if a namespace is defined as protected in the desired state file, false otherwise.
func (rs *helmRelease) getCurrentNamespaceProtection() bool {
	return s.Namespaces[rs.Namespace].Protected
}
