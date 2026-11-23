/*
 * Copyright (c) 2026 Brandon Miller-Mumford.
 * Licensed under the MIT License; see LICENSE.
 */
package atlas

import (
	"fmt"
	"net"

	"github.com/ORBTR/aether"
)

// Compile-time checks: DevPlatform satisfies both PlatformInfo and BindAddressResolver.
var _ PlatformInfo = (*DevPlatform)(nil)
var _ aether.BindAddressResolver = (*DevPlatform)(nil)

// DevPlatform provides sensible defaults for local development and bare metal.
type DevPlatform struct {
	region string // Can be set via config; defaults to "local"
	// defaulted records that region is the "local" placeholder, which KnownRegion
	// reports as no region.
	defaulted bool
}

// NewDevPlatform creates a DevPlatform with an optional region override.
// If region is empty, defaults to "local".
func NewDevPlatform(region string) *DevPlatform {
	if region == "" {
		return &DevPlatform{region: "local", defaulted: true}
	}
	return &DevPlatform{region: region}
}

func (d *DevPlatform) Region() string    { return d.region }
func (d *DevPlatform) MachineID() string { return "" }
func (d *DevPlatform) AppName() string   { return "" }
func (d *DevPlatform) PublicIP() string  { return "" }
func (d *DevPlatform) IsCloud() bool     { return false }
func (d *DevPlatform) Provider() string  { return "dev" }

// PrivateIP returns "" on dev — there's no authoritative platform source
// for the private IP, so callers fall through to interface enumeration
// (Runtime.discoverPrivateIP walks net.InterfaceAddrs for an fd00::/8
// ULA or first routable RFC1918 IPv4).
func (d *DevPlatform) PrivateIP() string { return "" }

func (d *DevPlatform) ResolveUDPBind(port int) string {
	return fmt.Sprintf(":%d", port)
}

func (d *DevPlatform) ResolveUDPBindDualStack(port int) (string, string) {
	return "", "" // dev: single dual-stack socket is sufficient
}

func (d *DevPlatform) OpenUDPListener(network, address string) (*net.UDPConn, error) {
	addr, err := net.ResolveUDPAddr(network, address)
	if err != nil {
		return nil, err
	}
	return net.ListenUDP(network, addr)
}

// EnvTags returns a single "provider=dev" tag on bare metal / dev machines.
func (d *DevPlatform) EnvTags() map[string]string {
	return map[string]string{"provider": "dev"}
}

// PreferredSources: on bare metal the interface list is usually the richest
// source; STUN tells us the reflexive address; UPnP is a last resort for
// home networks without admin access.
func (d *DevPlatform) PreferredSources() []AddressSource {
	return []AddressSource{
		SourceInterface,
		SourceSTUN,
		SourceUPnP,
		SourceTURN,
	}
}
