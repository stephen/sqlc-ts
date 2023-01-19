package sdk

import "github.com/stephen/sqlc-sql.js/internal/plugin"

func DataType(n *plugin.Identifier) string {
	if n.Schema != "" {
		return n.Schema + "." + n.Name
	} else {
		return n.Name
	}
}
