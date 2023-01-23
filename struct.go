package main

import (
	"strings"

	"github.com/stephen/sqlc-ts/internal/plugin"
	"github.com/stephen/sqlc-ts/internal/sdk"
)

type Struct struct {
	Table   plugin.Identifier
	Name    string
	Fields  []Field
	Comment string
}

func StructName(name string, settings *plugin.Settings) string {
	if rename := settings.Rename[name]; rename != "" {
		return rename
	}
	out := ""
	for _, p := range strings.Split(name, "_") {
		if p == "id" {
			out += p
		} else {
			out += strings.Title(p)
		}
	}
	return out
}

func FieldName(name string, settings *plugin.Settings) string {
	if rename := settings.Rename[name]; rename != "" {
		return rename
	}
	out := ""
	for i, p := range strings.Split(name, "_") {
		if i == 0 {
			out += sdk.LowerTitle(p)
		} else {
			out += sdk.Title(p)
		}
	}
	return out
}
