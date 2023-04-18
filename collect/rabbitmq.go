package collect

import (
	inventory "github.com/neticdk-k8s/k8s-inventory"
	"github.com/neticdk-k8s/k8s-inventory-client/kubernetes"
	rmqapi "github.com/rabbitmq/cluster-operator/api/v1beta1"
	ck "k8s.io/client-go/kubernetes"
)

func CollectRabbitMQClusters(cs *ck.Clientset) ([]*inventory.RabbitMQCluster, error) {
	rmqClusters := make([]*inventory.RabbitMQCluster, 0)
	res, found, err := kubernetes.GetK8SRESTResource(cs, "/apis/rabbitmq.com/v1beta1/rabbitmqclusters")
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	clusters := &rmqapi.RabbitmqClusterList{}
	if err := res.Into(clusters); err != nil {
		return nil, err
	}
	for _, o := range clusters.Items {
		rmqClusters = append(rmqClusters, CollectRabbitMQCluster(o))
	}
	return rmqClusters, nil
}

func CollectRabbitMQCluster(o rmqapi.RabbitmqCluster) *inventory.RabbitMQCluster {
	c := inventory.NewRabbitMQCluster()
	c.Name = o.Name
	c.Namespace = o.Namespace
	c.CreationTimestamp = o.CreationTimestamp
	c.Annotations = filterAnnotations(&o)
	labels := o.GetLabels()
	if len(labels) > 0 {
		c.Labels = labels
	}
	c.Image = o.Spec.Image
	return c
}
