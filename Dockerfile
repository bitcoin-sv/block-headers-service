FROM golang:1.23.5

ENV GOPATH=/
COPY ./ ./

RUN go mod download
RUN go build -o block-headers-service ./cmd/

CMD ["./block-headers-service"]
