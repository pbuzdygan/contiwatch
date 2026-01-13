#!/bin/sh
set -e

PUID="${PUID:-1000}"
PGID="${PGID:-1000}"
DOCKER_GID="${DOCKER_GID:-}"

if [ -n "${TZ:-}" ] && [ -f "/usr/share/zoneinfo/$TZ" ]; then
  ln -snf "/usr/share/zoneinfo/$TZ" /etc/localtime
  echo "$TZ" > /etc/timezone
fi

PRIMARY_GROUP_NAME="$(getent group "$PGID" | cut -d: -f1)"
if [ -z "$PRIMARY_GROUP_NAME" ]; then
  addgroup -g "$PGID" contiwatch >/dev/null 2>&1 || true
  PRIMARY_GROUP_NAME="contiwatch"
fi

USER_NAME="$(getent passwd "$PUID" | cut -d: -f1)"
if [ -z "$USER_NAME" ]; then
  adduser -D -H -u "$PUID" -G "$PRIMARY_GROUP_NAME" contiwatch >/dev/null 2>&1 || true
  USER_NAME="contiwatch"
fi

if [ -z "$DOCKER_GID" ] && [ -S /var/run/docker.sock ]; then
  DOCKER_GID="$(stat -c '%g' /var/run/docker.sock 2>/dev/null || true)"
fi

case "$DOCKER_GID" in
  "" | *[!0-9]*)
    DOCKER_GID=""
    ;;
esac

if [ -n "$DOCKER_GID" ] && [ "$DOCKER_GID" != "$PGID" ]; then
  DOCKER_GROUP_NAME="$(getent group "$DOCKER_GID" | cut -d: -f1)"
  if [ -z "$DOCKER_GROUP_NAME" ]; then
    addgroup -g "$DOCKER_GID" docker >/dev/null 2>&1 || true
    DOCKER_GROUP_NAME="docker"
  fi
  addgroup "$USER_NAME" "$DOCKER_GROUP_NAME" >/dev/null 2>&1 || true
fi

chown -R "$PUID:$PGID" /data

exec su-exec "$USER_NAME" /app/contiwatch
