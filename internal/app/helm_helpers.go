package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/Praqma/helmsman/internal/gcs"
)

type helmRepo struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

// helmCmd prepares a helm command to be executed
func helmCmd(args []string, desc string) command {
	return command{
		Cmd:         helmBin,
		Args:        args,
		Description: desc,
	}
}

// extractChartName extracts the Helm chart name from full chart name in the desired state.
func extractChartName(releaseChart string) string {
	cmd := helmCmd([]string{"show", "chart", releaseChart}, "Caching chart information for [ "+releaseChart+" ].")

	result := cmd.exec()
	if result.code != 0 {
		log.Fatal("While getting chart information: " + result.errors)
	}

	name := ""
	for _, v := range strings.Split(result.output, "\n") {
		split := strings.Split(v, ":")
		if len(split) == 2 && split[0] == "name" {
			name = strings.Trim(split[1], `"' `)
			break
		}
	}

	return name
}

// getHelmClientVersion returns Helm client Version
func getHelmVersion() string {
	cmd := helmCmd([]string{"version", "--short", "-c"}, "Checking Helm version")

	result := cmd.exec()
	if result.code != 0 {
		log.Fatal("While checking helm version: " + result.errors)
	}

	log.Verbose("Helm version " + result.output)
	return result.output
}

// helmPluginExists returns true if the plugin is present in the environment and false otherwise.
// It takes as input the plugin's name to check if it is recognizable or not. e.g. diff
func helmPluginExists(plugin string) bool {
	cmd := helmCmd([]string{"plugin", "list"}, "Validating that [ "+plugin+" ] is installed")

	result := cmd.exec()

	if result.code != 0 {
		return false
	}

	return strings.Contains(result.output, plugin)
}

// updateChartDep updates dependencies for a local chart
func updateChartDep(chartPath string) error {
	cmd := helmCmd([]string{"dependency", "update", chartPath}, "Updating dependency for local chart [ "+chartPath+" ]")

	result := cmd.exec()
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
	if reposResult := cmdList.exec(); reposResult.code == 0 {
		if err := json.Unmarshal([]byte(reposResult.output), &helmRepos); err != nil {
			log.Fatal(fmt.Sprintf("failed to unmarshal Helm CLI output: %s", err))
		}
		// create map of existing repositories
		for _, repo := range helmRepos {
			existingRepos[repo.Name] = repo.Url
		}
	} else {
		if !strings.Contains(reposResult.errors, "no repositories to show") {
			return fmt.Errorf("while listing helm repositories: %s", reposResult.errors)
		}
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

		}

		cmd := helmCmd(concat([]string{"repo", "add", repoName, repoLink}, basicAuthArgs), "Adding helm repository [ "+repoName+" ]")
		// check current repository against existing repositories map in order to make sure it's missing and needs to be added
		if existingRepoUrl, ok := existingRepos[repoName]; ok {
			if repoLink == existingRepoUrl {
				continue
			}
		}
		if result := cmd.exec(); result.code != 0 {
			return fmt.Errorf("While adding helm repository ["+repoName+"]: %s", result.errors)
		}
	}

	if len(repos) > 0 {
		cmd := helmCmd([]string{"repo", "update"}, "Updating helm repositories")

		if result := cmd.exec(); result.code != 0 {
			return errors.New("While updating helm repos : " + result.errors)
		}
	}

	return nil
}
