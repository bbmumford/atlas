/*
 * Copyright (c) 2026 Brandon Miller-Mumford.
 * Licensed under the MIT License; see LICENSE.
 */

package atlas

import (
	"testing"

	"github.com/bbmumford/loom/ports"
)

// TestMapAddressSourceSpellingDivergence is a REGRESSION LOCK for the
// host->loom AddressSource spelling map. host and loom independently declare
// the same enum with three DIFFERENT spellings — host "env"/"iface"/"k8s" are
// loom "platform-env"/"interface"/"k8s-downward". A bare string cast (the obvious, wrong
// implementation) silently emits a source string loom's PreferredSources
// ordering never matches, so the operator's preference for the
// FLY_PUBLIC_IP-derived source is dropped precisely on the cloud nodes that
// need it — the same address-advertisement class as the noise-UDP regression
// (#O-281). This is latent today (loom/node does not yet consume
// PreferredSources) but must be correct before it goes live, so the guard
// goes in the code at the site (#R-480) rather than a prose warning.
func TestMapAddressSourceSpellingDivergence(t *testing.T) {
	// The two genuinely-diverging spellings MUST be translated, not cast.
	if got := mapAddressSource(SourcePlatformEnv); got != ports.SourcePlatformEnv {
		t.Errorf("host %q must map to loom %q, got %q — a bare cast leaves it %q",
			SourcePlatformEnv, ports.SourcePlatformEnv, got, ports.AddressSource(SourcePlatformEnv))
	}
	if got := mapAddressSource(SourceInterface); got != ports.SourceInterface {
		t.Errorf("host %q must map to loom %q, got %q — a bare cast leaves it %q",
			SourceInterface, ports.SourceInterface, got, ports.AddressSource(SourceInterface))
	}
	if got := mapAddressSource(SourceK8sDownward); got != ports.SourceK8sDownward {
		t.Errorf("host %q must map to loom %q, got %q — a bare cast leaves it %q",
			SourceK8sDownward, ports.SourceK8sDownward, got, ports.AddressSource(SourceK8sDownward))
	}

	// Prove each divergence is real: a bare cast does NOT equal loom's value. If
	// host and loom ever converge a spelling this assertion fires, signalling the
	// map can be simplified — a deliberate tripwire, not a failure of correctness.
	if ports.AddressSource(SourcePlatformEnv) == ports.SourcePlatformEnv {
		t.Errorf("host/loom spellings for the platform-env source have converged (%q) — "+
			"the bare-cast hazard this map guards is gone; simplify mapAddressSource", SourcePlatformEnv)
	}
	if ports.AddressSource(SourceInterface) == ports.SourceInterface {
		t.Errorf("host/loom spellings for the interface source have converged (%q) — "+
			"the bare-cast hazard this map guards is gone; simplify mapAddressSource", SourceInterface)
	}
	if ports.AddressSource(SourceK8sDownward) == ports.SourceK8sDownward {
		t.Errorf("host/loom spellings for the k8s-downward source have converged (%q) — "+
			"the bare-cast hazard this map guards is gone; simplify mapAddressSource", SourceK8sDownward)
	}

	// Shared spellings pass through unchanged.
	for _, s := range []AddressSource{SourceIMDS, SourceDNS, SourceSTUN, SourceUPnP} {
		if got := mapAddressSource(s); string(got) != string(s) {
			t.Errorf("shared spelling %q must pass through unchanged, got %q", s, got)
		}
	}

	// host-only sources loom does not enumerate pass through verbatim rather
	// than being dropped, so loom still sees the full ordering intent.
	for _, s := range []AddressSource{SourceStatic, SourceTURN, SourceReflection} {
		if got := mapAddressSource(s); string(got) != string(s) {
			t.Errorf("host-only source %q must pass through verbatim, got %q", s, got)
		}
	}
}
