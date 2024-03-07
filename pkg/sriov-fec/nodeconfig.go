package sriovfec

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	fecv2 "github.com/smart-edge-open/sriov-fec-operator/sriov-fec/api/v2"
)

// NodeConfigBuilder provides struct for SriovFecNodeConfig object which contains connection to cluster and
// SriovFecNodeConfig definitions.
type NodeConfigBuilder struct {
	// Dynamically discovered SriovFecNodeConfig object.
	Objects *fecv2.SriovFecNodeConfig
	// apiClient opens api connection to the cluster.
	apiClient *clients.Settings
	// nodeName defines on what node SriovFecNodeConfig resource should be queried.
	nodeName string
	// nsName defines SriovFec operator namespace.
	nsName string
	// errorMsg used in discovery function before sending api request to cluster.
	errorMsg string
}

// NewNodeConfigBuilder creates new instance of NodeConfigBuilder.
func NewNodeConfigBuilder(apiClient *clients.Settings, nodeName, nsname string) *NodeConfigBuilder {
	glog.V(100).Infof(
		"Initializing new NodeConfigBuilder structure with the following params: %s, %s",
		nodeName, nsname)

	builder := &NodeConfigBuilder{
		apiClient: apiClient,
		nodeName:  nodeName,
		nsName:    nsname,
	}

	if nodeName == "" {
		glog.V(100).Infof("The name of the nodeName is empty")

		builder.errorMsg = "SriovFecNodeConfig 'nodeName' is empty"
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the SriovFecNodeConfig is empty")

		builder.errorMsg = "SriovFecNodeConfig 'nsname' is empty"
	}

	return builder
}

// ListNodeConfig returns SriovFecNodeConfigs inventory in the given namespace.
func ListNodeConfig(
	apiClient *clients.Settings, nsname string, options ...metaV1.ListOptions) ([]*NodeConfigBuilder, error) {
	if nsname == "" {
		glog.V(100).Infof("SriovFecNodeConfigs 'nsname' parameter can not be empty")

		return nil, fmt.Errorf("failed to list SriovFecNodeConfigs, 'nsname' parameter is empty")
	}

	logMessage := fmt.Sprintf("Listing SriovFecNodeConfigs in the namespace %s", nsname)
	passedOptions := metaV1.ListOptions{}

	if len(options) > 1 {
		glog.V(100).Infof("'options' parameter must be empty or single-valued")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	if len(options) == 1 {
		passedOptions = options[0]
		logMessage += fmt.Sprintf(" with the options %v", passedOptions)
	}

	glog.V(100).Infof(logMessage)

	nodeConfigList, err := apiClient.SriovFecNodeConfigs(nsname).List(context.Background(), passedOptions)

	if err != nil {
		glog.V(100).Infof("Failed to list SriovFecNodeConfigs in the namespace %s due to %s", nsname, err.Error())

		return nil, err
	}

	var nodeConfigObjects []*NodeConfigBuilder

	for _, nodeConfig := range nodeConfigList.Items {
		copiedNodeConfig := nodeConfig
		stateBuilder := &NodeConfigBuilder{
			apiClient: apiClient,
			Objects:   &copiedNodeConfig,
			nsName:    nsname,
			nodeName:  copiedNodeConfig.Name}

		nodeConfigObjects = append(nodeConfigObjects, stateBuilder)
	}

	return nodeConfigObjects, nil
}
