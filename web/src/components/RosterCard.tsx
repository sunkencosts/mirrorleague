import { useRosterCard } from "../hooks/useRosterCard";
import { slotLabel } from "../slots";
import type { Lineup, Player, Roster, WeekMatchup } from "../types";
import PlayerCard from "./PlayerCard";
import PlayerPickerItem from "./PlayerPickerItem";
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
			<div className={styles.officialCell}>
				{officialPlayer ? (
					<PlayerCard player={officialPlayer} points={officialPoints} dimmed={isOverridden} />
				) : (
					<div className={styles.emptySlot}>
						<div className={styles.emptyAvatar} />
						<span className={styles.emptyLabel}>Empty</span>
					</div>
				)}
			</div>

			<div className={styles.slotCell}>
				<span className={`${styles.slotPill} ${isOverridden ? styles.slotPillOverridden : ""}`}>
					{slotLabel(slot)}
				</span>
				{isOverridden && <span className={styles.slotArrow}> →</span>}
			</div>

			<div className={styles.pickCell}>
				{overridePlayer !== null ? (
					<button
						type="button"
						className={styles.overrideChipBtn}
						onClick={onClearOverride}
						title="Click to remove override"
					>
						<PlayerCard player={overridePlayer} reversed points={overridePoints} />
					</button>
				) : (
					<div className={styles.overrideCta}>
						<button type="button" className={styles.overrideBtn} onClick={onTogglePicker}>
							+ Override
						</button>
						{isPickerOpen && (
							<div className={styles.pickerDropdown}>
								{eligiblePicks.length > 0 ? (
									eligiblePicks.map((p) => (
										<PlayerPickerItem
											key={p.player_id}
											player={p}
											onClick={() => onPickOverride(p)}
										/>
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
	const {
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
	} = useRosterCard({ roster, weekMatchup, starterSlots, lineups, userId, leagueId, weekNumber });

	function getPoints(playerId: string) {
		return weekHasScoring ? (playerPoints[playerId] ?? 0) : undefined;
	}

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
							{{ user: "You Win", official: "You Lose", tie: "Tie" }[winner]}
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
								officialPoints={official ? getPoints(official.player_id) : undefined}
								overridePoints={overridePlayer ? getPoints(overridePlayer.player_id) : undefined}
							/>
						);
					})}
				</div>
			</div>

			<div className={styles.section}>
				<h3 className={styles.sectionLabel}>Bench · {bench.length}/{benchSlots}</h3>
				<div className={styles.playerList}>
					{bench.map((player) => (
						<PlayerCard
							key={player.player_id}
							player={player}
							points={getPoints(player.player_id)}
						/>
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
					<h3 className={styles.sectionLabel}>IR · {roster.reserve.length}/{irSlots}</h3>
					<div className={styles.playerList}>
						{roster.reserve.map((player) => (
							<PlayerCard
								key={player.player_id}
								player={player}
								points={getPoints(player.player_id)}
							/>
						))}
					</div>
				</div>
			)}

			{taxiSlots > 0 && (
				<div className={styles.section}>
					<h3 className={styles.sectionLabel}>Taxi · {roster.taxi.length}/{taxiSlots}</h3>
					<div className={styles.playerList}>
						{roster.taxi.map((player) => (
							<PlayerCard
								key={player.player_id}
								player={player}
								points={getPoints(player.player_id)}
							/>
						))}
					</div>
				</div>
			)}

			{selectedIndex !== null && (
				<button
					type="button"
					aria-label="Close"
					className={styles.backdrop}
					onMouseDown={handleCloseAllPickers}
				/>
			)}
		</div>
	);
}
