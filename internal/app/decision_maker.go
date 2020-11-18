package app

import (
	"fmt"
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
func buildState(s *state) *currentState {
	log.Info("Acquiring current Helm state from cluster")

	cs := newCurrentState()
	rel := getHelmReleases(s)

	wg := sync.WaitGroup{}
	sem := make(chan struct{}, resourcePool)

	for _, r := range rel {
		// aquire
		sem <- struct{}{}
		wg.Add(1)
		go func(r helmRelease) {
			defer func() {
				wg.Done()
				// release
				<-sem
			}()
			if flags.contextOverride == "" {
				r.HelmsmanContext = getReleaseContext(r.Name, r.Namespace, s.Settings.StorageBackend)
			} else {
				r.HelmsmanContext = flags.contextOverride
				log.Info("Overwrote Helmsman context for release [ " + r.Name + " ] to " + flags.contextOverride)
			}
			cs.Lock()
			cs.releases[r.key()] = r
			cs.Unlock()
		}(r)
	}
	wg.Wait()
	return cs
}

// makePlan creates a plan of the actions needed to make the desired state come true.
// TODO: this code needs to be simplified
func (cs *currentState) makePlan(s *state) *plan {
	p := createPlan()
	p.StorageBackend = s.Settings.StorageBackend
	p.ReverseDelete = s.Settings.ReverseDelete

	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}
	sem := make(chan struct{}, resourcePool)

	// We store the results of the helm commands
	extractedChartInfo := make(map[string]map[string]*chartInfo)

	// We get the charts and versions with the expensive helm commands first.
	// We can probably DRY this concurrency stuff up somehow.
	// We can also definitely DRY this with validateReleaseChart.
	// We should probably have a data structure earlier on that sorts this out properly.
	// Ideally we'd have a pipeline of helm command tasks with several stages that can all come home if one of them fails.
	// Is it better to fail early here? I am not sure.

	// Unique chart names and versions in the DSF
	charts := make(map[string]map[string]bool)
	// Initialize the rejigged data structures
	for _, r := range s.Apps {
		if charts[r.Chart] == nil {
			charts[r.Chart] = make(map[string]bool)
		}

		if extractedChartInfo[r.Chart] == nil {
			extractedChartInfo[r.Chart] = make(map[string]*chartInfo)
		}

		if r.isConsideredToRun() {
			charts[r.Chart][r.Version] = true
		}
	}

	// Concurrently extract chart names and versions
	for chart, versions := range charts {
		for version, shouldRun := range versions {
			if !shouldRun {
				continue
			}

			sem <- struct{}{}
			wg.Add(1)
			go func(chart, version string) {
				defer func() {
					wg.Done()
					<-sem
				}()

				info, err := getChartInfo(chart, version)
				if err != nil {
					log.Error(err.Error())
				} else {
					mutex.Lock()
					log.Verbose(fmt.Sprintf("Extracted chart information from chart [ %s ] with version [ %s ]: %s %s", chart, version, info.Name, info.Version))
					extractedChartInfo[chart][version] = info
					mutex.Unlock()
				}
			}(chart, version)
		}
	}

	wg.Wait()

	// Pass the extracted names and versions back to the apps to decide.
	// We still have to run decide on all the apps, even the ones we previously filtered out when extracting names and versions.
	// We can now proceed without trying lots of identical helm commands at the same time.
	for _, r := range s.Apps {
		// To be honest, the helmCmd function should probably pass back a channel at this point, making the resource pool "global", for all helm commands.
		// It would make more sense than parallelising *some of the workload* like we do here with r.checkChartDepUpdate(), leaving some helm commands outside the concurrent part.
		r.checkChartDepUpdate()
		sem <- struct{}{}
		wg.Add(1)
		go func(r *release, c *chartInfo) {
			defer func() {
				wg.Done()
				<-sem
			}()
			cs.decide(r, s.Namespaces[r.Namespace], p, c)
		}(r, extractedChartInfo[r.Chart][r.Version])
	}
	wg.Wait()

	return p
}

// decide makes a decision about what commands (actions) need to be executed
// to make a release section of the desired state come true.
func (cs *currentState) decide(r *release, n *namespace, p *plan, c *chartInfo) {
	// check for presence in defined targets or groups
	if !r.isConsideredToRun() {
		p.addDecision("Release [ "+r.Name+" ] ignored", r.Priority, ignored)
		return
	}

	if flags.destroy {
		if ok := cs.releaseExists(r, ""); ok {
			p.addDecision("Release [ "+r.Name+" ] will be DELETED (destroy flag enabled).", r.Priority, delete)
			r.uninstall(p)
		}
		return
	}

	if !r.Enabled {
		if ok := cs.releaseExists(r, ""); ok {
			if r.isProtected(cs, n) {
				p.addDecision("Release [ "+r.Name+" ] in namespace [ "+r.Namespace+" ] is PROTECTED. Operations are not allowed on this release until "+
					"protection is removed.", r.Priority, noop)
				return
			}
			p.addDecision("Release [ "+r.Name+" ] is desired to be DELETED.", r.Priority, delete)
			r.uninstall(p)
			return
		}
		p.addDecision("Release [ "+r.Name+" ] disabled", r.Priority, noop)
		return
	}

	if ok := cs.releaseExists(r, helmStatusDeployed); ok {
		if !r.isProtected(cs, n) {
			cs.inspectUpgradeScenario(r, p, c) // upgrade or move
		} else {
			p.addDecision("Release [ "+r.Name+" ] in namespace [ "+r.Namespace+" ] is PROTECTED. Operations are not allowed on this release until "+
				"you remove its protection.", r.Priority, noop)
		}
	} else if ok := cs.releaseExists(r, helmStatusUninstalled); ok {
		if !r.isProtected(cs, n) {
			r.rollback(cs, p) // rollback
		} else {
			p.addDecision("Release [ "+r.Name+" ] in namespace [ "+r.Namespace+" ] is PROTECTED. Operations are not allowed on this release until "+
				"you remove its protection.", r.Priority, noop)
		}
	} else if ok := cs.releaseExists(r, helmStatusFailed); ok {
		if !r.isProtected(cs, n) {
			p.addDecision("Release [ "+r.Name+" ] in namespace [ "+r.Namespace+" ] is in FAILED state. Upgrade is scheduled!", r.Priority, change)
			r.upgrade(p)
		} else {
			p.addDecision("Release [ "+r.Name+" ] in namespace [ "+r.Namespace+" ] is PROTECTED. Operations are not allowed on this release until "+
				"you remove its protection.", r.Priority, noop)
		}
	} else if cs.releaseExists(r, helmStatusPendingInstall) || cs.releaseExists(r, helmStatusPendingUpgrade) || cs.releaseExists(r, helmStatusPendingRollback) || cs.releaseExists(r, helmStatusUninstalling) {
		log.Fatal("Release [ " + r.Name + " ] in namespace [ " + r.Namespace + " ] is in a pending (install/upgrade/rollback or uninstalling) state. " +
			"This means application is being operated on outside of this Helmsman invocation's scope." +
			"Exiting, as this may cause issues when continuing...")
	} else {
		// If there is no release in the cluster with this name and in this namespace, then install it!
		if _, ok := cs.releases[r.key()]; !ok {
			p.addDecision("Release [ "+r.Name+" ] version [ "+r.Version+" ] will be installed in [ "+r.Namespace+" ] namespace", r.Priority, create)
			r.install(p)
		} else {
			// A release with the same name and in the same namespace exists, but it has a different context label (managed by another DSF)
			log.Fatal("Release [ " + r.Name + " ] in namespace [ " + r.Namespace + " ] already exists but is not managed by the" +
				" current context. Applying changes will likely cause conflicts. Change the release name or namespace.")
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

var resourceNameExtractor = regexp.MustCompile(`(^\w+/|\.v\d+$)`)
var releaseNameExtractor = regexp.MustCompile(`sh\.helm\.release\.v\d+\.`)

// getHelmsmanReleases returns a map of all releases that are labeled with "MANAGED-BY=HELMSMAN"
// The releases are categorized by the namespaces in which they are deployed
// The returned map format is: map[<namespace>:map[<helmRelease>:true]]
func (cs *currentState) getHelmsmanReleases(s *state) map[string]map[string]bool {
	const outputFmt = "custom-columns=NAME:.metadata.name,CTX:.metadata.labels.HELMSMAN_CONTEXT"
	var (
		wg    sync.WaitGroup
		mutex = &sync.Mutex{}
	)
	releases := make(map[string]map[string]bool)
	sem := make(chan struct{}, resourcePool)

	storageBackend := s.Settings.StorageBackend

	for ns, cfg := range s.Namespaces {
		if cfg.disabled {
			continue
		}
		// acquire
		sem <- struct{}{}
		wg.Add(1)
		go func(ns string) {
			var lines []string
			defer func() {
				wg.Done()
				// release
				<-sem
			}()

			cmd := kubectl([]string{"get", storageBackend, "-n", ns, "-l", "MANAGED-BY=HELMSMAN", "-o", outputFmt, "--no-headers"}, "Getting Helmsman-managed releases from namespace [ "+ns+" ]")
			result := cmd.RetryExec(3)
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
				rctx := defaultContextName
				if len(flds) > 1 {
					rctx = flds[1]
				}
				if len(s.TargetMap) > 0 {
					if use, ok := s.TargetMap[name]; !ok || !use {
						continue
					}
				}
				mutex.Lock()
				if _, ok := releases[ns]; !ok {
					releases[ns] = make(map[string]bool)
				}
				if !s.isNamespaceDefined(ns) || rctx != s.Context {
					// if the namespace is not managed by this desired state
					// or the release is not related to the current context we assume it's tracked
					releases[ns][name] = true
					mutex.Unlock()
					continue
				}
				releases[ns][name] = false
				for _, app := range s.Apps {
					if app.Name == name && app.Namespace == ns {
						releases[ns][name] = true
						break
					}
				}
				mutex.Unlock()
			}

		}(ns)
	}
	wg.Wait()
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
				p.addDecision("Untracked release [ "+r.Name+" ] found and it will be deleted", -1000, delete)
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
// it will be uninstalled and installed in the same namespace using the new chart.
// - If the release is NOT in the same namespace specified in the input,
// it will be purge deleted and installed in the new namespace.
func (cs *currentState) inspectUpgradeScenario(r *release, p *plan, c *chartInfo) {
	if c == nil || c.Name == "" || c.Version == "" {
		return
	}

	rs, ok := cs.releases[r.key()]
	if !ok {
		return
	}

	if r.Namespace == rs.Namespace {
		r.Version = c.Version

		if c.Name == rs.getChartName() && r.Version != rs.getChartVersion() {
			// upgrade
			r.diff()
			r.upgrade(p)
			p.addDecision("Release [ "+r.Name+" ] will be upgraded", r.Priority, change)

		} else if c.Name != rs.getChartName() {
			r.reInstall(p)
			p.addDecision("Release [ "+r.Name+" ] is desired to use a new chart [ "+r.Chart+
				" ]. Delete of the current release will be planned and new chart will be installed in namespace [ "+
				r.Namespace+" ]", r.Priority, change)
		} else {
			if flags.alwaysUpgrade {
				r.upgrade(p)
				p.addDecision("Release [ "+r.Name+" ] will be updated (forced)", r.Priority, change)
			} else if diff := r.diff(); diff != "" {
				r.upgrade(p)
				p.addDecision("Release [ "+r.Name+" ] will be updated", r.Priority, change)
			} else {
				p.addDecision("Release [ "+r.Name+" ] installed and up-to-date", r.Priority, noop)
			}
		}
	} else {
		r.reInstall(p, rs.Namespace)
		p.addDecision("Release [ "+r.Name+" ] is desired to be enabled in a new namespace [ "+r.Namespace+
			" ]. Uninstall of the current release from namespace [ "+rs.Namespace+" ] will be performed "+
			"and then installation in namespace [ "+r.Namespace+" ] will take place", r.Priority, change)
		p.addDecision("WARNING: moving release [ "+r.Name+" ] from [ "+rs.Namespace+" ] to [ "+r.Namespace+
			" ] might not correctly connect existing volumes. Check https://github.com/Praqma/helmsman/blob/master/docs/how_to/apps/moving_across_namespaces.md#note-on-persistent-volumes"+
			" for details if this release uses PV and PVC.", r.Priority, change)
	}
}
