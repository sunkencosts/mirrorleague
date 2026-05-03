import { useQuery } from "@tanstack/react-query";
import { useMemo } from "react";
import { useParams } from "react-router";
import { fetchJson } from "../api";
import styles from "./LeaguePage.module.css";
import LeagueSummary from "../components/LeagueSummary";
import RosterCard from "../components/RosterCard";
import { computePowerScore } from "../scoring";
import type { League, LeagueConfig, Lineup, Roster } from "../types";

const userId = "00000000-0000-0000-0000-000000000001";
const weekNumber = 1;

export default function LeaguePage() {
  const { leagueId = "" } = useParams();

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

  const { data: lineups = [], isLoading: lineupsLoading } = useQuery<Lineup[]>({
    queryKey: ["lineups", userId, leagueId, weekNumber],
    queryFn: () =>
      fetchJson(
        `/api/lineups?user_id=${userId}&league_id=${leagueId}&week_number=${weekNumber}`,
      ),
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
  if (leagueLoading || rostersLoading || lineupsLoading) {
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
      <LeagueSummary league={league} />
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
