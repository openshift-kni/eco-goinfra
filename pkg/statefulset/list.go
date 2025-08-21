package statefulset

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// List returns statefulset inventory in the given namespace.
func List(apiClient *clients.Settings, nsname string, options ...metaV1.ListOptions) ([]*Builder, error) {
	if nsname == "" {
		glog.V(100).Infof("statefulset 'nsname' parameter can not be empty")

		return nil, fmt.Errorf("failed to list statefulsets, 'nsname' parameter is empty")
	}

	passedOptions := metaV1.ListOptions{}
	logMessage := fmt.Sprintf("Listing statefulsets in the namespace %s", nsname)

	if len(options) > 1 {
		glog.V(100).Infof("'options' parameter must be empty or single-valued")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	if len(options) == 1 {
		passedOptions = options[0]
		logMessage += fmt.Sprintf(" with the options %v", passedOptions)
	}

	glog.V(100).Infof(logMessage)

	statefulsetList, err := apiClient.StatefulSets(nsname).List(context.Background(), passedOptions)

	if err != nil {
		glog.V(100).Infof("Failed to list statefulsets in the namespace %s due to %s", nsname, err.Error())

		return nil, err
	}

	var statefulsetObjects []*Builder

	for _, runningStatefulSet := range statefulsetList.Items {
		copiedStatefulSet := runningStatefulSet
		statefulsetBuilder := &Builder{
			apiClient:  apiClient,
			Object:     &copiedStatefulSet,
			Definition: &copiedStatefulSet,
		}

		statefulsetObjects = append(statefulsetObjects, statefulsetBuilder)
	}

	return statefulsetObjects, nil
}
