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

func CollectDeployments(cs *ck.Clientset) ([]*inventory.Deployment, error) {
	var err error

	deployments := make([]*inventory.Deployment, 0)
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

func CollectDeployment(o v1.Deployment) *inventory.Deployment {
	d := inventory.NewDeployment()
	d.Name = o.Name
	d.Namespace = o.Namespace
	d.CreationTimestamp = o.CreationTimestamp
	d.Replicas = o.Spec.Replicas
	d.Strategy = string(o.Spec.Strategy.Type)

	d.Annotations = filterAnnotations(&o)
	labels := o.GetLabels()
	if len(labels) > 0 {
		d.Labels = labels
	}
	d.Template.Containers = getContainerInfoFromContainers(o.Spec.Template.Spec.Containers)
	d.Template.InitContainers = getContainerInfoFromContainers(o.Spec.Template.Spec.InitContainers)

	return d
}
