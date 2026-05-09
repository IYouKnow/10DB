package project

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	types "github.com/pedro/10db-launch/apps/server/internal/types"
)

var (
	ErrInvalidAPIKey      = errors.New("invalid api key")
	ErrAPIKeyPermission   = errors.New("api key does not have permission for this action")
	identifierPattern     = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	apiKeyAlphabet        = []byte("abcdefghijklmnopqrstuvwxyz0123456789")
	allowedKeyPermissions = map[string]struct{}{"read": {}, "write": {}, "read_write": {}}
)

const (
	defaultDataLimit = 100
	maxDataLimit     = 500
	apiKeyPrefixLen  = 16
)

func (s *Service) ListAPIKeys(ctx context.Context, ownerUserID, databaseID string) ([]types.DatabaseAPIKey, error) {
	if _, err := s.store.GetProjectDatabaseByOwner(ctx, ownerUserID, databaseID); err != nil {
		return nil, err
	}
	return s.store.ListDatabaseAPIKeysByOwner(ctx, ownerUserID, databaseID)
}

func (s *Service) CreateAPIKey(ctx context.Context, ownerUserID, databaseID, name, permission string) (types.DatabaseAPIKeySecret, error) {
	if _, err := s.store.GetProjectDatabaseByOwner(ctx, ownerUserID, databaseID); err != nil {
		return types.DatabaseAPIKeySecret{}, err
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return types.DatabaseAPIKeySecret{}, errors.New("api key name is required")
	}
	permission = normalizeAPIKeyPermission(permission)
	if err := validateAPIKeyPermission(permission); err != nil {
		return types.DatabaseAPIKeySecret{}, err
	}

	key, err := generateDatabaseAPIKey()
	if err != nil {
		return types.DatabaseAPIKeySecret{}, err
	}

	now := time.Now().UTC()
	record := types.DatabaseAPIKey{
		ID:         uuid.NewString(),
		DatabaseID: databaseID,
		Name:       name,
		KeyHash:    hashAPIKey(key),
		KeyPrefix:  keyPrefix(key),
		Permission: permission,
		CreatedAt:  now,
	}
	if err := s.store.CreateDatabaseAPIKey(ctx, record); err != nil {
		return types.DatabaseAPIKeySecret{}, err
	}

	return types.DatabaseAPIKeySecret{
		ID:         record.ID,
		Name:       record.Name,
		Key:        key,
		KeyPrefix:  record.KeyPrefix,
		Permission: record.Permission,
		CreatedAt:  record.CreatedAt,
	}, nil
}

func (s *Service) RevokeAPIKey(ctx context.Context, ownerUserID, databaseID, keyID string) error {
	return s.store.RevokeDatabaseAPIKeyByOwner(ctx, ownerUserID, databaseID, keyID, time.Now().UTC())
}

func (s *Service) ListDataByAPIKey(ctx context.Context, token, method, table string, limit int) (types.TableRows, error) {
	projectConn, password, err := s.projectConnectionForAPIKey(ctx, token, method)
	if err != nil {
		return types.TableRows{}, err
	}
	if err := validateIdentifier(table); err != nil {
		return types.TableRows{}, err
	}
	return s.postgres.ListDataRows(ctx, projectConn, password, table, normalizeDataLimit(limit))
}

func (s *Service) GetDataByAPIKey(ctx context.Context, token, method, table, id string) (map[string]any, error) {
	projectConn, password, err := s.projectConnectionForAPIKey(ctx, token, method)
	if err != nil {
		return nil, err
	}
	if err := validateIdentifier(table); err != nil {
		return nil, err
	}
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("id is required")
	}
	return s.postgres.GetDataRow(ctx, projectConn, password, table, id)
}

func (s *Service) InsertDataByAPIKey(ctx context.Context, token, method, table string, values map[string]any) (map[string]any, error) {
	projectConn, password, err := s.projectConnectionForAPIKey(ctx, token, method)
	if err != nil {
		return nil, err
	}
	if err := validateIdentifier(table); err != nil {
		return nil, err
	}
	if err := validateColumnNames(values); err != nil {
		return nil, err
	}
	return s.postgres.InsertDataRow(ctx, projectConn, password, table, values)
}

func (s *Service) UpdateDataByAPIKey(ctx context.Context, token, method, table, id string, values map[string]any) (map[string]any, error) {
	projectConn, password, err := s.projectConnectionForAPIKey(ctx, token, method)
	if err != nil {
		return nil, err
	}
	if err := validateIdentifier(table); err != nil {
		return nil, err
	}
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("id is required")
	}
	if err := validateColumnNames(values); err != nil {
		return nil, err
	}
	return s.postgres.UpdateDataRow(ctx, projectConn, password, table, id, values)
}

func (s *Service) DeleteDataByAPIKey(ctx context.Context, token, method, table, id string) error {
	projectConn, password, err := s.projectConnectionForAPIKey(ctx, token, method)
	if err != nil {
		return err
	}
	if err := validateIdentifier(table); err != nil {
		return err
	}
	if strings.TrimSpace(id) == "" {
		return errors.New("id is required")
	}
	return s.postgres.DeleteDataRow(ctx, projectConn, password, table, id)
}

func (s *Service) projectConnectionForAPIKey(ctx context.Context, token, method string) (types.Project, string, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return types.Project{}, "", ErrInvalidAPIKey
	}

	record, err := s.store.GetActiveDatabaseAPIKeyByHash(ctx, hashAPIKey(token))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return types.Project{}, "", ErrInvalidAPIKey
		}
		return types.Project{}, "", err
	}
	if !apiKeyAllowsMethod(record.Permission, method) {
		return types.Project{}, "", ErrAPIKeyPermission
	}

	database, err := s.store.GetProjectDatabaseByID(ctx, record.DatabaseID)
	if err != nil {
		return types.Project{}, "", err
	}
	credential, err := s.store.GetActiveMainDatabaseCredential(ctx, database.ID)
	if err != nil {
		return types.Project{}, "", err
	}
	password, err := s.crypto.Decrypt(credential.Password)
	if err != nil {
		return types.Project{}, "", err
	}

	return types.Project{
		ID:             database.ProjectID,
		PGDatabaseName: database.PGDatabaseName,
		PGRoleName:     credential.Username,
		PGHost:         database.PGHost,
		PGPort:         database.PGPort,
		PGSSLMode:      database.PGSSLMode,
	}, password, nil
}

func normalizeAPIKeyPermission(permission string) string {
	permission = strings.TrimSpace(permission)
	if permission == "" {
		return "read_write"
	}
	return permission
}

func validateAPIKeyPermission(permission string) error {
	if _, ok := allowedKeyPermissions[permission]; !ok {
		return fmt.Errorf("invalid permission: %s", permission)
	}
	return nil
}

func apiKeyAllowsMethod(permission, method string) bool {
	switch permission {
	case "read":
		return method == "GET"
	case "write":
		return method == "POST" || method == "PATCH" || method == "DELETE"
	case "read_write":
		return method == "GET" || method == "POST" || method == "PATCH" || method == "DELETE"
	default:
		return false
	}
}

func normalizeDataLimit(limit int) int {
	if limit <= 0 {
		return defaultDataLimit
	}
	if limit > maxDataLimit {
		return maxDataLimit
	}
	return limit
}

func validateIdentifier(value string) error {
	if !identifierPattern.MatchString(value) {
		return fmt.Errorf("invalid identifier: %s", value)
	}
	return nil
}

func validateColumnNames(values map[string]any) error {
	for key := range values {
		if err := validateIdentifier(key); err != nil {
			return err
		}
	}
	return nil
}

func generateDatabaseAPIKey() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	token := make([]byte, len(buf))
	for i, value := range buf {
		token[i] = apiKeyAlphabet[int(value)%len(apiKeyAlphabet)]
	}
	return "tdb_live_" + string(token), nil
}

func hashAPIKey(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func keyPrefix(value string) string {
	if len(value) <= apiKeyPrefixLen {
		return value
	}
	return value[:apiKeyPrefixLen]
}
