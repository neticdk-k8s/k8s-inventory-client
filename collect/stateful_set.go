package collect

import (
	"context"
	"fmt"

	inventory "github.com/neticdk-k8s/k8s-inventory"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ck "k8s.io/client-go/kubernetes"
)

func CollectStatefulSets(cs *ck.Clientset) ([]*inventory.StatefulSet, error) {
	var err error

	ssets := make([]*inventory.StatefulSet, 0)

	statefulSetList, err := cs.AppsV1().
		StatefulSets("").
		List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("getting StatefulSets: %v", err)
	}
	for _, o := range statefulSetList.Items {
		ssets = append(ssets, CollectStatefulSet(o))
	}
	return ssets, nil
}

func CollectStatefulSet(o v1.StatefulSet) *inventory.StatefulSet {
	ss := inventory.NewStatefulSet()
	ss.Name = o.Name
	ss.Namespace = o.Namespace
	ss.CreationTimestamp = o.CreationTimestamp
	ss.Replicas = o.Spec.Replicas
	ss.UpdateStrategy = string(o.Spec.UpdateStrategy.Type)

	ss.Annotations = filterAnnotations(&o)
	labels := o.GetLabels()
	if len(labels) > 0 {
		ss.Labels = labels
	}
	ss.Template.Containers = getContainerInfoFromContainers(o.Spec.Template.Spec.Containers)
	ss.Template.InitContainers = getContainerInfoFromContainers(o.Spec.Template.Spec.InitContainers)

	return ss
}
