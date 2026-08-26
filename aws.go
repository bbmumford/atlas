/*
 * Copyright (c) 2026 HSTLES / ORBTR Pty Ltd. All Rights Reserved.
 * Queries: licensing@hstles.com
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

// Compile-time checks: AWSPlatform satisfies both PlatformInfo and BindAddressResolver.
var _ PlatformInfo = (*AWSPlatform)(nil)
var _ aether.BindAddressResolver = (*AWSPlatform)(nil)

// AWSPlatform reads AWS environment variables and queries IMDSv2 for
// authoritative reachability data (public IP, availability zone).
// Should be selected when AWS_REGION is set.
type AWSPlatform struct {
	region    string
	machineID string
	appName   string

	imdsOnce         sync.Once
	imdsPublic       string
	imdsPrivateOnce  sync.Once
	imdsPrivate      string
}

// NewAWSPlatform creates an AWSPlatform by reading AWS env vars at construction time.
func NewAWSPlatform() *AWSPlatform {
	return &AWSPlatform{
		region:    Normalize(os.Getenv("AWS_REGION")),
		machineID: os.Getenv("ECS_TASK_ID"),
		appName:   os.Getenv("ECS_SERVICE_NAME"),
	}
}

func (a *AWSPlatform) Region() string    { return a.region }
func (a *AWSPlatform) MachineID() string { return a.machineID }
func (a *AWSPlatform) AppName() string   { return a.appName }
func (a *AWSPlatform) IsCloud() bool     { return true }
func (a *AWSPlatform) Provider() string  { return "aws" }

// PrivateIP returns the instance's VPC private IPv4 from IMDS. Cached on
// first lookup. Returns "" if IMDS is unreachable — caller falls back
// to interface enumeration. Used by mesh node runtime for ourOrigin
// detection.
func (a *AWSPlatform) PrivateIP() string {
	a.imdsPrivateOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		a.imdsPrivate = fetchAWSPrivateIPv4(ctx)
	})
	return a.imdsPrivate
}

// PublicIP queries IMDSv2 for the instance's public IPv4 address.
// Cached after the first successful lookup. Returns "" when IMDS is
// unreachable (VPC-only instances, locked-down IMDS, non-EC2 runtime).
func (a *AWSPlatform) PublicIP() string {
	a.imdsOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		a.imdsPublic = fetchAWSPublicIPv4(ctx)
	})
	return a.imdsPublic
}

// EnvTags surfaces AWS runtime identifiers for reach provenance.
func (a *AWSPlatform) EnvTags() map[string]string {
	tags := map[string]string{"provider": "aws"}
	if a.appName != "" {
		tags["aws:service"] = a.appName
	}
	if a.machineID != "" {
		tags["aws:task"] = a.machineID
	}
	if a.region != "" {
		tags["aws:region"] = a.region
	}
	return tags
}

// PreferredSources: IMDS is authoritative on EC2; DNS and interface come
// next; STUN is a valid fallback for VPCs without an internet gateway.
func (a *AWSPlatform) PreferredSources() []AddressSource {
	return []AddressSource{
		SourceIMDS,
		SourceDNS,
		SourceInterface,
		SourceSTUN,
	}
}

// fetchAWSPublicIPv4 queries IMDSv2 for the public IPv4 of this instance.
// Returns "" on any failure; callers should fall through to DNS/STUN.
func fetchAWSPublicIPv4(ctx context.Context) string {
	client := &http.Client{Timeout: 2 * time.Second}

	// Step 1: obtain a short-lived IMDSv2 token (PUT /latest/api/token).
	req, err := http.NewRequestWithContext(ctx, http.MethodPut,
		"http://169.254.169.254/latest/api/token", nil)
	if err != nil {
		return ""
	}
	req.Header.Set("X-aws-ec2-metadata-token-ttl-seconds", "60")
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			_ = resp.Body.Close()
		}
		return ""
	}
	tokenBytes, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil || len(tokenBytes) == 0 {
		return ""
	}

	// Step 2: GET /latest/meta-data/public-ipv4 with the token.
	req2, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"http://169.254.169.254/latest/meta-data/public-ipv4", nil)
	if err != nil {
		return ""
	}
	req2.Header.Set("X-aws-ec2-metadata-token", string(tokenBytes))
	resp2, err := client.Do(req2)
	if err != nil || resp2.StatusCode != http.StatusOK {
		if resp2 != nil {
			_ = resp2.Body.Close()
		}
		return ""
	}
	defer resp2.Body.Close()
	ipBytes, err := io.ReadAll(resp2.Body)
	if err != nil {
		return ""
	}
	return string(ipBytes)
}

// fetchAWSPrivateIPv4 queries IMDSv2 for the instance's primary VPC
// private IPv4 address. Same IMDSv2 token flow as fetchAWSPublicIPv4 —
// returns "" on any failure so the caller can fall through to interface
// enumeration.
func fetchAWSPrivateIPv4(ctx context.Context) string {
	client := &http.Client{Timeout: 2 * time.Second}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut,
		"http://169.254.169.254/latest/api/token", nil)
	if err != nil {
		return ""
	}
	req.Header.Set("X-aws-ec2-metadata-token-ttl-seconds", "60")
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			_ = resp.Body.Close()
		}
		return ""
	}
	tokenBytes, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil || len(tokenBytes) == 0 {
		return ""
	}

	req2, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"http://169.254.169.254/latest/meta-data/local-ipv4", nil)
	if err != nil {
		return ""
	}
	req2.Header.Set("X-aws-ec2-metadata-token", string(tokenBytes))
	resp2, err := client.Do(req2)
	if err != nil || resp2.StatusCode != http.StatusOK {
		if resp2 != nil {
			_ = resp2.Body.Close()
		}
		return ""
	}
	defer resp2.Body.Close()
	ipBytes, err := io.ReadAll(resp2.Body)
	if err != nil {
		return ""
	}
	return string(ipBytes)
}

func (a *AWSPlatform) ResolveUDPBind(port int) string {
	return fmt.Sprintf(":%d", port)
}

func (a *AWSPlatform) ResolveUDPBindDualStack(port int) (string, string) {
	return "", "" // AWS: single dual-stack socket is sufficient
}

func (a *AWSPlatform) OpenUDPListener(network, address string) (*net.UDPConn, error) {
	addr, err := net.ResolveUDPAddr(network, address)
	if err != nil {
		return nil, err
	}
	return net.ListenUDP(network, addr)
}
