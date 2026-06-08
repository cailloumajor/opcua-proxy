#!/usr/bin/env bash

set -eu -o pipefail

readonly me="$0"
log_file=

function generate_self_signed {
    openssl req \
        -x509 \
        -config ./openssl.conf \
        -newkey rsa:2048 \
        -noenc \
        -days 365 \
        -outform der \
        "$@"
}

function teardown {
    if [ "$log_file" ]; then
        docker compose stop
        docker compose logs --timestamps > "$log_file"
    fi
    docker compose down --volumes
}

function log {
    echo >&2 "${me}:" "$@"
}

while :; do
    case ${1-} in
        -h|--help)
            echo "Usage: $me [--log-file path]"
            exit 2
            ;;
        --log-file)
            if [ "$2" ]; then
                if touch "$2"; then
                    log_file=$2
                    shift
                else
                    log "log file error"
                    exit 1
                fi
            else
                log '"--log-file" requires a non-empty option argument'
                exit 1
            fi
            ;;
        *)
            break
    esac
done

log "Creating fresh PKI tree"
rm -r pki-server-first pki-server-second pki-client || true
mkdir -p pki-server-first/{own,trusted}/certs
mkdir -p pki-server-second/{own,trusted}/certs
mkdir -p pki-client/{own,private,rejected,trusted}

log "Creating client certificate and private key"
generate_self_signed \
    -keyout pki-client/private/opcua-proxy-key.pem \
    -out pki-client/own/opcua-proxy-cert.der \
    -subj "/C=FR/L=Testing Land/O=Testing Corp./CN=OPC-UA proxy" \
    -addext "subjectAltName=URI:urn:opcua-proxy"
chmod +r pki-client/private/opcua-proxy-key.pem
cp pki-client/own/opcua-proxy-cert.der pki-server-first/trusted/certs/
cp pki-client/own/opcua-proxy-cert.der pki-server-second/trusted/certs/

# Set required variables for Docker Compose
OPCUA_SERVER_UID="$(id -u)"
OPCUA_SERVER_GID="$(id -g)"
export OPCUA_SERVER_UID OPCUA_SERVER_GID

log "Pulling images"
docker compose pull --quiet

log "Building service images"
docker compose build

trap teardown EXIT

log "Starting dependency services"
docker compose up -d --wait --wait-timeout 30 centrifugo deno opcua-server-first opcua-server-second

log "Adding OPC-UA servers certificates to opcua-proxy trusted"
for f in pki-server-{first,second}/own/certs/*.der; do
    filename=$(basename "${f}" | sed -r 's/(.*\[)([^]]*)(.*)/CN=\1\L\2\E\3/')
    cp "${f}" "pki-client/trusted/${filename}"
done

log "Starting opcua-proxy"
docker compose up -d --wait --wait-timeout 30 opcua-proxy

log "Running Centrifuge client tests"
docker compose exec deno deno run --allow-net --no-lock /app/centrifuge-test.ts

echo "$me: success 🎉"
