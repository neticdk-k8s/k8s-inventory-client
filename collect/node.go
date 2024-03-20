package collect

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	inventory "github.com/neticdk-k8s/k8s-inventory"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ck "k8s.io/client-go/kubernetes"
)

func collectNodes(cs *ck.Clientset, i *inventory.Inventory) error {
	nl := make([]*inventory.Node, 0)
	nodes, err := cs.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("getting nodes: %v", err)
	}
	var errs []error
	for _, o := range nodes.Items {
		node, err := collectNode(o)
		errs = append(errs, err)
		nl = append(nl, node)
	}
	i.Nodes = nl
	return errors.Join(errs...)
}

func collectNode(o v1.Node) (*inventory.Node, error) {
	r := inventory.NewNode()

	labels := o.GetLabels()

	criName, criVersion := parseContainerRuntimeVersion(o.Status.NodeInfo.ContainerRuntimeVersion)

	r.ObjectMeta = inventory.NewObjectMeta(o.ObjectMeta)

	r.Spec = inventory.NodeSpec{
		PodCIDRs:      o.Spec.PodCIDRs,
		ProviderID:    o.Spec.ProviderID,
		Unschedulable: o.Spec.Unschedulable,
		Taints:        o.Spec.Taints,
	}

	r.Status = inventory.NodeStatus{
		NodeInfo: o.Status.NodeInfo,
	}

	r.Role = strings.Join(rolesFromNodeLabels(labels), ",")
	r.KubeProxyVersion = o.Status.NodeInfo.KubeProxyVersion
	r.KubeletVersion = o.Status.NodeInfo.KubeletVersion
	r.KernelVersion = o.Status.NodeInfo.KernelVersion
	r.CRIName = criName
	r.CRIVersion = criVersion
	r.ContainerRuntimeVersion = o.Status.NodeInfo.ContainerRuntimeVersion
	r.IsControlPlane = (r.Role != "worker")
	r.Provider = providerNameFromProviderID(o.Spec.ProviderID)
	r.TopologyRegion = regionFromNodeLabels(labels)
	r.TopologyZone = zoneFromNodeLabels(labels)
	r.CPUCapacityMillis = o.Status.Capacity.Cpu().MilliValue()
	r.CPUAllocatableMillis = o.Status.Allocatable.Cpu().MilliValue()
	r.MemoryCapacityBytes = o.Status.Capacity.Memory().Value()
	r.MemoryAllocatableBytes = o.Status.Allocatable.Memory().Value()

	return r, nil
}

func regionFromNodeLabels(nodeLabels map[string]string) string {
	testLabels := []string{"topology.kubernetes.io/region", "failure-domain.beta.kubernetes.io/region"}
	for _, l := range testLabels {
		if nodeLabels[l] != "" {
			return nodeLabels[l]
		}
	}
	return ""
}

func zoneFromNodeLabels(nodeLabels map[string]string) string {
	testLabels := []string{"topology.kubernetes.io/zone", "failure-domain.beta.kubernetes.io/zone"}
	for _, l := range testLabels {
		if nodeLabels[l] != "" {
			return nodeLabels[l]
		}
	}
	return ""
}

func rolesFromNodeLabels(nodeLabels map[string]string) (roles []string) {
	testRoles := []string{"controlplane", "control-plane", "master", "etcd"}
	for _, role := range testRoles {
		if _, ok := nodeLabels["node-role.kubernetes.io/"+role]; ok {
			roles = append(roles, role)
		}
	}

	if len(roles) == 0 {
		roles = append(roles, "worker")
	}
	return
}

func parseContainerRuntimeVersion(criVersion string) (name string, version string) {
	url, err := url.Parse(criVersion)
	if err != nil {
		return "", criVersion
	}
	return url.Scheme, url.Host
}

func providerNameFromProviderID(providerID string) string {
	url, err := url.Parse(providerID)
	if err != nil {
		return providerID
	}
	return url.Scheme
}
