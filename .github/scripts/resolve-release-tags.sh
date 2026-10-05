#!/usr/bin/env bash
set -euo pipefail

# Release metadata is data, never shell source. Match Docker's tag grammar and
# bound the version value before passing it to the linker through build args.
tag="${RELEASE_TAG:?RELEASE_TAG is required}"
target="${RELEASE_TARGET:?RELEASE_TARGET is required}"
repository="${RELEASE_REPOSITORY:?RELEASE_REPOSITORY is required}"

if [[ ! "$tag" =~ ^[A-Za-z0-9_][A-Za-z0-9_.-]{0,123}$ ]]; then
  echo "Invalid release tag: use at most 124 Docker-tag characters." >&2
  exit 1
fi
if [[ ! "$repository" =~ ^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$ ]]; then
  echo "Invalid release repository." >&2
  exit 1
fi
image="ghcr.io/${repository,,}"

case "$target" in
  dev)
    echo "tags=${image}:dev_latest,${image}:dev_${tag}"
    if [[ "$tag" == dev* ]]; then
      echo "version=${tag}"
    else
      echo "version=dev${tag}"
    fi
    ;;
  main)
    echo "tags=${image}:latest,${image}:${tag}"
    echo "version=${tag}"
    ;;
  *)
    echo "Release target must be main or dev." >&2
    exit 1
    ;;
esac
