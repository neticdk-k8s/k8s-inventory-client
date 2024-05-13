package collect

import (
	"context"
	"errors"

	inventory "github.com/neticdk-k8s/k8s-inventory"
	ck "k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func collectWorkloads(ctx context.Context, cs *ck.Clientset, client client.Client, i *inventory.Inventory) error {
	i.Workloads = make([]*inventory.Workload, 0)

	deployments, deploymentsErr := collectDeployments(ctx, cs, client)
	i.Workloads = append(i.Workloads, deployments...)

	statefulSets, statefulSetsErr := collectStatefulSets(ctx, cs, client)
	i.Workloads = append(i.Workloads, statefulSets...)

	replicaSets, replicaSetsErr := collectReplicaSets(ctx, cs, client)
	i.Workloads = append(i.Workloads, replicaSets...)

	daemonSets, daemonSetsErr := collectDaemonSets(ctx, cs, client)
	i.Workloads = append(i.Workloads, daemonSets...)

	cronJobs, cronJobErr := collectCronJobs(ctx, cs, client)
	i.Workloads = append(i.Workloads, cronJobs...)

	jobs, jobsErr := collectJobs(ctx, cs, client)
	i.Workloads = append(i.Workloads, jobs...)

	pods, owners, podsErr := collectPods(ctx, cs, client)
	i.Workloads = append(i.Workloads, pods...)

	// Append all pod owners that is _not_ already part of the collection
	for _, o := range owners {
		found := false
		for _, w := range i.Workloads {
			if o.APIGroup == w.APIGroup &&
				o.APIVersion == w.APIVersion &&
				o.Kind == w.Kind &&
				o.Name == w.Name &&
				o.Namespace == w.Namespace {
				found = true
			}
		}
		if !found {
			i.Workloads = append(i.Workloads, o)
		}
	}

	return errors.Join(
		deploymentsErr,
		statefulSetsErr,
		replicaSetsErr,
		daemonSetsErr,
		cronJobErr,
		jobsErr,
		podsErr,
	)
}
