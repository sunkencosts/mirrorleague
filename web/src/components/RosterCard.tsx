import { useState } from "react";
import type { Roster, Player, SwapOption } from "../types";
import PlayerCard, { PROFILE_FALLBACK } from "./PlayerCard";
import RosterRarity from "./RosterRarity";
import styles from "./RosterCard.module.css";

const SLOT_DISPLAY: Record<string, string> = {
  SUPER_FLEX: "SF",
};

function slotLabel(slot: string): string {
  return SLOT_DISPLAY[slot] ?? slot;
}

const SLOT_ELIGIBILITY: Record<string, string[]> = {
  QB: ["QB"],
  RB: ["RB"],
  WR: ["WR"],
  TE: ["TE"],
  K: ["K"],
  DEF: ["DEF"],
  FLEX: ["RB", "WR", "TE"],
  SUPER_FLEX: ["QB", "RB", "WR", "TE"],
  IDP_FLEX: ["DL", "LB", "DB"],
  DL: ["DL"],
  LB: ["LB"],
  DB: ["DB"],
};

function canFillSlot(slot: string, player: Player): boolean {
  const eligible = SLOT_ELIGIBILITY[slot];
  if (!eligible) return false;
  return player.fantasy_positions.some((pos) => eligible.includes(pos));
}

interface Props {
  roster: Roster;
  starterSlots: string[];
  benchSlots: number;
  irSlots: number;
  taxiSlots: number;
  allScores: number[];
}

export default function RosterCard({ roster, starterSlots, benchSlots, irSlots, taxiSlots, allScores }: Props) {
  const [localStarters, setLocalStarters] = useState<(Player | null)[]>(
    () => Array.from({ length: starterSlots.length }, (_, i) => roster.starters[i] ?? null),
  );
  const [selectedIndex, setSelectedIndex] = useState<number | null>(null);

  const starterIds = new Set(localStarters.flatMap((p) => (p ? [p.player_id] : [])));
  const bench = roster.players.filter((p) => !starterIds.has(p.player_id));
  const hasEmptyBench = bench.length < benchSlots;

  function select(i: number) {
    setSelectedIndex((prev) => (prev === i ? null : i));
  }

  function handleMoveToEmpty(i: number) {
    setLocalStarters((prev) => prev.map((p, j) => (j === i ? null : p)));
    setSelectedIndex(null);
  }

  function handleFillEmpty(i: number, player: Player) {
    setLocalStarters((prev) => prev.map((p, j) => (j === i ? player : p)));
    setSelectedIndex(null);
  }

  function handleSwapSelect(i: number, swapTarget: Player) {
    const selectedPlayer = localStarters[i];
    const targetIndex = localStarters.findIndex((p) => p?.player_id === swapTarget.player_id);

    setLocalStarters((prev) =>
      prev.map((p, j) => {
        if (j === i) return swapTarget;
        if (j === targetIndex) return selectedPlayer;
        return p;
      }),
    );
    setSelectedIndex(null);
  }

  const filledStarters = localStarters.filter(Boolean).length;

  return (
    <div className={styles.rosterCard}>
      <h2>{roster.team_name || `Team ${roster.roster_id}`}</h2>

      <RosterRarity
        players={roster.players}
        starters={localStarters.filter((p): p is Player => p !== null)}
        allScores={allScores}
      />

      <div className={styles.section}>
        <h3 className={styles.sectionLabel}>Starters · {filledStarters}/{starterSlots.length}</h3>
        <div className={styles.playerList}>
          {localStarters.map((player, i) => {
            const slot = starterSlots[i];
            const isSelected = selectedIndex === i;

            if (!player) {
              const eligible = bench.filter((b) => canFillSlot(slot, b));
              return (
                <div key={`empty-starter-${i}`} className={styles.emptyStarterRow}>
                  <div className={styles.emptyAvatar} />
                  <div style={{ flex: 1 }} />
                  <button
                    type="button"
                    className={styles.emptyBtn}
                    onClick={() => select(i)}
                  >
                    Empty
                  </button>
                  {isSelected && (
                    <div className={styles.emptyDropdown}>
                      {eligible.length > 0 ? (
                        eligible.map((b) => (
                          <button
                            key={b.player_id}
                            type="button"
                            className={styles.emptyDropdownItem}
                            onClick={() => handleFillEmpty(i, b)}
                          >
                            <img
                              src={b.image_url}
                              alt={`${b.first_name} ${b.last_name}`}
                              onError={(e) => { e.currentTarget.onerror = null; e.currentTarget.src = PROFILE_FALLBACK; }}
                            />
                            <span className={styles.emptyDropdownName}>
                              {b.first_name} {b.last_name}
                            </span>
                            <span className={styles.emptyDropdownMeta}>
                              {b.fantasy_positions[0]} · {b.team}
                            </span>
                          </button>
                        ))
                      ) : (
                        <p className={styles.emptyDropdownEmpty}>No eligible bench players</p>
                      )}
                    </div>
                  )}
                </div>
              );
            }

            const eligibleSwaps: SwapOption[] | undefined = isSelected
              ? [
                  ...localStarters
                    .filter(
                      (s, j): s is Player =>
                        s !== null &&
                        j !== i &&
                        canFillSlot(starterSlots[j], player) &&
                        canFillSlot(slot, s),
                    )
                    .map((s) => ({ player: s, isBench: false })),
                  ...bench
                    .filter((b) => canFillSlot(slot, b))
                    .map((b) => ({ player: b, isBench: true })),
                ]
              : undefined;

            return (
              <PlayerCard
                key={player.player_id}
                player={player}
                swapLabel={slotLabel(slot)}
                isSelected={isSelected}
                eligibleSwaps={eligibleSwaps}
                onSwapClick={() => select(i)}
                onSwapSelect={(opt) => handleSwapSelect(i, opt.player)}
                onMoveToEmpty={hasEmptyBench ? () => handleMoveToEmpty(i) : undefined}
              />
            );
          })}
        </div>
      </div>

      <div className={styles.section}>
        <h3 className={styles.sectionLabel}>Bench · {bench.length}/{benchSlots}</h3>
        <div className={styles.playerList}>
          {bench.map((player) => (
            <PlayerCard key={player.player_id} player={player} />
          ))}
          {Array.from({ length: Math.max(0, benchSlots - bench.length) }).map((_, i) => (
            <div key={`empty-${i}`} className={styles.emptyStarterRow}>
              <div className={styles.emptyAvatar} />
              <span className={styles.emptyLabel}>Empty</span>
            </div>
          ))}
        </div>
      </div>

      {irSlots > 0 && (
        <div className={styles.section}>
          <h3 className={styles.sectionLabel}>IR · {roster.reserve.length}/{irSlots}</h3>
          <div className={styles.playerList}>
            {roster.reserve.map((player) => (
              <PlayerCard key={player.player_id} player={player} />
            ))}
            {Array.from({ length: Math.max(0, irSlots - roster.reserve.length) }).map((_, i) => (
              // biome-ignore lint/suspicious/noArrayIndexKey: empty placeholder rows have no identity
              <div key={`ir-empty-${i}`} className={styles.emptyStarterRow}>
                <div className={styles.emptyAvatar} />
                <span className={styles.emptyLabel}>Empty</span>
              </div>
            ))}
          </div>
        </div>
      )}

      {taxiSlots > 0 && (
        <div className={styles.section}>
          <h3 className={styles.sectionLabel}>Taxi · {roster.taxi.length}/{taxiSlots}</h3>
          <div className={styles.playerList}>
            {roster.taxi.map((player) => (
              <PlayerCard key={player.player_id} player={player} />
            ))}
            {Array.from({ length: Math.max(0, taxiSlots - roster.taxi.length) }).map((_, i) => (
              // biome-ignore lint/suspicious/noArrayIndexKey: empty placeholder rows have no identity
              <div key={`taxi-empty-${i}`} className={styles.emptyStarterRow}>
                <div className={styles.emptyAvatar} />
                <span className={styles.emptyLabel}>Empty</span>
              </div>
            ))}
          </div>
        </div>
      )}

      {selectedIndex !== null && (
        <button
          type="button"
          aria-label="Close"
          className={styles.backdrop}
          onMouseDown={() => setSelectedIndex(null)}
        />
      )}
    </div>
  );
}
