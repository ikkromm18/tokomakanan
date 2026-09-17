package service

import "github.com/ikkromm18/tokomakanan/internal/config"

const (
	defaultBcryptCost = 12
)

func toActorPtr(id uint64) *uint64 {
	if id > 0 {
		return &id
	}
	return nil
}

func getBcryptCost(cfg *config.Config) int {
	if cfg != nil && cfg.BcryptCost > 0 {
		return cfg.BcryptCost
	}
	return defaultBcryptCost
}
