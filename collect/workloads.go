package collect

import (
	"errors"

	inventory "github.com/neticdk-k8s/k8s-inventory"
	ck "k8s.io/client-go/kubernetes"
)

func CollectWorkloads(cs *ck.Clientset, i *inventory.Inventory) error {
	i.Workloads = make([]*inventory.Workload, 0)

	deployments, deploymentsErr := CollectDeployments(cs)
	i.Workloads = append(i.Workloads, deployments...)

	statefulSets, statefulSetsErr := CollectStatefulSets(cs)
	i.Workloads = append(i.Workloads, statefulSets...)

	replicaSets, replicaSetsErr := CollectReplicaSets(cs)
	i.Workloads = append(i.Workloads, replicaSets...)

	daemonSets, daemonSetsErr := CollectDaemonSets(cs)
	i.Workloads = append(i.Workloads, daemonSets...)

	cronJobs, cronJobErr := CollectCronJobs(cs)
	i.Workloads = append(i.Workloads, cronJobs...)

	jobs, jobsErr := CollectJobs(cs)
	i.Workloads = append(i.Workloads, jobs...)

	return errors.Join(
		deploymentsErr,
		statefulSetsErr,
		replicaSetsErr,
		daemonSetsErr,
		cronJobErr,
		jobsErr)
}
