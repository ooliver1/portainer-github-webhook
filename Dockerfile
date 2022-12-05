FROM golang:1.19-buster as builder

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY ./src ./src
RUN go build src/main.go


FROM debian:buster-slim

# Remove docker's default of removing cache after use.
RUN rm -f /etc/apt/apt.conf.d/docker-clean; echo 'Binary::apt::APT::Keep-Downloaded-Packages "true";' > /etc/apt/apt.conf.d/keep-cache
ENV PACKAGES ca-certificates
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    apt-get update && apt-get install -yqq --no-install-recommends \
    $PACKAGES && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/main /app/main

VOLUME /config.yaml

CMD ["/app/main"]
