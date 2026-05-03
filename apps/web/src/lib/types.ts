export type ProjectStatus = "creating" | "ready" | "apply_failed" | "deleting";

export interface Project {
  id: string;
  name: string;
  slug: string;
  description: string;
  status: ProjectStatus;
  pgDatabaseName: string;
  pgRoleName: string;
  pgHost: string;
  pgPort: number;
  pgSslMode: string;
  lastAppliedRevisionId?: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface ProjectConnection {
  host: string;
  port: number;
  database: string;
  username: string;
  password?: string;
  sslMode: string;
  dsn: string;
  envExample: string;
}

export interface Position {
  x: number;
  y: number;
}

export interface DefaultValue {
  kind: string;
  value: string;
}

export interface ColumnConfig {
  varcharLength?: number;
  decimalPrecision?: number;
  decimalScale?: number;
  generatedStrategy?: string;
}

export interface ColumnBlueprint {
  id: string;
  name: string;
  type: string;
  primaryKey: boolean;
  unique: boolean;
  nullable: boolean;
  default: DefaultValue | null;
  config: ColumnConfig;
}

export interface ForeignKeyBlueprint {
  id: string;
  columnNames: string[];
  refTable: string;
  refColumnNames: string[];
  onDelete: string;
  onUpdate: string;
}

export interface TableBlueprint {
  id: string;
  name: string;
  displayName?: string;
  position: Position;
  columns: ColumnBlueprint[];
  foreignKeys: ForeignKeyBlueprint[];
}

export interface SchemaBlueprint {
  version: 1;
  projectId: string;
  tables: TableBlueprint[];
}

export interface SchemaRevision {
  id: string;
  projectId: string;
  versionNumber: number;
  blueprint: SchemaBlueprint;
  blueprintHash: string;
  generatedSql: string;
  createdAt: string;
}

export interface TableInfo {
  name: string;
}

export interface ColumnInfo {
  name: string;
  dataType: string;
  nullable: boolean;
}

export interface TableRows {
  columns: string[];
  rows: Array<Record<string, unknown>>;
  limit: number;
  offset: number;
}
