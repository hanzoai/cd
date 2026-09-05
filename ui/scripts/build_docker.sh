#!/bin/bash

set -e

TAG=${IMAGE_TAG:-'latest'}

docker build --build-arg CD_VERSION="${TAG}" -t "${IMAGE_NAMESPACE:-$(whoami)}/hanzocd-ui:${TAG}" .

if [ "$DOCKER_PUSH" == "true" ]
then
    docker push "${IMAGE_NAMESPACE:-$(whoami)}/hanzocd-ui:${TAG}"
fi
