#!/bin/bash

# Default values for environment variables
KV_PORT="${CD_E2E_KV_PORT:-6379}"
KV_IMAGE_TAG=$(grep 'image: redis' manifests/base/redis/hanzocd-redis-deployment.yaml | cut -d':' -f3)

if [ "$CD_KV_LOCAL" = 'true' ]; then
    if ! command -v redis-server &>/dev/null; then
      echo "Redis server is not installed locally. Please install Redis or set CD_KV_LOCAL to false."
      exit 1
    fi

    # Start local Redis server with password if defined
    if [ -z "$KV_PASSWORD" ]; then
        echo "Starting local Redis server without password."
        redis-server --save '' --appendonly no --port "$KV_PORT"
    else
        echo "Starting local Redis server with password."
        redis-server --save '' --appendonly no --port "$KV_PORT" --requirepass "$KV_PASSWORD"
    fi
else
    # Run Redis in a Docker container with password if defined
    if [ -z "$KV_PASSWORD" ]; then
        echo "Starting Docker container without password."
        docker run --rm --name hanzocd-redis -i -p "$KV_PORT:$KV_PORT" docker.io/library/redis:"$KV_IMAGE_TAG" --save '' --appendonly no --port "$KV_PORT"
    else
        echo "Starting Docker container with password."
        docker run --rm --name hanzocd-redis -i -p "$KV_PORT:$KV_PORT" -e KV_PASSWORD="$KV_PASSWORD" docker.io/library/redis:"$KV_IMAGE_TAG" redis-server --save '' --requirepass "$KV_PASSWORD" --appendonly no --port "$KV_PORT"
    fi
fi