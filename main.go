package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"html/template"
	"log"
	"strings"

	"github.com/stephen/sqlc-sql.js/internal/plugin"
	"github.com/stephen/sqlc-sql.js/internal/sdk"
)

func main() {
	req := &plugin.CodeGenRequest{}
	// XXX: fail on req.Settings.Engine
	// XXX: fail on type overrides
	// XXX: fail on non-default schema
	enums := buildEnums(req)
	structs, err := buildStructs(req)
	if err != nil {
		panic(err)
	}

	queries, err := buildQueries(req, structs)
	if err != nil {
		panic(err)
	}

	log.Println(queries, enums)
}

type tmplCtx struct {
	Q         string
	Package   string
	Enums     []Enum
	Structs   []Struct
	GoQueries []Query

	// TODO: Race conditions
	SourceName string

	EmitJSONTags              bool
	EmitDBTags                bool
	EmitPreparedQueries       bool
	EmitInterface             bool
	EmitEmptySlices           bool
	EmitMethodsWithDBArgument bool
	UsesCopyFrom              bool
	UsesBatch                 bool
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
				"templates/*/*.tmpl",
			),
	)

	golang := req.Settings.Go
	tctx := tmplCtx{
		EmitInterface:             golang.EmitInterface,
		EmitJSONTags:              golang.EmitJsonTags,
		EmitDBTags:                golang.EmitDbTags,
		EmitPreparedQueries:       golang.EmitPreparedQueries,
		EmitEmptySlices:           golang.EmitEmptySlices,
		EmitMethodsWithDBArgument: golang.EmitMethodsWithDbArgument,
		Q:                         "`",
		Package:                   golang.Package,
		GoQueries:                 queries,
		Enums:                     enums,
		Structs:                   structs,
	}

	if tctx.UsesCopyFrom {
		return nil, errors.New(":copyfrom not supported")
	}

	if tctx.UsesBatch {
		return nil, errors.New(":batch* commands not supported")
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
		code, err := format.Source(b.Bytes())
		if err != nil {
			fmt.Println(b.String())
			return fmt.Errorf("source error: %w", err)
		}

		if templateName == "queryFile" && golang.OutputFilesSuffix != "" {
			name += golang.OutputFilesSuffix
		}

		if !strings.HasSuffix(name, ".go") {
			name += ".go"
		}
		output[name] = string(code)
		return nil
	}

	dbFileName := "db.go"
	if golang.OutputDbFileName != "" {
		dbFileName = golang.OutputDbFileName
	}
	modelsFileName := "models.go"
	if golang.OutputModelsFileName != "" {
		modelsFileName = golang.OutputModelsFileName
	}
	querierFileName := "querier.go"
	if golang.OutputQuerierFileName != "" {
		querierFileName = golang.OutputQuerierFileName
	}
	copyfromFileName := "copyfrom.go"
	// TODO(Jille): Make this configurable.

	batchFileName := "batch.go"

	if err := execute(dbFileName, "dbFile"); err != nil {
		return nil, err
	}
	if err := execute(modelsFileName, "modelsFile"); err != nil {
		return nil, err
	}
	if golang.EmitInterface {
		if err := execute(querierFileName, "interfaceFile"); err != nil {
			return nil, err
		}
	}
	if tctx.UsesCopyFrom {
		if err := execute(copyfromFileName, "copyfromFile"); err != nil {
			return nil, err
		}
	}
	if tctx.UsesBatch {
		if err := execute(batchFileName, "batchFile"); err != nil {
			return nil, err
		}
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
