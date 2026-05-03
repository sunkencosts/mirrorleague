import styles from "./WeekSelector.module.css";

interface Props {
  weekNumber: number;
  onWeekChange: (week: number) => void;
  min?: number;
  max?: number;
}

export default function WeekSelector({
  weekNumber,
  onWeekChange,
  min = 1,
  max = 18,
}: Props) {
  return (
    <div className={styles.container}>
      <button
        type="button"
        className={styles.chevron}
        disabled={weekNumber <= min}
        onClick={() => onWeekChange(Math.max(min, weekNumber - 1))}
      >
        &lsaquo;
      </button>
      <span className={styles.label}>WEEK {weekNumber}</span>
      <button
        type="button"
        className={styles.chevron}
        disabled={weekNumber >= max}
        onClick={() => onWeekChange(Math.min(max, weekNumber + 1))}
      >
        &rsaquo;
      </button>
    </div>
  );
}
