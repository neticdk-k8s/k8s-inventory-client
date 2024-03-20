package collect

import (
	"errors"

	inventory "github.com/neticdk-k8s/k8s-inventory"
	ck "k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func collectWorkloads(cs *ck.Clientset, client client.Client, i *inventory.Inventory) error {
	i.Workloads = make([]*inventory.Workload, 0)

	deployments, deploymentsErr := collectDeployments(cs, client)
	i.Workloads = append(i.Workloads, deployments...)

	statefulSets, statefulSetsErr := collectStatefulSets(cs, client)
	i.Workloads = append(i.Workloads, statefulSets...)

	replicaSets, replicaSetsErr := collectReplicaSets(cs, client)
	i.Workloads = append(i.Workloads, replicaSets...)

	daemonSets, daemonSetsErr := collectDaemonSets(cs, client)
	i.Workloads = append(i.Workloads, daemonSets...)

	cronJobs, cronJobErr := collectCronJobs(cs, client)
	i.Workloads = append(i.Workloads, cronJobs...)

	jobs, jobsErr := collectJobs(cs, client)
	i.Workloads = append(i.Workloads, jobs...)

	pods, podsErr := collectPods(cs, client)
	i.Workloads = append(i.Workloads, pods...)

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
