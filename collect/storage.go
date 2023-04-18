package collect

import (
	inventory "github.com/neticdk-k8s/k8s-inventory"
	ck "k8s.io/client-go/kubernetes"
)

func CollectStorage(cs *ck.Clientset, i *inventory.Inventory) (errors []error) {
	pvs, err := CollectPVs(cs)
	errors = appendError(errors, err)
	i.Storage.PersistentVolumes = pvs
	sclss, err := CollectStorageClasses(cs)
	errors = appendError(errors, err)
	i.Storage.StorageClasses = sclss
	return
}
