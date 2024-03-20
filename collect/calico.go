package collect

import (
	inventory "github.com/neticdk-k8s/k8s-inventory"
	"github.com/neticdk-k8s/k8s-inventory-client/kubernetes"
	calicoapi "github.com/projectcalico/api/pkg/apis/projectcalico/v3"
	ck "k8s.io/client-go/kubernetes"
)

func collectCalico(cs *ck.Clientset) (*inventory.CalicoClusterInformation, error) {
	r := inventory.NewCalicoClusterInformation()

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

	r.ObjectMeta = inventory.NewObjectMeta(o.ObjectMeta)

	r.Spec = inventory.CalicoClusterInformationSpec{
		Version: o.Spec.CalicoVersion,
	}

	return r, nil
}
