package storage

import (
	"fmt"
	"testing"

	odfoperatorv1alpha1 "github.com/red-hat-storage/odf-operator/api/v1alpha1"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultSystemODFName      = "ocs-cluster-system"
	defaultSystemODFNamespace = "openshift-odf"
)

func TestPullSystemODF(t *testing.T) {
	generateSystemODF := func(name, namespace string) *odfoperatorv1alpha1.StorageSystem {
		return &odfoperatorv1alpha1.StorageSystem{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
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
			name:                defaultSystemODFName,
			namespace:           defaultSystemODFNamespace,
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           defaultSystemODFNamespace,
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("SystemODF 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                defaultSystemODFName,
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("SystemODF 'namespace' cannot be empty"),
			client:              true,
		},
		{
			name:                "odftest",
			namespace:           defaultSystemODFNamespace,
			addToRuntimeObjects: false,
			expectedError:       fmt.Errorf("SystemODF object odftest does not exist in namespace openshift-odf"),
			client:              true,
		},
		{
			name:                "odftest",
			namespace:           defaultSystemODFNamespace,
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("SystemODF 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testSystemODF := generateSystemODF(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testSystemODF)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})
		}

		builderResult, err := PullSystemODF(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testSystemODF.Name, builderResult.Object.Name)
			assert.Equal(t, testSystemODF.Namespace, builderResult.Object.Namespace)
		}
	}
}

func TestNewSystemODFBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		expectedError string
		client        bool
	}{
		{
			name:          defaultSystemODFName,
			namespace:     defaultSystemODFNamespace,
			expectedError: "",
			client:        true,
		},
		{
			name:          "",
			namespace:     defaultSystemODFNamespace,
			expectedError: "SystemODF 'name' cannot be empty",
			client:        true,
		},
		{
			name:          defaultSystemODFName,
			namespace:     "",
			expectedError: "SystemODF 'nsname' cannot be empty",
			client:        true,
		},
		{
			name:          defaultSystemODFName,
			namespace:     defaultSystemODFNamespace,
			expectedError: "",
			client:        false,
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		testSystemODFBuilder := NewSystemODFBuilder(testSettings, testCase.name, testCase.namespace)

		if testCase.expectedError == "" {
			if testCase.client {
				assert.Equal(t, testCase.name, testSystemODFBuilder.Definition.Name)
				assert.Equal(t, testCase.namespace, testSystemODFBuilder.Definition.Namespace)
			} else {
				assert.Nil(t, testSystemODFBuilder)
			}
		} else {
			assert.Equal(t, testCase.expectedError, testSystemODFBuilder.errorMsg)
			assert.NotNil(t, testSystemODFBuilder.Definition)
		}
	}
}

func TestSystemODFExist(t *testing.T) {
	testCases := []struct {
		testSystemODF  *SystemODFBuilder
		expectedStatus bool
	}{
		{
			testSystemODF:  buildValidSystemODFBuilder(buildSystemODFClientWithDummyObject()),
			expectedStatus: true,
		},
		{
			testSystemODF:  buildInValidSystemODFBuilder(buildSystemODFClientWithDummyObject()),
			expectedStatus: false,
		},
		{
			testSystemODF:  buildValidSystemODFBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testSystemODF.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestSystemODFGet(t *testing.T) {
	testCases := []struct {
		testSystemODF *SystemODFBuilder
		expectedError error
	}{
		{
			testSystemODF: buildValidSystemODFBuilder(buildSystemODFClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testSystemODF: buildInValidSystemODFBuilder(buildSystemODFClientWithDummyObject()),
			expectedError: fmt.Errorf("SystemODF 'name' cannot be empty"),
		},
		{
			testSystemODF: buildValidSystemODFBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("storagesystems.odf.openshift.io \"ocs-cluster-system\" " +
				"not found"),
		},
	}

	for _, testCase := range testCases {
		systemODFObj, err := testCase.testSystemODF.Get()

		if testCase.expectedError == nil {
			assert.Equal(t, systemODFObj.Name, testCase.testSystemODF.Definition.Name)
			assert.Equal(t, systemODFObj.Namespace, testCase.testSystemODF.Definition.Namespace)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestSystemODFCreate(t *testing.T) {
	testCases := []struct {
		testSystemODF *SystemODFBuilder
		expectedError error
	}{
		{
			testSystemODF: buildValidSystemODFBuilder(buildSystemODFClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testSystemODF: buildInValidSystemODFBuilder(buildSystemODFClientWithDummyObject()),
			expectedError: fmt.Errorf("SystemODF 'name' cannot be empty"),
		},
		{
			testSystemODF: buildValidSystemODFBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		testSystemODFBuilder, err := testCase.testSystemODF.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testSystemODFBuilder.Definition.Name, testSystemODFBuilder.Object.Name)
			assert.Equal(t, testSystemODFBuilder.Definition.Namespace, testSystemODFBuilder.Object.Namespace)
		}
	}
}

func TestSystemODFDelete(t *testing.T) {
	testCases := []struct {
		testSystemODF *SystemODFBuilder
		expectedError error
	}{
		{
			testSystemODF: buildValidSystemODFBuilder(buildSystemODFClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testSystemODF: buildInValidSystemODFBuilder(buildSystemODFClientWithDummyObject()),
			expectedError: fmt.Errorf("SystemODF 'name' cannot be empty"),
		},
		{
			testSystemODF: buildValidSystemODFBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		err := testCase.testSystemODF.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testSystemODF.Object)
			assert.Nil(t, err)
		}
	}
}

func TestSystemODFWithSpec(t *testing.T) {
	testCases := []struct {
		testKind      odfoperatorv1alpha1.StorageKind
		testName      string
		testNamespace string
		expectedError string
	}{
		{
			testKind:      odfoperatorv1alpha1.StorageKind("cluster.ocs.openshift.io/v1"),
			testName:      "ocs-cluster",
			testNamespace: defaultSystemODFNamespace,
			expectedError: "",
		},
		{
			testKind:      odfoperatorv1alpha1.StorageKind(""),
			testName:      "ocs-cluster",
			testNamespace: defaultSystemODFNamespace,
			expectedError: "SystemODF spec 'kind' cannot be empty",
		},
		{
			testKind:      odfoperatorv1alpha1.StorageKind("cluster.ocs.openshift.io/v1"),
			testName:      "",
			testNamespace: defaultSystemODFNamespace,
			expectedError: "SystemODF spec 'name' cannot be empty",
		},
		{
			testKind:      odfoperatorv1alpha1.StorageKind("storagecluster.ocs.openshift.io/v1"),
			testName:      "ocs-cluster",
			testNamespace: "",
			expectedError: "SystemODF spec 'nsname' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidSystemODFBuilder(buildSystemODFClientWithDummyObject())

		result := testBuilder.WithSpec(testCase.testKind, testCase.testName, testCase.testNamespace)
		assert.Equal(t, testCase.expectedError, result.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testKind, result.Definition.Spec.Kind)
			assert.Equal(t, testCase.testName, result.Definition.Spec.Name)
			assert.Equal(t, testCase.testNamespace, result.Definition.Spec.Namespace)
		}
	}
}

func buildValidSystemODFBuilder(apiClient *clients.Settings) *SystemODFBuilder {
	systemODFBuilder := NewSystemODFBuilder(
		apiClient, defaultSystemODFName, defaultSystemODFNamespace)

	return systemODFBuilder
}

func buildInValidSystemODFBuilder(apiClient *clients.Settings) *SystemODFBuilder {
	systemODFBuilder := NewSystemODFBuilder(
		apiClient, "", defaultSystemODFNamespace)

	return systemODFBuilder
}

func buildSystemODFClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: buildDummySystemODF(),
	})
}

func buildDummySystemODF() []runtime.Object {
	return append([]runtime.Object{}, &odfoperatorv1alpha1.StorageSystem{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultSystemODFName,
			Namespace: defaultSystemODFNamespace,
		},
	})
}
