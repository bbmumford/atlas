/*
 * Copyright (c) 2026 HSTLES / ORBTR Pty Ltd. All Rights Reserved.
 * Queries: licensing@hstles.com
 */
package atlas

import "os"

// Detect auto-detects the current platform from environment signals.
// Returns the most specific platform available, falling back to DevPlatform.
//
// Detection order is deliberate — Kubernetes is checked BEFORE AWS/GCP
// because managed clusters (EKS, GKE) run on cloud nodes that also set
// AWS_REGION / GOOGLE_CLOUD_REGION. A pod should be detected as k8s so the
// downward-API paths take precedence; cloud-specific hints are still
// available to the K8sPlatform Region() fallback.
func Detect() PlatformInfo {
	return DetectWith()
}

// DetectOption adjusts DetectWith.
type DetectOption func(*detectOptions)

type detectOptions struct {
	fallbackRegion string
}

// WithFallbackRegion names the region of a host whose platform reports none: a host
// on none of the platforms Detect recognises (a machine in a data centre, say), a
// Kubernetes pod with no region hint, or a Fly, AWS or GCP host whose region
// variable is unset. The caller reads the value from wherever it declares its
// settings. An empty region changes nothing.
func WithFallbackRegion(region string) DetectOption {
	return func(o *detectOptions) { o.fallbackRegion = region }
}

// DetectWith is Detect with options applied; with none it returns what Detect does.
func DetectWith(opts ...DetectOption) PlatformInfo {
	var o detectOptions
	for _, opt := range opts {
		opt(&o)
	}
	if os.Getenv("FLY_REGION") != "" || os.Getenv("FLY_MACHINE_ID") != "" {
		p := NewFlyPlatform()
		if p.region == "" {
			p.region = o.fallbackRegion
		}
		return p
	}
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" || os.Getenv("POD_NAME") != "" {
		p := NewK8sPlatform()
		p.fallbackRegion = o.fallbackRegion
		return p
	}
	if os.Getenv("AWS_REGION") != "" || os.Getenv("ECS_TASK_ID") != "" {
		p := NewAWSPlatform()
		if p.region == "" {
			p.region = o.fallbackRegion
		}
		return p
	}
	if os.Getenv("GOOGLE_CLOUD_REGION") != "" || os.Getenv("K_SERVICE") != "" {
		p := NewGCPPlatform()
		if p.region == "" {
			p.region = o.fallbackRegion
		}
		return p
	}
	return NewDevPlatform(o.fallbackRegion)
}

// KnownRegion returns p's region, or "" where p reports a placeholder for a region it
// does not know: the "local" of a DevPlatform built with no region and the "unknown"
// of a K8sPlatform with no region hint. A caller that records or prices by region
// reads this, so an unknown region stays empty rather than becoming a word.
func KnownRegion(p PlatformInfo) string {
	switch v := p.(type) {
	case *DevPlatform:
		if v.defaulted {
			return ""
		}
	case *K8sPlatform:
		return v.knownRegion()
	}
	return p.Region()
}
