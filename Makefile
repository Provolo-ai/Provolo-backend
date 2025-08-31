# Makefile for Provolo CLI commands

seed-tiers:
	go run cmd/seed/seed_tier.go

migrate-users-tier:
	go run cmd/migrate/*.go -type=users_tier

migrate-prompt-quota:
	go run cmd/migrate/*.go -type=prompt_quota

