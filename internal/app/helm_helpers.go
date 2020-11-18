package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/Masterminds/semver"
	"github.com/Praqma/helmsman/internal/gcs"
)

type helmRepo struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type chartInfo struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

// helmCmd prepares a helm command to be executed
func helmCmd(args []string, desc string) Command {
	return Command{
		Cmd:         helmBin,
		Args:        args,
		Description: desc,
	}
}

// getChartInfo fetches the latest chart information (name, version) matching the semantic versioning constraints.
func getChartInfo(chartName, chartVersion string) (*chartInfo, error) {
	if isLocalChart(chartName) {
		log.Info("Chart [ " + chartName + " ] with version [ " + chartVersion + " ] was found locally.")
	}

	cmd := helmCmd([]string{"show", "chart", chartName, "--version", chartVersion}, "Getting latest non-local chart's version "+chartName+"-"+chartVersion+"")

	result := cmd.Exec()
	if result.code != 0 {
		maybeRepo := filepath.Base(filepath.Dir(chartName))
		message := strings.TrimSpace(result.errors)

		return nil, fmt.Errorf("chart [ %s ] version [ %s ] can't be found. Inspection returned error: \"%s\" -- If this is not a local chart, add the repo [ %s ] in your helmRepos stanza", chartName, chartVersion, message, maybeRepo)
	}

	c := &chartInfo{}
	if err := yaml.Unmarshal([]byte(result.output), &c); err != nil {
		log.Fatal(fmt.Sprint(err))
	}

	constraint, err := semver.NewConstraint(chartVersion)
	if err != nil {
		return nil, err
	}
	found, err := semver.NewVersion(c.Version)
	if err != nil {
		return nil, err
	}

	if !constraint.Check(found) {
		return nil, fmt.Errorf("chart [ %s ] with version [ %s ] was found with a mismatched version: %s", chartName, chartVersion, c.Version)
	}

	return c, nil
}

// getHelmClientVersion returns Helm client Version
func getHelmVersion() string {
	cmd := helmCmd([]string{"version", "--short", "-c"}, "Checking Helm version")

	result := cmd.Exec()
	if result.code != 0 {
		log.Fatal("While checking helm version: " + result.errors)
	}

	return result.output
}

func checkHelmVersion(constraint string) bool {
	helmVersion := strings.TrimSpace(getHelmVersion())
	extractedHelmVersion := helmVersion
	if !strings.HasPrefix(helmVersion, "v") {
		extractedHelmVersion = strings.TrimSpace(strings.Split(helmVersion, ":")[1])
	}
	v, err := semver.NewVersion(extractedHelmVersion)
	if err != nil {
		return false
	}

	jsonConstraint, err := semver.NewConstraint(constraint)
	if err != nil {
		return false
	}
	return jsonConstraint.Check(v)
}

// helmPluginExists returns true if the plugin is present in the environment and false otherwise.
// It takes as input the plugin's name to check if it is recognizable or not. e.g. diff
func helmPluginExists(plugin string) bool {
	if plugins == nil {
		pluginNameRegex := regexp.MustCompile(`^[^\s]+`)
		plugins = make(map[string]bool)
		cmd := helmCmd([]string{"plugin", "list"}, "Listing installed plugins")
		result := cmd.Exec()
		if result.code != 0 {
			log.Fatal("Couldn't get helm plugins: " + result.errors)
		}
		for i, line := range strings.Split(result.output, "\n") {
			if i > 0 {
				if name := pluginNameRegex.FindString(line); name != "" {
					plugins[name] = true
				}
			}
		}
	}

	log.Debug(fmt.Sprintf("Validating that plugin [ %s ] is installed", plugin))
	return plugins[plugin] == true
}

// updateChartDep updates dependencies for a local chart
func updateChartDep(chartPath string) error {
	cmd := helmCmd([]string{"dependency", "update", "--skip-refresh", chartPath}, "Updating dependency for local chart [ "+chartPath+" ]")

	result := cmd.Exec()
	if result.code != 0 {
		return errors.New(result.errors)
	}
	return nil
}

// addHelmRepos adds repositories to Helm if they don't exist already.
// Helm does not mind if a repo with the same name exists. It treats it as an update.
func addHelmRepos(repos map[string]string) error {
	var helmRepos []helmRepo
	existingRepos := make(map[string]string)

	// get existing helm repositories
	cmdList := helmCmd(concat([]string{"repo", "list", "--output", "json"}), "Listing helm repositories")
	if reposResult := cmdList.Exec(); reposResult.code == 0 {
		if err := json.Unmarshal([]byte(reposResult.output), &helmRepos); err != nil {
			log.Fatal(fmt.Sprintf("failed to unmarshal Helm CLI output: %s", err))
		}
		// create map of existing repositories
		for _, repo := range helmRepos {
			existingRepos[repo.Name] = repo.URL
		}
	} else {
		if !strings.Contains(reposResult.errors, "no repositories to show") {
			return fmt.Errorf("while listing helm repositories: %s", reposResult.errors)
		}
	}

	repoAddFlags := ""
	if checkHelmVersion(">=3.3.2") {
		repoAddFlags += "--force-update"
	}

	for repoName, repoLink := range repos {
		basicAuthArgs := []string{}
		// check if repo is in GCS, then perform GCS auth -- needed for private GCS helm repos
		// failed auth would not throw an error here, as it is possible that the repo is public and does not need authentication
		if strings.HasPrefix(repoLink, "gs://") {
			if !helmPluginExists("gcs") {
				log.Fatal(fmt.Sprintf("repository %s can't be used: helm-gcs plugin is missing", repoLink))
			}
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
			u.User = nil
			repoLink = u.String()
		}

		cmd := helmCmd(concat([]string{"repo", "add", repoAddFlags, repoName, repoLink}, basicAuthArgs), "Adding helm repository [ "+repoName+" ]")
		// check current repository against existing repositories map in order to make sure it's missing and needs to be added
		if existingRepoUrl, ok := existingRepos[repoName]; ok {
			if repoLink == existingRepoUrl {
				continue
			}
		}
		if result := cmd.Exec(); result.code != 0 {
			return fmt.Errorf("While adding helm repository ["+repoName+"]: %s", result.errors)
		}
	}

	if len(repos) > 0 {
		cmd := helmCmd([]string{"repo", "update"}, "Updating helm repositories")

		if result := cmd.Exec(); result.code != 0 {
			return errors.New("While updating helm repos : " + result.errors)
		}
	}

	return nil
}
