import type { Roster } from "../types";
import PlayerCard from "./PlayerCard";
import styles from "./RosterCard.module.css";

interface Props {
  roster: Roster;
}

export default function RosterCard({ roster }: Props) {
  const starterIds = new Set(roster.starters.map((p) => p.player_id));
  const bench = roster.players.filter((p) => !starterIds.has(p.player_id));

  return (
    <div className={styles.rosterCard}>
      <h2>{roster.team_name || `Team ${roster.roster_id}`}</h2>

      <div className={styles.section}>
        <h3 className={styles.sectionLabel}>Starters · {roster.starters.length}</h3>
        <div className={styles.playerList}>
          {roster.starters.map((player) => (
            <PlayerCard key={player.player_id} player={player} />
          ))}
        </div>
      </div>

      <div className={styles.section}>
        <h3 className={styles.sectionLabel}>Bench · {bench.length}</h3>
        <div className={styles.playerList}>
          {bench.map((player) => (
            <PlayerCard key={player.player_id} player={player} />
          ))}
        </div>
      </div>
    </div>
  );
}
