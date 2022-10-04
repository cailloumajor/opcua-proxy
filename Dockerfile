# syntax=docker/dockerfile:1.3

FROM --platform=$BUILDPLATFORM tonistiigi/xx:1.1.2 as xx

FROM --platform=$BUILDPLATFORM golang:1.19.2-bullseye AS builder

COPY --from=xx / /

WORKDIR /usr/src/app
COPY go.mod go.sum ./
COPY cmd ./cmd
COPY internal ./internal
ARG TARGETPLATFORM
ENV CGO_ENABLED=0
RUN --mount=type=cache,target=/root/.cache/go-build \
    xx-go build -o bin/ -v ./... && \
    xx-verify bin/*

# hadolint ignore=DL3006
FROM gcr.io/distroless/static-debian11

COPY --from=builder /usr/src/app/bin/* /usr/local/bin/

HEALTHCHECK CMD ["/usr/local/bin/healthcheck", "--port", "8080"]

USER nonroot
EXPOSE 8080
CMD ["/usr/local/bin/opcua-proxy"]
