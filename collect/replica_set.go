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
	r := inventory.NewReplicaSet()

	r.ObjectMeta = inventory.NewObjectMeta(o.ObjectMeta)

	r.Spec = inventory.ReplicaSetSpec{
		Replicas: o.Spec.Replicas,
		Template: &inventory.PodTemplate{
			Containers:     getContainerInfoFromContainers(o.Spec.Template.Spec.Containers),
			InitContainers: getContainerInfoFromContainers(o.Spec.Template.Spec.InitContainers),
		},
	}

	r.Status = inventory.ReplicaSetStatus{
		Replicas:             o.Status.Replicas,
		FullyLabeledReplicas: o.Status.FullyLabeledReplicas,
		ReadyReplicas:        o.Status.ReadyReplicas,
		AvailableReplicas:    o.Status.AvailableReplicas,
	}

	return r
}
