package collect

import (
	"context"
	"fmt"

	inventory "github.com/neticdk-k8s/k8s-inventory"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ck "k8s.io/client-go/kubernetes"
)

func CollectReplicaSets(cs *ck.Clientset) ([]*inventory.ReplicaSet, error) {
	var err error

	rsets := make([]*inventory.ReplicaSet, 0)

	replicaSetList, err := cs.AppsV1().
		ReplicaSets("").
		List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("getting ReplicaSets: %v", err)
	}
	for _, o := range replicaSetList.Items {
		rsets = append(rsets, CollectReplicaSet(o))
	}
	return rsets, nil
}

func CollectReplicaSet(o v1.ReplicaSet) *inventory.ReplicaSet {
	rs := inventory.NewReplicaSet()
	rs.Name = o.Name
	rs.Namespace = o.Namespace
	rs.CreationTimestamp = o.CreationTimestamp
	rs.Replicas = o.Spec.Replicas

	if len(o.OwnerReferences) > 0 {
		rs.OwnerKind = o.OwnerReferences[0].Kind
		rs.OwnerName = o.OwnerReferences[0].Name
	}

	rs.Annotations = filterAnnotations(&o)
	labels := o.GetLabels()
	if len(labels) > 0 {
		rs.Labels = labels
	}
	rs.Template.Containers = getContainerInfoFromContainers(o.Spec.Template.Spec.Containers)
	rs.Template.InitContainers = getContainerInfoFromContainers(o.Spec.Template.Spec.InitContainers)

	return rs
}
