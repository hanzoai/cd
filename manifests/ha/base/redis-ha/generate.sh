#!/bin/sh -xe

helm dependency update ./chart

AUTOGENMSG="# This is an auto-generated file. DO NOT EDIT"
echo "${AUTOGENMSG}" > ./chart/upstream.yaml

helm version
helm template hanzocd ./chart \
  --namespace hanzo-cd \
  --values ./chart/values.yaml \
  --no-hooks \
  >> ./chart/upstream.yaml
