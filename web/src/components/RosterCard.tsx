import { useState } from "react";
import type { Roster, Player, SwapOption } from "../types";
import PlayerCard from "./PlayerCard";
import styles from "./RosterCard.module.css";

interface Props {
  roster: Roster;
}

export default function RosterCard({ roster }: Props) {
  const [localStarters, setLocalStarters] = useState<Player[]>(roster.starters);
  const [selectedStarterId, setSelectedStarterId] = useState<string | null>(
    null,
  );
  const starterIds = new Set(localStarters.map((p) => p.player_id));
  const bench = roster.players.filter((p) => !starterIds.has(p.player_id));

  function handleSwapClick(playerId: string) {
    setSelectedStarterId((prev) => (prev === playerId ? null : playerId));
  }

  function handleSwapSelect(swapTarget: Player) {
    const selectedPlayer = localStarters.find(
      (p) => p.player_id === selectedStarterId,
    );
    const targetIsStarter = localStarters.some(
      (p) => p.player_id === swapTarget.player_id,
    );

    setLocalStarters((prev) => {
      if (targetIsStarter) {
        return prev.map((p) => {
          if (p.player_id === selectedStarterId) return swapTarget;
          if (p.player_id === swapTarget.player_id) return selectedPlayer!;
          return p;
        });
      }
      return prev.map((p) =>
        p.player_id === selectedStarterId ? swapTarget : p,
      );
    });
    setSelectedStarterId(null);
  }

  return (
    <div className={styles.rosterCard}>
      <h2>{roster.team_name || `Team ${roster.roster_id}`}</h2>

      <div className={styles.section}>
        <h3 className={styles.sectionLabel}>
          Starters · {localStarters.length}
        </h3>
        <div className={styles.playerList}>
          {localStarters.map((player) => {
            const isSelected = selectedStarterId === player.player_id;
            const eligibleSwaps: SwapOption[] | undefined = isSelected
              ? [
                  ...localStarters
                    .filter(
                      (s) =>
                        s.player_id !== player.player_id &&
                        s.fantasy_positions.some((pos) =>
                          player.fantasy_positions.includes(pos),
                        ),
                    )
                    .map((s) => ({ player: s, isBench: false })),
                  ...bench
                    .filter((b) =>
                      b.fantasy_positions.some((pos) =>
                        player.fantasy_positions.includes(pos),
                      ),
                    )
                    .map((b) => ({ player: b, isBench: true })),
                ]
              : undefined;
            return (
              <PlayerCard
                key={player.player_id}
                player={player}
                isSelected={isSelected}
                eligibleSwaps={eligibleSwaps}
                onSwapClick={() => handleSwapClick(player.player_id)}
                onSwapSelect={(opt) => handleSwapSelect(opt.player)}
              />
            );
          })}
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

      {selectedStarterId && (
        <button
          type="button"
          aria-label="Close"
          className={styles.backdrop}
          onMouseDown={() => setSelectedStarterId(null)}
        />
      )}
    </div>
  );
}
