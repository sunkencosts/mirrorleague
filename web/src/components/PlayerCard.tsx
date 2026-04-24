import type { Player, SwapOption } from "../types";
import styles from "./PlayerCard.module.css";

interface Props {
  player: Player;
  isSelected?: boolean;
  eligibleSwaps?: SwapOption[];
  onSwapClick?: () => void;
  onSwapSelect?: (opt: SwapOption) => void;
}

export default function PlayerCard({
  player,
  isSelected,
  eligibleSwaps,
  onSwapClick,
  onSwapSelect,
}: Props) {
  return (
    <div
      className={`${styles.playerCard} ${isSelected ? styles.selected : ""}`}
    >
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

      {onSwapClick && (
        <button
          type="button"
          className={styles.swapBtn}
          onClick={onSwapClick}
          title="Swap player"
        >
          ⇄
        </button>
      )}

      {isSelected && (
        <div className={styles.dropdown}>
          {eligibleSwaps && eligibleSwaps.length > 0 ? (
            eligibleSwaps.map((opt) => (
              <button
                key={opt.player.player_id}
                type="button"
                className={styles.dropdownItem}
                onClick={() => onSwapSelect?.(opt)}
              >
                <img
                  src={opt.player.image_url}
                  alt={`${opt.player.first_name} ${opt.player.last_name}`}
                  onError={(e) => {
                    e.currentTarget.style.visibility = "hidden";
                  }}
                />
                <span className={styles.dropdownName}>
                  {opt.player.first_name} {opt.player.last_name}
                </span>
                <span className={styles.dropdownMeta}>
                  {opt.player.fantasy_positions[0]} · {opt.player.team}
                </span>
                {opt.isBench && (
                  <span className={styles.benchBadge}>BENCH</span>
                )}
              </button>
            ))
          ) : (
            <p className={styles.dropdownEmpty}>No eligible players on bench</p>
          )}
        </div>
      )}
    </div>
  );
}
