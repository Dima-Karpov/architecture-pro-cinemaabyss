package config

import "os"

type Config struct {
	Port         string
	KafkaBrokers string
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	return Config{
		Port:         port,
		KafkaBrokers: os.Getenv("KAFKA_BROKERS"),
	}
}
