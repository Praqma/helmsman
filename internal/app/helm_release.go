package app

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

const (
	helmStatusDeployed = "deployed"
	helmStatusDeleted  = "deleted"
	helmStatusFailed   = "failed"
	helmStatusPending  = "pending-upgrade"
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
func getHelmReleases(s *state) []helmRelease {
	var (
		allReleases []helmRelease
		wg          sync.WaitGroup
		mutex       = &sync.Mutex{}
		namespaces  map[string]namespace
	)
	if len(s.TargetMap) > 0 {
		namespaces = s.TargetNamespaces
	} else {
		namespaces = s.Namespaces
	}
	for ns := range namespaces {
		wg.Add(1)
		go func(ns string) {
			var releases []helmRelease
			var targetReleases []helmRelease
			defer wg.Done()
			cmd := helmCmd([]string{"list", "--all", "--max", "0", "--output", "json", "-n", ns}, "Listing all existing releases in [ "+ns+" ] namespace...")
			result := cmd.exec()
			if result.code != 0 {
				log.Fatal("Failed to list all releases: " + result.errors)
			}
			if err := json.Unmarshal([]byte(result.output), &releases); err != nil {
				log.Fatal(fmt.Sprintf("failed to unmarshal Helm CLI output: %s", err))
			}
			if len(s.TargetMap) > 0 {
				for _, r := range releases {
					if use, ok := s.TargetMap[r.Name]; ok && use {
						targetReleases = append(targetReleases, r)
					}
				}
			} else {
				targetReleases = releases
			}
			mutex.Lock()
			allReleases = append(allReleases, targetReleases...)
			mutex.Unlock()
		}(ns)
	}
	wg.Wait()
	return allReleases
}

func (r *helmRelease) key() string {
	return fmt.Sprintf("%s-%s", r.Name, r.Namespace)
}

// uninstall creates the helm command to uninstall an untracked release
func (r *helmRelease) uninstall(p *plan) {
	cmd := helmCmd(concat([]string{"uninstall", r.Name, "--namespace", r.Namespace}, flags.getDryRunFlags()), "Delete untracked release [ "+r.Name+" ] in namespace [ "+r.Namespace+" ]")

	p.addCommand(cmd, -800, nil)
}

// getRevision returns the revision number for an existing helm release
func (r *helmRelease) getRevision() string {
	return strconv.Itoa(r.Revision)
}

// getChartName extracts and returns the Helm chart name from the chart info in a release state.
// example: chart in release state is "jenkins-0.9.0" and this function will extract "jenkins" from it.
func (r *helmRelease) getChartName() string {

	chart := r.Chart
	runes := []rune(chart)
	return string(runes[0:strings.LastIndexByte(chart[0:strings.IndexByte(chart, '.')], '-')])
}

// getChartVersion extracts and returns the Helm chart version from the chart info in a release state.
// example: chart in release state is returns "jenkins-0.9.0" and this functions will extract "0.9.0" from it.
// It should also handle semver-valid pre-release/meta information, example: in: jenkins-0.9.0-1, out: 0.9.0-1
// in the event of an error, an empty string is returned.
func (r *helmRelease) getChartVersion() string {
	chart := r.Chart
	re := regexp.MustCompile(`-(v?[0-9]+\.[0-9]+\.[0-9]+.*)`)
	matches := re.FindStringSubmatch(chart)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// getCurrentNamespaceProtection returns the protection state for the namespace where a release is currently installed.
// It returns true if a namespace is defined as protected in the desired state file, false otherwise.
func (r *helmRelease) getCurrentNamespaceProtection(s *state) bool {
	return s.Namespaces[r.Namespace].Protected
}
