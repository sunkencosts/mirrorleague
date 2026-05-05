import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router";
import { bookmarksKey, deleteJson, fetchJson, patchJson } from "../api";
import type { LeagueBookmark } from "../types";
import styles from "./LeagueBookmarks.module.css";

interface Props {
	userId: string;
}

export default function LeagueBookmarks({ userId }: Props) {
	const navigate = useNavigate();
	const queryClient = useQueryClient();
	const { data: bookmarks = [] } = useQuery<LeagueBookmark[]>({
		queryKey: bookmarksKey(userId),
		queryFn: () => fetchJson(`/api/league-bookmarks?user_id=${userId}`),
	});

	const [editingId, setEditingId] = useState<string | null>(null);
	const [editLabel, setEditLabel] = useState("");
	const editInputRef = useRef<HTMLInputElement>(null);

	useEffect(() => {
		if (editingId) {
			editInputRef.current?.focus();
		}
	}, [editingId]);

	const deleteMutation = useMutation({
		mutationFn: (leagueId: string) =>
			deleteJson(`/api/league-bookmarks/${leagueId}?user_id=${userId}`),
		onSuccess: () => queryClient.invalidateQueries({ queryKey: bookmarksKey(userId) }),
	});

	const patchMutation = useMutation({
		mutationFn: ({ leagueId, label }: { leagueId: string; label: string }) =>
			patchJson<LeagueBookmark>(`/api/league-bookmarks/${leagueId}`, { user_id: userId, label }),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: bookmarksKey(userId) });
			setEditingId(null);
		},
	});

	function startEdit(b: LeagueBookmark) {
		setEditingId(b.league_id);
		setEditLabel(b.label);
	}

	function cancelEdit() {
		setEditingId(null);
	}

	function saveEdit(leagueId: string) {
		patchMutation.mutate({ leagueId, label: editLabel.trim() });
	}

	function handleEditKeyDown(e: React.KeyboardEvent, leagueId: string) {
		if (e.key === "Enter") {
			saveEdit(leagueId);
		}
		if (e.key === "Escape") {
			cancelEdit();
		}
	}

	if (bookmarks.length === 0) {
		return null;
	}

	return (
		<section className={styles.section}>
			<h2 className={styles.heading}>Saved Leagues</h2>
			<ul className={styles.list}>
				{bookmarks.map((b) => {
					const isEditing = editingId === b.league_id;
					const isDeleting = deleteMutation.isPending && deleteMutation.variables === b.league_id;

					return (
						<li key={b.league_id}>
							<div className={styles.row}>
								{isEditing ? (
									<input
										ref={editInputRef}
										className={styles.editInput}
										value={editLabel}
										onChange={(e) => setEditLabel(e.target.value)}
										onKeyDown={(e) => handleEditKeyDown(e, b.league_id)}
									/>
								) : (
									<button
										type="button"
										className={styles.navigate}
										onClick={() => navigate(`/league/${b.league_id}`)}
									>
										<span className={styles.label}>{b.label || b.league_id}</span>
										{b.label && <span className={styles.id}>{b.league_id}</span>}
									</button>
								)}

								<div className={styles.actions}>
									{isEditing ? (
										<>
											<button
												type="button"
												className={`${styles.actionBtn} ${styles.save}`}
												onClick={() => saveEdit(b.league_id)}
												disabled={patchMutation.isPending}
												aria-label="Save"
												title="Save"
											>
												<svg
													width="15"
													height="15"
													viewBox="0 0 24 24"
													fill="none"
													stroke="currentColor"
													strokeWidth="2.5"
													strokeLinecap="round"
													strokeLinejoin="round"
													aria-hidden="true"
												>
													<polyline points="20,6 9,17 4,12" />
												</svg>
											</button>
											<button
												type="button"
												className={`${styles.actionBtn} ${styles.cancel}`}
												onClick={cancelEdit}
												aria-label="Cancel"
												title="Cancel"
											>
												<svg
													width="15"
													height="15"
													viewBox="0 0 24 24"
													fill="none"
													stroke="currentColor"
													strokeWidth="2"
													strokeLinecap="round"
													strokeLinejoin="round"
													aria-hidden="true"
												>
													<line x1="18" y1="6" x2="6" y2="18" />
													<line x1="6" y1="6" x2="18" y2="18" />
												</svg>
											</button>
										</>
									) : (
										<>
											<button
												type="button"
												className={styles.actionBtn}
												onClick={() => startEdit(b)}
												aria-label="Edit label"
												title="Edit label"
											>
												<svg
													width="15"
													height="15"
													viewBox="0 0 24 24"
													fill="none"
													stroke="currentColor"
													strokeWidth="2"
													strokeLinecap="round"
													strokeLinejoin="round"
													aria-hidden="true"
												>
													<path d="M12 20h9" />
													<path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4Z" />
												</svg>
											</button>
											<button
												type="button"
												className={`${styles.actionBtn} ${styles.delete}`}
												onClick={() => deleteMutation.mutate(b.league_id)}
												disabled={isDeleting}
												aria-label="Delete bookmark"
												title="Delete bookmark"
											>
												<svg
													width="15"
													height="15"
													viewBox="0 0 24 24"
													fill="none"
													stroke="currentColor"
													strokeWidth="2"
													strokeLinecap="round"
													strokeLinejoin="round"
													aria-hidden="true"
												>
													<polyline points="3,6 5,6 21,6" />
													<path d="M19 6l-.867 12.142A2 2 0 0 1 16.138 20H7.862a2 2 0 0 1-1.995-1.858L5 6" />
													<path d="M10 11v6M14 11v6" />
													<path d="M9 6V4a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2" />
												</svg>
											</button>
										</>
									)}
								</div>
							</div>
						</li>
					);
				})}
			</ul>
		</section>
	);
}
