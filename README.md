# atlas

Detects the hosting platform a process runs on (Fly.io, Kubernetes, AWS, GCP, or a development or bare-metal host) from its environment and cloud metadata, and answers the questions a mesh node asks about it: region, machine id, app name, public and private IP, how to bind a UDP socket, and which address-discovery sources to prefer.

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

## Consumers

ORBTR library, ai cortex and xyz, the fleet.orbtr.ai endpoint, and io agent and tools.
