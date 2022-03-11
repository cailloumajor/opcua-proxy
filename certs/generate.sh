#!/bin/sh

openssl genrsa -out key.pem 2048
openssl req -x509 -key key.pem -out cert.der -outform der -days 3650 -nodes -config ./openssl.conf
