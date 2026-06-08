# Vendored Centrifugo Protobuf Files

This directory contains protobuf definitions copied from the official
[Centrifugo](https://github.com/centrifugal/centrifugo) repository.

They are vendored to ensure deterministic builds and to avoid network
dependencies during compilation.

## Source

* Repository: <https://github.com/centrifugal/centrifugo>
* Revision: `v6.6.0`

### Files

* `internal/apiproto/api.proto`
* `internal/proxyproto/proxy.proto`

Please update the revision above and the crate version when upgrading these files.

## Updating

```bash
curl -L https://raw.githubusercontent.com/centrifugal/centrifugo/<TAG_OR_COMMIT>/internal/apiproto/api.proto \
  -o proto/centrifugo/api.proto

curl -L https://raw.githubusercontent.com/centrifugal/centrifugo/<TAG_OR_COMMIT>/internal/proxyproto/proxy.proto \
  -o proto/centrifugo/proxy.proto
```

Review upstream changes for breaking modifications before upgrading.

## License

Original files are licensed under the Apache-2.0 License.

See:
<https://github.com/centrifugal/centrifugo/blob/master/LICENSE>
