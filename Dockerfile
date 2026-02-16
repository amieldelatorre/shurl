FROM golang:1.26.0-trixie AS build
WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY ./cmd ./cmd
COPY ./internal ./internal
COPY ./main.go ./
RUN ls -alh
RUN CGO_ENABLED=0 GOOS=linux go build .

FROM debian:13.3 AS final
RUN apt-get update && apt-get install curl netcat-openbsd bind9-dnsutils -y
RUN useradd -ms /bin/sh -u 3333 shurl

COPY --from=build --chown=shurl:shurl /build/shurl /usr/local/bin/
WORKDIR /shurl
RUN chown -R shurl:shurl /shurl
USER shurl

EXPOSE 8080
CMD ["/usr/local/bin/shurl"]