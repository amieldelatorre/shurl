build:
	go build

run-dev:
	docker compose up app --watch

validatetestdata:
	./.github/scripts/validatetestdata.sh

test:
	date
	time go test -race -cover -parallel 4 ./... -v
	date

testff:
	date
	time go test -race -cover -failfast -parallel 4 ./...
	date