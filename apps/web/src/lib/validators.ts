import { SchemaBlueprint } from "./types";

export function hasTables(blueprint: SchemaBlueprint) {
  return blueprint.tables.length > 0;
}
