#!/usr/bin/env bash

me="$0"
log_file=

generate_self_signed () {
    openssl req \
        -x509 \
        -config ./openssl.conf \
        -newkey rsa:2048 \
        -noenc \
        -days 365 \
        -outform der \
        "$@"
}

teardown () {
    if [ "$log_file" ]; then
        docker compose stop
        docker compose logs --timestamps > "$log_file"
    fi
    docker compose down --volumes
}

die () {
    echo "$me: $1" >&2
    teardown
    exit 1
}

while :; do
    case $1 in
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
                    die "log file error"
                fi
            else
                die '"--log-file" requires a non-empty option argument'
            fi
            ;;
        *)
            break
    esac
done

set -eux

# Create a fresh pki tree
rm -r pki-server pki-client || true
mkdir -p pki-server/{own,trusted}/certs
mkdir -p pki-client/{own,private,rejected,trusted}

# Create self-signed certificate and private key for client
generate_self_signed \
    -keyout pki-client/private/opcua-proxy-key.pem \
    -out pki-client/own/opcua-proxy-cert.der \
    -subj "/C=FR/L=Testing Land/O=Testing Corp./CN=OPC-UA proxy" \
    -addext "subjectAltName=URI:urn:opcua-proxy"
chmod +r pki-client/private/opcua-proxy-key.pem
cp pki-client/own/opcua-proxy-cert.der pki-server/trusted/certs/

# Set required variables for Docker Compose
OPCUA_SERVER_UID="$(id -u)"
OPCUA_SERVER_GID="$(id -g)"
export OPCUA_SERVER_UID OPCUA_SERVER_GID

# Build services images
docker compose build

# Add MongoDB initial data
docker compose up -d --quiet-pull mongodb
max_attempts=3
try_success=
for i in $(seq 1 $max_attempts); do
    if docker compose exec mongodb mongosh --norc --quiet /usr/src/initial-data.mongodb; then
        try_success="true"
        break
    fi
    echo "$me: MongoDB initialization: try #$i failed" >&2
    [[ $i != "$max_attempts" ]] && sleep 5
done
if [ "$try_success" != "true" ]; then
    die "failure trying to initialize MongoDB"
fi

# Start config API
docker compose up -d --quiet-pull config-api
max_attempts=5
wait_success=
for i in $(seq 1 $max_attempts); do
    if docker compose exec config-api wget --spider http://127.0.0.1:3000/; then
        wait_success="true"
        break
    fi
    echo "$me: waiting for config API to be healthy: try #$i failed" >&2
    [[ $i != "$max_attempts" ]] && sleep 3
done
if [ "$wait_success" != "true" ]; then
    die "failure waiting for config API to be healthy"
fi

# Start OPC-UA servers
docker compose up -d --quiet-pull opcua-server-first opcua-server-second
max_attempts=3
wait_success=
for i in $(seq 1 $max_attempts); do
    if [[ $(find pki-server/own/certs -name "*.der" | wc -l) -eq 2 ]]; then
        wait_success="true"
        break
    fi
    echo "$me: waiting for OPC-UA server certificate creation: try #$i failed" >&2
    [[ $i != "$max_attempts" ]] && sleep 3
done
if [ "$wait_success" != "true" ]; then
    die "failure waiting for OPC-UA server certificate creation"
fi

# Add OPC-UA servers certificates to opcua-proxy trusted
for f in pki-server/own/certs/*.der; do
    filename=$(basename "$f" | sed -r 's/(.*\[)([^]]*)(.*)/\1\L\2\E\3/')
    cp "$f" "pki-client/trusted/$filename"
done

# Start opcua-proxy (no-value configuration)
docker compose up -d opcua-proxy

# Wait for OPC-UA proxy to be ready
max_attempts=8
wait_success=
for i in $(seq 1 $max_attempts); do
    if docker compose exec opcua-proxy /usr/local/bin/healthcheck; then
        wait_success="true"
        break
    fi
    echo "$me: waiting for OPC-UA proxy to be healthy: try #$i failed" >&2
    [[ $i != "$max_attempts" ]] && sleep 3
done
if [ "$wait_success" != "true" ]; then
    die "failure waiting for OPC-UA proxy to be healthy"
fi

# Run tests on MongoDB instance
if ! docker compose exec mongodb mongosh /usr/src/tests.mongodb --quiet --nodb --norc; then
    die "MongoDB tests (normal configuration) failed"
fi

echo "$me: success 🎉"
teardown
