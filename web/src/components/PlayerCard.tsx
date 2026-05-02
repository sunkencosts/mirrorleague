import type { Player, SwapOption } from "../types";
import { RARITY_GLOW } from "../rarity";
import styles from "./PlayerCard.module.css";

export const PROFILE_FALLBACK =
  "data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24'><rect width='24' height='24' fill='%231e3020'/><circle cx='12' cy='8' r='4' fill='%233a5c3a'/><path d='M5 20c0-4.2 3.5-7 7-7s7 2.8 7 7' fill='%233a5c3a'/></svg>";

interface Props {
  player: Player;
  isSelected?: boolean;
  eligibleSwaps?: SwapOption[];
  onSwapSelect?: (opt: SwapOption) => void;
  onMoveToEmpty?: () => void;
  onSelect?: () => void;
  reversed?: boolean;
}
export default function PlayerCard({
  player,
  isSelected,
  eligibleSwaps,
  onSwapSelect,
  onMoveToEmpty,
  onSelect,
  reversed,
}: Props) {
  return (
    <div
      className={`${styles.playerCard} ${isSelected ? styles.selected : ""} ${reversed ? styles.reversed : ""}`}
    >
      <img
        style={{
          boxShadow: RARITY_GLOW[player.rarity || "grey"],
          cursor: onSelect ? "pointer" : undefined,
        }}
        src={player.image_url}
        alt={`${player.first_name} ${player.last_name}`}
        role={onSelect ? "button" : undefined}
        tabIndex={onSelect ? 0 : undefined}
        onClick={onSelect}
        onKeyDown={
          onSelect
            ? (e) => {
                if (e.key === "Enter" || e.key === " ") onSelect();
              }
            : undefined
        }
        onError={(e) => {
          e.currentTarget.onerror = null;
          e.currentTarget.src = PROFILE_FALLBACK;
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
                    e.currentTarget.onerror = null;
                    e.currentTarget.src = PROFILE_FALLBACK;
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
          ) : !onMoveToEmpty ? (
            <p className={styles.dropdownEmpty}>No eligible players</p>
          ) : null}
          {onMoveToEmpty && (
            <button
              type="button"
              className={`${styles.dropdownItem} ${styles.moveToEmpty}`}
              onClick={onMoveToEmpty}
            >
              Move to bench
            </button>
          )}
        </div>
      )}
    </div>
  );
}
