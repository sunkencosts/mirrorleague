import { Link, useRouteError } from "react-router";
import styles from "./RouteError.module.css";

export default function RootError() {
	const error = useRouteError();
	const message = error instanceof Error ? error.message : "Something went wrong";
	return (
		<div className={styles.container}>
			<p className={styles.message}>{message}</p>
			<Link to="/">← Back to home</Link>
		</div>
	);
}
