package velero

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListBackupStorageLocationBuilder returns backupstoragelocation inventory in the given namespace.
func ListBackupStorageLocationBuilder(
	apiClient *clients.Settings, nsname string, options ...metav1.ListOptions) ([]*BackupStorageLocationBuilder, error) {
	if apiClient == nil {
		glog.V(100).Infof("The apiClient cannot be nil")

		return nil, fmt.Errorf("the apiClient is nil")
	}

	if nsname == "" {
		glog.V(100).Infof("backupstoragelocation 'nsname' parameter can not be empty")

		return nil, fmt.Errorf("failed to list backupstoragelocations, 'nsname' parameter is empty")
	}

	logMessage := fmt.Sprintf("Listing backupstoragelocations in the nsname %s", nsname)
	passedOptions := metav1.ListOptions{}

	if len(options) == 1 {
		passedOptions = options[0]
		logMessage += fmt.Sprintf(" with the options %v", passedOptions)
	} else if len(options) > 1 {
		glog.V(100).Infof("'options' parameter must be empty or single-valued")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	glog.V(100).Infof(logMessage)

	bslList, err := apiClient.BackupStorageLocations(nsname).List(context.TODO(), passedOptions)

	if err != nil {
		glog.V(100).Infof("Failed to list backupstoragelocations in the nsname %s due to %s", nsname, err.Error())

		return nil, err
	}

	var bslObjects []*BackupStorageLocationBuilder

	for _, runningBsl := range bslList.Items {
		copiedBsl := runningBsl
		bslBuilder := &BackupStorageLocationBuilder{
			apiClient:  apiClient.VeleroClient,
			Object:     &copiedBsl,
			Definition: &copiedBsl,
		}

		bslObjects = append(bslObjects, bslBuilder)
	}

	return bslObjects, nil
}
