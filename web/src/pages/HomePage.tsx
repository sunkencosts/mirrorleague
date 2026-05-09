import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { useNavigate } from "react-router";
import { bookmarksKey, postJson } from "../api";
import LeagueBookmarks from "../components/LeagueBookmarks";
import { useUserId } from "../hooks/useUserId";
import type { LeagueBookmark } from "../types";
import styles from "./HomePage.module.css";

export default function HomePage() {
	const [leagueId, setLeagueId] = useState("1182073403987832832");
	const [label, setLabel] = useState("");
	const navigate = useNavigate();
	const userId = useUserId();
	const queryClient = useQueryClient();

	const saveBookmark = useMutation({
		mutationFn: () =>
			postJson<LeagueBookmark>("/api/league-bookmarks", {
				user_id: userId,
				league_id: leagueId.trim(),
				label: label.trim(),
				source: "sleeper",
			}),
		onSuccess: () => queryClient.invalidateQueries({ queryKey: bookmarksKey(userId) }),
	});

	function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
		e.preventDefault();
		const id = leagueId.trim();
		if (!id) {
			return;
		}
		saveBookmark.mutate();
		navigate(`/league/${id}`);
	}

	return (
		<>
			<LeagueBookmarks userId={userId} />
			<h2 className={styles.formHeading}>Connect League</h2>
			<form onSubmit={handleSubmit} className={styles.leagueForm}>
				<input
					type="text"
					placeholder="Enter Sleeper league ID"
					value={leagueId}
					onChange={(e) => setLeagueId(e.target.value)}
					required
				/>
				<input
					type="text"
					placeholder="Label (optional)"
					value={label}
					onChange={(e) => setLabel(e.target.value)}
					className={styles.labelInput}
				/>
				<button type="submit">Load League</button>
			</form>
		</>
	);
}
