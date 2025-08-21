package ptp

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	ptpv1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/ptp/v1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ListPtpConfigs returns a list of PtpConfigs in all namespaces, using the provided options.
func ListPtpConfigs(
	apiClient *clients.Settings, options ...runtimeclient.ListOptions) ([]*PtpConfigBuilder, error) {
	if apiClient == nil {
		glog.V(100).Info("PtpConfigs 'apiClient' parameter cannot be nil")

		return nil, fmt.Errorf("failed to list PtpConfigs, 'apiClient' parameter is nil")
	}

	err := apiClient.AttachScheme(ptpv1.AddToScheme)
	if err != nil {
		glog.V(100).Info("Failed to add ptp v1 scheme to client schemes")

		return nil, err
	}

	logMessage := "Listing PtpConfigs in all namespaces"
	passedOptions := runtimeclient.ListOptions{}

	if len(options) > 1 {
		glog.V(100).Info("PtpConfigs 'options' parameter must be empty or single-valued")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	if len(options) == 1 {
		passedOptions = options[0]
		logMessage += fmt.Sprintf(" with the options %v", passedOptions)
	}

	glog.V(100).Info(logMessage)

	ptpConfigList := new(ptpv1.PtpConfigList)
	err = apiClient.List(context.TODO(), ptpConfigList, &passedOptions)

	if err != nil {
		glog.V(100).Infof("Failed to list PtpConfigs in all namespaces due to %v", err)

		return nil, err
	}

	var ptpConfigObjects []*PtpConfigBuilder

	for _, ptpConfig := range ptpConfigList.Items {
		copiedPtpConfig := ptpConfig
		ptpConfigBuilder := &PtpConfigBuilder{
			apiClient:  apiClient.Client,
			Object:     &copiedPtpConfig,
			Definition: &copiedPtpConfig,
		}

		ptpConfigObjects = append(ptpConfigObjects, ptpConfigBuilder)
	}

	return ptpConfigObjects, nil
}
