package oran

import (
	"fmt"
	"testing"

	pluginsv1alpha1 "github.com/openshift-kni/oran-o2ims/api/hardwaremanagement/plugins/v1alpha1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	defaultNARName      = "test-node-allocation-request"
	defaultNARNamespace = "test-namespace"
)

func TestPullNodeAllocationRequest(t *testing.T) {
	testCases := []struct {
		name                string
		nsname              string
		addToRuntimeObjects bool
		client              bool
		expectedError       error
	}{
		{
			name:                defaultNARName,
			nsname:              defaultNARNamespace,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       nil,
		},
		{
			name:                "",
			nsname:              defaultNARNamespace,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("nodeAllocationRequest 'name' cannot be empty"),
		},
		{
			name:                defaultNARName,
			nsname:              "",
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("nodeAllocationRequest 'nsname' cannot be empty"),
		},
		{
			name:                defaultNARName,
			nsname:              defaultNARNamespace,
			addToRuntimeObjects: false,
			client:              true,
			expectedError: fmt.Errorf("nodeAllocationRequest object %s does not exist in namespace %s",
				defaultNARName, defaultNARNamespace),
		},
		{
			name:                defaultNARName,
			nsname:              defaultNARNamespace,
			addToRuntimeObjects: true,
			client:              false,
			expectedError:       fmt.Errorf("nodeAllocationRequest 'apiClient' cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects,
				buildDummyNAR(defaultNARName, defaultNARNamespace))
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: pluginsTestSchemes,
			})
		}

		testBuilder, err := PullNodeAllocationRequest(testSettings, testCase.name, testCase.nsname)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, testBuilder.Definition.Name)
			assert.Equal(t, testCase.nsname, testBuilder.Definition.Namespace)
		}
	}
}

func TestNodeAllocationRequestGet(t *testing.T) {
	testCases := []struct {
		testBuilder   *NARBuilder
		expectedError string
	}{
		{
			testBuilder:   buildValidNARTestBuilder(buildTestClientWithDummyNAR()),
			expectedError: "",
		},
		{
			testBuilder: buildValidNARTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Sprintf(
				"nodeallocationrequests.plugins.clcm.openshift.io \"%s\" not found", defaultNARName),
		},
	}

	for _, testCase := range testCases {
		nodeAllocationRequest, err := testCase.testBuilder.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, testCase.testBuilder.Definition.Name, nodeAllocationRequest.Name)
			assert.Equal(t, testCase.testBuilder.Definition.Namespace, nodeAllocationRequest.Namespace)
		} else {
			assert.EqualError(t, err, testCase.expectedError)
		}
	}
}

func TestNodeAllocationRequestExists(t *testing.T) {
	testCases := []struct {
		testBuilder *NARBuilder
		exists      bool
	}{
		{
			testBuilder: buildValidNARTestBuilder(buildTestClientWithDummyNAR()),
			exists:      true,
		},
		{
			testBuilder: buildValidNARTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

// buildDummyNAR returns a NodeAllocationRequest with the provided name and nsname.
func buildDummyNAR(name, nsname string) *pluginsv1alpha1.NodeAllocationRequest {
	return &pluginsv1alpha1.NodeAllocationRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsname,
		},
	}
}

// buildTestClientWithDummyNAR returns an apiClient with the correct schemes and a NodeAllocationRequest with default
// name and namespace.
func buildTestClientWithDummyNAR() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyNAR(defaultNARName, defaultNARNamespace),
		},
		SchemeAttachers: pluginsTestSchemes,
	})
}

// buildValidNARTestBuilder returns a valid NARBuilder with all defaults and the provided apiClient.
func buildValidNARTestBuilder(apiClient *clients.Settings) *NARBuilder {
	_ = apiClient.AttachScheme(pluginsv1alpha1.AddToScheme)

	return &NARBuilder{
		Definition: buildDummyNAR(defaultNARName, defaultNARNamespace),
		apiClient:  apiClient,
	}
}
