package dbutil

import "strings"

// QualifiedTable returns a normalized PostgreSQL schema-qualified table name.
func QualifiedTable(schema, table string) string {
	schema = strings.ToLower(strings.TrimSpace(schema))
	if schema == "" {
		return table
	}
	return schema + "." + table
}
