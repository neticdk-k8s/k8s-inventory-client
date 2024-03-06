package collect

import (
	"context"
	"errors"
	"fmt"

	inventory "github.com/neticdk-k8s/k8s-inventory"
	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ck "k8s.io/client-go/kubernetes"
)

func CollectNamespaces(cs *ck.Clientset, i *inventory.Inventory) error {
	npl := make([]*inventory.NetworkPolicy, 0)
	networkPolicies, err := cs.NetworkingV1().NetworkPolicies("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("getting network policies: %v", err)
	}
	var errs []error
	for _, o := range networkPolicies.Items {
		np, err := CollectNetworkPolicy(o)
		errs = append(errs, err)
		npl = append(npl, np)
	}
	i.NetworkPolicies = npl
	return errors.Join(errs...)
}

func CollectNetworkPolicy(o v1.NetworkPolicy) (*inventory.NetworkPolicy, error) {
	r := inventory.NewNetworkPolicy()
	r.ObjectMeta = inventory.NewObjectMeta(o.ObjectMeta)

	r.Spec.PodSelector.MatchLabels = o.Spec.PodSelector.MatchLabels
	if len(o.Spec.PodSelector.MatchExpressions) > 0 {
		r.Spec.PodSelector.MatchExpressions = make([]inventory.LabelSelectorRequirement, 0)
	}
	for _, me := range o.Spec.PodSelector.MatchExpressions {
		lse := inventory.LabelSelectorRequirement{
			Key:      me.Key,
			Operator: string(me.Operator),
			Values:   me.Values,
		}
		r.Spec.PodSelector.MatchExpressions = append(r.Spec.PodSelector.MatchExpressions, lse)
	}
	for _, ig := range o.Spec.Ingress {
		ingress := inventory.NetworkPolicyIngressRule{
			Ports: make([]inventory.NetworkPolicyPort, 0),
			From:  make([]inventory.NetworkPolicyPeer, 0),
		}
		for _, port := range ig.Ports {
			prot := string(*port.Protocol)
			ingress.Ports = append(ingress.Ports, inventory.NetworkPolicyPort{
				Protocol: &prot,
				Port: &inventory.IntOrString{
					Type:   int(port.Port.Type),
					IntVal: port.Port.IntVal,
					StrVal: port.Port.StrVal,
				},
				EndPort: port.EndPort,
			})
		}
		for _, from := range ig.From {
			ingressFrom := inventory.NetworkPolicyPeer{}
			if from.PodSelector != nil {
				ingressFrom.PodSelector = &inventory.LabelSelector{
					MatchLabels:      make(map[string]string),
					MatchExpressions: make([]inventory.LabelSelectorRequirement, 0),
				}
				ingressFrom.PodSelector.MatchLabels = from.PodSelector.MatchLabels
				for _, me := range from.PodSelector.MatchExpressions {
					ingressFrom.PodSelector.MatchExpressions = append(ingressFrom.PodSelector.MatchExpressions, inventory.LabelSelectorRequirement{
						Key:      me.Key,
						Operator: string(me.Operator),
						Values:   me.Values,
					})
				}
			}
			if from.NamespaceSelector != nil {
				ingressFrom.NamespaceSelector = &inventory.LabelSelector{
					MatchLabels:      make(map[string]string),
					MatchExpressions: make([]inventory.LabelSelectorRequirement, 0),
				}
				ingressFrom.NamespaceSelector.MatchLabels = from.NamespaceSelector.MatchLabels
				for _, me := range from.NamespaceSelector.MatchExpressions {
					ingressFrom.PodSelector.MatchExpressions = append(ingressFrom.NamespaceSelector.MatchExpressions, inventory.LabelSelectorRequirement{
						Key:      me.Key,
						Operator: string(me.Operator),
						Values:   me.Values,
					})
				}
			}
			if from.IPBlock != nil {
				ingressFrom.IPBlock = &inventory.IPBlock{
					CIDR:   from.IPBlock.CIDR,
					Except: from.IPBlock.Except,
				}
			}
			ingress.From = append(ingress.From, ingressFrom)
		}
		r.Spec.Ingress = append(r.Spec.Ingress, ingress)
	}
	for _, eg := range o.Spec.Egress {
		egress := inventory.NetworkPolicyEgressRule{
			Ports: make([]inventory.NetworkPolicyPort, 0),
			To:    make([]inventory.NetworkPolicyPeer, 0),
		}
		for _, port := range eg.Ports {
			prot := string(*port.Protocol)
			egress.Ports = append(egress.Ports, inventory.NetworkPolicyPort{
				Protocol: &prot,
				Port: &inventory.IntOrString{
					Type:   int(port.Port.Type),
					IntVal: port.Port.IntVal,
					StrVal: port.Port.StrVal,
				},
				EndPort: port.EndPort,
			})
		}
		for _, to := range eg.To {
			egressTo := inventory.NetworkPolicyPeer{}
			if to.PodSelector != nil {
				egressTo.PodSelector = &inventory.LabelSelector{
					MatchLabels:      make(map[string]string),
					MatchExpressions: make([]inventory.LabelSelectorRequirement, 0),
				}
				egressTo.PodSelector.MatchLabels = to.PodSelector.MatchLabels
				for _, me := range to.PodSelector.MatchExpressions {
					egressTo.PodSelector.MatchExpressions = append(egressTo.PodSelector.MatchExpressions, inventory.LabelSelectorRequirement{
						Key:      me.Key,
						Operator: string(me.Operator),
						Values:   me.Values,
					})
				}
			}
			if to.NamespaceSelector != nil {
				egressTo.NamespaceSelector = &inventory.LabelSelector{
					MatchLabels:      make(map[string]string),
					MatchExpressions: make([]inventory.LabelSelectorRequirement, 0),
				}
				egressTo.NamespaceSelector.MatchLabels = to.NamespaceSelector.MatchLabels
				for _, me := range to.NamespaceSelector.MatchExpressions {
					egressTo.PodSelector.MatchExpressions = append(egressTo.NamespaceSelector.MatchExpressions, inventory.LabelSelectorRequirement{
						Key:      me.Key,
						Operator: string(me.Operator),
						Values:   me.Values,
					})
				}
			}
			if to.IPBlock != nil {
				egressTo.IPBlock = &inventory.IPBlock{
					CIDR:   to.IPBlock.CIDR,
					Except: to.IPBlock.Except,
				}
			}
			egress.To = append(egress.To, egressTo)
		}
		r.Spec.Egress = append(r.Spec.Egress, egress)
	}
	r.Spec.PolicyTypes = make([]string, 0)
	for _, pt := range o.Spec.PolicyTypes {
		r.Spec.PolicyTypes = append(r.Spec.PolicyTypes, string(pt))
	}
	return r, nil
}
