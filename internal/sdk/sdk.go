package sdk

import "github.com/stephen/sqlc-ts/internal/plugin"

func DataType(n *plugin.Identifier) string {
	if n.Schema != "" {
		return n.Schema + "." + n.Name
	} else {
		return n.Name
	}
}

func SameTableName(tableID, f *plugin.Identifier, defaultSchema string) bool {
	if tableID == nil {
		return false
	}
	schema := tableID.Schema
	if tableID.Schema == "" {
		schema = defaultSchema
	}
	return tableID.Catalog == f.Catalog && schema == f.Schema && tableID.Name == f.Name
}
