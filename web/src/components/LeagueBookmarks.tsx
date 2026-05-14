import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router";
import { bookmarksKey, deleteJson, fetchJson, patchJson } from "../api";
import type { LeagueBookmark } from "../types";
import { onImageError } from "../utils/playerImage";
import { CheckIcon, PencilIcon, TrashIcon, XIcon } from "./icons";
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
		mutationFn: ({ leagueId, source }: { leagueId: string; source: string }) =>
			deleteJson(`/api/league-bookmarks/${leagueId}?user_id=${userId}&source=${source}`),
		onSuccess: () => queryClient.invalidateQueries({ queryKey: bookmarksKey(userId) }),
	});

	const patchMutation = useMutation({
		mutationFn: ({
			leagueId,
			label,
			source,
		}: {
			leagueId: string;
			label: string;
			source: string;
		}) =>
			patchJson<LeagueBookmark>(`/api/league-bookmarks/${leagueId}?source=${source}`, {
				user_id: userId,
				label,
			}),
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
		const source = bookmarks.find((b) => b.league_id === leagueId)?.source ?? "";
		patchMutation.mutate({ leagueId, label: editLabel.trim(), source });
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
					const isDeleting =
						deleteMutation.isPending && deleteMutation.variables?.leagueId === b.league_id;

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
										{b.icon_url && (
											<img src={b.icon_url} alt="" className={styles.icon} onError={onImageError} />
										)}
										<span className={styles.info}>
											<span className={styles.label}>{b.label || b.league_id}</span>
											{b.label && <span className={styles.id}>{b.league_id}</span>}
										</span>
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
												<CheckIcon />
											</button>
											<button
												type="button"
												className={`${styles.actionBtn} ${styles.cancel}`}
												onClick={cancelEdit}
												aria-label="Cancel"
												title="Cancel"
											>
												<XIcon />
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
												<PencilIcon />
											</button>
											<button
												type="button"
												className={`${styles.actionBtn} ${styles.delete}`}
												onClick={() =>
													deleteMutation.mutate({ leagueId: b.league_id, source: b.source })
												}
												disabled={isDeleting}
												aria-label="Delete bookmark"
												title="Delete bookmark"
											>
												<TrashIcon />
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
