import { useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { Link, Outlet } from "react-router";
import { deleteJson } from "../api";
import { useAuth } from "../context/AuthContext";
import type { AuthUser } from "../types";
import styles from "./Layout.module.css";

interface AuthControlsProps {
	user: AuthUser | null;
	isLoading: boolean;
	onLogout: () => void;
	usernameCls: string;
	buttonCls: string;
}

function AuthControls({ user, isLoading, onLogout, usernameCls, buttonCls }: AuthControlsProps) {
	if (isLoading) {
		return null;
	}
	if (user) {
		return (
			<>
				<span className={usernameCls}>{user.username}</span>
				<button type="button" className={buttonCls} onClick={onLogout}>
					Log out
				</button>
			</>
		);
	}
	return (
		<button
			type="button"
			className={buttonCls}
			onClick={() => {
				window.location.href = `${import.meta.env.VITE_API_URL ?? ""}/auth/google`;
			}}
		>
			Log in
		</button>
	);
}

export default function Layout() {
	const [menuOpen, setMenuOpen] = useState(false);
	const { user, isLoading } = useAuth();
	const queryClient = useQueryClient();

	async function handleLogout() {
		await deleteJson("/auth/logout");
		queryClient.invalidateQueries({ queryKey: ["auth"] });
	}

	return (
		<div className={styles.shell}>
			<header className={styles.header}>
				<div className={styles.headerTop}>
					<Link to="/" className={styles.titleLink}>
						<h1>Mirror League</h1>
					</Link>
					<div className={styles.headerActions}>
						<AuthControls
							user={user}
							isLoading={isLoading}
							onLogout={handleLogout}
							usernameCls={styles.username}
							buttonCls={styles.authBtn}
						/>
						<button
							type="button"
							className={styles.hamburger}
							onClick={() => setMenuOpen((o) => !o)}
							aria-label="Toggle menu"
						>
							<span aria-hidden="true">{menuOpen ? "✕" : "☰"}</span>
						</button>
					</div>
				</div>
				<p>Mirror a Sleeper league and set your own lineup</p>
				<nav className={`${styles.nav} ${menuOpen ? styles.navOpen : ""}`}>
					<Link to="/" onClick={() => setMenuOpen(false)}>
						Home
					</Link>
					<AuthControls
						user={user}
						isLoading={isLoading}
						onLogout={handleLogout}
						usernameCls={styles.navUsername}
						buttonCls={styles.navAuthBtn}
					/>
				</nav>
			</header>
			<main className={styles.content}>
				<Outlet />
			</main>
		</div>
	);
}
