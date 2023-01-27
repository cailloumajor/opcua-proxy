#!/usr/bin/env sh

if ! xx-info is-cross; then
    return 0
fi

apt-get install -y --no-install-recommends \
    "gcc-$(xx-info)" \
    "libc6-dev-$(xx-info debian-arch)-cross"

export "CFLAGS_$(xx-cargo --print-target-triple | tr - _)=--sysroot=/usr/$(xx-info)"
