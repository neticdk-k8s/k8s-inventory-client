package collect

import (
	inventory "github.com/neticdk-k8s/k8s-inventory"
	"github.com/neticdk-k8s/k8s-inventory-client/kubernetes"
	veleroapi "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	ck "k8s.io/client-go/kubernetes"
)

func CollectVeleroBackups(cs *ck.Clientset) ([]*inventory.VeleroBackup, error) {
	veleroBackups := make([]*inventory.VeleroBackup, 0)

	res, found, err := kubernetes.GetK8SRESTResource(cs, "/apis/velero.io/v1/backups")
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}

	backups := &veleroapi.BackupList{}
	if err := res.Into(backups); err != nil {
		return nil, err
	}

	for _, b := range backups.Items {
		var itemsBackedUp, totalItems int
		if b.Status.Progress != nil {
			itemsBackedUp = b.Status.Progress.ItemsBackedUp
			totalItems = b.Status.Progress.TotalItems
		}
		veleroBackup := &inventory.VeleroBackup{
			Name:                b.ObjectMeta.Name,
			Namespace:           b.ObjectMeta.Namespace,
			ScheduleName:        b.ObjectMeta.GetLabels()["velero.io/schedule-name"],
			ExcludedNamespaces:  b.Spec.ExcludedNamespaces,
			StorageLocation:     b.Spec.StorageLocation,
			SnapshotVolumes:     b.Spec.SnapshotVolumes,
			TTL:                 b.Spec.TTL,
			StartTimestamp:      b.Status.StartTimestamp,
			CompletionTimestamp: b.Status.CompletionTimestamp,
			Expiration:          b.Status.Expiration,
			Phase:               string(b.Status.Phase),
			ItemsBackedUp:       itemsBackedUp,
			TotalItems:          totalItems,
			Warnings:            b.Status.Warnings,
			Errors:              b.Status.Errors,
			Version:             b.Status.Version,
		}
		veleroBackups = append(veleroBackups, veleroBackup)
	}
	return veleroBackups, nil
}

func CollectVeleroSchedules(cs *ck.Clientset) ([]*inventory.VeleroSchedule, error) {
	veleroSchedules := make([]*inventory.VeleroSchedule, 0)
	res, found, err := kubernetes.GetK8SRESTResource(cs, "/apis/velero.io/v1/schedules")
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}

	schedules := &veleroapi.ScheduleList{}
	if err := res.Into(schedules); err != nil {
		return nil, err
	}

	for _, s := range schedules.Items {
		veleroSchedule := &inventory.VeleroSchedule{
			Name:               s.ObjectMeta.Name,
			Namespace:          s.ObjectMeta.Namespace,
			Schedule:           s.Spec.Schedule,
			ExcludedNamespaces: s.Spec.Template.ExcludedNamespaces,
			SnapshotVolumes:    s.Spec.Template.SnapshotVolumes,
			TTL:                s.Spec.Template.TTL,
			LastBackup:         s.Status.LastBackup,
			Phase:              string(s.Status.Phase),
		}
		veleroSchedules = append(veleroSchedules, veleroSchedule)
	}
	return veleroSchedules, nil
}
