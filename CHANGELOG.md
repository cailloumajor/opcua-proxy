# Changelog

## [0.2.0](https://github.com/cailloumajor/opcua-centrifugo/compare/v0.1.1...v0.2.0) (2022-03-17)


### ⚠ BREAKING CHANGES

* flexible monitored nodes type

### Features

* add OPC-UA client retrying code ([945c2ba](https://github.com/cailloumajor/opcua-centrifugo/commit/945c2baa47837833020f0dd38c1a85705c9889d9))


### Bug Fixes

* add missing mutex lock ([0a8c6de](https://github.com/cailloumajor/opcua-centrifugo/commit/0a8c6de112e38ae0c4d63c5c423e718a3afe5346))
* create a dummy subscription ([dd32433](https://github.com/cailloumajor/opcua-centrifugo/commit/dd32433a7a1177deb91e462c76f4393ccc65b3cc))
* decoding error caused by nil field ([103259b](https://github.com/cailloumajor/opcua-centrifugo/commit/103259b4040a07ad74af3c6f7a3b6fdad68bc643))
* **deps:** update module github.com/gopcua/opcua to v0.3.2 ([72b3d80](https://github.com/cailloumajor/opcua-centrifugo/commit/72b3d802a5acc353b07f51d19d8e123c0fc7ba52))
* do not call Publish if GetDataChange returns ([9016a80](https://github.com/cailloumajor/opcua-centrifugo/commit/9016a80353abf02a428d9afd170775227a30625d))
* keep monitor at the end of actors stack ([7af2ba4](https://github.com/cailloumajor/opcua-centrifugo/commit/7af2ba49798e034bdd05b1fdb4c93e597a86bc78))
* refactor monitor internal data ([cab7429](https://github.com/cailloumajor/opcua-centrifugo/commit/cab74293e74aea76cc61279027c89205f2bce6d5))


### Code Refactoring

* flexible monitored nodes type ([59d54c2](https://github.com/cailloumajor/opcua-centrifugo/commit/59d54c21d87c432cd7406ea0b7c684f1bc30f32c))

### [0.1.1](https://github.com/cailloumajor/opcua-centrifugo/compare/v0.1.0...v0.1.1) (2022-03-10)


### Bug Fixes

* disable cgo ([8c9e68c](https://github.com/cailloumajor/opcua-centrifugo/commit/8c9e68cdfbaf3b33eb420e9a6aec479ca3dfb0d1))

## 0.1.0 (2022-03-10)


### ⚠ BREAKING CHANGES

* change Centrifugo subscription contract

### Features

* implement opcua-centrifugo command ([2500e39](https://github.com/cailloumajor/opcua-centrifugo/commit/2500e3947357217d7e804ffc431b8ceb8f9354e5))


### Bug Fixes

* **deps:** update module github.com/gopcua/opcua to v0.2.6 ([32a2aa6](https://github.com/cailloumajor/opcua-centrifugo/commit/32a2aa6c309caf0d09389ddb17c29e3f9962ead7))


### Reverts

* **ci:** lint only changed files ([fa7c089](https://github.com/cailloumajor/opcua-centrifugo/commit/fa7c089af071a30ce58cb88ca83f472c5e5b19c0))


### Miscellaneous Chores

* change Centrifugo subscription contract ([f785280](https://github.com/cailloumajor/opcua-centrifugo/commit/f785280b9746a1b64b896d9a9721d8adfbf16c8a))


### Continuous Integration

* force release-please language ([a5ea8f0](https://github.com/cailloumajor/opcua-centrifugo/commit/a5ea8f057c4c63fb27b1fcd2f85428d8a40800b3))
