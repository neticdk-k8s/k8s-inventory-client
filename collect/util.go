package collect

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ck "k8s.io/client-go/kubernetes"
)

func readConfigMapByName(cs *ck.Clientset, ns string, name string) (*v1.ConfigMap, error) {
	res, err := cs.CoreV1().
		ConfigMaps(ns).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return res, nil
}
