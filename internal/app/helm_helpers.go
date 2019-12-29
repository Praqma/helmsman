package app

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/Praqma/helmsman/internal/gcs"
)

// helmCmd prepares a helm command to be executed
func helmCmd(args []string, desc string) command {
	return command{
		Cmd:         helmBin,
		Args:        args,
		Description: desc,
	}
}

var chartNameExtractor = regexp.MustCompile(`[\\/]([^\\/]+)$`)

// extractChartName extracts the Helm chart name from full chart name in the desired state.
// example: it extracts "chartY" from "repoX/chartY" and "chartZ" from "c:\charts\chartZ"
func extractChartName(releaseChart string) string {

	m := chartNameExtractor.FindStringSubmatch(releaseChart)
	if len(m) == 2 {
		return m[1]
	}

	return ""
}

// getHelmClientVersion returns Helm client Version
func getHelmVersion() string {
	cmd := helmCmd([]string{"version", "--short", "-c"}, "Checking Helm version")

	exitCode, result, _ := cmd.exec(debug, false)
	if exitCode != 0 {
		log.Fatal("While checking helm version: " + result)
	}
	return result
}

// helmPluginExists returns true if the plugin is present in the environment and false otherwise.
// It takes as input the plugin's name to check if it is recognizable or not. e.g. diff
func helmPluginExists(plugin string) bool {
	cmd := helmCmd([]string{"plugin", "list"}, "Validating that [ "+plugin+" ] is installed")

	exitCode, result, _ := cmd.exec(debug, false)

	if exitCode != 0 {
		return false
	}

	return strings.Contains(result, plugin)
}

// updateChartDep updates dependencies for a local chart
func updateChartDep(chartPath string) (bool, string) {
	cmd := helmCmd([]string{"dependency", "update", chartPath}, "Updating dependency for local chart [ "+chartPath+" ]")

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

		cmd := helmCmd(concat([]string{"repo", "add", repoName, repoLink}, basicAuthArgs), "Adding helm repository [ "+repoName+" ]")

		if exitCode, err, _ := cmd.exec(debug, verbose); exitCode != 0 {
			return false, "While adding helm repository [" + repoName + "]: " + err
		}

	}

	cmd := helmCmd([]string{"repo", "update"}, "Updating helm repositories")

	if exitCode, err, _ := cmd.exec(debug, verbose); exitCode != 0 {
		return false, "While updating helm repos : " + err
	}

	return true, ""
}
