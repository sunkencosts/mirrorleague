import { useState } from "react";
import { useNavigate } from "react-router";
import styles from "./HomePage.module.css";

export default function HomePage() {
	const [leagueId, setLeagueId] = useState("1322995024962543616");
	const navigate = useNavigate();

	function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
		e.preventDefault();
		if (leagueId.trim()) {
			navigate(`/league/${leagueId.trim()}`);
		}
	}

	return (
		<form onSubmit={handleSubmit} className={styles.leagueForm}>
			<input
				type="text"
				placeholder="Enter Sleeper league ID"
				value={leagueId}
				onChange={(e) => setLeagueId(e.target.value)}
				required
			/>
			<button type="submit">Load League</button>
		</form>
	);
}
