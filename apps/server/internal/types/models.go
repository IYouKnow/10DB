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

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type ProjectConnection struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Database   string `json:"database"`
	Username   string `json:"username"`
	Password   string `json:"password,omitempty"`
	SSLMode    string `json:"sslMode"`
	DSN        string `json:"dsn"`
	EnvExample string `json:"envExample"`
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
	Version   int              `json:"version"`
	ProjectID string           `json:"projectId"`
	DatabaseID string          `json:"databaseId"`
	Tables    []TableBlueprint `json:"tables"`
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
