/*
 * Copyright (c) 2026 Brandon Miller-Mumford.
 * Licensed under the MIT License; see LICENSE.
 */

package atlas

import (
	"testing"

	"github.com/bbmumford/loom/ports"
)

// TestEveryDetectorIsALoomPlatform pins that each platform atlas detects is loom's ports.PlatformInfo itself, so an
// endpoint hands Detect's result to loom's node.Config with no adapter between them.
func TestEveryDetectorIsALoomPlatform(t *testing.T) {
	for name, p := range map[string]any{
		"fly": NewFlyPlatform(), "aws": NewAWSPlatform(), "gcp": NewGCPPlatform(),
		"k8s": NewK8sPlatform(), "dev": NewDevPlatform("syd"),
	} {
		if _, ok := p.(ports.PlatformInfo); !ok {
			t.Errorf("the %s platform is not a loom ports.PlatformInfo", name)
		}
	}
}

// TestDetectorsNameSourcesInLoomsSpelling pins the three sources whose spelling loom's ordering matches by value:
// Fly's FLY_PUBLIC_IP source, the k8s downward API and interface enumeration. A detector that names them any other
// way leaves loom's node unable to honour the platform's preference for them.
func TestDetectorsNameSourcesInLoomsSpelling(t *testing.T) {
	cases := []struct {
		name string
		got  []ports.AddressSource
		want ports.AddressSource
	}{
		{"fly", toPorts(NewFlyPlatform().PreferredSources()), "platform-env"},
		{"k8s", toPorts(NewK8sPlatform().PreferredSources()), "k8s-downward"},
		{"aws", toPorts(NewAWSPlatform().PreferredSources()), "interface"},
	}
	for _, c := range cases {
		found := false
		for _, s := range c.got {
			found = found || s == c.want
		}
		if !found {
			t.Errorf("the %s platform's preferred sources %v lack loom's %q", c.name, c.got, c.want)
		}
	}
}

// toPorts reads a detector's sources by their string values, the form loom's ordering compares.
func toPorts[S ~string](in []S) []ports.AddressSource {
	out := make([]ports.AddressSource, len(in))
	for i, s := range in {
		out[i] = ports.AddressSource(s)
	}
	return out
}
