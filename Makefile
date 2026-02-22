.PHONY: dev setup build docker-up docker-down

# Start backend + frontend together (Ctrl+C stops both)
dev:
	@if [ ! -f server/.env ]; then cp server/.env.example server/.env && echo "Created server/.env — edit PROXERA_ENCRYPTION_KEY before running"; exit 1; fi
	@trap 'kill %1 2>/dev/null; exit 0' INT TERM EXIT; \
	 cd server && go run . & \
	 cd "$(CURDIR)" && bun run dev

# First-time setup
setup:
	@if [ ! -f server/.env ]; then \
	  cp server/.env.example server/.env; \
	  KEY=$$(openssl rand -hex 32); \
	  sed -i.bak "s/your_64_char_hex_key_here/$$KEY/" server/.env && rm -f server/.env.bak; \
	  echo "✓ Created server/.env with a generated encryption key"; \
	fi
	@cd server && go mod tidy
	@bun install

# Production build
build:
	bun run build
	cd server && CGO_ENABLED=0 go build -ldflags="-w -s" -o proxera .

# Docker full stack
docker-up:
	bun run build
	docker-compose up --build

docker-down:
	docker-compose down
