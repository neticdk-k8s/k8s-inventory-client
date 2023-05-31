package collect

import (
	"errors"

	inventory "github.com/neticdk-k8s/k8s-inventory"
	ck "k8s.io/client-go/kubernetes"
)

func CollectStorage(cs *ck.Clientset, i *inventory.Inventory) error {
	pvs, pvsErr := CollectPVs(cs)
	i.Storage.PersistentVolumes = pvs
	sclss, sclssErr := CollectStorageClasses(cs)
	i.Storage.StorageClasses = sclss

	return errors.Join(pvsErr, sclssErr)
}
