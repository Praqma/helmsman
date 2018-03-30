package main

import (
	"fmt"
	"log"
	"time"
)

// plan type representing the plan of actions to make the desired state come true.
type plan struct {
	Commands  []command
	Decisions []string
	Created   time.Time
}

// createPlan initializes an empty plan
func createPlan() plan {

	p := plan{
		Commands:  []command{},
		Decisions: []string{},
		Created:   time.Now(),
	}
	return p
}

// addCommand adds a command type to the plan
func (p *plan) addCommand(c command) {

	p.Commands = append(p.Commands, c)
}

// addDecision adds a decision type to the plan
func (p *plan) addDecision(decision string) {

	p.Decisions = append(p.Decisions, decision)
}

// execPlan executes the commands (actions) which were added to the plan.
func (p plan) execPlan() {
	log.Println("INFO: Executing the following plan ... ")
	p.printPlan()
	for _, cmd := range p.Commands {
		log.Println("INFO: attempting: --  ", cmd.Description)
		if exitCode, msg := cmd.exec(debug); exitCode != 0 {
			log.Fatal("Command returned with exit code: " + string(exitCode) + ". And error message: " + msg)
		}
	}
}

// printPlanCmds prints the actual commands that will be executed as part of a plan.
func (p plan) printPlanCmds() {
	fmt.Println("Printing the commands of the current plan ...")
	for _, Cmd := range p.Commands {
		fmt.Println(Cmd.Description)
	}
}

// printPlan prints the decisions made in a plan.
func (p plan) printPlan() {
	fmt.Println("----------------------")
	fmt.Printf("Plan generated at: %s \n", p.Created.Format("Mon Jan _2 2006 15:04:05"))
	for _, decision := range p.Decisions {
		fmt.Println(decision)
	}
}
