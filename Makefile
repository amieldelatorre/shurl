build:
	go build

run-dev:
	docker compose up app --watch

test:
	date
	time go test -race -cover -parallel 4 ./... -v
	date