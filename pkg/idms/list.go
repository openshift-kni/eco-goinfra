package idms

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	configv1 "github.com/openshift/api/config/v1"

	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ListImageDigestMirrorSets returns a cluster-wide imagedigestmirrorset inventory.
func ListImageDigestMirrorSets(
	apiClient *clients.Settings,
	options ...runtimeClient.ListOptions) ([]*Builder, error) {
	passedOptions := runtimeClient.ListOptions{}
	logMessage := "Listing all imagedigestmirrorsets"

	if apiClient == nil {
		glog.V(100).Infof("The apiClient is nil")

		return nil, fmt.Errorf("apiClient cannot be nil")
	}

	if err := apiClient.AttachScheme(configv1.AddToScheme); err != nil {
		glog.V(100).Infof(
			"Failed to add configv1 scheme to client schemes")

		return nil, fmt.Errorf("failed to add configv1 to client schemes")
	}

	if len(options) > 1 {
		glog.V(100).Infof("'options' parameter must be empty or single-valued")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	if len(options) == 1 {
		passedOptions = options[0]
		logMessage += fmt.Sprintf(" with the options %v", passedOptions)
	}

	glog.V(100).Infof(logMessage)

	imageDigestMirrorSets := new(configv1.ImageDigestMirrorSetList)
	err := apiClient.List(context.TODO(), imageDigestMirrorSets, &passedOptions)

	if err != nil {
		glog.V(100).Infof("Failed to list all imageDigestMirrorSets due to %s", err.Error())

		return nil, err
	}

	var IDMSObjects []*Builder

	for _, idms := range imageDigestMirrorSets.Items {
		copiedIDMS := idms
		idmsBuilder := &Builder{
			apiClient:  apiClient.Client,
			Object:     &copiedIDMS,
			Definition: &copiedIDMS,
		}

		IDMSObjects = append(IDMSObjects, idmsBuilder)
	}

	return IDMSObjects, nil
}
