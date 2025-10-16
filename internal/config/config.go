package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

type Config struct {
	Port int
}

func GetConfig() Config {
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Printf("invalid port: %v, using default port 8080", err)
		port = 8080
	}

	return Config{
		Port: port,
	}
}

func (c *Config) String() string {
	return fmt.Sprintf("port: %d", c.Port)
}
