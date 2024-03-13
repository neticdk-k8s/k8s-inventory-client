package collect

import (
	"context"
	"errors"
	"fmt"

	inventory "github.com/neticdk-k8s/k8s-inventory"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ck "k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CollectPods(cs *ck.Clientset, client client.Client, i *inventory.Inventory) error {
	options := metav1.ListOptions{Limit: 500}
	pods := make([]*inventory.Pod, 0)
	var errs []error
	for {
		podList, err := cs.CoreV1().
			Pods("").
			List(context.Background(), options)
		if err != nil && !k8serrors.IsNotFound(err) {
			errs = append(errs, fmt.Errorf("getting Pods: %v", err))
		}
		for _, o := range podList.Items {
			pod, err := CollectPod(client, o)
			errs = append(errs, err)
			pods = append(pods, pod)
		}
		if podList.Continue == "" {
			break
		}
		options.Continue = podList.Continue
	}
	i.Pods = pods
	return errors.Join(errs...)
}

func CollectPod(client client.Client, o v1.Pod) (*inventory.Pod, error) {
	r := inventory.NewPod()

	r.ObjectMeta = inventory.NewObjectMeta(o.ObjectMeta)

	r.Spec = inventory.PodSpec{
		InitContainers:     getContainers(o.Spec.InitContainers),
		Containers:         getContainers(o.Spec.Containers),
		RestartPolicy:      string(o.Spec.RestartPolicy),
		ServiceAccountName: o.Spec.ServiceAccountName,
		NodeName:           o.Spec.NodeName,
		HostNetwork:        o.Spec.HostNetwork,
		PriorityClassName:  o.Spec.PriorityClassName,
		Priority:           o.Spec.Priority,
	}
	if o.Spec.SecurityContext != nil {
		r.Spec.SecurityContext = &inventory.PodSecurityContext{
			RunAsUser:          o.Spec.SecurityContext.RunAsUser,
			RunAsGroup:         o.Spec.SecurityContext.RunAsUser,
			RunAsNonRoot:       o.Spec.SecurityContext.RunAsNonRoot,
			SupplementalGroups: make([]int64, 0),
		}
		r.Spec.SecurityContext.SupplementalGroups = append(r.Spec.SecurityContext.SupplementalGroups, o.Spec.SecurityContext.SupplementalGroups...)
	}
	if o.Spec.PreemptionPolicy != nil {
		policy := string(*o.Spec.PreemptionPolicy)
		r.Spec.PreemptionPolicy = &policy
	}
	for _, v := range o.Spec.Volumes {
		vol := inventory.Volume{
			Name:   v.Name,
			Source: volumeSource(v),
		}
		r.Spec.Volumes = append(r.Spec.Volumes, vol)
	}

	r.Status = inventory.PodStatus{
		Phase:                 string(o.Status.Phase),
		Conditions:            make([]inventory.PodCondition, 0),
		PodIP:                 o.Status.PodIP,
		StartTime:             o.Status.StartTime,
		InitContainerStatuses: getContainerStatuses(o.Status.InitContainerStatuses),
		ContainerStatuses:     getContainerStatuses(o.Status.ContainerStatuses),
		QOSClass:              string(o.Status.QOSClass),
	}

	for _, c := range o.Status.Conditions {
		r.Status.Conditions = append(r.Status.Conditions, inventory.PodCondition{
			Type:    string(c.Type),
			Status:  string(c.Status),
			Message: c.Message,
		})
	}

	rootObj, err := resolveOwnerChain(client, &o)
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
		r.RootOwner = rootOwner
	}

	return r, nil
}

func volumeSource(v v1.Volume) string {
	switch {
	case v.HostPath != nil:
		return "HostPath"
	case v.EmptyDir != nil:
		return "EmptyDir"
	case v.GCEPersistentDisk != nil:
		return "GCEPersistentDisk"
	case v.AWSElasticBlockStore != nil:
		return "AWSElasticBlockStore"
	case v.GitRepo != nil:
		return "GitRepo"
	case v.Secret != nil:
		return "Secret"
	case v.NFS != nil:
		return "NFS"
	case v.ISCSI != nil:
		return "ISCSI"
	case v.Glusterfs != nil:
		return "Glusterfs"
	case v.PersistentVolumeClaim != nil:
		return "PersistentVolumeClaim"
	case v.RBD != nil:
		return "RBD"
	case v.FlexVolume != nil:
		return "FlexVolume"
	case v.Cinder != nil:
		return "Cinder"
	case v.CephFS != nil:
		return "CephFS"
	case v.Flocker != nil:
		return "Flocker"
	case v.DownwardAPI != nil:
		return "DownwardAPI"
	case v.FC != nil:
		return "FC"
	case v.AzureFile != nil:
		return "AzureFile"
	case v.ConfigMap != nil:
		return "ConfigMap"
	case v.VsphereVolume != nil:
		return "VsphereVolume"
	case v.Quobyte != nil:
		return "Quobyte"
	case v.AzureDisk != nil:
		return "AzureDisk"
	case v.PhotonPersistentDisk != nil:
		return "PhotonPersistentDisk"
	case v.Projected != nil:
		return "Projected"
	case v.PortworxVolume != nil:
		return "PortworxVolume"
	case v.ScaleIO != nil:
		return "ScaleIO"
	case v.StorageOS != nil:
		return "StorageOS"
	case v.CSI != nil:
		return "CSI"
	case v.Ephemeral != nil:
		return "Ephemeral"
	default:
		return "Unknown"
	}
}
