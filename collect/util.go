package collect

import (
	"context"
	"fmt"
	"strings"

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

func resolveRootOwner(ctx context.Context, kc client.Client, obj client.Object) (*inventory.RootOwner, *inventory.Workload, error) {
	owner := metav1.GetControllerOf(obj)
	if owner == nil {
		return nil, nil, nil
	}
	rootObj, err := resolveOwnerChain(ctx, kc, obj.GetNamespace(), owner)
	if err != nil {
		return nil, nil, err
	}

	if rootObj != nil {
		rootOwner := &inventory.RootOwner{
			Kind:       rootObj.GetObjectKind().GroupVersionKind().Kind,
			APIGroup:   rootObj.GetObjectKind().GroupVersionKind().Group,
			APIVersion: rootObj.GetObjectKind().GroupVersionKind().Version,
			Name:       rootObj.GetName(),
			Namespace:  rootObj.GetNamespace(),
		}

		meta := metav1.ObjectMeta{
			Name:              rootObj.GetName(),
			Namespace:         rootObj.GetNamespace(),
			Labels:            rootObj.GetLabels(),
			Annotations:       rootObj.GetAnnotations(),
			CreationTimestamp: rootObj.GetCreationTimestamp(),
			OwnerReferences:   rootObj.GetOwnerReferences(),
		}

		apiGroup := rootOwner.APIGroup
		if apiGroup == "" {
			apiGroup = "core"
		}
		ownerWorkload := &inventory.Workload{
			TypeMeta: inventory.TypeMeta{
				Kind:         rootOwner.Kind,
				APIGroup:     apiGroup,
				APIVersion:   rootOwner.APIVersion,
				ResourceType: strings.ToLower(rootOwner.Kind),
			},
			ObjectMeta: inventory.NewObjectMeta(meta),
			Spec:       map[string]interface{}{},
			Status:     map[string]interface{}{},
		}

		return rootOwner, ownerWorkload, nil
	}

	return nil, nil, nil
}

func resolveOwnerChain(ctx context.Context, kc client.Client, namespace string, owner *metav1.OwnerReference) (*unstructured.Unstructured, error) {
	obj := &unstructured.Unstructured{}
	obj.SetAPIVersion(owner.APIVersion)
	obj.SetKind(owner.Kind)
	err := kc.Get(ctx, client.ObjectKey{Namespace: namespace, Name: owner.Name}, obj)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting object from object ref: %w", err)
	}
	owner = metav1.GetControllerOf(obj)
	if owner != nil {
		return resolveOwnerChain(ctx, kc, namespace, owner)
	} else {
		return obj, nil
	}
}
