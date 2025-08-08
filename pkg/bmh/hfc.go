package bmh

import (
	bmhv1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/internal/common"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// HFCBuilder provides a struct to interface with HostFirmwareComponents resources on a specific cluster.
type HFCBuilder struct {
	common.EmbeddableBuilder[bmhv1alpha1.HostFirmwareComponents, *bmhv1alpha1.HostFirmwareComponents]
}

// PullHFC pulls an existing HostFirmwareComponents from the cluster.
func PullHFC(apiClient *clients.Settings, name, nsname string) (*HFCBuilder, error) {
	return common.PullNamespacedBuilder[bmhv1alpha1.HostFirmwareComponents, HFCBuilder](
		apiClient, bmhv1alpha1.AddToScheme, name, nsname)
}

// GetGVK returns the Group, Version, and Kind for the HostFirmwareComponents resource.
func (builder *HFCBuilder) GetGVK() schema.GroupVersionKind {
	return bmhv1alpha1.GroupVersion.WithKind("HostFirmwareComponents")
}
