import { useQuery } from "@tanstack/react-query";
import { ColumnInfo, TableInfo, TableRows } from "../lib/types";
import { apiGet } from "./client";

export function useTables(projectId: string) {
  return useQuery({
    queryKey: ["projects", projectId, "tables"],
    queryFn: () => apiGet<{ tables: TableInfo[] }>(`/api/v1/projects/${projectId}/tables`),
    enabled: Boolean(projectId)
  });
}

export function useColumns(projectId: string, tableName: string) {
  return useQuery({
    queryKey: ["projects", projectId, "tables", tableName, "columns"],
    queryFn: () => apiGet<{ columns: ColumnInfo[] }>(`/api/v1/projects/${projectId}/tables/${tableName}/columns`),
    enabled: Boolean(projectId && tableName)
  });
}

export function useRows(projectId: string, tableName: string, limit = 50, offset = 0) {
  return useQuery({
    queryKey: ["projects", projectId, "tables", tableName, "rows", limit, offset],
    queryFn: () => apiGet<TableRows>(`/api/v1/projects/${projectId}/tables/${tableName}/rows?limit=${limit}&offset=${offset}`),
    enabled: Boolean(projectId && tableName)
  });
}
