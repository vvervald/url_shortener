FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/api

FROM alpine:3.22

WORKDIR /opt/goapp

COPY --from=builder /app/app ./app
COPY --from=builder /app/config ./config

ENV CONFIG_PATH=/opt/goapp/config/config.yml

CMD ["./app"]