// go:build exclude

package main

import (
	"encoding/json"
	"os"

	"github.com/Praqma/helmsman/internal/app"
	"github.com/invopop/jsonschema"
)

func main() {
	r := new(jsonschema.Reflector)
	r.AllowAdditionalProperties = true
	if err := r.AddGoComments("github.com/Praqma/helmsman", "./internal/app"); err != nil {
		panic(err)
	}
	s := r.Reflect(&app.State{})
	data, _ := json.MarshalIndent(s, "", "  ")
	os.WriteFile("schema.json", data, 0o644)
}
