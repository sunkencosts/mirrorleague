import type { Player } from '../types'

interface Props {
  player: Player
}

export default function PlayerCard({ player }: Props) {
  return (
    <div className="player-card">
      <img
        src={player.image_url}
        alt={`${player.first_name} ${player.last_name}`}
        onError={(e) => { (e.target as HTMLImageElement).style.visibility = 'hidden' }}
      />
      <div className="player-info">
        <span className="player-name">{player.first_name} {player.last_name}</span>
        <span className="player-meta">{player.fantasy_positions[0]} · {player.team}</span>
      </div>
    </div>
  )
}
