package collect

import (
	inventory "github.com/neticdk-k8s/k8s-inventory"
	v1 "k8s.io/api/core/v1"
)

func getContainerInfoFromContainers(containers []v1.Container) (ret []*inventory.Container) {
	ret = []*inventory.Container{}
	for _, c := range containers {
		i := &inventory.Container{
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
