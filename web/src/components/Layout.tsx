import { useState } from "react";
import { Link, Outlet } from "react-router";
import styles from "./Layout.module.css";

export default function Layout() {
	const [menuOpen, setMenuOpen] = useState(false);

	return (
		<div className={styles.shell}>
			<header className={styles.header}>
				<div className={styles.headerTop}>
					<Link to="/" className={styles.titleLink}>
						<h1>Mirror League</h1>
					</Link>
					<button
						type="button"
						className={styles.hamburger}
						onClick={() => setMenuOpen((o) => !o)}
						aria-label="Toggle menu"
					>
						<span aria-hidden="true">{menuOpen ? "✕" : "☰"}</span>
					</button>
				</div>
				<p>Mirror a Sleeper league and set your own lineup</p>
				<nav className={`${styles.nav} ${menuOpen ? styles.navOpen : ""}`}>
					<Link to="/" onClick={() => setMenuOpen(false)}>Home</Link>
				</nav>
			</header>
			<main className={styles.content}>
				<Outlet />
			</main>
		</div>
	);
}
