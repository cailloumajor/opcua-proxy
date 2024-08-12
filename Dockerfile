# syntax=docker/dockerfile:1.3

FROM --platform=$BUILDPLATFORM tonistiigi/xx:1.4.0 AS xx

FROM --platform=$BUILDPLATFORM rust:1.80.1-bookworm AS builder

COPY --from=xx / /

WORKDIR /usr/src/app

# hadolint ignore=DL3008
RUN apt-get update && apt-get install -y --no-install-recommends clang lld

COPY Cargo.lock Cargo.toml cross-compile.sh ./
COPY src ./src

RUN --mount=type=cache,target=/usr/local/cargo/git/db \
    --mount=type=cache,target=/usr/local/cargo/registry/index \
    --mount=type=cache,target=/usr/local/cargo/registry/cache \
    cargo fetch

ARG TARGETPLATFORM
# hadolint ignore=SC1091
RUN --mount=type=cache,target=/usr/local/cargo/git/db \
    --mount=type=cache,target=/usr/local/cargo/registry/index \
    --mount=type=cache,target=/usr/local/cargo/registry/cache \
    . ./cross-compile.sh && \
    xx-cargo install --locked --path . --root . && \
    xx-verify bin/*

# hadolint ignore=DL3006
FROM gcr.io/distroless/cc-debian12

WORKDIR /app

COPY --from=builder /usr/src/app/bin/* /usr/local/bin/

HEALTHCHECK CMD ["/usr/local/bin/healthcheck"]

USER nonroot
CMD ["/usr/local/bin/opcua-proxy"]
