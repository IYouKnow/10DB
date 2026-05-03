package sqlgen

import (
	"strings"
	"testing"

	"github.com/pedro/10db-launch/apps/server/internal/models"
)

func TestGenerateSQL(t *testing.T) {
	blueprint := models.SchemaBlueprint{
		Version:   1,
		ProjectID: "proj_1",
		Tables: []models.TableBlueprint{
			{
				ID:   "tbl_users",
				Name: "users",
				Columns: []models.ColumnBlueprint{
					{ID: "col_id", Name: "id", Type: "id", PrimaryKey: true, Unique: true, Nullable: false, Config: models.ColumnConfig{}},
					{ID: "col_email", Name: "email", Type: "varchar", Nullable: false, Unique: true, Config: models.ColumnConfig{VarcharLength: 255}},
				},
			},
		},
	}

	sql, err := Generate(blueprint)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	for _, fragment := range []string{
		`CREATE EXTENSION IF NOT EXISTS pgcrypto;`,
		`CREATE TABLE "users"`,
		`"id" BIGSERIAL NOT NULL`,
		`PRIMARY KEY ("id")`,
		`UNIQUE ("email")`,
	} {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("Generate() missing fragment %q in SQL:\n%s", fragment, sql)
		}
	}
}
