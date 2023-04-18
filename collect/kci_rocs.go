package collect

import (
	dboperatorapi "github.com/kloeckner-i/db-operator/api/v1alpha1"
	inventory "github.com/neticdk-k8s/k8s-inventory"
	"github.com/neticdk-k8s/k8s-inventory-client/kubernetes"
	ck "k8s.io/client-go/kubernetes"
)

func CollectKCIRocksDBInstances(cs *ck.Clientset) ([]*inventory.KCIRocksDBInstance, error) {
	instances := make([]*inventory.KCIRocksDBInstance, 0)
	res, found, err := kubernetes.GetK8SRESTResource(cs, "/apis/kci.rocks/v1alpha1/dbinstances")
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	dbInstances := &dboperatorapi.DbInstanceList{}
	if err := res.Into(dbInstances); err != nil {
		return nil, err
	}
	for _, o := range dbInstances.Items {
		instances = append(instances, CollectKCIRocksDBInstance(o))
	}
	return instances, nil
}

func CollectKCIRocksDBInstance(o dboperatorapi.DbInstance) *inventory.KCIRocksDBInstance {
	dbi := inventory.NewKCIRocksDBInstance()
	dbi.Name = o.Name
	dbi.CreationTimestamp = o.CreationTimestamp
	dbi.Annotations = filterAnnotations(&o)
	labels := o.GetLabels()
	if len(labels) > 0 {
		dbi.Labels = labels
	}
	dbi.Engine = o.Spec.Engine
	dbi.Host = o.Spec.Generic.Host
	dbi.Port = o.Spec.Generic.Port
	dbi.Status = o.Status.Phase

	return dbi
}
