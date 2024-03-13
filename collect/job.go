package collect

import (
	"context"
	"errors"
	"fmt"

	inventory "github.com/neticdk-k8s/k8s-inventory"
	v1 "k8s.io/api/batch/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ck "k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CollectJobs(cs *ck.Clientset, client client.Client) ([]*inventory.Workload, error) {
	jobs := make([]*inventory.Workload, 0)
	options := metav1.ListOptions{Limit: 500}
	var errs []error
	for {
		jobList, err := cs.BatchV1().
			Jobs("").
			List(context.Background(), options)
		if err != nil && !k8serrors.IsNotFound(err) {
			return nil, fmt.Errorf("getting Jobs/v1: %v", err)
		}
	items:
		for _, o := range jobList.Items {
			if o.OwnerReferences != nil {
				for _, r := range o.OwnerReferences {
					if r.Kind == "CronJob" {
						continue items
					}
				}
			}
			job, err := CollectJob(client, o)
			errs = append(errs, err)
			jobs = append(jobs, job)
		}
		if jobList.Continue == "" {
			break
		}
		options.Continue = jobList.Continue
	}
	return jobs, errors.Join(errs...)
}

func CollectJob(client client.Client, o v1.Job) (*inventory.Workload, error) {
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

	rootOwner, err := resolveRootOwner(client, &o)
	if err != nil {
		return nil, err
	}
	r.RootOwner = rootOwner

	return r, nil
}
