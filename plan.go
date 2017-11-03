package main

import (
	"fmt"
	"log"
	"time"
)

type plan struct {
	Commands  []command
	Decisions []string
	Created   time.Time
}

func createPlan() plan {

	p := plan{
		Commands:  []command{},
		Decisions: []string{},
		Created:   time.Now(),
	}
	return p
}

func (p *plan) addCommand(c command) {

	p.Commands = append(p.Commands, c)
}

func (p *plan) addDecision(decision string) {

	p.Decisions = append(p.Decisions, decision)
}

func (p plan) execPlan() {
	log.Println("INFO: Executing the following plan ... ")
	p.printPlan()
	for _, cmd := range p.Commands {
		log.Println("INFO: attempting: --  ", cmd.Description)
		cmd.exec(debug)
	}
}

func (p plan) printPlanCmds() {
	fmt.Println("Printing the commands of the current plan ...")
	for _, Cmd := range p.Commands {
		fmt.Println(Cmd.Description)
	}
}

func (p plan) printPlan() {
	fmt.Println("---------------")
	fmt.Printf("Ok, I have generated a plan for you at: %s \n", p.Created)
	for _, decision := range p.Decisions {
		fmt.Println(decision)
	}
}
