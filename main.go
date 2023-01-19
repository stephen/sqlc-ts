package main

import (
	"fmt"
	"sort"

	"github.com/jinzhu/inflection"
	"github.com/stephen/sqlc-sql.js/internal/plugin"
)

func main() {
	req := &plugin.CodeGenRequest{}
	// XXX: fail on req.Settings.Engine
	// XXX: fail on type overrides
	buildStructs(req)
}

func buildStructs(req *plugin.CodeGenRequest) ([]Struct, error) {
	var structs []Struct
	for _, schema := range req.Catalog.Schemas {
		for _, table := range schema.Tables {
			var tableName string
			if schema.Name == req.Catalog.DefaultSchema {
				tableName = table.Rel.Name
			} else {
				return nil, fmt.Errorf("sql.js does not support non-default schema")
			}
			// XXX: go codegen has req.Settings.Go.EmitExactTableNames knob.
			structName := inflection.Singular(tableName)
			s := Struct{
				Table:   plugin.Identifier{Schema: schema.Name, Name: table.Rel.Name},
				Name:    StructName(structName, req.Settings),
				Comment: table.Comment,
			}
			for _, column := range table.Columns {
				tags := map[string]string{}
				if req.Settings.Go.EmitDbTags {
					tags["db:"] = column.Name
				}
				if req.Settings.Go.EmitJsonTags {
					tags["json:"] = JSONTagName(column.Name, req.Settings)
				}
				s.Fields = append(s.Fields, Field{
					Name:    StructName(column.Name, req.Settings),
					Type:    tsType(req, column),
					Tags:    tags,
					Comment: column.Comment,
				})
			}
			structs = append(structs, s)
		}
	}
	if len(structs) > 0 {
		sort.Slice(structs, func(i, j int) bool { return structs[i].Name < structs[j].Name })
	}
	return structs, nil
}
