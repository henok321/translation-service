FROM golang:1.25.0-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN make build

FROM debian:bookworm-slim

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates curl && \
    apt-get clean && rm -rf /var/lib/apt/lists/* &&  \
    groupadd --gid 1001 appgroup && \
    useradd --uid 1001 --gid appgroup --create-home appuser

WORKDIR /home/appuser

COPY --from=builder /app/translation-service /home/appuser/translation-service
RUN chown appuser:appgroup /home/appuser/translation-service

EXPOSE 8080

ENV ENVIRONMENT=production

USER appuser

CMD ["/home/appuser/translation-service"]
