package assisted

import (
	"context"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	assistedv1beta1 "github.com/openshift/assisted-service/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ListNmStateConfigsInAllNamespaces returns a a cluster-wide NMStateConfig list.
func ListNmStateConfigsInAllNamespaces(apiClient *clients.Settings) ([]*NmStateConfigBuilder, error) {
	nmStateConfigList := &assistedv1beta1.NMStateConfigList{}

	err := apiClient.List(context.Background(), nmStateConfigList, &client.ListOptions{})

	if err != nil {
		glog.V(100).Infof("Failed to list nmStateConfigs due to %s", err.Error())

		return nil, err
	}

	var nmstateConfigObjects []*NmStateConfigBuilder

	for _, nmStateConfigObj := range nmStateConfigList.Items {
		nmStateConf := nmStateConfigObj
		nmStateConfBuilder := &NmStateConfigBuilder{
			apiClient:  apiClient,
			Definition: &nmStateConf,
			Object:     &nmStateConf,
		}

		nmstateConfigObjects = append(nmstateConfigObjects, nmStateConfBuilder)
	}

	return nmstateConfigObjects, err
}

// ListNmStateConfigs returns a NMStateConfig list in a given namespace.
func ListNmStateConfigs(apiClient *clients.Settings, namespace string) ([]*NmStateConfigBuilder, error) {
	nmStateConfigList := &assistedv1beta1.NMStateConfigList{}

	err := apiClient.List(context.Background(), nmStateConfigList, &client.ListOptions{Namespace: namespace})

	if err != nil {
		glog.V(100).Infof("Failed to list nmStateConfig due to %s in namespace: %s",
			err.Error(), namespace)

		return nil, err
	}

	var nmstateConfigObjects []*NmStateConfigBuilder

	for _, nmStateConfigObj := range nmStateConfigList.Items {
		nmStateConf := nmStateConfigObj
		nmStateConfBuilder := &NmStateConfigBuilder{
			apiClient:  apiClient,
			Definition: &nmStateConf,
			Object:     &nmStateConf,
		}

		nmstateConfigObjects = append(nmstateConfigObjects, nmStateConfBuilder)
	}

	return nmstateConfigObjects, err
}
