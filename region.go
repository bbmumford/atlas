/*
 * Copyright (c) 2026 Brandon Miller-Mumford.
 * Licensed under the MIT License; see LICENSE.
 */
package atlas

import (
	"strings"
)

// Deprecated: Region is replaced by PlatformInfo.Region().
// Use host.Detect().Region() or inject PlatformInfo via config.
// Kept for backward compatibility — calls Detect().Region() internally.
func Region() string {
	return Detect().Region()
}

// regionMap maps provider-specific region codes (AWS, GCP) to Fly.io-style
// 3-letter IATA codes. Fly.io regions pass through unchanged since they are
// already in the target format.
var regionMap = map[string]string{
	// AWS → Fly
	"us-east-1":      "iad", // N. Virginia
	"us-east-2":      "ord", // Ohio → Chicago
	"us-west-1":      "sjc", // N. California
	"us-west-2":      "sea", // Oregon → Seattle
	"ca-central-1":   "yul", // Montreal
	"eu-west-1":      "dub", // Ireland
	"eu-west-2":      "lhr", // London
	"eu-west-3":      "cdg", // Paris
	"eu-central-1":   "fra", // Frankfurt
	"eu-north-1":     "arn", // Stockholm
	"ap-southeast-1": "sin", // Singapore
	"ap-southeast-2": "syd", // Sydney
	"ap-northeast-1": "nrt", // Tokyo
	"ap-northeast-2": "icn", // Seoul
	"ap-south-1":     "bom", // Mumbai
	"sa-east-1":      "gru", // São Paulo

	// GCP → Fly
	"us-east1":             "iad", // South Carolina
	"us-central1":          "ord", // Iowa
	"us-west1":             "sea", // Oregon
	"europe-west1":         "ams", // Belgium
	"europe-west2":         "lhr", // London
	"europe-west3":         "fra", // Frankfurt
	"asia-northeast1":      "nrt", // Tokyo
	"asia-southeast1":      "sin", // Singapore
	"australia-southeast1": "syd", // Sydney
}

// Normalize maps provider-specific region codes to a standard set (Fly.io-style
// 3-letter IATA codes). Fly.io regions pass through unchanged. AWS and GCP
// region codes are mapped to their nearest Fly.io equivalent.
func Normalize(region string) string {
	region = strings.ToLower(region)
	if mapped, ok := regionMap[region]; ok {
		return mapped
	}
	return region
}
