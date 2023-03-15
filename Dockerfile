FROM golang:1.20.0

ENV GOPATH=/
COPY ./ ./

RUN go mod download
RUN go build -o p2p-headers ./cmd/

COPY ./data/sql/migrations/ /migrations

CMD ["./p2p-headers"]
