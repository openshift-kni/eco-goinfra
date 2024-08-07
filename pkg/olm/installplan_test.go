package olm

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	oplmV1alpha1 "github.com/openshift-kni/eco-goinfra/pkg/schemes/olm/operators/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestNewInstallPlanBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		expectedError string
		client        bool
	}{
		{
			name:          "installplan",
			namespace:     "test-namespace",
			client:        true,
			expectedError: "",
		},
		{
			name:          "",
			namespace:     "test-namespace",
			client:        true,
			expectedError: "installplan 'name' cannot be empty",
		},
		{
			name:          "installplan",
			namespace:     "",
			client:        true,
			expectedError: "installplan 'nsname' cannot be empty",
		},
		{
			name:          "installplan",
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

		installPlan := NewInstallPlanBuilder(testSettings, testCase.name, testCase.namespace)

		if testCase.client {
			assert.Equal(t, testCase.expectedError, installPlan.errorMsg)

			if testCase.expectedError == "" {
				assert.NotNil(t, installPlan.Definition)
				assert.Equal(t, testCase.name, installPlan.Definition.Name)
				assert.Equal(t, testCase.namespace, installPlan.Definition.Namespace)
			}
		} else {
			assert.Nil(t, installPlan)
		}
	}
}

func TestPullInstallPlan(t *testing.T) {
	installPlan := func(name, namespace string) *oplmV1alpha1.InstallPlan {
		return &oplmV1alpha1.InstallPlan{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: oplmV1alpha1.InstallPlanSpec{
				CatalogSource: "test",
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
			name:                "installplan",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("installPlan 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                "installplan",
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("installPlan 'nsName' cannot be empty"),
			client:              true,
		},
		{
			name:                "installplan",
			namespace:           "test-namespace",
			addToRuntimeObjects: false,
			expectedError: fmt.Errorf(
				"installPlan object named installplan does not exist in namespace test-namespace"),
			client: true,
		},
		{
			name:                "installplan",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("installPlan 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		installPlan := installPlan(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, installPlan)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects, SchemeAttachers: testSchemes})
		}

		builderResult, err := PullInstallPlan(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, builderResult.Object.Name)
			assert.Equal(t, testCase.namespace, builderResult.Object.Namespace)
		}
	}
}

func TestInstallPlanGet(t *testing.T) {
	testCases := []struct {
		installPlan   *InstallPlanBuilder
		expectedError string
	}{
		{
			installPlan:   buildValidInstallPlanBuilder(buildInstallPlanTestClientWithDummyObject()),
			expectedError: "",
		},
		{
			installPlan:   buildInValidInstallPlanBuilder(buildInstallPlanTestClientWithDummyObject()),
			expectedError: "installplan 'nsname' cannot be empty",
		},
		{
			installPlan: buildValidInstallPlanBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})),
			expectedError: "installplans.operators.coreos.com \"installplan\" not found",
		},
	}

	for _, testCase := range testCases {
		installPlan, err := testCase.installPlan.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, installPlan.Name, testCase.installPlan.Definition.Name)
		} else {
			assert.Equal(t, err.Error(), testCase.expectedError)
		}
	}
}

func TestInstallPlanExist(t *testing.T) {
	testCases := []struct {
		installPlan    *InstallPlanBuilder
		expectedStatus bool
	}{
		{
			installPlan:    buildValidInstallPlanBuilder(buildInstallPlanTestClientWithDummyObject()),
			expectedStatus: true,
		},
		{
			installPlan:    buildInValidInstallPlanBuilder(buildInstallPlanTestClientWithDummyObject()),
			expectedStatus: false,
		},
		{
			installPlan: buildValidInstallPlanBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.installPlan.Exists()
		assert.Equal(t, exist, testCase.expectedStatus)
	}
}

func TestInstallPlanCreate(t *testing.T) {
	testCases := []struct {
		installPlan   *InstallPlanBuilder
		expectedError error
	}{
		{
			installPlan:   buildValidInstallPlanBuilder(buildInstallPlanTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			installPlan:   buildInValidInstallPlanBuilder(buildInstallPlanTestClientWithDummyObject()),
			expectedError: fmt.Errorf("installplan 'nsname' cannot be empty"),
		},
		{
			installPlan: buildValidInstallPlanBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		installPlanBuilder, err := testCase.installPlan.Create()
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil {
			assert.Equal(t, installPlanBuilder.Definition.Name, installPlanBuilder.Object.Name)
		}
	}
}

func TestInstallPlanDelete(t *testing.T) {
	testCases := []struct {
		installPlan   *InstallPlanBuilder
		expectedError error
	}{
		{
			installPlan:   buildValidInstallPlanBuilder(buildInstallPlanTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			installPlan:   buildInValidInstallPlanBuilder(buildInstallPlanTestClientWithDummyObject()),
			expectedError: fmt.Errorf("installplan 'nsname' cannot be empty"),
		},
		{
			installPlan: buildValidInstallPlanBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		err := testCase.installPlan.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.installPlan.Object)
		}
	}
}

func TestInstallPlanUpdate(t *testing.T) {
	testCases := []struct {
		installPlan            *InstallPlanBuilder
		expectedError          error
		catalogSourceNamespace string
	}{
		{
			installPlan:            buildValidInstallPlanBuilder(buildInstallPlanTestClientWithDummyObject()),
			expectedError:          nil,
			catalogSourceNamespace: "test",
		},
		{
			installPlan:            buildInValidInstallPlanBuilder(buildInstallPlanTestClientWithDummyObject()),
			expectedError:          fmt.Errorf("installplan 'nsname' cannot be empty"),
			catalogSourceNamespace: "",
		},
		{
			installPlan: buildValidInstallPlanBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})),
			expectedError: fmt.Errorf(
				"installPlan named installplan in namespace test-namespace does not exist"),
			catalogSourceNamespace: "test",
		},
	}

	for _, testCase := range testCases {
		assert.Empty(t, testCase.installPlan.Definition.Spec.CatalogSourceNamespace)
		assert.Nil(t, nil, testCase.installPlan.Object)
		testCase.installPlan.Definition.Spec.CatalogSourceNamespace = testCase.catalogSourceNamespace
		testCase.installPlan.Definition.ObjectMeta.ResourceVersion = "999"
		_, err := testCase.installPlan.Update()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.catalogSourceNamespace, testCase.installPlan.Object.Spec.CatalogSourceNamespace)
		}
	}
}

func buildInValidInstallPlanBuilder(apiClient *clients.Settings) *InstallPlanBuilder {
	return NewInstallPlanBuilder(apiClient, "installplan", "")
}

func buildValidInstallPlanBuilder(apiClient *clients.Settings) *InstallPlanBuilder {
	return NewInstallPlanBuilder(apiClient, "installplan", "test-namespace")
}

func buildInstallPlanTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyInstallPlan(),
		SchemeAttachers: testSchemes,
	})
}

func buildDummyInstallPlan() []runtime.Object {
	return append([]runtime.Object{}, &oplmV1alpha1.InstallPlan{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "installplan",
			Namespace: "test-namespace",
		},
		Spec: oplmV1alpha1.InstallPlanSpec{
			CatalogSource: "test",
		},
	})
}
