#!/bin/sh

set -e

openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 | openssl rsa -traditional -out key.pem
openssl req -x509 -key key.pem -out cert.der -outform der -days 3650 -config ./openssl.conf
