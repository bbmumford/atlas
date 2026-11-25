/*
 * Copyright (c) 2026 Brandon Miller-Mumford.
 * Licensed under the MIT License; see LICENSE.
 */
package atlas

import "github.com/bbmumford/loom/ports"

// MaxMachinesPerApp is the per-app machine-count ceiling assumed by the
// per-machine UDP port allocator. Each machine gets one of N consecutive
// ports starting at the configured base; the endpoint's fly.toml must
// declare all N ports as UDP services. 16 covers current fleet sizing
// (typically 1-3 machines per app) with headroom and keeps the port
// range small enough to be readable in fly.toml.
const MaxMachinesPerApp = 16

// AddressSource is loom's address-source type, so a detected platform's
// PreferredSources is the slice loom's node reads with no conversion.
type AddressSource = ports.AddressSource

// The sources loom enumerates carry loom's spellings; static, turn and
// reflection are sources atlas's detectors name that loom does not enumerate.
const (
	SourceStatic      AddressSource = "static"
	SourcePlatformEnv               = ports.SourcePlatformEnv
	SourceDNS                       = ports.SourceDNS
	SourceIMDS                      = ports.SourceIMDS
	SourceK8sDownward               = ports.SourceK8sDownward
	SourceInterface                 = ports.SourceInterface
	SourceSTUN                      = ports.SourceSTUN
	SourceTURN        AddressSource = "turn"
	SourceReflection  AddressSource = "reflection"
	SourceUPnP                      = ports.SourceUPnP
)

// PlatformInfo is loom's host-platform seam, which every detector here
// implements, so Detect's result goes into loom's node.Config with no adapter.
// Implementations read env vars or cloud metadata APIs; consumers never call
// os.Getenv for platform detection. It also satisfies
// aether.BindAddressResolver: the transport layer receives a PlatformInfo
// directly without type assertions.
type PlatformInfo = ports.PlatformInfo
