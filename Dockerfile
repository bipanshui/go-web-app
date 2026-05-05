FROM golang:1.22 AS builder
WORKDIR /app

COPY go.mod .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:3.20 AS runtime
WORKDIR /app

COPY --from=builder /app/main /app/main
COPY --from=builder /app/static /app/static

EXPOSE 8080

CMD ["/app/main"]
