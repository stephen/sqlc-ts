package main

import (
	"bufio"
	"bytes"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/stephen/sqlc-sql.js/internal/plugin"
	"github.com/stephen/sqlc-sql.js/internal/sdk"
)

func Generate(req *plugin.CodeGenRequest) (*plugin.CodeGenResponse, error) {
	enums := buildEnums(req)
	structs := buildStructs(req)
	queries, err := buildQueries(req, structs)
	if err != nil {
		return nil, errors.Errorf("error generating queries: %w", err)
	}
	return generate(req, enums, structs, queries)
}

type tmplCtx struct {
	Q       string
	Enums   []Enum
	Structs []Struct
	Queries []Query

	// XXX: race
	SourceName string
}

func (t *tmplCtx) OutputQuery(sourceName string) bool {
	return t.SourceName == sourceName
}

func generate(req *plugin.CodeGenRequest, enums []Enum, structs []Struct, queries []Query) (*plugin.CodeGenResponse, error) {
	// i := &importer{
	// 	Settings: req.Settings,
	// 	Queries:  queries,
	// 	Enums:    enums,
	// 	Structs:  structs,
	// }

	funcMap := template.FuncMap{
		"lowerTitle": sdk.LowerTitle,
		"comment":    sdk.DoubleSlashComment,
		"escape":     sdk.EscapeBacktick,
		// "imports":    i.Imports,
		"hasPrefix": strings.HasPrefix,
	}

	tmpl := template.Must(
		template.New("table").
			Funcs(funcMap).
			ParseFS(
				templates,
				"templates/*.tmpl",
			),
	)

	tctx := tmplCtx{
		Q:       "`",
		Queries: queries,
		Enums:   enums,
		Structs: structs,
	}

	output := map[string]string{}

	execute := func(name, templateName string) error {
		var b bytes.Buffer
		w := bufio.NewWriter(&b)
		tctx.SourceName = name
		err := tmpl.ExecuteTemplate(w, templateName, &tctx)
		w.Flush()
		if err != nil {
			return err
		}
		code := b.Bytes()
		if !strings.HasSuffix(name, ".ts") {
			name += ".ts"
		}
		output[name] = string(code)
		return nil
	}

	files := map[string]struct{}{}
	for _, gq := range queries {
		files[gq.SourceName] = struct{}{}
	}

	for source := range files {
		if err := execute(source, "queryFile"); err != nil {
			return nil, err
		}
	}
	resp := plugin.CodeGenResponse{}

	for filename, code := range output {
		resp.Files = append(resp.Files, &plugin.File{
			Name:     filename,
			Contents: []byte(code),
		})
	}

	return &resp, nil
}
