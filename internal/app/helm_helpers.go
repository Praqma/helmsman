package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Praqma/helmsman/internal/gcs"
)

var currentState map[string]releaseState

// releaseState represents the current state of a release
type releaseState struct {
	Name            string
	Revision        int
	Updated         time.Time
	Status          string
	Chart           string
	Namespace       string
	HelmsmanContext string
}

type releaseInfo struct {
	Name       string `json:"Name"`
	Namespace  string `json:"Namespace"`
	Revision   string `json:"Revision"`
	Updated    string `json:"Updated"`
	Status     string `json:"Status"`
	Chart      string `json:"Chart"`
	AppVersion string `json:"AppVersion,omitempty"`
}

type chartVersion struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	AppVersion  string `json:"app_version"`
	Description string `json:"description"`
}

// getHelmClientVersion returns Helm client Version
func getHelmVersion() string {
	cmd := command{
		Cmd:         helmBin,
		Args:        []string{"version", "--short", "-c"},
		Description: "Checking Helm version",
	}

	exitCode, result, _ := cmd.exec(debug, false)
	if exitCode != 0 {
		log.Fatal("While checking helm version: " + result)
	}
	return result
}

// getHelmReleases fetches a list of all releases in a k8s cluster
func getHelmReleases() []releaseInfo {
	var allReleases []releaseInfo
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

// buildState builds the currentState map containing information about all releases existing in a k8s cluster
func buildState() {
	log.Info("Acquiring current Helm state from cluster...")

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
			log.Fatal("While converting release time: " + err.Error())
		}
		revision, _ := strconv.Atoi(rel[i].Revision)
		currentState[fmt.Sprintf("%s-%s", rel[i].Name, rel[i].Namespace)] = releaseState{
			Name:            rel[i].Name,
			Revision:        revision,
			Updated:         date,
			Status:          rel[i].Status,
			Chart:           rel[i].Chart,
			Namespace:       rel[i].Namespace,
			HelmsmanContext: getReleaseContext(rel[i].Name, rel[i].Namespace),
		}
	}
}

// isReleaseExisting checks if a Helm release is/was deployed in a k8s cluster.
// It searches the Current State for releases.
// The key format for releases uniqueness is:  <release name - release namespace>
// If status is provided as an input [deployed, deleted, failed], then the search will verify the release status matches the search status.
func isReleaseExisting(r *release, status string) bool {
	v, ok := currentState[fmt.Sprintf("%s-%s", r.Name, r.Namespace)]
	if !ok || v.HelmsmanContext != s.Context {
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
func validateReleaseCharts(apps map[string]*release) error {
	versionExtractor := regexp.MustCompile(`version:\s?(.*)`)

	wg := sync.WaitGroup{}
	c := make(chan string, len(apps))
	for app, r := range apps {
		wg.Add(1)
		go func(app string, r *release, wg *sync.WaitGroup, c chan string) {
			defer wg.Done()
			validateCurrentChart := true
			if !r.isReleaseConsideredToRun() {
				validateCurrentChart = false
			}
			if validateCurrentChart {
				if isLocalChart(r.Chart) {
					cmd := command{
						Cmd:         helmBin,
						Args:        []string{"inspect", "chart", r.Chart},
						Description: "Validating [ " + r.Chart + " ] chart's availability",
					}

					var output string
					var exitCode int
					if exitCode, output, _ = cmd.exec(debug, verbose); exitCode != 0 {
						maybeRepo := filepath.Base(filepath.Dir(r.Chart))
						c <- "Chart [ " + r.Chart + " ] for app [" + app + "] can't be found. Did you mean to add a repo [ " + maybeRepo + " ]?"
						return
					}
					matches := versionExtractor.FindStringSubmatch(output)
					if len(matches) == 2 {
						version := matches[1]
						if r.Version != version {
							c <- "Chart [ " + r.Chart + " ] with version [ " + r.Version + " ] is specified for " +
								"app [" + app + "] but the chart found at that path has version [ " + version + " ] which does not match."
							return
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
						Description: "Validating [ " + r.Chart + " ] chart's version [ " + r.Version + " ] availability",
					}

					if exitCode, result, _ := cmd.exec(debug, verbose); exitCode != 0 || strings.Contains(result, "No results found") {
						c <- "Chart [ " + r.Chart + " ] with version [ " + r.Version + " ] is specified for " +
							"app [" + app + "] but was not found"
						return
					}
				}
			}
		}(app, r, &wg, c)
	}
	wg.Wait()
	if len(c) > 0 {
		err := <-c
		if err != "" {
			return errors.New(err)
		}
	}
	return nil
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
		Description: "Getting latest chart's version " + r.Chart + "-" + r.Version + "",
	}

	var (
		exitCode int
		result   string
	)

	if exitCode, result, _ = cmd.exec(debug, verbose); exitCode != 0 {
		return "", "Chart [ " + r.Chart + " ] with version [ " + r.Version + " ] is specified but not found in the helm repositories"
	}

	chartVersions := make([]chartVersion, 0)
	if err := json.Unmarshal([]byte(result), &chartVersions); err != nil {
		log.Fatal(fmt.Sprint(err))
	}

	if len(chartVersions) < 1 {
		return "", "Chart [ " + r.Chart + " ] with version [ " + r.Version + " ] is specified but not found in the helm repositories"
	} else if len(chartVersions) > 1 {
		return "", "Multiple versions of chart [ " + r.Chart + " ] with version [ " + r.Version + " ] found in the helm repositories"
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
			log.Fatal("failed to add helm repo:  " + err.Error())
		}
		if u.User != nil {
			p, ok := u.User.Password()
			if !ok {
				log.Fatal("helm repo " + repoName + " has incomplete basic auth info. Missing the password!")
			}
			basicAuthArgs = append(basicAuthArgs, "--username", u.User.Username(), "--password", p)

		}

		cmd := command{
			Cmd:         helmBin,
			Args:        concat([]string{"repo", "add", repoName, repoLink}, basicAuthArgs),
			Description: "Adding helm repository [ " + repoName + " ]",
		}

		if exitCode, err, _ := cmd.exec(debug, verbose); exitCode != 0 {
			return false, "While adding helm repository [" + repoName + "]: " + err
		}

	}

	cmd := command{
		Cmd:         helmBin,
		Args:        []string{"repo", "update"},
		Description: "Updating helm repositories",
	}

	if exitCode, err, _ := cmd.exec(debug, verbose); exitCode != 0 {
		return false, "While updating helm repos : " + err
	}

	return true, ""
}

// cleanUntrackedReleases checks for any releases that are managed by Helmsman and are no longer tracked by the desired state
// It compares the currently deployed releases labeled with "MANAGED-BY=HELMSMAN" with Apps defined in the desired state
// For all untracked releases found, a decision is made to uninstall them and is added to the Helmsman plan
// NOTE: Untracked releases don't benefit from either namespace or application protection.
// NOTE: Removing/Commenting out an app from the desired state makes it untracked.
func cleanUntrackedReleases() {
	toDelete := make(map[string]map[releaseState]bool)
	log.Info("Checking if any Helmsman managed releases are no longer tracked by your desired state ...")
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
					toDelete[ns] = make(map[releaseState]bool)
				}
				toDelete[ns][r] = true
			}
		}
	}

	if len(toDelete) == 0 {
		log.Info("No untracked releases found")
	} else {
		for _, releases := range toDelete {
			for r := range releases {
				logDecision("Untracked release [ "+r.Name+" ] found and it will be deleted", -800, delete)
				uninstallUntrackedRelease(r)
			}
		}
	}
}

// uninstallUntrackedRelease creates the helm command to uninstall an untracked release
func uninstallUntrackedRelease(release releaseState) {
	cmd := command{
		Cmd:         helmBin,
		Args:        concat([]string{"uninstall", release.Name, "--namespace", release.Namespace}, getDryRunFlags()),
		Description: "Deleting untracked release [ " + release.Name + " ] in namespace [ " + release.Namespace + " ]",
	}

	outcome.addCommand(cmd, -800, nil)
}

// decrypt a helm secret file
func decryptSecret(name string) error {
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
			return errors.New(output)
		}
	}

	if exitCode != 0 {
		return errors.New(output)
	} else if stderr != "" {
		return errors.New(output)
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
			log.Fatal("Can't write [ " + outfile + " ] file")
		}
	}
	return nil
}

// updateChartDep updates dependencies for a local chart
func updateChartDep(chartPath string) (bool, string) {
	cmd := command{
		Cmd:         helmBin,
		Args:        []string{"dependency", "update", chartPath},
		Description: "Updating dependency for local chart [ " + chartPath + " ]",
	}

	exitCode, err, _ := cmd.exec(debug, verbose)

	if exitCode != 0 {
		return false, err
	}
	return true, ""
}

// helmPluginExists returns true if the plugin is present in the environment and false otherwise.
// It takes as input the plugin's name to check if it is recognizable or not. e.g. diff
func helmPluginExists(plugin string) bool {
	cmd := command{
		Cmd:         helmBin,
		Args:        []string{"plugin", "list"},
		Description: "Validating that [ " + plugin + " ] is installed",
	}

	exitCode, result, _ := cmd.exec(debug, false)

	if exitCode != 0 {
		return false
	}

	return strings.Contains(result, plugin)
}
