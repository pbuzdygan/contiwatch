FROM golang:1.22-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

COPY go.* ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN CGO_ENABLED=0 \
    GOOS="${TARGETOS:-linux}" \
    GOARCH="${TARGETARCH:-amd64}" \
    GOARM="${TARGETVARIANT#v}" \
    go build -mod=mod -o /out/contiwatch ./cmd/contiwatch

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
