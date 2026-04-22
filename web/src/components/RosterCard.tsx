import type { Roster } from '../types'
import PlayerCard from './PlayerCard'

interface Props {
  roster: Roster
}

export default function RosterCard({ roster }: Props) {
  return (
    <div className="roster-card">
      <h2>{roster.team_name || `Team ${roster.roster_id}`}</h2>
      <div className="player-list">
        {roster.starters.map((player) => (
          <PlayerCard key={player.player_id} player={player} />
        ))}
      </div>
    </div>
  )
}
