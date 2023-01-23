package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/stephen/sqlc-ts/internal/plugin"
	"github.com/stephen/sqlc-ts/internal/sdk"
)

func tsTypecheckTemplate(req *plugin.CodeGenRequest, col *plugin.Column) string {
	typ := sqliteType(req, col)
	cond := fmt.Sprintf(`typeof %% !== "%s"`, typ)
	if !col.NotNull {
		cond = fmt.Sprintf(`%s && %% !== null`, cond)
	}

	if col.IsArray {
		cond = fmt.Sprintf(`!Array.isArray(%%) || %%.some(e => %s)`, strings.ReplaceAll(cond, "%%", "e"))
	}

	return cond
}

func tsType(req *plugin.CodeGenRequest, col *plugin.Column) string {
	typ := sqliteType(req, col)
	if !col.NotNull {
		typ = fmt.Sprintf("%s | null", typ)
	}

	if col.IsArray {
		typ = fmt.Sprintf("%s[]", typ)
	}

	return typ
}

func sqliteType(req *plugin.CodeGenRequest, col *plugin.Column) string {
	dt := strings.ToLower(sdk.DataType(col.Type))

	switch dt {
	case "int", "integer", "tinyint", "smallint", "mediumint", "bigint", "unsignedbigint", "int2", "int8", "numeric", "decimal", "real", "double", "doubleprecision", "float":
		return "number"

	case "blob":
		return "Uint8Array"

	case "boolean":
		return "boolean"

	case "date", "datetime", "timestamp":
		return "Date"

	case "any":
		// XXX: is this right?
		return "unknown"
	}

	switch {
	case strings.HasPrefix(dt, "character"),
		strings.HasPrefix(dt, "varchar"),
		strings.HasPrefix(dt, "varyingcharacter"),
		strings.HasPrefix(dt, "nchar"),
		strings.HasPrefix(dt, "nativecharacter"),
		strings.HasPrefix(dt, "nvarchar"),
		dt == "text",
		dt == "clob":
		return "string"

	default:
		log.Printf("unknown SQLite type: %s\n", dt)
		// XXX: is this right? or prefer any?
		return "unknown"
	}
}
