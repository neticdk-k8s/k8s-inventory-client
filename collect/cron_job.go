package collect

import (
	"context"
	"fmt"

	inventory "github.com/neticdk-k8s/k8s-inventory"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/batch/v1"
	v1beta1 "k8s.io/api/batch/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ck "k8s.io/client-go/kubernetes"
)

func CollectCronJobs(cs *ck.Clientset) (cjs []*inventory.CronJob, errors []error) {
	cjs = make([]*inventory.CronJob, 0)
	v1Jobs, err := CollectCronJobsV1(cs)
	errors = appendError(errors, err)
	cjs = append(cjs, v1Jobs...)
	v1BetaJobs, err := CollectCronJobsV1beta1(cs)
	errors = appendError(errors, err)
	cjs = append(cjs, v1BetaJobs...)
	return
}

func CollectCronJobsV1beta1(cs *ck.Clientset) ([]*inventory.CronJob, error) {
	cjs := make([]*inventory.CronJob, 0)
	cronJobList, err := cs.BatchV1beta1().
		CronJobs("").
		List(context.Background(), metav1.ListOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return nil, fmt.Errorf("getting CronJobs/v1beta1: %v", err)
	}
	for _, o := range cronJobList.Items {
		cj := inventory.NewCronJob()
		CollectCronJob(cj, o)
		cjs = append(cjs, cj)
	}
	return cjs, nil
}

func CollectCronJobsV1(cs *ck.Clientset) ([]*inventory.CronJob, error) {
	cjs := make([]*inventory.CronJob, 0)
	cronJobList, err := cs.BatchV1().
		CronJobs("").
		List(context.Background(), metav1.ListOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return nil, fmt.Errorf("getting CronJobs/v1: %v", err)
	}
	for _, o := range cronJobList.Items {
		cj := CollectCronJob(inventory.NewCronJob(), o)
		cjs = append(cjs, cj)
	}
	return cjs, nil
}

func CollectCronJob(cj *inventory.CronJob, o interface{}) *inventory.CronJob {
	switch obj := o.(type) {
	case v1beta1.CronJob:
		return CollectCronJobV1Beta1(obj)
	case v1.CronJob:
		return CollectCronJobV1(obj)
	default:
		log.Warningf("api/resource: %v not supported", obj)
	}
	return cj
}

func CollectCronJobV1(o v1.CronJob) *inventory.CronJob {
	cj := inventory.NewCronJob()
	cj.Name = o.Name
	cj.Namespace = o.Namespace
	cj.CreationTimestamp = o.CreationTimestamp
	cj.Annotations = filterAnnotations(&o)
	labels := o.GetLabels()
	if len(labels) > 0 {
		cj.Labels = labels
	}
	cj.Schedule = o.Spec.Schedule
	cj.ConcurrencyPolicy = string(o.Spec.ConcurrencyPolicy)

	cj.Template.Containers = getContainerInfoFromContainers(o.Spec.JobTemplate.Spec.Template.Spec.Containers)
	cj.Template.InitContainers = getContainerInfoFromContainers(o.Spec.JobTemplate.Spec.Template.Spec.InitContainers)
	return cj
}

func CollectCronJobV1Beta1(o v1beta1.CronJob) *inventory.CronJob {
	cj := inventory.NewCronJob()
	cj.Name = o.Name
	cj.Namespace = o.Namespace
	cj.CreationTimestamp = o.CreationTimestamp
	cj.Annotations = filterAnnotations(&o)
	labels := o.GetLabels()
	if len(labels) > 0 {
		cj.Labels = labels
	}
	cj.Schedule = o.Spec.Schedule
	cj.ConcurrencyPolicy = string(o.Spec.ConcurrencyPolicy)

	cj.Template.Containers = getContainerInfoFromContainers(o.Spec.JobTemplate.Spec.Template.Spec.Containers)
	cj.Template.InitContainers = getContainerInfoFromContainers(o.Spec.JobTemplate.Spec.Template.Spec.InitContainers)
	return cj
}
