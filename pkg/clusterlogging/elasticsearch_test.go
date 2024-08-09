package clusterlogging

import (
	"fmt"
	"testing"

	eskv1 "github.com/openshift/elasticsearch-operator/apis/logging/v1"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	eskAPIGroup                         = "logging.openshift.io"
	eskAPIVersion                       = "v1"
	eskKind                             = "Elasticsearch"
	defaultElasticsearchName            = "elasticsearch"
	defaultElasticsearchNamespace       = "openshift-logging"
	defaultElasticsearchManagementState = eskv1.ManagementState("")
	eskv1TestSchemes                    = []clients.SchemeAttacher{
		eskv1.AddToScheme,
	}
)

func TestElasticsearchPull(t *testing.T) {
	generateElasticsearch := func(name, namespace string) *eskv1.Elasticsearch {
		return &eskv1.Elasticsearch{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: eskv1.ElasticsearchSpec{
				ManagementState: eskv1.ManagementStateManaged,
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
			name:                "test",
			namespace:           "openshift-logging",
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           "openshift-logging",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("elasticsearch 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                "test",
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("elasticsearch 'nsname' cannot be empty"),
			client:              true,
		},
		{
			name:                "esktest",
			namespace:           "openshift-logging",
			addToRuntimeObjects: false,
			expectedError:       fmt.Errorf("elasticsearch object esktest does not exist in namespace openshift-logging"),
			client:              true,
		},
		{
			name:                "esktest",
			namespace:           "openshift-logging",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("elasticsearch 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testElasticsearch := generateElasticsearch(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testElasticsearch)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: eskv1TestSchemes,
			})
		}

		builderResult, err := PullElasticsearch(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		} else {
			assert.Equal(t, testElasticsearch.Name, builderResult.Object.Name)
		}
	}
}

func TestNewElasticsearchBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		expectedError string
	}{
		{
			name:          defaultElasticsearchName,
			namespace:     defaultElasticsearchNamespace,
			expectedError: "",
		},
		{
			name:          "",
			namespace:     defaultElasticsearchNamespace,
			expectedError: "elasticsearch 'name' cannot be empty",
		},
		{
			name:          defaultElasticsearchName,
			namespace:     "",
			expectedError: "elasticsearch 'nsname' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})
		testElasticsearchBuilder := NewElasticsearchBuilder(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, testElasticsearchBuilder.errorMsg)
		assert.NotNil(t, testElasticsearchBuilder.Definition)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.name, testElasticsearchBuilder.Definition.Name)
			assert.Equal(t, testCase.namespace, testElasticsearchBuilder.Definition.Namespace)
		}
	}
}

func TestElasticsearchExist(t *testing.T) {
	testCases := []struct {
		testElasticsearch *ElasticsearchBuilder
		expectedStatus    bool
	}{
		{
			testElasticsearch: buildValidElasticsearchBuilder(buildElasticsearchClientWithDummyObject()),
			expectedStatus:    true,
		},
		{
			testElasticsearch: buildInValidElasticsearchBuilder(buildElasticsearchClientWithDummyObject()),
			expectedStatus:    false,
		},
		{
			testElasticsearch: buildValidElasticsearchBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus:    false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testElasticsearch.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestElasticsearchGet(t *testing.T) {
	testCases := []struct {
		testElasticsearch *ElasticsearchBuilder
		expectedError     error
	}{
		{
			testElasticsearch: buildValidElasticsearchBuilder(buildElasticsearchClientWithDummyObject()),
			expectedError:     nil,
		},
		{
			testElasticsearch: buildInValidElasticsearchBuilder(buildElasticsearchClientWithDummyObject()),
			expectedError:     fmt.Errorf("elasticsearchs.logging.openshift.io \"\" not found"),
		},
		{
			testElasticsearch: buildValidElasticsearchBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:     fmt.Errorf("elasticsearchs.logging.openshift.io \"elasticsearch\" not found"),
		},
	}

	for _, testCase := range testCases {
		elasticsearchObj, err := testCase.testElasticsearch.Get()

		if testCase.expectedError == nil {
			assert.Equal(t, elasticsearchObj, testCase.testElasticsearch.Definition)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestElasticsearchCreate(t *testing.T) {
	testCases := []struct {
		testElasticsearch *ElasticsearchBuilder
		expectedError     string
	}{
		{
			testElasticsearch: buildValidElasticsearchBuilder(buildElasticsearchClientWithDummyObject()),
			expectedError:     "",
		},
		{
			testElasticsearch: buildInValidElasticsearchBuilder(buildElasticsearchClientWithDummyObject()),
			expectedError: fmt.Sprintf("Elasticsearch.logging.openshift.io \"\" is invalid: %s",
				metaDataNameErrorMgs),
		},
		{
			testElasticsearch: buildValidElasticsearchBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:     "resourceVersion can not be set for Create requests",
		},
	}

	for _, testCase := range testCases {
		testElasticsearchBuilder, err := testCase.testElasticsearch.Create()

		if testCase.expectedError == "" {
			assert.Equal(t, testElasticsearchBuilder.Definition, testElasticsearchBuilder.Object)
		} else {
			assert.Equal(t, testCase.expectedError, err.Error())
		}
	}
}

func TestElasticsearchDelete(t *testing.T) {
	testCases := []struct {
		testElasticsearch *ElasticsearchBuilder
		expectedError     error
	}{
		{
			testElasticsearch: buildValidElasticsearchBuilder(buildElasticsearchClientWithDummyObject()),
			expectedError:     nil,
		},
		{
			testElasticsearch: buildInValidElasticsearchBuilder(buildElasticsearchClientWithDummyObject()),
			expectedError:     fmt.Errorf("elasticsearch cannot be deleted because it does not exist"),
		},
		{
			testElasticsearch: buildValidElasticsearchBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:     nil,
		},
	}

	for _, testCase := range testCases {
		err := testCase.testElasticsearch.Delete()

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testElasticsearch.Object)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestElasticsearchUpdate(t *testing.T) {
	testCases := []struct {
		testElasticsearch *ElasticsearchBuilder
		expectedError     string
		managementState   eskv1.ManagementState
	}{
		{
			testElasticsearch: buildValidElasticsearchBuilder(buildElasticsearchClientWithDummyObject()),
			expectedError:     "",
			managementState:   eskv1.ManagementStateManaged,
		},
		{
			testElasticsearch: buildInValidElasticsearchBuilder(buildElasticsearchClientWithDummyObject()),
			expectedError: fmt.Sprintf("Elasticsearch.logging.openshift.io \"\" is invalid: %s",
				metaDataNameErrorMgs),
			managementState: eskv1.ManagementStateManaged,
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, defaultElasticsearchManagementState, testCase.testElasticsearch.Definition.Spec.ManagementState)
		assert.Nil(t, nil, testCase.testElasticsearch.Object)
		testCase.testElasticsearch.WithManagementState(testCase.managementState)
		_, err := testCase.testElasticsearch.Update()

		if testCase.expectedError != "" {
			assert.Equal(t, testCase.expectedError, err.Error())
		} else {
			assert.Equal(t, testCase.managementState, testCase.testElasticsearch.Definition.Spec.ManagementState)
		}
	}
}

func TestElasticsearchWithManagementState(t *testing.T) {
	testCases := []struct {
		testManagementState eskv1.ManagementState
		expectedError       bool
		expectedErrorText   string
	}{
		{
			testManagementState: eskv1.ManagementStateUnmanaged,
			expectedError:       false,
			expectedErrorText:   "",
		},
		{
			testManagementState: eskv1.ManagementStateManaged,
			expectedError:       false,
			expectedErrorText:   "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidElasticsearchBuilder(buildElasticsearchClientWithDummyObject())

		result := testBuilder.WithManagementState(testCase.testManagementState)

		if testCase.expectedError {
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
			}
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testManagementState, result.Definition.Spec.ManagementState)
		}
	}
}

func TestElasticsearchGetManagementState(t *testing.T) {
	testCases := []struct {
		testElasticsearch *ElasticsearchBuilder
		expectedError     error
	}{
		{
			testElasticsearch: buildValidElasticsearchBuilder(buildElasticsearchClientWithDummyObject()),
			expectedError:     nil,
		},
		{
			testElasticsearch: buildValidElasticsearchBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:     fmt.Errorf("elasticsearch object does not exist"),
		},
	}

	for _, testCase := range testCases {
		currentManagementState, err := testCase.testElasticsearch.GetManagementState()

		if testCase.expectedError == nil {
			assert.Equal(t, *currentManagementState, testCase.testElasticsearch.Object.Spec.ManagementState)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func buildValidElasticsearchBuilder(apiClient *clients.Settings) *ElasticsearchBuilder {
	elasticsearchBuilder := NewElasticsearchBuilder(
		apiClient, defaultElasticsearchName, defaultElasticsearchNamespace)
	elasticsearchBuilder.Definition.ResourceVersion = "999"
	elasticsearchBuilder.Definition.Spec.ManagementState = ""
	elasticsearchBuilder.Definition.TypeMeta = metav1.TypeMeta{
		Kind:       eskKind,
		APIVersion: fmt.Sprintf("%s/%s", eskAPIGroup, eskAPIVersion),
	}

	return elasticsearchBuilder
}

func buildInValidElasticsearchBuilder(apiClient *clients.Settings) *ElasticsearchBuilder {
	elasticsearchBuilder := NewElasticsearchBuilder(
		apiClient, "", defaultElasticsearchNamespace)
	elasticsearchBuilder.Definition.ResourceVersion = "999"
	elasticsearchBuilder.Definition.Spec.ManagementState = ""
	elasticsearchBuilder.Definition.TypeMeta = metav1.TypeMeta{
		Kind:       eskKind,
		APIVersion: fmt.Sprintf("%s/%s", eskAPIGroup, eskAPIVersion),
	}

	return elasticsearchBuilder
}

func buildElasticsearchClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyElasticsearch(),
		SchemeAttachers: eskv1TestSchemes,
	})
}

func buildDummyElasticsearch() []runtime.Object {
	return append([]runtime.Object{}, &eskv1.Elasticsearch{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultElasticsearchName,
			Namespace: defaultElasticsearchNamespace,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       eskKind,
			APIVersion: fmt.Sprintf("%s/%s", eskAPIGroup, eskAPIVersion),
		},
		Spec: eskv1.ElasticsearchSpec{
			ManagementState: "",
		},
	})
}
