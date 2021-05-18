package app

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
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

	res, err := cmd.Exec()
	if err != nil {
		maybeRepo := filepath.Base(filepath.Dir(chartName))
		return nil, fmt.Errorf("chart [ %s ] version [ %s ] can't be found. If this is not a local chart, add the repo [ %s ] in your helmRepos stanza. Error output: %w", chartName, chartVersion, maybeRepo, err)
	}

	c := &chartInfo{}
	if err := yaml.Unmarshal([]byte(res.output), &c); err != nil {
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
	cmd := helmCmd([]string{"version", "--short", "--client"}, "Checking Helm version")

	res, err := cmd.Exec()
	if err != nil {
		log.Fatalf("While checking helm version: %v", err)
	}

	version := strings.TrimSpace(res.output)
	if !strings.HasPrefix(version, "v") {
		version = strings.SplitN(version, ":", 2)[1]
	}
	return version
}

func checkHelmVersion(constraint string) bool {
	return checkVersion(getHelmVersion(), constraint)
}

// helmPluginExists returns true if the plugin is present in the environment and false otherwise.
// It takes as input the plugin's name to check if it is recognizable or not. e.g. diff
func helmPluginExists(plugin string) bool {
	cmd := helmCmd([]string{"plugin", "list"}, "Validating that [ "+plugin+" ] is installed")

	res, err := cmd.Exec()
	if err != nil {
		return false
	}

	return strings.Contains(res.output, plugin)
}

// updateChartDep updates dependencies for a local chart
func updateChartDep(chartPath string) error {
	cmd := helmCmd([]string{"dependency", "update", "--skip-refresh", chartPath}, "Updating dependency for local chart [ "+chartPath+" ]")

	if _, err := cmd.Exec(); err != nil {
		return err
	}
	return nil
}

// helmExportChart pulls chart and exports it to the specified destination
func helmExportChart(chart, dest string) error {
	cmd := helmCmd([]string{"chart", "pull", chart}, "Pulling chart [ "+chart+" ] to local registry cache")
	if _, err := cmd.Exec(); err != nil {
		return err
	}
	cmd = helmCmd([]string{"chart", "export", chart, "-d", dest}, "Exporting chart [ "+chart+" ] to "+dest)
	if _, err := cmd.Exec(); err != nil {
		return err
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
	if reposResult, err := cmdList.Exec(); err == nil {
		if err := json.Unmarshal([]byte(reposResult.output), &helmRepos); err != nil {
			log.Fatal(fmt.Sprintf("failed to unmarshal Helm CLI output: %s", err))
		}
		// create map of existing repositories
		for _, repo := range helmRepos {
			existingRepos[repo.Name] = repo.URL
		}
	} else if !strings.Contains(reposResult.errors, "no repositories to show") {
		return fmt.Errorf("while listing helm repositories: %w", err)
	}

	forceUpdateFlag := ""
	if checkHelmVersion(">=3.3.2") && !flags.noUpdate {
		forceUpdateFlag += "--force-update"
	}

	for repoName, repoURL := range repos {
		basicAuthArgs := []string{}
		u, err := url.Parse(repoURL)
		if err != nil {
			log.Fatal("failed to add helm repo:  " + err.Error())
		}
		// check if repo is in GCS, then perform GCS auth -- needed for private GCS helm repos
		// failed auth would not throw an error here, as it is possible that the repo is public and does not need authentication
		if u.Scheme == "gs" {
			if !helmPluginExists("gcs") {
				log.Fatal(fmt.Sprintf("repository %s can't be used: helm-gcs plugin is missing", repoURL))
			}
			msg, err := gcs.Auth()
			if err != nil {
				log.Fatal(msg)
			}
		}

		if u.User != nil {
			p, ok := u.User.Password()
			if !ok {
				log.Fatal("helm repo " + repoName + " has incomplete basic auth info. Missing the password!")
			}
			basicAuthArgs = append(basicAuthArgs, "--username", u.User.Username(), "--password", p)
			u.User = nil
			repoURL = u.String()
		}

		// check current repository against existing repositories map in order to make sure it's missing and needs to be added
		if existingRepoURL, ok := existingRepos[repoName]; ok {
			if repoURL == existingRepoURL {
				continue
			}
		}
		cmd := helmCmd(concat([]string{"repo", "add", forceUpdateFlag, repoName, repoURL}, basicAuthArgs), "Adding helm repository [ "+repoName+" ]")
		if _, err := cmd.Exec(); err != nil {
			return fmt.Errorf("while adding helm repository [%s]]: %w", repoName, err)
		}
	}

	if !flags.noUpdate && len(repos) > 0 {
		cmd := helmCmd([]string{"repo", "update"}, "Updating helm repositories")

		if _, err := cmd.Exec(); err != nil {
			return fmt.Errorf("err updating helm repos: %w", err)
		}
	}

	return nil
}
