export interface Player {
  player_id: string;
  first_name: string;
  last_name: string;
  number: number;
  age: number;
  team: string;
  active: boolean;
  fantasy_positions: string[];
  image_url: string;
  rarity?: import("./rarity").Rarity;
}

export interface SwapOption {
  player: Player;
  isBench: boolean;
}

export interface Roster {
  roster_id: number;
  owner_id: string;
  team_name: string;
  players: Player[];
  starters: Player[];
  reserve: Player[];
  taxi: Player[];
}

export interface League {
  roster_positions: string[];
  settings: {
    reserve_slots: number;
    taxi_slots: number;
  };
}
