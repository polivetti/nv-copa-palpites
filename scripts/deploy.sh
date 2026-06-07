#!/bin/sh
set -eu

if [ -z "${DEPLOY_PATH:-}" ]; then
  echo "DEPLOY_PATH nao definido"
  exit 1
fi

if [ -z "${IMAGE:-}" ]; then
  echo "IMAGE nao definida"
  exit 1
fi

if [ -z "${GHCR_USERNAME:-}" ] || [ -z "${GHCR_TOKEN:-}" ]; then
  echo "credenciais do GHCR nao definidas"
  exit 1
fi

echo "$GHCR_TOKEN" | docker login ghcr.io -u "$GHCR_USERNAME" --password-stdin

cd "$DEPLOY_PATH"
docker pull "$IMAGE"
IMAGE="$IMAGE" docker compose -f docker-compose.prod.yml up -d
docker image prune -f
