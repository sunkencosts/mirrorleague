import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect } from "react";
import { postJson, fetchJson } from "../api";
import type { AuthUser } from "../types";
import { ANON_USER_ID_KEY, useUserId } from "./useUserId";

// Internal — consume auth state via useAuth() from context/AuthContext instead.
export function useAuthState() {
	const queryClient = useQueryClient();
	const { data: user = null, isLoading } = useQuery({
		queryKey: ["auth"],
		queryFn: async () => {
			try {
				return await fetchJson<AuthUser>("/auth/me");
			} catch {
				return null;
			}
		},
		staleTime: 5 * 60 * 1000,
	});

	useEffect(() => {
		if (!user) {
			return;
		}
		const anonId = localStorage.getItem(ANON_USER_ID_KEY);
		if (!anonId) {
			return;
		}
		async function merge() {
			try {
				await postJson("/auth/merge", { anonymous_id: anonId });
				localStorage.removeItem(ANON_USER_ID_KEY);
				queryClient.invalidateQueries();
			} catch {
				// no-op — backend is idempotent, will retry on next render
			}
		}
		merge();
	}, [user, queryClient]);

	const anonId = useUserId();
	const userId = user?.id ?? anonId;

	return { user, isLoading, userId };
}
