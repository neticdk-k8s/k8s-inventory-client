package detect

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	detector "github.com/rancher/kubernetes-provider-detector"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ck "k8s.io/client-go/kubernetes"
)

type infProviders struct {
	AWS   bool
	Azure bool
	GCP   bool
}

func DetectKubernetesProvider(cs *ck.Clientset) string {
	provider, err := detector.DetectProvider(context.TODO(), cs)
	if err != nil {
		log.Info("Could not detect cluster provider")
		provider = "undetected"
	}
	return provider
}

func DetectInfrastructureProvider(cs *ck.Clientset, kubernetesProvider string) string {
	switch kubernetesProvider {
	case "aks":
		return "azure"
	case "eks":
		return "aws"
	case "gcp":
		return "gcp"
	}
	var i infProviders
	var wg sync.WaitGroup
	var hc = &http.Client{Timeout: 300 * time.Millisecond}
	wg.Add(3)
	go func() {
		defer wg.Done()
		i.AWS = detectAWS(hc)
	}()
	go func() {
		defer wg.Done()
		i.Azure = detectAzure(hc)
	}()
	go func() {
		defer wg.Done()
		i.GCP = detectGCP(hc)
	}()
	wg.Wait()

	if i.AWS {
		return "aws"
	}
	if i.Azure {
		return "azure"
	}
	if i.GCP {
		return "gcp"
	}

	log.Infof("Collecting node information to detect additional cluster information")
	if nodes, err := cs.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{}); err == nil {
		for p := range nodes.Items {
			node := nodes.Items[p]
			labels := node.ObjectMeta.GetLabels()
			if strings.Contains(labels["topology.kubernetes.io/region"], "netic") || strings.Contains(labels["failure-domain.beta.kubernetes.io/region"], "aalborg") {
				return "netic"
			}
			if strings.HasPrefix(node.Spec.ProviderID, "kind") {
				return "docker"
			}
		}
	}

	return "undetected"
}

func detectAWS(client *http.Client) bool {
	resp, err := client.Get("http://169.254.169.254/latest/")
	if err != nil {
		return false
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	return resp.StatusCode == http.StatusOK
}

func detectAzure(client *http.Client) bool {
	resp, err := client.Get("http://169.254.169.254/metadata/v1/InstanceInfo")
	if err != nil {
		return false
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	return resp.StatusCode == http.StatusOK
}

func detectGCP(client *http.Client) bool {
	resp, err := http.NewRequest("GET", "http://metadata.google.internal/computeMetadata/v1/instance/tags", nil)
	if err != nil {
		return false
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	resp.Header.Add("Metadata-Flavor", "Google")
	resp2, err := client.Do(resp)
	if err != nil {
		return false
	}
	if resp2.Body != nil {
		defer resp2.Body.Close()
	}
	return resp2.StatusCode == http.StatusOK
}
