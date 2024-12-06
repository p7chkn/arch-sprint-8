FROM golang:1.23.3 as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main .

FROM debian:bookworm-slim

ENV TZ=Etc/UTC

COPY --from=builder /app/main /app/main

WORKDIR /app

EXPOSE 8000

CMD ["./main"]
