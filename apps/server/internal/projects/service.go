package projects

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/pedro/10db-launch/apps/server/internal/crypto"
	"github.com/pedro/10db-launch/apps/server/internal/models"
	"github.com/pedro/10db-launch/apps/server/internal/postgres"
	"github.com/pedro/10db-launch/apps/server/internal/schema"
	"github.com/pedro/10db-launch/apps/server/internal/sqlgen"
	"github.com/pedro/10db-launch/apps/server/internal/store"
)

type Service struct {
	store    *store.Store
	postgres *postgres.Service
	crypto   *crypto.Service
	pgConfig postgres.AdminConfig
}

func New(store *store.Store, postgresService *postgres.Service, cryptoService *crypto.Service, pgConfig postgres.AdminConfig) *Service {
	return &Service{store: store, postgres: postgresService, crypto: cryptoService, pgConfig: pgConfig}
}

type CreateProjectInput struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func (s *Service) List(ctx context.Context) ([]models.Project, error) {
	return s.store.ListProjects(ctx)
}

func (s *Service) Get(ctx context.Context, id string) (models.Project, error) {
	return s.store.GetProject(ctx, id)
}

func (s *Service) Create(ctx context.Context, input CreateProjectInput) (models.Project, error) {
	name := strings.TrimSpace(input.Name)
	slug := strings.TrimSpace(strings.ToLower(input.Slug))
	if name == "" || slug == "" {
		return models.Project{}, errors.New("name and slug are required")
	}
	if !slugPattern.MatchString(slug) {
		return models.Project{}, errors.New("slug must contain lowercase letters, numbers, and hyphens only")
	}

	suffix := strings.ReplaceAll(uuid.NewString()[:8], "-", "")
	dbName := fmt.Sprintf("p_%s_%s", strings.ReplaceAll(slug, "-", "_"), suffix)
	roleName := fmt.Sprintf("u_%s_%s", strings.ReplaceAll(slug, "-", "_"), suffix)
	password, err := crypto.GeneratePassword(24)
	if err != nil {
		return models.Project{}, err
	}
	encryptedPassword, err := s.crypto.Encrypt(password)
	if err != nil {
		return models.Project{}, err
	}

	now := time.Now().UTC()
	project := models.Project{
		ID:                  uuid.NewString(),
		Name:                name,
		Slug:                slug,
		Description:         strings.TrimSpace(input.Description),
		Status:              models.ProjectStatusCreating,
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
		return models.Project{}, err
	}
	if err := s.postgres.CreateProjectDatabase(ctx, dbName, roleName, password); err != nil {
		_ = s.store.DeleteProject(ctx, project.ID)
		return models.Project{}, err
	}
	project.Status = models.ProjectStatusReady
	project.UpdatedAt = time.Now().UTC()
	if err := s.store.UpdateProject(ctx, project); err != nil {
		return models.Project{}, err
	}
	return project, nil
}

func (s *Service) Connection(ctx context.Context, projectID string) (models.ProjectConnection, error) {
	project, err := s.store.GetProject(ctx, projectID)
	if err != nil {
		return models.ProjectConnection{}, err
	}
	password, err := s.crypto.Decrypt(project.PGPasswordEncrypted)
	if err != nil {
		return models.ProjectConnection{}, err
	}
	return postgres.BuildConnection(project, password), nil
}

func (s *Service) SaveSchema(ctx context.Context, projectID string, blueprint models.SchemaBlueprint) (models.SchemaRevision, map[string]string, error) {
	blueprint = schema.Normalize(blueprint, projectID)
	errs := schema.Validate(blueprint)
	if len(errs) > 0 {
		return models.SchemaRevision{}, errs, nil
	}
	hash, err := schema.HashBlueprint(blueprint)
	if err != nil {
		return models.SchemaRevision{}, nil, err
	}
	sql, err := sqlgen.Generate(blueprint)
	if err != nil {
		return models.SchemaRevision{}, nil, err
	}
	revision, err := s.store.SaveSchemaRevision(ctx, projectID, blueprint, hash, sql)
	return revision, nil, err
}

func (s *Service) LatestSchema(ctx context.Context, projectID string) (models.SchemaRevision, error) {
	return s.store.GetLatestSchemaRevision(ctx, projectID)
}

func (s *Service) Revisions(ctx context.Context, projectID string) ([]models.SchemaRevision, error) {
	return s.store.ListSchemaRevisions(ctx, projectID)
}

func (s *Service) ValidateBlueprint(_ context.Context, projectID string, blueprint models.SchemaBlueprint) (models.SchemaBlueprint, map[string]string) {
	blueprint = schema.Normalize(blueprint, projectID)
	return blueprint, schema.Validate(blueprint)
}

func (s *Service) PreviewSQL(ctx context.Context, projectID string, blueprint *models.SchemaBlueprint) (string, map[string]string, error) {
	if blueprint != nil {
		normalized := schema.Normalize(*blueprint, projectID)
		errs := schema.Validate(normalized)
		if len(errs) > 0 {
			return "", errs, nil
		}
		sql, err := sqlgen.Generate(normalized)
		return sql, nil, err
	}
	revision, err := s.store.GetLatestSchemaRevision(ctx, projectID)
	if err != nil {
		return "", nil, err
	}
	return revision.GeneratedSQL, nil, nil
}

func (s *Service) Apply(ctx context.Context, projectID string) (models.ApplyRun, error) {
	project, err := s.store.GetProject(ctx, projectID)
	if err != nil {
		return models.ApplyRun{}, err
	}
	revision, err := s.store.GetLatestSchemaRevision(ctx, projectID)
	if err != nil {
		return models.ApplyRun{}, err
	}
	password, err := s.crypto.Decrypt(project.PGPasswordEncrypted)
	if err != nil {
		return models.ApplyRun{}, err
	}
	tableCount, err := s.postgres.PublicTableCount(ctx, project, password)
	if err != nil {
		return models.ApplyRun{}, err
	}
	if tableCount > 0 {
		return models.ApplyRun{}, errors.New("schema apply only supports empty project schemas in v1; reset the project first")
	}
	now := time.Now().UTC()
	run := models.ApplyRun{
		ID:               uuid.NewString(),
		ProjectID:        projectID,
		SchemaRevisionID: revision.ID,
		Status:           "pending",
		SQLExecuted:      revision.GeneratedSQL,
		StartedAt:        now,
	}
	if err := s.store.CreateApplyRun(ctx, run); err != nil {
		return models.ApplyRun{}, err
	}
	if err := s.postgres.ApplySQL(ctx, project, password, revision.GeneratedSQL); err != nil {
		finished := time.Now().UTC()
		run.Status = "failed"
		run.ErrorMessage = err.Error()
		run.FinishedAt = &finished
		project.Status = models.ProjectStatusApplyFailed
		project.UpdatedAt = finished
		_ = s.store.UpdateProject(ctx, project)
		_ = s.store.UpdateApplyRun(ctx, run)
		return run, err
	}
	finished := time.Now().UTC()
	run.Status = "success"
	run.FinishedAt = &finished
	project.Status = models.ProjectStatusReady
	project.LastAppliedRevisionID = &revision.ID
	project.UpdatedAt = finished
	if err := s.store.UpdateProject(ctx, project); err != nil {
		return models.ApplyRun{}, err
	}
	if err := s.store.UpdateApplyRun(ctx, run); err != nil {
		return models.ApplyRun{}, err
	}
	return run, nil
}

func (s *Service) Reset(ctx context.Context, projectID string) error {
	project, err := s.store.GetProject(ctx, projectID)
	if err != nil {
		return err
	}
	password, err := s.crypto.Decrypt(project.PGPasswordEncrypted)
	if err != nil {
		return err
	}
	return s.postgres.ResetProjectSchema(ctx, project, password)
}

func (s *Service) Delete(ctx context.Context, projectID string) error {
	project, err := s.store.GetProject(ctx, projectID)
	if err != nil {
		return err
	}
	if err := s.postgres.DropProjectDatabase(ctx, project.PGDatabaseName, project.PGRoleName); err != nil {
		return err
	}
	return s.store.DeleteProject(ctx, project.ID)
}

func (s *Service) ListTables(ctx context.Context, projectID string) ([]models.TableInfo, error) {
	project, err := s.store.GetProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	password, err := s.crypto.Decrypt(project.PGPasswordEncrypted)
	if err != nil {
		return nil, err
	}
	return s.postgres.ListTables(ctx, project, password)
}

func (s *Service) ListColumns(ctx context.Context, projectID, tableName string) ([]models.ColumnInfo, error) {
	project, err := s.store.GetProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	password, err := s.crypto.Decrypt(project.PGPasswordEncrypted)
	if err != nil {
		return nil, err
	}
	return s.postgres.ListColumns(ctx, project, password, tableName)
}

func (s *Service) ListRows(ctx context.Context, projectID, tableName string, limit, offset int) (models.TableRows, error) {
	project, err := s.store.GetProject(ctx, projectID)
	if err != nil {
		return models.TableRows{}, err
	}
	password, err := s.crypto.Decrypt(project.PGPasswordEncrypted)
	if err != nil {
		return models.TableRows{}, err
	}
	return s.postgres.ListRows(ctx, project, password, tableName, limit, offset)
}
