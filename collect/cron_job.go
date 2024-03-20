package collect

import (
	"context"
	"errors"
	"fmt"

	inventory "github.com/neticdk-k8s/k8s-inventory"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/batch/v1"
	v1beta1 "k8s.io/api/batch/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ck "k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func collectCronJobs(cs *ck.Clientset, client client.Client) ([]*inventory.Workload, error) {
	cjs := make([]*inventory.Workload, 0)
	v1Jobs, v1Err := collectCronJobsV1(cs, client)
	cjs = append(cjs, v1Jobs...)
	var (
		v1BetaErr  error
		v1BetaJobs []*inventory.Workload
	)
	if len(cjs) == 0 {
		v1BetaJobs, v1BetaErr = collectCronJobsV1beta1(cs, client)
		cjs = append(cjs, v1BetaJobs...)
	}
	return cjs, errors.Join(v1Err, v1BetaErr)
}

func collectCronJobsV1beta1(cs *ck.Clientset, client client.Client) ([]*inventory.Workload, error) {
	cjs := make([]*inventory.Workload, 0)
	cronJobList, err := cs.BatchV1beta1().
		CronJobs("").
		List(context.Background(), metav1.ListOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return nil, fmt.Errorf("getting CronJobs/v1beta1: %v", err)
	}
	var errs []error
	for _, o := range cronJobList.Items {
		cj, err := collectCronJob(inventory.NewCronJob(), client, o)
		errs = append(errs, err)
		cjs = append(cjs, cj)
	}
	return cjs, errors.Join(errs...)
}

func collectCronJobsV1(cs *ck.Clientset, client client.Client) ([]*inventory.Workload, error) {
	cjs := make([]*inventory.Workload, 0)
	cronJobList, err := cs.BatchV1().
		CronJobs("").
		List(context.Background(), metav1.ListOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return nil, fmt.Errorf("getting CronJobs/v1: %v", err)
	}
	var errs []error
	for _, o := range cronJobList.Items {
		cj, err := collectCronJob(inventory.NewCronJob(), client, o)
		errs = append(errs, err)
		cjs = append(cjs, cj)
	}
	return cjs, errors.Join(errs...)
}

func collectCronJob(cj *inventory.Workload, client client.Client, o interface{}) (*inventory.Workload, error) {
	switch obj := o.(type) {
	case v1beta1.CronJob:
		return collectCronJobV1Beta1(obj, client)
	case v1.CronJob:
		return collectCronJobV1(obj, client)
	default:
		log.Warn().Msgf("api/resource: %v not supported", obj)
	}
	return cj, nil
}

func collectCronJobV1(o v1.CronJob, client client.Client) (*inventory.Workload, error) {
	r := inventory.NewCronJob()

	r.ObjectMeta = inventory.NewObjectMeta(o.ObjectMeta)

	r.Spec = inventory.CronJobSpec{
		Schedule:          o.Spec.Schedule,
		ConcurrencyPolicy: string(o.Spec.ConcurrencyPolicy),
		JobTemplate: &inventory.PodTemplate{
			Containers:     getContainerInfoFromContainers(o.Spec.JobTemplate.Spec.Template.Spec.Containers),
			InitContainers: getContainerInfoFromContainers(o.Spec.JobTemplate.Spec.Template.Spec.InitContainers),
		},
	}

	r.Status = inventory.CronJobStatus{
		LastScheduleTime:   o.Status.LastScheduleTime,
		LastSuccessfulTime: o.Status.LastSuccessfulTime,
	}

	rootOwner, err := resolveRootOwner(client, &o)
	if err != nil {
		return nil, err
	}
	r.RootOwner = rootOwner

	return r, nil
}

func collectCronJobV1Beta1(o v1beta1.CronJob, client client.Client) (*inventory.Workload, error) {
	r := inventory.NewCronJob()
	r.APIVersion = "v1beta1"

	r.ObjectMeta = inventory.NewObjectMeta(o.ObjectMeta)

	r.Spec = inventory.CronJobSpec{
		Schedule:          o.Spec.Schedule,
		ConcurrencyPolicy: string(o.Spec.ConcurrencyPolicy),
		JobTemplate: &inventory.PodTemplate{
			Containers:     getContainerInfoFromContainers(o.Spec.JobTemplate.Spec.Template.Spec.Containers),
			InitContainers: getContainerInfoFromContainers(o.Spec.JobTemplate.Spec.Template.Spec.InitContainers),
		},
	}

	r.Status = inventory.CronJobStatus{
		LastScheduleTime:   o.Status.LastScheduleTime,
		LastSuccessfulTime: o.Status.LastSuccessfulTime,
	}

	rootOwner, err := resolveRootOwner(client, &o)
	if err != nil {
		return nil, err
	}
	r.RootOwner = rootOwner

	return r, nil
}
