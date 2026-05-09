package types

import "time"

type ProjectStatus string

const (
	ProjectStatusDraft       ProjectStatus = "draft"
	ProjectStatusCreating    ProjectStatus = "creating"
	ProjectStatusReady       ProjectStatus = "ready"
	ProjectStatusApplyFailed ProjectStatus = "apply_failed"
	ProjectStatusDeleting    ProjectStatus = "deleting"
)

type Project struct {
	ID                    string            `json:"id"`
	OwnerUserID           string            `json:"-"`
	Name                  string            `json:"name"`
	Slug                  string            `json:"slug"`
	Description           string            `json:"description"`
	Status                ProjectStatus     `json:"status"`
	Databases             []ProjectDatabase `json:"databases,omitempty"`
	PGDatabaseName        string            `json:"pgDatabaseName"`
	PGRoleName            string            `json:"pgRoleName"`
	PGPasswordEncrypted   string            `json:"-"`
	PGHost                string            `json:"pgHost"`
	PGPort                int               `json:"pgPort"`
	PGSSLMode             string            `json:"pgSslMode"`
	LastAppliedRevisionID *string           `json:"lastAppliedRevisionId,omitempty"`
	CreatedAt             time.Time         `json:"createdAt"`
	UpdatedAt             time.Time         `json:"updatedAt"`
}

type ProjectDatabase struct {
	ID                  string    `json:"id"`
	ProjectID           string    `json:"projectId"`
	ServerID            *string   `json:"serverId,omitempty"`
	Engine              string    `json:"engine"`
	Name                string    `json:"name"`
	Status              string    `json:"status"`
	PGDatabaseName      string    `json:"pgDatabaseName"`
	PGRoleName          string    `json:"pgRoleName"`
	PGPasswordEncrypted string    `json:"-"`
	PGHost              string    `json:"pgHost"`
	PGPort              int       `json:"pgPort"`
	PGSSLMode           string    `json:"pgSslMode"`
	PositionX           float64   `json:"positionX"`
	PositionY           float64   `json:"positionY"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

type DatabaseCredential struct {
	ID         string     `json:"id"`
	DatabaseID string     `json:"databaseId"`
	Label      string     `json:"label"`
	Username   string     `json:"username"`
	Password   string     `json:"-"`
	Type       string     `json:"type"`
	RevokedAt  *time.Time `json:"revokedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
}

type DatabaseAPIKey struct {
	ID         string     `json:"id"`
	DatabaseID string     `json:"databaseId"`
	Name       string     `json:"name"`
	KeyHash    string     `json:"-"`
	KeyPrefix  string     `json:"keyPrefix"`
	Permission string     `json:"permission"`
	RevokedAt  *time.Time `json:"revokedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
}

type DatabaseAPIKeySecret struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Key        string    `json:"key"`
	KeyPrefix  string    `json:"keyPrefix"`
	Permission string    `json:"permission"`
	CreatedAt  time.Time `json:"createdAt"`
}

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	Role         string    `json:"role"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

const (
	UserRoleAdmin = "admin"
	UserRoleUser  = "user"
)

type DatabaseServer struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Engine          string    `json:"engine"`
	Host            string    `json:"host"`
	Port            int       `json:"port"`
	AdminUsername   string    `json:"adminUsername"`
	AdminPassword   string    `json:"-"` // TODO: encrypt stored server credentials before production use.
	SSLMode         string    `json:"sslMode"`
	DefaultDatabase string    `json:"defaultDatabase"`
	IsDefault       bool      `json:"isDefault"`
	IsActive        bool      `json:"isActive"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type DatabaseServerView struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Engine          string    `json:"engine"`
	Host            string    `json:"host"`
	Port            int       `json:"port"`
	AdminUsername   string    `json:"adminUsername"`
	HasPassword     bool      `json:"hasPassword"`
	SSLMode         string    `json:"sslMode"`
	DefaultDatabase string    `json:"defaultDatabase"`
	IsDefault       bool      `json:"isDefault"`
	IsActive        bool      `json:"isActive"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type AdminOverview struct {
	TotalUsers            int `json:"total_users"`
	TotalProjects         int `json:"total_projects"`
	TotalDatabases        int `json:"total_databases"`
	TotalDatabaseServers  int `json:"total_database_servers"`
	FailedDatabases       int `json:"failed_databases"`
	ActiveDatabaseServers int `json:"active_database_servers"`
}

type SchemaRevision struct {
	ID            string          `json:"id"`
	ProjectID     string          `json:"projectId"`
	DatabaseID    string          `json:"databaseId"`
	VersionNumber int             `json:"versionNumber"`
	Blueprint     SchemaBlueprint `json:"blueprint"`
	BlueprintHash string          `json:"blueprintHash"`
	GeneratedSQL  string          `json:"generatedSql"`
	CreatedAt     time.Time       `json:"createdAt"`
}

type ApplyRun struct {
	ID               string     `json:"id"`
	ProjectID        string     `json:"projectId"`
	SchemaRevisionID string     `json:"schemaRevisionId"`
	Status           string     `json:"status"`
	SQLExecuted      string     `json:"sqlExecuted"`
	ErrorMessage     string     `json:"errorMessage,omitempty"`
	StartedAt        time.Time  `json:"startedAt"`
	FinishedAt       *time.Time `json:"finishedAt,omitempty"`
}

type SchemaBlueprint struct {
	Version    int              `json:"version"`
	ProjectID  string           `json:"projectId"`
	DatabaseID string           `json:"databaseId"`
	Tables     []TableBlueprint `json:"tables"`
}

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type TableBlueprint struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	DisplayName string                `json:"displayName,omitempty"`
	Position    Position              `json:"position"`
	Width       float64               `json:"width,omitempty"`
	Height      float64               `json:"height,omitempty"`
	Columns     []ColumnBlueprint     `json:"columns"`
	ForeignKeys []ForeignKeyBlueprint `json:"foreignKeys"`
}

type ColumnBlueprint struct {
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	Type       string        `json:"type"`
	PrimaryKey bool          `json:"primaryKey"`
	Unique     bool          `json:"unique"`
	Nullable   bool          `json:"nullable"`
	Default    *DefaultValue `json:"default"`
	Config     ColumnConfig  `json:"config"`
}

type DraftTableColumn struct {
	ID           string    `json:"id"`
	TableID      string    `json:"tableId"`
	Name         string    `json:"name"`
	Type         string    `json:"type"`
	Nullable     bool      `json:"nullable"`
	PrimaryKey   bool      `json:"primaryKey"`
	DefaultValue string    `json:"defaultValue"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type DefaultValue struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

type ColumnConfig struct {
	VarcharLength     int    `json:"varcharLength,omitempty"`
	DecimalPrecision  int    `json:"decimalPrecision,omitempty"`
	DecimalScale      int    `json:"decimalScale,omitempty"`
	GeneratedStrategy string `json:"generatedStrategy,omitempty"`
}

type ForeignKeyBlueprint struct {
	ID             string   `json:"id"`
	ColumnNames    []string `json:"columnNames"`
	RefTable       string   `json:"refTable"`
	RefColumnNames []string `json:"refColumnNames"`
	OnDelete       string   `json:"onDelete"`
	OnUpdate       string   `json:"onUpdate"`
}

type TableInfo struct {
	Name string `json:"name"`
}

type ColumnInfo struct {
	Name     string `json:"name"`
	DataType string `json:"dataType"`
	Nullable bool   `json:"nullable"`
}

type TableRows struct {
	Columns []string         `json:"columns"`
	Rows    []map[string]any `json:"rows"`
	Limit   int              `json:"limit"`
	Offset  int              `json:"offset"`
}
