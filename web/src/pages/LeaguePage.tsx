import { useQuery } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import { useParams } from "react-router";
import { fetchJson } from "../api";
import LeagueSummary from "../components/LeagueSummary";
import RosterCard from "../components/RosterCard";
import { computePowerScore } from "../scoring";
import type { League, LeagueConfig, Lineup, Roster } from "../types";
import styles from "./LeaguePage.module.css";

const userId = "00000000-0000-0000-0000-000000000001";

export default function LeaguePage() {
	const { leagueId = "" } = useParams();
	const [weekNumber, setWeekNumber] = useState(1);

	const {
		data: league,
		isLoading: leagueLoading,
		error: leagueError,
	} = useQuery<League>({
		queryKey: ["league", leagueId],
		queryFn: () => fetchJson(`/api/league/${leagueId}`),
		enabled: !!leagueId,
	});

	const {
		data: rosters = [],
		isLoading: rostersLoading,
		error: rostersError,
	} = useQuery<Roster[]>({
		queryKey: ["rosters", leagueId],
		queryFn: () => fetchJson(`/api/league/${leagueId}/rosters`),
		enabled: !!leagueId,
	});

	const { data: lineups = [] } = useQuery<Lineup[]>({
		queryKey: ["lineups", userId, leagueId, weekNumber],
		queryFn: () =>
			fetchJson(`/api/lineups?user_id=${userId}&league_id=${leagueId}&week_number=${weekNumber}`),
		enabled: !!leagueId,
	});

	const leagueConfig = useMemo<LeagueConfig | null>(() => {
		if (!league) return null;
		const starterSlots = league.roster_positions.filter((p) => p !== "BN");
		return {
			starterSlots,
			benchSlots: league.roster_positions.length - starterSlots.length,
			irSlots: league.settings.reserve_slots,
			taxiSlots: league.settings.taxi_slots,
		};
	}, [league]);

	const allScores = useMemo(
		() => rosters.map((r) => computePowerScore(r.players, r.starters)),
		[rosters],
	);

	const error = leagueError ?? rostersError;
	if (leagueLoading || rostersLoading) {
		return <p>Loading…</p>;
	}
	if (error)
		return (
			<p className={styles.error}>
				{error instanceof Error ? error.message : "Something went wrong"}
			</p>
		);
	if (!leagueConfig || !league) {
		return null;
	}

	return (
		<>
			<LeagueSummary league={league} weekNumber={weekNumber} onWeekChange={setWeekNumber} />
			<div className={styles.rosterList}>
				{rosters.map((roster) => (
					<RosterCard
						key={roster.roster_id}
						roster={roster}
						{...leagueConfig}
						allScores={allScores}
						leagueId={leagueId}
						userId={userId}
						lineups={lineups}
						weekNumber={weekNumber}
					/>
				))}
			</div>
		</>
	);
}
