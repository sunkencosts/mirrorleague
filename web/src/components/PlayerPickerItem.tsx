import type { Player } from "../types";
import PlayerCard from "./PlayerCard";
import styles from "./PlayerPickerItem.module.css";

interface Props {
	player: Player;
	onClick: () => void;
}

export default function PlayerPickerItem({ player, onClick }: Props) {
	return (
		<button type="button" className={styles.pickerItem} onClick={onClick}>
			<PlayerCard player={player} compact />
		</button>
	);
}
