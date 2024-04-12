package bmh

import (
	"context"
	"fmt"
	"time"

	goclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/wait"

	bmhv1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
)

const (
	fiveScds time.Duration = 5 * time.Second
)

// List returns bareMetalHosts inventory in the given namespace.
func List(apiClient *clients.Settings, nsname string, options ...goclient.ListOptions) ([]*BmhBuilder, error) {
	if nsname == "" {
		glog.V(100).Infof("bareMetalHost 'nsname' parameter can not be empty")

		return nil, fmt.Errorf("failed to list bareMetalHosts, 'nsname' parameter is empty")
	}

	logMessage := fmt.Sprintf("Listing bareMetalHosts in the namespace %s", nsname)
	passedOptions := goclient.ListOptions{}

	if len(options) > 1 {
		glog.V(100).Infof("'options' parameter must be empty or single-valued")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	if len(options) == 1 {
		passedOptions = options[0]
		logMessage += fmt.Sprintf(" with the options %v", passedOptions)
	}

	passedOptions.Namespace = nsname

	glog.V(100).Infof(logMessage)

	var bmhList bmhv1alpha1.BareMetalHostList
	err := apiClient.List(context.TODO(), &bmhList, &passedOptions)

	if err != nil {
		glog.V(100).Infof("Failed to list bareMetalHosts in the nsname %s due to %s", nsname, err.Error())

		return nil, err
	}

	var bmhObjects []*BmhBuilder

	for _, baremetalhost := range bmhList.Items {
		copiedBmh := baremetalhost
		bmhBuilder := &BmhBuilder{
			apiClient:  apiClient,
			Object:     &copiedBmh,
			Definition: &copiedBmh,
		}

		bmhObjects = append(bmhObjects, bmhBuilder)
	}

	return bmhObjects, nil
}

// WaitForAllBareMetalHostsInGoodOperationalState waits for all baremetalhosts to be in good Operational State
// for a time duration up to the timeout.
func WaitForAllBareMetalHostsInGoodOperationalState(apiClient *clients.Settings,
	nsname string,
	timeout time.Duration,
	options ...goclient.ListOptions) (bool, error) {
	glog.V(100).Infof("Waiting for all bareMetalHosts in %s namespace to have OK operationalStatus",
		nsname)

	bmhList, err := List(apiClient, nsname, options...)
	if err != nil {
		glog.V(100).Infof("Failed to list all bareMetalHosts in the %s namespace due to %s",
			nsname, err.Error())

		return false, err
	}

	// Wait 5 secs in each iteration before condition function () returns true or errors or times out
	// after availableDuration
	err = wait.PollUntilContextTimeout(
		context.TODO(), fiveScds, timeout, true, func(ctx context.Context) (bool, error) {
			for _, baremetalhost := range bmhList {
				status := baremetalhost.GetBmhOperationalState()

				if status != bmhv1alpha1.OperationalStatusOK {
					glog.V(100).Infof("The %s bareMetalHost in namespace %s has an unexpected operational status: %s",
						baremetalhost.Object.Name, baremetalhost.Object.Namespace, status)

					return false, nil
				}
			}

			return true, nil
		})

	if err == nil {
		glog.V(100).Infof("All baremetalhosts were found in the good Operational State "+
			"during defined timeout: %v", timeout)

		return true, nil
	}

	// Here err is "timed out waiting for the condition"
	glog.V(100).Infof("Not all baremetalhosts were found in the good Operational State "+
		"during defined timeout: %v", timeout)

	return false, err
}
