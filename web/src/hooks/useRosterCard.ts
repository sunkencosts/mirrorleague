import { useMemo, useState } from "react";
import { canFillSlot } from "../slots";
import type { Lineup, Player, Roster, WeekMatchup } from "../types";
import { useLineup } from "./useLineup";

interface Params {
	roster: Roster;
	weekMatchup?: WeekMatchup | null;
	starterSlots: string[];
	lineups: Lineup[];
	userId: string;
	leagueId: string;
	weekNumber: number;
}

export function useRosterCard({
	roster,
	weekMatchup,
	starterSlots,
	lineups,
	userId,
	leagueId,
	weekNumber,
}: Params) {
	const [selectedIndex, setSelectedIndex] = useState<number | null>(null);

	const activePlayers = weekMatchup?.players ?? roster.players;
	const activeStarters = weekMatchup?.starters ?? roster.starters;

	const existingLineup = useMemo(
		() => lineups.find((l) => l.roster_id === roster.roster_id) ?? null,
		[lineups, roster.roster_id],
	);

	const playerPoints = useMemo(() => weekMatchup?.player_points ?? {}, [weekMatchup]);
	const weekHasScoring = (weekMatchup?.points ?? 0) > 0;

	const userTotal = useMemo(() => {
		if (!existingLineup || !weekHasScoring) return null;
		return existingLineup.starters.reduce((sum, id) => sum + (playerPoints[id] ?? 0), 0);
	}, [existingLineup, playerPoints, weekHasScoring]);

	const winner = useMemo((): "user" | "official" | "tie" | null => {
		if (userTotal === null || !weekMatchup) return null;
		const officialPoints = weekMatchup.custom_points ?? weekMatchup.points;
		if (userTotal > officialPoints) return "user";
		if (officialPoints > userTotal) return "official";
		return "tie";
	}, [userTotal, weekMatchup]);

	const { overrides, applyOverride, saveStatus } = useLineup({
		userId,
		leagueId,
		rosterId: roster.roster_id,
		weekNumber,
		players: activePlayers,
		initialStarters: activeStarters,
		slotCount: starterSlots.length,
		existingLineup,
	});

	const bench = useMemo(() => {
		const officialIds = new Set(activeStarters.map((p) => p.player_id));
		const usedIds = new Set(Object.values(overrides).map((p) => p.player_id));
		return activePlayers.filter((p) => !officialIds.has(p.player_id) && !usedIds.has(p.player_id));
	}, [activePlayers, activeStarters, overrides]);

	const eligiblePicksBySlot = useMemo(
		() => starterSlots.map((slot) => bench.filter((p) => canFillSlot(slot, p))),
		[bench, starterSlots],
	);

	const slotKeys = useMemo(
		() => starterSlots.map((slot, i) => `${slot}-${i}`),
		[starterSlots],
	);

	const isSaving = saveStatus === "saving";

	function handleTogglePicker(i: number) {
		if (isSaving) return;
		setSelectedIndex((prev) => (prev === i ? null : i));
	}

	function handlePickOverride(i: number, player: Player) {
		if (isSaving) return;
		applyOverride(i, player);
		setSelectedIndex(null);
	}

	function handleClearOverride(i: number) {
		if (isSaving) return;
		applyOverride(i, null);
	}

	function handleCloseAllPickers() {
		setSelectedIndex(null);
	}

	return {
		activePlayers,
		activeStarters,
		playerPoints,
		weekHasScoring,
		userTotal,
		winner,
		overrides,
		bench,
		eligiblePicksBySlot,
		slotKeys,
		selectedIndex,
		handleTogglePicker,
		handlePickOverride,
		handleClearOverride,
		handleCloseAllPickers,
	};
}
