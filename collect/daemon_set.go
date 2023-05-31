package collect

import (
	"context"
	"fmt"

	inventory "github.com/neticdk-k8s/k8s-inventory"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ck "k8s.io/client-go/kubernetes"
)

func CollectDaemonSets(cs *ck.Clientset) ([]*inventory.Workload, error) {
	dsets := make([]*inventory.Workload, 0)

	daemonSetList, err := cs.AppsV1().
		DaemonSets("").
		List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("getting DaemonSets: %v", err)
	}
	for _, o := range daemonSetList.Items {
		dsets = append(dsets, CollectDaemonSet(o))
	}
	return dsets, nil
}

func CollectDaemonSet(o v1.DaemonSet) *inventory.Workload {
	r := inventory.NewDaemonSet()

	r.ObjectMeta = inventory.NewObjectMeta(o.ObjectMeta)

	r.Spec = inventory.DaemonSetSpec{
		UpdateStrategy: string(o.Spec.UpdateStrategy.Type),
		Template: &inventory.PodTemplate{
			Containers:     getContainerInfoFromContainers(o.Spec.Template.Spec.Containers),
			InitContainers: getContainerInfoFromContainers(o.Spec.Template.Spec.InitContainers),
		},
	}

	r.Status = inventory.DaemonSetStatus{
		CurrentNumberScheduled: o.Status.CurrentNumberScheduled,
		NumberMisscheduled:     o.Status.NumberMisscheduled,
		DesiredNumberScheduled: o.Status.DesiredNumberScheduled,
	}

	return r
}
