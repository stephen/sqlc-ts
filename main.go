package main

import (
	"log"

	"github.com/stephen/sqlc-sql.js/internal/plugin"
)

func main() {
	req := &plugin.CodeGenRequest{}
	// XXX: fail on req.Settings.Engine
	// XXX: fail on type overrides
	structs, err := buildStructs(req)
	if err != nil {
		panic(err)
	}

	queries, err := buildQueries(req, structs)
	if err != nil {
		panic(err)
	}

	log.Println(queries)
}
