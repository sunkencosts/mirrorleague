import { useMemo, useRef, useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { Lineup, Player } from "../types";

interface Params {
  userId: string;
  leagueId: string;
  rosterId: number;
  weekNumber: number;
  players: Player[];
  initialStarters: Player[];
  slotCount: number;
  existingLineup: Lineup | null;
}

export function useLineup({
  userId,
  leagueId,
  rosterId,
  weekNumber,
  players,
  initialStarters,
  slotCount,
  existingLineup,
}: Params) {
  const queryKey = ["lineups", userId, leagueId, weekNumber];

  const playerMap = useMemo(
    () => new Map(players.map((p) => [p.player_id, p])),
    [players],
  );

  const [localStarters, setLocalStarters] = useState<(Player | null)[]>(() => {
    if (existingLineup) {
      return existingLineup.starters.map((id) => playerMap.get(id) ?? null);
    }
    return Array.from({ length: slotCount }, (_, i) => initialStarters[i] ?? null);
  });

  const lineupIdRef = useRef<string | null>(existingLineup?.id ?? null);

  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (starters: string[]) => {
      if (lineupIdRef.current === null) {
        return fetch("/api/lineups", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            user_id: userId,
            league_id: leagueId,
            source: "sleeper",
            roster_id: rosterId,
            week_number: weekNumber,
            starters,
          }),
        }).then((r) => {
          if (!r.ok) throw new Error(`${r.status}`);
          return r.json();
        });
      }
      return fetch(`/api/lineups/${lineupIdRef.current}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ user_id: userId, starters }),
      }).then((r) => {
        if (!r.ok) throw new Error(`${r.status}`);
        return r.json();
      });
    },
    onSuccess: (data) => {
      if (lineupIdRef.current === null) lineupIdRef.current = data.id;
      queryClient.invalidateQueries({ queryKey });
    },
  });

  return {
    localStarters,
    setLocalStarters,
    saveStatus: mutation.isPending
      ? "saving"
      : mutation.isError
        ? "error"
        : mutation.isSuccess
          ? "saved"
          : "idle",
    saveStarters: (next: (Player | null)[]) => {
      if (next.some((p) => p === null)) return;
      mutation.mutate((next as Player[]).map((p) => p.player_id));
    },
  } as const;
}
