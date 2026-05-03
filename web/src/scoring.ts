import { RARITY_POINTS, type Rarity } from "./rarity";
import type { Player, Roster } from "./types";

const POSITION_MULTIPLIERS: Record<string, number> = {
	QB: 1.4,
	RB: 1.2,
	WR: 1.2,
	TE: 1.0,
	K: 0.7,
	DEF: 0.7,
};

export function computePowerScore(players: Player[], starters: Player[]): number {
	const starterIds = new Set(starters.map((p) => p.player_id));
	let total = 0;
	let baseline = 0;
	for (const p of players) {
		if (!starterIds.has(p.player_id)) {
			continue;
		}
		const primaryPos = p.fantasy_positions?.[0];
		if (!primaryPos || !(primaryPos in POSITION_MULTIPLIERS)) {
			continue;
		}
		const posMult = POSITION_MULTIPLIERS[primaryPos];
		total += RARITY_POINTS[(p.rarity || "grey") as Rarity] * posMult;
		baseline += RARITY_POINTS.purple * posMult;
	}
	if (baseline === 0) {
		return 0;
	}
	return Math.min(10, (total / baseline) * 5);
}

export type Tier = "S" | "A" | "B" | "C" | "D";

const ABSOLUTE_THRESHOLDS = { s: 9.0, a: 7.0, b: 5.0, c: 3.0 };
const RELATIVE_THRESHOLDS = { s: 9.0, a: 7.0, b: 3.0, c: 1.0 };

function getTierWithThresholds(score: number, t: typeof ABSOLUTE_THRESHOLDS): Tier {
	if (score >= t.s) {
		return "S";
	}
	if (score >= t.a) {
		return "A";
	}
	if (score >= t.b) {
		return "B";
	}
	if (score >= t.c) {
		return "C";
	}
	return "D";
}

function tierFromRange(score: number, min: number, max: number): Tier {
	return getTierWithThresholds(((score - min) / (max - min)) * 10, RELATIVE_THRESHOLDS);
}

export function getTier(score: number): Tier {
	return getTierWithThresholds(score, ABSOLUTE_THRESHOLDS);
}

export function getTierRelative(score: number, allScores: number[]): Tier {
	if (allScores.length <= 1) {
		return getTier(score);
	}
	const min = Math.min(...allScores);
	const max = Math.max(...allScores);
	if (max === min) {
		return getTier(score);
	}
	return tierFromRange(score, min, max);
}

export interface RosterScore {
	roster_id: number;
	score: number;
	tier: Tier;
}

export function scoreRosters(rosters: Roster[]): RosterScore[] {
	const scores = rosters.map((r) => computePowerScore(r.players, r.starters));
	const min = Math.min(...scores);
	const max = Math.max(...scores);
	return rosters.map((r, i) => ({
		roster_id: r.roster_id,
		score: scores[i],
		tier: min === max ? getTier(scores[i]) : tierFromRange(scores[i], min, max),
	}));
}

export const TIER_COLORS: Record<Tier, string> = {
	S: "#ffd700",
	A: "#f97316",
	B: "#a855f7",
	C: "#3b82f6",
	D: "#4b5563",
};
