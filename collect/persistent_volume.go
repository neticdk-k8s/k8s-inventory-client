package collect

import (
	"context"
	"fmt"

	inventory "github.com/neticdk-k8s/k8s-inventory"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ck "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/pkg/apis/core/v1/helper"
)

func CollectPVs(cs *ck.Clientset) ([]*inventory.PV, error) {
	var err error

	pvs := make([]*inventory.PV, 0)
	pvList, err := cs.CoreV1().
		PersistentVolumes().
		List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("getting PersistentVolumes: %v", err)
	}
	for _, o := range pvList.Items {
		pvs = append(pvs, CollectPV(o))
	}
	return pvs, nil
}

func CollectPV(o v1.PersistentVolume) *inventory.PV {
	pv := inventory.NewPV()
	pv.Name = o.Name
	pv.CreationTimestamp = o.CreationTimestamp
	pv.StorageClass = o.Spec.StorageClassName
	pv.Claim = fmt.Sprintf("%s/%s/%s", o.Spec.ClaimRef.Namespace, o.Spec.ClaimRef.Kind, o.Spec.ClaimRef.Name)
	pv.Status = string(o.Status.Phase)
	pv.AccessModes = helper.GetAccessModesAsString(o.Spec.AccessModes)
	pv.VolumeMode = string(*o.Spec.VolumeMode)
	pv.Capacity = o.Spec.Capacity.Storage().Value()
	pv.SetPersistentVolumeSource(o)

	pv.Annotations = filterAnnotations(&o)
	labels := o.GetLabels()
	if len(labels) > 0 {
		pv.Labels = labels
	}

	return pv
}
