package collect

import (
	"context"
	"fmt"

	inventory "github.com/neticdk-k8s/k8s-inventory"
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

func resolveRootOwner(kc client.Client, obj client.Object) (*inventory.RootOwner, error) {
	if len(obj.GetOwnerReferences()) == 0 {
		return nil, nil
	}
	rootObj, err := resolveOwnerChain(kc, obj)
	if err != nil {
		return nil, err
	}

	if rootObj != nil {
		rootOwner := &inventory.RootOwner{
			Kind:       rootObj.GetObjectKind().GroupVersionKind().Kind,
			APIGroup:   rootObj.GetObjectKind().GroupVersionKind().Group,
			APIVersion: rootObj.GetObjectKind().GroupVersionKind().Version,
			Name:       rootObj.GetName(),
			Namespace:  rootObj.GetNamespace(),
		}
		return rootOwner, nil
	}
	return nil, nil
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
