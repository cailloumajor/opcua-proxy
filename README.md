<!-- markdownlint-configure-file { "MD033": { "allowed_elements": [ "br" ] } } -->
# OPC-UA proxy

[![Tests and code quality](https://github.com/cailloumajor/opcua-proxy/actions/workflows/tests.yml/badge.svg)](https://github.com/cailloumajor/opcua-proxy/actions/workflows/tests.yml)
[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-yellow.svg)](https://conventionalcommits.org)

A proxy microservice connecting to an OPC-UA server and offering:

- Data change subscriptions through Centrifugo;
- InfluxDB metrics.

## Specifications

### Nodes object

OPC-UA nodes are represented as a JSON object with following fields:

- *namespaceURI*: namespace URI for nodes to monitor (string)
- *nodes*: array of node identifiers, with mapping below:

| JSON type                       | [NodeID type][3] |
| ------------------------------- | ---------------- |
| Integer (positive whole number) | Numeric          |
| String                          | String           |

### Centrifugo subscriptions

[1]: https://centrifugal.dev/docs/server/proxy#subscribe-proxy
[2]: https://centrifugal.dev/docs/server/channels#channel-namespaces
[3]: https://reference.opcfoundation.org/v105/Core/docs/Part3/8.2.3/

- A Centrifugo server (at least v3.1.1) is configured to [proxy subscriptions][1] to this service (`/centrifugo/subscribe` endpoint).
- Clients interested in OPC-UA values changes subscribe to Centrifugo with following characteristics:
  - *Channel*: name and interval, separated by `@`, e.g. `my_nodes@2000`, with:
    - *name*: unique identifier for each nodes set.
    - *interval*: publishing interval in milliseconds.
    - **Note**: channel [namespace][2] is reserved for configuring the proxy endpoint.
  - *Data*: Nodes object (see [above](#nodes-object)).

### InfluxDB metrics endpoint

A `GET` request on `/influxdb-metrics` endpoint returns the configured nodes data values, in InfluxDB line protocol format. Measurement name must be given as the value of `measurement` URL parameter. Other query parameters are expected to each have one value and will be emitted as tags.

The nodes to be read are defined in a JSON file, located in configured path, and containing an array of Nodes objects (see [above](#nodes-object)).

## Data flow

Connection to OPC-UA server and session establishment are considered to have been done successfully.

### Subscriptions

```mermaid
sequenceDiagram
    participant Client
    participant Centrifugo as Centrifugo server
    participant Proxy as Centrifugo / OPC-UA<br>proxy
    participant OPCServer as OPC-UA server
    alt unrecognized channel
        Client->>+Centrifugo: Subscribes to a channel
        Centrifugo->>+Proxy: Proxies the subscription request
        Proxy-->>-Centrifugo: Subscription allowed
        Centrifugo-->>-Client: Success
    else Heartbeat channel
        Client->>+Centrifugo: Subscribes to heartbeat channel
        Centrifugo->>+Proxy: Proxies the subscription request
        Proxy-->>-Centrifugo: Subscription allowed
        Centrifugo-->>-Client: Success
        loop Each heartbeat interval
            Proxy-)Centrifugo: Publishes status
            Centrifugo-)Client: Notifies status
        end
    else OPC-UA related channel
        Client->>+Centrifugo: Subscribes to a channel
        Centrifugo->>+Proxy: Proxies the subscription request
        opt No subscription for this refresh interval
            Proxy->>+OPCServer: Create subscription
            OPCServer-->>-Proxy: Subscription created
        end
        Proxy->>+OPCServer: Create monitored item
        OPCServer-->>-Proxy: Monitored item created
        Proxy-->>-Centrifugo: Subscription allowed
        Centrifugo-->>-Client: Success
        loop Each publishing interval with data change
            OPCServer-)Proxy: Data change notification
            activate Proxy
            Proxy-)Centrifugo: Publish
            deactivate Proxy
            activate Centrifugo
            Centrifugo-)Client: Notifies data change
            deactivate Centrifugo
        end
    end
```

### InfluxDB metrics

```mermaid
sequenceDiagram
    participant Client
    participant Proxy as Centrifugo / OPC-UA<br>proxy
    participant OPCServer as OPC-UA server
    Client->>+Proxy: Requests InfluxDB metrics
    Proxy->>+OPCServer: Reads nodes data values
    OPCServer-->>-Proxy: Sends data values
    Proxy-->>-Client: Sends values in line protocol format
```

## Configuration

:construction: WIP :construction:

```ShellSession
```
