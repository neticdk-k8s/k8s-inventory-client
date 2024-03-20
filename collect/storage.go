package collect

import (
	"errors"

	inventory "github.com/neticdk-k8s/k8s-inventory"
	ck "k8s.io/client-go/kubernetes"
)

func collectStorage(cs *ck.Clientset, i *inventory.Inventory) error {
	pvs, pvsErr := collectPVs(cs)
	i.Storage.PersistentVolumes = pvs
	sclss, sclssErr := collectStorageClasses(cs)
	i.Storage.StorageClasses = sclss

	return errors.Join(pvsErr, sclssErr)
}
