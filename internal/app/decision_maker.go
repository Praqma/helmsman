package app

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

var outcome plan

type currentState map[string]helmRelease

// logDecision adds the decisions made to the plan.
// Depending on the debug flag being set or not, it will either log the the decision to output or not.
func logDecision(decision string, priority int, decisionType decisionType) {
	outcome.addDecision(decision, priority, decisionType)
}

// buildState builds the currentState map containing information about all releases existing in a k8s cluster
func buildState() currentState {
	log.Info("Acquiring current Helm state from cluster...")

	cs := make(map[string]helmRelease)
	rel := getHelmReleases()

	for _, r := range rel {
		r.HelmsmanContext = getReleaseContext(r.Name, r.Namespace)
		cs[fmt.Sprintf("%s-%s", r.Name, r.Namespace)] = r
	}
	return cs
}

// makePlan creates a plan of the actions needed to make the desired state come true.
func (cs *currentState) makePlan(s *state) *plan {
	outcome = createPlan()

	wg := sync.WaitGroup{}
	for _, r := range s.Apps {
		r.checkChartDepUpdate()
		wg.Add(1)
		go cs.decide(r, s, &wg)
	}
	wg.Wait()

	return &outcome
}

// decide makes a decision about what commands (actions) need to be executed
// to make a release section of the desired state come true.
func (cs *currentState) decide(r *release, s *state, wg *sync.WaitGroup) {
	defer wg.Done()
	// check for presence in defined targets or groups
	if !r.isConsideredToRun() {
		logDecision("Release [ "+r.Name+" ] ignored", r.Priority, ignored)
		return
	}

	if destroy {
		if ok := cs.releaseExists(r, ""); ok {
			r.uninstall()
		}
		return
	}

	if !r.Enabled {
		if ok := cs.releaseExists(r, ""); ok {

			if cs.isProtected(r) {

				logDecision("Release [ "+r.Name+" ] in namespace [ "+r.Namespace+" ] is PROTECTED. Operations are not allowed on this release until "+
					"protection is removed.", r.Priority, noop)
				return
			}
			r.uninstall()
			return
		}
		logDecision("Release [ "+r.Name+" ] disabled", r.Priority, noop)
		return

	}

	if ok := cs.releaseExists(r, "deployed"); ok {
		if !cs.isProtected(r) {
			cs.inspectUpgradeScenario(r) // upgrade or move

		} else {
			logDecision("Release [ "+r.Name+" ] in namespace [ "+r.Namespace+" ] is PROTECTED. Operations are not allowed on this release until "+
				"you remove its protection.", r.Priority, noop)
		}

	} else if ok := cs.releaseExists(r, "deleted"); ok {
		if !cs.isProtected(r) {

			r.rollback(cs) // rollback

		} else {
			logDecision("Release [ "+r.Name+" ] in namespace [ "+r.Namespace+" ] is PROTECTED. Operations are not allowed on this release until "+
				"you remove its protection.", r.Priority, noop)
		}

	} else if ok := cs.releaseExists(r, "failed"); ok {

		if !cs.isProtected(r) {

			logDecision("Release [ "+r.Name+" ] in namespace [ "+r.Namespace+" ] is in FAILED state. Upgrade is scheduled!", r.Priority, change)
			r.upgrade()

		} else {
			logDecision("Release [ "+r.Name+" ] in namespace [ "+r.Namespace+" ] is PROTECTED. Operations are not allowed on this release until "+
				"you remove its protection.", r.Priority, noop)
		}
	} else {
		// If there is no release in the cluster with this name and in this namespace, then install it!
		if _, ok := (*cs)[fmt.Sprintf("%s-%s", r.Name, r.Namespace)]; !ok {
			r.install()
		} else {
			// A release with the same name and in the same namespace exists, but it has a different context label (managed by another DSF)
			log.Fatal("Release [ " + r.Name + " ] in namespace [ " + r.Namespace + " ] already exists but is not managed by the" +
				" current context: [ " + s.Context + " ]. Applying changes will likely cause conflicts. Change the release name or namespace.")
		}
	}
}

// releaseExists checks if a Helm release is/was deployed in a k8s cluster.
// It searches the Current State for releases.
// The key format for releases uniqueness is:  <release name - release namespace>
// If status is provided as an input [deployed, deleted, failed], then the search will verify the release status matches the search status.
func (cs *currentState) releaseExists(r *release, status string) bool {
	v, ok := (*cs)[fmt.Sprintf("%s-%s", r.Name, r.Namespace)]
	if !ok || v.HelmsmanContext != s.Context {
		return false
	}

	if status != "" {
		return v.Status == status
	}
	return true
}

// getHelmsmanReleases returns a map of all releases that are labeled with "MANAGED-BY=HELMSMAN"
// The releases are categorized by the namespaces in which they are deployed
// The returned map format is: map[<namespace>:map[<helmRelease>:true]]
func (cs *currentState) getHelmsmanReleases() map[string]map[helmRelease]bool {
	var lines []string
	releases := make(map[string]map[helmRelease]bool)
	storageBackend := s.Settings.StorageBackend

	for ns := range s.Namespaces {
		cmd := command{
			Cmd:         "kubectl",
			Args:        []string{"get", storageBackend, "-n", ns, "-l", "MANAGED-BY=HELMSMAN", "-l", "HELMSMAN_CONTEXT=" + s.Context, "-o", "name"},
			Description: "Getting Helmsman-managed releases in namespace [ " + ns + " ]",
		}

		exitCode, output, _ := cmd.exec(debug, verbose)

		if exitCode != 0 {
			log.Fatal(output)
		}
		if strings.EqualFold("No resources found.", strings.TrimSpace(output)) {
			lines = strings.Split(output, "\n")
		}

		for _, r := range lines {
			if r == "" {
				continue
			}
			if _, ok := releases[ns]; !ok {
				releases[ns] = make(map[helmRelease]bool)
			}
			releaseName := strings.Split(strings.Split(r, "/")[1], ".")[4]
			releases[ns][(*cs)[releaseName+"-"+ns]] = true
		}
	}
	return releases
}

// cleanUntrackedReleases checks for any releases that are managed by Helmsman and are no longer tracked by the desired state
// It compares the currently deployed releases labeled with "MANAGED-BY=HELMSMAN" with Apps defined in the desired state
// For all untracked releases found, a decision is made to uninstall them and is added to the Helmsman plan
// NOTE: Untracked releases don't benefit from either namespace or application protection.
// NOTE: Removing/Commenting out an app from the desired state makes it untracked.
func (cs *currentState) cleanUntrackedReleases() {
	toDelete := make(map[string]map[helmRelease]bool)
	log.Info("Checking if any Helmsman managed releases are no longer tracked by your desired state ...")
	for ns, releases := range cs.getHelmsmanReleases() {
		for r := range releases {
			tracked := false
			for _, app := range s.Apps {
				if app.Name == r.Name && app.Namespace == r.Namespace {
					tracked = true
				}
			}
			if !tracked {
				if _, ok := toDelete[ns]; !ok {
					toDelete[ns] = make(map[helmRelease]bool)
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
				r.uninstall()
			}
		}
	}
}

// inspectUpgradeScenario evaluates if a release should be upgraded.
// - If the release is already in the same namespace specified in the input,
// it will be upgraded using the values file specified in the release info.
// - If the release is already in the same namespace specified in the input but is using a different chart,
// it will be purge deleted and installed in the same namespace using the new chart.
// - If the release is NOT in the same namespace specified in the input,
// it will be purge deleted and installed in the new namespace.
func (cs *currentState) inspectUpgradeScenario(r *release) {

	rs, ok := (*cs)[fmt.Sprintf("%s-%s", r.Name, r.Namespace)]
	if !ok {
		return
	}

	if r.Namespace == rs.Namespace {

		version, msg := r.getChartVersion()
		if msg != "" {
			log.Fatal(msg)
			return
		}
		r.Version = version

		if extractChartName(r.Chart) == rs.getChartName() && r.Version != rs.getChartVersion() {
			// upgrade
			r.diff()
			r.upgrade()
			logDecision("Release [ "+r.Name+" ] will be updated", r.Priority, change)

		} else if extractChartName(r.Chart) != rs.getChartName() {
			r.reInstall(rs)
			logDecision("Release [ "+r.Name+" ] is desired to use a new chart [ "+r.Chart+
				" ]. Delete of the current release will be planned and new chart will be installed in namespace [ "+
				r.Namespace+" ]", r.Priority, change)
		} else {
			if diff := r.diff(); diff != "" {
				r.upgrade()
				logDecision("Release [ "+r.Name+" ] will be updated", r.Priority, change)
			} else {
				logDecision("Release [ "+r.Name+" ] installed and up-to-date", r.Priority, noop)
			}
		}
	} else {
		r.reInstall(rs)
		logDecision("Release [ "+r.Name+" ] is desired to be enabled in a new namespace [ "+r.Namespace+
			" ]. Uninstall of the current release from namespace [ "+rs.Namespace+" ] will be performed "+
			"and then installation in namespace [ "+r.Namespace+" ] will take place", r.Priority, change)
		logDecision("WARNING: moving release [ "+r.Name+" ] from [ "+rs.Namespace+" ] to [ "+r.Namespace+
			" ] might not correctly connect existing volumes. Check https://github.com/Praqma/helmsman/blob/master/docs/how_to/move_charts_across_namespaces.md"+
			" for details if this release uses PV and PVC.", r.Priority, change)
	}
}

// isProtected checks if a release is protected or not.
// A protected is release is either: a) deployed in a protected namespace b) flagged as protected in the desired state file
// Any release in a protected namespace is protected by default regardless of its flag
// returns true if a release is protected, false otherwise
func (cs *currentState) isProtected(r *release) bool {

	// if the release does not exist in the cluster, it is not protected
	if ok := cs.releaseExists(r, ""); !ok {
		return false
	}

	if s.Namespaces[r.Namespace].Protected || r.Protected {
		return true
	}

	return false

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
