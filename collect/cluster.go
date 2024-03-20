package collect

import (
	"fmt"

	"github.com/Masterminds/semver"
	inventory "github.com/neticdk-k8s/k8s-inventory"
	"github.com/neticdk-k8s/k8s-inventory-client/detect"
	ck "k8s.io/client-go/kubernetes"
)

func collectCluster(cs *ck.Clientset, i *inventory.Inventory) error {
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

func collectSCSMetadata(cs *ck.Clientset, i *inventory.Inventory) error {
	cm, err := readConfigMapByName(cs, "netic-metadata-system", "cluster-id")
	if err != nil {
		return err
	}

	i.Cluster.Name = cm.Data["cluster-name"]
	i.Cluster.FQDN = cm.Data["cluster-fqdn"]
	i.Cluster.ProviderName = cm.Data["provider-name"]

	return nil
}
