import { useState, useMemo } from "react";
import { Routes, Route, Navigate, useNavigate, useParams } from "react-router";
import type { Lineup, Roster, League } from "./types";
import RosterCard from "./components/RosterCard";
import { computePowerScore } from "./scoring";
import styles from "./App.module.css";
import { useQuery } from "@tanstack/react-query";

interface LeagueConfig {
  starterSlots: string[];
  benchSlots: number;
  irSlots: number;
  taxiSlots: number;
}

function HomeForm() {
  const [leagueId, setLeagueId] = useState("1322995024962543616");
  const navigate = useNavigate();

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (leagueId.trim()) navigate(`/league/${leagueId.trim()}`);
  }

  return (
    <form onSubmit={handleSubmit} className={styles.leagueForm}>
      <input
        type="text"
        placeholder="Enter Sleeper league ID"
        value={leagueId}
        onChange={(e) => setLeagueId(e.target.value)}
        required
      />
      <button type="submit">Load League</button>
    </form>
  );
}

function LeaguePage() {
  const { leagueId = "" } = useParams();
  const userId = "00000000-0000-0000-0000-000000000001";
  const weekNumber = 1;

  const { data: league, isLoading: leagueLoading, error: leagueError } =
    useQuery<League>({
      queryKey: ["league", leagueId],
      queryFn: () =>
        fetch(`/api/league/${leagueId}`).then((r) => {
          if (!r.ok) throw new Error(`${r.status} ${r.statusText}`);
          return r.json();
        }),
      enabled: !!leagueId,
    });

  const { data: rosters = [], isLoading: rostersLoading, error: rostersError } =
    useQuery<Roster[]>({
      queryKey: ["rosters", leagueId],
      queryFn: () =>
        fetch(`/api/league/${leagueId}/rosters`).then((r) => {
          if (!r.ok) throw new Error(`${r.status} ${r.statusText}`);
          return r.json();
        }),
      enabled: !!leagueId,
    });

  const { data: lineups = [], isLoading: lineupsLoading } = useQuery<Lineup[]>({
    queryKey: ["lineups", userId, leagueId, weekNumber],
    queryFn: () =>
      fetch(
        `/api/lineups?user_id=${userId}&league_id=${leagueId}&week_number=${weekNumber}`,
      ).then((r) => {
        if (!r.ok) throw new Error(`${r.status}`);
        return r.json();
      }),
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
  if (leagueLoading || rostersLoading || lineupsLoading) return <p>Loading…</p>;
  if (error) return <p className={styles.error}>{error instanceof Error ? error.message : "Something went wrong"}</p>;
  if (!leagueConfig) return null;

  return (
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
  );
}

export default function App() {
  return (
    <div className={styles.app}>
      <header className={styles.header}>
        <h1>Mirror Me</h1>
        <p>Mirror a Sleeper league and set your own lineup</p>
      </header>

      <Routes>
        <Route path="/" element={<HomeForm />} />
        <Route path="/league/:leagueId" element={<LeaguePage />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </div>
  );
}
