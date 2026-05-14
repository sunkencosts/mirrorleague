import { useState } from "react";

export const ANON_USER_ID_KEY = "mirror_me_user_id";

export function useUserId(): string {
	const [userId] = useState(() => {
		const existing = localStorage.getItem(ANON_USER_ID_KEY);
		if (existing) {
			return existing;
		}
		const id = crypto.randomUUID();
		localStorage.setItem(ANON_USER_ID_KEY, id);
		return id;
	});
	return userId;
}
