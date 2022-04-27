#!/bin/sh
docker compose run --rm --entrypoint cypress -e DISPLAY -e http_proxy -e https_proxy -e no_proxy="localhost,centrifugo,opcua-proxy" cypress open --project .
