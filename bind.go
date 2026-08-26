/*
 * Copyright (c) 2026 HSTLES / ORBTR Pty Ltd. All Rights Reserved.
 * Queries: licensing@hstles.com
 */
package host

// Deprecated: FlyBindResolver is replaced by FlyPlatform which satisfies
// both PlatformInfo and aether.BindAddressResolver.
// Use host.Detect() or host.NewFlyPlatform() instead.
type FlyBindResolver = FlyPlatform

// Deprecated: DefaultBindResolver is replaced by host.Detect().
// Use host.Detect() which returns a PlatformInfo that also satisfies
// aether.BindAddressResolver.
func DefaultBindResolver() PlatformInfo {
	return Detect()
}
