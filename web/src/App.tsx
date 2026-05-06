import { Navigate, Route, Routes } from "react-router";
import Layout from "./components/Layout";
import HomePage from "./pages/HomePage";
import LeaguePage from "./pages/LeaguePage";

export default function App() {
	return (
		<Routes>
			<Route element={<Layout />}>
				<Route path="/" element={<HomePage />} />
				<Route path="/league/:leagueId" element={<LeaguePage />} />
				<Route path="/league/:leagueId/week/:week" element={<LeaguePage />} />
				<Route path="*" element={<Navigate to="/" replace />} />
			</Route>
		</Routes>
	);
}
