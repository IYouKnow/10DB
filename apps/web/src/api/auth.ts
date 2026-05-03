import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { apiGet, apiSend } from "./client";

export function useMe() {
  return useQuery({
    queryKey: ["me"],
    queryFn: () => apiGet<{ username: string }>("/api/v1/auth/me")
  });
}

export function useLogin() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (body: { username: string; password: string }) => apiSend<{ ok: boolean }>("/api/v1/auth/login", "POST", body),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["me"] });
    }
  });
}

export function useLogout() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => apiSend<{ ok: boolean }>("/api/v1/auth/logout", "POST", {}),
    onSuccess: () => {
      queryClient.removeQueries();
    }
  });
}
