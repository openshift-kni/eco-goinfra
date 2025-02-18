package bmh

import (
	"fmt"
	"testing"

	bmhv1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	dataImageTestName      = "dataimage-test-name"
	dataImageTestNamespace = "dataimage-test-namespace"
)

func TestPullDataImage(t *testing.T) {
	testCases := []struct {
		name                string
		namespace           string
		addToRuntimeObjects bool
		expectedError       error
		client              bool
	}{
		{
			name:                dataImageTestName,
			namespace:           dataImageTestNamespace,
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           dataImageTestNamespace,
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("dataimage 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                dataImageTestName,
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("dataimage 'namespace' cannot be empty"),
			client:              true,
		},
		{
			name:                dataImageTestName,
			namespace:           dataImageTestNamespace,
			addToRuntimeObjects: false,
			expectedError: fmt.Errorf(
				"dataimage object %s does not exist in namespace %s", dataImageTestName, dataImageTestNamespace),
			client: true,
		},
		{
			name:                dataImageTestName,
			namespace:           dataImageTestNamespace,
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("dataimage 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, buildDummyDataImageObject()...)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		builderResult, err := PullDataImage(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, builderResult.Object.Name)
			assert.Equal(t, testCase.namespace, builderResult.Object.Namespace)
		}
	}
}

func TestDataImageDelete(t *testing.T) {
	testCases := []struct {
		testDi        *DataImageBuilder
		expectedError error
	}{
		{
			testDi:        buildValidDataImageBuilder(buildDataImageTestClientWithDummyObject(buildDummyDataImageObject())),
			expectedError: nil,
		},
		{
			testDi:        buildValidDataImageBuilder(buildDataImageTestClientWithDummyObject([]runtime.Object{})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		_, err := testCase.testDi.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testDi.Object)
		}
	}
}

func TestDataImageGet(t *testing.T) {
	testCases := []struct {
		testDataImage *DataImageBuilder
		expectedError error
	}{
		{
			testDataImage: buildValidDataImageBuilder(buildDataImageTestClientWithDummyObject(buildDummyDataImageObject())),
			expectedError: nil,
		},
		{
			testDataImage: buildValidDataImageBuilder(buildDataImageTestClientWithDummyObject([]runtime.Object{})),
			expectedError: fmt.Errorf("dataimages.metal3.io \"%s\" not found", dataImageTestName),
		},
	}

	for _, testCase := range testCases {
		dataImage, err := testCase.testDataImage.Get()

		if testCase.expectedError == nil {
			assert.Equal(t, dataImage.Name, testCase.testDataImage.Definition.Name)
			assert.Equal(t, dataImage.Namespace, testCase.testDataImage.Definition.Namespace)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestDataImageExists(t *testing.T) {
	testCases := []struct {
		testDataImage  *DataImageBuilder
		expectedStatus bool
	}{
		{
			testDataImage:  buildValidDataImageBuilder(buildDataImageTestClientWithDummyObject(buildDummyDataImageObject())),
			expectedStatus: true,
		},
		{
			testDataImage:  buildValidDataImageBuilder(buildDataImageTestClientWithDummyObject([]runtime.Object{})),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testDataImage.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestValidate(t *testing.T) {
	testCases := []struct {
		builderNil    bool
		definitionNil bool
		apiClientNil  bool
		expectedError string
	}{
		{
			builderNil:    true,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: "error: received nil dataimage builder",
		},
		{
			builderNil:    false,
			definitionNil: true,
			apiClientNil:  false,
			expectedError: "can not redefine the undefined dataimage",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  true,
			expectedError: "dataimage builder cannot have nil apiClient",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidDataImageBuilder(buildDataImageTestClientWithDummyObject(buildDummyDataImageObject()))

		if testCase.builderNil {
			testBuilder = nil
		}

		if testCase.definitionNil {
			testBuilder.Definition = nil
		}

		if testCase.apiClientNil {
			testBuilder.apiClient = nil
		}

		result, err := testBuilder.validate()
		if testCase.expectedError != "" {
			assert.NotNil(t, err)
			assert.Equal(t, testCase.expectedError, err.Error())
			assert.False(t, result)
		} else {
			assert.Nil(t, err)
			assert.True(t, result)
		}
	}
}

func buildValidDataImageBuilder(apiClient *clients.Settings) *DataImageBuilder {
	return &DataImageBuilder{
		apiClient:  apiClient.Client,
		Definition: buildDummyDataImage(),
	}
}

func buildDataImageTestClientWithDummyObject(objects []runtime.Object) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  objects,
		SchemeAttachers: testSchemes,
	})
}

func buildDummyDataImageObject() []runtime.Object {
	return append([]runtime.Object{}, buildDummyDataImage())
}

func buildDummyDataImage() *bmhv1alpha1.DataImage {
	return &bmhv1alpha1.DataImage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dataImageTestName,
			Namespace: dataImageTestNamespace,
		},
		Spec: bmhv1alpha1.DataImageSpec{
			URL: "http://test.com",
		},
	}
}
