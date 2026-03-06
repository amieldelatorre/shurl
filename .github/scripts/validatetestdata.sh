#!/usr/bin/env bash
set -eou pipefail

rm shurl
docker compose down
rm -rf .docker_data

docker compose up postgresql -d
sleep 5

go build .
export SERVER_AUTH_JWT_KEY=$(openssl genpkey -algorithm EC -pkeyopt ec_paramgen_curve:secp521r1)
./shurl run-migrations -f example-config.yaml

cat internal/test/testdata.sql | docker exec -i postgresql psql -U shurl -v ON_ERROR_STOP=1 && docker compose down