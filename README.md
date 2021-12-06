# OPC-UA / Centrifugo proxy

[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-yellow.svg)](https://conventionalcommits.org)

A microservice to proxy OPC-UA data change subscription through Centrifugo.

## Specifications

[1]: https://centrifugal.dev/docs/server/channels#channel-namespaces

* A [Centrifugo](https://centrifugal.dev/) server is configured to proxy subscriptions to this service.
* Clients must subscribe to Centrifugo channels with following characteristics:

  | Part           | Value                                                            |
  | -------------- | ---------------------------------------------------------------- |
  | [Namespace][1] | `opc-ua`                                                         |
  | Channel name   | percent-encoded (path type) string identifier of the OPC-UA node |

  Example: `opc‑ua:%22myDB%22.%22signal%22`
