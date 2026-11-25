/*
 * Copyright (c) 2026 Brandon Miller-Mumford.
 * Licensed under the MIT License; see LICENSE.
 */

package atlas

import "github.com/bbmumford/loom/ports"

// LoomPlatform returns p unchanged, or Detect() when p is nil. atlas's platforms
// are loom's ports.PlatformInfo, so a composition root that wraps Detect's
// result in it hands loom's node.Config the detected platform itself.
func LoomPlatform(p PlatformInfo) ports.PlatformInfo {
	if p == nil {
		return Detect()
	}
	return p
}
