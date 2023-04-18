package kubernetes

import (
	"context"
	"fmt"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"

	ck "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func CreateK8SClient(impersonate string) (*ck.Clientset, error) {
	var err error
	var conf *restclient.Config
	log.Info("Configuring k8s client using in-cluster config")
	conf, err = restclient.InClusterConfig()
	if err != nil {
		log.Info("Could not load in-cluster config")
		log.Debug(err)
		log.Info("Trying out-of-cluster config")
		conf, err = clientcmd.BuildConfigFromFlags("", os.Getenv("HOME")+"/.kube/config")
		if err != nil {
			log.Error(err)
			return nil, fmt.Errorf("creating k8s client: %v", err)
		}
	}
	log.Info("K8s client configured")
	if impersonate != "" {
		log.Infof("Setting up impersonation as user: %v", impersonate)
		conf.Impersonate = restclient.ImpersonationConfig{UserName: impersonate}
	}

	clientset, err := ck.NewForConfig(conf)
	if err != nil {
		return nil, err
	}
	return clientset, err
}

func GetK8SRESTResource(cs *ck.Clientset, path string) (res restclient.Result, found bool, err error) {
	res = cs.Discovery().RESTClient().
		Get().
		AbsPath(path).
		Do(context.TODO())

	statusCode := 0
	res.StatusCode(&statusCode)
	if statusCode == http.StatusOK {
		found = true
	} else if statusCode == http.StatusNotFound {
		log.Infof("No %v resources found", path)
	} else {
		err = fmt.Errorf("expected %v, got %v", http.StatusOK, statusCode)
	}
	return res, found, err
}
