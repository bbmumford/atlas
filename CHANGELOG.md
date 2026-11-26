# Changelog

## [0.2.0](https://github.com/bbmumford/atlas/compare/v0.1.0...v0.2.0) (2026-11-26)


### ⚠ BREAKING CHANGES

* LoomPlatform is removed; pass Detect() (or a DetectWith result) where it was passed.
* SourcePlatformEnv, SourceInterface and SourceK8sDownward change value, and PlatformInfo is loom's interface type.

### Features

* atlas's PlatformInfo and AddressSource are loom's ports types, and every detector names sources in loom's spellings ([b617cd9](https://github.com/bbmumford/atlas/commit/b617cd9c16b4381da50f27eb0ccfa7fcc350fad7))
* LoomPlatform is removed, since Detect's platform is loom's ports.PlatformInfo itself ([ef4983b](https://github.com/bbmumford/atlas/commit/ef4983bfae57d995e86c31748d752a23690c673d))
