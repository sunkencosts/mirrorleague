import type { League } from "../types";
import styles from "./LeagueSummary.module.css";
import WeekSelector from "./WeekSelector";

interface Props {
	league: League;
	weekNumber: number;
	onWeekChange: (week: number) => void;
}

const pprLabel = (rec: number) => {
	if (rec === 1) {
		return "FULL PPR";
	}
	if (rec === 0.5) {
		return "HALF PPR";
	}
	return null;
};

export default function LeagueSummary({
	league: {
		name,
		roster_positions,
		scoring_settings,
		settings: { num_teams },
	},
	weekNumber,
	onWeekChange,
}: Props) {
	const ppr = pprLabel(scoring_settings.rec);
	const tep = scoring_settings.bonus_rec_te > 0;
	const superflex = roster_positions.includes("SUPER_FLEX");

	return (
		<div className={styles.container}>
			<div className={styles.header}>
				<span className={styles.leagueName}>{name}</span>
				<WeekSelector weekNumber={weekNumber} onWeekChange={onWeekChange} />
			</div>
			<div className={styles.badges}>
				<span className={styles.badge} style={{ color: "#94a3b8", borderColor: "#94a3b8" }}>
					{num_teams} TEAMS
				</span>
				{ppr && (
					<span className={styles.badge} style={{ color: "#4ade80", borderColor: "#4ade80" }}>
						{ppr}
					</span>
				)}
				{tep && (
					<span className={styles.badge} style={{ color: "#fbbf24", borderColor: "#fbbf24" }}>
						TEP +{scoring_settings.bonus_rec_te}
					</span>
				)}
				{superflex && (
					<span className={styles.badge} style={{ color: "#a78bfa", borderColor: "#a78bfa" }}>
						SUPERFLEX
					</span>
				)}
			</div>
		</div>
	);
}
