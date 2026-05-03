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

type ProvisionPostgresInput struct {
	Name string `json:"name"`
}

type UpdateProjectDatabaseInput struct {
	Name string `json:"name"`
}

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

const (
	defaultDatabasePositionX = 80.0
	defaultDatabasePositionY = 80.0
	databaseColumnGap        = 360.0
	databaseRowGap           = 220.0
	databasesPerRow          = 3
)

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

	now := time.Now().UTC()
	project := types.Project{
		ID:                  uuid.NewString(),
		OwnerUserID:         ownerUserID,
		Name:                name,
		Slug:                slug,
		Description:         strings.TrimSpace(input.Description),
		Status:              types.ProjectStatusDraft,
		PGDatabaseName:      "",
		PGRoleName:          "",
		PGPasswordEncrypted: "",
		PGHost:              "",
		PGPort:              0,
		PGSSLMode:           "",
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	if err := s.store.CreateProject(ctx, project); err != nil {
		return types.Project{}, err
	}
	return project, nil
}

func (s *Service) ProvisionPostgres(ctx context.Context, ownerUserID, projectID string, input ProvisionPostgresInput) (types.Project, error) {
	project, err := s.store.GetProject(ctx, ownerUserID, projectID)
	if err != nil {
		return types.Project{}, err
	}

	suffix := strings.ReplaceAll(uuid.NewString()[:8], "-", "")
	dbName := fmt.Sprintf("p_%s_%s", strings.ReplaceAll(project.Slug, "-", "_"), suffix)
	roleName := fmt.Sprintf("u_%s_%s", strings.ReplaceAll(project.Slug, "-", "_"), suffix)
	password, err := crypto.GeneratePassword(24)
	if err != nil {
		return types.Project{}, err
	}
	encryptedPassword, err := s.crypto.Encrypt(password)
	if err != nil {
		return types.Project{}, err
	}

	index := len(project.Databases)
	column := index % databasesPerRow
	row := index / databasesPerRow
	displayName := strings.TrimSpace(input.Name)
	if displayName == "" {
		displayName = fmt.Sprintf("PostgreSQL Database %d", index+1)
	}
	now := time.Now().UTC()
	database := types.ProjectDatabase{
		ID:                  uuid.NewString(),
		ProjectID:           project.ID,
		Engine:              "postgresql",
		Name:                displayName,
		Status:              string(types.ProjectStatusCreating),
		PGDatabaseName:      dbName,
		PGRoleName:          roleName,
		PGPasswordEncrypted: encryptedPassword,
		PGHost:              s.pgConfig.Host,
		PGPort:              s.pgConfig.Port,
		PGSSLMode:           s.pgConfig.SSLMode,
		PositionX:           defaultDatabasePositionX + float64(column)*databaseColumnGap,
		PositionY:           defaultDatabasePositionY + float64(row)*databaseRowGap,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	if err := s.store.CreateProjectDatabase(ctx, database); err != nil {
		return types.Project{}, err
	}

	project.Status = types.ProjectStatusCreating
	project.UpdatedAt = now
	if err := s.store.UpdateProject(ctx, project); err != nil {
		return types.Project{}, err
	}

	if err := s.postgres.CreateProjectDatabase(ctx, dbName, roleName, password); err != nil {
		_ = s.store.DeleteProjectDatabase(ctx, project.ID, database.ID)
		project.Status = types.ProjectStatusDraft
		project.UpdatedAt = time.Now().UTC()
		_ = s.store.UpdateProject(ctx, project)
		return types.Project{}, err
	}

	database.Status = string(types.ProjectStatusReady)
	database.UpdatedAt = time.Now().UTC()
	if err := s.store.UpdateProjectDatabase(ctx, database); err != nil {
		return types.Project{}, err
	}
	project.Status = types.ProjectStatusReady
	project.UpdatedAt = time.Now().UTC()
	if err := s.store.UpdateProject(ctx, project); err != nil {
		return types.Project{}, err
	}
	return s.store.GetProject(ctx, ownerUserID, projectID)
}

func (s *Service) UpdateProjectDatabase(ctx context.Context, ownerUserID, projectID, databaseID string, input UpdateProjectDatabaseInput) (types.Project, error) {
	project, err := s.store.GetProject(ctx, ownerUserID, projectID)
	if err != nil {
		return types.Project{}, err
	}

	database, err := s.store.GetProjectDatabase(ctx, project.ID, databaseID)
	if err != nil {
		return types.Project{}, err
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return types.Project{}, errors.New("database name is required")
	}

	database.Name = name
	database.UpdatedAt = time.Now().UTC()
	if err := s.store.UpdateProjectDatabase(ctx, database); err != nil {
		return types.Project{}, err
	}

	return s.store.GetProject(ctx, ownerUserID, projectID)
}

func (s *Service) RemoveProvisionedPostgres(ctx context.Context, ownerUserID, projectID, databaseID string) (types.Project, error) {
	project, err := s.store.GetProject(ctx, ownerUserID, projectID)
	if err != nil {
		return types.Project{}, err
	}
	database, err := s.store.GetProjectDatabase(ctx, project.ID, databaseID)
	if err != nil {
		return types.Project{}, err
	}

	if err := s.postgres.DropProjectDatabase(ctx, database.PGDatabaseName, database.PGRoleName); err != nil {
		return types.Project{}, err
	}
	if err := s.store.DeleteProjectDatabase(ctx, project.ID, database.ID); err != nil {
		return types.Project{}, err
	}

	remainingDatabases, err := s.store.ListProjectDatabases(ctx, project.ID)
	if err != nil {
		return types.Project{}, err
	}
	if len(remainingDatabases) == 0 {
		project.Status = types.ProjectStatusDraft
	} else {
		project.Status = types.ProjectStatusReady
	}
	project.UpdatedAt = time.Now().UTC()
	if err := s.store.UpdateProject(ctx, project); err != nil {
		return types.Project{}, err
	}
	return s.store.GetProject(ctx, ownerUserID, projectID)
}

func (s *Service) Connection(ctx context.Context, ownerUserID, projectID string) (types.ProjectConnection, error) {
	project, err := s.store.GetProject(ctx, ownerUserID, projectID)
	if err != nil {
		return types.ProjectConnection{}, err
	}
	if project.PGDatabaseName == "" {
		return types.ProjectConnection{}, errors.New("project does not have a provisioned database yet")
	}
	password, err := s.crypto.Decrypt(project.PGPasswordEncrypted)
	if err != nil {
		return types.ProjectConnection{}, err
	}
	return postgres.BuildConnection(project, password), nil
}

func (s *Service) SaveSchema(ctx context.Context, ownerUserID, projectID, databaseID string, blueprint types.SchemaBlueprint) (types.SchemaRevision, map[string]string, error) {
	if _, err := s.store.GetProject(ctx, ownerUserID, projectID); err != nil {
		return types.SchemaRevision{}, nil, err
	}
	if _, err := s.store.GetProjectDatabase(ctx, projectID, databaseID); err != nil {
		return types.SchemaRevision{}, nil, err
	}
	blueprint = Normalize(blueprint, projectID, databaseID)
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

func (s *Service) LatestSchema(ctx context.Context, ownerUserID, projectID, databaseID string) (types.SchemaRevision, error) {
	if _, err := s.store.GetProject(ctx, ownerUserID, projectID); err != nil {
		return types.SchemaRevision{}, err
	}
	if _, err := s.store.GetProjectDatabase(ctx, projectID, databaseID); err != nil {
		return types.SchemaRevision{}, err
	}
	return s.store.GetLatestSchemaRevision(ctx, projectID, databaseID)
}

func (s *Service) Revisions(ctx context.Context, ownerUserID, projectID, databaseID string) ([]types.SchemaRevision, error) {
	if _, err := s.store.GetProject(ctx, ownerUserID, projectID); err != nil {
		return nil, err
	}
	if _, err := s.store.GetProjectDatabase(ctx, projectID, databaseID); err != nil {
		return nil, err
	}
	return s.store.ListSchemaRevisions(ctx, projectID, databaseID)
}

func (s *Service) ValidateBlueprint(_ context.Context, projectID, databaseID string, blueprint types.SchemaBlueprint) (types.SchemaBlueprint, map[string]string) {
	blueprint = Normalize(blueprint, projectID, databaseID)
	return blueprint, Validate(blueprint)
}

func (s *Service) PreviewSQL(ctx context.Context, ownerUserID, projectID, databaseID string, blueprint *types.SchemaBlueprint) (string, map[string]string, error) {
	if _, err := s.store.GetProject(ctx, ownerUserID, projectID); err != nil {
		return "", nil, err
	}
	if _, err := s.store.GetProjectDatabase(ctx, projectID, databaseID); err != nil {
		return "", nil, err
	}
	if blueprint != nil {
		normalized := Normalize(*blueprint, projectID, databaseID)
		errs := Validate(normalized)
		if len(errs) > 0 {
			return "", errs, nil
		}
		sql, err := Generate(normalized)
		return sql, nil, err
	}
	revision, err := s.store.GetLatestSchemaRevision(ctx, projectID, databaseID)
	if err != nil {
		return "", nil, err
	}
	return revision.GeneratedSQL, nil, nil
}

func (s *Service) Apply(ctx context.Context, ownerUserID, projectID, databaseID string) (types.ApplyRun, error) {
	database, projectConn, password, err := s.loadDatabaseConnection(ctx, ownerUserID, projectID, databaseID)
	if err != nil {
		return types.ApplyRun{}, err
	}
	revision, err := s.store.GetLatestSchemaRevision(ctx, projectID, databaseID)
	if err != nil {
		return types.ApplyRun{}, err
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
	if err := s.postgres.ApplySQL(ctx, projectConn, password, revision.GeneratedSQL); err != nil {
		finished := time.Now().UTC()
		run.Status = "failed"
		run.ErrorMessage = err.Error()
		run.FinishedAt = &finished
		database.Status = string(types.ProjectStatusApplyFailed)
		database.UpdatedAt = finished
		_ = s.store.UpdateProjectDatabase(ctx, database)
		_ = s.store.UpdateApplyRun(ctx, run)
		return run, err
	}
	finished := time.Now().UTC()
	run.Status = "success"
	run.FinishedAt = &finished
	database.Status = string(types.ProjectStatusReady)
	database.UpdatedAt = finished
	if err := s.store.UpdateProjectDatabase(ctx, database); err != nil {
		return types.ApplyRun{}, err
	}
	if err := s.store.UpdateApplyRun(ctx, run); err != nil {
		return types.ApplyRun{}, err
	}
	return run, nil
}

func (s *Service) ApplyTable(ctx context.Context, ownerUserID, projectID, databaseID, tableID string) error {
	database, projectConn, password, err := s.loadDatabaseConnection(ctx, ownerUserID, projectID, databaseID)
	if err != nil {
		return err
	}
	revision, err := s.store.GetLatestSchemaRevision(ctx, projectID, databaseID)
	if err != nil {
		return err
	}

	var target *types.TableBlueprint
	for index := range revision.Blueprint.Tables {
		if revision.Blueprint.Tables[index].ID == tableID {
			target = &revision.Blueprint.Tables[index]
			break
		}
	}
	if target == nil {
		return errors.New("table draft not found")
	}

	sql, err := Generate(types.SchemaBlueprint{
		Version:    1,
		ProjectID:  projectID,
		DatabaseID: databaseID,
		Tables:     []types.TableBlueprint{*target},
	})
	if err != nil {
		return err
	}

	database.Status = string(types.ProjectStatusReady)
	database.UpdatedAt = time.Now().UTC()
	if err := s.postgres.ApplySQL(ctx, projectConn, password, sql); err != nil {
		return err
	}
	return s.store.UpdateProjectDatabase(ctx, database)
}

func (s *Service) DeleteTable(ctx context.Context, ownerUserID, projectID, databaseID, tableID string) error {
	_, projectConn, password, err := s.loadDatabaseConnection(ctx, ownerUserID, projectID, databaseID)
	if err != nil {
		return err
	}
	revision, err := s.store.GetLatestSchemaRevision(ctx, projectID, databaseID)
	if err != nil {
		return err
	}

	var tableName string
	for _, table := range revision.Blueprint.Tables {
		if table.ID == tableID {
			tableName = table.Name
			break
		}
	}
	if tableName == "" {
		return errors.New("table draft not found")
	}

	return s.postgres.DropTable(ctx, projectConn, password, tableName)
}

func (s *Service) ListDatabaseTables(ctx context.Context, ownerUserID, projectID, databaseID string) ([]types.TableInfo, error) {
	_, projectConn, password, err := s.loadDatabaseConnection(ctx, ownerUserID, projectID, databaseID)
	if err != nil {
		return nil, err
	}
	return s.postgres.ListTables(ctx, projectConn, password)
}

func (s *Service) loadDatabaseConnection(ctx context.Context, ownerUserID, projectID, databaseID string) (types.ProjectDatabase, types.Project, string, error) {
	if _, err := s.store.GetProject(ctx, ownerUserID, projectID); err != nil {
		return types.ProjectDatabase{}, types.Project{}, "", err
	}
	database, err := s.store.GetProjectDatabase(ctx, projectID, databaseID)
	if err != nil {
		return types.ProjectDatabase{}, types.Project{}, "", err
	}
	password, err := s.crypto.Decrypt(database.PGPasswordEncrypted)
	if err != nil {
		return types.ProjectDatabase{}, types.Project{}, "", err
	}
	projectConn := types.Project{
		ID:             projectID,
		PGDatabaseName: database.PGDatabaseName,
		PGRoleName:     database.PGRoleName,
		PGHost:         database.PGHost,
		PGPort:         database.PGPort,
		PGSSLMode:      database.PGSSLMode,
	}
	return database, projectConn, password, nil
}

func (s *Service) Reset(ctx context.Context, ownerUserID, projectID string) error {
	project, err := s.store.GetProject(ctx, ownerUserID, projectID)
	if err != nil {
		return err
	}
	if project.PGDatabaseName == "" {
		return errors.New("project does not have a provisioned database yet")
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
	if project.PGDatabaseName != "" && project.PGRoleName != "" {
		if err := s.postgres.DropProjectDatabase(ctx, project.PGDatabaseName, project.PGRoleName); err != nil {
			return err
		}
	}
	return s.store.DeleteProject(ctx, project.ID)
}

func (s *Service) ListTables(ctx context.Context, ownerUserID, projectID string) ([]types.TableInfo, error) {
	project, err := s.store.GetProject(ctx, ownerUserID, projectID)
	if err != nil {
		return nil, err
	}
	if project.PGDatabaseName == "" {
		return nil, errors.New("project does not have a provisioned database yet")
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
	if project.PGDatabaseName == "" {
		return nil, errors.New("project does not have a provisioned database yet")
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
	if project.PGDatabaseName == "" {
		return types.TableRows{}, errors.New("project does not have a provisioned database yet")
	}
	password, err := s.crypto.Decrypt(project.PGPasswordEncrypted)
	if err != nil {
		return types.TableRows{}, err
	}
	return s.postgres.ListRows(ctx, project, password, tableName, limit, offset)
}
