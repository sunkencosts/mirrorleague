import { useState } from "react";
import type { Roster } from "./types";
import RosterCard from "./components/RosterCard";
import styles from "./App.module.css";

export default function App() {
  const [leagueId, setLeagueId] = useState("1322995024962543616");
  const [rosters, setRosters] = useState<Roster[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function fetchRosters(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setLoading(true);
    setError(null);
    setRosters([]);

    try {
      const res = await fetch(`/api/league/${leagueId}/rosters`);
      if (!res.ok) throw new Error(`${res.status} ${res.statusText}`);
      const data: Roster[] = await res.json();
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
          <RosterCard key={roster.roster_id} roster={roster} />
        ))}
      </div>
    </div>
  );
}
