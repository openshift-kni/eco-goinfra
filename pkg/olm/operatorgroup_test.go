package olm

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	v1 "github.com/openshift-kni/eco-goinfra/pkg/schemes/olm/operators/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	operatorTestSchemes = []clients.SchemeAttacher{
		v1.AddToScheme,
	}
)

func TestNewOperatorGroupBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		expectedError string
		client        bool
	}{
		{
			name:          "operatorgroup",
			namespace:     "test-namespace",
			client:        true,
			expectedError: "",
		},
		{
			name:          "",
			namespace:     "test-namespace",
			client:        true,
			expectedError: "operatorGroup 'groupName' cannot be empty",
		},
		{
			name:          "operatorgroup",
			namespace:     "",
			client:        true,
			expectedError: "operatorGroup 'Namespace' cannot be empty",
		},
		{
			name:          "operatorgroup",
			namespace:     "test-namespace",
			client:        false,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})
		}

		operatorGroup := NewOperatorGroupBuilder(testSettings, testCase.name, testCase.namespace)

		if testCase.client {
			assert.Equal(t, testCase.expectedError, operatorGroup.errorMsg)

			if testCase.expectedError == "" {
				assert.NotNil(t, operatorGroup.Definition)
				assert.Equal(t, testCase.name, operatorGroup.Definition.Name)
				assert.Equal(t, testCase.namespace, operatorGroup.Definition.Namespace)
			}
		} else {
			assert.Nil(t, operatorGroup)
		}
	}
}

func TestPullIOperatorGroup(t *testing.T) {
	operatorGroup := func(name, namespace string) *v1.OperatorGroup {
		return &v1.OperatorGroup{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: v1.OperatorGroupSpec{
				ServiceAccountName: "test",
			},
		}
	}

	testCases := []struct {
		name                string
		namespace           string
		addToRuntimeObjects bool
		expectedError       error
		client              bool
	}{
		{
			name:                "operatorgroup",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("operatorGroup 'Name' cannot be empty"),
			client:              true,
		},
		{
			name:                "operatorgroup",
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("operatorGroup 'Namespace' cannot be empty"),
			client:              true,
		},
		{
			name:                "operatorgroup",
			namespace:           "test-namespace",
			addToRuntimeObjects: false,
			expectedError:       fmt.Errorf("operatorGroup object named test-namespace does not exist"),
			client:              true,
		},
		{
			name:                "operatorgroup",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("operatorGroup 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		operatorGroupObj := operatorGroup(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, operatorGroupObj)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: operatorTestSchemes})
		}

		builderResult, err := PullOperatorGroup(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, builderResult.Object.Name)
			assert.Equal(t, testCase.namespace, builderResult.Object.Namespace)
		}
	}
}

func TestOperatorGroupGet(t *testing.T) {
	testCases := []struct {
		operatorGroup *OperatorGroupBuilder
		expectedError string
	}{
		{
			operatorGroup: buildValidOperatorGroupBuilder(buildOperatorGroupTestClientWithDummyObject()),
			expectedError: "",
		},
		{
			operatorGroup: buildInValidOperatorGroupBuilder(buildOperatorGroupTestClientWithDummyObject()),
			expectedError: "operatorGroup 'Namespace' cannot be empty",
		},
		{
			operatorGroup: buildValidOperatorGroupBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})),
			expectedError: "operatorgroups.operators.coreos.com \"operatorgroup\" not found",
		},
	}

	for _, testCase := range testCases {
		operatorGroup, err := testCase.operatorGroup.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, operatorGroup.Name, testCase.operatorGroup.Definition.Name)
		} else {
			assert.Equal(t, testCase.expectedError, err.Error())
		}
	}
}

func TestOperatorGroupExist(t *testing.T) {
	testCases := []struct {
		operatorGroup  *OperatorGroupBuilder
		expectedStatus bool
	}{
		{
			operatorGroup:  buildValidOperatorGroupBuilder(buildOperatorGroupTestClientWithDummyObject()),
			expectedStatus: true,
		},
		{
			operatorGroup:  buildInValidOperatorGroupBuilder(buildOperatorGroupTestClientWithDummyObject()),
			expectedStatus: false,
		},
		{
			operatorGroup: buildValidOperatorGroupBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: operatorTestSchemes})),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.operatorGroup.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestOperatorGroupCreate(t *testing.T) {
	testCases := []struct {
		operatorGroup *OperatorGroupBuilder
		expectedError error
	}{
		{
			operatorGroup: buildValidOperatorGroupBuilder(buildOperatorGroupTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			operatorGroup: buildInValidOperatorGroupBuilder(buildOperatorGroupTestClientWithDummyObject()),
			expectedError: fmt.Errorf("operatorGroup 'Namespace' cannot be empty"),
		},
		{
			operatorGroup: buildValidOperatorGroupBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: operatorTestSchemes})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		operatorGroupBuilder, err := testCase.operatorGroup.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, operatorGroupBuilder.Definition.Name, operatorGroupBuilder.Object.Name)
		}
	}
}

func TestOperatorGroupDelete(t *testing.T) {
	testCases := []struct {
		operatorGroup *OperatorGroupBuilder
		expectedError error
	}{
		{
			operatorGroup: buildValidOperatorGroupBuilder(buildOperatorGroupTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			operatorGroup: buildInValidOperatorGroupBuilder(buildOperatorGroupTestClientWithDummyObject()),
			expectedError: fmt.Errorf("operatorGroup 'Namespace' cannot be empty"),
		},
		{
			operatorGroup: buildValidOperatorGroupBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: operatorTestSchemes})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		err := testCase.operatorGroup.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.operatorGroup.Object)
		}
	}
}

func TestOperatorGroupUpdate(t *testing.T) {
	testCases := []struct {
		operatorGroup  *OperatorGroupBuilder
		expectedError  error
		serviceAccount string
	}{
		{
			operatorGroup:  buildValidOperatorGroupBuilder(buildOperatorGroupTestClientWithDummyObject()),
			expectedError:  nil,
			serviceAccount: "test",
		},
		{
			operatorGroup:  buildInValidOperatorGroupBuilder(buildOperatorGroupTestClientWithDummyObject()),
			expectedError:  fmt.Errorf("operatorGroup 'Namespace' cannot be empty"),
			serviceAccount: "test",
		},
		{
			operatorGroup: buildValidOperatorGroupBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: operatorTestSchemes})),
			expectedError:  fmt.Errorf("cannot update non-existent operatorgroup"),
			serviceAccount: "test",
		},
	}

	for _, testCase := range testCases {
		assert.Empty(t, testCase.operatorGroup.Definition.Spec.ServiceAccountName)
		assert.Nil(t, nil, testCase.operatorGroup.Object)
		testCase.operatorGroup.Definition.Spec.ServiceAccountName = testCase.serviceAccount
		testCase.operatorGroup.Definition.ObjectMeta.ResourceVersion = "999"
		_, err := testCase.operatorGroup.Update()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.serviceAccount, testCase.operatorGroup.Object.Spec.ServiceAccountName)
		}
	}
}

func buildInValidOperatorGroupBuilder(apiClient *clients.Settings) *OperatorGroupBuilder {
	return NewOperatorGroupBuilder(apiClient, "operatorgroup", "")
}

func buildValidOperatorGroupBuilder(apiClient *clients.Settings) *OperatorGroupBuilder {
	return NewOperatorGroupBuilder(apiClient, "operatorgroup", "test-namespace")
}

func buildOperatorGroupTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyOperatorGroup(),
		SchemeAttachers: operatorTestSchemes,
	})
}

func buildDummyOperatorGroup() []runtime.Object {
	return append([]runtime.Object{}, &v1.OperatorGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "operatorgroup",
			Namespace: "test-namespace",
		},
		Spec: v1.OperatorGroupSpec{},
	})
}
