/*
 * Copyright (c) 2026 Brandon Miller-Mumford.
 * Licensed under the MIT License; see LICENSE.
 */
package atlas

import (
	"fmt"
	"net"
	"os"

	"github.com/ORBTR/aether"
)

// Compile-time checks: K8sPlatform satisfies both PlatformInfo and BindAddressResolver.
var _ PlatformInfo = (*K8sPlatform)(nil)
var _ aether.BindAddressResolver = (*K8sPlatform)(nil)

// K8sPlatform reads Kubernetes downward-API environment variables.
// Should be selected when KUBERNETES_SERVICE_HOST is set. Compatible with
// both vanilla k8s and managed variants (EKS, GKE, AKS) — those providers'
// native PlatformInfo implementations can delegate to K8sPlatform when the
// cluster environment is present.
type K8sPlatform struct {
	region    string
	nodeName  string
	namespace string
	podIP     string
	hostIP    string
	podName   string
	// fallbackRegion is DetectWith's WithFallbackRegion value, the region Region
	// reports when the pod carries no region hint.
	fallbackRegion string
}

// NewK8sPlatform reads the standard downward-API env vars (conventional
// names — the pod spec must wire these):
//
//	POD_NAME, POD_IP, HOST_IP, NODE_NAME, NAMESPACE, K8S_REGION (optional)
func NewK8sPlatform() *K8sPlatform {
	return &K8sPlatform{
		region:    Normalize(os.Getenv("K8S_REGION")),
		nodeName:  os.Getenv("NODE_NAME"),
		namespace: os.Getenv("NAMESPACE"),
		podIP:     os.Getenv("POD_IP"),
		hostIP:    os.Getenv("HOST_IP"),
		podName:   os.Getenv("POD_NAME"),
	}
}

func (k *K8sPlatform) Region() string {
	if r := k.knownRegion(); r != "" {
		return r
	}
	return "unknown"
}

// knownRegion is K8S_REGION, else an inherited AWS / GCP region env, else the
// fallback region DetectWith was given, else "": Region without its placeholder,
// which KnownRegion reads.
func (k *K8sPlatform) knownRegion() string {
	if k.region != "" {
		return k.region
	}
	if r := os.Getenv("AWS_REGION"); r != "" {
		return Normalize(r)
	}
	if r := os.Getenv("GOOGLE_CLOUD_REGION"); r != "" {
		return Normalize(r)
	}
	return k.fallbackRegion
}

func (k *K8sPlatform) MachineID() string { return k.podName }
func (k *K8sPlatform) AppName() string   { return k.namespace }
func (k *K8sPlatform) IsCloud() bool     { return true }
func (k *K8sPlatform) Provider() string  { return "k8s" }

// PublicIP returns the HOST_IP when the pod runs on hostNetwork, else "".
// Downstream discoverers will fall through to STUN/TURN for pods behind
// ClusterIP.
func (k *K8sPlatform) PublicIP() string {
	// Heuristic: if hostNetwork is on, HOST_IP == POD_IP and is routable.
	if k.hostIP != "" && k.hostIP == k.podIP {
		return k.hostIP
	}
	return ""
}

// PrivateIP returns POD_IP from the k8s downward-API. Stable for the pod
// lifetime. Returns "" if the downward API isn't wired into the pod
// spec — callers fall through to interface enumeration.
func (k *K8sPlatform) PrivateIP() string { return k.podIP }

func (k *K8sPlatform) ResolveUDPBind(port int) string {
	return fmt.Sprintf(":%d", port)
}

func (k *K8sPlatform) ResolveUDPBindDualStack(port int) (string, string) {
	return "", "" // single dual-stack socket suffices
}

func (k *K8sPlatform) OpenUDPListener(network, address string) (*net.UDPConn, error) {
	addr, err := net.ResolveUDPAddr(network, address)
	if err != nil {
		return nil, err
	}
	return net.ListenUDP(network, addr)
}

// EnvTags surfaces k8s identifiers for reach provenance.
func (k *K8sPlatform) EnvTags() map[string]string {
	tags := map[string]string{"provider": "k8s"}
	if k.podName != "" {
		tags["k8s:pod"] = k.podName
	}
	if k.namespace != "" {
		tags["k8s:namespace"] = k.namespace
	}
	if k.nodeName != "" {
		tags["k8s:node"] = k.nodeName
	}
	return tags
}

// PreferredSources: k8s downward API is authoritative for pod/host IPs;
// interface enumeration finds the pod CNI address; STUN/TURN cover
// ClusterIP-only pods.
func (k *K8sPlatform) PreferredSources() []AddressSource {
	return []AddressSource{
		SourceK8sDownward,
		SourceInterface,
		SourceSTUN,
		SourceTURN,
	}
}
