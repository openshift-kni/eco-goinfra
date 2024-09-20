package metallb

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/schemes/metallb/frrtypes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ListFrrNodeState returns frr node state inventory in the given namespace.
func ListFrrNodeState(
	apiClient *clients.Settings, nsname string, options ...client.ListOptions) ([]*FrrNodeStateBuilder, error) {
	if apiClient == nil {
		glog.V(100).Infof("FrrNodeStates 'apiClient' parameter can not be empty")

		return nil, fmt.Errorf("failed to list FrrNodeStates, 'apiClient' parameter is empty")
	}

	err := apiClient.AttachScheme(frrtypes.AddToScheme)
	if err != nil {
		glog.V(100).Infof("Failed to add frrk8 scheme to client schemes")

		return nil, err
	}

	if nsname == "" {
		glog.V(100).Infof("FrrNodeStates 'nsname' parameter can not be empty")

		return nil, fmt.Errorf("failed to list FrrNodeStates, 'nsname' parameter is empty")
	}

	logMessage := fmt.Sprintf("Listing FrrNodeStates in the namespace %s", nsname)
	passedOptions := client.ListOptions{}

	if len(options) > 1 {
		glog.V(100).Infof("'options' parameter must be empty or single-valued")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	if len(options) == 1 {
		passedOptions = options[0]
		logMessage += fmt.Sprintf(" with the options %v", passedOptions)
	}

	glog.V(100).Infof(logMessage)

	frrNodeStateList := new(frrtypes.FRRNodeStateList)
	err = apiClient.List(context.TODO(), frrNodeStateList, &passedOptions)

	if err != nil {
		glog.V(100).Infof("Failed to list FrrNodeStates in the namespace %s due to %s", nsname, err.Error())

		return nil, err
	}

	var frrNodeStateObjects []*FrrNodeStateBuilder

	for _, networkNodeState := range frrNodeStateList.Items {
		copiedNetworkNodeState := networkNodeState
		stateBuilder := &FrrNodeStateBuilder{
			apiClient: apiClient.Client,
			Objects:   &copiedNetworkNodeState,
			nsName:    nsname,
			nodeName:  copiedNetworkNodeState.Name}

		frrNodeStateObjects = append(frrNodeStateObjects, stateBuilder)
	}

	return frrNodeStateObjects, nil
}
