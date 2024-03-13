package collect

import (
	inventory "github.com/neticdk-k8s/k8s-inventory"
	v1 "k8s.io/api/core/v1"
)

func getContainerInfoFromContainers(containers []v1.Container) (ret []*inventory.PodTemplateContainer) {
	ret = []*inventory.PodTemplateContainer{}
	for _, c := range containers {
		i := &inventory.PodTemplateContainer{
			Image:          c.Image,
			LimitsCPU:      c.Resources.Limits.Cpu().MilliValue(),
			LimitsMemory:   c.Resources.Limits.Memory().Value(),
			RequestsCPU:    c.Resources.Requests.Cpu().MilliValue(),
			RequestsMemory: c.Resources.Requests.Memory().Value(),
		}
		ret = append(ret, i)
	}
	return
}

func getContainers(containers []v1.Container) (ret []inventory.Container) {
	ret = []inventory.Container{}
	for _, c := range containers {
		i := inventory.Container{
			Name:       c.Name,
			Image:      c.Image,
			WorkingDir: c.WorkingDir,
			Ports:      make([]inventory.ContainerPort, 0),
			Resources: inventory.ResourceRequirements{
				LimitsCPU:                c.Resources.Limits.Cpu().MilliValue(),
				RequestsCPU:              c.Resources.Requests.Cpu().MilliValue(),
				LimitsMemory:             c.Resources.Limits.Memory().Value(),
				RequestsMemory:           c.Resources.Requests.Memory().Value(),
				LimitsStorage:            c.Resources.Limits.Storage().Value(),
				RequestsStorage:          c.Resources.Requests.Storage().Value(),
				LimitsStorageEphemeral:   c.Resources.Limits.StorageEphemeral().Value(),
				RequestsStorageEphemeral: c.Resources.Requests.StorageEphemeral().Value(),
			},
			VolumeMounts:    make([]inventory.VolumeMount, 0),
			ImagePullPolicy: string(c.ImagePullPolicy),
		}
		i.Command = append(i.Command, c.Command...)
		i.Args = append(i.Args, c.Args...)
		for _, p := range c.Ports {
			i.Ports = append(i.Ports, inventory.ContainerPort{
				Name:          p.Name,
				HostPort:      p.HostPort,
				ContainerPort: p.ContainerPort,
				Protocol:      string(p.Protocol),
				HostIP:        p.HostIP,
			})
		}
		if c.RestartPolicy != nil {
			policy := string(*c.RestartPolicy)
			i.RestartPolicy = &policy
		}
		for _, m := range c.VolumeMounts {
			i.VolumeMounts = append(i.VolumeMounts, inventory.VolumeMount{
				Name:        m.Name,
				ReadOnly:    m.ReadOnly,
				MountPath:   m.MountPath,
				SubPath:     m.SubPath,
				SubPathExpr: m.SubPathExpr,
			})
		}
		if c.SecurityContext != nil {
			i.SecurityContext = &inventory.SecurityContext{
				Privileged:               c.SecurityContext.Privileged,
				RunAsUser:                c.SecurityContext.RunAsUser,
				RunAsGroup:               c.SecurityContext.RunAsGroup,
				RunAsNonRoot:             c.SecurityContext.RunAsNonRoot,
				ReadOnlyRootFilesystem:   c.SecurityContext.ReadOnlyRootFilesystem,
				AllowPrivilegeEscalation: c.SecurityContext.AllowPrivilegeEscalation,
			}
			if c.SecurityContext.Capabilities != nil {
				i.SecurityContext.Capabilities = &inventory.Capabilities{
					Add:  make([]string, 0),
					Drop: make([]string, 0),
				}
				for _, c := range c.SecurityContext.Capabilities.Add {
					i.SecurityContext.Capabilities.Add = append(i.SecurityContext.Capabilities.Add, string(c))
				}
				for _, c := range c.SecurityContext.Capabilities.Drop {
					i.SecurityContext.Capabilities.Drop = append(i.SecurityContext.Capabilities.Drop, string(c))
				}
			}
		}

		ret = append(ret, i)
	}
	return ret
}

func getContainerStatuses(containerstatuses []v1.ContainerStatus) (ret []inventory.ContainerStatus) {
	ret = []inventory.ContainerStatus{}
	for _, c := range containerstatuses {
		i := inventory.ContainerStatus{
			Name:    c.Name,
			Ready:   c.Ready,
			Image:   c.Image,
			ImageID: c.ImageID,
		}
		i.State = inventory.ContainerState{}
		if c.State.Waiting != nil {
			i.State.Waiting = &inventory.ContainerStateWaiting{
				Reason:  c.State.Waiting.Reason,
				Message: c.State.Waiting.Message,
			}
		}
		if c.State.Running != nil {
			i.State.Running = &inventory.ContainerStateRunning{
				StartedAt: c.State.Running.StartedAt,
			}
		}
		if c.State.Terminated != nil {
			i.State.Terminated = &inventory.ContainerStateTerminated{
				ExitCode:    c.State.Terminated.ExitCode,
				Signal:      c.State.Terminated.Signal,
				Reason:      c.State.Terminated.Reason,
				Message:     c.State.Terminated.Message,
				StartedAt:   c.State.Terminated.StartedAt,
				FinishedAt:  c.State.Terminated.FinishedAt,
				ContainerID: c.State.Terminated.ContainerID,
			}
		}

		ret = append(ret, i)
	}
	return ret
}
