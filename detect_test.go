/*
 * Copyright (c) 2026 HSTLES / ORBTR Pty Ltd. All Rights Reserved.
 * Queries: licensing@hstles.com
 */
package atlas

import (
	"os"
	"testing"
)

// detectionEnv is every variable Detect and the platforms' Region read.
var detectionEnv = []string{
	"FLY_REGION", "FLY_MACHINE_ID", "KUBERNETES_SERVICE_HOST", "POD_NAME", "K8S_REGION",
	"AWS_REGION", "ECS_TASK_ID", "GOOGLE_CLOUD_REGION", "K_SERVICE",
}

// clearDetectionEnv unsets every detection variable for the test, restoring them
// afterwards.
func clearDetectionEnv(t *testing.T) {
	t.Helper()
	for _, name := range detectionEnv {
		t.Setenv(name, "")
		os.Unsetenv(name)
	}
}

func TestDetectWithFallbackRegionAndKnownRegion(t *testing.T) {
	for _, c := range []struct {
		name      string
		env       map[string]string
		fallback  string
		region    string // DetectWith(...).Region()
		known     string // KnownRegion(DetectWith(...))
		detectReg string // Detect().Region(), unchanged by the options
	}{
		{"no platform, no fallback", nil, "", "local", "", "local"},
		{"no platform, fallback", nil, "dc-syd1", "dc-syd1", "dc-syd1", "local"},
		{"fly region wins over the fallback", map[string]string{"FLY_REGION": "SYD"}, "dc-syd1", "syd", "syd", "syd"},
		{"fly machine with no region takes the fallback", map[string]string{"FLY_MACHINE_ID": "m1"}, "dc-syd1", "dc-syd1", "dc-syd1", ""},
		{"kubernetes with no hint, no fallback", map[string]string{"KUBERNETES_SERVICE_HOST": "10.0.0.1"}, "", "unknown", "", "unknown"},
		{"kubernetes with no hint takes the fallback", map[string]string{"KUBERNETES_SERVICE_HOST": "10.0.0.1"}, "dc-syd1", "dc-syd1", "dc-syd1", "unknown"},
		{"kubernetes hint wins over the fallback", map[string]string{"KUBERNETES_SERVICE_HOST": "10.0.0.1", "K8S_REGION": "us-east-1"}, "dc-syd1", "iad", "iad", "iad"},
		{"aws region", map[string]string{"AWS_REGION": "eu-west-2"}, "dc-syd1", "lhr", "lhr", "lhr"},
		{"aws task with no region takes the fallback", map[string]string{"ECS_TASK_ID": "t1"}, "dc-syd1", "dc-syd1", "dc-syd1", ""},
		{"gcp service with no region takes the fallback", map[string]string{"K_SERVICE": "s1"}, "dc-syd1", "dc-syd1", "dc-syd1", ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			clearDetectionEnv(t)
			for k, v := range c.env {
				t.Setenv(k, v)
			}
			p := DetectWith(WithFallbackRegion(c.fallback))
			if got := p.Region(); got != c.region {
				t.Errorf("Region() = %q, want %q", got, c.region)
			}
			if got := KnownRegion(p); got != c.known {
				t.Errorf("KnownRegion = %q, want %q", got, c.known)
			}
			if got := Detect().Region(); got != c.detectReg {
				t.Errorf("Detect().Region() = %q, want %q", got, c.detectReg)
			}
		})
	}
}

func TestKnownRegionOfAnExplicitDevRegion(t *testing.T) {
	if got := KnownRegion(NewDevPlatform("local")); got != "local" {
		t.Errorf("a region given as \"local\" is a region: got %q", got)
	}
	if got := KnownRegion(NewDevPlatform("")); got != "" {
		t.Errorf("the defaulted placeholder is no region: got %q", got)
	}
}
