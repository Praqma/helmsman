package main

import (
	"os"

	"github.com/Praqma/helmsman/internal/app"
)

func main() {
	exitCode := app.Main()
	os.Exit(exitCode)
}
