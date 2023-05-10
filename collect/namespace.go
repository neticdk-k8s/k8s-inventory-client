package collect

import (
	"context"
	"fmt"

	inventory "github.com/neticdk-k8s/k8s-inventory"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ck "k8s.io/client-go/kubernetes"
)

func CollectNamespaces(cs *ck.Clientset, i *inventory.Inventory) (errors []error) {
	nl := make([]*inventory.Namespace, 0)
	namespaces, err := cs.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return []error{fmt.Errorf("getting namespaces: %v", err)}
	}
	for _, o := range namespaces.Items {
		ns, err := CollectNamespace(o)
		errors = appendError(errors, err)
		nl = append(nl, ns)
	}
	i.Namespaces = nl
	return
}

func CollectNamespace(o v1.Namespace) (*inventory.Namespace, error) {
	r := inventory.NewNamespace()

	r.ObjectMeta = inventory.NewObjectMeta(o.ObjectMeta)

	return r, nil
}
