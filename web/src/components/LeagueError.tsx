import { Link, useNavigate, useRouteError } from "react-router";
import type { ApiError } from "../api";
import styles from "./RouteError.module.css";

export default function LeagueError() {
	const error = useRouteError();
	const navigate = useNavigate();
	const status = (error as ApiError)?.status;

	if (status === 404) {
		return (
			<div className={styles.container}>
				<p className={styles.message}>League not found. Check that the league ID is correct.</p>
				<Link to="/">← Back to home</Link>
			</div>
		);
	}

	return (
		<div className={styles.container}>
			<p className={styles.message}>
				{error instanceof Error ? error.message : "Something went wrong"}
			</p>
			<div className={styles.actions}>
				<button type="button" onClick={() => navigate(0)}>Try again</button>
				<Link to="/">← Back to home</Link>
			</div>
		</div>
	);
}
