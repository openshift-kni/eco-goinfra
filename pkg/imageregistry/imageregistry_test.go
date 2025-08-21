package imageregistry

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/glog"

	imageregistryV1 "github.com/openshift/api/imageregistry/v1"
	operatorV1 "github.com/openshift/api/operator/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	imageRegistryGVK = schema.GroupVersionKind{
		Group:   APIGroup,
		Version: APIVersion,
		Kind:    APIKind,
	}
	defaultImageRegistryName      = "cluster"
	defaultManagementState        = operatorV1.Managed
	defaultImageRegistryCondition = operatorV1.OperatorCondition{
		Type:   "Available",
		Status: operatorV1.ConditionTrue,
		Reason: "Ready",
	}
)

func TestImageRegistryPull(t *testing.T) {
	generateImageRegistry := func(name string) *imageregistryV1.Config {
		return &imageregistryV1.Config{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: imageregistryV1.ImageRegistrySpec{
				OperatorSpec: operatorV1.OperatorSpec{
					ManagementState: operatorV1.Removed,
				},
				Storage: imageregistryV1.ImageRegistryConfigStorage{},
			},
		}
	}

	testCases := []struct {
		name                string
		addToRuntimeObjects bool
		expectedError       error
		client              bool
	}{
		{
			name:                "test",
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("imageRegistry 'imageRegistryObjName' cannot be empty"),
			client:              true,
		},
		{
			name:                "irtest",
			addToRuntimeObjects: false,
			expectedError:       fmt.Errorf("imageRegistry object irtest does not exist"),
			client:              true,
		},
		{
			name:                "irtest",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("imageRegistry Config 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testImageRegistry := generateImageRegistry(testCase.name)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testImageRegistry)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})
		}

		builderResult, err := Pull(testSettings, testCase.name)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError, err)
		} else {
			assert.Equal(t, testImageRegistry.Name, builderResult.Object.Name)
		}
	}
}

func TestImageRegistryExist(t *testing.T) {
	testCases := []struct {
		testImageRegistryConfig *Builder
		expectedStatus          bool
	}{
		{
			testImageRegistryConfig: buildValidImageRegistryBuilder(buildImageRegistryClientWithDummyObject()),
			expectedStatus:          true,
		},
		{
			testImageRegistryConfig: buildInValidImageRegistryBuilder(buildImageRegistryClientWithDummyObject()),
			expectedStatus:          false,
		},
		{
			testImageRegistryConfig: buildValidImageRegistryBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus:          false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testImageRegistryConfig.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestImageRegistryGet(t *testing.T) {
	testCases := []struct {
		testImageRegistry *Builder
		expectedError     error
	}{
		{
			testImageRegistry: buildValidImageRegistryBuilder(buildImageRegistryClientWithDummyObject()),
			expectedError:     nil,
		},
		{
			testImageRegistry: buildInValidImageRegistryBuilder(buildImageRegistryClientWithDummyObject()),
			expectedError:     fmt.Errorf("the imageRegistry 'name' cannot be empty"),
		},
		{
			testImageRegistry: buildValidImageRegistryBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:     fmt.Errorf("configs.imageregistry.operator.openshift.io \"cluster\" not found"),
		},
	}

	for _, testCase := range testCases {
		imageRegistryObj, err := testCase.testImageRegistry.Get()

		if testCase.expectedError == nil {
			assert.Equal(t, imageRegistryObj, testCase.testImageRegistry.Definition)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestImageRegistryUpdate(t *testing.T) {
	testCases := []struct {
		testImageRegistry *Builder
		expectedError     error
		managementState   operatorV1.ManagementState
	}{
		{
			testImageRegistry: buildValidImageRegistryBuilder(buildImageRegistryClientWithDummyObject()),
			expectedError:     nil,
			managementState:   operatorV1.Managed,
		},
		{
			testImageRegistry: buildInValidImageRegistryBuilder(buildImageRegistryClientWithDummyObject()),
			expectedError:     fmt.Errorf("the imageRegistry 'name' cannot be empty"),
			managementState:   operatorV1.Managed,
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, defaultManagementState, testCase.testImageRegistry.Definition.Spec.ManagementState)
		assert.Nil(t, nil, testCase.testImageRegistry.Object)
		testCase.testImageRegistry.WithManagementState(testCase.managementState)
		_, err := testCase.testImageRegistry.Update()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.managementState, testCase.testImageRegistry.Definition.Spec.ManagementState)
		}
	}
}

func TestImageRegistryWithManagementState(t *testing.T) {
	testCases := []struct {
		testManagementState operatorV1.ManagementState
		expectedError       bool
		expectedErrorText   string
	}{
		{
			testManagementState: operatorV1.Removed,
			expectedError:       false,
			expectedErrorText:   "",
		},
		{
			testManagementState: operatorV1.Managed,
			expectedError:       false,
			expectedErrorText:   "",
		},
		{
			testManagementState: operatorV1.Unmanaged,
			expectedError:       false,
			expectedErrorText:   "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidImageRegistryBuilder(buildImageRegistryClientWithDummyObject())

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

func TestImageRegistryGetManagementState(t *testing.T) {
	testCases := []struct {
		testImageRegistry *Builder
		expectedError     error
	}{
		{
			testImageRegistry: buildValidImageRegistryBuilder(buildImageRegistryClientWithDummyObject()),
			expectedError:     nil,
		},
		{
			testImageRegistry: buildValidImageRegistryBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:     fmt.Errorf("imageRegistry object does not exist"),
		},
	}

	for _, testCase := range testCases {
		currentManagementState, err := testCase.testImageRegistry.GetManagementState()

		if testCase.expectedError == nil {
			assert.Equal(t, *currentManagementState, testCase.testImageRegistry.Object.Spec.ManagementState)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestImageRegistryWithStorage(t *testing.T) {
	testCases := []struct {
		testStorageConfig imageregistryV1.ImageRegistryConfigStorage
		expectedError     bool
		expectedErrorText string
	}{
		{
			testStorageConfig: imageregistryV1.ImageRegistryConfigStorage{
				EmptyDir: &imageregistryV1.ImageRegistryConfigStorageEmptyDir{},
			},
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			testStorageConfig: imageregistryV1.ImageRegistryConfigStorage{},
			expectedError:     false,
			expectedErrorText: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidImageRegistryBuilder(buildImageRegistryClientWithDummyObject())

		result := testBuilder.WithStorage(testCase.testStorageConfig)

		if testCase.expectedError {
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
			}
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testStorageConfig, result.Definition.Spec.Storage)
		}
	}
}

func TestImageRegistryGetStorageConfig(t *testing.T) {
	testCases := []struct {
		testImageRegistry *Builder
		expectedError     error
	}{
		{
			testImageRegistry: buildValidImageRegistryBuilder(buildImageRegistryClientWithDummyObject()),
			expectedError:     nil,
		},
		{
			testImageRegistry: buildValidImageRegistryBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:     fmt.Errorf("imageRegistry object does not exist"),
		},
	}

	for _, testCase := range testCases {
		currentStorageConfig, err := testCase.testImageRegistry.GetStorageConfig()

		if testCase.expectedError == nil {
			assert.Equal(t, *currentStorageConfig, testCase.testImageRegistry.Object.Spec.Storage)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestImageRegistryWaitForCondition(t *testing.T) {
	testCases := []struct {
		exists        bool
		conditionMet  bool
		valid         bool
		expectedError error
	}{
		{
			exists:        true,
			conditionMet:  true,
			valid:         true,
			expectedError: nil,
		},
		{
			exists:        false,
			conditionMet:  true,
			valid:         true,
			expectedError: fmt.Errorf("imageRegistry object %s does not exist", defaultImageRegistryName),
		},
		{
			exists:        true,
			conditionMet:  false,
			valid:         true,
			expectedError: context.DeadlineExceeded,
		},
		{
			exists:        true,
			conditionMet:  true,
			valid:         false,
			expectedError: fmt.Errorf("the imageRegistry 'name' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testBuilder    *Builder
		)

		if testCase.exists {
			imageRegistry := buildDummyImageRegistry(defaultImageRegistryName, defaultManagementState)

			if testCase.conditionMet {
				imageRegistry.Status.Conditions = append(imageRegistry.Status.Conditions, defaultImageRegistryCondition)
			}

			runtimeObjects = append(runtimeObjects, imageRegistry)
		}

		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		if testCase.valid {
			testBuilder = buildValidImageRegistryBuilder(testSettings)
		} else {
			testBuilder = buildInValidImageRegistryBuilder(testSettings)
		}

		_, err := testBuilder.WaitForCondition(defaultImageRegistryCondition, time.Second)
		assert.Equal(t, testCase.expectedError, err)
	}
}

func buildValidImageRegistryBuilder(apiClient *clients.Settings) *Builder {
	return newBuilder(apiClient, defaultImageRegistryName, defaultManagementState)
}

func buildInValidImageRegistryBuilder(apiClient *clients.Settings) *Builder {
	return newBuilder(apiClient, "", defaultManagementState)
}

func buildImageRegistryClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: buildDummyImageRegistryObject(),
		GVK:            []schema.GroupVersionKind{imageRegistryGVK},
	})
}

func buildDummyImageRegistry(name string, managementState operatorV1.ManagementState) *imageregistryV1.Config {
	return &imageregistryV1.Config{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: imageregistryV1.ImageRegistrySpec{
			OperatorSpec: operatorV1.OperatorSpec{
				ManagementState: managementState,
			},
			Storage: imageregistryV1.ImageRegistryConfigStorage{},
		},
	}
}

func buildDummyImageRegistryObject() []runtime.Object {
	return append([]runtime.Object{}, buildDummyImageRegistry(defaultImageRegistryName, defaultManagementState))
}

// newBuilder method creates new instance of builder (for the unit test propose only).
func newBuilder(apiClient *clients.Settings, name string, managementState operatorV1.ManagementState) *Builder {
	glog.V(100).Infof("Initializing new Builder structure with the name: %s", name)

	builder := &Builder{
		apiClient:  apiClient.Client,
		Definition: buildDummyImageRegistry(name, managementState),
	}

	builder.Definition.ResourceVersion = "999"

	if name == "" {
		glog.V(100).Infof("The name of the imageRegistry is empty")

		builder.errorMsg = "the imageRegistry 'name' cannot be empty"

		return builder
	}

	return builder
}
