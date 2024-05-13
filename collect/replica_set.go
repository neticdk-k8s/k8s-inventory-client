package collect

import (
	"context"
	"errors"
	"fmt"

	inventory "github.com/neticdk-k8s/k8s-inventory"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ck "k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func collectReplicaSets(ctx context.Context, cs *ck.Clientset, client client.Client) ([]*inventory.Workload, error) {
	rsets := make([]*inventory.Workload, 0)

	replicaSetList, err := cs.AppsV1().
		ReplicaSets("").
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("getting ReplicaSets: %v", err)
	}
	var errs []error
	for _, o := range replicaSetList.Items {
		rset, err := collectReplicaSet(ctx, client, o)
		errs = append(errs, err)
		rsets = append(rsets, rset)
	}
	return rsets, errors.Join(errs...)
}

func collectReplicaSet(ctx context.Context, client client.Client, o v1.ReplicaSet) (*inventory.Workload, error) {
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

	rootOwner, _, err := resolveRootOwner(ctx, client, &o)
	if err != nil {
		return nil, err
	}
	r.RootOwner = rootOwner

	return r, nil
}
