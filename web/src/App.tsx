import { useState, useMemo } from "react";
import type { Roster, League } from "./types";
import RosterCard from "./components/RosterCard";
import { computePowerScore } from "./scoring";
import styles from "./App.module.css";
import { useQuery } from "@tanstack/react-query";
import type { Lineup } from "./types";

interface LeagueConfig {
  starterSlots: string[];
  benchSlots: number;
  irSlots: number;
  taxiSlots: number;
}

export default function App() {
  const [userId, setUserId] = useState("00000000-0000-0000-0000-000000000001");
  const [weekNumber] = useState(1);
  const [leagueId, setLeagueId] = useState("1322995024962543616");
  const [rosters, setRosters] = useState<Roster[]>([]);
  const [leagueConfig, setLeagueConfig] = useState<LeagueConfig | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function fetchRosters(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setLoading(true);
    setError(null);
    setRosters([]);
    setLeagueConfig(null);

    try {
      const [leagueRes, rostersRes] = await Promise.all([
        fetch(`/api/league/${leagueId}`),
        fetch(`/api/league/${leagueId}/rosters`),
      ]);
      if (!leagueRes.ok)
        throw new Error(`${leagueRes.status} ${leagueRes.statusText}`);
      if (!rostersRes.ok)
        throw new Error(`${rostersRes.status} ${rostersRes.statusText}`);
      const league: League = await leagueRes.json();
      const data: Roster[] = await rostersRes.json();
      setLeagueConfig({
        starterSlots: league.roster_positions.filter((p) => p !== "BN"),
        benchSlots: league.roster_positions.filter((p) => p === "BN").length,
        irSlots: league.settings.reserve_slots,
        taxiSlots: league.settings.taxi_slots,
      });
      setRosters(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Something went wrong");
    } finally {
      setLoading(false);
    }
  }

  const allScores = useMemo(
    () => rosters.map((r) => computePowerScore(r.players, r.starters)),
    [rosters],
  );
  const { data: lineups = [], isLoading: lineupsLoading } = useQuery<Lineup[]>({
    queryKey: ["lineups", userId, leagueId, weekNumber],
    queryFn: () =>
      fetch(
        `/api/lineups?user_id=${userId}&league_id=${leagueId}&week_number=${weekNumber}`,
      ).then((r) => {
        if (!r.ok) throw new Error(`${r.status}`);
        return r.json();
      }),
    enabled: rosters.length > 0,
  });

  return (
    <div className={styles.app}>
      <header className={styles.header}>
        <h1>Mirror Me</h1>
        <p>Mirror a Sleeper league and set your own lineup</p>
      </header>

      <form onSubmit={fetchRosters} className={styles.leagueForm}>
        <input
          type="text"
          placeholder="Enter Sleeper league ID"
          value={leagueId}
          onChange={(e) => setLeagueId(e.target.value)}
          required
        />
        <input
          type="text"
          placeholder="Your user ID"
          value={userId}
          onChange={(e) => setUserId(e.target.value)}
        />
        <button type="submit" disabled={loading}>
          {loading ? "Loading…" : "Load League"}
        </button>
      </form>

      {error && <p className={styles.error}>{error}</p>}

      <div className={styles.rosterList}>
        {leagueConfig && !lineupsLoading &&
          rosters.map((roster) => (
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
    </div>
  );
}
