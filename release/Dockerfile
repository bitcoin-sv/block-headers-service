# syntax=docker/dockerfile:1

# Pinned by digest for reproducibility and security (Scorecard requirement).
# This is a manifest list digest, so it works with --platform for multi-arch builds.
# Update periodically: docker pull ubuntu:24.04 && docker inspect --format='{{index .RepoDigests 0}}' ubuntu:24.04
FROM --platform=$TARGETPLATFORM ubuntu@sha256:c098726dd9b9733f3e885f8de0ef375206a311b0a317eadc40d347b1115fb944

RUN apt-get update && apt-get install -y wget && rm -rf /var/lib/apt/lists/*

WORKDIR /app

ENV BHS_DB_SCHEMA_PATH=./migrations
COPY ./database/migrations $BHS_DB_SCHEMA_PATH
COPY ./release/entrypoint.sh .
ARG project_name=block-headers-service
ARG repository_owner=bsv-blockchain
ARG TARGETPLATFORM
ENV APP_BINARY=${project_name}
COPY ./dist/$TARGETPLATFORM/$APP_BINARY .

ARG version
ARG gh_repository=$repository_owner/$project_name
ARG tag=${version:+v$version}
ENV VERSION=${version:-develop}
ENV TAG=${tag:-main}
ENV PRELOADED_DB_URL="https://raw.githubusercontent.com/${gh_repository}/${TAG}/data/blockheaders.csv.gz"

VOLUME ["/app/data"]

ENTRYPOINT ["./entrypoint.sh"]
