package kubernetes

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/rs/zerolog/log"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ck "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func CreateK8SClient(impersonate string) (*ck.Clientset, client.Client, error) {
	var err error
	var conf *restclient.Config
	conf, err = restclient.InClusterConfig()
	if err != nil {
		log.Debug().Err(err).Msg("")
		log.Info().Msg("using out-of-cluster config")
		conf, err = clientcmd.BuildConfigFromFlags("", os.Getenv("HOME")+"/.kube/config")
		if err != nil {
			return nil, nil, err
		}
	}
	log.Info().Msg("using in-cluster config")
	if impersonate != "" {
		log.Info().Str("user", impersonate).Msg("impersonating as user")
		conf.Impersonate = restclient.ImpersonationConfig{UserName: impersonate}
	}

	clientset, err := ck.NewForConfig(conf)
	if err != nil {
		return nil, nil, err
	}
	cl, err := client.New(conf, client.Options{})
	if err != nil {
		return nil, nil, err
	}

	return clientset, cl, err
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
		log.Info().Msgf("No %v resources found", path)
	} else {
		err = fmt.Errorf("expected %v, got %v", http.StatusOK, statusCode)
	}
	return res, found, err
}
