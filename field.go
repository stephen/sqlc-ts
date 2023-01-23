package main

import (
	"fmt"
	"strings"
)

type Field struct {
	Name              string // CamelCased name for Go
	DBName            string // Name as used in the DB
	Type              string
	TypecheckTemplate string
	Comment           string
}

func (f Field) Typecheck(i int) string {
	return strings.ReplaceAll(f.TypecheckTemplate, "%", fmt.Sprintf("row[%d]", i))
}

func SetCaseStyle(name string, style string) string {
	switch style {
	case "camel":
		return toCamelCase(name)
	case "pascal":
		return toPascalCase(name)
	case "snake":
		return toSnakeCase(name)
	default:
		panic(fmt.Sprintf("unsupported JSON tags case style: '%s'", style))
	}
}

func toSnakeCase(s string) string {
	return strings.ToLower(s)
}

func toCamelCase(s string) string {
	return toCamelInitCase(s, false)
}

func toPascalCase(s string) string {
	return toCamelInitCase(s, true)
}

func toCamelInitCase(name string, initUpper bool) string {
	out := ""
	for i, p := range strings.Split(name, "_") {
		if !initUpper && i == 0 {
			out += p
			continue
		}
		if p == "id" {
			out += "ID"
		} else {
			out += strings.Title(p)
		}
	}
	return out
}
