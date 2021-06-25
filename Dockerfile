# Stage 1 - the build process
FROM jwahab/go-ps-build:0.1.0 AS build-env

WORKDIR /app
COPY . .

RUN VER=$(git describe --tags) && \
  GIT_COMMIT=$(git rev-parse HEAD) && \
  CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X main.commit=${GIT_COMMIT} -X main.version=${VER}" ./cmd/grpc-server

# Stage 2 - the production environment
FROM scratch
WORKDIR /app
COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build-env /bin/grpc_health_probe /bin/
COPY --from=build-env /app/grpc-server /app/
COPY --from=build-env /app/settings.conf /app/
EXPOSE 9020

CMD ["/app/grpc-server"]