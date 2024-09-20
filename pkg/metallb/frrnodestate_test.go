package metallb

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/schemes/metallb/frrtypes"
	"github.com/stretchr/testify/assert"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultNodeName           = "worker-0"
	defaultFrrNodeStateNsName = "test-namespace"
)

func TestNewFrrNodeStateBuilder(t *testing.T) {
	generateNetworkBuilder := NewFrrNodeStateBuilder

	testCases := []struct {
		nodeName          string
		nsName            string
		expectedErrorText string
		client            bool
	}{
		{
			nodeName: defaultNodeName,
			nsName:   defaultFrrNodeStateNsName,
			client:   true,
		},
		{
			nodeName:          "",
			nsName:            defaultFrrNodeStateNsName,
			expectedErrorText: "FrrNodeState 'nodeName' is empty",
			client:            true,
		},
		{
			nodeName:          defaultNodeName,
			nsName:            "",
			expectedErrorText: "FrrNodeState 'nsname' is empty",
			client:            true,
		},
	}
	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})
		testFrrStructure := generateNetworkBuilder(
			testSettings, testCase.nodeName, testCase.nsName)
		assert.NotNil(t, testFrrStructure)

		if len(testCase.expectedErrorText) > 0 {
			assert.Equal(t, testFrrStructure.errorMsg, testCase.expectedErrorText)
		}
	}
}

func TestFrrNodeStateDiscovery(t *testing.T) {
	generateFrrNodeState := func(name, namespace string) *frrtypes.FRRNodeState {
		return &frrtypes.FRRNodeState{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec:   frrtypes.FRRNodeStateSpec{},
			Status: frrtypes.FRRNodeStateStatus{},
		}
	}

	testCases := []struct {
		nodeName            string
		nsName              string
		expectedError       error
		addToRuntimeObjects bool
	}{
		{
			nodeName:            defaultNodeName,
			nsName:              defaultFrrNodeStateNsName,
			addToRuntimeObjects: true,
		},
		{
			nodeName:            "",
			nsName:              defaultFrrNodeStateNsName,
			expectedError:       fmt.Errorf("FrrNodeState 'nodeName' is empty"),
			addToRuntimeObjects: true,
		},
		{
			nodeName:            defaultNodeName,
			nsName:              "",
			expectedError:       fmt.Errorf("FrrNodeState 'nsname' is empty"),
			addToRuntimeObjects: true,
		},
		{
			nodeName:            defaultNodeName,
			nsName:              defaultFrrNodeStateNsName,
			expectedError:       fmt.Errorf("frrnodestates.frrk8s.metallb.io \"\" not found"),
			addToRuntimeObjects: false,
		},
	}
	for _, testCase := range testCases {
		var (
			testSettings   *clients.Settings
			runtimeObjects []runtime.Object
		)

		if testCase.addToRuntimeObjects {
			FrrNodeState := generateFrrNodeState(testCase.nodeName, testCase.nsName)
			runtimeObjects = append(runtimeObjects, FrrNodeState)
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  runtimeObjects,
			SchemeAttachers: frrTestSchemes,
		})

		frrNodeStateBuilder := NewFrrNodeStateBuilder(testSettings, testCase.nodeName, testCase.nsName)
		err := frrNodeStateBuilder.Discover()

		if testCase.addToRuntimeObjects {
			assert.Equal(t, testCase.expectedError, err)
		} else {
			assert.True(t, k8serrors.IsNotFound(err))
		}

		if testCase.expectedError == nil {
			assert.NotNil(t, frrNodeStateBuilder.Objects)
		} else {
			assert.Nil(t, frrNodeStateBuilder.Objects)
		}
	}
}
