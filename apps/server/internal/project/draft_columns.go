package project

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	types "github.com/pedro/10db-launch/apps/server/internal/types"
)

var (
	draftColumnNamePattern    = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	supportedDraftColumnTypes = map[string]struct{}{
		"text":      {},
		"integer":   {},
		"boolean":   {},
		"uuid":      {},
		"timestamp": {},
		"jsonb":     {},
	}
)

type DraftColumnInput struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Nullable     bool   `json:"nullable"`
	PrimaryKey   bool   `json:"primaryKey"`
	DefaultValue string `json:"defaultValue"`
}

func (s *Service) ListDraftColumns(ctx context.Context, ownerUserID, tableID string) ([]types.DraftTableColumn, error) {
	draftTable, err := s.store.FindOwnedDraftTable(ctx, ownerUserID, tableID)
	if err != nil {
		return nil, err
	}

	columns, err := s.store.ListDraftTableColumns(ctx, tableID)
	if err != nil {
		return nil, err
	}
	if len(columns) == 0 {
		columns = draftColumnsFromTable(draftTable.Table)
		if err := s.store.BackfillDraftTableColumns(ctx, tableID, columns); err != nil {
			return nil, err
		}
	}
	return columns, nil
}

func (s *Service) CreateDraftColumn(ctx context.Context, ownerUserID, tableID string, input DraftColumnInput) (types.DraftTableColumn, error) {
	draftTable, err := s.store.FindOwnedDraftTable(ctx, ownerUserID, tableID)
	if err != nil {
		return types.DraftTableColumn{}, err
	}
	columns, err := s.ListDraftColumns(ctx, ownerUserID, tableID)
	if err != nil {
		return types.DraftTableColumn{}, err
	}
	name, columnType, err := validateDraftColumnInput(input)
	if err != nil {
		return types.DraftTableColumn{}, err
	}
	for _, column := range columns {
		if strings.EqualFold(column.Name, name) {
			return types.DraftTableColumn{}, errors.New("duplicate column name")
		}
	}

	now := time.Now().UTC()
	column := types.DraftTableColumn{
		ID:           uuid.NewString(),
		TableID:      tableID,
		Name:         name,
		Type:         columnType,
		Nullable:     input.Nullable,
		PrimaryKey:   input.PrimaryKey,
		DefaultValue: strings.TrimSpace(input.DefaultValue),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	columns = append(columns, column)
	if err := s.saveDraftColumns(ctx, draftTable, columns); err != nil {
		return types.DraftTableColumn{}, err
	}
	return column, nil
}

func (s *Service) UpdateDraftColumn(ctx context.Context, ownerUserID, tableID, columnID string, input DraftColumnInput) (types.DraftTableColumn, error) {
	draftTable, err := s.store.FindOwnedDraftTable(ctx, ownerUserID, tableID)
	if err != nil {
		return types.DraftTableColumn{}, err
	}
	columns, err := s.ListDraftColumns(ctx, ownerUserID, tableID)
	if err != nil {
		return types.DraftTableColumn{}, err
	}
	name, columnType, err := validateDraftColumnInput(input)
	if err != nil {
		return types.DraftTableColumn{}, err
	}

	found := false
	updatedAt := time.Now().UTC()
	for index := range columns {
		if columns[index].ID == columnID {
			found = true
			columns[index].Name = name
			columns[index].Type = columnType
			columns[index].Nullable = input.Nullable
			columns[index].PrimaryKey = input.PrimaryKey
			columns[index].DefaultValue = strings.TrimSpace(input.DefaultValue)
			columns[index].UpdatedAt = updatedAt
			continue
		}
		if strings.EqualFold(columns[index].Name, name) {
			return types.DraftTableColumn{}, errors.New("duplicate column name")
		}
	}
	if !found {
		return types.DraftTableColumn{}, sql.ErrNoRows
	}

	if err := s.saveDraftColumns(ctx, draftTable, columns); err != nil {
		return types.DraftTableColumn{}, err
	}
	for _, column := range columns {
		if column.ID == columnID {
			return column, nil
		}
	}
	return types.DraftTableColumn{}, sql.ErrNoRows
}

func (s *Service) DeleteDraftColumn(ctx context.Context, ownerUserID, tableID, columnID string) error {
	draftTable, err := s.store.FindOwnedDraftTable(ctx, ownerUserID, tableID)
	if err != nil {
		return err
	}
	columns, err := s.ListDraftColumns(ctx, ownerUserID, tableID)
	if err != nil {
		return err
	}
	nextColumns := make([]types.DraftTableColumn, 0, len(columns))
	found := false
	for _, column := range columns {
		if column.ID == columnID {
			found = true
			continue
		}
		nextColumns = append(nextColumns, column)
	}
	if !found {
		return sql.ErrNoRows
	}
	return s.saveDraftColumns(ctx, draftTable, nextColumns)
}

func (s *Service) saveDraftColumns(ctx context.Context, draftTable ownedDraftTable, columns []types.DraftTableColumn) error {
	blueprint := draftTable.Revision.Blueprint
	for tableIndex := range blueprint.Tables {
		if blueprint.Tables[tableIndex].ID != draftTable.Table.ID {
			continue
		}
		blueprint.Tables[tableIndex].Columns = make([]types.ColumnBlueprint, 0, len(columns))
		for _, column := range columns {
			var defaultValue *types.DefaultValue
			if strings.TrimSpace(column.DefaultValue) != "" {
				defaultValue = &types.DefaultValue{Kind: "expression", Value: column.DefaultValue}
			}
			blueprint.Tables[tableIndex].Columns = append(blueprint.Tables[tableIndex].Columns, types.ColumnBlueprint{
				ID:         column.ID,
				Name:       column.Name,
				Type:       column.Type,
				PrimaryKey: column.PrimaryKey,
				Nullable:   column.Nullable,
				Default:    defaultValue,
				Config:     types.ColumnConfig{},
			})
		}
		break
	}

	hash, err := HashBlueprint(blueprint)
	if err != nil {
		return err
	}
	sqlText, err := Generate(blueprint)
	if err != nil {
		return err
	}
	if _, err := s.store.SaveSchemaRevision(ctx, draftTable.ProjectID, blueprint, hash, sqlText); err != nil {
		return err
	}
	return s.store.ReplaceDraftTableColumns(ctx, draftTable.Table.ID, columns)
}

func validateDraftColumnInput(input DraftColumnInput) (string, string, error) {
	name := strings.TrimSpace(input.Name)
	if !draftColumnNamePattern.MatchString(name) {
		return "", "", fmt.Errorf("invalid column name: %s", name)
	}
	columnType := strings.TrimSpace(input.Type)
	if _, ok := supportedDraftColumnTypes[columnType]; !ok {
		return "", "", fmt.Errorf("unsupported column type: %s", columnType)
	}
	return name, columnType, nil
}

func draftColumnsFromTable(table types.TableBlueprint) []types.DraftTableColumn {
	columns := make([]types.DraftTableColumn, 0, len(table.Columns))
	for _, column := range table.Columns {
		defaultValue := ""
		if column.Default != nil {
			defaultValue = column.Default.Value
		}
		columns = append(columns, types.DraftTableColumn{
			ID:           column.ID,
			TableID:      table.ID,
			Name:         column.Name,
			Type:         column.Type,
			Nullable:     column.Nullable,
			PrimaryKey:   column.PrimaryKey,
			DefaultValue: defaultValue,
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		})
	}
	return columns
}
