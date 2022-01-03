# OPC-UA / Centrifugo proxy

[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-yellow.svg)](https://conventionalcommits.org)

A microservice to proxy OPC-UA data change subscription through Centrifugo.

## Specifications

[1]: https://centrifugal.dev/docs/server/proxy#subscribe-proxy
[2]: https://centrifugal.dev/docs/server/channels#channel-namespaces

* A Centrifugo server is configured to [proxy subscriptions][1] to this service.
* Clients must subscribe to Centrifugo channels with following characteristics:

  | Part           | Value                                                            |
  | -------------- | ---------------------------------------------------------------- |
  | [Namespace][2] | `opc-ua`                                                         |
  | Channel name   | percent-encoded (path type) string identifier of the OPC-UA node |

  Example: `opc‑ua:%22myDB%22.%22signal%22`

## Data flow

![Data flow](/assets/data_flow.png)

## Configuration

The service is configured with following environment variables.

| Key              | Description                                                   |
|------------------|---------------------------------------------------------------|
| OPCUA_SERVER_URL | URL of the OPC-UA server endpoint                             |
| OPCUA_USER       | (Optional) OPC-UA authentication username                     |
| OPCUA_PASSWORD   | (Optional) OPC-UA authentication password                     |
| OPCUA_CERT_FILE  | (Optional) Path of the OPC-UA secure channel certificate file |
| OPCUA_KEY_FILE   | (Optional) Path of the OPC-UA secure channel private key file |
