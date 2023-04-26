package collect

import (
	"fmt"
	"strconv"

	"github.com/Masterminds/semver"
	inventory "github.com/neticdk-k8s/k8s-inventory"
	"github.com/neticdk-k8s/k8s-inventory-client/detect"
	log "github.com/sirupsen/logrus"
	ck "k8s.io/client-go/kubernetes"
)

func CollectCluster(cs *ck.Clientset, i *inventory.Inventory) (errors []error) {
	v, err := cs.Discovery().ServerVersion()
	if err != nil {
		return []error{fmt.Errorf("getting server version: %v", err)}
	}

	semVer, _ := semver.NewVersion(v.String())
	i.Cluster.Version = fmt.Sprintf("%d.%d.%d", semVer.Major(), semVer.Minor(), semVer.Patch())
	i.Cluster.FullVersion = v.String()
	i.Cluster.GitCommit = v.GitCommit
	i.Cluster.BuildDate = v.BuildDate

	i.Cluster.KubernetesProvider = detect.DetectKubernetesProvider(cs)
	i.Cluster.InfrastructureProvider = detect.DetectInfrastructureProvider(cs, i.Cluster.KubernetesProvider)

	if i.Cluster.InfrastructureProvider == "docker" {
		i.Cluster.KubernetesProvider = "kind"
	}

	return
}

func CollectSCSMetadata(cs *ck.Clientset, i *inventory.Inventory) (errors []error) {
	metaDataConfigMaps, err := readConfigMapsByLabel(cs, "netic-metadata-system", "netic.dk/owned-by=operator")

	if err != nil {
		return []error{err}
	}

	if len(metaDataConfigMaps) < 1 {
		log.Warnf("no ConfigMap with label %s found", "netic.dk/owned-by=operator")
		return nil
	}

	cm := metaDataConfigMaps[0]
	i.Cluster.ClusterName = cm["cluster-name"]
	i.Cluster.ClusterFQDN = cm["cluster-fqdn"]
	i.Cluster.ClusterType = cm["cluster-type"]
	i.Cluster.ClusterDescription = cm["cluster-description"]
	i.Cluster.ClusterResilienceZone = cm["cluster-resilience-zone"]
	i.Cluster.EnvironmentName = cm["environment-name"]
	i.Cluster.InfrastructureEnvironmentType = cm["infrastructure-environment-type"]
	if i.Cluster.EnvironmentName == "" {
		i.Cluster.EnvironmentName = "NA"
	}
	i.Cluster.OperatorName = cm["operator-name"]
	i.Cluster.OperatorSubscriptionID, _ = strconv.Atoi(cm["operator-subscription-id"])
	i.Cluster.ProviderName = cm["provider-name"]
	i.Cluster.ProviderSubscriptionID, _ = strconv.Atoi(cm["provider-subscription-id"])
	i.Cluster.CustomerName = cm["customer-name"]
	i.Cluster.CustomerID, _ = strconv.Atoi(cm["customer-id"])
	i.Cluster.BillingSubject = cm["billing-subject"]
	i.Cluster.BillingGranularity = cm["billing-granularity"]
	i.Cluster.HasTechnicalOperations, _ = strconv.ParseBool(cm["has-technical-operations"])
	i.Cluster.HasTechnicalManagement, _ = strconv.ParseBool(cm["has-technical-management"])
	i.Cluster.HasApplicationOperations, _ = strconv.ParseBool(cm["has-application-operations"])
	i.Cluster.HasApplicationManagement, _ = strconv.ParseBool(cm["has-application-management"])
	i.Cluster.HasCapacityManagement, _ = strconv.ParseBool(cm["has-capacity-management"])
	i.Cluster.HasCustomOperations, _ = strconv.ParseBool(cm["has-custom-operations"])
	i.Cluster.CustomOperationsURL = cm["custom-operations-url"]

	if i.Cluster.ClusterFQDN == "" && i.Cluster.ClusterName != "" && i.Cluster.ProviderName != "" && i.Cluster.ClusterType != "" {
		i.Cluster.ClusterFQDN = fmt.Sprintf("%s.%s.%s.k8s.netic.dk", i.Cluster.ClusterName, i.Cluster.ProviderName, i.Cluster.ClusterType)
	}

	return
}

func CollectSCSTenants(cs *ck.Clientset, i *inventory.Inventory) (errors []error) {
	tenantDataConfigMaps, err := readConfigMapsByLabel(cs, "netic-metadata-system", "netic.dk/owned-by=tenant")

	if err != nil {
		return []error{err}
	}

	if len(tenantDataConfigMaps) < 1 {
		log.Warnf("no ConfigMap with label %s found", "netic.dk/owned-by=tenant")
		return nil
	}

	i.Tenants = nil
	for n := range tenantDataConfigMaps {
		tenant := inventory.NewTenant()
		tenant.Name = tenantDataConfigMaps[n]["tenant-name"]
		tenant.Namespace = tenantDataConfigMaps[n]["tenant-ns"]
		tenant.BusinessUnitID = tenantDataConfigMaps[n]["business-unit-id"]
		tenant.SubscriptionID, _ = strconv.Atoi(tenantDataConfigMaps[n]["tenant-subscription-id"])
		tenant.HasApplicationOperations, _ = strconv.ParseBool(tenantDataConfigMaps[n]["has-application-operations"])
		tenant.HasApplicationManagement, _ = strconv.ParseBool(tenantDataConfigMaps[n]["has-application-management"])
		tenant.HasCapacityManagement, _ = strconv.ParseBool(tenantDataConfigMaps[n]["has-capacity-management"])
		i.Tenants = append(i.Tenants, tenant)
	}

	return
}
