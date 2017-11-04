package main

var s state
var debug bool
var file string
var apply bool

func main() {

	p := makePlan(&s)

	if !apply {
		p.printPlan()
	} else {
		p.execPlan()
	}

}
