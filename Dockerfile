FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o vesperis-mp ./cmd/vesperis-mp

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/vesperis-mp .
EXPOSE 25565
CMD ["./vesperis-mp"]