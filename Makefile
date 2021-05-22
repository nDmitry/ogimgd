.SILENT:

test:
	go test ./... -coverprofile cover.out

race:
	go test ./... -race

coverage:
	go tool cover -func cover.out

up:
	docker-compose build app
	docker-compose up -d app
	docker image prune --force
