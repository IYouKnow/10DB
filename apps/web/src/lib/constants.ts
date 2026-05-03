import { SchemaBlueprint } from "./types";

export const COLUMN_TYPES = [
  "id",
  "uuid",
  "text",
  "varchar",
  "integer",
  "decimal",
  "boolean",
  "timestamp",
  "date",
  "jsonb"
] as const;

export const DEFAULT_BLUEPRINT = (projectId: string): SchemaBlueprint => ({
  version: 1,
  projectId,
  tables: []
});
