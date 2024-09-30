package resourcequotas

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// List returns resource quota inventory in the given namespace.
func List(apiClient *clients.Settings, nsname string, options ...metav1.ListOptions) ([]*Builder, error) {
	glog.V(100).Infof("Listing resource quotas in the namespace %s", nsname)

	if nsname == "" {
		glog.V(100).Infof("resource quota 'nsname' parameter can not be empty")

		return nil, fmt.Errorf("failed to list resource quotas, 'nsname' parameter is empty")
	}

	passedOptions := metav1.ListOptions{}
	logMessage := fmt.Sprintf("Listing resource quotas in the namespace %s", nsname)

	if len(options) > 1 {
		glog.V(100).Infof("'options' parameter must be empty or single-valued")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	if len(options) == 1 {
		passedOptions = options[0]
		logMessage += fmt.Sprintf(" with the options %v", passedOptions)
	}

	glog.V(100).Infof(logMessage)

	resourceQuotaList, err := apiClient.ResourceQuotas(nsname).List(context.TODO(), passedOptions)

	if err != nil {
		glog.V(100).Infof("Failed to list resource quotas in the namespace %s due to %s", nsname, err.Error())

		return nil, err
	}

	var resourceQuotaObjects []*Builder

	for _, runningResourceQuota := range resourceQuotaList.Items {
		copiedResourceQuota := runningResourceQuota
		resourceQuotaBuilder := &Builder{
			apiClient:  apiClient.CoreV1Interface,
			Object:     &copiedResourceQuota,
			Definition: &copiedResourceQuota,
		}

		resourceQuotaObjects = append(resourceQuotaObjects, resourceQuotaBuilder)
	}

	return resourceQuotaObjects, nil
}
