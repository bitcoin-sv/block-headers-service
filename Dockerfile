FROM golang:1.20.0

ENV GOPATH=/
COPY ./ ./

RUN go mod download
RUN go build -o pulse ./cmd/

CMD ["./pulse"]
