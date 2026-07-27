package config

import (
	"errors"
	"os"
)

type Controller struct {
	Address string
	Secret  string
}

func LoadControllerFromEnv() (Controller, error) {
	address := os.Getenv("MIHOMO_URL")
	if address == "" {
		address = "127.0.0.1:9090"
	}

	secret := os.Getenv("MIHOMO_SECRET")
	if secret == "" {
		return Controller{}, errors.New("MIHOMO_SECRET is required")
	}

	return Controller{
		address,
		secret,
	}, nil
}
