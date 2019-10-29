package main

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
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
	Command       command
	Priority      int
	targetRelease *release
}

// plan type representing the plan of actions to make the desired state come true.
type plan struct {
	Commands  []orderedCommand
	Decisions []orderedDecision
	Created   time.Time
}

// createPlan initializes an empty plan
func createPlan() plan {

	p := plan{
		Commands:  []orderedCommand{},
		Decisions: []orderedDecision{},
		Created:   time.Now().UTC(),
	}
	return p
}

// addCommand adds a command type to the plan
func (p *plan) addCommand(cmd command, priority int, r *release) {
	oc := orderedCommand{
		Command:       cmd,
		Priority:      priority,
		targetRelease: r,
	}

	p.Commands = append(p.Commands, oc)
}

// addDecision adds a decision type to the plan
func (p *plan) addDecision(decision string, priority int, decisionType decisionType) {
	od := orderedDecision{
		Description: decision,
		Priority:    priority,
		Type:        decisionType,
	}
	p.Decisions = append(p.Decisions, od)
}

// execPlan executes the commands (actions) which were added to the plan.
func (p plan) execPlan() {
	p.sortPlan()
	if len(p.Commands) > 0 {
		logs.Info("Executing the plan ... ")
	} else {
		logs.Info("Nothing to execute.")
	}

	for _, cmd := range p.Commands {
		logs.Notice(cmd.Command.Description)
		if exitCode, msg, _ := cmd.Command.exec(debug, verbose); exitCode != 0 {
			var errorMsg string
			if errorMsg = msg; !verbose {
				errorMsg = strings.Split(msg, "---")[0]
			}
			logError(fmt.Sprintf("Command returned with exit code: %d. And error message: %s ", exitCode, errorMsg))
		} else {
			logs.Notice(msg)
			if cmd.targetRelease != nil && !dryRun {
				labelResource(cmd.targetRelease)
			}
			logs.Notice("Finished " + cmd.Command.Description)
			if _, err := url.ParseRequestURI(s.Settings.SlackWebhook); err == nil {
				notifySlack(cmd.Command.Description+" ... SUCCESS!", s.Settings.SlackWebhook, false, true)
			}
		}
	}

	if len(p.Commands) > 0 {
		logs.Info("Plan applied.")
	}
}

// printPlanCmds prints the actual commands that will be executed as part of a plan.
func (p plan) printPlanCmds() {
	fmt.Println("Printing the commands of the current plan ...")
	for _, cmd := range p.Commands {
		fmt.Println(cmd.Command.Args[1])
	}
}

// printPlan prints the decisions made in a plan.
func (p plan) printPlan() {
	logs.Notice("-------- PLAN starts here --------------")
	for _, decision := range p.Decisions {
		if decision.Type == ignored {
			logs.Info(decision.Description + " -- priority: " + strconv.Itoa(decision.Priority))
		} else if decision.Type == noop {
			logs.Info(decision.Description+" -- priority: "+strconv.Itoa(decision.Priority))
		} else if decision.Type == delete {
			logs.Warning(decision.Description+" -- priority: "+strconv.Itoa(decision.Priority))
		} else {
			logs.Notice(decision.Description+" -- priority: "+strconv.Itoa(decision.Priority))
		}
	}
	logs.Notice("-------- PLAN ends here --------------")
}

// sendPlanToSlack sends the description of plan commands to slack if a webhook is provided.
func (p plan) sendPlanToSlack() {
	if _, err := url.ParseRequestURI(s.Settings.SlackWebhook); err == nil {
		str := ""
		for _, c := range p.Commands {
			str = str + c.Command.Description + "\n"
		}

		notifySlack(strings.TrimRight(str, "\n"), s.Settings.SlackWebhook, false, false)
	}

}

// sortPlan sorts the slices of commands and decisions based on priorities
// the lower the priority value the earlier a command should be attempted
func (p plan) sortPlan() {
	logs.Debug("Sorting the commands in the plan based on priorities (order flags) ... ")

	sort.SliceStable(p.Commands, func(i, j int) bool {
		return p.Commands[i].Priority < p.Commands[j].Priority
	})

	sort.SliceStable(p.Decisions, func(i, j int) bool {
		return p.Decisions[i].Priority < p.Decisions[j].Priority
	})
}
