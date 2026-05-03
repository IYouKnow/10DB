package schema

import (
	"testing"

	"github.com/pedro/10db-launch/apps/server/internal/models"
)

func TestValidateBlueprint(t *testing.T) {
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

	errs := Validate(blueprint)
	if len(errs) != 0 {
		t.Fatalf("Validate() errors = %v, want none", errs)
	}
}

func TestValidateBlueprintRejectsBadIdentifiers(t *testing.T) {
	blueprint := models.SchemaBlueprint{
		Version: 1,
		Tables: []models.TableBlueprint{
			{
				Name: "Users",
				Columns: []models.ColumnBlueprint{
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
