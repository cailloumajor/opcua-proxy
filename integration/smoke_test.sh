#!/usr/bin/env bash

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
    echo "$1" >&2
    teardown
    exit 1
}

while :; do
    case $1 in
        -h|--help)
            echo "Usage: $0 [--log-file path]"
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
rm -r pki || true
mkdir -p pki/{own,private,rejected,trusted}

# Create self-signed certificate and private key for server
common_name="open62541Server@opcua-server"
generate_self_signed \
    -keyout pki/private/server_key.pem \
    -out pki/own/server_cert.der \
    -subj "/C=DE/L=Here/O=open62541/CN=$common_name" \
    -addext "subjectAltName=URI:urn:open62541.server.application,DNS:opcua-server"
openssl rsa \
    -traditional \
    -inform PEM \
    -in pki/private/server_key.pem \
    -outform DER \
    -out pki/private/server_key.der
rm pki/private/server_key.pem
openssl_fingerprint=$(openssl x509 -in pki/own/server_cert.der -noout -fingerprint -sha1)
[[ $openssl_fingerprint =~ ([[:xdigit:]]{2}:){19}[[:xdigit:]]{2} ]]
fingerprint=$(echo "${BASH_REMATCH[0]}" | tr "[:upper:]" "[:lower:]" | tr -d ":")
cp pki/own/server_cert.der pki/trusted/"$common_name [$fingerprint]".der

# Create self-signed certificate and private key for client
generate_self_signed \
    -keyout pki/private/private.pem \
    -out pki/own/cert.der \
    -subj "/C=FR/L=Testing Land/O=Testing Corp./CN=OPC-UA proxy" \
    -addext "subjectAltName=URI:urn:opcua-proxy:integration-tests"
chmod +r pki/private/private.pem

# Prevent docker compose warnings about unset environment variable.
export CONFIG_API_URL=""

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
    echo "MongoDB initialization: try #$i failed" >&2
    [[ $i != "$max_attempts" ]] && sleep 5
done
if [ "$try_success" != "true" ]; then
    die "Failure trying to initialize MongoDB"
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
    echo "Waiting for config API to be healthy: try #$i failed" >&2
    [[ $i != "$max_attempts" ]] && sleep 3
done
if [ "$wait_success" != "true" ]; then
    die "Failure waiting for config API to be healthy"
fi

# Start OPC-UA server
docker compose up -d --quiet-pull opcua-server

# Start opcua-proxy (no-value configuration)
CONFIG_API_URL=http://config-api:3000/novalue/ \
docker compose up -d opcua-proxy

# Wait for OPC-UA proxy to be ready
max_attempts=5
wait_success=
for i in $(seq 1 $max_attempts); do
    if docker compose exec opcua-proxy /usr/local/bin/healthcheck; then
        wait_success="true"
        break
    fi
    echo "Waiting for OPC-UA proxy to be healthy: try #$i failed" >&2
    [[ $i != "$max_attempts" ]] && sleep 5
done
if [ "$wait_success" != "true" ]; then
    die "Failure waiting for OPC-UA proxy to be healthy"
fi

# Run tests on MongoDB instance
if ! docker compose exec mongodb mongosh /usr/src/tests-nodata.mongodb --quiet --nodb --norc; then
    die "MongoDB tests (no-value configuration) failed"
fi

# Restart opcua-proxy ("normal" configuration)
docker compose rm --force --stop opcua-proxy
CONFIG_API_URL=http://config-api:3000/normal/ \
docker compose up -d opcua-proxy

# Wait for OPC-UA proxy to be ready
max_attempts=5
wait_success=
for i in $(seq 1 $max_attempts); do
    if docker compose exec opcua-proxy /usr/local/bin/healthcheck; then
        wait_success="true"
        break
    fi
    echo "Waiting for OPC-UA proxy to be healthy: try #$i failed" >&2
    [[ $i != "$max_attempts" ]] && sleep 5
done
if [ "$wait_success" != "true" ]; then
    die "Failure waiting for OPC-UA proxy to be healthy"
fi

# Run tests on MongoDB instance
if ! docker compose exec mongodb mongosh /usr/src/tests.mongodb --quiet --nodb --norc; then
    die "MongoDB tests (normal configuration) failed"
fi

echo "🎉 success"
teardown
