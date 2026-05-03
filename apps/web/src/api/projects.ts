import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Project, ProjectConnection } from "../lib/types";
import { apiGet, apiSend } from "./client";

export function useProjects() {
  return useQuery({
    queryKey: ["projects"],
    queryFn: () => apiGet<{ projects: Project[] }>("/api/v1/projects")
  });
}

export function useProject(projectId: string) {
  return useQuery({
    queryKey: ["projects", projectId],
    queryFn: () => apiGet<Project>(`/api/v1/projects/${projectId}`),
    enabled: Boolean(projectId)
  });
}

export function useCreateProject() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (body: { name: string; slug: string; description: string }) => apiSend<Project>("/api/v1/projects", "POST", body),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["projects"] });
    }
  });
}

export function useDeleteProject() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (projectId: string) => apiSend<{ ok: boolean }>(`/api/v1/projects/${projectId}`, "DELETE"),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["projects"] });
    }
  });
}

export function useResetProject() {
  return useMutation({
    mutationFn: (projectId: string) => apiSend<{ ok: boolean }>(`/api/v1/projects/${projectId}/reset`, "POST", {})
  });
}

export function useProjectConnection(projectId: string) {
  return useQuery({
    queryKey: ["projects", projectId, "connection"],
    queryFn: () => apiGet<ProjectConnection>(`/api/v1/projects/${projectId}/connection`),
    enabled: Boolean(projectId)
  });
}
