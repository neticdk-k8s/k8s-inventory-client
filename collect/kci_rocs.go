package collect

import (
	dboperatorapi "github.com/db-operator/db-operator/api/v1alpha1"
	inventory "github.com/neticdk-k8s/k8s-inventory"
	"github.com/neticdk-k8s/k8s-inventory-client/kubernetes"
	ck "k8s.io/client-go/kubernetes"
)

func collectKCIRocksDBInstances(cs *ck.Clientset) ([]*inventory.KCIRocksDBInstance, error) {
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
		instances = append(instances, collectKCIRocksDBInstance(o))
	}
	return instances, nil
}

func collectKCIRocksDBInstance(o dboperatorapi.DbInstance) *inventory.KCIRocksDBInstance {
	r := inventory.NewKCIRocksDBInstance()

	r.ObjectMeta = inventory.NewObjectMeta(o.ObjectMeta)

	r.Spec = inventory.KCIRocksDBInstanceSpec{
		Engine: o.Spec.Engine,
		Host:   o.Spec.Generic.Host,
		Port:   o.Spec.Generic.Port,
	}

	r.Status = inventory.KCIRocksDBInstanceStatus{
		Phase: o.Status.Phase,
	}

	return r
}
