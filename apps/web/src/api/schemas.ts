import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { SchemaBlueprint, SchemaRevision } from "../lib/types";
import { apiGet, apiSend } from "./client";

export function useSchema(projectId: string) {
  return useQuery({
    queryKey: ["projects", projectId, "schema"],
    queryFn: async () => {
      const payload = await apiGet<SchemaRevision | SchemaBlueprint>(`/api/v1/projects/${projectId}/schema`);
      if ("blueprint" in payload) {
        return payload.blueprint;
      }
      return payload;
    },
    enabled: Boolean(projectId)
  });
}

export function useSaveSchema(projectId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (blueprint: SchemaBlueprint) => apiSend<SchemaRevision>(`/api/v1/projects/${projectId}/schema`, "PUT", blueprint),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["projects", projectId, "schema"] });
      void queryClient.invalidateQueries({ queryKey: ["projects", projectId, "schema", "revisions"] });
    }
  });
}

export function useSchemaRevisions(projectId: string) {
  return useQuery({
    queryKey: ["projects", projectId, "schema", "revisions"],
    queryFn: () => apiGet<{ revisions: SchemaRevision[] }>(`/api/v1/projects/${projectId}/schema/revisions`),
    enabled: Boolean(projectId)
  });
}

export function useValidateSchema(projectId: string) {
  return useMutation({
    mutationFn: (blueprint: SchemaBlueprint) =>
      apiSend<{ blueprint: SchemaBlueprint; errors: Record<string, string>; valid: boolean }>(
        `/api/v1/projects/${projectId}/schema/validate`,
        "POST",
        blueprint
      )
  });
}
