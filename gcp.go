/*
 * Copyright (c) 2026 Brandon Miller-Mumford.
 * Licensed under the MIT License; see LICENSE.
 */
package atlas

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/ORBTR/aether"
)

// Compile-time checks: GCPPlatform satisfies both PlatformInfo and BindAddressResolver.
var _ PlatformInfo = (*GCPPlatform)(nil)
var _ aether.BindAddressResolver = (*GCPPlatform)(nil)

// GCPPlatform reads Google Cloud environment variables and queries the
// instance metadata service for the external IP. Should be selected when
// GOOGLE_CLOUD_REGION or K_SERVICE is set.
type GCPPlatform struct {
	region    string
	machineID string
	appName   string

	imdsOnce        sync.Once
	imdsPublic      string
	imdsPrivateOnce sync.Once
	imdsPrivate     string
}

// NewGCPPlatform creates a GCPPlatform by reading GCP env vars at construction time.
func NewGCPPlatform() *GCPPlatform {
	return &GCPPlatform{
		region:    Normalize(os.Getenv("GOOGLE_CLOUD_REGION")),
		machineID: os.Getenv("K_REVISION"),
		appName:   os.Getenv("K_SERVICE"),
	}
}

func (g *GCPPlatform) Region() string    { return g.region }
func (g *GCPPlatform) MachineID() string { return g.machineID }
func (g *GCPPlatform) AppName() string   { return g.appName }
func (g *GCPPlatform) IsCloud() bool     { return true }
func (g *GCPPlatform) Provider() string  { return "gcp" }

// PrivateIP queries the GCP metadata service for the primary NIC's
// internal IP. Cached after first success. Returns "" on failure —
// callers fall through to interface enumeration.
func (g *GCPPlatform) PrivateIP() string {
	g.imdsPrivateOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		g.imdsPrivate = fetchGCPInternalIP(ctx)
	})
	return g.imdsPrivate
}

// PublicIP queries metadata.google.internal for the external IP.
// Cached after first success. Returns "" on Cloud Run (no external IP)
// and on any metadata-service error.
func (g *GCPPlatform) PublicIP() string {
	g.imdsOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		g.imdsPublic = fetchGCPExternalIP(ctx)
	})
	return g.imdsPublic
}

// EnvTags surfaces GCP runtime identifiers for reach provenance.
func (g *GCPPlatform) EnvTags() map[string]string {
	tags := map[string]string{"provider": "gcp"}
	if g.appName != "" {
		tags["gcp:service"] = g.appName
	}
	if g.machineID != "" {
		tags["gcp:revision"] = g.machineID
	}
	if g.region != "" {
		tags["gcp:region"] = g.region
	}
	return tags
}

// PreferredSources: IMDS is authoritative; STUN/TURN tends to be the only
// workable option on Cloud Run (no external IP per instance).
func (g *GCPPlatform) PreferredSources() []AddressSource {
	return []AddressSource{
		SourceIMDS,
		SourceDNS,
		SourceInterface,
		SourceSTUN,
		SourceTURN,
	}
}

// fetchGCPInternalIP queries the GCP metadata service for the primary
// NIC's internal (VPC) IPv4 address. Returns "" on any failure.
func fetchGCPInternalIP(ctx context.Context) string {
	client := &http.Client{Timeout: 2 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"http://metadata.google.internal/computeMetadata/v1/instance/network-interfaces/0/ip", nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Metadata-Flavor", "Google")
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			_ = resp.Body.Close()
		}
		return ""
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	return string(body)
}

// fetchGCPExternalIP queries the GCP metadata service for the primary NIC's
// first access-config external IP. Returns "" on any failure.
func fetchGCPExternalIP(ctx context.Context) string {
	client := &http.Client{Timeout: 2 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"http://metadata.google.internal/computeMetadata/v1/instance/network-interfaces/0/access-configs/0/external-ip", nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Metadata-Flavor", "Google")
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			_ = resp.Body.Close()
		}
		return ""
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	return string(body)
}

func (g *GCPPlatform) ResolveUDPBind(port int) string {
	return fmt.Sprintf(":%d", port)
}

func (g *GCPPlatform) ResolveUDPBindDualStack(port int) (string, string) {
	return "", "" // GCP: single dual-stack socket is sufficient
}

func (g *GCPPlatform) OpenUDPListener(network, address string) (*net.UDPConn, error) {
	addr, err := net.ResolveUDPAddr(network, address)
	if err != nil {
		return nil, err
	}
	return net.ListenUDP(network, addr)
}
