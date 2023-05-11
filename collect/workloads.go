package collect

import (
	inventory "github.com/neticdk-k8s/k8s-inventory"
	ck "k8s.io/client-go/kubernetes"
)

func CollectWorkloads(cs *ck.Clientset, i *inventory.Inventory) (errors []error) {
	i.Workloads = make([]*inventory.Workload, 0)

	deployments, err := CollectDeployments(cs)
	errors = appendError(errors, err)
	i.Workloads = append(i.Workloads, deployments...)

	stateful_sets, err := CollectStatefulSets(cs)
	errors = appendError(errors, err)
	i.Workloads = append(i.Workloads, stateful_sets...)

	replica_sets, err := CollectReplicaSets(cs)
	errors = appendError(errors, err)
	i.Workloads = append(i.Workloads, replica_sets...)

	daemon_sets, err := CollectDaemonSets(cs)
	errors = appendError(errors, err)
	i.Workloads = append(i.Workloads, daemon_sets...)

	cron_jobs, errs := CollectCronJobs(cs)
	for _, e := range errs {
		errors = appendError(errors, e)
	}
	i.Workloads = append(i.Workloads, cron_jobs...)

	jobs, err := CollectJobs(cs)
	errors = appendError(errors, err)
	i.Workloads = append(i.Workloads, jobs...)

	return
}
