package collect

import (
	inventory "github.com/neticdk-k8s/k8s-inventory"
	ck "k8s.io/client-go/kubernetes"
)

func CollectWorkloads(cs *ck.Clientset, i *inventory.Inventory) (errors []error) {
	deployments, err := CollectDeployments(cs)
	errors = appendError(errors, err)
	i.Workloads.Deployments = deployments
	stateful_sets, err := CollectStatefulSets(cs)
	errors = appendError(errors, err)
	i.Workloads.StatefulSets = stateful_sets
	replica_sets, err := CollectReplicaSets(cs)
	errors = appendError(errors, err)
	i.Workloads.ReplicaSets = replica_sets
	daemon_sets, err := CollectDaemonSets(cs)
	errors = appendError(errors, err)
	i.Workloads.DaemonSets = daemon_sets
	cron_jobs, errs := CollectCronJobs(cs)
	for _, e := range errs {
		errors = appendError(errors, e)
	}
	i.Workloads.CronJobs = cron_jobs
	jobs, err := CollectJobs(cs)
	errors = appendError(errors, err)
	i.Workloads.Jobs = jobs
	return
}
