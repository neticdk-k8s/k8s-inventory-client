package collect

import (
	"context"
	"fmt"

	inventory "github.com/neticdk-k8s/k8s-inventory"
	v1 "k8s.io/api/apps/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ck "k8s.io/client-go/kubernetes"
)

func CollectDeployments(cs *ck.Clientset) ([]*inventory.Workload, error) {
	var err error

	deployments := make([]*inventory.Workload, 0)
	deploymentList, err := cs.AppsV1().
		Deployments("").
		List(context.Background(), metav1.ListOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return nil, fmt.Errorf("getting Deployments: %v", err)
	}
	for _, o := range deploymentList.Items {
		deployments = append(deployments, CollectDeployment(o))
	}
	return deployments, nil
}

func CollectDeployment(o v1.Deployment) *inventory.Workload {
	r := inventory.NewDeployment()

	r.ObjectMeta = inventory.NewObjectMeta(o.ObjectMeta)

	r.Spec = inventory.DeploymentSpec{
		Strategy: string(o.Spec.Strategy.Type),
		Replicas: o.Spec.Replicas,
		Template: &inventory.PodTemplate{
			Containers:     getContainerInfoFromContainers(o.Spec.Template.Spec.Containers),
			InitContainers: getContainerInfoFromContainers(o.Spec.Template.Spec.InitContainers),
		},
	}

	r.Status = inventory.DeploymentStatus{
		Replicas:            o.Status.Replicas,
		ReadyReplicas:       o.Status.ReadyReplicas,
		UpdatedReplicas:     o.Status.UpdatedReplicas,
		AvailableReplicas:   o.Status.AvailableReplicas,
		UnavailableReplicas: o.Status.UnavailableReplicas,
	}

	return r
}
