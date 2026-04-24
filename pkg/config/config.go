package config

type Config struct {
	Port           string
	SleeperBaseURL string
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
	return Config{
		Port:           port,
		SleeperBaseURL: sleeperBaseURL,
	}
}
