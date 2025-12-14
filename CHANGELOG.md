# Changelog

## [6.0.9](https://github.com/cailloumajor/opcua-proxy/compare/v6.0.8...v6.0.9) (2025-12-14)


### Bug Fixes

* **deps:** remove useless bson v2 indirect dependency ([6c1e36a](https://github.com/cailloumajor/opcua-proxy/commit/6c1e36ae53a3210ddcd991f3aa66649639d832a7))
* **deps:** switch to BSON v3 ([d29a98b](https://github.com/cailloumajor/opcua-proxy/commit/d29a98b5df23c7533bda78edb20809fba9a4a83c))
* **deps:** update base images to trixie ([75cccdc](https://github.com/cailloumajor/opcua-proxy/commit/75cccdcdd2a51b0ef611ac922cc78934c0676071))
* **deps:** update rust crate anyhow to v1.0.100 ([#711](https://github.com/cailloumajor/opcua-proxy/issues/711)) ([0080ff5](https://github.com/cailloumajor/opcua-proxy/commit/0080ff55fbd66b162999352607ad1b8498c7bac0))
* **deps:** update rust crate clap to v4.5.46 ([#690](https://github.com/cailloumajor/opcua-proxy/issues/690)) ([1e81878](https://github.com/cailloumajor/opcua-proxy/commit/1e81878ca11be7bd26e9640934c1e9457181a347))
* **deps:** update rust crate clap to v4.5.47 ([#698](https://github.com/cailloumajor/opcua-proxy/issues/698)) ([5b2dfbe](https://github.com/cailloumajor/opcua-proxy/commit/5b2dfbeb0f507aa45bdac2f5b14471311a0b7254))
* **deps:** update rust crate clap to v4.5.48 ([#712](https://github.com/cailloumajor/opcua-proxy/issues/712)) ([5c4e984](https://github.com/cailloumajor/opcua-proxy/commit/5c4e9840ae656f835ff3a822686db2c22a1de559))
* **deps:** update rust crate clap-verbosity-flag to v3.0.4 ([#682](https://github.com/cailloumajor/opcua-proxy/issues/682)) ([61e7c48](https://github.com/cailloumajor/opcua-proxy/commit/61e7c48f93a3f550feef3e9b53b72bdca55a9299))
* **deps:** update rust crate mongodb to v3.2.5 ([#687](https://github.com/cailloumajor/opcua-proxy/issues/687)) ([88cc2db](https://github.com/cailloumajor/opcua-proxy/commit/88cc2db2cf429605f3a26351561af9a3352083c0))
* **deps:** update rust crate mongodb to v3.3.0 ([5bb86a8](https://github.com/cailloumajor/opcua-proxy/commit/5bb86a8ec6b97edd3f605412ce117a2597bd8610))
* **deps:** update rust crate serde to v1.0.221 ([#704](https://github.com/cailloumajor/opcua-proxy/issues/704)) ([9141bea](https://github.com/cailloumajor/opcua-proxy/commit/9141bea72299f3c940c10151ec520e5dc185f66b))
* **deps:** update rust crate serde to v1.0.223 ([#705](https://github.com/cailloumajor/opcua-proxy/issues/705)) ([5e68656](https://github.com/cailloumajor/opcua-proxy/commit/5e686568844decf0cb6f3c0ff18edf924e9ea143))
* **deps:** update rust crate serde to v1.0.225 ([#707](https://github.com/cailloumajor/opcua-proxy/issues/707)) ([f2e6f95](https://github.com/cailloumajor/opcua-proxy/commit/f2e6f95b6569371151dc1a47d00dc2ed0e3ce76d))
* **deps:** update rust crate serde to v1.0.226 ([#713](https://github.com/cailloumajor/opcua-proxy/issues/713)) ([30168d0](https://github.com/cailloumajor/opcua-proxy/commit/30168d09f85800ea2c82304fd716561c71a5a7b6))
* **deps:** update rust crate tracing-subscriber to v0.3.20 ([#693](https://github.com/cailloumajor/opcua-proxy/issues/693)) ([8ae74dd](https://github.com/cailloumajor/opcua-proxy/commit/8ae74dd9deac9831f9abd43b3ecea11c22f44b74))
* **deps:** update rust crate url to v2.5.6 ([#686](https://github.com/cailloumajor/opcua-proxy/issues/686)) ([b596567](https://github.com/cailloumajor/opcua-proxy/commit/b596567c6c582d7baaf806f8e5473273241458cb))
* **deps:** update rust crate url to v2.5.7 ([#688](https://github.com/cailloumajor/opcua-proxy/issues/688)) ([c74d017](https://github.com/cailloumajor/opcua-proxy/commit/c74d017f15591d9fa120017b2461b485c1e0535d))
* **deps:** update rust docker tag to v1.90.0 ([031d93a](https://github.com/cailloumajor/opcua-proxy/commit/031d93ae28a29bb0607edf3426c637e01f28f52c))
* **deps:** update rust docker tag to v1.91.0 ([f227d07](https://github.com/cailloumajor/opcua-proxy/commit/f227d074028d6382624ff78bb0a954aa850a298e))
* **deps:** update rust docker tag to v1.91.1 ([0bc2d6e](https://github.com/cailloumajor/opcua-proxy/commit/0bc2d6e95ca29aa89da49279ee0e3aa32be341b4))
* **deps:** update rust docker tag to v1.92.0 ([757bb7c](https://github.com/cailloumajor/opcua-proxy/commit/757bb7c3481b915ab70cfb11520d2f6dd4749e46))
* **deps:** update tonistiigi/xx docker tag to v1.7.0 ([934acfd](https://github.com/cailloumajor/opcua-proxy/commit/934acfd1f90f2d436f3f819433b96ff51e84df77))
* **deps:** update tonistiigi/xx docker tag to v1.8.0 ([6d3c181](https://github.com/cailloumajor/opcua-proxy/commit/6d3c1815252e62be8359a66ebe3b27f5adb06169))
* **deps:** update tonistiigi/xx docker tag to v1.9.0 ([00ea606](https://github.com/cailloumajor/opcua-proxy/commit/00ea60632dbdc84e39be8a492a07eb1f7af85bfa))

## [6.0.8](https://github.com/cailloumajor/opcua-proxy/compare/v6.0.7...v6.0.8) (2025-08-14)


### Bug Fixes

* **deps:** update rust crate anyhow to v1.0.99 ([#673](https://github.com/cailloumajor/opcua-proxy/issues/673)) ([9578ba8](https://github.com/cailloumajor/opcua-proxy/commit/9578ba8902580b4c788e946b93174d7e1abc7bad))
* **deps:** update rust crate clap to v4.5.41 ([#657](https://github.com/cailloumajor/opcua-proxy/issues/657)) ([47fa8e0](https://github.com/cailloumajor/opcua-proxy/commit/47fa8e060c4833f73a8b74f01b994a1ff219dd2e))
* **deps:** update rust crate clap to v4.5.42 ([#664](https://github.com/cailloumajor/opcua-proxy/issues/664)) ([4dcc083](https://github.com/cailloumajor/opcua-proxy/commit/4dcc083c18912f52d053bda05013de3e58576145))
* **deps:** update rust crate clap to v4.5.43 ([#668](https://github.com/cailloumajor/opcua-proxy/issues/668)) ([f417708](https://github.com/cailloumajor/opcua-proxy/commit/f41770850ab0a365e179f243ec6cfd32babfd586))
* **deps:** update rust crate clap to v4.5.44 ([#672](https://github.com/cailloumajor/opcua-proxy/issues/672)) ([1680f7b](https://github.com/cailloumajor/opcua-proxy/commit/1680f7baef9486c0486d2b602000f6f37f66a43c))
* **deps:** update rust crate clap to v4.5.45 ([#675](https://github.com/cailloumajor/opcua-proxy/issues/675)) ([be25527](https://github.com/cailloumajor/opcua-proxy/commit/be25527decb524f867453dc83ca51834fd23109a))
* **deps:** update rust crate reqwest to v0.12.23 ([#674](https://github.com/cailloumajor/opcua-proxy/issues/674)) ([7bb3384](https://github.com/cailloumajor/opcua-proxy/commit/7bb338478b3c5aa536bdd15d67e8d58aaf0f56fc))
* **deps:** update rust crate tokio to v1.47.1 ([#662](https://github.com/cailloumajor/opcua-proxy/issues/662)) ([650bb46](https://github.com/cailloumajor/opcua-proxy/commit/650bb4693887ece23676dcb596c3d26da57c046a))
* **deps:** update rust crate tokio-util to v0.7.16 ([#666](https://github.com/cailloumajor/opcua-proxy/issues/666)) ([59642c5](https://github.com/cailloumajor/opcua-proxy/commit/59642c5fbd04e5450f74e9436abf7f530ab18275))
* **deps:** update rust docker tag to v1.89.0 ([#669](https://github.com/cailloumajor/opcua-proxy/issues/669)) ([33cdadf](https://github.com/cailloumajor/opcua-proxy/commit/33cdadf8bccd4b12b035de775fa6f65eaf9d2bb7))

## [6.0.7](https://github.com/cailloumajor/opcua-proxy/compare/v6.0.6...v6.0.7) (2025-07-08)


### Bug Fixes

* **deps:** update rust crate anyhow to v1.0.98 ([#620](https://github.com/cailloumajor/opcua-proxy/issues/620)) ([32d011a](https://github.com/cailloumajor/opcua-proxy/commit/32d011afbf980f83396f86818f534a8834a90c4e))
* **deps:** update rust crate clap to v4.5.35 ([#611](https://github.com/cailloumajor/opcua-proxy/issues/611)) ([f2dbc7c](https://github.com/cailloumajor/opcua-proxy/commit/f2dbc7c6a14bcfa31b6555cfcbb7f40aeb98583f))
* **deps:** update rust crate clap to v4.5.36 ([#619](https://github.com/cailloumajor/opcua-proxy/issues/619)) ([649e9a8](https://github.com/cailloumajor/opcua-proxy/commit/649e9a829ed5ff93c5f202c6521abed4e4253e16))
* **deps:** update rust crate clap to v4.5.37 ([#623](https://github.com/cailloumajor/opcua-proxy/issues/623)) ([5c308c7](https://github.com/cailloumajor/opcua-proxy/commit/5c308c7a47f1e0332453f22f6621bfd2565f54e3))
* **deps:** update rust crate clap to v4.5.39 ([#638](https://github.com/cailloumajor/opcua-proxy/issues/638)) ([5e27d74](https://github.com/cailloumajor/opcua-proxy/commit/5e27d74624cf3123936c674451b04470fbf75980))
* **deps:** update rust crate clap to v4.5.40 ([#644](https://github.com/cailloumajor/opcua-proxy/issues/644)) ([17a819d](https://github.com/cailloumajor/opcua-proxy/commit/17a819d6f88998b512d37428a938cfeb252ce372))
* **deps:** update rust crate clap-verbosity-flag to v3.0.3 ([#636](https://github.com/cailloumajor/opcua-proxy/issues/636)) ([ad24b61](https://github.com/cailloumajor/opcua-proxy/commit/ad24b61c2646aee6166ce1135bde781eec686276))
* **deps:** update rust crate env_logger to v0.11.8 ([#612](https://github.com/cailloumajor/opcua-proxy/issues/612)) ([88a63f4](https://github.com/cailloumajor/opcua-proxy/commit/88a63f46cd9c36cfb97189fdc18696417d3d6d19))
* **deps:** update rust crate mongodb to v3.2.4 ([#647](https://github.com/cailloumajor/opcua-proxy/issues/647)) ([4858485](https://github.com/cailloumajor/opcua-proxy/commit/4858485b2421d4e92e185092ca896502699c02e9))
* **deps:** update rust crate reqwest to v0.12.16 ([#639](https://github.com/cailloumajor/opcua-proxy/issues/639)) ([17d5a19](https://github.com/cailloumajor/opcua-proxy/commit/17d5a196879163fd89b06534b59eba7cb3c42e11))
* **deps:** update rust crate reqwest to v0.12.18 ([#640](https://github.com/cailloumajor/opcua-proxy/issues/640)) ([b8e79c2](https://github.com/cailloumajor/opcua-proxy/commit/b8e79c2526d1b5887bb1cb54bfacabe1c0596a10))
* **deps:** update rust crate reqwest to v0.12.19 ([#642](https://github.com/cailloumajor/opcua-proxy/issues/642)) ([c065ff3](https://github.com/cailloumajor/opcua-proxy/commit/c065ff3080275ca0f7de82dd467eb61dc87a87fb))
* **deps:** update rust crate reqwest to v0.12.20 ([#645](https://github.com/cailloumajor/opcua-proxy/issues/645)) ([114788d](https://github.com/cailloumajor/opcua-proxy/commit/114788d6c14cc23d1fdad3ed79ae0f35ad5163ec))
* **deps:** update rust crate reqwest to v0.12.21 ([#648](https://github.com/cailloumajor/opcua-proxy/issues/648)) ([843f57c](https://github.com/cailloumajor/opcua-proxy/commit/843f57c88083d45a210eeaf9371277979c62e173))
* **deps:** update rust crate reqwest to v0.12.22 ([#650](https://github.com/cailloumajor/opcua-proxy/issues/650)) ([8df62fe](https://github.com/cailloumajor/opcua-proxy/commit/8df62fe1e0976d6e4dd719277d7f946d10c0fc05))
* **deps:** update rust crate signal-hook to v0.3.18 ([#629](https://github.com/cailloumajor/opcua-proxy/issues/629)) ([4817192](https://github.com/cailloumajor/opcua-proxy/commit/48171921d1009a0eeea447b02b7392d79e9f4910))
* **deps:** update rust crate tokio to v1.44.2 ([#616](https://github.com/cailloumajor/opcua-proxy/issues/616)) ([81d5c93](https://github.com/cailloumajor/opcua-proxy/commit/81d5c93ff9556b546883a4b56b54d00ee0edc99d))
* **deps:** update rust crate tokio to v1.45.0 ([#628](https://github.com/cailloumajor/opcua-proxy/issues/628)) ([d4213e5](https://github.com/cailloumajor/opcua-proxy/commit/d4213e59eb3ffd0b1b3c64b1a0f914c4635229ff))
* **deps:** update rust crate tokio to v1.45.1 ([#637](https://github.com/cailloumajor/opcua-proxy/issues/637)) ([55f9329](https://github.com/cailloumajor/opcua-proxy/commit/55f9329875db464031ddc4ec43ae784d8e9a6d89))
* **deps:** update rust crate tokio to v1.46.0 ([#651](https://github.com/cailloumajor/opcua-proxy/issues/651)) ([f163eb8](https://github.com/cailloumajor/opcua-proxy/commit/f163eb8193251cb374c7f245244461eb1c984719))
* **deps:** update rust crate tokio to v1.46.1 ([#654](https://github.com/cailloumajor/opcua-proxy/issues/654)) ([536363d](https://github.com/cailloumajor/opcua-proxy/commit/536363da9ac2b2f53aee32e5d24a762d5470faf4))
* **deps:** update rust crate tokio-util to v0.7.15 ([#625](https://github.com/cailloumajor/opcua-proxy/issues/625)) ([51b8ed7](https://github.com/cailloumajor/opcua-proxy/commit/51b8ed7c4bc575686dba6aec3f7187e4128534a8))
* **deps:** update rust docker tag to v1.86.0 ([#614](https://github.com/cailloumajor/opcua-proxy/issues/614)) ([8ffbc81](https://github.com/cailloumajor/opcua-proxy/commit/8ffbc814cd8018934dfc8288b885d366a8b1a255))
* **deps:** update rust docker tag to v1.87.0 ([#632](https://github.com/cailloumajor/opcua-proxy/issues/632)) ([122097e](https://github.com/cailloumajor/opcua-proxy/commit/122097eab78ad0d6dab41a41715287e9e98b635f))
* **deps:** update rust docker tag to v1.88.0 ([#646](https://github.com/cailloumajor/opcua-proxy/issues/646)) ([7712ab5](https://github.com/cailloumajor/opcua-proxy/commit/7712ab5e9681dc295acb8b0bbe2819b17e5f0989))

## [6.0.6](https://github.com/cailloumajor/opcua-proxy/compare/v6.0.5...v6.0.6) (2025-03-30)


### Bug Fixes

* **deps:** update rust crate clap to v4.5.32 ([#593](https://github.com/cailloumajor/opcua-proxy/issues/593)) ([ecdfe93](https://github.com/cailloumajor/opcua-proxy/commit/ecdfe939efeb1a21e1c6e673ae8e0589609c9244))
* **deps:** update rust crate clap to v4.5.34 ([#607](https://github.com/cailloumajor/opcua-proxy/issues/607)) ([0a33c48](https://github.com/cailloumajor/opcua-proxy/commit/0a33c48230d15763d365310124809b1129940241))
* **deps:** update rust crate env_logger to v0.11.7 ([#592](https://github.com/cailloumajor/opcua-proxy/issues/592)) ([5d2e591](https://github.com/cailloumajor/opcua-proxy/commit/5d2e591fa0dc7e04748f43549938f5390ff27e4d))
* **deps:** update rust crate mongodb to v3.2.2 ([#585](https://github.com/cailloumajor/opcua-proxy/issues/585)) ([6cc070b](https://github.com/cailloumajor/opcua-proxy/commit/6cc070b07d19babfd686ec8e494bfb4ae04d17ab))
* **deps:** update rust crate mongodb to v3.2.3 ([#602](https://github.com/cailloumajor/opcua-proxy/issues/602)) ([793bfff](https://github.com/cailloumajor/opcua-proxy/commit/793bfff7ac105fbc473504d31db110cb2feddd25))
* **deps:** update rust crate reqwest to v0.12.13 ([#594](https://github.com/cailloumajor/opcua-proxy/issues/594)) ([1620d82](https://github.com/cailloumajor/opcua-proxy/commit/1620d823d3df5c642152bc99a6798f1437203f0f))
* **deps:** update rust crate reqwest to v0.12.14 ([#595](https://github.com/cailloumajor/opcua-proxy/issues/595)) ([071cf4d](https://github.com/cailloumajor/opcua-proxy/commit/071cf4da2e0007e4bc521eebdd960d9ba0f089ce))
* **deps:** update rust crate reqwest to v0.12.15 ([#600](https://github.com/cailloumajor/opcua-proxy/issues/600)) ([022b164](https://github.com/cailloumajor/opcua-proxy/commit/022b16458be0100babaa37c3a5f9bb400d1a522a))
* **deps:** update rust crate tokio to v1.44.0 ([#589](https://github.com/cailloumajor/opcua-proxy/issues/589)) ([aaaf70d](https://github.com/cailloumajor/opcua-proxy/commit/aaaf70d7c055708a79a7ccbe4dd2d277bcf1b65f))
* **deps:** update rust crate tokio to v1.44.1 ([#596](https://github.com/cailloumajor/opcua-proxy/issues/596)) ([ac5c9c2](https://github.com/cailloumajor/opcua-proxy/commit/ac5c9c2ded23f62fae967a73e2faf13f4510204a))
* **deps:** update rust docker tag to v1.85.1 ([#601](https://github.com/cailloumajor/opcua-proxy/issues/601)) ([3f7b87c](https://github.com/cailloumajor/opcua-proxy/commit/3f7b87c8fdf5383f9cbd871fc9e007dd8a906300))

## [6.0.5](https://github.com/cailloumajor/opcua-proxy/compare/v6.0.4...v6.0.5) (2025-03-06)


### Bug Fixes

* **deps:** update rust crate anyhow to v1.0.94 ([#533](https://github.com/cailloumajor/opcua-proxy/issues/533)) ([ba48253](https://github.com/cailloumajor/opcua-proxy/commit/ba48253bc30317129e42608771859e577116c0ce))
* **deps:** update rust crate anyhow to v1.0.95 ([#547](https://github.com/cailloumajor/opcua-proxy/issues/547)) ([6ec3588](https://github.com/cailloumajor/opcua-proxy/commit/6ec3588e865a2b28121178617f60ceac434f10b2))
* **deps:** update rust crate anyhow to v1.0.96 ([#575](https://github.com/cailloumajor/opcua-proxy/issues/575)) ([98d025c](https://github.com/cailloumajor/opcua-proxy/commit/98d025c3efedad84c153d31a5509cabcabfbe028))
* **deps:** update rust crate anyhow to v1.0.97 ([#581](https://github.com/cailloumajor/opcua-proxy/issues/581)) ([724a075](https://github.com/cailloumajor/opcua-proxy/commit/724a075d9bc9cec8f195ca29fb0e057e93964309))
* **deps:** update rust crate clap to v4.5.22 ([#534](https://github.com/cailloumajor/opcua-proxy/issues/534)) ([fbff7c8](https://github.com/cailloumajor/opcua-proxy/commit/fbff7c8219e0146cc774dcc18daa8e4d88519bc1))
* **deps:** update rust crate clap to v4.5.24 ([#556](https://github.com/cailloumajor/opcua-proxy/issues/556)) ([035cf48](https://github.com/cailloumajor/opcua-proxy/commit/035cf487c220f71cb43814486883da84d0309058))
* **deps:** update rust crate clap to v4.5.26 ([#559](https://github.com/cailloumajor/opcua-proxy/issues/559)) ([2a8b1e4](https://github.com/cailloumajor/opcua-proxy/commit/2a8b1e41dd0186579e5b092ad79545c70e5b6110))
* **deps:** update rust crate clap to v4.5.27 ([#563](https://github.com/cailloumajor/opcua-proxy/issues/563)) ([b2b5aed](https://github.com/cailloumajor/opcua-proxy/commit/b2b5aed6f34f92080f8094f83f136d088d63fc9a))
* **deps:** update rust crate clap to v4.5.28 ([#568](https://github.com/cailloumajor/opcua-proxy/issues/568)) ([28f8456](https://github.com/cailloumajor/opcua-proxy/commit/28f8456e657b43fea843b8fbab3f40b15103b8fe))
* **deps:** update rust crate clap to v4.5.29 ([#570](https://github.com/cailloumajor/opcua-proxy/issues/570)) ([240c2b1](https://github.com/cailloumajor/opcua-proxy/commit/240c2b1205f234f96343072da62f76eb28d4e395))
* **deps:** update rust crate clap to v4.5.30 ([#574](https://github.com/cailloumajor/opcua-proxy/issues/574)) ([0da6ecc](https://github.com/cailloumajor/opcua-proxy/commit/0da6ecc7a61ea4eef40d243d59c92f960c7f022f))
* **deps:** update rust crate clap to v4.5.31 ([#579](https://github.com/cailloumajor/opcua-proxy/issues/579)) ([0764403](https://github.com/cailloumajor/opcua-proxy/commit/076440346ebff847102aa5a33eab88c4b9dc3fbc))
* **deps:** update rust crate clap-verbosity-flag to v3 ([4d836df](https://github.com/cailloumajor/opcua-proxy/commit/4d836df78e4e6ce69a68425e2f8bde6810cf24ae))
* **deps:** update rust crate clap-verbosity-flag to v3 ([522a742](https://github.com/cailloumajor/opcua-proxy/commit/522a7426c7d1fd3e0464faca7a29da49b8875933))
* **deps:** update rust crate clap-verbosity-flag to v3.0.2 ([#552](https://github.com/cailloumajor/opcua-proxy/issues/552)) ([b5f8593](https://github.com/cailloumajor/opcua-proxy/commit/b5f8593e0b054d565577526bac78750e8075ac4b))
* **deps:** update rust crate env_logger to v0.11.6 ([#546](https://github.com/cailloumajor/opcua-proxy/issues/546)) ([8474a4c](https://github.com/cailloumajor/opcua-proxy/commit/8474a4c7dadb040ab2201590ebe39dacc1dd94b8))
* **deps:** update rust crate reqwest to v0.12.10 ([#549](https://github.com/cailloumajor/opcua-proxy/issues/549)) ([9c45117](https://github.com/cailloumajor/opcua-proxy/commit/9c451170cf56d4fd793f6b0691fbe66c522d987a))
* **deps:** update rust crate reqwest to v0.12.11 ([#550](https://github.com/cailloumajor/opcua-proxy/issues/550)) ([3ee6e6d](https://github.com/cailloumajor/opcua-proxy/commit/3ee6e6dc066e509087e74dff2d6878f17ec0bc9d))
* **deps:** update rust crate serde to v1.0.216 ([#542](https://github.com/cailloumajor/opcua-proxy/issues/542)) ([8661ba8](https://github.com/cailloumajor/opcua-proxy/commit/8661ba89710e305bc4630c19953aaa7ff2344382))
* **deps:** update rust crate serde to v1.0.217 ([#551](https://github.com/cailloumajor/opcua-proxy/issues/551)) ([38459bd](https://github.com/cailloumajor/opcua-proxy/commit/38459bd960f8395697d0ebb30fa5690e1ea2e20b))
* **deps:** update rust crate serde to v1.0.218 ([#576](https://github.com/cailloumajor/opcua-proxy/issues/576)) ([8debe2b](https://github.com/cailloumajor/opcua-proxy/commit/8debe2b87693ab01145ef3eb8dae904582994753))
* **deps:** update rust crate tokio to v1.42.0 ([4889d11](https://github.com/cailloumajor/opcua-proxy/commit/4889d11bfac9c7a5d83d207572029ae47a3a0a1a))
* **deps:** update rust crate tokio to v1.42.0 ([7da9141](https://github.com/cailloumajor/opcua-proxy/commit/7da91416f46dca55d1eb2f5f736861de92f4b136))
* **deps:** update rust crate tokio to v1.43.0 ([54d7467](https://github.com/cailloumajor/opcua-proxy/commit/54d7467f91f0c9a6eb9825a760a4d378c5bebc02))
* **deps:** update rust crate tokio to v1.43.0 ([769ab30](https://github.com/cailloumajor/opcua-proxy/commit/769ab3032c71d07778bc1ed6741377eba698c54b))
* **deps:** update rust crate tokio-util to v0.7.13 ([#536](https://github.com/cailloumajor/opcua-proxy/issues/536)) ([1b9894b](https://github.com/cailloumajor/opcua-proxy/commit/1b9894bab30d39df2d87e2cb5b68d593e241b9fc))
* **deps:** update rust docker tag to v1.82.0 ([793ad52](https://github.com/cailloumajor/opcua-proxy/commit/793ad525134e80133125dba196444ffac3045b06))
* **deps:** update rust docker tag to v1.82.0 ([4400230](https://github.com/cailloumajor/opcua-proxy/commit/44002301cad53ab8578424d5fd2d7dc565bfd901))
* **deps:** update rust docker tag to v1.83.0 ([81c62fc](https://github.com/cailloumajor/opcua-proxy/commit/81c62fc3935e4f4e4c2188c40e3c2dc814beca47))
* **deps:** update rust docker tag to v1.83.0 ([25172e5](https://github.com/cailloumajor/opcua-proxy/commit/25172e5ea23489c461f1c7c1e030ed53e4f617b8))
* **deps:** update rust docker tag to v1.84.0 ([ef0d232](https://github.com/cailloumajor/opcua-proxy/commit/ef0d23258a2876040ae625a6ae235bd2c93c2003))
* **deps:** update rust docker tag to v1.84.0 ([4156d6f](https://github.com/cailloumajor/opcua-proxy/commit/4156d6f588ac0857bb78720c31beab186f333d8e))
* **deps:** update rust docker tag to v1.84.1 ([8a19794](https://github.com/cailloumajor/opcua-proxy/commit/8a19794092dae7dc173711b3551a99f78771a30c))
* **deps:** update rust docker tag to v1.84.1 ([ac9e467](https://github.com/cailloumajor/opcua-proxy/commit/ac9e4672074c084998690818255740616cbc2d55))
* **deps:** update Rust version, edition and dependencies ([#583](https://github.com/cailloumajor/opcua-proxy/issues/583)) ([a064b44](https://github.com/cailloumajor/opcua-proxy/commit/a064b44703fc000cd47c5c615806d4d697b08240))
* **deps:** update tonistiigi/xx docker tag to v1.6.1 ([7ec048e](https://github.com/cailloumajor/opcua-proxy/commit/7ec048eac0f3433e791079e9c59c5d74dea75648))
* **deps:** update tonistiigi/xx docker tag to v1.6.1 ([aebaa4e](https://github.com/cailloumajor/opcua-proxy/commit/aebaa4e3ec56634ba14ba6204533a44767b80096))

## [6.0.4](https://github.com/cailloumajor/opcua-proxy/compare/v6.0.3...v6.0.4) (2024-09-28)


### Bug Fixes

* **deps:** update rust crate tokio to v1.38.0 ([9219406](https://github.com/cailloumajor/opcua-proxy/commit/921940635d1375f0d82cbb4ba9f29e72203b564b))
* **deps:** update rust docker tag to v1.80.1 ([9828422](https://github.com/cailloumajor/opcua-proxy/commit/9828422571db47bbb0ff29ab4ebd111852566f66))
* **deps:** update rust docker tag to v1.81.0 ([d166211](https://github.com/cailloumajor/opcua-proxy/commit/d166211fb33f52512cd67f971d560e0addee4f5f))
* **deps:** update tonistiigi/xx docker tag to v1.5.0 ([fc8f964](https://github.com/cailloumajor/opcua-proxy/commit/fc8f964a30472dc1a4ed0d8075eaacf861f3f156))
* **deps:** upgrade Rust docker tag to bookworm ([a7ebac9](https://github.com/cailloumajor/opcua-proxy/commit/a7ebac9d628ad92b70cc4a5e2a71fe7dd11cd1be))

## [6.0.3](https://github.com/cailloumajor/opcua-proxy/compare/v6.0.2...v6.0.3) (2024-05-25)


### Bug Fixes

* **deps:** update rust crate url to v2.5.0 ([51723c1](https://github.com/cailloumajor/opcua-proxy/commit/51723c1a51a1d0375afe7e621549167125b44039))

## [6.0.2](https://github.com/cailloumajor/opcua-proxy/compare/v6.0.1...v6.0.2) (2024-05-24)


### Bug Fixes

* changes in Array API (opcua 0.12.0) ([5363efd](https://github.com/cailloumajor/opcua-proxy/commit/5363efd63dbf1255d19e6cbe71610b67233c3aab))
* **deps:** update rust crate anyhow to v1.0.86 ([45af502](https://github.com/cailloumajor/opcua-proxy/commit/45af502ae62e702680d08e8bbbfbcaed51b4bda6))
* **deps:** update rust crate clap to 4.5.4 ([f19bd44](https://github.com/cailloumajor/opcua-proxy/commit/f19bd44c1068783519084b3a215f6b165892d44e))
* **deps:** update rust crate clap-verbosity-flag to v2.2.0 ([04745bc](https://github.com/cailloumajor/opcua-proxy/commit/04745bc12d9642cc1a6c978c5af9771e92d31395))
* **deps:** update rust crate env_logger to 0.11.0 ([fcaaf14](https://github.com/cailloumajor/opcua-proxy/commit/fcaaf14e0d821f32c5206dd6567f4a5e409d1c33))
* **deps:** update rust crate futures-util to 0.3.30 ([b61f4db](https://github.com/cailloumajor/opcua-proxy/commit/b61f4dbf5be68e998c7a7ee412a99f751c5d2a2e))
* **deps:** update rust crate mongodb to v2.8.2 ([b1631e3](https://github.com/cailloumajor/opcua-proxy/commit/b1631e377f1defa7a260679778eb0681dc364824))
* **deps:** update rust crate opcua to 0.12.0 ([b5d2984](https://github.com/cailloumajor/opcua-proxy/commit/b5d298415918a3279a86023cdd1a1caf22078284))
* **deps:** update rust crate reqwest to 0.12.0 ([c3349bf](https://github.com/cailloumajor/opcua-proxy/commit/c3349bf3214ee08988e57d7c98eeca2fd64f4d43))
* **deps:** update rust crate serde to 1.0.198 ([c60f137](https://github.com/cailloumajor/opcua-proxy/commit/c60f137d8f599f93bd862f08d471c9aa29e97870))
* **deps:** update rust crate serde to v1.0.202 ([ce1eefa](https://github.com/cailloumajor/opcua-proxy/commit/ce1eefac4a72b38ef0da9b498499a7edbfd97e88))
* **deps:** update rust crate tokio to 1.37.0 ([aadb6bb](https://github.com/cailloumajor/opcua-proxy/commit/aadb6bb7c1f78dd32b52db87e997f295068361b2))
* **deps:** update rust crate tokio-util to 0.7.10 ([ea7a55d](https://github.com/cailloumajor/opcua-proxy/commit/ea7a55d96433cb37ab6ab02d021cf80d21d8fc50))
* **deps:** update rust crate tokio-util to v0.7.11 ([cf56da0](https://github.com/cailloumajor/opcua-proxy/commit/cf56da07453b35decd2c3f1724a9f0e39090cfed))
* **deps:** update rust crate url to v2.5.0 ([df86bbe](https://github.com/cailloumajor/opcua-proxy/commit/df86bbed9c78b0a549e61f2778faa7c30a6ed1ab))
* **deps:** update rust docker tag to v1.77.2 ([14613e2](https://github.com/cailloumajor/opcua-proxy/commit/14613e2d23283d371e24d97d756a40e1fe7d6298))
* **deps:** update rust docker tag to v1.78.0 ([719e38f](https://github.com/cailloumajor/opcua-proxy/commit/719e38fb81edb61072ce807e5e1e10eeb912b2b4))
* **deps:** update tokio-tracing monorepo ([2a13064](https://github.com/cailloumajor/opcua-proxy/commit/2a130648eda833c5e4fc7289aa0b8a70ee6a4f1a))
* **deps:** update tonistiigi/xx docker tag to v1.4.0 ([188739b](https://github.com/cailloumajor/opcua-proxy/commit/188739b922943d81035322bdeefd065a933f809e))
* remove usage of arcstr ([96f8b0a](https://github.com/cailloumajor/opcua-proxy/commit/96f8b0afad6863b8f6c7ed6787b565fe54fd3f88))

## [6.0.1](https://github.com/cailloumajor/opcua-proxy/compare/v6.0.0...v6.0.1) (2023-10-19)


### Bug Fixes

* **deps:** update rust crate tracing to 0.1.39 ([e5f69f1](https://github.com/cailloumajor/opcua-proxy/commit/e5f69f1511978cdaad57470fdba53f1b56dccd10))
* **deps:** update rust crate tracing to 0.1.40 ([0dc0a3a](https://github.com/cailloumajor/opcua-proxy/commit/0dc0a3adc45c68552a9e3e0c57ce9d407b4e9639))
* **deps:** update tonistiigi/xx docker tag to v1.3.0 ([569582c](https://github.com/cailloumajor/opcua-proxy/commit/569582c6837a61c0eeb0d2967284967d347b7674))

## [6.0.0](https://github.com/cailloumajor/opcua-proxy/compare/v5.3.3...v6.0.0) (2023-10-13)


### ⚠ BREAKING CHANGES

* implement multi-session management

### Features

* implement multi-session management ([e499d47](https://github.com/cailloumajor/opcua-proxy/commit/e499d4702760c23d6a219bb624fd59853e492a1f))


### Bug Fixes

* **deps:** update rust crate reqwest to 0.11.21 ([098a053](https://github.com/cailloumajor/opcua-proxy/commit/098a0539c42b3df467d10fa569bffa7934e181be))
* **deps:** update rust crate reqwest to 0.11.22 ([38d1978](https://github.com/cailloumajor/opcua-proxy/commit/38d1978a0d1da0e9929d09c240fe110ec3818a51))
* **deps:** update rust crate serde to 1.0.189 ([3d1f799](https://github.com/cailloumajor/opcua-proxy/commit/3d1f7992c6a57da86178007520ad51bdc1e7c5b9))
* **deps:** update rust crate tokio to 1.33.0 ([8fa4264](https://github.com/cailloumajor/opcua-proxy/commit/8fa426444aff72ba97fc7fc6fffdf03ddd7b4dbf))
* **deps:** update rust docker tag to v1.73.0 ([59acd9f](https://github.com/cailloumajor/opcua-proxy/commit/59acd9f6d61494e5a19ddcaab523b8ceb94477c8))

## [5.3.3](https://github.com/cailloumajor/opcua-proxy/compare/v5.3.2...v5.3.3) (2023-10-02)


### Bug Fixes

* **deps:** update rust crate clap to 4.4.5 ([6bd7115](https://github.com/cailloumajor/opcua-proxy/commit/6bd7115850c16d19ed31b84e4fee81367b0c9459))
* **deps:** update rust crate clap to 4.4.6 ([94c3961](https://github.com/cailloumajor/opcua-proxy/commit/94c39614ea21e1758368ad76a3e6cc0d9969f5b8))
* **deps:** update rust crate mongodb to 2.7.0 ([8257641](https://github.com/cailloumajor/opcua-proxy/commit/8257641ffcb32ad23a9f5c98006c70e0ee64c0f9))

## [5.3.2](https://github.com/cailloumajor/opcua-proxy/compare/v5.3.1...v5.3.2) (2023-09-20)


### Bug Fixes

* **deps:** update rust crate clap to 4.4.4 ([8f40629](https://github.com/cailloumajor/opcua-proxy/commit/8f406298b45ed6a9cd8b6d01167d9e6995accae1))
* **deps:** update rust docker tag to v1.72.1 ([b6c3f63](https://github.com/cailloumajor/opcua-proxy/commit/b6c3f631a03dd59efdb106accdea555b467efdff))

## [5.3.1](https://github.com/cailloumajor/opcua-proxy/compare/v5.3.0...v5.3.1) (2023-09-13)


### Bug Fixes

* serialize null strings as empty strings ([850e42b](https://github.com/cailloumajor/opcua-proxy/commit/850e42b5e8bd54217fd7f8b53bde7528973e791c))

## [5.3.0](https://github.com/cailloumajor/opcua-proxy/compare/v5.2.9...v5.3.0) (2023-09-13)


### Features

* accept `Organizes` tag set container nodes ([9aed8ea](https://github.com/cailloumajor/opcua-proxy/commit/9aed8eaf388284261bab303cbe623bb5443cedbe))

## [5.2.9](https://github.com/cailloumajor/opcua-proxy/compare/v5.2.8...v5.2.9) (2023-09-13)


### Bug Fixes

* commit 0d5caea2facefcf91d384c1bc1fbbbc4197f0937 should have been a fix ([e95c0cd](https://github.com/cailloumajor/opcua-proxy/commit/e95c0cd4198184b55d4398f806abc60c359b1ef0))
* **deps:** update rust crate clap to 4.4.3 ([83a5167](https://github.com/cailloumajor/opcua-proxy/commit/83a51670ea91ca47b35fbe1b5b96ad77e1e32404))

## [5.2.8](https://github.com/cailloumajor/opcua-proxy/compare/v5.2.7...v5.2.8) (2023-09-04)


### Bug Fixes

* **deps:** update rust crate clap to 4.4.2 ([171b0b7](https://github.com/cailloumajor/opcua-proxy/commit/171b0b716551f0fbc49d138d5884c0c16a03f705))
* **deps:** update rust crate serde to 1.0.188 ([ca6d785](https://github.com/cailloumajor/opcua-proxy/commit/ca6d785d2d0aced66db3cdd60cf753c2b2f2af1e))
* **deps:** update rust crate url to 2.4.1 ([82ac240](https://github.com/cailloumajor/opcua-proxy/commit/82ac24007c0f66177d6a5563ae99bb64365a5d89))

## [5.2.7](https://github.com/cailloumajor/opcua-proxy/compare/v5.2.6...v5.2.7) (2023-08-25)


### Bug Fixes

* **deps:** update rust crate clap to 4.4.0 ([2e5c484](https://github.com/cailloumajor/opcua-proxy/commit/2e5c484848be268870e7d41c8e2759c7838d2e62))
* **deps:** update rust crate reqwest to 0.11.19 ([5caaeb5](https://github.com/cailloumajor/opcua-proxy/commit/5caaeb54eb46380738a9ba07d3c863b017cc6cbd))
* **deps:** update rust crate reqwest to 0.11.20 ([3636159](https://github.com/cailloumajor/opcua-proxy/commit/3636159edadb568b2aeeae4db80acff65c6cffca))
* **deps:** update rust crate serde to 1.0.186 ([542bfd5](https://github.com/cailloumajor/opcua-proxy/commit/542bfd5531d6a8241789a0f5e8ca928cae290861))
* **deps:** update rust docker tag to v1.72.0 ([09c3b03](https://github.com/cailloumajor/opcua-proxy/commit/09c3b0356f7f96297d70f7e8999f9efdb1c80ca2))
* remove useless Arc's ([2b92606](https://github.com/cailloumajor/opcua-proxy/commit/2b9260626c4028ad5030d1b0e74407dff45d32ad))

## [5.2.6](https://github.com/cailloumajor/opcua-proxy/compare/v5.2.5...v5.2.6) (2023-08-21)


### Bug Fixes

* **deps:** update rust crate clap to 4.3.23 ([aa9f2d8](https://github.com/cailloumajor/opcua-proxy/commit/aa9f2d8be05b3bca11de6ce888553f3e33b87432))

## [5.2.5](https://github.com/cailloumajor/opcua-proxy/compare/v5.2.4...v5.2.5) (2023-08-21)


### Bug Fixes

* **deps:** update rust crate anyhow to 1.0.75 ([1a5d6a7](https://github.com/cailloumajor/opcua-proxy/commit/1a5d6a70ba8e041efb0d3ad06d5237019925fac9))
* **deps:** update rust crate clap to 4.3.22 ([1e64f06](https://github.com/cailloumajor/opcua-proxy/commit/1e64f064ebf9597437ce349f20483d874aeb13ec))
* **deps:** update rust crate clap to 4.3.23 ([6e23e49](https://github.com/cailloumajor/opcua-proxy/commit/6e23e4981269ff1754634299be6e0434ba12544c))
* **deps:** update rust crate serde to 1.0.185 ([1242c71](https://github.com/cailloumajor/opcua-proxy/commit/1242c716d95e1ce793d587bbe675021d22dcf252))
* **deps:** update rust crate tokio to 1.32.0 ([00fa4cb](https://github.com/cailloumajor/opcua-proxy/commit/00fa4cbff8ea285b685c41b5caba524d66ea0cd2))

## [5.2.4](https://github.com/cailloumajor/opcua-proxy/compare/v5.2.3...v5.2.4) (2023-08-16)


### Bug Fixes

* **deps:** update rust crate anyhow to 1.0.74 ([2fbba46](https://github.com/cailloumajor/opcua-proxy/commit/2fbba4646dd65479b58d5a9d741822e2cd9ddf16))
* **deps:** update rust crate mongodb to 2.6.1 ([6b48178](https://github.com/cailloumajor/opcua-proxy/commit/6b481782a258a7da948ab462fd9fa7fd9fc1cdb2))

## [5.2.3](https://github.com/cailloumajor/opcua-proxy/compare/v5.2.2...v5.2.3) (2023-08-14)


### Bug Fixes

* **deps:** update rust crate clap to 4.3.21 ([5bd14a6](https://github.com/cailloumajor/opcua-proxy/commit/5bd14a6fe43756e5144d9d93169d364bd20a24c0))
* **deps:** update rust crate serde to 1.0.183 ([092c43c](https://github.com/cailloumajor/opcua-proxy/commit/092c43cc92aca9de435da1509038f6168ce72b4b))
* **deps:** update rust crate tokio to 1.31.0 ([93dfa23](https://github.com/cailloumajor/opcua-proxy/commit/93dfa23bd9862233fe36301368c616dfe35759ab))
* **deps:** update rust docker tag to v1.71.1 ([90539ea](https://github.com/cailloumajor/opcua-proxy/commit/90539eaa890321522624d0f51bf5c5569f052c56))

## [5.2.2](https://github.com/cailloumajor/opcua-proxy/compare/v5.2.1...v5.2.2) (2023-07-25)


### Bug Fixes

* **deps:** update rust crate clap to 4.3.15 ([39e7b1f](https://github.com/cailloumajor/opcua-proxy/commit/39e7b1f2e76721ea1a498508f864daafbe609c69))
* **deps:** update rust crate clap to 4.3.17 ([e512207](https://github.com/cailloumajor/opcua-proxy/commit/e5122075c1834715f3b242e6dffadc7ca3a37254))
* **deps:** update rust crate clap to 4.3.19 ([300e9d6](https://github.com/cailloumajor/opcua-proxy/commit/300e9d63377d801c4842a7366e26e06d4c74230c))
* **deps:** update rust crate serde to 1.0.173 ([4b72904](https://github.com/cailloumajor/opcua-proxy/commit/4b729041966e02a4e6bb3c79d1499a90cc40bf0e))
* **deps:** update rust crate serde to 1.0.174 ([8e8827d](https://github.com/cailloumajor/opcua-proxy/commit/8e8827d4ae94cd9d0d3164556c091f74244a9502))
* **deps:** update rust crate serde to 1.0.175 ([10f77da](https://github.com/cailloumajor/opcua-proxy/commit/10f77da64f2338f62e55d03bb1dabd8dc74fe481))
* **deps:** update rust crate signal-hook to 0.3.17 ([9891526](https://github.com/cailloumajor/opcua-proxy/commit/989152634d21625a55c6babbe451d276eb83b36f))

## [5.2.1](https://github.com/cailloumajor/opcua-proxy/compare/v5.2.0...v5.2.1) (2023-07-17)


### Bug Fixes

* **deps:** update rust crate anyhow to 1.0.72 ([62f808a](https://github.com/cailloumajor/opcua-proxy/commit/62f808ac6b415195071e6821e962e32b2ed48286))
* **deps:** update rust crate clap to 4.3.1 ([5c53e99](https://github.com/cailloumajor/opcua-proxy/commit/5c53e99a3a03e7195a8e7f51490a640913bb7318))
* **deps:** update rust crate clap to 4.3.11 ([3c0aba8](https://github.com/cailloumajor/opcua-proxy/commit/3c0aba8d027c1eea860654b6d375aa839b070d14))
* **deps:** update rust crate clap to 4.3.12 ([3ab98b8](https://github.com/cailloumajor/opcua-proxy/commit/3ab98b8da46fc1987c6cd4c98ecc42968570ac45))
* **deps:** update rust crate clap to 4.3.2 ([181910c](https://github.com/cailloumajor/opcua-proxy/commit/181910c697b588038782108f93a818a8b1a778c6))
* **deps:** update rust crate clap to 4.3.3 ([273d758](https://github.com/cailloumajor/opcua-proxy/commit/273d758f6d0c87250a57ba044f3f914bac13a38e))
* **deps:** update rust crate mongodb to 2.6.0 ([1898e31](https://github.com/cailloumajor/opcua-proxy/commit/1898e310d6638a23825a6a2777ff47a39a9f00ae))
* **deps:** update rust crate serde to 1.0.164 ([008ca80](https://github.com/cailloumajor/opcua-proxy/commit/008ca80d965ffcd0839c5acb3845a89e27aa238c))
* **deps:** update rust crate serde to 1.0.171 ([53affbc](https://github.com/cailloumajor/opcua-proxy/commit/53affbcff927f11768a578afab82ae1db7c19466))
* **deps:** update rust crate signal-hook to 0.3.16 ([fe615bb](https://github.com/cailloumajor/opcua-proxy/commit/fe615bb0dcd0fbab72801b5114a4ce25ce8a6927))
* **deps:** update rust crate tokio to 1.29.1 ([744cc59](https://github.com/cailloumajor/opcua-proxy/commit/744cc59b3337d00c639a6a13bf04421e29806609))
* **deps:** update rust crate url to 2.4.0 ([8a99cdf](https://github.com/cailloumajor/opcua-proxy/commit/8a99cdfd39aae31dcad71f058a8b5a7a13679bef))
* **deps:** update rust docker tag to v1.70.0 ([e402d0a](https://github.com/cailloumajor/opcua-proxy/commit/e402d0a8f1fa91b92222a2b46d2abf02ac01c8fd))
* **deps:** update rust docker tag to v1.71.0 ([dfbc589](https://github.com/cailloumajor/opcua-proxy/commit/dfbc589397bbe19e63e8ad9bfec27f9fe94e4f41))

## [5.2.0](https://github.com/cailloumajor/opcua-proxy/compare/v5.1.0...v5.2.0) (2023-06-01)


### Features

* remove tags age recording config passthrough ([c49e1ef](https://github.com/cailloumajor/opcua-proxy/commit/c49e1efb29869ef41176a4461b77b3014b896d47))


### Bug Fixes

* **deps:** update rust crate tokio to 1.28.2 ([9d2a97c](https://github.com/cailloumajor/opcua-proxy/commit/9d2a97c4f403f5d9d3e28ff3e9a22c78f2f1c5ab))

## [5.1.0](https://github.com/cailloumajor/opcua-proxy/compare/v5.0.2...v5.1.0) (2023-05-22)


### Features

* use specific cert and key files naming ([a4539b5](https://github.com/cailloumajor/opcua-proxy/commit/a4539b5953c95ffad05e4a35c20cdc20dd4af91c))

## [5.0.2](https://github.com/cailloumajor/opcua-proxy/compare/v5.0.1...v5.0.2) (2023-05-21)


### Bug Fixes

* **deps:** update rust crate clap to 4.3.0 ([0f73a26](https://github.com/cailloumajor/opcua-proxy/commit/0f73a2698c7e131159726f1ab3b13f2df2a39d14))

## [5.0.1](https://github.com/cailloumajor/opcua-proxy/compare/v5.0.0...v5.0.1) (2023-05-17)


### Bug Fixes

* **deps:** update rust crate reqwest to 0.11.18 ([dbe43e7](https://github.com/cailloumajor/opcua-proxy/commit/dbe43e759562e29c85cb74d6d7ffd6ac453119a6))

## [5.0.0](https://github.com/cailloumajor/opcua-proxy/compare/v4.6.0...v5.0.0) (2023-05-16)


### ⚠ BREAKING CHANGES

* implement single tags configuration

### Features

* delete health collection at start and stop ([b42de61](https://github.com/cailloumajor/opcua-proxy/commit/b42de61d00f09105d344ccb2ee81b8ca84b74474))
* implement single tags configuration ([a9f5761](https://github.com/cailloumajor/opcua-proxy/commit/a9f5761b40ec4a32afb339ac0d715dda6d66a76a))


### Bug Fixes

* add error context ([c027edb](https://github.com/cailloumajor/opcua-proxy/commit/c027edb680dab24dcab36edc7f5831ecf13210ad))
* initialize data document with minimal structure ([38960c6](https://github.com/cailloumajor/opcua-proxy/commit/38960c6fcba3a842f3c9b13bebed0d09e664b25a))
* prevent emitting empty tags changes message ([dbe9fd3](https://github.com/cailloumajor/opcua-proxy/commit/dbe9fd36c2ca387c00f30460f276bb30f3a5765d))

## [4.6.0](https://github.com/cailloumajor/opcua-proxy/compare/v4.5.0...v4.6.0) (2023-05-12)


### Features

* use reqwest instead of awc ([ffa208a](https://github.com/cailloumajor/opcua-proxy/commit/ffa208ae79f16124cf23fa0fcabc8ba63f0a1e95))


### Bug Fixes

* **deps:** update rust crate serde to 1.0.163 ([eaaf310](https://github.com/cailloumajor/opcua-proxy/commit/eaaf310bd9cae36aefbe716ad90ce533d7651f40))
* **deps:** update rust crate tokio to 1.28.1 ([071263d](https://github.com/cailloumajor/opcua-proxy/commit/071263d230c0cc00d988f3b69587b7b796633f15))

## [4.5.0](https://github.com/cailloumajor/opcua-proxy/compare/v4.4.0...v4.5.0) (2023-05-05)


### Features

* use actix web client instead of trillium-client ([087382c](https://github.com/cailloumajor/opcua-proxy/commit/087382ce404757ee3fb1c40602ddd0429141913b))


### Bug Fixes

* **deps:** downgrade tracing crate to v0.1.37 ([8f58271](https://github.com/cailloumajor/opcua-proxy/commit/8f5827127b12eef604bbeb58621ccb01f41f6c5c))
* **deps:** update rust crate anyhow to 1.0.71 ([90d185b](https://github.com/cailloumajor/opcua-proxy/commit/90d185b77c61753d2bc3146a10be1941760d231c))
* **deps:** update rust crate clap to 4.2.2 ([1f4193e](https://github.com/cailloumajor/opcua-proxy/commit/1f4193e9dc7a2b862f0ee94788db7d586f5b3ae6))
* **deps:** update rust crate clap to 4.2.4 ([7b2e12a](https://github.com/cailloumajor/opcua-proxy/commit/7b2e12a4b55fb32f73545fd67901300d59bae3f2))
* **deps:** update rust crate clap to 4.2.5 ([5739aba](https://github.com/cailloumajor/opcua-proxy/commit/5739aba445dc28092af490268ec20cc82a93a5e0))
* **deps:** update rust crate clap to 4.2.7 ([593a822](https://github.com/cailloumajor/opcua-proxy/commit/593a822f05f7882642e224a4409c36416c38dcf5))
* **deps:** update rust crate mongodb to 2.5.0 ([82d75cb](https://github.com/cailloumajor/opcua-proxy/commit/82d75cbc88098bcc2b65d046000169816ea4e5dc))
* **deps:** update rust crate serde to 1.0.160 ([6280e33](https://github.com/cailloumajor/opcua-proxy/commit/6280e33001d4b9897d24a0b8013e5e30126b4703))
* **deps:** update rust crate serde to 1.0.162 ([af35f8c](https://github.com/cailloumajor/opcua-proxy/commit/af35f8cdc61158db66d15c8df214d036f350f1f3))
* **deps:** update rust crate tokio to 1.28.0 ([d01b213](https://github.com/cailloumajor/opcua-proxy/commit/d01b213bf8a4c806808573ce1ec94e17cb1ee4fc))
* **deps:** update rust crate tracing to 0.1.38 ([9da0ae8](https://github.com/cailloumajor/opcua-proxy/commit/9da0ae836aed35891296891f0f574dc8ee446bce))
* **deps:** update rust crate tracing-subscriber to 0.3.17 ([29b084f](https://github.com/cailloumajor/opcua-proxy/commit/29b084fa53aaa653c529d18a12706912d9ac3bd3))
* **deps:** update rust crate trillium-client to 0.4.4 ([a3d0ffb](https://github.com/cailloumajor/opcua-proxy/commit/a3d0ffb3b0b829eb939c29654e1ea3db10b33a88))
* **deps:** update rust docker tag to v1.69.0 ([d13d860](https://github.com/cailloumajor/opcua-proxy/commit/d13d860b1c8bdea1224acca9720f60d65615592b))
* **deps:** upgrade trillium crates ([5ba9053](https://github.com/cailloumajor/opcua-proxy/commit/5ba90537fdfe5b3b6073ecdf17f30c40ceda32e3))

## [4.4.0](https://github.com/cailloumajor/opcua-proxy/compare/v4.3.0...v4.4.0) (2023-04-03)


### Features

* implement tags age recording configuration ([6ffee9d](https://github.com/cailloumajor/opcua-proxy/commit/6ffee9d8e8d9435e7c1605ec2a1c3de03555a7b4))

## [4.3.0](https://github.com/cailloumajor/opcua-proxy/compare/v4.2.3...v4.3.0) (2023-04-03)


### Features

* use env_logger to handle logs ([8151f18](https://github.com/cailloumajor/opcua-proxy/commit/8151f184491edbeb583772ba3166eab60ceb68e8))


### Bug Fixes

* **deps:** update rust crate clap-verbosity-flag to 2.0.1 ([f8550c7](https://github.com/cailloumajor/opcua-proxy/commit/f8550c78ba249dbb20fdf0e62801df3b00e9f365))

## [4.2.3](https://github.com/cailloumajor/opcua-proxy/compare/v4.2.2...v4.2.3) (2023-03-31)


### Bug Fixes

* **deps:** update rust crate futures-util to 0.3.28 ([576a957](https://github.com/cailloumajor/opcua-proxy/commit/576a9570372ca1e093db163aa9acc7b522503102))

## [4.2.2](https://github.com/cailloumajor/opcua-proxy/compare/v4.2.1...v4.2.2) (2023-03-30)


### Bug Fixes

* **deps:** update rust crate clap to 4.2.0 ([04d6859](https://github.com/cailloumajor/opcua-proxy/commit/04d6859922ee99570a1e5f78e1bf173bf404a53a))
* **deps:** update rust crate clap to 4.2.1 ([1024e12](https://github.com/cailloumajor/opcua-proxy/commit/1024e12f80bd640119d6904b0eae525c216b7796))
* **deps:** update rust crate serde to 1.0.159 ([eecb169](https://github.com/cailloumajor/opcua-proxy/commit/eecb16948cde3adf27a9e4c4b1a3819aee37d984))
* **deps:** update rust crate tokio to 1.27.0 ([52fe602](https://github.com/cailloumajor/opcua-proxy/commit/52fe602d3b0c5a34535392054f668f8f36db93da))
* **deps:** update rust crate trillium-client to 0.3.1 ([0c1a201](https://github.com/cailloumajor/opcua-proxy/commit/0c1a201a7c41b1b41d81ba4cd05b04e61e807965))
* **deps:** update rust docker tag to v1.68.2 ([15c5c70](https://github.com/cailloumajor/opcua-proxy/commit/15c5c70ec55236bbcdd52c99372b0f2c53a35e16))

## [4.2.1](https://github.com/cailloumajor/opcua-proxy/compare/v4.2.0...v4.2.1) (2023-03-27)


### Bug Fixes

* expect a root `tags` key from config API ([bbe7907](https://github.com/cailloumajor/opcua-proxy/commit/bbe79070cd757bdb1f77c669a8ef10663b9e93b7))

## [4.2.0](https://github.com/cailloumajor/opcua-proxy/compare/v4.1.4...v4.2.0) (2023-03-27)


### Features

* implement tags configuration fetching from API ([6eeacdd](https://github.com/cailloumajor/opcua-proxy/commit/6eeacdd522d982a7160b2fb02edd446dd23cc3af))


### Bug Fixes

* **deps:** update rust crate anyhow to 1.0.70 ([06eb2ac](https://github.com/cailloumajor/opcua-proxy/commit/06eb2ac70a3e8ebcfe5d7f681e87a1ece1b21cf8))
* **deps:** update rust crate clap to 4.1.10 ([ddd11ee](https://github.com/cailloumajor/opcua-proxy/commit/ddd11ee45670f9d806cfc01d8767b9e369802c4b))
* **deps:** update rust crate clap to 4.1.11 ([9f2a00b](https://github.com/cailloumajor/opcua-proxy/commit/9f2a00b358a608b09384fb12862a2c6d2d714898))
* **deps:** update rust crate clap to 4.1.12 ([b2c2bbb](https://github.com/cailloumajor/opcua-proxy/commit/b2c2bbb0943c0fd631561691ea6f6b5fcf3e1645))
* **deps:** update rust crate clap to 4.1.13 ([1b56f30](https://github.com/cailloumajor/opcua-proxy/commit/1b56f30776f239e4f4ccde6f88d6516f673b13f5))
* **deps:** update rust crate clap to 4.1.9 ([850267f](https://github.com/cailloumajor/opcua-proxy/commit/850267f239e5b1576f20f687cac6196b4c6375a1))
* **deps:** update rust crate futures-util to 0.3.27 ([01dbee1](https://github.com/cailloumajor/opcua-proxy/commit/01dbee17a6efd3a332ca576b109fd717e2f17637))
* **deps:** update rust crate serde to 1.0.153 ([8ea2653](https://github.com/cailloumajor/opcua-proxy/commit/8ea26530497bf95ff2982219186703c4bfc5fd07))
* **deps:** update rust crate serde to 1.0.154 ([c466fb9](https://github.com/cailloumajor/opcua-proxy/commit/c466fb9686b6fddf33eb26c3daabed8d78f42ea5))
* **deps:** update rust crate serde to 1.0.155 ([05cdf6e](https://github.com/cailloumajor/opcua-proxy/commit/05cdf6e13ead0a2d9fc36af3aa09d0edeba2e7f5))
* **deps:** update rust crate serde to 1.0.156 ([d1bc1c6](https://github.com/cailloumajor/opcua-proxy/commit/d1bc1c61da1b760e78ef99147729dc01a6164733))
* **deps:** update rust crate serde to 1.0.157 ([681c9bc](https://github.com/cailloumajor/opcua-proxy/commit/681c9bc7794ed843087c6996baebb806a06f3356))
* **deps:** update rust crate serde to 1.0.158 ([89e7a60](https://github.com/cailloumajor/opcua-proxy/commit/89e7a60e2bd6b2c2857245fd7377320460d677df))
* **deps:** update rust docker tag to v1.68.0 ([1e26153](https://github.com/cailloumajor/opcua-proxy/commit/1e2615352f32f6b2db667a396e0b7ad8f955f131))
* **deps:** update rust docker tag to v1.68.1 ([5d796e0](https://github.com/cailloumajor/opcua-proxy/commit/5d796e09f2994312404766246c8c9bc582c138b3))
* implement localized text serialization ([95b1e69](https://github.com/cailloumajor/opcua-proxy/commit/95b1e69190d35287b142972ca91ddf00a261a0d8))
* prevent nesting opcua client's async runtime ([cf21a58](https://github.com/cailloumajor/opcua-proxy/commit/cf21a58f99a71df264e9c6f829301217edf46a0e))
* serialize unsupported variants as null ([d00c1b9](https://github.com/cailloumajor/opcua-proxy/commit/d00c1b967b0fc97aaba998524bdd428fc37b1ea3))
* simplify messages ([ce47ead](https://github.com/cailloumajor/opcua-proxy/commit/ce47eadb2c5d0fc6076a7d2e2b1a89a0b3fabace))


### Reverts

* do not use Cargo sparse protocol yet ([447ee96](https://github.com/cailloumajor/opcua-proxy/commit/447ee968e64bd38ee8dab45ab4c3bd765c268d84))

## [4.1.4](https://github.com/cailloumajor/opcua-proxy/compare/v4.1.3...v4.1.4) (2023-03-06)


### Bug Fixes

* apply verbosity to opcua crate's logs ([dbce4ca](https://github.com/cailloumajor/opcua-proxy/commit/dbce4ca7de3acf3b6018e4f5700fa455a94d131f))
* **deps:** update rust crate serde_json to 1.0.94 ([1e7837a](https://github.com/cailloumajor/opcua-proxy/commit/1e7837a09510a75e1ff3d4e1442bf21581fed01f))

## [4.1.3](https://github.com/cailloumajor/opcua-proxy/compare/v4.1.2...v4.1.3) (2023-03-02)


### Bug Fixes

* **deps:** update rust crate clap to 4.1.7 ([2e5a352](https://github.com/cailloumajor/opcua-proxy/commit/2e5a352e3f9d7dde25995449fa031597b484e674))
* **deps:** update rust crate clap to 4.1.8 ([2fb351c](https://github.com/cailloumajor/opcua-proxy/commit/2fb351cea0e1a3f78010cd8222e8839ec3de39e4))
* **deps:** update rust crate mongodb to 2.4.0 ([cb5487a](https://github.com/cailloumajor/opcua-proxy/commit/cb5487ac87fd3037b4a9351ae9739fceb046d83e))
* **deps:** update rust crate tokio to 1.26.0 ([ed6b385](https://github.com/cailloumajor/opcua-proxy/commit/ed6b385464aa1225c00fb4d950dfc7cb1011ee48))
* use clone wherever to_owned is not needed ([f56e39d](https://github.com/cailloumajor/opcua-proxy/commit/f56e39d60d4ea92ac1cfb40e0d384942540e6bba))

## [4.1.2](https://github.com/cailloumajor/opcua-proxy/compare/v4.1.1...v4.1.2) (2023-02-24)


### Bug Fixes

* enhance async tasks organization ([cb94cb7](https://github.com/cailloumajor/opcua-proxy/commit/cb94cb75f03831f59507a0bd522252020a270d50))

## [4.1.1](https://github.com/cailloumajor/opcua-proxy/compare/v4.1.0...v4.1.1) (2023-02-16)


### Bug Fixes

* **deps:** update rust crate clap to 4.1.6 ([3ee0dc7](https://github.com/cailloumajor/opcua-proxy/commit/3ee0dc75240f712e3ca517996d42b09ea4af59ba))
* **deps:** update rust crate signal-hook to 0.3.15 ([86b6384](https://github.com/cailloumajor/opcua-proxy/commit/86b63847cde76710d175b0fd0d948d379dba5671))
* **deps:** update tonistiigi/xx docker tag to v1.2.1 ([f34218a](https://github.com/cailloumajor/opcua-proxy/commit/f34218a2d72d64846db4f730ebc642015d4db1a6))
* express feature flag requirement ([a5d73a3](https://github.com/cailloumajor/opcua-proxy/commit/a5d73a3e04b51e1e8397313f70295ce25c2ad3f4))

## [4.1.0](https://github.com/cailloumajor/opcua-proxy/compare/v4.0.2...v4.1.0) (2023-02-15)


### Features

* switch to idiomatic pipeline model ([9e8b56f](https://github.com/cailloumajor/opcua-proxy/commit/9e8b56f74ea640d7cf8659eb70a4b1d8f8fccffe))


### Bug Fixes

* **deps:** update rust crate serde_json to 1.0.93 ([7d64b3f](https://github.com/cailloumajor/opcua-proxy/commit/7d64b3f7c70739156f755ded3590c18061a951e1))
* **deps:** update rust docker tag to v1.67.1 ([79305bd](https://github.com/cailloumajor/opcua-proxy/commit/79305bd6f7a668fdfdcee1a2e90c9739ecd67ffc))

## [4.0.2](https://github.com/cailloumajor/opcua-proxy/compare/v4.0.1...v4.0.2) (2023-02-06)


### Bug Fixes

* **deps:** update rust crate anyhow to 1.0.69 ([8163da6](https://github.com/cailloumajor/opcua-proxy/commit/8163da6792abe300acdebe4b5287cc20f3028bd5))
* **deps:** update rust crate clap to 4.1.2 ([b459028](https://github.com/cailloumajor/opcua-proxy/commit/b459028d752c60b2439e06cb19c4a724c31e1754))
* **deps:** update rust crate clap to 4.1.4 ([6b33b62](https://github.com/cailloumajor/opcua-proxy/commit/6b33b62d99dd9b7171fe5aed5e228b59f769f646))
* **deps:** update rust crate futures-util to 0.3.26 ([7c8482d](https://github.com/cailloumajor/opcua-proxy/commit/7c8482df981f1a29e70f1f8ae5407851c947c0e5))
* **deps:** update rust crate serde_json to 1.0.92 ([037c2c5](https://github.com/cailloumajor/opcua-proxy/commit/037c2c5ea76e221acd7796a4bb9899292524ef22))
* **deps:** update rust crate tokio to 1.24.2 ([983950d](https://github.com/cailloumajor/opcua-proxy/commit/983950d82f8f075f5ea80aa847ceffdbf875f2c4))
* **deps:** update rust crate tokio to 1.25.0 ([40398d9](https://github.com/cailloumajor/opcua-proxy/commit/40398d93df5b9b69bf1db9e8361c10b9961c0ff5))
* **deps:** update rust docker tag to v1.67.0 ([7d156e2](https://github.com/cailloumajor/opcua-proxy/commit/7d156e2e9ee519e072112a6420a2c7c62a34f18d))
* **deps:** update tonistiigi/xx docker tag to v1.2.0 ([622ad5a](https://github.com/cailloumajor/opcua-proxy/commit/622ad5ad9f0153e459bdaaa556add578c3e9cc2c))
* use xx-cargo ([9923ee6](https://github.com/cailloumajor/opcua-proxy/commit/9923ee6233c41d31f40f7d8a2e8230e45a7514ed))

## [4.0.1](https://github.com/cailloumajor/opcua-proxy/compare/v4.0.0...v4.0.1) (2023-01-15)


### Bug Fixes

* log count of deleted documents at start ([9fea2a9](https://github.com/cailloumajor/opcua-proxy/commit/9fea2a97cdb5a7d7a711f2e097e446aa86f9e71d))

## [4.0.0](https://github.com/cailloumajor/opcua-proxy/compare/v3.1.0...v4.0.0) (2023-01-15)


### ⚠ BREAKING CHANGES

* delete MongoDB data document when starting
* change data collection members naming

### Features

* change data collection members naming ([ccc5e73](https://github.com/cailloumajor/opcua-proxy/commit/ccc5e73dfa4628802478345910f6a7d74fbe509b))
* delete MongoDB data document when starting ([744d216](https://github.com/cailloumajor/opcua-proxy/commit/744d216d3ec7549741a471da8843a9c6beb0cabf))


### Bug Fixes

* change MongoDB URI argument default value ([5ad8af5](https://github.com/cailloumajor/opcua-proxy/commit/5ad8af5c84d3cb65e334e8378b99c0b6f02e720a))
* **deps:** update rust crate clap to 4.1.0 ([756e4a0](https://github.com/cailloumajor/opcua-proxy/commit/756e4a0b128238ca43fe517d1e5c99ece2f8b8c7))
* **deps:** update rust crate clap to 4.1.1 ([07ce32e](https://github.com/cailloumajor/opcua-proxy/commit/07ce32ebd94a326880bd2f9fc8b1a91b2bb651e8))
* **deps:** update rust docker tag to v1.66.1 ([13d2a92](https://github.com/cailloumajor/opcua-proxy/commit/13d2a9203d9a3101b81a7ce523ed8b12809c8ba5))

## [3.1.0](https://github.com/cailloumajor/opcua-proxy/compare/v3.0.7...v3.1.0) (2023-01-10)


### Features

* add source timestamps ([473123a](https://github.com/cailloumajor/opcua-proxy/commit/473123accda141e506c64a59e3b4af32469ef03a))


### Bug Fixes

* leverage Docker build caching ([b5cbf2c](https://github.com/cailloumajor/opcua-proxy/commit/b5cbf2c20b73fa951b92fbd81d4e2220ed67dd3d))

## [3.0.7](https://github.com/cailloumajor/opcua-proxy/compare/v3.0.6...v3.0.7) (2023-01-06)


### Bug Fixes

* **deps:** update rust crate anyhow to 1.0.68 ([d8aeecc](https://github.com/cailloumajor/opcua-proxy/commit/d8aeeccb0a4b316c2d700578f1e3e90b21d0161b))
* **deps:** update rust crate clap to 4.0.32 ([1c82d79](https://github.com/cailloumajor/opcua-proxy/commit/1c82d797ec33ed0045666e06e5ab0cb2c400f8f7))
* **deps:** update rust crate serde to 1.0.152 ([612186e](https://github.com/cailloumajor/opcua-proxy/commit/612186eb117da4a6511c5e30d492576a35aa650d))
* **deps:** update rust crate serde_json to 1.0.91 ([25323ca](https://github.com/cailloumajor/opcua-proxy/commit/25323ca36f9e69973fa858017dfd4b9d2ee40b19))
* **deps:** update rust crate tokio to 1.24.1 ([987a059](https://github.com/cailloumajor/opcua-proxy/commit/987a05981e4399a35d4b7c1db432169223556fa9))

## [3.0.6](https://github.com/cailloumajor/opcua-proxy/compare/v3.0.5...v3.0.6) (2022-12-16)


### Bug Fixes

* **deps:** update rust crate clap to 4.0.29 ([be4711c](https://github.com/cailloumajor/opcua-proxy/commit/be4711cf7c7ff9fc6895c74e78c48921b87d6740))
* **deps:** update rust crate serde to 1.0.149 ([8372f6b](https://github.com/cailloumajor/opcua-proxy/commit/8372f6bdc87b7cad55e662742ba0da7837e9ec19))
* **deps:** update rust crate serde to 1.0.150 ([105423b](https://github.com/cailloumajor/opcua-proxy/commit/105423b2afd5e6017a705afa23a6186b4a86e38e))
* **deps:** update rust crate serde_json to 1.0.89 ([3baec2a](https://github.com/cailloumajor/opcua-proxy/commit/3baec2a9c8a3b53bf02dc4dcc45c67050eef3a4f))
* **deps:** update rust crate tokio to 1.23.0 ([5650ff4](https://github.com/cailloumajor/opcua-proxy/commit/5650ff4bd16f01d4dabdb00253b976ff2ee40e98))
* **deps:** update rust docker tag to v1.66.0 ([4381ede](https://github.com/cailloumajor/opcua-proxy/commit/4381ede3bec3460bb92c6bebc5ef1359e7560ec3))

## [3.0.5](https://github.com/cailloumajor/opcua-proxy/compare/v3.0.4...v3.0.5) (2022-11-20)


### Bug Fixes

* **deps:** update rust crate clap to 4.0.26 ([66931a4](https://github.com/cailloumajor/opcua-proxy/commit/66931a4ba7ee77f12e613d15df5564a769061331))
* **deps:** update rust crate serde_json to 1.0.88 ([59c060c](https://github.com/cailloumajor/opcua-proxy/commit/59c060c7808a4d995a2f2f010fdb1cfa354b3f34))
* **deps:** update rust crate tokio to 1.22.0 ([167c7d2](https://github.com/cailloumajor/opcua-proxy/commit/167c7d204b3d525a21e925070c82c55a93b43790))

## [3.0.4](https://github.com/cailloumajor/opcua-proxy/compare/v3.0.3...v3.0.4) (2022-11-14)


### Bug Fixes

* **deps:** update rust crate clap to 4.0.23 ([3ee8570](https://github.com/cailloumajor/opcua-proxy/commit/3ee857052b9026235e9d2a66f457edac7d1e6863))
* remove meaningless port expose ([c99040a](https://github.com/cailloumajor/opcua-proxy/commit/c99040ad9c02a569d60deca0cb7f0c16688808ff))

## [3.0.3](https://github.com/cailloumajor/opcua-proxy/compare/v3.0.2...v3.0.3) (2022-11-09)


### Bug Fixes

* **deps:** update rust crate clap to 4.0.22 ([9d7e843](https://github.com/cailloumajor/opcua-proxy/commit/9d7e843264444b069ca0c2a66499a415884ca213))
* **deps:** update rust docker tag to v1.65.0 ([4011756](https://github.com/cailloumajor/opcua-proxy/commit/4011756fe9759609a4516453156b470d30caf95b))

## [3.0.2](https://github.com/cailloumajor/opcua-proxy/compare/v3.0.1...v3.0.2) (2022-10-24)


### Bug Fixes

* **deps:** update rust crate serde to 1.0.147 ([5ed6c71](https://github.com/cailloumajor/opcua-proxy/commit/5ed6c71fb8071f39583c087a92e8048b1746d641))
* prevent conversion with the help of zip ([3be37ed](https://github.com/cailloumajor/opcua-proxy/commit/3be37edcc2ff2a5299c1c7374d89d166fe7ca26f))

## [3.0.1](https://github.com/cailloumajor/opcua-proxy/compare/v3.0.0...v3.0.1) (2022-10-21)


### Bug Fixes

* **deps:** update rust crate clap to 4.0.18 ([b5eb656](https://github.com/cailloumajor/opcua-proxy/commit/b5eb6562b23ab51517ba28d475325f376e6e9c33))
* **deps:** update rust crate futures-util to 0.3.25 ([abb0088](https://github.com/cailloumajor/opcua-proxy/commit/abb0088ce5d4442cd7e4ff7c3b7f11c2c8d838e9))
* **deps:** update rust crate mongodb to 2.3.1 ([9e252d3](https://github.com/cailloumajor/opcua-proxy/commit/9e252d32b817e948c80daf52d1fd6978c508b4da))
* **deps:** update rust crate serde to 1.0.146 ([d571f6f](https://github.com/cailloumajor/opcua-proxy/commit/d571f6f6c77f6acd4cc7ec946041a48bea4e2b52))
* **deps:** update rust crate serde_json to 1.0.87 ([08a3efb](https://github.com/cailloumajor/opcua-proxy/commit/08a3efb50a05003a3d578587257dfcb9dae95414))
* **deps:** update rust crate tracing to 0.1.37 ([0485907](https://github.com/cailloumajor/opcua-proxy/commit/0485907da1054d3c0729456af2c4255df77a9b78))
* **deps:** update rust crate tracing-subscriber to 0.3.16 ([2b66e03](https://github.com/cailloumajor/opcua-proxy/commit/2b66e0315bea6a90d1917a2a67cb7d8cd1c5ca9c))
* **deps:** update rust docker tag to v1.64.0 ([028789c](https://github.com/cailloumajor/opcua-proxy/commit/028789c7e973d55eef432f8ab448b18f2c8a831a))

## [3.0.0](https://github.com/cailloumajor/opcua-proxy/compare/v2.1.4...v3.0.0) (2022-10-21)


### ⚠ BREAKING CHANGES

* rewrite in Rust

### Features

* add instrumentation ([c4f114b](https://github.com/cailloumajor/opcua-proxy/commit/c4f114b3a7c3ea02c54a9a0854a81418db67b995))
* add verbosity flags ([1a03819](https://github.com/cailloumajor/opcua-proxy/commit/1a03819a7f8a1543703550ef1b762bdbf05d68ec))
* implement health check ([d81a264](https://github.com/cailloumajor/opcua-proxy/commit/d81a264e14290e752d27259870ac18f4cf0c9e46))
* implement integration tests ([4edede2](https://github.com/cailloumajor/opcua-proxy/commit/4edede2e0fbb3add1d0329bcd619c413813fda58))
* rewrite in Rust ([839d587](https://github.com/cailloumajor/opcua-proxy/commit/839d5877790b6aca8471db3fa39cb548d17dfc95))


### Bug Fixes

* add partner id to MongoDB app name ([d709bb6](https://github.com/cailloumajor/opcua-proxy/commit/d709bb642968de7557e007498b23960ab965fbbb))
* add session retry policy ([d9904de](https://github.com/cailloumajor/opcua-proxy/commit/d9904de27c6527e6962c0f84e8747c623a0655f5))
* **ci:** binary name for usage check ([15a46e5](https://github.com/cailloumajor/opcua-proxy/commit/15a46e57741be4a5e5c5910230f054c0563aa254))
* **ci:** ensure OpenSSL v3 ([e7b6a8a](https://github.com/cailloumajor/opcua-proxy/commit/e7b6a8a100289be632170f9d92bd71b9f54e57bf))
* **deps:** update golang.org/x/exp digest to 334a238 ([4f5129a](https://github.com/cailloumajor/opcua-proxy/commit/4f5129a7a7c7b1f94843a56b3d48b0bf3d2852ea))
* **deps:** update golang.org/x/exp digest to 4cc3b17 ([d70b55f](https://github.com/cailloumajor/opcua-proxy/commit/d70b55f3ca13d98c414ec02f91ce1253dec52e21))
* **deps:** update golang.org/x/exp digest to 807a232 ([64aec50](https://github.com/cailloumajor/opcua-proxy/commit/64aec509096beed4f05751b6b545ac6bde2e7288))
* **deps:** update golang.org/x/exp digest to bd9bcdd ([b747037](https://github.com/cailloumajor/opcua-proxy/commit/b7470371a34b836eb8d467c011b3cec7d079324e))
* serialize DateTime as a string ([93b32f1](https://github.com/cailloumajor/opcua-proxy/commit/93b32f15de1cf2ea71391518136c2c253556dd61))
* synchronize project version ([fc1a61c](https://github.com/cailloumajor/opcua-proxy/commit/fc1a61ce5f9216165e5c87d729d4b2ebdaef453b))
* update usage in README.md ([05bd502](https://github.com/cailloumajor/opcua-proxy/commit/05bd502e8ca2ab73fceb17752f94e92bd7e3211f))

## [2.1.4](https://github.com/cailloumajor/opcua-proxy/compare/v2.1.3...v2.1.4) (2022-08-22)


### Bug Fixes

* generalize health checking and isolate gocent ([2537487](https://github.com/cailloumajor/opcua-proxy/commit/25374871dbb8e8dc656776c62c1e0b4f127c61d9))
* refactor lineprotocol package ([50a003d](https://github.com/cailloumajor/opcua-proxy/commit/50a003dea4dd16f88923095126b7b2b0b561a6f9))

## [2.1.3](https://github.com/cailloumajor/opcua-proxy/compare/v2.1.2...v2.1.3) (2022-08-19)


### Bug Fixes

* **deps:** update dependency golang to v1.19.0 ([c7ae289](https://github.com/cailloumajor/opcua-proxy/commit/c7ae2895236485ee5d5a05e37dbdba3055faed43))
* refactor to only return concrete types ([85563d2](https://github.com/cailloumajor/opcua-proxy/commit/85563d2c5c8883c84ba399336e7ed11737ff93d2))
* remove unneeded method ([7cff50d](https://github.com/cailloumajor/opcua-proxy/commit/7cff50d704a6ede3239c73c3fa17d45a4c87df00))
* return concrete type in interface ([e831f6a](https://github.com/cailloumajor/opcua-proxy/commit/e831f6a1380ad7986e52006b5143e94996d6655b))
* tidy-up top-level functions dependencies implementation ([a464082](https://github.com/cailloumajor/opcua-proxy/commit/a464082d81e604d21c921e010c7dafcb2845c8de))

## [2.1.2](https://github.com/cailloumajor/opcua-proxy/compare/v2.1.1...v2.1.2) (2022-08-01)


### Bug Fixes

* **deps:** update golang.org/x/exp digest to a9213ee ([c947efc](https://github.com/cailloumajor/opcua-proxy/commit/c947efcb2518c917b93ddb34455c41741ebee696))

## [2.1.1](https://github.com/cailloumajor/opcua-proxy/compare/v2.1.0...v2.1.1) (2022-07-22)


### Bug Fixes

* **deps:** update dependency tonistiigi/xx to v1.1.2 ([41080d4](https://github.com/cailloumajor/opcua-proxy/commit/41080d4a293be9c616ca163fcdd59b822c8600b0))

## [2.1.0](https://github.com/cailloumajor/opcua-proxy/compare/v2.0.2...v2.1.0) (2022-07-19)


### Features

* also sort fields by name in metrics response ([4cc0c6f](https://github.com/cailloumajor/opcua-proxy/commit/4cc0c6f78fee45144df344fdbbdd660ff3e8dbd4))


### Bug Fixes

* **ci:** run golangci-lint with the same go version as tests ([646b2bf](https://github.com/cailloumajor/opcua-proxy/commit/646b2bf0c6f670f62f79860cb609977b5db02823))

## [2.0.2](https://github.com/cailloumajor/opcua-proxy/compare/v2.0.1...v2.0.2) (2022-07-13)


### Bug Fixes

* **deps:** update dependency golang to v1.18.4 ([8ac8af4](https://github.com/cailloumajor/opcua-proxy/commit/8ac8af4bd2f64d8a0f0abf8629087b1ffb63ef23))
* do not set line protocol encoder precision ([0f3b05d](https://github.com/cailloumajor/opcua-proxy/commit/0f3b05d9bc295e22db56c280ec32b19ceffc1d6d))

## [2.0.1](https://github.com/cailloumajor/opcua-proxy/compare/v2.0.0...v2.0.1) (2022-07-06)


### Bug Fixes

* sort tags ([5f1851e](https://github.com/cailloumajor/opcua-proxy/commit/5f1851ef2979cc8da29f68919002dcb6c128f3d3))

## [2.0.0](https://github.com/cailloumajor/opcua-proxy/compare/v1.1.1...v2.0.0) (2022-07-05)


### ⚠ BREAKING CHANGES

* implement InfluxDB metrics endpoint

### Features

* implement InfluxDB metrics endpoint ([1bc2a9c](https://github.com/cailloumajor/opcua-proxy/commit/1bc2a9c6d42619727e7e737ff41e2623971f613d))

## [1.1.1](https://github.com/cailloumajor/opcua-proxy/compare/v1.1.0...v1.1.1) (2022-06-16)


### Bug Fixes

* **deps:** update module github.com/gopcua/opcua to v0.3.5 ([0cd431e](https://github.com/cailloumajor/opcua-proxy/commit/0cd431eaad52cae8dc5e7c6e2cb4a03ebacce762))

## [1.1.0](https://github.com/cailloumajor/opcua-proxy/compare/v1.0.0...v1.1.0) (2022-06-14)


### Features

* use cross-compilation to build image ([f7b493f](https://github.com/cailloumajor/opcua-proxy/commit/f7b493f7f68696581bb672d122e06194718ba587))

## [1.0.0](https://github.com/cailloumajor/opcua-proxy/compare/v0.7.4...v1.0.0) (2022-06-14)


### Features

* get nodes to read from HTTP ([d4c0c3b](https://github.com/cailloumajor/opcua-proxy/commit/d4c0c3bca721e8df3c0f741d722590095ad21f3a))


### Bug Fixes

* **deps:** update module github.com/avast/retry-go/v4 to v4.1.0 ([754b4d1](https://github.com/cailloumajor/opcua-proxy/commit/754b4d1cbcab501c51bc4c4b385e9f9adef5f43a))


### Miscellaneous Chores

* release 1.0.0 ([fabeab9](https://github.com/cailloumajor/opcua-proxy/commit/fabeab9b125a4faafde94cc6ee5453c0bce83f72))

## [0.7.4](https://github.com/cailloumajor/opcua-proxy/compare/v0.7.3...v0.7.4) (2022-06-02)


### Bug Fixes

* **deps:** update dependency golang to v1.18.3 ([2330967](https://github.com/cailloumajor/opcua-proxy/commit/23309671b729f62453bed461b444eab968346d25))

### [0.7.3](https://github.com/cailloumajor/opcua-proxy/compare/v0.7.2...v0.7.3) (2022-05-18)


### Bug Fixes

* **deps:** update module github.com/avast/retry-go/v4 to v4.0.5 ([dfaeced](https://github.com/cailloumajor/opcua-proxy/commit/dfaeced2f873775e90644f8c4773e5a66adb8a4e))
* **deps:** update module github.com/go-kit/log to v0.2.1 ([218a893](https://github.com/cailloumajor/opcua-proxy/commit/218a893e7fc5db689675bd9d84c40daa279d116d))

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
