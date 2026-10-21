FROM golang@sha256:f24164268fc1e741bbdc0066dbadcf1c8426c5a8489f0729c02c3905a53032fd AS build-stage

ENV GOPATH=/
COPY ./ ./

RUN go mod download
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /block-headers-service ./cmd/

FROM debian@sha256:913f6706df59a68922d1dd08f78c2476560a8d367897200a6005b00e5f67c2d5

WORKDIR /service

COPY --from=build-stage /block-headers-service /service/block-headers-service
COPY --from=build-stage /go/data/blockheaders.csv.gz /service/data/blockheaders.csv.gz
COPY --from=build-stage /go/data/blockheaders-testnet.csv.gz /service/data/blockheaders-testnet.csv.gz
COPY --from=build-stage /go/database/migrations /service/database/migrations

CMD ["/service/block-headers-service"]
