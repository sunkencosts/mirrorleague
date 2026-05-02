import { useMemo, useState } from "react";
import { useLineup } from "../hooks/useLineup";
import { canFillSlot, slotLabel } from "../slots";
import type { Player, Roster, SwapOption, Lineup } from "../types";
import PlayerCard, { PROFILE_FALLBACK } from "./PlayerCard";
import styles from "./RosterCard.module.css";
import RosterRarity from "./RosterRarity";

interface Props {
  roster: Roster;
  starterSlots: string[];
  benchSlots: number;
  irSlots: number;
  taxiSlots: number;
  allScores: number[];
  leagueId: string;
  weekNumber: number;
  userId: string;
  lineups: Lineup[];
}

interface StarterRowProps {
  slot: string;
  officialPlayer: Player | null;
  myPlayer: Player | null;
  isSelected: boolean;
  eligibleSwaps?: SwapOption[];
  bench: Player[];
  hasEmptyBench: boolean;
  onSelect: () => void;
  onFillEmpty: (player: Player) => void;
  onSwapSelect: (opt: SwapOption) => void;
  onMoveToEmpty: () => void;
}

function StarterRow({
  slot,
  officialPlayer,
  myPlayer,
  isSelected,
  eligibleSwaps,
  bench,
  hasEmptyBench,
  onSelect,
  onFillEmpty,
  onSwapSelect,
  onMoveToEmpty,
}: StarterRowProps) {
  const eligible = bench.filter((b) => canFillSlot(slot, b));

  return (
    <div className={styles.starterRow}>
      <div className={styles.officialStarters}>
        {officialPlayer ? (
          <PlayerCard player={officialPlayer} />
        ) : (
          <div className={styles.emptyStarterRow}>
            <div className={styles.emptyAvatar} />
            <span className={styles.emptyLabel}>Empty</span>
          </div>
        )}
      </div>

      <button
        type="button"
        className={`${styles.slotBadge} ${isSelected ? styles.slotBadgeSelected : ""}`}
        onClick={onSelect}
      >
        {slotLabel(slot)}
      </button>

      <div className={styles.editableCell}>
        {myPlayer ? (
          <PlayerCard
            player={myPlayer}
            isSelected={isSelected}
            eligibleSwaps={eligibleSwaps}
            onSwapSelect={onSwapSelect}
            onMoveToEmpty={hasEmptyBench ? onMoveToEmpty : undefined}
            reversed
          />
        ) : (
          <>
            <div className={`${styles.emptyStarterRow} ${styles.reversed}`}>
              <div className={styles.emptyAvatar} />
            </div>
            {isSelected && (
              <div className={styles.emptyDropdown}>
                {eligible.length > 0 ? (
                  eligible.map((b) => (
                    <button
                      key={b.player_id}
                      type="button"
                      className={styles.emptyDropdownItem}
                      onClick={() => onFillEmpty(b)}
                    >
                      <img
                        src={b.image_url}
                        alt={`${b.first_name} ${b.last_name}`}
                        onError={(e) => {
                          e.currentTarget.onerror = null;
                          e.currentTarget.src = PROFILE_FALLBACK;
                        }}
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
                  <p className={styles.emptyDropdownEmpty}>
                    No eligible bench players
                  </p>
                )}
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}

export default function RosterCard({
  roster,
  starterSlots,
  benchSlots,
  irSlots,
  taxiSlots,
  allScores,
  leagueId,
  weekNumber,
  userId,
  lineups,
}: Props) {
  const [selectedIndex, setSelectedIndex] = useState<number | null>(null);
  const existingLineup =
    lineups.find((l) => l.roster_id === roster.roster_id) ?? null;

  const { localStarters, setLocalStarters, saveStatus, saveStarters } =
    useLineup({
      userId,
      leagueId,
      rosterId: roster.roster_id,
      weekNumber,
      players: roster.players,
      initialStarters: roster.starters,
      slotCount: starterSlots.length,
      existingLineup,
    });

  const bench = useMemo(() => {
    const starterIds = new Set(localStarters.flatMap((p) => (p ? [p.player_id] : [])));
    return roster.players.filter((p) => !starterIds.has(p.player_id));
  }, [localStarters, roster.players]);
  const hasEmptyBench = bench.length < benchSlots;

  const eligibleSwaps = useMemo<SwapOption[] | null>(() => {
    if (selectedIndex === null) return null;
    const myPlayer = localStarters[selectedIndex];
    const slot = starterSlots[selectedIndex];
    if (!myPlayer) return null;
    return [
      ...localStarters
        .filter(
          (s, j): s is Player =>
            s !== null &&
            j !== selectedIndex &&
            canFillSlot(starterSlots[j], myPlayer) &&
            canFillSlot(slot, s),
        )
        .map((s) => ({ player: s, isBench: false })),
      ...bench
        .filter((b) => canFillSlot(slot, b))
        .map((b) => ({ player: b, isBench: true })),
    ];
  }, [selectedIndex, localStarters, starterSlots, bench]);

  function select(i: number) {
    setSelectedIndex((prev) => (prev === i ? null : i));
  }

  function handleMoveToEmpty(i: number) {
    const next = localStarters.map((p, j) => (j === i ? null : p));
    setLocalStarters(next);
    setSelectedIndex(null);
    saveStarters(next);
  }

  function handleFillEmpty(i: number, player: Player) {
    const next = localStarters.map((p, j) => (j === i ? player : p));
    setLocalStarters(next);
    setSelectedIndex(null);
    saveStarters(next);
  }

  function handleSwapSelect(i: number, swapTarget: Player) {
    const displaced = localStarters[i];
    const next = localStarters.map((p, j) => {
      if (j === i) return swapTarget;
      if (p?.player_id === swapTarget.player_id) return displaced;
      return p;
    });
    setLocalStarters(next);
    setSelectedIndex(null);
    saveStarters(next);
  }

  const filledStarters = localStarters.filter(Boolean).length;

  return (
    <div className={styles.rosterCard}>
      <div className={styles.cardHeader}>
        <h2>{roster.team_name || `Team ${roster.roster_id}`}</h2>
      </div>

      <RosterRarity
        players={roster.players}
        starters={roster.starters}
        allScores={allScores}
      />

      <div className={styles.section}>
        <div className={styles.starterRow}>
          <h3 className={styles.sectionLabel}>Official Starters</h3>
          <span />
          <h3 className={styles.sectionLabel}>
            Your Picks · {filledStarters}/{starterSlots.length}
          </h3>
        </div>
        <div className={styles.starterGrid}>
          {starterSlots.map((slot, i) => {
            return (
              // biome-ignore lint/suspicious/noArrayIndexKey: slot rows are positional by design
              <StarterRow
                key={`starter-row-${i}`}
                slot={slot}
                officialPlayer={roster.starters[i] ?? null}
                myPlayer={localStarters[i] ?? null}
                isSelected={selectedIndex === i}
                eligibleSwaps={
                  selectedIndex === i ? (eligibleSwaps ?? undefined) : undefined
                }
                bench={bench}
                hasEmptyBench={hasEmptyBench}
                onSelect={() => select(i)}
                onFillEmpty={(player) => handleFillEmpty(i, player)}
                onSwapSelect={(opt) => handleSwapSelect(i, opt.player)}
                onMoveToEmpty={() => handleMoveToEmpty(i)}
              />
            );
          })}
        </div>
      </div>

      <div className={styles.section}>
        <h3 className={styles.sectionLabel}>
          Bench · {bench.length}/{benchSlots}
        </h3>
        <div className={styles.playerList}>
          {bench.map((player) => (
            <PlayerCard key={player.player_id} player={player} />
          ))}
          {Array.from({ length: Math.max(0, benchSlots - bench.length) }).map(
            (_, i) => (
              // biome-ignore lint/suspicious/noArrayIndexKey: empty placeholder rows have no identity
              <div key={`empty-${i}`} className={styles.emptyStarterRow}>
                <div className={styles.emptyAvatar} />
                <span className={styles.emptyLabel}>Empty</span>
              </div>
            ),
          )}
        </div>
      </div>

      {irSlots > 0 && (
        <div className={styles.section}>
          <h3 className={styles.sectionLabel}>
            IR · {roster.reserve.length}/{irSlots}
          </h3>
          <div className={styles.playerList}>
            {roster.reserve.map((player) => (
              <PlayerCard key={player.player_id} player={player} />
            ))}
          </div>
        </div>
      )}

      {taxiSlots > 0 && (
        <div className={styles.section}>
          <h3 className={styles.sectionLabel}>
            Taxi · {roster.taxi.length}/{taxiSlots}
          </h3>
          <div className={styles.playerList}>
            {roster.taxi.map((player) => (
              <PlayerCard key={player.player_id} player={player} />
            ))}
          </div>
        </div>
      )}

      {saveStatus !== "idle" && (
        <div className={styles.saveFooter}>
          {saveStatus === "saving" && <span className={styles.statusSaved}>Saving…</span>}
          {saveStatus === "saved" && <span className={styles.statusSaved}>Saved</span>}
          {saveStatus === "error" && <span className={styles.statusError}>Save failed</span>}
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
