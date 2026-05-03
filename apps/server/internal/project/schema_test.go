package project

import (
	"testing"

	types "github.com/pedro/10db-launch/apps/server/internal/types"
)

func TestValidateBlueprint(t *testing.T) {
	blueprint := types.SchemaBlueprint{
		Version:   1,
		ProjectID: "proj_1",
		Tables: []types.TableBlueprint{
			{
				ID:   "tbl_users",
				Name: "users",
				Columns: []types.ColumnBlueprint{
					{ID: "col_id", Name: "id", Type: "id", PrimaryKey: true, Unique: true, Nullable: false, Config: types.ColumnConfig{}},
					{ID: "col_email", Name: "email", Type: "varchar", Nullable: false, Unique: true, Config: types.ColumnConfig{VarcharLength: 255}},
				},
			},
		},
	}

	errs := Validate(blueprint)
	if len(errs) != 0 {
		t.Fatalf("Validate() errors = %v, want none", errs)
	}
}

func TestValidateBlueprintRejectsBadIdentifiers(t *testing.T) {
	blueprint := types.SchemaBlueprint{
		Version: 1,
		Tables: []types.TableBlueprint{
			{
				Name: "Users",
				Columns: []types.ColumnBlueprint{
					{Name: "Bad Name", Type: "text"},
				},
			},
		},
	}

	errs := Validate(blueprint)
	if len(errs) == 0 {
		t.Fatalf("Validate() errors = none, want validation failures")
	}
}
