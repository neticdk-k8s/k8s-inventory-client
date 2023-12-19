package collect

import (
	"fmt"

	"github.com/Masterminds/semver"
	inventory "github.com/neticdk-k8s/k8s-inventory"
	"github.com/neticdk-k8s/k8s-inventory-client/detect"
	"github.com/rs/zerolog/log"
	ck "k8s.io/client-go/kubernetes"
)

func CollectCluster(cs *ck.Clientset, i *inventory.Inventory) error {
	v, err := cs.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("getting server version: %v", err)
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

	return nil
}

func CollectSCSMetadata(cs *ck.Clientset, i *inventory.Inventory) error {
	metaDataConfigMaps, err := readConfigMapsByLabel(cs, "netic-metadata-system", "netic.dk/owned-by=operator")
	if err != nil {
		return err
	}

	if len(metaDataConfigMaps) < 1 {
		log.Warn().Str("label", "netic.dk/owned-by=operator").Msg("ConfigMap with label matching not found")
		return nil
	}

	cm := metaDataConfigMaps[0]
	i.Cluster.Name = cm["cluster-name"]
	i.Cluster.ClusterSubscriptionID = parseHorrorID(cm["cluster-subscription-id"], 744)
	i.Cluster.FQDN = cm["cluster-fqdn"]
	i.Cluster.ClusterType = cm["cluster-type"]
	i.Cluster.Description = cm["cluster-description"]
	i.Cluster.ResilienceZone = cm["cluster-resilience-zone"]
	i.Cluster.EnvironmentName = cm["environment-name"]
	i.Cluster.InfrastructureEnvironmentType = cm["infrastructure-environment-type"]
	if i.Cluster.EnvironmentName == "" {
		i.Cluster.EnvironmentName = "NA"
	}
	i.Cluster.OperatorName = cm["operator-name"]
	i.Cluster.OperatorSubscriptionID = parseHorrorID(cm["operator-subscription-id"], 744)
	i.Cluster.ProviderName = cm["provider-name"]
	i.Cluster.ProviderSubscriptionID = parseHorrorID(cm["provider-subscription-id"], 744)
	i.Cluster.CustomerName = cm["customer-name"]
	i.Cluster.CustomerID = parseHorrorID(cm["customer-id"], 744)
	i.Cluster.BillingSubject = cm["billing-subject"]
	i.Cluster.BillingGranularity = cm["billing-granularity"]
	i.Cluster.HasTechnicalOperations = parseHorrorBool(cm["has-technical-operations"])
	i.Cluster.HasTechnicalManagement = parseHorrorBool(cm["has-technical-management"])
	i.Cluster.HasApplicationOperations = parseHorrorBool(cm["has-application-operations"])
	i.Cluster.HasApplicationManagement = parseHorrorBool(cm["has-application-management"])
	i.Cluster.HasCapacityManagement = parseHorrorBool(cm["has-capacity-management"])
	i.Cluster.HasCustomOperations = parseHorrorBool(cm["has-custom-operations"])
	i.Cluster.CustomOperationsURL = cm["custom-operations-url"]

	if i.Cluster.FQDN == "" && i.Cluster.Name != "" && i.Cluster.ProviderName != "" && i.Cluster.ClusterType != "" {
		i.Cluster.FQDN = fmt.Sprintf("%s.%s.%s.k8s.netic.dk", i.Cluster.Name, i.Cluster.ProviderName, i.Cluster.ClusterType)
	}

	return nil
}

func CollectSCSTenants(cs *ck.Clientset, i *inventory.Inventory) error {
	tenantDataConfigMaps, err := readConfigMapsByLabel(cs, "netic-metadata-system", "netic.dk/owned-by=tenant")
	if err != nil {
		return err
	}

	if len(tenantDataConfigMaps) < 1 {
		log.Warn().Str("label", "netic.dk/owned-by=tenant").Msg("ConfigMap with label matching not found")
		return nil
	}

	i.Tenants = nil
	for n := range tenantDataConfigMaps {
		tenant := inventory.NewTenant()
		tenant.Name = tenantDataConfigMaps[n]["tenant-name"]
		tenant.Namespace = tenantDataConfigMaps[n]["tenant-ns"]
		tenant.BusinessUnitID = tenantDataConfigMaps[n]["business-unit-id"]
		tenant.SubscriptionID = parseHorrorID(tenantDataConfigMaps[n]["tenant-subscription-id"], 744)
		tenant.HasApplicationOperations = parseHorrorBool(tenantDataConfigMaps[n]["has-application-operations"])
		tenant.HasApplicationManagement = parseHorrorBool(tenantDataConfigMaps[n]["has-application-management"])
		tenant.HasCapacityManagement = parseHorrorBool(tenantDataConfigMaps[n]["has-capacity-management"])
		i.Tenants = append(i.Tenants, tenant)
	}

	return nil
}
