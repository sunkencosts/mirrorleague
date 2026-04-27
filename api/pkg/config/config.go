package config

type Config struct {
	Port           string
	SleeperBaseURL string
	RankingsCSVURL string
	DatabaseURL    string
	MigrationsURL  string
}

func Load(getenv func(string) string) Config {
	port := getenv("PORT")
	if port == "" {
		port = "8080"
	}
	databaseURL := getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://mirrorme:mirrorme@localhost:5433/mirrorme"
	}
	sleeperBaseURL := getenv("SLEEPER_BASE_URL")
	if sleeperBaseURL == "" {
		sleeperBaseURL = "https://api.sleeper.app/v1"
	}
	rankingsCSVURL := getenv("RANKINGS_CSV_URL")
	if rankingsCSVURL == "" {
		rankingsCSVURL = "https://raw.githubusercontent.com/dynastyprocess/data/master/files/db_fpecr_latest.csv"
	}
	migrationsURL := getenv("MIGRATIONS_URL")
	if migrationsURL == "" {
		migrationsURL = "file://migrations"
	}
	return Config{
		Port:           port,
		SleeperBaseURL: sleeperBaseURL,
		RankingsCSVURL: rankingsCSVURL,
		DatabaseURL:    databaseURL,
		MigrationsURL:  migrationsURL,
	}
}
