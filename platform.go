/*
 * Copyright (c) 2026 HSTLES / ORBTR Pty Ltd. All Rights Reserved.
 * Queries: licensing@hstles.com
 */
package host

import "net"

// MaxMachinesPerApp is the per-app machine-count ceiling assumed by the
// per-machine UDP port allocator. Each machine gets one of N consecutive
// ports starting at the configured base; the endpoint's fly.toml must
// declare all N ports as UDP services. 16 covers current fleet sizing
// (typically 1-3 machines per app) with headroom and keeps the port
// range small enough to be readable in fly.toml.
const MaxMachinesPerApp = 16

// AddressSource identifies the discovery source of a reach address when
// PlatformInfo.PreferredSources is consulted by the reach publisher.
// String values match the reach package's `AddressSource.String()` so
// consumers don't need a conversion layer.
type AddressSource string

const (
	SourceStatic       AddressSource = "static"
	SourcePlatformEnv  AddressSource = "env"
	SourceDNS          AddressSource = "dns"
	SourceIMDS         AddressSource = "imds"
	SourceK8sDownward  AddressSource = "k8s"
	SourceInterface    AddressSource = "iface"
	SourceSTUN         AddressSource = "stun"
	SourceTURN         AddressSource = "turn"
	SourceReflection   AddressSource = "reflection"
	SourceUPnP         AddressSource = "upnp"
)

// PlatformInfo provides platform-specific metadata to library packages.
// Implementations read env vars or cloud metadata APIs; consumers never
// call os.Getenv for platform detection.
//
// PlatformInfo also satisfies aether.BindAddressResolver — the transport
// layer receives a PlatformInfo directly without type assertions.
type PlatformInfo interface {
	// Region returns the normalized 3-letter region code (e.g., "iad", "syd").
	// Returns "unknown" if region cannot be determined.
	Region() string

	// MachineID returns the platform-specific machine/instance identifier.
	// Returns "" if not available (e.g., local dev).
	MachineID() string

	// AppName returns the application/service name as known to the platform.
	// Returns "" if not available.
	AppName() string

	// PublicIP returns the node's public IP as known to the platform.
	// Returns "" if not available (caller should fall back to STUN).
	PublicIP() string

	// PrivateIP returns the node's private mesh IP as known to the platform.
	// On Fly this is FLY_PRIVATE_IP (the 6PN fdaa: address). On AWS/GCP this
	// is the primary VPC IP from IMDS. On k8s it is POD_IP (downward API).
	// On dev/bare-metal it is the first non-loopback interface address.
	// Returns "" if not available; callers should fall back to interface
	// enumeration. Consumed by the mesh node runtime for ourOrigin
	// detection (used by the cross-org forwarder to distinguish same-org
	// from cross-org peers) and by the reach publisher's interface
	// discoverer to filter to the routable private address.
	PrivateIP() string

	// ResolveUDPBind returns the correct UDP bind address for a given port.
	// On Fly.io returns "fly-global-services:<port>".
	// On other platforms returns ":<port>" (bind all interfaces).
	ResolveUDPBind(port int) string

	// ResolveUDPBindDualStack returns separate IPv4 and IPv6 bind addresses.
	// Returns ("", "") if dual-stack isn't needed (single socket handles both).
	// When both are non-empty, two sockets are opened on the same port.
	ResolveUDPBindDualStack(port int) (ipv4Addr, ipv6Addr string)

	// OpenUDPListener creates a UDP socket bound to the platform-appropriate address.
	// Platforms that need special bind semantics (e.g., Fly.io's eBPF proxy)
	// override the default net.ResolveUDPAddr + net.ListenUDP behavior here.
	OpenUDPListener(network, address string) (*net.UDPConn, error)

	// IsCloud returns true if running on a cloud platform (not local dev).
	IsCloud() bool

	// Provider returns the platform name ("fly", "aws", "gcp", "k8s", "dev").
	Provider() string

	// EnvTags returns provenance tags the reach publisher attaches to
	// addresses it emits on this platform (e.g., "fly:app=X", "fly:machine=Y",
	// "k8s:pod=abc"). Empty map is acceptable — provider names already convey
	// most context. Read once at publisher startup.
	EnvTags() map[string]string

	// PreferredSources returns the ordered list of reach address sources
	// the publisher should prefer on this platform. Earlier entries win in
	// the discoverer deduplication pass. Empty slice = let the publisher
	// decide from its defaults.
	//
	// Rationale per provider:
	//   Fly  : PlatformEnv (FLY_PUBLIC_IP) > DNS > Interface (IPv6 fdaa) — STUN banned
	//   AWS  : IMDS > DNS > Interface > STUN
	//   GCP  : IMDS > DNS > Interface > STUN (Cloud Run tends to fall through to STUN/TURN)
	//   k8s  : K8sDownward > Interface > STUN
	//   dev  : Interface > STUN > UPnP (STUN is PRIMARY on bare metal / dev)
	PreferredSources() []AddressSource
}
