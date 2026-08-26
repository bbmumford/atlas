/*
 * Copyright (c) 2026 HSTLES / ORBTR Pty Ltd. All Rights Reserved.
 * Queries: licensing@hstles.com
 */
package atlas

import (
	"fmt"
	"net"
	"os"

	"github.com/ORBTR/aether"
)

// Compile-time checks: FlyPlatform satisfies both PlatformInfo and BindAddressResolver.
var _ PlatformInfo = (*FlyPlatform)(nil)
var _ aether.BindAddressResolver = (*FlyPlatform)(nil)

// FlyPlatform reads Fly.io environment variables.
// Should be selected when FLY_REGION is set.
type FlyPlatform struct {
	region    string
	machineID string
	appName   string
	publicIP  string
}

// NewFlyPlatform creates a FlyPlatform by reading Fly.io env vars at construction time.
func NewFlyPlatform() *FlyPlatform {
	return &FlyPlatform{
		region:    Normalize(os.Getenv("FLY_REGION")),
		machineID: os.Getenv("FLY_MACHINE_ID"),
		appName:   os.Getenv("FLY_APP_NAME"),
		publicIP:  os.Getenv("FLY_PUBLIC_IP"),
	}
}

func (f *FlyPlatform) Region() string    { return f.region }
func (f *FlyPlatform) MachineID() string { return f.machineID }
func (f *FlyPlatform) AppName() string   { return f.appName }
func (f *FlyPlatform) PublicIP() string  { return f.publicIP }
func (f *FlyPlatform) IsCloud() bool     { return true }
func (f *FlyPlatform) Provider() string  { return "fly" }

// PrivateIP returns the 6PN fdaa: mesh IP from Fly's FLY_PRIVATE_IP env
// var, set on every Fly machine at boot. Stable for the machine's
// lifetime. Used by the mesh node runtime for ourOrigin detection.
func (f *FlyPlatform) PrivateIP() string { return os.Getenv("FLY_PRIVATE_IP") }

func (f *FlyPlatform) ResolveUDPBind(port int) string {
	return fmt.Sprintf("fly-global-services:%d", port)
}

func (f *FlyPlatform) ResolveUDPBindDualStack(port int) (string, string) {
	// Fly: fly-global-services binds IPv4 for public traffic.
	// [::] binds IPv6 for same-org fdaa: private traffic.
	return fmt.Sprintf("fly-global-services:%d", port), fmt.Sprintf("[::]:%d", port)
}

// OpenUDPListener binds UDP on Fly.io using fly-global-services.
//
// Per Fly docs, apps MUST bind to fly-global-services for inbound public UDP.
// 0.0.0.0 causes Linux to use incorrect source addresses in replies.
// Go resolves fly-global-services to the machine's private IP (172.19.x.x)
// before the bind syscall — this is correct, as Fly's proxy delivers to that IP.
func (f *FlyPlatform) OpenUDPListener(network, address string) (*net.UDPConn, error) {
	pc, err := net.ListenPacket(network, address)
	if err != nil {
		return nil, err
	}
	conn, ok := pc.(*net.UDPConn)
	if !ok {
		pc.Close()
		return nil, fmt.Errorf("fly: ListenPacket returned %T, not *net.UDPConn", pc)
	}
	return conn, nil
}

// EnvTags surfaces Fly runtime identifiers for reach record provenance.
func (f *FlyPlatform) EnvTags() map[string]string {
	tags := map[string]string{"provider": "fly"}
	if f.appName != "" {
		tags["fly:app"] = f.appName
	}
	if f.machineID != "" {
		tags["fly:machine"] = f.machineID
	}
	if f.region != "" {
		tags["fly:region"] = f.region
	}
	if allocID := os.Getenv("FLY_ALLOC_ID"); allocID != "" {
		tags["fly:alloc"] = allocID
	}
	return tags
}

// PreferredSources orders discovery so PlatformEnv (FLY_PUBLIC_IP) wins
// over DNS and interface enumeration. STUN is intentionally absent — on
// Fly the egress IP differs from the ingress anycast IP.
func (f *FlyPlatform) PreferredSources() []AddressSource {
	return []AddressSource{
		SourcePlatformEnv,
		SourceDNS,
		SourceInterface,
	}
}
