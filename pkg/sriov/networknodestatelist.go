package sriov

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListNetworkNodeState returns SriovNetworkNodeStates inventory in the given namespace.
func ListNetworkNodeState(
	apiClient *clients.Settings, nsname string, options ...metaV1.ListOptions) ([]*NetworkNodeStateBuilder, error) {
	passedOptions := metaV1.ListOptions{}

	if len(options) == 1 {
		passedOptions = options[0]
	} else if len(options) > 1 {

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	glog.V(100).Infof("Listing SriovNetworkNodeStates in the namespace %s with the options %v", nsname, passedOptions)

	if nsname == "" {
		glog.V(100).Infof("SriovNetworkNodeStates 'nsname' parameter can not be empty")

		return nil, fmt.Errorf("failed to list SriovNetworkNodeStates, 'nsname' parameter is empty")
	}

	networkNodeStateList, err := apiClient.SriovNetworkNodeStates(nsname).List(context.Background(), passedOptions)

	if err != nil {
		glog.V(100).Infof("Failed to list SriovNetworkNodeStates in the namespace %s due to %s", nsname, err.Error())

		return nil, err
	}

	var networkNodeStateObjects []*NetworkNodeStateBuilder

	for _, networkNodeState := range networkNodeStateList.Items {
		copiedNetworkNodeState := networkNodeState
		stateBuilder := &NetworkNodeStateBuilder{
			apiClient: apiClient,
			Objects:   &copiedNetworkNodeState,
			nsName:    nsname,
			nodeName:  copiedNetworkNodeState.Name}

		networkNodeStateObjects = append(networkNodeStateObjects, stateBuilder)
	}

	return networkNodeStateObjects, nil
}
