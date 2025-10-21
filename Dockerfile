FROM golang:1.25.3-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download && go mod verify

COPY system-scraper/ ./system-scraper/

WORKDIR /app/system-scraper
RUN go build -o system-scraper main.go



FROM alpine:latest


WORKDIR /app

COPY --from=builder /app/system-scraper .

EXPOSE 8081

CMD ["./system-scraper"]