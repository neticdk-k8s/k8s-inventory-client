package collect

import (
	"context"
	"fmt"

	inventory "github.com/neticdk-k8s/k8s-inventory"
	v1 "k8s.io/api/batch/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ck "k8s.io/client-go/kubernetes"
)

func CollectJobs(cs *ck.Clientset) ([]*inventory.Workload, error) {
	jobs := make([]*inventory.Workload, 0)
	jobList, err := cs.BatchV1().
		Jobs("").
		List(context.Background(), metav1.ListOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return nil, fmt.Errorf("getting Jobs/v1: %v", err)
	}
	for _, o := range jobList.Items {
		jobs = append(jobs, CollectJob(o))
	}
	return jobs, nil
}

func CollectJob(o v1.Job) *inventory.Workload {
	r := inventory.NewJob()

	r.ObjectMeta = inventory.NewObjectMeta(o.ObjectMeta)

	r.Spec = inventory.JobSpec{
		Parallelism:  o.Spec.Parallelism,
		Completions:  o.Spec.Completions,
		BackoffLimit: o.Spec.BackoffLimit,
		Template: &inventory.PodTemplate{
			Containers:     getContainerInfoFromContainers(o.Spec.Template.Spec.Containers),
			InitContainers: getContainerInfoFromContainers(o.Spec.Template.Spec.InitContainers),
		},
	}

	r.Status = inventory.JobStatus{
		StartTime:      o.Status.StartTime,
		CompletionTime: o.Status.CompletionTime,
		Active:         o.Status.Active,
		Ready:          o.Status.Ready,
		Succeeded:      o.Status.Succeeded,
		Failed:         o.Status.Failed,
	}

	return r
}
