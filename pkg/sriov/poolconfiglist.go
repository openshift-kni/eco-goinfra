package sriov

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	srIovV1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ListPoolConfigs returns a sriovNetworkPoolConfig list in a given namespace.
func ListPoolConfigs(apiClient *clients.Settings, namespace string) ([]*PoolConfigBuilder, error) {
	sriovNetworkPoolConfigList := &srIovV1.SriovNetworkPoolConfigList{}

	if namespace == "" {
		glog.V(100).Infof("sriovNetworkPoolConfigs 'namespace' parameter can not be empty")

		return nil, fmt.Errorf("failed to list sriovNetworkPoolConfigs, 'namespace' parameter is empty")
	}

	err := apiClient.List(context.TODO(), sriovNetworkPoolConfigList, &client.ListOptions{Namespace: namespace})

	if err != nil {
		glog.V(100).Infof("Failed to list SriovNetworkPoolConfigs in namespace: %s due to %s",
			namespace, err.Error())

		return nil, err
	}

	var poolConfigBuilderObjects []*PoolConfigBuilder

	for _, sriovNetworkPoolConfigObj := range sriovNetworkPoolConfigList.Items {
		sriovNetworkPoolConfig := sriovNetworkPoolConfigObj
		sriovNetworkPoolConfBuilder := &PoolConfigBuilder{
			apiClient:  apiClient,
			Definition: &sriovNetworkPoolConfig,
			Object:     &sriovNetworkPoolConfig,
		}

		poolConfigBuilderObjects = append(poolConfigBuilderObjects, sriovNetworkPoolConfBuilder)
	}

	return poolConfigBuilderObjects, nil
}

// CleanAllNonDefaultPoolConfigs removes all sriovNetworkPoolConfigs that are not set as default.
func CleanAllNonDefaultPoolConfigs(
	apiClient *clients.Settings, operatornsname string) error {
	glog.V(100).Infof("Cleaning up SriovNetworkPoolConfigs in the %s namespace", operatornsname)

	if operatornsname == "" {
		glog.V(100).Infof("'operatornsname' parameter can not be empty")

		return fmt.Errorf("failed to clean up SriovNetworkPoolConfigs, 'operatornsname' parameter is empty")
	}

	poolConfigs, err := ListPoolConfigs(apiClient, operatornsname)

	if err != nil {
		glog.V(100).Infof("Failed to list SriovNetworkPoolConfigs in namespace: %s", operatornsname)

		return err
	}

	for _, poolConfig := range poolConfigs {
		// The "default" sriovNetworkPoolConfig is both mandatory and the default option.
		if poolConfig.Object.Name != "default" {
			err = poolConfig.Delete()

			if err != nil {
				glog.V(100).Infof("Failed to delete SriovNetworkPoolConfigs: %s", poolConfig.Object.Name)

				return err
			}
		}
	}

	return nil
}
