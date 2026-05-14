import { memo } from "react";
import { RARITY_GLOW } from "../rarity";
import type { Player } from "../types";
import { onImageError } from "../utils/playerImage";
import styles from "./PlayerCard.module.css";

interface Props {
	player: Player;
	reversed?: boolean;
	points?: number;
	dimmed?: boolean;
	compact?: boolean;
}

function PlayerCard({ player, reversed, points, dimmed, compact }: Props) {
	const cardClass = [
		styles.playerCard,
		reversed && styles.reversed,
		dimmed && styles.dimmed,
		compact && styles.compact,
	]
		.filter(Boolean)
		.join(" ");

	return (
		<div className={cardClass}>
			<img
				style={{
					boxShadow:
						dimmed ? "0 0 0 1px var(--border-ui)" : RARITY_GLOW[player.rarity || "grey"],
				}}
				src={player.image_url}
				alt={`${player.first_name} ${player.last_name}`}
				width={42}
				height={40}
				onError={onImageError}
			/>
			<div className={styles.playerInfo}>
				<span
					className={[styles.playerName, dimmed && styles.nameStruck].filter(Boolean).join(" ")}
				>
					{player.first_name} {player.last_name}
				</span>
				<span className={styles.playerMeta}>
					{player.fantasy_positions[0]} · {player.team}
				</span>
				{typeof points === "number" && (
					<span className={styles.points}>{points.toFixed(1)} pts</span>
				)}
			</div>
		</div>
	);
}

export default memo(PlayerCard);
