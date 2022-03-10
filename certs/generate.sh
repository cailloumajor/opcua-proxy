#!/bin/sh

openssl req -x509 -out cert.pem -days 3650 -nodes -config ./openssl.conf
openssl x509 -outform der -in cert.pem -out cert.der
rm cert.pem
