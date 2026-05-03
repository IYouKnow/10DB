package project

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/pedro/10db-launch/apps/server/internal/platform/crypto"
	"github.com/pedro/10db-launch/apps/server/internal/platform/postgres"
	types "github.com/pedro/10db-launch/apps/server/internal/types"
)

type Service struct {
	store    *Store
	postgres *postgres.Service
	crypto   *crypto.Service
	pgConfig postgres.AdminConfig
}

func New(store *Store, postgresService *postgres.Service, cryptoService *crypto.Service, pgConfig postgres.AdminConfig) *Service {
	return &Service{store: store, postgres: postgresService, crypto: cryptoService, pgConfig: pgConfig}
}

type CreateProjectInput struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func (s *Service) List(ctx context.Context, ownerUserID string) ([]types.Project, error) {
	return s.store.ListProjects(ctx, ownerUserID)
}

func (s *Service) Get(ctx context.Context, ownerUserID, id string) (types.Project, error) {
	return s.store.GetProject(ctx, ownerUserID, id)
}

func (s *Service) Create(ctx context.Context, ownerUserID string, input CreateProjectInput) (types.Project, error) {
	name := strings.TrimSpace(input.Name)
	slug := strings.TrimSpace(strings.ToLower(input.Slug))
	if name == "" || slug == "" {
		return types.Project{}, errors.New("name and slug are required")
	}
	if !slugPattern.MatchString(slug) {
		return types.Project{}, errors.New("slug must contain lowercase letters, numbers, and hyphens only")
	}

	suffix := strings.ReplaceAll(uuid.NewString()[:8], "-", "")
	dbName := fmt.Sprintf("p_%s_%s", strings.ReplaceAll(slug, "-", "_"), suffix)
	roleName := fmt.Sprintf("u_%s_%s", strings.ReplaceAll(slug, "-", "_"), suffix)
	password, err := crypto.GeneratePassword(24)
	if err != nil {
		return types.Project{}, err
	}
	encryptedPassword, err := s.crypto.Encrypt(password)
	if err != nil {
		return types.Project{}, err
	}

	now := time.Now().UTC()
	project := types.Project{
		ID:                  uuid.NewString(),
		OwnerUserID:         ownerUserID,
		Name:                name,
		Slug:                slug,
		Description:         strings.TrimSpace(input.Description),
		Status:              types.ProjectStatusCreating,
		PGDatabaseName:      dbName,
		PGRoleName:          roleName,
		PGPasswordEncrypted: encryptedPassword,
		PGHost:              s.pgConfig.Host,
		PGPort:              s.pgConfig.Port,
		PGSSLMode:           s.pgConfig.SSLMode,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	if err := s.store.CreateProject(ctx, project); err != nil {
		return types.Project{}, err
	}
	if err := s.postgres.CreateProjectDatabase(ctx, dbName, roleName, password); err != nil {
		_ = s.store.DeleteProject(ctx, project.ID)
		return types.Project{}, err
	}
	project.Status = types.ProjectStatusReady
	project.UpdatedAt = time.Now().UTC()
	if err := s.store.UpdateProject(ctx, project); err != nil {
		return types.Project{}, err
	}
	return project, nil
}

func (s *Service) Connection(ctx context.Context, ownerUserID, projectID string) (types.ProjectConnection, error) {
	project, err := s.store.GetProject(ctx, ownerUserID, projectID)
	if err != nil {
		return types.ProjectConnection{}, err
	}
	password, err := s.crypto.Decrypt(project.PGPasswordEncrypted)
	if err != nil {
		return types.ProjectConnection{}, err
	}
	return postgres.BuildConnection(project, password), nil
}

func (s *Service) SaveSchema(ctx context.Context, ownerUserID, projectID string, blueprint types.SchemaBlueprint) (types.SchemaRevision, map[string]string, error) {
	if _, err := s.store.GetProject(ctx, ownerUserID, projectID); err != nil {
		return types.SchemaRevision{}, nil, err
	}
	blueprint = Normalize(blueprint, projectID)
	errs := Validate(blueprint)
	if len(errs) > 0 {
		return types.SchemaRevision{}, errs, nil
	}
	hash, err := HashBlueprint(blueprint)
	if err != nil {
		return types.SchemaRevision{}, nil, err
	}
	sql, err := Generate(blueprint)
	if err != nil {
		return types.SchemaRevision{}, nil, err
	}
	revision, err := s.store.SaveSchemaRevision(ctx, projectID, blueprint, hash, sql)
	return revision, nil, err
}

func (s *Service) LatestSchema(ctx context.Context, ownerUserID, projectID string) (types.SchemaRevision, error) {
	if _, err := s.store.GetProject(ctx, ownerUserID, projectID); err != nil {
		return types.SchemaRevision{}, err
	}
	return s.store.GetLatestSchemaRevision(ctx, projectID)
}

func (s *Service) Revisions(ctx context.Context, ownerUserID, projectID string) ([]types.SchemaRevision, error) {
	if _, err := s.store.GetProject(ctx, ownerUserID, projectID); err != nil {
		return nil, err
	}
	return s.store.ListSchemaRevisions(ctx, projectID)
}

func (s *Service) ValidateBlueprint(_ context.Context, projectID string, blueprint types.SchemaBlueprint) (types.SchemaBlueprint, map[string]string) {
	blueprint = Normalize(blueprint, projectID)
	return blueprint, Validate(blueprint)
}

func (s *Service) PreviewSQL(ctx context.Context, ownerUserID, projectID string, blueprint *types.SchemaBlueprint) (string, map[string]string, error) {
	if _, err := s.store.GetProject(ctx, ownerUserID, projectID); err != nil {
		return "", nil, err
	}
	if blueprint != nil {
		normalized := Normalize(*blueprint, projectID)
		errs := Validate(normalized)
		if len(errs) > 0 {
			return "", errs, nil
		}
		sql, err := Generate(normalized)
		return sql, nil, err
	}
	revision, err := s.store.GetLatestSchemaRevision(ctx, projectID)
	if err != nil {
		return "", nil, err
	}
	return revision.GeneratedSQL, nil, nil
}

func (s *Service) Apply(ctx context.Context, ownerUserID, projectID string) (types.ApplyRun, error) {
	project, err := s.store.GetProject(ctx, ownerUserID, projectID)
	if err != nil {
		return types.ApplyRun{}, err
	}
	revision, err := s.store.GetLatestSchemaRevision(ctx, projectID)
	if err != nil {
		return types.ApplyRun{}, err
	}
	password, err := s.crypto.Decrypt(project.PGPasswordEncrypted)
	if err != nil {
		return types.ApplyRun{}, err
	}
	tableCount, err := s.postgres.PublicTableCount(ctx, project, password)
	if err != nil {
		return types.ApplyRun{}, err
	}
	if tableCount > 0 {
		return types.ApplyRun{}, errors.New("schema apply only supports empty project schemas in v1; reset the project first")
	}
	now := time.Now().UTC()
	run := types.ApplyRun{
		ID:               uuid.NewString(),
		ProjectID:        projectID,
		SchemaRevisionID: revision.ID,
		Status:           "pending",
		SQLExecuted:      revision.GeneratedSQL,
		StartedAt:        now,
	}
	if err := s.store.CreateApplyRun(ctx, run); err != nil {
		return types.ApplyRun{}, err
	}
	if err := s.postgres.ApplySQL(ctx, project, password, revision.GeneratedSQL); err != nil {
		finished := time.Now().UTC()
		run.Status = "failed"
		run.ErrorMessage = err.Error()
		run.FinishedAt = &finished
		project.Status = types.ProjectStatusApplyFailed
		project.UpdatedAt = finished
		_ = s.store.UpdateProject(ctx, project)
		_ = s.store.UpdateApplyRun(ctx, run)
		return run, err
	}
	finished := time.Now().UTC()
	run.Status = "success"
	run.FinishedAt = &finished
	project.Status = types.ProjectStatusReady
	project.LastAppliedRevisionID = &revision.ID
	project.UpdatedAt = finished
	if err := s.store.UpdateProject(ctx, project); err != nil {
		return types.ApplyRun{}, err
	}
	if err := s.store.UpdateApplyRun(ctx, run); err != nil {
		return types.ApplyRun{}, err
	}
	return run, nil
}

func (s *Service) Reset(ctx context.Context, ownerUserID, projectID string) error {
	project, err := s.store.GetProject(ctx, ownerUserID, projectID)
	if err != nil {
		return err
	}
	password, err := s.crypto.Decrypt(project.PGPasswordEncrypted)
	if err != nil {
		return err
	}
	return s.postgres.ResetProjectSchema(ctx, project, password)
}

func (s *Service) Delete(ctx context.Context, ownerUserID, projectID string) error {
	project, err := s.store.GetProject(ctx, ownerUserID, projectID)
	if err != nil {
		return err
	}
	if err := s.postgres.DropProjectDatabase(ctx, project.PGDatabaseName, project.PGRoleName); err != nil {
		return err
	}
	return s.store.DeleteProject(ctx, project.ID)
}

func (s *Service) ListTables(ctx context.Context, ownerUserID, projectID string) ([]types.TableInfo, error) {
	project, err := s.store.GetProject(ctx, ownerUserID, projectID)
	if err != nil {
		return nil, err
	}
	password, err := s.crypto.Decrypt(project.PGPasswordEncrypted)
	if err != nil {
		return nil, err
	}
	return s.postgres.ListTables(ctx, project, password)
}

func (s *Service) ListColumns(ctx context.Context, ownerUserID, projectID, tableName string) ([]types.ColumnInfo, error) {
	project, err := s.store.GetProject(ctx, ownerUserID, projectID)
	if err != nil {
		return nil, err
	}
	password, err := s.crypto.Decrypt(project.PGPasswordEncrypted)
	if err != nil {
		return nil, err
	}
	return s.postgres.ListColumns(ctx, project, password, tableName)
}

func (s *Service) ListRows(ctx context.Context, ownerUserID, projectID, tableName string, limit, offset int) (types.TableRows, error) {
	project, err := s.store.GetProject(ctx, ownerUserID, projectID)
	if err != nil {
		return types.TableRows{}, err
	}
	password, err := s.crypto.Decrypt(project.PGPasswordEncrypted)
	if err != nil {
		return types.TableRows{}, err
	}
	return s.postgres.ListRows(ctx, project, password, tableName, limit, offset)
}
