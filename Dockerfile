FROM golang:1.16.6-buster as builder

WORKDIR /app
COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" ./cmd/http-server

# Create appuser.
ENV USER=appuser
ENV UID=10001
# See https://stackoverflow.com/a/55757473/12429735RUN
RUN adduser \
    --disabled-password \
    --gecos "" \
    --no-create-home \
    --shell "/sbin/nologin" \
    --uid "${UID}" \
    "${USER}"

FROM bitnami/minideb:buster

COPY --from=builder /app/http-server /bin/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/data/sql/migrations/ /migrations
USER appuser:appuser

EXPOSE 8442

CMD ["http-server"]