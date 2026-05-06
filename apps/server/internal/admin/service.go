package admin

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/pedro/10db-launch/apps/server/internal/platform/postgres"
	types "github.com/pedro/10db-launch/apps/server/internal/types"
)

type userCounter interface {
	Count(ctx context.Context) (int, error)
}

type projectStore interface {
	CountProjects(ctx context.Context) (int, error)
	CountProjectDatabasesTotal(ctx context.Context) (int, error)
	CountFailedProjectDatabases(ctx context.Context) (int, error)
	CountDatabaseServers(ctx context.Context) (int, error)
	CountActiveDatabaseServers(ctx context.Context) (int, error)
	ListDatabaseServers(ctx context.Context) ([]types.DatabaseServer, error)
	GetDatabaseServer(ctx context.Context, id string) (types.DatabaseServer, error)
	CreateDatabaseServer(ctx context.Context, server types.DatabaseServer) error
	UpdateDatabaseServer(ctx context.Context, server types.DatabaseServer) error
	SetDefaultDatabaseServer(ctx context.Context, id string) error
	DeleteDatabaseServer(ctx context.Context, id string) error
	CountDatabasesByServerID(ctx context.Context, serverID string) (int, error)
}

type postgresTester interface {
	TestAdminConnection(ctx context.Context, cfg postgres.AdminConfig) error
}

type Service struct {
	users    userCounter
	projects projectStore
	tester   postgresTester
}

func New(userCounter userCounter, projectStore projectStore, tester postgresTester) *Service {
	return &Service{users: userCounter, projects: projectStore, tester: tester}
}

type CreateServerInput struct {
	Name            string `json:"name"`
	Host            string `json:"host"`
	Port            int    `json:"port"`
	AdminUsername   string `json:"admin_username"`
	AdminPassword   string `json:"admin_password"`
	SSLMode         string `json:"ssl_mode"`
	DefaultDatabase string `json:"default_database"`
	IsDefault       bool   `json:"is_default"`
	IsActive        *bool  `json:"is_active"`
}

type UpdateServerInput struct {
	Name            *string `json:"name"`
	Host            *string `json:"host"`
	Port            *int    `json:"port"`
	AdminUsername   *string `json:"admin_username"`
	AdminPassword   *string `json:"admin_password"`
	SSLMode         *string `json:"ssl_mode"`
	DefaultDatabase *string `json:"default_database"`
	IsDefault       *bool   `json:"is_default"`
	IsActive        *bool   `json:"is_active"`
}

type ServerTestResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (s *Service) Overview(ctx context.Context) (types.AdminOverview, error) {
	totalUsers, err := s.users.Count(ctx)
	if err != nil {
		return types.AdminOverview{}, err
	}
	totalProjects, err := s.projects.CountProjects(ctx)
	if err != nil {
		return types.AdminOverview{}, err
	}
	totalDatabases, err := s.projects.CountProjectDatabasesTotal(ctx)
	if err != nil {
		return types.AdminOverview{}, err
	}
	totalDatabaseServers, err := s.projects.CountDatabaseServers(ctx)
	if err != nil {
		return types.AdminOverview{}, err
	}
	failedDatabases, err := s.projects.CountFailedProjectDatabases(ctx)
	if err != nil {
		return types.AdminOverview{}, err
	}
	activeDatabaseServers, err := s.projects.CountActiveDatabaseServers(ctx)
	if err != nil {
		return types.AdminOverview{}, err
	}

	return types.AdminOverview{
		TotalUsers:            totalUsers,
		TotalProjects:         totalProjects,
		TotalDatabases:        totalDatabases,
		TotalDatabaseServers:  totalDatabaseServers,
		FailedDatabases:       failedDatabases,
		ActiveDatabaseServers: activeDatabaseServers,
	}, nil
}

func (s *Service) ListServers(ctx context.Context) ([]types.DatabaseServerView, error) {
	servers, err := s.projects.ListDatabaseServers(ctx)
	if err != nil {
		return nil, err
	}
	views := make([]types.DatabaseServerView, 0, len(servers))
	for _, server := range servers {
		views = append(views, toDatabaseServerView(server))
	}
	return views, nil
}

func (s *Service) CreateServer(ctx context.Context, input CreateServerInput) (types.DatabaseServerView, error) {
	normalized, err := normalizeCreateInput(input)
	if err != nil {
		return types.DatabaseServerView{}, err
	}

	count, err := s.projects.CountDatabaseServers(ctx)
	if err != nil {
		return types.DatabaseServerView{}, err
	}

	now := time.Now().UTC()
	server := types.DatabaseServer{
		ID:              uuid.NewString(),
		Name:            normalized.Name,
		Engine:          "postgres",
		Host:            normalized.Host,
		Port:            normalized.Port,
		AdminUsername:   normalized.AdminUsername,
		AdminPassword:   normalized.AdminPassword,
		SSLMode:         normalized.SSLMode,
		DefaultDatabase: normalized.DefaultDatabase,
		IsDefault:       count == 0 || normalized.IsDefault,
		IsActive:        normalized.IsActive,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if !server.IsActive && server.IsDefault {
		return types.DatabaseServerView{}, errors.New("Default server must be active.")
	}

	if err := s.projects.CreateDatabaseServer(ctx, server); err != nil {
		return types.DatabaseServerView{}, err
	}
	if server.IsDefault {
		if err := s.projects.SetDefaultDatabaseServer(ctx, server.ID); err != nil {
			return types.DatabaseServerView{}, err
		}
		server.UpdatedAt = time.Now().UTC()
	}
	return toDatabaseServerView(server), nil
}

func (s *Service) UpdateServer(ctx context.Context, id string, input UpdateServerInput) (types.DatabaseServerView, error) {
	server, err := s.projects.GetDatabaseServer(ctx, id)
	if err != nil {
		return types.DatabaseServerView{}, err
	}

	if input.Name != nil {
		server.Name = strings.TrimSpace(*input.Name)
	}
	if input.Host != nil {
		server.Host = strings.TrimSpace(*input.Host)
	}
	if input.Port != nil {
		server.Port = *input.Port
	}
	if input.AdminUsername != nil {
		server.AdminUsername = strings.TrimSpace(*input.AdminUsername)
	}
	if input.AdminPassword != nil && strings.TrimSpace(*input.AdminPassword) != "" {
		server.AdminPassword = *input.AdminPassword
	}
	if input.SSLMode != nil {
		server.SSLMode = strings.TrimSpace(*input.SSLMode)
	}
	if input.DefaultDatabase != nil {
		server.DefaultDatabase = strings.TrimSpace(*input.DefaultDatabase)
	}
	if input.IsActive != nil {
		if server.IsDefault && !*input.IsActive {
			return types.DatabaseServerView{}, errors.New("Default server must remain active until another default is set.")
		}
		server.IsActive = *input.IsActive
	}

	if err := validateServer(server, true); err != nil {
		return types.DatabaseServerView{}, err
	}

	server.UpdatedAt = time.Now().UTC()
	if err := s.projects.UpdateDatabaseServer(ctx, server); err != nil {
		return types.DatabaseServerView{}, err
	}

	if input.IsDefault != nil && *input.IsDefault {
		if !server.IsActive {
			return types.DatabaseServerView{}, errors.New("Only active servers can be default.")
		}
		if err := s.projects.SetDefaultDatabaseServer(ctx, server.ID); err != nil {
			return types.DatabaseServerView{}, err
		}
		server.IsDefault = true
		server.UpdatedAt = time.Now().UTC()
	}

	return toDatabaseServerView(server), nil
}

func (s *Service) TestServer(ctx context.Context, id string) (ServerTestResult, error) {
	server, err := s.projects.GetDatabaseServer(ctx, id)
	if err != nil {
		return ServerTestResult{}, err
	}

	err = s.tester.TestAdminConnection(ctx, postgres.AdminConfig{
		Host:     server.Host,
		Port:     server.Port,
		DBName:   server.DefaultDatabase,
		User:     server.AdminUsername,
		Password: server.AdminPassword,
		SSLMode:  server.SSLMode,
	})
	if err != nil {
		return ServerTestResult{
			Success: false,
			Message: "Connection failed: " + err.Error(),
		}, nil
	}
	return ServerTestResult{
		Success: true,
		Message: "Connection successful.",
	}, nil
}

func (s *Service) SetDefaultServer(ctx context.Context, id string) (types.DatabaseServerView, error) {
	server, err := s.projects.GetDatabaseServer(ctx, id)
	if err != nil {
		return types.DatabaseServerView{}, err
	}
	if !server.IsActive {
		return types.DatabaseServerView{}, errors.New("Only active servers can be default.")
	}
	if err := s.projects.SetDefaultDatabaseServer(ctx, id); err != nil {
		return types.DatabaseServerView{}, err
	}
	server.IsDefault = true
	server.UpdatedAt = time.Now().UTC()
	return toDatabaseServerView(server), nil
}

func (s *Service) DeleteServer(ctx context.Context, id string) error {
	server, err := s.projects.GetDatabaseServer(ctx, id)
	if err != nil {
		return err
	}
	if server.IsDefault {
		return errors.New("Cannot delete the default server.")
	}
	count, err := s.projects.CountDatabasesByServerID(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("Cannot delete server while databases are using it.")
	}
	return s.projects.DeleteDatabaseServer(ctx, id)
}

type normalizedCreateInput struct {
	Name            string
	Host            string
	Port            int
	AdminUsername   string
	AdminPassword   string
	SSLMode         string
	DefaultDatabase string
	IsDefault       bool
	IsActive        bool
}

func normalizeCreateInput(input CreateServerInput) (normalizedCreateInput, error) {
	normalized := normalizedCreateInput{
		Name:            strings.TrimSpace(input.Name),
		Host:            strings.TrimSpace(input.Host),
		Port:            input.Port,
		AdminUsername:   strings.TrimSpace(input.AdminUsername),
		AdminPassword:   input.AdminPassword,
		SSLMode:         strings.TrimSpace(input.SSLMode),
		DefaultDatabase: strings.TrimSpace(input.DefaultDatabase),
		IsDefault:       input.IsDefault,
		IsActive:        true,
	}
	if input.IsActive != nil {
		normalized.IsActive = *input.IsActive
	}
	if normalized.Port == 0 {
		normalized.Port = 5432
	}
	if normalized.SSLMode == "" {
		normalized.SSLMode = "disable"
	}
	if normalized.DefaultDatabase == "" {
		normalized.DefaultDatabase = "postgres"
	}
	server := types.DatabaseServer{
		Name:            normalized.Name,
		Host:            normalized.Host,
		Port:            normalized.Port,
		AdminUsername:   normalized.AdminUsername,
		AdminPassword:   normalized.AdminPassword,
		SSLMode:         normalized.SSLMode,
		DefaultDatabase: normalized.DefaultDatabase,
		IsActive:        normalized.IsActive,
	}
	if err := validateServer(server, false); err != nil {
		return normalizedCreateInput{}, err
	}
	return normalized, nil
}

func validateServer(server types.DatabaseServer, allowEmptyPassword bool) error {
	if strings.TrimSpace(server.Name) == "" {
		return errors.New("name is required")
	}
	if strings.TrimSpace(server.Host) == "" {
		return errors.New("host is required")
	}
	if server.Port <= 0 {
		return errors.New("port must be greater than 0")
	}
	if strings.TrimSpace(server.AdminUsername) == "" {
		return errors.New("admin_username is required")
	}
	if !allowEmptyPassword && strings.TrimSpace(server.AdminPassword) == "" {
		return errors.New("admin_password is required")
	}
	if strings.TrimSpace(server.SSLMode) == "" {
		return errors.New("ssl_mode is required")
	}
	if strings.TrimSpace(server.DefaultDatabase) == "" {
		return errors.New("default_database is required")
	}
	return nil
}

func toDatabaseServerView(server types.DatabaseServer) types.DatabaseServerView {
	return types.DatabaseServerView{
		ID:              server.ID,
		Name:            server.Name,
		Engine:          server.Engine,
		Host:            server.Host,
		Port:            server.Port,
		AdminUsername:   server.AdminUsername,
		HasPassword:     strings.TrimSpace(server.AdminPassword) != "",
		SSLMode:         server.SSLMode,
		DefaultDatabase: server.DefaultDatabase,
		IsDefault:       server.IsDefault,
		IsActive:        server.IsActive,
		CreatedAt:       server.CreatedAt,
		UpdatedAt:       server.UpdatedAt,
	}
}

type PostgresTester struct{}

func (PostgresTester) TestAdminConnection(ctx context.Context, cfg postgres.AdminConfig) error {
	return postgres.TestAdminConnection(ctx, cfg)
}

func IsNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
