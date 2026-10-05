// Command openapi prints the OpenAPI document of the HTTP API to stdout.
package main

import (
	"encoding/json"
	"log"
	"os"
	"xarantolus/sensibleHub/web/api"

	"github.com/gorilla/mux"
)

func main() {
	a := api.New(mux.NewRouter(), nil)

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(a.OpenAPI()); err != nil {
		log.Fatal(err)
	}
}
