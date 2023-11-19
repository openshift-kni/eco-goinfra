package ocm

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	placementrulev1 "open-cluster-management.io/multicloud-operators-subscription/pkg/apis/apps/placementrule/v1"
)

// PlacementRuleBuilder provides struct for the PlacementRule object containing connection to
// the cluster and the PlacementRule definitions.
type PlacementRuleBuilder struct {
	// PlacementRule Definition, used to create the PlacementRule object.
	Definition *placementrulev1.PlacementRule
	// created PlacementRule object.
	Object *placementrulev1.PlacementRule
	// api client to interact with the cluster.
	apiClient *clients.Settings
	// used to store latest error message upon defining or mutating PlacementRule definition.
	errorMsg string
}

// Pull pulls existing placementrule into Builder struct.
func PullPlacementRule(apiClient *clients.Settings, name, nsname string) (*PlacementRuleBuilder, error) {
	glog.V(100).Infof("Pulling existing placementrule name %s under namespace %s from cluster", name, nsname)

	builder := PlacementRuleBuilder{
		apiClient: apiClient,
		Definition: &placementrulev1.PlacementRule{
			ObjectMeta: metaV1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the placementrule is empty")

		builder.errorMsg = "placementrule's 'name' cannot be empty"
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the placementrule is empty")

		builder.errorMsg = "placementrule's 'namespace' cannot be empty"
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("placementrule object %s doesn't exist in namespace %s", name, nsname)
	}

	builder.Definition = builder.Object

	return &builder, nil
}
