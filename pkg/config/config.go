package config

type Config struct {
	Port           string
	SleeperBaseURL string
	RankingsCSVURL string
}

func Load(getenv func(string) string) Config {
	port := getenv("PORT")
	if port == "" {
		port = "8080"
	}
	sleeperBaseURL := getenv("SLEEPER_BASE_URL")
	if sleeperBaseURL == "" {
		sleeperBaseURL = "https://api.sleeper.app/v1"
	}
	rankingsCSVURL := getenv("RANKINGS_CSV_URL")
	if rankingsCSVURL == "" {
		rankingsCSVURL = "https://raw.githubusercontent.com/dynastyprocess/data/master/files/db_fpecr_latest.csv"
	}
	return Config{
		Port:           port,
		SleeperBaseURL: sleeperBaseURL,
		RankingsCSVURL: rankingsCSVURL,
	}
}
