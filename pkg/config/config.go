package config

import "os"

type Config struct {
	Port           string
	SleeperBaseURL string
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	sleeperBaseUrl := os.Getenv("SLEEPER_BASE_URL")
	if sleeperBaseUrl == "" {
		sleeperBaseUrl = "https://api.sleeper.app/v1"
	}
	return Config{
		Port:           port,
		SleeperBaseURL: sleeperBaseUrl,
	}
}
