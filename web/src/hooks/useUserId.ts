import { useState } from "react";

const KEY = "mirror_me_user_id";

export function useUserId(): string {
	const [userId] = useState(() => {
		const existing = localStorage.getItem(KEY);
		if (existing) {
			return existing;
		}
		const id = crypto.randomUUID();
		localStorage.setItem(KEY, id);
		return id;
	});
	return userId;
}
