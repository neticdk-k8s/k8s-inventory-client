package collect

import (
	"errors"

	inventory "github.com/neticdk-k8s/k8s-inventory"
	ck "k8s.io/client-go/kubernetes"
)

func CollectCustomResources(cs *ck.Clientset, i *inventory.Inventory) error {
	var errs []error

	resourceMap := make(map[string]bool)

	_, rl, err := cs.Discovery().ServerGroupsAndResources()
	for _, l := range rl {
		for _, r := range l.APIResources {
			resourceMap[l.GroupVersion+"/"+r.Name] = true
		}
	}

	if err != nil {
		errs = append(errs, err)
	} else {
		i.CustomResources.HasVelero = resourceMap["velero.io/v1/backups"]
		i.CustomResources.HasKCIRocks = resourceMap["kci.rocks/v1alpha1/dbinstances"]
		i.CustomResources.HasRabbitMQ = resourceMap["rabbitmq.com/v1beta1/rabbitmqclusters"]
		i.CustomResources.HasCalico = resourceMap["crd.projectcalico.org/v1/clusterinformations"]
		i.CustomResources.HasContour = resourceMap["projectcontour.io/v1/httpproxies"]
		i.CustomResources.HasExternalSecrets = resourceMap["external-secrets.io/v1alpha1/secretstores"]
		i.CustomResources.HasCertManager = resourceMap["cert-manager.io/v1/issuers"]
		i.CustomResources.HasGitOpsToolkit = resourceMap["source.toolkit.fluxcd.io/v1beta2/gitrepositories"]
		i.CustomResources.HasPrometheus = resourceMap["monitoring.coreos.com/v1/prometheuses"]
	}

	if i.CustomResources.HasVelero {
		velero_backups, err := CollectVeleroBackups(cs)
		errs = append(errs, err)
		i.CustomResources.Velero.Backups = velero_backups
		velero_schedules, err := CollectVeleroSchedules(cs)
		errs = append(errs, err)
		i.CustomResources.Velero.Schedules = velero_schedules
	}
	if i.CustomResources.HasKCIRocks {
		kcirocks_db_instances, err := CollectKCIRocksDBInstances(cs)
		errs = append(errs, err)
		i.CustomResources.KCIRocks.DBInstances = kcirocks_db_instances
	}
	if i.CustomResources.HasRabbitMQ {
		rabbitmq_clusters, err := CollectRabbitMQClusters(cs)
		errs = append(errs, err)
		i.CustomResources.RabbitMQ.Clusters = rabbitmq_clusters
	}
	if i.CustomResources.HasCalico {
		calico, err := CollectCalico(cs)
		errs = append(errs, err)
		i.CustomResources.CalicoCluster = calico
	}

	return errors.Join(errs...)
}
