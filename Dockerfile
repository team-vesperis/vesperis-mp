FROM golang:1.24.4-bullseye AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/vesperis-mp

FROM alpine:3.20

WORKDIR /app
RUN apk update && apk upgrade && apk add --no-cache ca-certificates

COPY --from=builder /app/server /app/server
COPY --from=builder /app/cmd/vesperis-mp/config /app/config

EXPOSE 25565
ENTRYPOINT ["/app/server"]
