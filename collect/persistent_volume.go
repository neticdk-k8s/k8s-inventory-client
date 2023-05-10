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

func CollectPVs(cs *ck.Clientset) ([]*inventory.PersistentVolume, error) {
	var err error

	pvs := make([]*inventory.PersistentVolume, 0)
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

func CollectPV(o v1.PersistentVolume) *inventory.PersistentVolume {
	r := inventory.NewPersistentVolume()

	r.ObjectMeta = inventory.NewObjectMeta(o.ObjectMeta)

	r.Spec = inventory.PersistentVolumeSpec{
		Capacity:         o.Spec.Capacity.Storage().Value(),
		AccessModes:      helper.GetAccessModesAsString(o.Spec.AccessModes),
		Claim:            fmt.Sprintf("%s/%s/%s", o.Spec.ClaimRef.Namespace, o.Spec.ClaimRef.Kind, o.Spec.ClaimRef.Name),
		StorageClassName: o.Spec.StorageClassName,
		VolumeMode:       string(*o.Spec.VolumeMode),
	}

	r.Status = inventory.PersistentVolumeStatus{
		Phase:   string(o.Status.Phase),
		Message: o.Status.Message,
	}

	r.SetPersistentVolumeSource(o)

	return r
}
