/*
 * Copyright (c) 2026 HSTLES / ORBTR Pty Ltd. All Rights Reserved.
 * Queries: licensing@hstles.com
 */
package host

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
	if os.Getenv("FLY_REGION") != "" || os.Getenv("FLY_MACHINE_ID") != "" {
		return NewFlyPlatform()
	}
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" || os.Getenv("POD_NAME") != "" {
		return NewK8sPlatform()
	}
	if os.Getenv("AWS_REGION") != "" || os.Getenv("ECS_TASK_ID") != "" {
		return NewAWSPlatform()
	}
	if os.Getenv("GOOGLE_CLOUD_REGION") != "" || os.Getenv("K_SERVICE") != "" {
		return NewGCPPlatform()
	}
	return NewDevPlatform("")
}
