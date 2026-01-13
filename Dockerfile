FROM golang:1.22-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git

COPY go.mod .
RUN go mod download

COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=mod -o /out/contiwatch ./cmd/contiwatch

FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache su-exec tzdata
COPY --from=builder /out/contiwatch /app/contiwatch
COPY web/static /app/web/static
COPY entrypoint.sh /app/entrypoint.sh
RUN mkdir -p /data && chmod 755 /app/entrypoint.sh
ENV CONTIWATCH_ADDR=:8080
ENV CONTIWATCH_CONFIG=/data/config.json
EXPOSE 8080
ENTRYPOINT ["/app/entrypoint.sh"]
