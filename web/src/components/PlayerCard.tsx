import type { Player } from "../types";
import styles from "./PlayerCard.module.css";

interface Props {
  player: Player;
}

export default function PlayerCard({ player }: Props) {
  return (
    <div className={styles.playerCard}>
      <img
        src={player.image_url}
        alt={`${player.first_name} ${player.last_name}`}
        onError={(e) => {
          e.currentTarget.style.visibility = "hidden";
        }}
      />
      <div className={styles.playerInfo}>
        <span className={styles.playerName}>
          {player.first_name} {player.last_name}
        </span>
        <span className={styles.playerMeta}>
          {player.fantasy_positions[0]} · {player.team}
        </span>
      </div>
    </div>
  );
}
