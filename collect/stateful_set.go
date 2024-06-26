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

func collectStatefulSets(ctx context.Context, cs *ck.Clientset, client client.Client) ([]*inventory.Workload, error) {
	ssets := make([]*inventory.Workload, 0)

	statefulSetList, err := cs.AppsV1().
		StatefulSets("").
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("getting StatefulSets: %v", err)
	}
	var errs []error
	for _, o := range statefulSetList.Items {
		sset, err := collectStatefulSet(ctx, client, o)
		errs = append(errs, err)
		ssets = append(ssets, sset)
	}
	return ssets, errors.Join(errs...)
}

func collectStatefulSet(ctx context.Context, client client.Client, o v1.StatefulSet) (*inventory.Workload, error) {
	r := inventory.NewStatefulSet()

	r.ObjectMeta = inventory.NewObjectMeta(o.ObjectMeta)

	r.Spec = inventory.StatefulSetSpec{
		Replicas:    o.Spec.Replicas,
		ServiceName: o.Spec.ServiceName,
		Template: &inventory.PodTemplate{
			Containers:     getContainerInfoFromContainers(o.Spec.Template.Spec.Containers),
			InitContainers: getContainerInfoFromContainers(o.Spec.Template.Spec.InitContainers),
		},
		UpdateStrategy: string(o.Spec.UpdateStrategy.Type),
	}

	r.Status = inventory.StatefulSetStatus{
		Replicas:          o.Status.Replicas,
		ReadyReplicas:     o.Status.ReadyReplicas,
		CurrentReplicas:   o.Status.CurrentReplicas,
		UpdatedReplicas:   o.Status.UpdatedReplicas,
		AvailableReplicas: o.Status.AvailableReplicas,
	}

	rootOwner, _, err := resolveRootOwner(ctx, client, &o)
	if err != nil {
		return nil, err
	}
	r.RootOwner = rootOwner

	return r, nil
}
