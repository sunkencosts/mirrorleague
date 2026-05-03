import { useMemo } from "react";
import { RARITY_COLORS, RARITY_LABELS, RARITY_ORDER } from "../rarity";
import { computePowerScore, getTierRelative, TIER_COLORS } from "../scoring";
import type { Player } from "../types";
import styles from "./RosterRarity.module.css";

interface Props {
	players: Player[];
	starters: Player[];
	allScores: number[];
}

export default function RosterRarity({ players, starters, allScores }: Props) {
	const rarityCounts = useMemo(
		() =>
			players.reduce<Record<string, number>>((acc, p) => {
				const key = p.rarity ?? "grey";
				acc[key] = (acc[key] ?? 0) + 1;
				return acc;
			}, {}),
		[players],
	);

	const score = useMemo(() => computePowerScore(players, starters), [players, starters]);

	const tier = useMemo(() => getTierRelative(score, allScores), [score, allScores]);
	const tierColor = TIER_COLORS[tier];
	const present = RARITY_ORDER.filter((r) => rarityCounts[r] > 0);

	return (
		<div className={styles.rarityContainer}>
			<div className={styles.header}>
				<span className={styles.tierBadge} style={{ color: tierColor, borderColor: tierColor }}>
					{tier}-TIER
				</span>
				<span className={styles.powerScore}>
					<span className={styles.powerLabel}>ROSTER POWER</span>
					<span className={styles.powerValue}>{score.toFixed(1)}/10</span>
				</span>
			</div>

			<div className={styles.rarityBar}>
				{present.map((r) => (
					<div
						key={r}
						className={styles.raritySegment}
						style={{ background: RARITY_COLORS[r], flex: rarityCounts[r] }}
						title={`${RARITY_LABELS[r]}: ${rarityCounts[r]}`}
					>
						{rarityCounts[r]}
					</div>
				))}
			</div>

			<div className={styles.legend}>
				{present.map((r) => (
					<span key={r} className={styles.legendItem}>
						<span className={styles.legendSwatch} style={{ background: RARITY_COLORS[r] }} />
						{RARITY_LABELS[r]}
					</span>
				))}
			</div>
		</div>
	);
}
