import { useMemo, useState } from "react";
import { useLineup } from "../hooks/useLineup";
import { RARITY_GLOW } from "../rarity";
import { canFillSlot, slotLabel } from "../slots";
import type { Lineup, Player, Roster, WeekMatchup } from "../types";
import PlayerCard, { onImageError } from "./PlayerCard";
import styles from "./RosterCard.module.css";
import RosterRarity from "./RosterRarity";

interface Props {
	roster: Roster;
	weekMatchup?: WeekMatchup | null;
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
	overridePlayer: Player | null;
	isPickerOpen: boolean;
	eligiblePicks: Player[];
	onTogglePicker: () => void;
	onPickOverride: (player: Player) => void;
	onClearOverride: () => void;
	officialPoints?: number;
	overridePoints?: number;
}

function OverrideChip({
	player,
	onClear,
	points,
}: { player: Player; onClear: () => void; points?: number }) {
	return (
		<button
			type="button"
			className={`${styles.overrideChip} ${styles.overrideChipBtn}`}
			onClick={onClear}
			title="Click to remove override"
		>
			<img
				style={{ boxShadow: RARITY_GLOW[player.rarity || "grey"] }}
				src={player.image_url}
				alt={`${player.first_name} ${player.last_name}`}
				onError={onImageError}
			/>
			<div className={styles.pickInfo}>
				<span className={styles.pickName}>
					{player.first_name} {player.last_name}
				</span>
				<span className={styles.pickMeta}>
					{player.fantasy_positions[0]} · {player.team}
				</span>
				{typeof points === "number" && (
					<span className={`${styles.playerPoints} ${styles.playerPointsRight}`}>
						{points.toFixed(1)} pts
					</span>
				)}
			</div>
		</button>
	);
}

function StarterRow({
	slot,
	officialPlayer,
	overridePlayer,
	isPickerOpen,
	eligiblePicks,
	onTogglePicker,
	onPickOverride,
	onClearOverride,
	officialPoints,
	overridePoints,
}: StarterRowProps) {
	const isOverridden = overridePlayer !== null;

	return (
		<div className={styles.starterRow}>
			{/* LEFT: official starter */}
			<div className={`${styles.officialCell} ${isOverridden ? styles.officialDimmed : ""}`}>
				{officialPlayer ? (
					<>
						<img
							style={{
								boxShadow: isOverridden
									? "0 0 0 1px #4b5563"
									: RARITY_GLOW[officialPlayer.rarity || "grey"],
							}}
							src={officialPlayer.image_url}
							alt={`${officialPlayer.first_name} ${officialPlayer.last_name}`}
							onError={onImageError}
						/>
						<div className={styles.playerInfo}>
							<span
								className={`${styles.playerName} ${isOverridden ? styles.playerNameStruck : ""}`}
							>
								{officialPlayer.first_name} {officialPlayer.last_name}
							</span>
							<span className={styles.playerMeta}>
								{officialPlayer.fantasy_positions[0]} · {officialPlayer.team}
							</span>
							{typeof officialPoints === "number" && (
								<span className={styles.playerPoints}>{officialPoints.toFixed(1)} pts</span>
							)}
						</div>
					</>
				) : (
					<div className={styles.emptySlot}>
						<div className={styles.emptyAvatar} />
						<span className={styles.emptyLabel}>Empty</span>
					</div>
				)}
			</div>

			{/* CENTER: slot pill */}
			<div className={styles.slotCell}>
				<span className={`${styles.slotPill} ${isOverridden ? styles.slotPillOverridden : ""}`}>
					{slotLabel(slot)}
				</span>
				{isOverridden && <span className={styles.slotArrow}> →</span>}
			</div>

			{/* RIGHT: override chip or CTA */}
			<div className={styles.pickCell}>
				{overridePlayer !== null ? (
					<OverrideChip player={overridePlayer} onClear={onClearOverride} points={overridePoints} />
				) : (
					<div className={styles.overrideCta}>
						<button type="button" className={styles.overrideBtn} onClick={onTogglePicker}>
							+ Override
						</button>
						{isPickerOpen && (
							<div className={styles.pickerDropdown}>
								{eligiblePicks.length > 0 ? (
									eligiblePicks.map((p) => (
										<button
											key={p.player_id}
											type="button"
											className={styles.pickerItem}
											onClick={() => onPickOverride(p)}
										>
											<img
												src={p.image_url}
												alt={`${p.first_name} ${p.last_name}`}
												onError={onImageError}
											/>
											<span className={styles.pickerName}>
												{p.first_name} {p.last_name}
											</span>
											<span className={styles.pickerMeta}>
												{p.fantasy_positions[0]} · {p.team}
											</span>
										</button>
									))
								) : (
									<p className={styles.pickerEmpty}>No eligible bench players</p>
								)}
							</div>
						)}
					</div>
				)}
			</div>
		</div>
	);
}

export default function RosterCard({
	roster,
	weekMatchup,
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

	const activePlayers = weekMatchup?.players ?? roster.players;
	const activeStarters = weekMatchup?.starters ?? roster.starters;

	const existingLineup = useMemo(
		() => lineups.find((l) => l.roster_id === roster.roster_id) ?? null,
		[lineups, roster.roster_id],
	);

	const playerPoints = useMemo(
		() => weekMatchup?.player_points ?? {},
		[weekMatchup],
	);
	const weekHasScoring = (weekMatchup?.points ?? 0) > 0;

	const userTotal = useMemo(() => {
		if (!existingLineup || !weekHasScoring) return null;
		return existingLineup.starters.reduce((sum, id) => sum + (playerPoints[id] ?? 0), 0);
	}, [existingLineup, playerPoints, weekHasScoring]);

	const winner = useMemo((): "user" | "official" | "tie" | null => {
		if (userTotal === null || !weekMatchup) return null;
		if (userTotal > weekMatchup.points) return "user";
		if (weekMatchup.points > userTotal) return "official";
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

	function handleTogglePicker(i: number) {
		if (saveStatus === "saving") {
			return;
		}
		setSelectedIndex((prev) => (prev === i ? null : i));
	}

	function handlePickOverride(i: number, player: Player) {
		if (saveStatus === "saving") {
			return;
		}
		applyOverride(i, player);
		setSelectedIndex(null);
	}

	function handleClearOverride(i: number) {
		applyOverride(i, null);
	}

	// Stable keys derived outside the map so biome's noArrayIndexKey rule doesn't fire.
	// Slot names can repeat (e.g. two FLEX slots), so we include the position.
	const slotKeys = useMemo(() => starterSlots.map((slot, i) => `${slot}-${i}`), [starterSlots]);

	return (
		<div className={styles.rosterCard}>
			<div className={styles.teamHeader}>
				<h2>{roster.team_name || `Team ${roster.roster_id}`}</h2>
				<div className={styles.headerScores}>
					{weekMatchup && (
						<span className={styles.officialScore}>
							Official: {(weekMatchup.custom_points ?? weekMatchup.points).toFixed(2)}
						</span>
					)}
					{userTotal !== null && (
						<span className={styles.userScore}>You: {userTotal.toFixed(2)}</span>
					)}
					{winner && (
						<span className={`${styles.winnerBadge} ${styles[winner]}`}>
							{winner === "user" ? "You Win" : winner === "official" ? "You Lose" : "Tie"}
						</span>
					)}
				</div>
			</div>

			<RosterRarity players={activePlayers} starters={activeStarters} allScores={allScores} />

			<div className={styles.section}>
				<div className={styles.columnHeaders}>
					<span className={styles.colHeaderOfficial}>Official</span>
					<span />
					<span className={styles.colHeaderPick}>Your Pick</span>
				</div>

				<div className={styles.starterGrid}>
					{starterSlots.map((slot, i) => {
						const official = activeStarters[i] ?? null;
						const overridePlayer = overrides[i] ?? null;
						return (
							<StarterRow
								key={slotKeys[i]}
								slot={slot}
								officialPlayer={official}
								overridePlayer={overridePlayer}
								isPickerOpen={selectedIndex === i}
								eligiblePicks={eligiblePicksBySlot[i]}
								onTogglePicker={() => handleTogglePicker(i)}
								onPickOverride={(player) => handlePickOverride(i, player)}
								onClearOverride={() => handleClearOverride(i)}
								officialPoints={weekHasScoring && official ? (playerPoints[official.player_id] ?? 0) : undefined}
								overridePoints={weekHasScoring && overridePlayer ? (playerPoints[overridePlayer.player_id] ?? 0) : undefined}
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
					{Array.from({ length: Math.max(0, benchSlots - bench.length) }).map((_, i) => (
						// biome-ignore lint/suspicious/noArrayIndexKey: empty placeholder rows have no identity
						<div key={`empty-${i}`} className={styles.emptyStarterRow}>
							<div className={styles.emptyAvatar} />
							<span className={styles.emptyLabel}>Empty</span>
						</div>
					))}
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
