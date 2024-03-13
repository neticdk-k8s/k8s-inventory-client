package collect

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	ck "k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func readConfigMapByName(cs *ck.Clientset, ns string, name string) (*v1.ConfigMap, error) {
	res, err := cs.CoreV1().
		ConfigMaps(ns).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func resolveOwnerChain(kc client.Client, obj client.Object) (client.Object, error) {
	owner := metav1.GetControllerOf(obj)
	if owner != nil {
		o := &unstructured.Unstructured{}
		o.SetAPIVersion(owner.APIVersion)
		o.SetKind(owner.Kind)
		err := kc.Get(context.TODO(), client.ObjectKey{Namespace: obj.GetNamespace(), Name: owner.Name}, o)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return nil, fmt.Errorf("owner has been deleted: %w", err)
			}
			return nil, fmt.Errorf("getting object from object ref: %w", err)
		}
		return resolveOwnerChain(kc, o)
	}
	return obj, nil
}
