package provider

type User struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

type League struct {
	LeagueID string `json:"league_id"`
}

type Roster struct {
	RosterID int      `json:"roster_id"`
	OwnerID  string   `json:"owner_id"`
	Players  []string `json:"players"`
	Starters []string `json:"starters"`
}

type Provider interface {
	GetRosters(leagueID string) ([]Roster, error)
}
