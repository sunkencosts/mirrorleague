import { useMemo, useState } from "react";
import { useAuth } from "../context/AuthContext";
import { canFillSlot } from "../slots";
import type { Lineup, Player, Roster, WeekMatchup } from "../types";
import { useLineup } from "./useLineup";

const POSITION_ORDER: Record<string, number> = { QB: 0, RB: 1, WR: 2, TE: 3, K: 4, DEF: 5, DST: 5 };
const EMPTY_PLAYERS: Player[] = [];

interface Params {
	roster: Roster;
	weekMatchup?: WeekMatchup | null;
	starterSlots: string[];
	lineups: Lineup[];
	userId: string;
	leagueId: string;
	weekNumber: number;
	currentWeek: number;
}

export function useRosterCard({
	roster,
	weekMatchup,
	starterSlots,
	lineups,
	userId,
	leagueId,
	weekNumber,
	currentWeek,
}: Params) {
	const [selectedIndex, setSelectedIndex] = useState<number | null>(null);
	const [showAuthPrompt, setShowAuthPrompt] = useState(false);
	const { user } = useAuth();
	const isAuthenticated = user !== null;

	const activePlayers = weekMatchup?.players ?? roster.players;
	const activeStarters = weekMatchup?.starters ?? roster.starters;

	// Sleeper provides no historical IR/taxi data — only show them for the current week.
	const isCurrentWeek = weekNumber === currentWeek;
	const activeReserve = isCurrentWeek ? roster.reserve : EMPTY_PLAYERS;
	const activeTaxi = isCurrentWeek ? roster.taxi : EMPTY_PLAYERS;

	const existingLineup = useMemo(
		() => lineups.find((l) => l.roster_id === roster.roster_id) ?? null,
		[lineups, roster.roster_id],
	);

	const playerPoints = useMemo(() => weekMatchup?.player_points ?? {}, [weekMatchup]);
	const weekHasScoring = (weekMatchup?.points ?? 0) > 0;

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

	const hasOverrides = useMemo(() => Object.values(overrides).some(Boolean), [overrides]);

	const officialPoints = useMemo(
		() => (weekMatchup ? (weekMatchup.custom_points ?? weekMatchup.points) : null),
		[weekMatchup],
	);

	const userTotal = useMemo(() => {
		if (!weekHasScoring || !hasOverrides) {
			return null;
		}
		return activeStarters.reduce((sum, player, i) => {
			const effective = overrides[i] ?? player;
			return sum + (effective ? (playerPoints[effective.player_id] ?? 0) : 0);
		}, 0);
	}, [weekHasScoring, hasOverrides, activeStarters, overrides, playerPoints]);

	const winner = useMemo((): "user" | "official" | "tie" | null => {
		if (userTotal === null || officialPoints === null) {
			return null;
		}
		if (userTotal > officialPoints) {
			return "user";
		}
		if (officialPoints > userTotal) {
			return "official";
		}
		return "tie";
	}, [userTotal, officialPoints]);

	const diff = useMemo(
		() => (officialPoints !== null && userTotal !== null ? userTotal - officialPoints : null),
		[officialPoints, userTotal],
	);

	const bench = useMemo(() => {
		const officialIds = new Set(activeStarters.map((p) => p.player_id));
		const usedIds = new Set(Object.values(overrides).map((p) => p.player_id));
		const taxiIds = new Set(activeTaxi.map((p) => p.player_id));
		const reserveIds = new Set(activeReserve.map((p) => p.player_id));
		return activePlayers
			.filter((p) => !officialIds.has(p.player_id) && !usedIds.has(p.player_id) && !taxiIds.has(p.player_id) && !reserveIds.has(p.player_id))
			.sort((a, b) => {
				const aPos = POSITION_ORDER[a.fantasy_positions[0]] ?? 99;
				const bPos = POSITION_ORDER[b.fantasy_positions[0]] ?? 99;
				return aPos - bPos;
			});
	}, [activePlayers, activeReserve, activeTaxi, activeStarters, overrides]);

	const eligiblePicksBySlot = useMemo(
		() => starterSlots.map((slot) => bench.filter((p) => canFillSlot(slot, p))),
		[bench, starterSlots],
	);

	const slotKeys = useMemo(() => starterSlots.map((slot, i) => `${slot}-${i}`), [starterSlots]);

	const isSaving = saveStatus === "saving";

	function handleTogglePicker(i: number) {
		if (isSaving) {
			return;
		}
		if (!isAuthenticated) {
			setShowAuthPrompt(true);
			return;
		}
		setShowAuthPrompt(false);
		setSelectedIndex((prev) => (prev === i ? null : i));
	}

	function handlePickOverride(i: number, player: Player) {
		if (isSaving) {
			return;
		}
		applyOverride(i, player);
		setSelectedIndex(null);
	}

	function handleClearOverride(i: number) {
		if (isSaving) {
			return;
		}
		applyOverride(i, null);
	}

	function handleCloseAllPickers() {
		setSelectedIndex(null);
	}

	return {
		activePlayers,
		activeStarters,
		activeReserve,
		activeTaxi,
		playerPoints,
		weekHasScoring,
		officialPoints,
		userTotal,
		diff,
		winner,
		hasOverrides,
		overrides,
		bench,
		eligiblePicksBySlot,
		slotKeys,
		selectedIndex,
		showAuthPrompt,
		handleTogglePicker,
		handlePickOverride,
		handleClearOverride,
		handleCloseAllPickers,
	};
}
