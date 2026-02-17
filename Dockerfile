# syntax=docker/dockerfile:1.7

FROM golang:1.24-bullseye AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS=linux
ARG TARGETARCH=amd64
ENV CGO_ENABLED=0
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /bin/f6n ./cmd/f6n

FROM debian:12-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates ncurses-base && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /bin/f6n /usr/local/bin/f6n

# Default envs can be overridden at runtime
ENV STAGE=dev \
    AWS_REGION=us-east-1

ENTRYPOINT ["f6n"]
