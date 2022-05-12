# Changelog

### [0.7.2](https://github.com/cailloumajor/opcua-proxy/compare/v0.7.1...v0.7.2) (2022-05-12)


### Bug Fixes

* **deps:** update dependency golang to v1.18.2 ([2e4b9cf](https://github.com/cailloumajor/opcua-proxy/commit/2e4b9cf6dffb18a80eb081ac8c70632522cd1fe2))

### [0.7.1](https://github.com/cailloumajor/opcua-proxy/compare/v0.7.0...v0.7.1) (2022-05-06)


### Bug Fixes

* **deps:** update module github.com/gopcua/opcua to v0.3.4 ([6f0aa70](https://github.com/cailloumajor/opcua-proxy/commit/6f0aa70c48928943e08b719135b3cd2b0fef79b3))

## [0.7.0](https://github.com/cailloumajor/opcua-proxy/compare/v0.6.0...v0.7.0) (2022-04-29)


### Features

* add a message to centrifugo subscribe success reply ([58087e0](https://github.com/cailloumajor/opcua-proxy/commit/58087e0ec54f8f46568496b4c6576554fae6c4e8))

## [0.6.0](https://github.com/cailloumajor/opcua-proxy/compare/v0.5.0...v0.6.0) (2022-04-28)


### ⚠ BREAKING CHANGES

* change heartbeat status code order

### Features

* change heartbeat status code order ([729a632](https://github.com/cailloumajor/opcua-proxy/commit/729a632d59e49e1296fdeccbaee36aec5deba6c6))

## [0.5.0](https://github.com/cailloumajor/opcua-proxy/compare/v0.4.1...v0.5.0) (2022-04-27)


### Features

* add heartbeat publication ([182e743](https://github.com/cailloumajor/opcua-proxy/commit/182e7433103e5a9847addfab56f37c7024e51fe3))
* **docs:** add heartbeat in the flow chart ([7ca3ad5](https://github.com/cailloumajor/opcua-proxy/commit/7ca3ad521f86ed76596dd48d1cc6901f2725fa54))

### [0.4.1](https://github.com/cailloumajor/opcua-proxy/compare/v0.4.0...v0.4.1) (2022-04-25)


### Bug Fixes

* **deps:** update module github.com/avast/retry-go/v4 to v4.0.4 ([a12b2ea](https://github.com/cailloumajor/opcua-proxy/commit/a12b2ea1465393662e92aed3a31129169fdfcf98))
* **deps:** update module github.com/gopcua/opcua to v0.3.3 ([a0cd79c](https://github.com/cailloumajor/opcua-proxy/commit/a0cd79c7192a7fcd37f5878eed111a1f479b6997))

## [0.4.0](https://github.com/cailloumajor/opcua-proxy/compare/v0.3.0...v0.4.0) (2022-03-25)


### ⚠ BREAKING CHANGES

* add nodes data values on-demand reading
* rename the project

### Features

* add nodes data values on-demand reading ([b1fce4d](https://github.com/cailloumajor/opcua-proxy/commit/b1fce4d396f32505933cc8b60fa444c8c5c6bd98))


### Bug Fixes

* disable client MIME sniffing ([36e2515](https://github.com/cailloumajor/opcua-proxy/commit/36e2515aa39e1d85c79f50551593acf5e08bc2d5))
* prevent variable capture ([c4a0fbf](https://github.com/cailloumajor/opcua-proxy/commit/c4a0fbfa57852836bb85dfd5b3563b99745d1911))


### Code Refactoring

* rename the project ([93dfa35](https://github.com/cailloumajor/opcua-proxy/commit/93dfa35fa206aeccacc08c0813c0702f13c9fe8e))

## [0.3.0](https://github.com/cailloumajor/opcua-proxy/compare/v0.2.0...v0.3.0) (2022-03-23)


### Features

* add healthcheck tooling ([ea3b43e](https://github.com/cailloumajor/opcua-proxy/commit/ea3b43edd3b29896f9c198ac9acdd331a2df168e))
* check centrifugo server for health status ([c34c1ce](https://github.com/cailloumajor/opcua-proxy/commit/c34c1ce618439aa5a5620e268826d5eb4b7dbe3b))


### Bug Fixes

* check Centrifugo address and namespace ([e8eacf0](https://github.com/cailloumajor/opcua-proxy/commit/e8eacf0aecc55f2d6de877304e98bf0a103f3fa2))
* more consistent exit code ([4027831](https://github.com/cailloumajor/opcua-proxy/commit/40278311243460dfded28be38cf2144b80e69fb4))
* outdated log message ([9595278](https://github.com/cailloumajor/opcua-proxy/commit/9595278f1729c06a4f57d5f2856aade7cdc4df7d))
* skip tidy logic if monitor does not have ([6a082a3](https://github.com/cailloumajor/opcua-proxy/commit/6a082a33c4f4d352c3b0e6b2b490102f8f3ad4b7))

## [0.2.0](https://github.com/cailloumajor/opcua-proxy/compare/v0.1.1...v0.2.0) (2022-03-17)


### ⚠ BREAKING CHANGES

* flexible monitored nodes type

### Features

* add OPC-UA client retrying code ([945c2ba](https://github.com/cailloumajor/opcua-proxy/commit/945c2baa47837833020f0dd38c1a85705c9889d9))


### Bug Fixes

* add missing mutex lock ([0a8c6de](https://github.com/cailloumajor/opcua-proxy/commit/0a8c6de112e38ae0c4d63c5c423e718a3afe5346))
* create a dummy subscription ([dd32433](https://github.com/cailloumajor/opcua-proxy/commit/dd32433a7a1177deb91e462c76f4393ccc65b3cc))
* decoding error caused by nil field ([103259b](https://github.com/cailloumajor/opcua-proxy/commit/103259b4040a07ad74af3c6f7a3b6fdad68bc643))
* **deps:** update module github.com/gopcua/opcua to v0.3.2 ([72b3d80](https://github.com/cailloumajor/opcua-proxy/commit/72b3d802a5acc353b07f51d19d8e123c0fc7ba52))
* do not call Publish if GetDataChange returns ([9016a80](https://github.com/cailloumajor/opcua-proxy/commit/9016a80353abf02a428d9afd170775227a30625d))
* keep monitor at the end of actors stack ([7af2ba4](https://github.com/cailloumajor/opcua-proxy/commit/7af2ba49798e034bdd05b1fdb4c93e597a86bc78))
* refactor monitor internal data ([cab7429](https://github.com/cailloumajor/opcua-proxy/commit/cab74293e74aea76cc61279027c89205f2bce6d5))


### Code Refactoring

* flexible monitored nodes type ([59d54c2](https://github.com/cailloumajor/opcua-proxy/commit/59d54c21d87c432cd7406ea0b7c684f1bc30f32c))

### [0.1.1](https://github.com/cailloumajor/opcua-proxy/compare/v0.1.0...v0.1.1) (2022-03-10)


### Bug Fixes

* disable cgo ([8c9e68c](https://github.com/cailloumajor/opcua-proxy/commit/8c9e68cdfbaf3b33eb420e9a6aec479ca3dfb0d1))

## 0.1.0 (2022-03-10)


### ⚠ BREAKING CHANGES

* change Centrifugo subscription contract

### Features

* implement opcua-proxy command ([2500e39](https://github.com/cailloumajor/opcua-proxy/commit/2500e3947357217d7e804ffc431b8ceb8f9354e5))


### Bug Fixes

* **deps:** update module github.com/gopcua/opcua to v0.2.6 ([32a2aa6](https://github.com/cailloumajor/opcua-proxy/commit/32a2aa6c309caf0d09389ddb17c29e3f9962ead7))


### Reverts

* **ci:** lint only changed files ([fa7c089](https://github.com/cailloumajor/opcua-proxy/commit/fa7c089af071a30ce58cb88ca83f472c5e5b19c0))


### Miscellaneous Chores

* change Centrifugo subscription contract ([f785280](https://github.com/cailloumajor/opcua-proxy/commit/f785280b9746a1b64b896d9a9721d8adfbf16c8a))


### Continuous Integration

* force release-please language ([a5ea8f0](https://github.com/cailloumajor/opcua-proxy/commit/a5ea8f057c4c63fb27b1fcd2f85428d8a40800b3))
