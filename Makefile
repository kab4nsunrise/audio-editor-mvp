.PHONY: run build test docker clean fmt

run:
	cd backend && go run .

build:
	cd backend && CGO_ENABLED=0 go build -o ../bin/audio-editor .

test:
	cd backend && go test ./...

fmt:
	cd backend && go fmt ./...

docker:
	docker compose up --build

docker-down:
	docker compose down

clean:
	rm -rf bin uploads
	docker compose down --rmi local 2>/dev/null || true
