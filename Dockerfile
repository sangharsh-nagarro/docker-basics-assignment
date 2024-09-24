FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main main.go

FROM alpine:edge

WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]
