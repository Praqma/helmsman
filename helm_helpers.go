package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"helmsman/gcs"
)

var currentState map[string]releaseState

// releaseState represents the current state of a release
type releaseState struct {
	Revision        int
	Updated         time.Time
	Status          string
	Chart           string
	Namespace       string
}

type releaseInfo struct {
	Name            string `json:"Name"`
	Namespace       string `json:"Namespace"`
	Revision        string `json:"Revision"`
	Updated         string `json:"Updated"`
	Status          string `json:"Status"`
	Chart           string `json:"Chart"`
	AppVersion      string `json:"AppVersion,omitempty"`
}

type chartVersion struct {
	Name            string `json:"name"`
	Version         string `json:"version"`
	AppVersion      string `json:"app_version"`
	Description     string `json:"description"`
}

// getHelmClientVersion returns Helm client Version
func getHelmVersion() string {
	cmd := command{
		Cmd:         helmBin,
		Args:        []string{"version", "--short"},
		Description: "checking Helm version ",
	}

	exitCode, result, _ := cmd.exec(debug, false)
	if exitCode != 0 {
		logError("while checking helm version: " + result)
	}
	return result
}

// getHelmReleases fetches a list of all releases in a k8s cluster
func getHelmReleases() []releaseInfo {
	var allReleases []releaseInfo
	cmd := command{
		Cmd:         helmBin,
		Args:        []string{"list", "--all", "--max", "0", "--output", "json", "--all-namespaces"},
		Description: "listing all existing releases...",
	}
	exitCode, result, _ := cmd.exec(debug, verbose)
	if exitCode != 0 {
		logError("failed to list all releases: " + result)
	}
	if err := json.Unmarshal([]byte(result), &allReleases); err != nil {
		logError(fmt.Sprintf("failed to unmarshal Helm CLI output: %s", err))
	}
	return allReleases
}

// buildState builds the currentState map containing information about all releases existing in a k8s cluster
func buildState() {
	logs.Info("Acquiring current Helm state from cluster...")

	currentState = make(map[string]releaseState)
	rel := getHelmReleases()

	for i := 0; i < len(rel); i++ {
		// we need to split the time into parts and make sure milliseconds len = 6, it happens to skip trailing zeros
		updatedFields := strings.Fields(rel[i].Updated)
		updatedHour := strings.Split(updatedFields[1], ".")
		milliseconds := updatedHour[1]
		for i := len(milliseconds); i < 9; i++ {
			milliseconds = fmt.Sprintf("%s0", milliseconds)
		}
		date, err := time.Parse("2006-01-02 15:04:05.000000000 -0700 MST",
			fmt.Sprintf("%s %s.%s %s %s", updatedFields[0], updatedHour[0], milliseconds, updatedFields[2], updatedFields[3]))
		if err != nil {
			logError("while converting release time: " + err.Error())
		}
		revision, _ := strconv.Atoi(rel[i].Revision)
		currentState[fmt.Sprintf("%s-%s", rel[i].Name, rel[i].Namespace)] = releaseState{
			Revision:        revision,
			Updated:         date,
			Status:          rel[i].Status,
			Chart:           rel[i].Chart,
			Namespace:       rel[i].Namespace,
		}
	}
}

// helmRealseExists checks if a Helm release is/was deployed in a k8s cluster.
// It searches the Current State for releases.
// The key format for releases uniqueness is:  <release name - the Tiller namespace where it should be deployed >
// If status is provided as an input [deployed, deleted, failed], then the search will verify the release status matches the search status.
func isReleaseExisting(r *release, status string) bool {
	v, ok := currentState[fmt.Sprintf("%s-%s", r.Name, r.Namespace)]
	if !ok {
		return false
	}

	if status != "" {
		if v.Status == status {
			return true
		}
		return false
	}
	return true
}

// getReleaseRevision returns the revision number for a release
func getReleaseRevision(rs releaseState) string {
	return strconv.Itoa(rs.Revision)
}

// getReleaseChartName extracts and returns the Helm chart name from the chart info in a release state.
// example: chart in release state is "jenkins-0.9.0" and this function will extract "jenkins" from it.
func getReleaseChartName(rs releaseState) string {

	chart := rs.Chart
	runes := []rune(chart)
	return string(runes[0:strings.LastIndexByte(chart[0:strings.IndexByte(chart, '.')], '-')])
}

// getReleaseChartVersion extracts and returns the Helm chart version from the chart info in a release state.
// example: chart in release state is returns "jenkins-0.9.0" and this functions will extract "0.9.0" from it.
// It should also handle semver-valid pre-release/meta information, example: in: jenkins-0.9.0-1, out: 0.9.0-1
// in the event of an error, an empty string is returned.
func getReleaseChartVersion(rs releaseState) string {
	chart := rs.Chart
	re := regexp.MustCompile("-(v?[0-9]+\\.[0-9]+\\.[0-9]+.*)")
	matches := re.FindStringSubmatch(chart)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// validateReleaseCharts validates if the charts defined in a release are valid.
// Valid charts are the ones that can be found in the defined repos.
// This function uses Helm search to verify if the chart can be found or not.
func validateReleaseCharts(apps map[string]*release) (bool, string) {
	versionExtractor := regexp.MustCompile(`version:\s?(.*)`)

	for app, r := range apps {
		validateCurrentChart := true
		if !r.isReleaseConsideredToRun() {
			validateCurrentChart = false
		}
		if validateCurrentChart {
			if isLocalChart(r.Chart) {
				cmd := command{
					Cmd:         helmBin,
					Args:        []string{"inspect", "chart", r.Chart},
					Description: "validating if chart at " + r.Chart + " is available.",
				}

				var output string
				var exitCode int
				if exitCode, output, _ = cmd.exec(debug, verbose); exitCode != 0 {
					maybeRepo := filepath.Base(filepath.Dir(r.Chart))
					return false, "chart at " + r.Chart + " for app [" + app + "] could not be found. Did you mean to add a repo named '" + maybeRepo + "'?"
				}
				matches := versionExtractor.FindStringSubmatch(output)
				if len(matches) == 2 {
					version := matches[1]
					if r.Version != version {
						return false, "chart " + r.Chart + " with version " + r.Version + " is specified for " +
							"app [" + app + "] but the chart found at that path has version " + version + " which does not match."
					}
				}

			} else {
				version := r.Version
				if len(version) == 0 {
					version = "*"
				}
				cmd := command{
					Cmd:         helmBin,
					Args:        []string{"search", "repo", r.Chart, "--version", version, "-l"},
					Description: "validating if chart " + r.Chart + " with version " + r.Version + " is available in the defined repos.",
				}

				if exitCode, result, _ := cmd.exec(debug, verbose); exitCode != 0 || strings.Contains(result, "No results found") {
					return false, "chart " + r.Chart + " with version " + r.Version + " is specified for " +
						"app [" + app + "] but is not found in the defined repos."
				}
			}
		}
	}
	return true, ""
}

// getChartVersion fetches the lastest chart version matching the semantic versioning constraints.
// If chart is local, returns the given release version
func getChartVersion(r *release) (string, string) {
	if isLocalChart(r.Chart) {
		return r.Version, ""
	}
	cmd := command{
		Cmd:         helmBin,
		Args:        []string{"search", "repo", r.Chart, "--version", r.Version, "-o", "json"},
		Description: "getting latest chart version " + r.Chart + "-" + r.Version + "",
	}

	var (
		exitCode int
		result   string
	)

	if exitCode, result, _ = cmd.exec(debug, verbose); exitCode != 0 {
		return "", "chart " + r.Chart + " with version " + r.Version + " is specified but not found in the helm repos."
	}

	chartVersions := make([]chartVersion, 0)
	if err := json.Unmarshal([]byte(result), &chartVersions); err != nil {
		logs.Fatal(fmt.Sprint(err))
	}

	if len(chartVersions) < 1 {
		return "", "chart " + r.Chart + " with version " + r.Version + " is specified but not found in the helm repos."
	} else if len(chartVersions) > 1 {
		return "", "multiple versions of chart " + r.Chart + " with version " + r.Version + " found in the helm repos."
	}
	return chartVersions[0].Version, ""
}


// addHelmRepos adds repositories to Helm if they don't exist already.
// Helm does not mind if a repo with the same name exists. It treats it as an update.
func addHelmRepos(repos map[string]string) (bool, string) {

	for repoName, repoLink := range repos {
		basicAuthArgs := []string{}
		// check if repo is in GCS, then perform GCS auth -- needed for private GCS helm repos
		// failed auth would not throw an error here, as it is possible that the repo is public and does not need authentication
		if strings.HasPrefix(repoLink, "gs://") {
			msg, err := gcs.Auth()
			if err != nil {
				log.Fatal(msg)
			}
		}

		u, err := url.Parse(repoLink)
		if err != nil {
			logError("failed to add helm repo:  " + err.Error())
		}
		if u.User != nil {
			p, ok := u.User.Password()
			if !ok {
				logError("helm repo " + repoName + " has incomplete basic auth info. Missing the password!")
			}
			basicAuthArgs = append(basicAuthArgs, "--username", u.User.Username(), "--password", p)

		}

		cmd := command{
			Cmd:         helmBin,
			Args:        concat([]string{"repo", "add", repoName, repoLink}, basicAuthArgs),
			Description: "adding repo " + repoName,
		}

		if exitCode, err, _ := cmd.exec(debug, verbose); exitCode != 0 {
			return false, "while adding repo [" + repoName + "]: " + err
		}

	}

	cmd := command{
		Cmd:         helmBin,
		Args:        []string{"repo", "update"},
		Description: "updating helm repos",
	}

	if exitCode, err, _ := cmd.exec(debug, verbose); exitCode != 0 {
		return false, "while updating helm repos : " + err
	}

	return true, ""
}


// cleanUntrackedReleases checks for any releases that are managed by Helmsman and are no longer tracked by the desired state
// It compares the currently deployed releases with "MANAGED-BY=HELMSMAN" labels with Apps defined in the desired state
// For all untracked releases found, a decision is made to delete them and is added to the Helmsman plan
// NOTE: Untracked releases don't benefit from either namespace or application protection.
// NOTE: Removing/Commenting out an app from the desired state makes it untracked.
func cleanUntrackedReleases() {
	toDelete := make(map[string]map[*release]bool)
	logs.Info("Checking if any Helmsman managed releases are no longer tracked by your desired state ...")
	for ns, releases := range getHelmsmanReleases() {
		for r := range releases {
			tracked := false
			for _, app := range s.Apps {
				if app.Name == r.Name && app.Namespace == r.Namespace {
					tracked = true
				}
			}
			if !tracked {
				if _, ok := toDelete[ns]; !ok {
					toDelete[ns] = make(map[*release]bool)
				}
				toDelete[ns][r] = true
			}
		}
	}

	if len(toDelete) == 0 {
		logs.Info("No untracked releases found.")
	} else {
		for _, releases := range toDelete {
			for r := range releases {
				if r.isReleaseConsideredToRun() {
					logDecision("Untracked release [ "+r.Name+" ] is ignored by target flag. Skipping.", -800, ignored)
				} else {
					logDecision("Untracked release found: release [ "+r.Name+" ]. It will be deleted", -800, delete)
					deleteUntrackedRelease(r)
				}
			}
		}
	}
}

// deleteUntrackedRelease creates the helm command to purge delete an untracked release
func deleteUntrackedRelease(release *release) {
	cmd := command{
		Cmd:         helmBin,
		Args:        concat([]string{"delete", release.Name, "--namespace", release.Namespace}, getDryRunFlags()),
		Description: "deleting not tracked release [ "+release.Name+" ] in namespace [[ "+release.Namespace+" ]]",
	}

	outcome.addCommand(cmd, -800, nil)
}

// decrypt a helm secret file
func decryptSecret(name string) bool {
	cmd := helmBin
	args := []string{"secrets", "dec", name}

	if settings.EyamlEnabled {
		cmd = "eyaml"
		args = []string{"decrypt", "-f", name}
		if settings.EyamlPrivateKeyPath != "" && settings.EyamlPublicKeyPath != "" {
			args = append(args, []string{"--pkcs7-private-key", settings.EyamlPrivateKeyPath, "--pkcs7-public-key", settings.EyamlPublicKeyPath}...)
		}
	}

	command := command{
		Cmd:         cmd,
		Args:        args,
		Description: "Decrypting " + name,
	}

	exitCode, output, stderr := command.exec(debug, false)
	if !settings.EyamlEnabled {
		_, fileNotFound := os.Stat(name + ".dec")
		if fileNotFound != nil && !isOfType(name, []string{".dec"}) {
			logs.Error(output)
			return false
		}
	}

	if exitCode != 0 {
		logs.Error(output)
		return false
	} else if stderr != "" {
		logs.Error(stderr)
		return false
	}

	if settings.EyamlEnabled {
		var outfile string
		if isOfType(name, []string{".dec"}) {
			outfile = name
		} else {
			outfile = name + ".dec"
		}
		err := writeStringToFile(outfile, output)
		if err != nil {
			logError("could not write [ " + outfile + " ] file")
		}
	}
	return true
}

// updateChartDep updates dependencies for a local chart
func updateChartDep(chartPath string) (bool, string) {
	cmd := command{
		Cmd:         helmBin,
		Args:        []string{"dependency", "update", chartPath},
		Description: "Updating dependency for local chart " + chartPath,
	}

	exitCode, err, _ := cmd.exec(debug, verbose)

	if exitCode != 0 {
		return false, err
	}
	return true, ""
}
