import { Outlet } from "react-router";
import styles from "./Layout.module.css";

export default function Layout() {
	return (
		<div className={styles.shell}>
			<header className={styles.header}>
				<h1>Mirror League</h1>
				<p>Mirror a Sleeper league and set your own lineup</p>
			</header>
			<main className={styles.content}>
				<Outlet />
			</main>
		</div>
	);
}
