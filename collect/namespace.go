package collect

import (
	"context"
	"errors"
	"fmt"

	inventory "github.com/neticdk-k8s/k8s-inventory"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ck "k8s.io/client-go/kubernetes"
)

func CollectNamespaces(cs *ck.Clientset, i *inventory.Inventory) error {
	npl := make([]*inventory.NetworkPolicy, 0)
	networkPolicies, err := cs.NetworkingV1().NetworkPolicies("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("getting network policies: %v", err)
	}
	var errs []error
	for _, o := range networkPolicies.Items {
		np, err := CollectNetworkPolicy(o)
		errs = append(errs, err)
		npl = append(npl, np)
	}
	i.NetworkPolicies = npl
	return errors.Join(errs...)
}

func CollectNamespace(o v1.Namespace) (*inventory.Namespace, error) {
	r := inventory.NewNamespace()
	r.ObjectMeta = inventory.NewObjectMeta(o.ObjectMeta)
	return r, nil
}
