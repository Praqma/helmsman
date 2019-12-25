package app

import (
	"net/url"
	"strings"

	"github.com/Praqma/helmsman/internal/gcs"
)

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
