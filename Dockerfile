# syntax=docker/dockerfile:1.3

FROM --platform=$BUILDPLATFORM tonistiigi/xx:1.1.2 AS xx

FROM --platform=$BUILDPLATFORM rust:1.63.0-bullseye AS builder

COPY --from=xx / /

WORKDIR /usr/src/app

ENV PKG_CONFIG_ALLOW_CROSS=1 \
    CARGO_TARGET_AARCH64_UNKNOWN_LINUX_GNU_LINKER=aarch64-linux-gnu-gcc

COPY Cargo.lock Cargo.toml ./
COPY src ./src
ARG TARGETPLATFORM
# hadolint ignore=DL3008,SC2155
RUN --mount=type=cache,target=/usr/local/cargo/git/db \
    --mount=type=cache,target=/usr/local/cargo/registry/index \
    --mount=type=cache,target=/usr/local/cargo/registry/cache \
    --mount=type=cache,sharing=private,target=/usr/src/app/target \
    xx-apt-get install -y --no-install-recommends libssl-dev && \
    apt-get install -y --no-install-recommends "crossbuild-essential-$(xx-info debian-arch)" && \
    export RUST_TRIPLE="$(xx-info march)-unknown-$(xx-info os)-$(xx-info libc)" && \
    rustup target add "$RUST_TRIPLE" && \
    cargo install --target "$RUST_TRIPLE" --locked --path . --root . && \
    xx-verify bin/*

# hadolint ignore=DL3006
FROM gcr.io/distroless/cc-debian11

WORKDIR /app

COPY --from=builder /usr/src/app/bin/* /usr/local/bin/

HEALTHCHECK CMD ["/usr/local/bin/healthcheck"]

USER nonroot
EXPOSE 8080
CMD ["/usr/local/bin/opcua-proxy"]
