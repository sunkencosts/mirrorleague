import { Navigate, createBrowserRouter } from "react-router";
import Layout from "./components/Layout";
import LeagueError from "./components/LeagueError";
import RootError from "./components/RootError";
import HomePage from "./pages/HomePage";
import LeaguePage from "./pages/LeaguePage";

export const router = createBrowserRouter([
	{
		path: "/",
		element: <Layout />,
		errorElement: <RootError />,
		children: [
			{ index: true, element: <HomePage /> },
			{
				errorElement: <LeagueError />,
				children: [
					{ path: "league/:leagueId", element: <LeaguePage /> },
					{ path: "league/:leagueId/week/:week", element: <LeaguePage /> },
				],
			},
			{ path: "*", element: <Navigate to="/" replace /> },
		],
	},
]);
