FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod ./
COPY main.go ./
RUN go build -o mock_services main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/mock_services .

EXPOSE 15000 16000 17000

CMD ["./mock_services"]
