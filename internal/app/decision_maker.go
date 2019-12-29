package app

import (
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

	cs := currentState(make(map[string]helmRelease))
	rel := getHelmReleases()

	var wg sync.WaitGroup
	for _, r := range rel {
		wg.Add(1)
		go func(c currentState, r helmRelease, wg *sync.WaitGroup) {
			defer wg.Done()
			r.HelmsmanContext = getReleaseContext(r.Name, r.Namespace)
			c[r.key()] = r
		}(cs, r, &wg)
	}
	wg.Wait()
	return cs
}

// makePlan creates a plan of the actions needed to make the desired state come true.
func (cs currentState) makePlan(s *state) *plan {
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
func (cs currentState) decide(r *release, s *state, wg *sync.WaitGroup) {
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

			if r.isProtected(cs) {

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
		if !r.isProtected(cs) {
			cs.inspectUpgradeScenario(r) // upgrade or move

		} else {
			logDecision("Release [ "+r.Name+" ] in namespace [ "+r.Namespace+" ] is PROTECTED. Operations are not allowed on this release until "+
				"you remove its protection.", r.Priority, noop)
		}

	} else if ok := cs.releaseExists(r, "deleted"); ok {
		if !r.isProtected(cs) {

			r.rollback(cs) // rollback

		} else {
			logDecision("Release [ "+r.Name+" ] in namespace [ "+r.Namespace+" ] is PROTECTED. Operations are not allowed on this release until "+
				"you remove its protection.", r.Priority, noop)
		}

	} else if ok := cs.releaseExists(r, "failed"); ok {

		if !r.isProtected(cs) {

			logDecision("Release [ "+r.Name+" ] in namespace [ "+r.Namespace+" ] is in FAILED state. Upgrade is scheduled!", r.Priority, change)
			r.upgrade()

		} else {
			logDecision("Release [ "+r.Name+" ] in namespace [ "+r.Namespace+" ] is PROTECTED. Operations are not allowed on this release until "+
				"you remove its protection.", r.Priority, noop)
		}
	} else {
		// If there is no release in the cluster with this name and in this namespace, then install it!
		if _, ok := cs[r.key()]; !ok {
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
func (cs currentState) releaseExists(r *release, status string) bool {
	v, ok := cs[r.key()]
	if !ok || v.HelmsmanContext != s.Context {
		return false
	}

	if status != "" {
		return v.Status == status
	}
	return true
}

var resourceNameExtractor = regexp.MustCompile(`(^\w+\/|\.v\d+$)`)
var releaseNameExtractor = regexp.MustCompile(`sh\.helm\.release\.v\d+\.`)

// getHelmsmanReleases returns a map of all releases that are labeled with "MANAGED-BY=HELMSMAN"
// The releases are categorized by the namespaces in which they are deployed
// The returned map format is: map[<namespace>:map[<helmRelease>:true]]
func (cs currentState) getHelmsmanReleases() map[string]map[string]bool {
	var lines []string
	releases := make(map[string]map[string]bool)
	storageBackend := s.Settings.StorageBackend

	for ns := range s.Namespaces {
		cmd := kubectl([]string{"get", storageBackend, "-n", ns, "-l", "MANAGED-BY=HELMSMAN", "-l", "HELMSMAN_CONTEXT=" + s.Context, "-o", "name"}, "Getting Helmsman-managed releases in namespace [ "+ns+" ]")

		exitCode, output, _ := cmd.exec(debug, verbose)

		if exitCode != 0 {
			log.Fatal(output)
		}
		if !strings.EqualFold("No resources found.", strings.TrimSpace(output)) {
			lines = strings.Split(output, "\n")
		}

		for _, name := range lines {
			if name == "" {
				continue
			}
			name = resourceNameExtractor.ReplaceAllString(name, "")
			name = releaseNameExtractor.ReplaceAllString(name, "")
			if _, ok := releases[ns]; !ok {
				releases[ns] = make(map[string]bool)
			}
			releases[ns][name] = false
			for _, app := range s.Apps {
				if app.Name == name && app.Namespace == ns {
					releases[ns][name] = true
				}
			}
		}
	}
	return releases
}

// cleanUntrackedReleases checks for any releases that are managed by Helmsman and are no longer tracked by the desired state
// It compares the currently deployed releases labeled with "MANAGED-BY=HELMSMAN" with Apps defined in the desired state
// For all untracked releases found, a decision is made to uninstall them and is added to the Helmsman plan
// NOTE: Untracked releases don't benefit from either namespace or application protection.
// NOTE: Removing/Commenting out an app from the desired state makes it untracked.
func (cs currentState) cleanUntrackedReleases() {
	toDelete := 0
	log.Info("Checking if any Helmsman managed releases are no longer tracked by your desired state ...")
	for ns, hr := range cs.getHelmsmanReleases() {
		for name, tracked := range hr {
			if !tracked {
				toDelete++
				r := cs[name+"-"+ns]
				logDecision("Untracked release [ "+r.Name+" ] found and it will be deleted", -800, delete)
				r.uninstall()
			}
		}
	}
	if toDelete == 0 {
		log.Info("No untracked releases found")
	}
}

// inspectUpgradeScenario evaluates if a release should be upgraded.
// - If the release is already in the same namespace specified in the input,
// it will be upgraded using the values file specified in the release info.
// - If the release is already in the same namespace specified in the input but is using a different chart,
// it will be purge deleted and installed in the same namespace using the new chart.
// - If the release is NOT in the same namespace specified in the input,
// it will be purge deleted and installed in the new namespace.
func (cs currentState) inspectUpgradeScenario(r *release) {

	rs, ok := cs[r.key()]
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
