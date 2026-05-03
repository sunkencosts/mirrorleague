import { Navigate, Route, Routes } from "react-router";
import styles from "./App.module.css";
import HomePage from "./pages/HomePage";
import LeaguePage from "./pages/LeaguePage";

export default function App() {
	return (
		<div className={styles.app}>
			<header className={styles.header}>
				<h1>Mirror League</h1>
				<p>Mirror a Sleeper league and set your own lineup</p>
			</header>

			<Routes>
				<Route path="/" element={<HomePage />} />
				<Route path="/league/:leagueId" element={<LeaguePage />} />
				<Route path="*" element={<Navigate to="/" replace />} />
			</Routes>
		</div>
	);
}
