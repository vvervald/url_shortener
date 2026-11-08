FROM golang:1.27-alpine AS builder 

WORKDIR /app 

COPY go.mod go.sum ./ 

RUN go mod download 

COPY . . 

RUN go build -o app ./cmd/api 

FROM alpine:3.22

WORKDIR /opt/goapp 

COPY --from=builder /app/app app 

CMD ["/opt/goapp/app"]


