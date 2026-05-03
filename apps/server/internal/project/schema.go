package project

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"

	types "github.com/pedro/10db-launch/apps/server/internal/types"
)

var identPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

var supportedTypes = map[string]struct{}{
	"id":        {},
	"uuid":      {},
	"text":      {},
	"varchar":   {},
	"integer":   {},
	"decimal":   {},
	"boolean":   {},
	"timestamp": {},
	"date":      {},
	"jsonb":     {},
}

func Normalize(bp types.SchemaBlueprint, projectID string) types.SchemaBlueprint {
	bp.Version = 1
	bp.ProjectID = projectID
	if bp.Tables == nil {
		bp.Tables = []types.TableBlueprint{}
	}
	for i := range bp.Tables {
		bp.Tables[i].Name = strings.TrimSpace(strings.ToLower(bp.Tables[i].Name))
		for j := range bp.Tables[i].Columns {
			bp.Tables[i].Columns[j].Name = strings.TrimSpace(strings.ToLower(bp.Tables[i].Columns[j].Name))
		}
	}
	return bp
}

func Validate(bp types.SchemaBlueprint) map[string]string {
	errs := map[string]string{}
	if bp.Version != 1 {
		errs["version"] = "only version 1 is supported"
	}
	tableNames := map[string]struct{}{}
	for ti, table := range bp.Tables {
		key := fmt.Sprintf("tables.%d.name", ti)
		if !identPattern.MatchString(table.Name) {
			errs[key] = "table name must start with a letter and contain only lowercase letters, numbers, and underscores"
		}
		if _, exists := tableNames[table.Name]; exists {
			errs[key] = "duplicate table name"
		}
		tableNames[table.Name] = struct{}{}

		columnNames := map[string]struct{}{}
		pkCount := 0
		for ci, column := range table.Columns {
			colKey := fmt.Sprintf("tables.%d.columns.%d.name", ti, ci)
			if !identPattern.MatchString(column.Name) {
				errs[colKey] = "column name must start with a letter and contain only lowercase letters, numbers, and underscores"
			}
			if _, exists := columnNames[column.Name]; exists {
				errs[colKey] = "duplicate column name"
			}
			columnNames[column.Name] = struct{}{}
			if _, ok := supportedTypes[column.Type]; !ok {
				errs[fmt.Sprintf("tables.%d.columns.%d.type", ti, ci)] = "unsupported column type"
			}
			if column.PrimaryKey {
				pkCount++
			}
			if column.Type == "varchar" && column.Config.VarcharLength <= 0 {
				errs[fmt.Sprintf("tables.%d.columns.%d.config.varcharLength", ti, ci)] = "varchar length must be greater than zero"
			}
			if column.Type == "decimal" && (column.Config.DecimalPrecision <= 0 || column.Config.DecimalScale < 0 || column.Config.DecimalScale > column.Config.DecimalPrecision) {
				errs[fmt.Sprintf("tables.%d.columns.%d.config.decimalPrecision", ti, ci)] = "decimal precision/scale is invalid"
			}
			if column.PrimaryKey && column.Nullable {
				errs[fmt.Sprintf("tables.%d.columns.%d.nullable", ti, ci)] = "primary key columns cannot be nullable"
			}
		}
		if len(table.Columns) == 0 {
			errs[fmt.Sprintf("tables.%d.columns", ti)] = "table must have at least one column"
		}
		if pkCount > 1 {
			errs[fmt.Sprintf("tables.%d.columns", ti)] = "only one primary key column is supported in v1"
		}
		for fi, fk := range table.ForeignKeys {
			fkKey := fmt.Sprintf("tables.%d.foreignKeys.%d", ti, fi)
			if len(fk.ColumnNames) != 1 || len(fk.RefColumnNames) != 1 {
				errs[fkKey] = "only single-column foreign keys are supported"
				continue
			}
			if _, ok := columnNames[fk.ColumnNames[0]]; !ok {
				errs[fkKey+".columnNames.0"] = "foreign key column does not exist"
			}
			if _, ok := tableNames[fk.RefTable]; !ok && fk.RefTable != table.Name {
			}
			if !slices.Contains([]string{"NO ACTION", "CASCADE", "SET NULL", "RESTRICT"}, strings.ToUpper(fk.OnDelete)) {
				errs[fkKey+".onDelete"] = "invalid onDelete action"
			}
			if !slices.Contains([]string{"NO ACTION", "CASCADE", "SET NULL", "RESTRICT"}, strings.ToUpper(fk.OnUpdate)) {
				errs[fkKey+".onUpdate"] = "invalid onUpdate action"
			}
		}
	}

	tableByName := map[string]types.TableBlueprint{}
	for _, table := range bp.Tables {
		tableByName[table.Name] = table
	}
	for ti, table := range bp.Tables {
		for fi, fk := range table.ForeignKeys {
			refTable, ok := tableByName[fk.RefTable]
			if !ok {
				errs[fmt.Sprintf("tables.%d.foreignKeys.%d.refTable", ti, fi)] = "referenced table does not exist"
				continue
			}
			refColumn := fk.RefColumnNames[0]
			found := false
			for _, column := range refTable.Columns {
				if column.Name == refColumn {
					found = true
					break
				}
			}
			if !found {
				errs[fmt.Sprintf("tables.%d.foreignKeys.%d.refColumnNames.0", ti, fi)] = "referenced column does not exist"
			}
		}
	}
	return errs
}

func HashBlueprint(bp types.SchemaBlueprint) (string, error) {
	raw, err := json.Marshal(bp)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}
