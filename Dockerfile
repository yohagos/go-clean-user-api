FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git make ca-certificates

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -v -o main ./cmd/api

CMD ["go", "run", "./cmd/api"]

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

ENV TZ=UTC

WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations

RUN chmod +x /root/main

EXPOSE 8080

CMD ["./main"]