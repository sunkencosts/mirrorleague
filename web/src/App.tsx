import { useState } from "react";
import type { Roster, League } from "./types";
import RosterCard from "./components/RosterCard";
import styles from "./App.module.css";

export default function App() {
  const [leagueId, setLeagueId] = useState("1322995024962543616");
  const [rosters, setRosters] = useState<Roster[]>([]);
  const [starterSlots, setStarterSlots] = useState<string[]>([]);
  const [benchSlots, setBenchSlots] = useState(0);
  const [irSlots, setIrSlots] = useState(0);
  const [taxiSlots, setTaxiSlots] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function fetchRosters(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setLoading(true);
    setError(null);
    setRosters([]);
    setStarterSlots([]);
    setBenchSlots(0);
    setIrSlots(0);
    setTaxiSlots(0);

    try {
      const [leagueRes, rostersRes] = await Promise.all([
        fetch(`/api/league/${leagueId}`),
        fetch(`/api/league/${leagueId}/rosters`),
      ]);
      if (!leagueRes.ok) throw new Error(`${leagueRes.status} ${leagueRes.statusText}`);
      if (!rostersRes.ok) throw new Error(`${rostersRes.status} ${rostersRes.statusText}`);
      const league: League = await leagueRes.json();
      const data: Roster[] = await rostersRes.json();
      setStarterSlots(league.roster_positions.filter((p) => p !== "BN"));
      setBenchSlots(league.roster_positions.filter((p) => p === "BN").length);
      setIrSlots(league.settings.reserve_slots);
      setTaxiSlots(league.settings.taxi_slots);
      setRosters(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Something went wrong");
    } finally {
      setLoading(false);
    }
  }

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
        <button type="submit" disabled={loading}>
          {loading ? "Loading…" : "Load League"}
        </button>
      </form>

      {error && <p className={styles.error}>{error}</p>}

      <div className={styles.rosterGrid}>
        {rosters.map((roster) => (
          <RosterCard key={roster.roster_id} roster={roster} starterSlots={starterSlots} benchSlots={benchSlots} irSlots={irSlots} taxiSlots={taxiSlots} />
        ))}
      </div>
    </div>
  );
}
