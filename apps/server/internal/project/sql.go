package project

import (
	"fmt"
	"sort"
	"strings"

	types "github.com/pedro/10db-launch/apps/server/internal/types"
)

func Generate(bp types.SchemaBlueprint) (string, error) {
	var statements []string
	tables := append([]types.TableBlueprint(nil), bp.Tables...)
	sort.Slice(tables, func(i, j int) bool { return tables[i].Name < tables[j].Name })

	for _, table := range tables {
		var defs []string
		var uniqueConstraints []string
		var primaryKey string
		for _, column := range table.Columns {
			def, isPK, isUnique, err := columnSQL(column)
			if err != nil {
				return "", err
			}
			defs = append(defs, def)
			if isPK {
				primaryKey = fmt.Sprintf("PRIMARY KEY (%s)", quoteIdent(column.Name))
			}
			if isUnique && !column.PrimaryKey {
				uniqueConstraints = append(uniqueConstraints, fmt.Sprintf("UNIQUE (%s)", quoteIdent(column.Name)))
			}
		}
		if primaryKey != "" {
			defs = append(defs, primaryKey)
		}
		defs = append(defs, uniqueConstraints...)
		statements = append(statements, fmt.Sprintf("CREATE TABLE %s (\n  %s\n);", quoteIdent(table.Name), strings.Join(defs, ",\n  ")))
	}

	for _, table := range tables {
		for _, fk := range table.ForeignKeys {
			statement := fmt.Sprintf(
				"ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s) ON DELETE %s ON UPDATE %s;",
				quoteIdent(table.Name),
				quoteIdent(fk.ID),
				quoteIdent(fk.ColumnNames[0]),
				quoteIdent(fk.RefTable),
				quoteIdent(fk.RefColumnNames[0]),
				strings.ToUpper(fk.OnDelete),
				strings.ToUpper(fk.OnUpdate),
			)
			statements = append(statements, statement)
		}
	}
	return "BEGIN;\nCREATE EXTENSION IF NOT EXISTS pgcrypto;\n" + strings.Join(statements, "\n") + "\nCOMMIT;\n", nil
}

func columnSQL(column types.ColumnBlueprint) (string, bool, bool, error) {
	columnType, err := mapType(column)
	if err != nil {
		return "", false, false, err
	}
	parts := []string{quoteIdent(column.Name), columnType}
	if !column.Nullable || column.PrimaryKey || column.Type == "id" {
		parts = append(parts, "NOT NULL")
	}
	if column.Default != nil && column.Type != "id" {
		parts = append(parts, "DEFAULT "+defaultSQL(*column.Default))
	}
	return strings.Join(parts, " "), column.PrimaryKey || column.Type == "id", column.Unique, nil
}

func mapType(column types.ColumnBlueprint) (string, error) {
	switch column.Type {
	case "id":
		return "BIGSERIAL", nil
	case "uuid":
		return "UUID", nil
	case "text":
		return "TEXT", nil
	case "varchar":
		return fmt.Sprintf("VARCHAR(%d)", column.Config.VarcharLength), nil
	case "integer":
		return "INTEGER", nil
	case "decimal":
		return fmt.Sprintf("DECIMAL(%d,%d)", column.Config.DecimalPrecision, column.Config.DecimalScale), nil
	case "boolean":
		return "BOOLEAN", nil
	case "timestamp":
		return "TIMESTAMPTZ", nil
	case "date":
		return "DATE", nil
	case "jsonb":
		return "JSONB", nil
	default:
		return "", fmt.Errorf("unsupported type %s", column.Type)
	}
}

func defaultSQL(def types.DefaultValue) string {
	switch def.Kind {
	case "expression":
		return def.Value
	case "number":
		return def.Value
	case "boolean":
		return strings.ToUpper(def.Value)
	case "json":
		return fmt.Sprintf("'%s'::jsonb", strings.ReplaceAll(def.Value, "'", "''"))
	default:
		return fmt.Sprintf("'%s'", strings.ReplaceAll(def.Value, "'", "''"))
	}
}

func quoteIdent(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}
