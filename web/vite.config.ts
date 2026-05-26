import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

export default defineConfig({
	plugins: [react()],
	server: {
		proxy: {
			"/auth": "http://localhost:8080",
			"/dev": "http://localhost:8080",
			"/league": "http://localhost:8080",
			"/lineups": "http://localhost:8080",
			"/players": "http://localhost:8080",
			"/league-bookmarks": "http://localhost:8080",
			"/admin": "http://localhost:8080",
		},
	},
});
