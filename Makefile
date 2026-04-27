-include api/.env
export

.PHONY: db db-stop db-reset migrate-up migrate-down migrate-version migrate-create test lint dev

db:
	docker compose up -d

db-stop:
	docker compose down

db-reset:
	docker compose down -v
	docker compose up -d

migrate-up:
	cd api && go run ./cmd/migrate up

migrate-down:
	cd api && go run ./cmd/migrate down

migrate-version:
	cd api && go run ./cmd/migrate version

migrate-create:
	@if [ -z "$(name)" ]; then echo "usage: make migrate-create name=<migration_name>"; exit 1; fi
	@next=$$(printf '%06d' $$(( $$(ls api/migrations/*.sql 2>/dev/null | wc -l) / 2 + 1 ))); \
	touch api/migrations/$${next}_$(name).up.sql; \
	touch api/migrations/$${next}_$(name).down.sql; \
	echo "created api/migrations/$${next}_$(name).{up,down}.sql"

test:
	cd api && go test ./...

lint:
	cd api && go vet ./...
	cd web && npm run lint

dev:
	./dev.sh
