# sqlc-ts

sqlc-ts is a [sqlc plugin](https://docs.sqlc.dev/en/stable/guides/plugins.html) to generate queries that are compatible with [@stephen/sql.js](https://github.com/stephen/sql.js).

This plugin is early-stage and is not quite production ready.

## Installation & Usage

```bash
$ go install github.com/kyleconroy/sqlc/cmd/@latest
```

```json5
// sqlc.json
{
  "version": "2",
  "plugins": [
    {
      "name": "ts",
      "process": {
        "cmd": "sqlc-ts"
      }
    }
  ],
   "sql": [
    {
      "schema": "schema.sql",
      "queries": "query.sql",
      "engine": "sqlite",
      "codegen": [
        {
          "out": "gen",
          "plugin": "ts"
        }
      ]
    }
  ]
}
```

See `testdata/` for a full example that can be run with:
```bash
$ sqlc generate -f ./testdata/sqlc.json && cat ./testdata/gen/query.sql.ts | less
```
