package main

import (
	"github.com/Praqma/helmsman/internal/app"
	"os"
)

func main() {
	exitCode := app.Main()
	os.Exit(exitCode)
}
