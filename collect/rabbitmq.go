package collect

import (
	inventory "github.com/neticdk-k8s/k8s-inventory"
	"github.com/neticdk-k8s/k8s-inventory-client/kubernetes"
	rmqapi "github.com/rabbitmq/cluster-operator/api/v1beta1"
	ck "k8s.io/client-go/kubernetes"
)

func collectRabbitMQClusters(cs *ck.Clientset) ([]*inventory.RabbitMQCluster, error) {
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
		rmqClusters = append(rmqClusters, collectRabbitMQCluster(o))
	}
	return rmqClusters, nil
}

func collectRabbitMQCluster(o rmqapi.RabbitmqCluster) *inventory.RabbitMQCluster {
	r := inventory.NewRabbitMQCluster()

	r.ObjectMeta = inventory.NewObjectMeta(o.ObjectMeta)

	r.Spec = inventory.RabbitMQClusterSpec{
		Replicas: o.Spec.Replicas,
		Image:    o.Spec.Image,
	}
	return r
}
