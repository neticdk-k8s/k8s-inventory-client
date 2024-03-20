package collect

import (
	"context"
	"fmt"

	inventory "github.com/neticdk-k8s/k8s-inventory"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ck "k8s.io/client-go/kubernetes"
)

func collectStorageClasses(cs *ck.Clientset) ([]*inventory.StorageClass, error) {
	sclss := make([]*inventory.StorageClass, 0)
	scList, err := cs.StorageV1().
		StorageClasses().
		List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("getting StorageClasses: %v", err)
	}
	for _, o := range scList.Items {
		sclss = append(sclss, collectStorageClass(o))
	}
	return sclss, nil
}

func collectStorageClass(o storagev1.StorageClass) *inventory.StorageClass {
	r := inventory.NewStorageClass()
	r.ObjectMeta = inventory.NewObjectMeta(o.ObjectMeta)
	r.Provisioner = o.Provisioner
	return r
}
