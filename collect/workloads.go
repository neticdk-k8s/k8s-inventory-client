package collect

import (
	"errors"

	inventory "github.com/neticdk-k8s/k8s-inventory"
	ck "k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CollectWorkloads(cs *ck.Clientset, client client.Client, i *inventory.Inventory) error {
	i.Workloads = make([]*inventory.Workload, 0)

	deployments, deploymentsErr := CollectDeployments(cs, client)
	i.Workloads = append(i.Workloads, deployments...)

	statefulSets, statefulSetsErr := CollectStatefulSets(cs, client)
	i.Workloads = append(i.Workloads, statefulSets...)

	replicaSets, replicaSetsErr := CollectReplicaSets(cs, client)
	i.Workloads = append(i.Workloads, replicaSets...)

	daemonSets, daemonSetsErr := CollectDaemonSets(cs, client)
	i.Workloads = append(i.Workloads, daemonSets...)

	cronJobs, cronJobErr := CollectCronJobs(cs, client)
	i.Workloads = append(i.Workloads, cronJobs...)

	jobs, jobsErr := CollectJobs(cs, client)
	i.Workloads = append(i.Workloads, jobs...)

	pods, podsErr := CollectPods(cs, client)
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
