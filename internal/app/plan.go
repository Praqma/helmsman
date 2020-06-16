package app

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// decisionType type representing type of Decision for console output
type decisionType int

const (
	create decisionType = iota + 1
	change
	delete
	noop
	ignored
)

// orderedDecision type representing a Decision and it's priority weight
type orderedDecision struct {
	Description string
	Priority    int
	Type        decisionType
}

// orderedCommand type representing a Command and it's priority weight and the targeted release from the desired state
type orderedCommand struct {
	Command        command
	Priority       int
	targetRelease  *release
	beforeCommands []command
	afterCommands  []command
}

// plan type representing the plan of actions to make the desired state come true.
type plan struct {
	sync.Mutex
	Commands  []orderedCommand
	Decisions []orderedDecision
	Created   time.Time
}

// createPlan initializes an empty plan
func createPlan() *plan {
	return &plan{
		Commands:  []orderedCommand{},
		Decisions: []orderedDecision{},
		Created:   time.Now().UTC(),
	}
}

// addCommand adds a command type to the plan
func (p *plan) addCommand(cmd command, priority int, r *release, beforeCommands []command, afterCommands []command) {
	p.Lock()
	defer p.Unlock()
	oc := orderedCommand{
		Command:        cmd,
		Priority:       priority,
		targetRelease:  r,
		beforeCommands: beforeCommands,
		afterCommands:  afterCommands,
	}

	p.Commands = append(p.Commands, oc)
}

// addDecision adds a decision type to the plan
func (p *plan) addDecision(decision string, priority int, decisionType decisionType) {
	p.Lock()
	defer p.Unlock()
	od := orderedDecision{
		Description: decision,
		Priority:    priority,
		Type:        decisionType,
	}
	p.Decisions = append(p.Decisions, od)
}

// exec executes the commands (actions) which were added to the plan.
func (p *plan) exec() {
	p.sort()
	if len(p.Commands) > 0 {
		log.Info("Executing plan... ")
	} else {
		log.Info("Nothing to execute")
	}

	wg := sync.WaitGroup{}
	sem := make(chan struct{}, flags.parallel)
	var fail bool
	var priorities []int
	pl := make(map[int][]orderedCommand)
	for _, cmd := range p.Commands {
		pl[cmd.Priority] = append(pl[cmd.Priority], cmd)
	}
	for priority := range pl {
		priorities = append(priorities, priority)
	}
	sort.Ints(priorities)

	for _, priority := range priorities {
		errorsChan := make(chan error, len(pl[priority]))
		for _, cmd := range pl[priority] {
			sem <- struct{}{}
			wg.Add(1)
			go releaseWithHooks(cmd, &wg, sem, errorsChan)
		}
		wg.Wait()
		close(errorsChan)
		for err := range errorsChan {
			if err != nil {
				fail = true
				log.Error(err.Error())
			}
		}
		if fail {
			log.Fatal("Plan execution failed")
		}
	}

	if len(p.Commands) > 0 {
		log.Info("Plan applied")
	}
}

func releaseWithHooks(cmd orderedCommand, wg *sync.WaitGroup, sem chan struct{}, errors chan error) {
	defer func() {
		wg.Done()
		<-sem
	}()
	for _, c := range cmd.beforeCommands {
		if err := execOne(c, cmd.targetRelease); err != nil {
			errors <- err
			log.Verbose(err.Error())
			return
		}
	}
	if err:= execOne(cmd.Command, cmd.targetRelease); err != nil {
		errors <- err
		log.Verbose(err.Error())
		return
	}
	if cmd.targetRelease != nil && !flags.dryRun && !flags.destroy {
		cmd.targetRelease.label()
	}
	for _, c := range cmd.afterCommands {
		if err := execOne(c, cmd.targetRelease); err != nil {
			errors <- err
			log.Verbose(err.Error())
			return
		}
	}
}

// execOne executes a single ordered command
func execOne(cmd command, targetRelease *release) error {
	log.Notice(cmd.Description)
	result := cmd.exec()
	if result.code != 0 {
		errorMsg := result.errors
		if !flags.verbose {
			errorMsg = strings.Split(result.errors, "---")[0]
		}
		if targetRelease != nil {
			return fmt.Errorf("command for release [%s] returned [ %d ] exit code and error message [ %s ]",
				targetRelease.Name, result.code, strings.TrimSpace(errorMsg))
		} else {
			return fmt.Errorf("%s returned [ %d ] exit code and error message [ %s ]",
					cmd.Description, result.code, strings.TrimSpace(errorMsg))
		}

	} else {
		log.Notice(result.output)
		successMsg := "Finished: " + cmd.Description
		log.Notice(successMsg)
		if _, err := url.ParseRequestURI(settings.SlackWebhook); err == nil {
			notifySlack(cmd.Description+" ... SUCCESS!", settings.SlackWebhook, false, true)
		}
		return nil
	}
}

// printPlanCmds prints the actual commands that will be executed as part of a plan.
func (p *plan) printCmds() {
	log.Info("Printing the commands of the current plan ...")
	for _, cmd := range p.Commands {
		for _, c := range cmd.beforeCommands {
			fmt.Println(c.String())
		}
		fmt.Println(cmd.Command.String())
		for _, c := range cmd.afterCommands {
			fmt.Println(c.String())
		}
	}
}

// printPlan prints the decisions made in a plan.
func (p *plan) print() {
	log.Notice("-------- PLAN starts here --------------")
	for _, decision := range p.Decisions {
		if decision.Type == ignored || decision.Type == noop {
			log.Info(decision.Description + " -- priority: " + strconv.Itoa(decision.Priority))
		} else if decision.Type == delete {
			log.Warning(decision.Description + " -- priority: " + strconv.Itoa(decision.Priority))
		} else {
			log.Notice(decision.Description + " -- priority: " + strconv.Itoa(decision.Priority))
		}
	}
	log.Notice("-------- PLAN ends here --------------")
}

// sendPlanToSlack sends the description of plan commands to slack if a webhook is provided.
func (p *plan) sendToSlack() {
	if _, err := url.ParseRequestURI(settings.SlackWebhook); err == nil {
		str := ""
		for _, c := range p.Commands {
			str = str + c.Command.Description + "\n"
		}
		notifySlack(strings.TrimRight(str, "\n"), settings.SlackWebhook, false, false)
	}
}

// sortPlan sorts the slices of commands and decisions based on priorities
// the lower the priority value the earlier a command should be attempted
func (p *plan) sort() {
	log.Verbose("Sorting the commands in the plan based on priorities (order flags) ... ")

	sort.SliceStable(p.Commands, func(i, j int) bool {
		return p.Commands[i].Priority < p.Commands[j].Priority
	})

	sort.SliceStable(p.Decisions, func(i, j int) bool {
		return p.Decisions[i].Priority < p.Decisions[j].Priority
	})
}
