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
	for _, p := range strings.Split(name, "_") {
		if p == "id" {
			out += p
		} else {
			out += sdk.LowerTitle(p)
		}
	}
	return out
}
