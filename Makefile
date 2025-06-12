build:
	go build -o bin/main main.go

run:
	go run main.go

docker-compose:
	docker compose up -d --build
