package provider

type User struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

type League struct {
	LeagueID string `json:"league_id"`
}

type Player struct {
	PlayerID         string   `json:"player_id"`
	FirstName        string   `json:"first_name"`
	LastName         string   `json:"last_name"`
	Number           int      `json:"number"`
	Age              int      `json:"age"`
	Team             string   `json:"team"`
	Active           bool     `json:"active"`
	FantasyPositions []string `json:"fantasy_positions"`
	ImageURL         string   `json:"image_url"`
}

type Roster struct {
	RosterID int      `json:"roster_id"`
	OwnerID  string   `json:"owner_id"`
	TeamName string   `json:"team_name"`
	Players  []Player `json:"players"`
	Starters []Player `json:"starters"`
}
type Provider interface {
	GetRosters(leagueID string) ([]Roster, error)
}
