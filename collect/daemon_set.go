package collect

import (
	"context"
	"fmt"

	inventory "github.com/neticdk-k8s/k8s-inventory"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ck "k8s.io/client-go/kubernetes"
)

func CollectDaemonSets(cs *ck.Clientset) ([]*inventory.DaemonSet, error) {
	var err error

	dsets := make([]*inventory.DaemonSet, 0)

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

func CollectDaemonSet(o v1.DaemonSet) *inventory.DaemonSet {
	ds := inventory.NewDaemonSet()
	ds.Name = o.Name
	ds.Namespace = o.Namespace
	ds.CreationTimestamp = o.CreationTimestamp
	ds.UpdateStrategy = string(o.Spec.UpdateStrategy.Type)

	ds.Annotations = filterAnnotations(&o)
	labels := o.GetLabels()
	if len(labels) > 0 {
		ds.Labels = labels
	}
	ds.Template.Containers = getContainerInfoFromContainers(o.Spec.Template.Spec.Containers)
	ds.Template.InitContainers = getContainerInfoFromContainers(o.Spec.Template.Spec.InitContainers)

	return ds
}
