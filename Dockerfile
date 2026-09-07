FROM golang:1.27.1 AS build-stage

ENV GOPATH=/
COPY ./ ./

RUN go mod download
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /block-headers-service ./cmd/

FROM debian:sid-slim@sha256:17d1843dae9ca66f1617c1e464f40d06059ccefb0c606e8b54687f253af1684e

WORKDIR /service

COPY --from=build-stage /block-headers-service /service/block-headers-service
COPY --from=build-stage /go/data/blockheaders.csv.gz /service/data/blockheaders.csv.gz
COPY --from=build-stage /go/data/blockheaders-testnet.csv.gz /service/data/blockheaders-testnet.csv.gz
COPY --from=build-stage /go/database/migrations /service/database/migrations

CMD ["/service/block-headers-service"]
