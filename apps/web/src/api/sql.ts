import { useMutation } from "@tanstack/react-query";
import { SchemaBlueprint } from "../lib/types";
import { apiSend } from "./client";

export function useSqlPreview(projectId: string) {
  return useMutation({
    mutationFn: (blueprint: SchemaBlueprint) => apiSend<{ sql: string }>(`/api/v1/projects/${projectId}/schema/sql-preview`, "POST", blueprint)
  });
}

export function useApplySchema(projectId: string) {
  return useMutation({
    mutationFn: () => apiSend(`/api/v1/projects/${projectId}/schema/apply`, "POST", {})
  });
}
