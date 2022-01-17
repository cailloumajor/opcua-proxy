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

This project uses standard library's [flag](https://pkg.go.dev/flag) and <https://github.com/peterbourgon/ff>
packages, configuration can be provided by flags or environment variables.

```ShellSession
$ opcua-centrifugo -help
USAGE
  opcua-centrifugo [options]

OPTIONS
  Flag               Env Var           Description
  -debug                               log debug information (default: false)
  -opcua-cert-file   OPCUA_CERT_FILE   certificate file path for OPC-UA secure channel (optional)
  -opcua-key-file    OPCUA_KEY_FILE    private key file path for OPC-UA secure channel (optional)
  -opcua-password    OPCUA_PASSWORD    password for OPC-UA authentication (optional)
  -opcua-server-url  OPCUA_SERVER_URL  OPC-UA server endpoint URL (default: opc.tcp://127.0.0.1:4840)
  -opcua-user        OPCUA_USER        user name for OPC-UA authentication (optional)
```
