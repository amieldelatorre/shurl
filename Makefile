build:
	go build

run-dev:
	docker compose up app --watch

testsqltestdata:
	rm shurl && \
	docker compose down && \
	rm -rf .docker_data && \
	docker compose up postgresql -d && \
	sleep 5 && \
	go build . && \
	./shurl run-migrations -f example-config.yaml && \
	export SERVER_AUTH_JWT_KEY=$(openssl genpkey -algorithm EC -pkeyopt ec_paramgen_curve:secp521r1 ) && \
	which docker && \
	cat internal/test/testdata.sql | docker exec -i postgresql psql -U shurl -v ON_ERROR_STOP=1 && docker compose down

testsqltestdataclean:
	rm shurl && \
	docker compose down && \
	rm -rf .docker_data	

test:
	date
	time go test -race -cover -parallel 4 ./... -v
	date