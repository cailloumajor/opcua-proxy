# syntax=docker/dockerfile:1.3

FROM golang:1.17.6-bullseye AS builder

WORKDIR /usr/src/app
COPY go.mod go.sum ./
COPY cmd ./cmd
COPY internal ./internal
RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -o bin/ -v ./...


# hadolint ignore=DL3006
FROM gcr.io/distroless/static-debian11

COPY --from=builder /usr/src/app/bin/* /usr/local/bin/

USER nonroot
CMD ["/usr/local/bin/opcua-centrifugo"]
