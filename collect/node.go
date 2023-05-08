package collect

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	inventory "github.com/neticdk-k8s/k8s-inventory"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ck "k8s.io/client-go/kubernetes"
)

func CollectNodes(cs *ck.Clientset, i *inventory.Inventory) (errors []error) {
	nl := make([]*inventory.Node, 0)
	nodes, err := cs.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return []error{fmt.Errorf("getting nodes: %v", err)}
	}
	for _, o := range nodes.Items {
		node, err := CollectNode(o)
		errors = appendError(errors, err)
		nl = append(nl, node)
	}
	i.Nodes = nl
	return
}

func CollectNode(o v1.Node) (*inventory.Node, error) {
	n := inventory.NewNode()

	labels := o.GetLabels()

	criName, criVersion := parseContainerRuntimeVersion(o.Status.NodeInfo.ContainerRuntimeVersion)

	n.Name = o.GetName()
	n.Annotations = filterAnnotations(&o)
	if len(labels) > 0 {
		n.Labels = labels
	}
	n.Role = strings.Join(rolesFromNodeLabels(labels), ",")
	n.KubeProxyVersion = o.Status.NodeInfo.KubeProxyVersion
	n.KubeletVersion = o.Status.NodeInfo.KubeletVersion
	n.KernelVersion = o.Status.NodeInfo.KernelVersion
	n.CRIName = criName
	n.CRIVersion = criVersion
	n.ContainerRuntimeVersion = o.Status.NodeInfo.ContainerRuntimeVersion
	n.IsControlPlane = (n.Role != "worker")
	n.Provider = providerNameFromProviderID(o.Spec.ProviderID)
	n.TopologyRegion = regionFromNodeLabels(labels)
	n.TopologyZone = zoneFromNodeLabels(labels)
	n.CPUCapacityMillis = o.Status.Capacity.Cpu().MilliValue()
	n.CPUAllocatableMillis = o.Status.Allocatable.Cpu().MilliValue()
	n.MemoryCapacityBytes = o.Status.Capacity.Memory().Value()
	n.MemoryAllocatableBytes = o.Status.Allocatable.Memory().Value()

	return n, nil
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
