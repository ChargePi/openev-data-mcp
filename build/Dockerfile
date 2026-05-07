FROM golang:1.26-alpine AS build
WORKDIR /app/src

ENV GOCACHE=/root/.cache/go-build
ENV GOMODCACHE=/root/.cache/go-build
ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/root/.cache/go-build go mod download
RUN go mod verify

COPY . .

RUN --mount=type=cache,target="/root/.cache/go-build" go build -o ../openev-data-mcp ./cmd/app

# Debug stage using Delve
FROM build AS debug

COPY --from=build /app/src /service/src
WORKDIR /service/src

RUN go install github.com/go-delve/delve/cmd/dlv@latest
RUN go build -gcflags="all=-N -l" -o ./openev-data-mcp ./cmd/app

CMD ["/go/bin/dlv", "--listen=:40000", "--headless=true", "--api-version=2", "--accept-multiclient", "exec", "./openev-data-mcp"]

# Fetch grpc_health_probe for the final image health check
FROM alpine:3.21 AS health-probe
RUN wget -qO /grpc_health_probe \
    https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.28/grpc_health_probe-linux-amd64 \
    && chmod +x /grpc_health_probe

FROM alpine:3.21 AS app

ARG user=app
ARG group=docker

RUN addgroup $group && \
    adduser \
        --disabled-password \
        --gecos "" \
        --ingroup $group \
        $user

COPY --chown=$user:$group --from=build /app/openev-data-mcp /usr/local/bin/openev-data-mcp
COPY --from=health-probe /grpc_health_probe /usr/local/bin/grpc_health_probe

USER $user
WORKDIR /usr/local/openev-data-mcp

HEALTHCHECK --interval=10s --timeout=3s --retries=3 \
    CMD ["/usr/local/bin/grpc_health_probe", "-addr=:9090"]

CMD ["/usr/local/bin/openev-data-mcp"]