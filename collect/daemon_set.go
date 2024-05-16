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

func collectDaemonSets(ctx context.Context, cs *ck.Clientset, client client.Client) ([]*inventory.Workload, error) {
	dsets := make([]*inventory.Workload, 0)

	daemonSetList, err := cs.AppsV1().
		DaemonSets("").
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("getting DaemonSets: %v", err)
	}
	var errs []error
	for _, o := range daemonSetList.Items {
		dset, err := collectDaemonSet(ctx, client, o)
		errs = append(errs, err)
		dsets = append(dsets, dset)
	}
	return dsets, errors.Join(errs...)
}

func collectDaemonSet(ctx context.Context, client client.Client, o v1.DaemonSet) (*inventory.Workload, error) {
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

	rootOwner, _, err := resolveRootOwner(ctx, client, &o)
	if err != nil {
		return nil, err
	}
	r.RootOwner = rootOwner

	return r, nil
}
