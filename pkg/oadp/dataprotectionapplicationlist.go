package oadp

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	oadpv1alpha1 "github.com/openshift-kni/eco-goinfra/pkg/schemes/oadp/api/v1alpha1"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ListDataProtectionApplication returns dataprotectionapplication inventory in the given namespace.
func ListDataProtectionApplication(
	apiClient *clients.Settings, nsname string, options ...runtimeClient.ListOptions) ([]*DPABuilder, error) {
	glog.V(100).Infof("Listing dataprotectionapplications in namespace %s", nsname)

	if apiClient == nil {
		glog.V(100).Infof("The apiClient cannot be nil")

		return nil, fmt.Errorf("the apiClient is nil")
	}

	if err := apiClient.AttachScheme(oadpv1alpha1.AddToScheme); err != nil {
		glog.V(100).Infof(
			"Failed to add oadpv1alpha1 scheme to client schemes")

		return nil, err
	}

	if nsname == "" {
		glog.V(100).Infof("dataprotectionapplication 'nsname' parameter can not be empty")

		return nil, fmt.Errorf("failed to list dataprotectionapplications, 'nsname' parameter is empty")
	}

	logMessage := fmt.Sprintf("Listing dataprotectionapplications in the nsname %s", nsname)
	passedOptions := runtimeClient.ListOptions{}

	if len(options) == 1 {
		passedOptions = options[0]
		logMessage += fmt.Sprintf(" with the options %v", passedOptions)
	} else if len(options) > 1 {
		glog.V(100).Infof("'options' parameter must be empty or single-valued")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	glog.V(100).Infof(logMessage)

	dataprotectionapplications := new(oadpv1alpha1.DataProtectionApplicationList)
	err := apiClient.List(context.TODO(), dataprotectionapplications, &passedOptions)

	if err != nil {
		glog.V(100).Infof("Failed to list all dataprotectionapplications due to %s", err.Error())

		return nil, err
	}

	var dpaObjects []*DPABuilder

	for _, dataprotectionapplication := range dataprotectionapplications.Items {
		copiedDPA := dataprotectionapplication
		builder := &DPABuilder{
			apiClient:  apiClient.Client,
			Object:     &copiedDPA,
			Definition: &copiedDPA,
		}

		dpaObjects = append(dpaObjects, builder)
	}

	return dpaObjects, nil
}
