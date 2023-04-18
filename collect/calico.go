package collect

import (
	inventory "github.com/neticdk-k8s/k8s-inventory"
	"github.com/neticdk-k8s/k8s-inventory-client/kubernetes"
	calicoapi "github.com/projectcalico/api/pkg/apis/projectcalico/v3"
	ck "k8s.io/client-go/kubernetes"
)

func CollectCalico(cs *ck.Clientset) (*inventory.Calico, error) {
	calico := inventory.NewCalico()

	res, found, err := kubernetes.GetK8SRESTResource(cs, "/apis/crd.projectcalico.org/v1/clusterinformations/default")
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}

	o := &calicoapi.ClusterInformation{}
	if err := res.Into(o); err != nil {
		return nil, err
	}

	calico.CreationTimestamp = o.CreationTimestamp

	labels := o.GetLabels()
	if len(labels) > 0 {
		calico.Labels = labels
	}
	calico.Version = o.Spec.CalicoVersion

	return calico, nil
}
