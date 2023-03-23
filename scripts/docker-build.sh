#!/bin/bash

docker build \
  --pull \
  --platform=linux/amd64 \
  --label=org.opencontainers.image.title=sgpt \
  --label=org.opencontainers.image.description=sgpt \
  --label=org.opencontainers.image.url=https://github.com/tbckr/sgpt \
  --label=org.opencontainers.image.source=https://github.com/tbckr/sgpt \
  --label=org.opencontainers.image.licenses=MIT \
  --build-arg TARGETOS=linux \
  --build-arg TARGETARCH=amd64 \
  -t sgpt:latest \
  .
