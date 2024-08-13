# syntax=docker/dockerfile:1

FROM --platform=$TARGETPLATFORM ubuntu:24.04

RUN apt-get update && apt-get install -y wget

WORKDIR /app

ENV BHS_DB_SCHEMA_PATH=./migrations
COPY ./database/migrations $BHS_DB_SCHEMA_PATH
COPY ./release/entrypoint.sh .
ARG project_name=block-header-service
ARG repository_owner=bitcoin-sv
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
