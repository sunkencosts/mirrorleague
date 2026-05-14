import { keepPreviousData, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useMemo, useRef } from "react";
import { useNavigate, useParams } from "react-router";
import { bookmarksKey, fetchJson, patchJson } from "../api";
import LeagueSummary from "../components/LeagueSummary";
import PlayerSearch from "../components/PlayerSearch";
import RosterCard from "../components/RosterCard";
import { useAuth } from "../context/AuthContext";
import { computePowerScore } from "../scoring";
import type { League, LeagueBookmark, LeagueConfig, Lineup, Roster, WeekMatchup } from "../types";
import styles from "./LeaguePage.module.css";

export default function LeaguePage() {
	const { leagueId = "", week } = useParams();
	const weekNumber = week ? parseInt(week, 10) : 1;
	const navigate = useNavigate();
	const { userId } = useAuth();
	const queryClient = useQueryClient();

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

	const { data: weekMatchups = [] } = useQuery<WeekMatchup[]>({
		queryKey: ["week-matchups", leagueId, weekNumber],
		queryFn: () => fetchJson(`/api/league/${leagueId}/week/${weekNumber}`),
		select: (data) => data ?? [],
		placeholderData: keepPreviousData,
		enabled: !!leagueId,
	});

	const matchupByRosterId = useMemo(
		() => new Map(weekMatchups.map((m) => [m.roster_id, m])),
		[weekMatchups],
	);

	const { data: lineups = [] } = useQuery<Lineup[]>({
		queryKey: ["lineups", userId, leagueId, weekNumber],
		queryFn: () =>
			fetchJson(`/api/lineups?user_id=${userId}&league_id=${leagueId}&week_number=${weekNumber}`),
		enabled: !!leagueId,
	});

	const { data: bookmarks = [] } = useQuery<LeagueBookmark[]>({
		queryKey: bookmarksKey(userId),
		queryFn: () => fetchJson(`/api/league-bookmarks?user_id=${userId}`),
	});

	const { mutate: patchLabel } = useMutation({
		mutationFn: ({ label, source }: { label: string; source: string }) =>
			patchJson<LeagueBookmark>(`/api/league-bookmarks/${leagueId}?source=${source}`, {
				user_id: userId,
				label,
			}),
		onSuccess: () => queryClient.invalidateQueries({ queryKey: bookmarksKey(userId) }),
	});

	const patched = useRef(false);
	const highlightTimer = useRef<ReturnType<typeof setTimeout>>();
	useEffect(() => () => clearTimeout(highlightTimer.current), []);

	useEffect(() => {
		if (patched.current || !league?.name) {
			return;
		}
		const bookmark = bookmarks.find((b) => b.league_id === leagueId);
		if (!bookmark || bookmark.label) {
			return;
		}
		patched.current = true;
		patchLabel({ label: league.name, source: bookmark.source });
	}, [league, bookmarks, leagueId, patchLabel]);

	const leagueConfig = useMemo<LeagueConfig | null>(() => {
		if (!league) {
			return null;
		}
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
	if (error) {
		return (
			<p className={styles.error}>
				{error instanceof Error ? error.message : "Something went wrong"}
			</p>
		);
	}
	if (!leagueConfig || !league) {
		return null;
	}

	function scrollToRoster(rosterId: number) {
		const el = document.getElementById(`roster-${rosterId}`);
		if (!el) return;
		el.scrollIntoView({ behavior: "smooth", block: "start" });
		el.classList.add(styles.highlighted);
		clearTimeout(highlightTimer.current);
		highlightTimer.current = setTimeout(() => el.classList.remove(styles.highlighted), 1500);
	}

	return (
		<>
			<LeagueSummary
				league={league}
				weekNumber={weekNumber}
				onWeekChange={(w) => navigate(`/league/${leagueId}/week/${w}`)}
			/>
			<PlayerSearch rosters={rosters} onScrollToRoster={scrollToRoster} />
			<div className={styles.rosterList}>
				{rosters.map((roster) => (
					<div key={roster.roster_id} id={`roster-${roster.roster_id}`} className={styles.rosterWrapper}>
						<RosterCard
							roster={roster}
							weekMatchup={matchupByRosterId.get(roster.roster_id) ?? null}
							{...leagueConfig}
							allScores={allScores}
							leagueId={leagueId}
							userId={userId}
							lineups={lineups}
							weekNumber={weekNumber}
						/>
					</div>
				))}
			</div>
		</>
	);
}
