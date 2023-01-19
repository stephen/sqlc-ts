package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jinzhu/inflection"
	"github.com/stephen/sqlc-sql.js/internal/plugin"
	"github.com/stephen/sqlc-sql.js/internal/sdk"
)

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

func buildQueries(req *plugin.CodeGenRequest, structs []Struct) ([]Query, error) {
	qs := make([]Query, 0, len(req.Queries))
	for _, query := range req.Queries {
		if query.Name == "" {
			continue
		}
		if query.Cmd == "" {
			continue
		}

		var constantName string
		if req.Settings.Go.EmitExportedQueries {
			constantName = sdk.Title(query.Name)
		} else {
			constantName = sdk.LowerTitle(query.Name)
		}

		gq := Query{
			Cmd:          query.Cmd,
			ConstantName: constantName,
			FieldName:    sdk.LowerTitle(query.Name) + "Stmt",
			MethodName:   query.Name,
			SourceName:   query.Filename,
			SQL:          query.Text,
			Comments:     query.Comments,
			Table:        query.InsertIntoTable,
		}

		if len(query.Params) == 1 {
			p := query.Params[0]
			gq.Arg = QueryValue{
				Name: paramName(p),
				Typ:  tsType(req, p.Column),
			}
		} else if len(query.Params) > 1 {
			var cols []column
			for _, p := range query.Params {
				cols = append(cols, column{
					id:     int(p.Number),
					Column: p.Column,
				})
			}
			s, err := columnsToStruct(req, gq.MethodName+"Params", cols, false)
			if err != nil {
				return nil, err
			}
			gq.Arg = QueryValue{
				Emit:   true,
				Name:   "arg",
				Struct: s,
			}
		}

		if len(query.Columns) == 1 {
			c := query.Columns[0]
			name := columnName(c, 0)
			if c.IsFuncCall {
				name = strings.Replace(name, "$", "_", -1)
			}
			gq.Ret = QueryValue{
				Name: name,
				Typ:  tsType(req, c),
			}
		} else if len(query.Columns) > 1 {
			var gs *Struct
			var emit bool

			for _, s := range structs {
				if len(s.Fields) != len(query.Columns) {
					continue
				}
				same := true
				for i, f := range s.Fields {
					c := query.Columns[i]
					sameName := f.Name == StructName(columnName(c, i), req.Settings)
					sameType := f.Type == tsType(req, c)
					sameTable := sdk.SameTableName(c.Table, &s.Table, req.Catalog.DefaultSchema)
					if !sameName || !sameType || !sameTable {
						same = false
					}
				}
				if same {
					gs = &s
					break
				}
			}

			if gs == nil {
				var columns []column
				for i, c := range query.Columns {
					columns = append(columns, column{
						id:     i,
						Column: c,
					})
				}
				var err error
				gs, err = columnsToStruct(req, gq.MethodName+"Row", columns, true)
				if err != nil {
					return nil, err
				}
				emit = true
			}
			gq.Ret = QueryValue{
				Emit:   emit,
				Name:   "i",
				Struct: gs,
			}
		}

		qs = append(qs, gq)
	}
	sort.Slice(qs, func(i, j int) bool { return qs[i].MethodName < qs[j].MethodName })
	return qs, nil
}

func columnName(c *plugin.Column, pos int) string {
	if c.Name != "" {
		return c.Name
	}
	return fmt.Sprintf("column_%d", pos+1)
}

func argName(name string) string {
	out := ""
	for i, p := range strings.Split(name, "_") {
		if i == 0 {
			out += strings.ToLower(p)
		} else if p == "id" {
			out += "ID"
		} else {
			out += strings.Title(p)
		}
	}
	return out
}

func paramName(p *plugin.Parameter) string {
	if p.Column.Name != "" {
		return argName(p.Column.Name)
	}
	return fmt.Sprintf("dollar_%d", p.Number)
}

type column struct {
	id int
	*plugin.Column
}

// It's possible that this method will generate duplicate JSON tag values
//
//	Columns: count, count,   count_2
//	 Fields: Count, Count_2, Count2
//
// JSON tags: count, count_2, count_2
//
// This is unlikely to happen, so don't fix it yet
func columnsToStruct(req *plugin.CodeGenRequest, name string, columns []column, useID bool) (*Struct, error) {
	gs := Struct{
		Name: name,
	}
	seen := map[string][]int{}
	suffixes := map[int]int{}
	for i, c := range columns {
		colName := columnName(c.Column, i)
		tagName := colName
		fieldName := StructName(colName, req.Settings)
		baseFieldName := fieldName
		// Track suffixes by the ID of the column, so that columns referring to the same numbered parameter can be
		// reused.
		suffix := 0
		if o, ok := suffixes[c.id]; ok && useID {
			suffix = o
		} else if v := len(seen[fieldName]); v > 0 && !c.IsNamedParam {
			suffix = v + 1
		}
		suffixes[c.id] = suffix
		if suffix > 0 {
			tagName = fmt.Sprintf("%s_%d", tagName, suffix)
			fieldName = fmt.Sprintf("%s_%d", fieldName, suffix)
		}
		tags := map[string]string{}
		if req.Settings.Go.EmitDbTags {
			tags["db:"] = tagName
		}
		if req.Settings.Go.EmitJsonTags {
			tags["json:"] = JSONTagName(tagName, req.Settings)
		}
		gs.Fields = append(gs.Fields, Field{
			Name:   fieldName,
			DBName: colName,
			Type:   tsType(req, c.Column),
			Tags:   tags,
		})
		if _, found := seen[baseFieldName]; !found {
			seen[baseFieldName] = []int{i}
		} else {
			seen[baseFieldName] = append(seen[baseFieldName], i)
		}
	}

	// If a field does not have a known type, but another
	// field with the same name has a known type, assign
	// the known type to the field without a known type
	for i, field := range gs.Fields {
		if len(seen[field.Name]) > 1 && field.Type == "interface{}" {
			for _, j := range seen[field.Name] {
				if i == j {
					continue
				}
				otherField := gs.Fields[j]
				if otherField.Type != field.Type {
					field.Type = otherField.Type
				}
				gs.Fields[i] = field
			}
		}
	}

	err := checkIncompatibleFieldTypes(gs.Fields)
	if err != nil {
		return nil, err
	}

	return &gs, nil
}

func checkIncompatibleFieldTypes(fields []Field) error {
	fieldTypes := map[string]string{}
	for _, field := range fields {
		if fieldType, found := fieldTypes[field.Name]; !found {
			fieldTypes[field.Name] = field.Type
		} else if field.Type != fieldType {
			return fmt.Errorf("named param %s has incompatible types: %s, %s", field.Name, field.Type, fieldType)
		}
	}
	return nil
}
