package app

import (
	"regexp"
	"strings"
	"sync"
)

type currentState struct {
	sync.Mutex
	releases map[string]helmRelease
	plan     *plan
}

func newCurrentState() *currentState {
	return &currentState{
		releases: map[string]helmRelease{},
	}
}

// buildState builds the currentState map containing information about all releases existing in a k8s cluster
func buildState() *currentState {
	log.Info("Acquiring current Helm state from cluster...")

	cs := newCurrentState()
	rel := getHelmReleases()

	var wg sync.WaitGroup
	for _, r := range rel {
		wg.Add(1)
		go func(c *currentState, r helmRelease, wg *sync.WaitGroup) {
			c.Lock()
			defer c.Unlock()
			defer wg.Done()
			r.HelmsmanContext = getReleaseContext(r.Name, r.Namespace)
			c.releases[r.key()] = r
		}(cs, r, &wg)
	}
	wg.Wait()
	return cs
}

// makePlan creates a plan of the actions needed to make the desired state come true.
func (cs *currentState) makePlan(s *state) *plan {
	p := createPlan()

	wg := sync.WaitGroup{}
	for _, r := range s.Apps {
		r.checkChartDepUpdate()
		wg.Add(1)
		go cs.decide(r, s, p, &wg)
	}
	wg.Wait()

	return p
}

// decide makes a decision about what commands (actions) need to be executed
// to make a release section of the desired state come true.
func (cs *currentState) decide(r *release, s *state, p *plan, wg *sync.WaitGroup) {
	defer wg.Done()
	// check for presence in defined targets or groups
	if !r.isConsideredToRun(s) {
		p.addDecision("Release [ "+r.Name+" ] ignored", r.Priority, ignored)
		return
	}

	if flags.destroy {
		if ok := cs.releaseExists(r, ""); ok {
			r.uninstall(p)
		}
		return
	}

	if !r.Enabled {
		if ok := cs.releaseExists(r, ""); ok {

			if r.isProtected(cs, s) {

				p.addDecision("Release [ "+r.Name+" ] in namespace [ "+r.Namespace+" ] is PROTECTED. Operations are not allowed on this release until "+
					"protection is removed.", r.Priority, noop)
				return
			}
			r.uninstall(p)
			return
		}
		p.addDecision("Release [ "+r.Name+" ] disabled", r.Priority, noop)
		return

	}

	if ok := cs.releaseExists(r, "deployed"); ok {
		if !r.isProtected(cs, s) {
			cs.inspectUpgradeScenario(r, p) // upgrade or move

		} else {
			p.addDecision("Release [ "+r.Name+" ] in namespace [ "+r.Namespace+" ] is PROTECTED. Operations are not allowed on this release until "+
				"you remove its protection.", r.Priority, noop)
		}

	} else if ok := cs.releaseExists(r, "deleted"); ok {
		if !r.isProtected(cs, s) {

			r.rollback(cs, p) // rollback

		} else {
			p.addDecision("Release [ "+r.Name+" ] in namespace [ "+r.Namespace+" ] is PROTECTED. Operations are not allowed on this release until "+
				"you remove its protection.", r.Priority, noop)
		}

	} else if ok := cs.releaseExists(r, "failed"); ok {

		if !r.isProtected(cs, s) {

			p.addDecision("Release [ "+r.Name+" ] in namespace [ "+r.Namespace+" ] is in FAILED state. Upgrade is scheduled!", r.Priority, change)
			r.upgrade(p)

		} else {
			p.addDecision("Release [ "+r.Name+" ] in namespace [ "+r.Namespace+" ] is PROTECTED. Operations are not allowed on this release until "+
				"you remove its protection.", r.Priority, noop)
		}
	} else {
		// If there is no release in the cluster with this name and in this namespace, then install it!
		if _, ok := cs.releases[r.key()]; !ok {
			r.install(p)
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
	v, ok := cs.releases[r.key()]
	if !ok || v.HelmsmanContext != curContext {
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
func (cs *currentState) getHelmsmanReleases(s *state) map[string]map[string]bool {
	var lines []string
	const outputFmt = "custom-columns=NAME:.metadata.name,NS:.metadata.namespace,CTX:.metadata.labels.HELMSMAN_CONTEXT"
	releases := make(map[string]map[string]bool)
	storageBackend := s.Settings.StorageBackend

	cmd := kubectl([]string{"get", storageBackend, "--all-namespaces", "-l", "MANAGED-BY=HELMSMAN", "-o", outputFmt, "--no-headers"}, "Getting Helmsman-managed releases")
	result := cmd.exec()

	if result.code != 0 {
		log.Fatal(result.errors)
	}
	if !strings.EqualFold("No resources found.", strings.TrimSpace(result.output)) {
		lines = strings.Split(result.output, "\n")
	}

	for _, line := range lines {
		if line == "" {
			continue
		}
		flds := strings.Fields(line)
		name := resourceNameExtractor.ReplaceAllString(flds[0], "")
		name = releaseNameExtractor.ReplaceAllString(name, "")
		ns := flds[1]
		rctx := defaultContextName
		if len(flds) > 2 {
			rctx = flds[2]
		}
		if _, ok := releases[ns]; !ok {
			releases[ns] = make(map[string]bool)
		}
		if !s.isNamespaceDefined(ns) {
			// if the namespace is not managed by this desired state we assume it's tracked
			releases[ns][name] = true
			continue
		}
		if rctx != s.Context {
			// if the release is not related to the current context we assume it's tracked
			releases[ns][name] = true
			continue
		}
		releases[ns][name] = false
		for _, app := range s.Apps {
			if app.Name == name && app.Namespace == ns {
				releases[ns][name] = true
				break
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
func (cs *currentState) cleanUntrackedReleases(s *state, p *plan) {
	toDelete := 0
	log.Info("Checking if any Helmsman managed releases are no longer tracked by your desired state ...")
	for ns, hr := range cs.getHelmsmanReleases(s) {
		for name, tracked := range hr {
			if !tracked {
				toDelete++
				r := cs.releases[name+"-"+ns]
				p.addDecision("Untracked release [ "+r.Name+" ] found and it will be deleted", -800, delete)
				r.uninstall(p)
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
func (cs *currentState) inspectUpgradeScenario(r *release, p *plan) {

	rs, ok := cs.releases[r.key()]
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
			r.upgrade(p)
			p.addDecision("Release [ "+r.Name+" ] will be updated", r.Priority, change)

		} else if extractChartName(r.Chart) != rs.getChartName() {
			r.reInstall(rs, p)
			p.addDecision("Release [ "+r.Name+" ] is desired to use a new chart [ "+r.Chart+
				" ]. Delete of the current release will be planned and new chart will be installed in namespace [ "+
				r.Namespace+" ]", r.Priority, change)
		} else {
			if diff := r.diff(); diff != "" {
				r.upgrade(p)
				p.addDecision("Release [ "+r.Name+" ] will be updated", r.Priority, change)
			} else {
				p.addDecision("Release [ "+r.Name+" ] installed and up-to-date", r.Priority, noop)
			}
		}
	} else {
		r.reInstall(rs, p)
		p.addDecision("Release [ "+r.Name+" ] is desired to be enabled in a new namespace [ "+r.Namespace+
			" ]. Uninstall of the current release from namespace [ "+rs.Namespace+" ] will be performed "+
			"and then installation in namespace [ "+r.Namespace+" ] will take place", r.Priority, change)
		p.addDecision("WARNING: moving release [ "+r.Name+" ] from [ "+rs.Namespace+" ] to [ "+r.Namespace+
			" ] might not correctly connect existing volumes. Check https://github.com/Praqma/helmsman/blob/master/docs/how_to/move_charts_across_namespaces.md"+
			" for details if this release uses PV and PVC.", r.Priority, change)
	}
}
