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
	for _, Cmd := range p.Commands {
		log.Println("INFO: attempting: --  ", Cmd.Description)
		Cmd.exec()
	}
}

func (p plan) printPlanCmds() {
	fmt.Println("Printing the commands of the current plan ...")
	for _, Cmd := range p.Commands {
		fmt.Println(Cmd.Description)
	}
}

func (p plan) printPlan() {
	fmt.Printf("Printing the current plan which was generated at: %s \n", p.Created)
	for _, decision := range p.Decisions {
		fmt.Println(decision)
	}
}
