package olm

import (
	"fmt"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	oplmV1alpha1 "github.com/openshift-kni/eco-goinfra/pkg/schemes/olm/operators/v1alpha1"
	"github.com/stretchr/testify/assert"
)

var (
	testSchemes = []clients.SchemeAttacher{
		oplmV1alpha1.AddToScheme,
	}
)

func TestNewCatalogSourceBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		expectedError string
		client        bool
	}{
		{
			name:          "catalogsource",
			namespace:     "test-namespace",
			client:        true,
			expectedError: "",
		},
		{
			name:          "",
			namespace:     "test-namespace",
			client:        true,
			expectedError: "catalogsource 'name' cannot be empty",
		},
		{
			name:          "catalogsource",
			namespace:     "",
			client:        true,
			expectedError: "catalogsource 'nsname' cannot be empty",
		},
		{
			name:          "catalogsource",
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

		catalogSource := NewCatalogSourceBuilder(testSettings, testCase.name, testCase.namespace)

		if testCase.client {
			assert.Equal(t, testCase.expectedError, catalogSource.errorMsg)

			if testCase.expectedError == "" {
				assert.NotNil(t, catalogSource.Definition)
				assert.Equal(t, testCase.name, catalogSource.Definition.Name)
				assert.Equal(t, testCase.namespace, catalogSource.Definition.Namespace)
			}
		} else {
			assert.Nil(t, catalogSource)
		}
	}
}
func TestPullCatalogSource(t *testing.T) {
	generateCatalogSource := func(name, namespace string) *oplmV1alpha1.CatalogSource {
		return &oplmV1alpha1.CatalogSource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: oplmV1alpha1.CatalogSourceSpec{
				Image: "test",
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
			name:                "catalogsource",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("catalogsource 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                "catalogsource",
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("catalogsource 'namespace' cannot be empty"),
			client:              true,
		},
		{
			name:                "catalogsource",
			namespace:           "test-namespace",
			addToRuntimeObjects: false,
			expectedError:       fmt.Errorf("catalogsource object catalogsource does not exist in namespace test-namespace"),
			client:              true,
		},
		{
			name:                "catalogsource",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("catalogsource 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		catalogSource := generateCatalogSource(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, catalogSource)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects, SchemeAttachers: testSchemes})
		}

		builderResult, err := PullCatalogSource(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, builderResult.Object.Name)
			assert.Equal(t, testCase.namespace, builderResult.Object.Namespace)
		}
	}
}

func TestCatalogSourceGet(t *testing.T) {
	testCases := []struct {
		catalogSource *CatalogSourceBuilder
		expectedError string
	}{
		{
			catalogSource: buildValidCatalogSourceBuilder(buildTestClientWithDummyObject()),
			expectedError: "",
		},
		{
			catalogSource: buildInValidCatalogSourceBuilder(buildTestClientWithDummyObject()),
			expectedError: "catalogsource 'nsname' cannot be empty",
		},
		{
			catalogSource: buildValidCatalogSourceBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})),
			expectedError: "catalogsources.operators.coreos.com \"catalogsource\" not found",
		},
	}

	for _, testCase := range testCases {
		catalogSource, err := testCase.catalogSource.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, catalogSource.Name, testCase.catalogSource.Definition.Name)
		} else {
			assert.Equal(t, err.Error(), testCase.expectedError)
		}
	}
}

func TestCatalogSourceExist(t *testing.T) {
	testCases := []struct {
		catalogSource  *CatalogSourceBuilder
		expectedStatus bool
	}{
		{
			catalogSource:  buildValidCatalogSourceBuilder(buildTestClientWithDummyObject()),
			expectedStatus: true,
		},
		{
			catalogSource:  buildInValidCatalogSourceBuilder(buildTestClientWithDummyObject()),
			expectedStatus: false,
		},
		{
			catalogSource: buildValidCatalogSourceBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.catalogSource.Exists()
		assert.Equal(t, exist, testCase.expectedStatus)
	}
}

func TestCatalogSourceCreate(t *testing.T) {
	testCases := []struct {
		catalogSource *CatalogSourceBuilder
		expectedError error
	}{
		{
			catalogSource: buildValidCatalogSourceBuilder(buildTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			catalogSource: buildInValidCatalogSourceBuilder(buildTestClientWithDummyObject()),
			expectedError: fmt.Errorf("catalogsource 'nsname' cannot be empty"),
		},
		{
			catalogSource: buildValidCatalogSourceBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		catalogSourceBuilder, err := testCase.catalogSource.Create()
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil {
			assert.Equal(t, catalogSourceBuilder.Definition.Name, catalogSourceBuilder.Object.Name)
		}
	}
}

func TestCatalogSourceDelete(t *testing.T) {
	testCases := []struct {
		catalogSource *CatalogSourceBuilder
		expectedError error
	}{
		{
			catalogSource: buildValidCatalogSourceBuilder(buildTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			catalogSource: buildInValidCatalogSourceBuilder(buildTestClientWithDummyObject()),
			expectedError: fmt.Errorf("catalogsource 'nsname' cannot be empty"),
		},
		{
			catalogSource: buildValidCatalogSourceBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		err := testCase.catalogSource.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.catalogSource.Object)
		}
	}
}

func TestCatalogSourceUpdate(t *testing.T) {
	testCases := []struct {
		catalogSource *CatalogSourceBuilder
		expectedError error
		address       string
		force         bool
	}{
		{
			catalogSource: buildValidCatalogSourceBuilder(buildTestClientWithDummyObject()),
			expectedError: nil,
			address:       "test",
			force:         false,
		},
		{
			catalogSource: buildInValidCatalogSourceBuilder(buildTestClientWithDummyObject()),
			expectedError: fmt.Errorf("catalogsource 'nsname' cannot be empty"),
			address:       "",
			force:         false,
		},
		{
			catalogSource: buildValidCatalogSourceBuilder(buildTestClientWithDummyObject()),
			expectedError: nil,
			address:       "test",
			force:         true,
		},
	}

	for _, testCase := range testCases {
		assert.Empty(t, testCase.catalogSource.Definition.Spec.Address)
		assert.Nil(t, nil, testCase.catalogSource.Object)
		testCase.catalogSource.Definition.Spec.Address = testCase.address
		testCase.catalogSource.Definition.ObjectMeta.ResourceVersion = "999"
		_, err := testCase.catalogSource.Update(false)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.address, testCase.catalogSource.Definition.Spec.Address)
		}
	}
}

func buildValidCatalogSourceBuilder(apiClient *clients.Settings) *CatalogSourceBuilder {
	return NewCatalogSourceBuilder(apiClient, "catalogsource", "test-namespace")
}

func buildInValidCatalogSourceBuilder(apiClient *clients.Settings) *CatalogSourceBuilder {
	return NewCatalogSourceBuilder(apiClient, "catalogsource", "")
}

func buildTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyCatalogSource(),
		SchemeAttachers: testSchemes,
	})
}

func buildDummyCatalogSource() []runtime.Object {
	return append([]runtime.Object{}, &oplmV1alpha1.CatalogSource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "catalogsource",
			Namespace: "test-namespace",
		},
		Spec: oplmV1alpha1.CatalogSourceSpec{
			Image: "test",
		},
	})
}
