export async function fetchJson<T>(url: string): Promise<T> {
	const r = await fetch(url);
	if (!r.ok) {
		throw new Error(`${r.status} ${r.statusText}`);
	}
	return r.json();
}

async function mutateJson<T>(method: string, url: string, body: unknown): Promise<T> {
	const r = await fetch(url, {
		method,
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify(body),
	});
	if (!r.ok) {
		throw new Error(`${r.status} ${r.statusText}`);
	}
	if (r.status === 204) {
		return undefined as T;
	}
	return r.json();
}

export function postJson<T>(url: string, body: unknown): Promise<T> {
	return mutateJson("POST", url, body);
}

export function patchJson<T>(url: string, body: unknown): Promise<T> {
	return mutateJson("PATCH", url, body);
}

export async function deleteJson(url: string): Promise<void> {
	const r = await fetch(url, { method: "DELETE" });
	if (!r.ok) {
		throw new Error(`${r.status} ${r.statusText}`);
	}
}

export function bookmarksKey(userId: string): ["bookmarks", string] {
	return ["bookmarks", userId];
}
