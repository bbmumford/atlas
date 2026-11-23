# atlas

atlas detects the hosting platform a process runs on (Fly.io, Kubernetes, AWS, GCP, or a development or bare-metal host) from its environment and cloud metadata, and answers the questions a mesh node asks about it: region, machine id, app name, public and private IP, how to bind a UDP socket, and which address-discovery sources to prefer.

Kubernetes is checked before AWS and GCP, because managed clusters run on cloud nodes that also set the cloud region variables; a host none of the platforms recognise is a `DevPlatform`.

## Use

```sh
go get github.com/bbmumford/atlas
```

```go
p := atlas.Detect()                 // or atlas.DetectWith(atlas.WithFallbackRegion("syd"))
region := p.Region()                // Fly-style region code; atlas.Normalize maps AWS and GCP codes
bind := p.ResolveUDPBind(4242)      // on Fly: "fly-global-services:4242"
conn, err := p.OpenUDPListener("udp", bind)

node := atlas.LoomPlatform(p)       // the same platform as loom's ports.PlatformInfo
```

`PlatformInfo` is the interface every platform implements (`Region`, `MachineID`, `AppName`, `PublicIP`, `PrivateIP`, `ResolveUDPBind`, `ResolveUDPBindDualStack`, `OpenUDPListener`, `IsCloud`, `Provider`, `EnvTags`, `PreferredSources`). `KnownRegion` returns a platform's region only when the platform knows one.

## Dependencies

- `github.com/bbmumford/loom/ports`: `LoomPlatform` adapts an atlas platform
  to loom's `ports.PlatformInfo`.
- `github.com/ORBTR/aether`: the bind resolvers implement aether's
  `BindAddressResolver`.
- `go.mod` replaces aether and loom with the sibling checkouts `../aether` and
  `../loom`.

## Used by

Importers in the ORBTR estate:

- `github.com/ORBTR/Library/endpoint`: every endpoint detects its platform at
  boot.
- `github.com/ORBTR/cortex/observability`.
- `github.com/ORBTR/agent/internal/core/wire`.

## Platforms and build

Pure Go, no cgo and no build tags. The module's `go` directive is 1.25.0.

## Licence

MIT. Copyright (c) 2026 Brandon Miller-Mumford. The full text is in
[LICENSE](LICENSE).

## Contributing

[AGENTS.md](AGENTS.md) holds the repository's two rules, for people and coding
agents alike: no temporary documents in the repository, and conventional commit
messages that stand on their own. [CLAUDE.md](CLAUDE.md) imports it for Claude
Code.
