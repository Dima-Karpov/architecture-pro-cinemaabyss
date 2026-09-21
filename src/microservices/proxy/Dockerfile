FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o proxy-service ./cmd/api

FROM alpine:latest

RUN apk --no-cache add ca-certificates wget

WORKDIR /root/

COPY --from=builder /app/proxy-service .

EXPOSE 8000

CMD ["./proxy-service"]
