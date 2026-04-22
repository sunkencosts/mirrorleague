.PHONY: dev

dev:
	go run ./cmd/server & cd web && npm run dev
