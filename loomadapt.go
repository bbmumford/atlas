/*
 * Copyright (c) 2026 Brandon Miller-Mumford.
 * Licensed under the MIT License; see LICENSE.
 */

package atlas

import (
	"net"

	"github.com/bbmumford/loom/ports"
)

// LoomPlatform adapts a host.PlatformInfo (Fly/AWS/GCP/k8s/dev detectors)
// to loom's ports.PlatformInfo seam. Every method forwards verbatim; only
// PreferredSources needs a type map because loom declares its own
// AddressSource. Wire it at endpoint boot:
//
//	cfg := node.Config{Platform: host.LoomPlatform(host.Detect()), …}
//
// ⚠ Config.Platform nil is NOT auto-detect in loom (it falls back to a
// generic dev platform with no Fly UDP-bind/6PN/region semantics) — Fly
// deployments MUST pass this adapter or they silently lose the machinery
// that makes the mesh work on Fly.
func LoomPlatform(p PlatformInfo) ports.PlatformInfo {
	if p == nil {
		p = Detect()
	}
	return loomPlatform{p}
}

type loomPlatform struct{ p PlatformInfo }

func (a loomPlatform) Region() string    { return a.p.Region() }
func (a loomPlatform) MachineID() string { return a.p.MachineID() }
func (a loomPlatform) AppName() string   { return a.p.AppName() }
func (a loomPlatform) PublicIP() string  { return a.p.PublicIP() }
func (a loomPlatform) PrivateIP() string { return a.p.PrivateIP() }

func (a loomPlatform) ResolveUDPBind(port int) string { return a.p.ResolveUDPBind(port) }

func (a loomPlatform) ResolveUDPBindDualStack(port int) (string, string) {
	return a.p.ResolveUDPBindDualStack(port)
}

func (a loomPlatform) OpenUDPListener(network, address string) (*net.UDPConn, error) {
	return a.p.OpenUDPListener(network, address)
}

func (a loomPlatform) IsCloud() bool              { return a.p.IsCloud() }
func (a loomPlatform) Provider() string           { return a.p.Provider() }
func (a loomPlatform) EnvTags() map[string]string { return a.p.EnvTags() }

func (a loomPlatform) PreferredSources() []ports.AddressSource {
	src := a.p.PreferredSources()
	out := make([]ports.AddressSource, len(src))
	for i, s := range src {
		out[i] = mapAddressSource(s)
	}
	return out
}

// mapAddressSource translates a host AddressSource spelling to loom's. Three
// spellings genuinely diverge — host "env"/"iface"/"k8s" are loom
// "platform-env"/"interface"/"k8s-downward" — so a bare string cast would
// silently emit a source loom's ordering never matches, dropping the operator's
// preference for exactly the FLY_PUBLIC_IP-derived source that matters on a
// cloud node. The shared spellings (imds/dns/stun/upnp) and host-only sources
// loom does not enumerate (static/turn/reflection) pass through verbatim rather
// than being dropped, so loom still sees the full ordering intent. (#O-281.)
func mapAddressSource(s AddressSource) ports.AddressSource {
	switch s {
	case SourcePlatformEnv: // host "env"
		return ports.SourcePlatformEnv // loom "platform-env"
	case SourceInterface: // host "iface"
		return ports.SourceInterface // loom "interface"
	case SourceK8sDownward: // host "k8s"
		return ports.SourceK8sDownward // loom "k8s-downward"
	default:
		return ports.AddressSource(s)
	}
}
