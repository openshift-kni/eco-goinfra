package storage

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	storageV1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	defaultStorageClassProvisioner = "test-provisioner"
	defaultParameterKey            = "readOnly"
	defaultParameterValue          = "false"
)

func TestNewClassBuilder(t *testing.T) {
	testCases := []struct {
		storageClassName        string
		storageClassProvisioner string
		client                  bool
		expectedErrorText       string
	}{
		{
			storageClassName:        defaultStorageClassName,
			storageClassProvisioner: defaultStorageClassProvisioner,
			client:                  true,
			expectedErrorText:       "",
		},
		{
			storageClassName:        "",
			storageClassProvisioner: defaultStorageClassProvisioner,
			client:                  true,
			expectedErrorText:       "storageclass 'name' cannot be empty",
		},
		{
			storageClassName:        defaultStorageClassName,
			storageClassProvisioner: "",
			client:                  true,
			expectedErrorText:       "storageclass 'provisioner' cannot be empty",
		},
		{
			storageClassName:        defaultStorageClassName,
			storageClassProvisioner: defaultStorageClassProvisioner,
			client:                  false,
			expectedErrorText:       "",
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		classBuilder := NewClassBuilder(testSettings, testCase.storageClassName, testCase.storageClassProvisioner)
		if testCase.client {
			assert.Equal(t, testCase.expectedErrorText, classBuilder.errorMsg)

			if testCase.expectedErrorText == "" {
				assert.Equal(t, testCase.storageClassName, classBuilder.Definition.Name)
				assert.Equal(t, testCase.storageClassProvisioner, classBuilder.Definition.Provisioner)
			}
		} else {
			assert.Nil(t, classBuilder)
		}
	}
}

func TestClassWithReclaimPolicy(t *testing.T) {
	testCases := []struct {
		reclaimPolicy     corev1.PersistentVolumeReclaimPolicy
		expectedErrorText string
	}{
		{
			reclaimPolicy:     corev1.PersistentVolumeReclaimDelete,
			expectedErrorText: "",
		},
		{
			reclaimPolicy:     "",
			expectedErrorText: "storageclass 'reclaimPolicy' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})
		classBuilder := buildValidClassTestBuilder(testSettings).WithReclaimPolicy(testCase.reclaimPolicy)
		assert.Equal(t, testCase.expectedErrorText, classBuilder.errorMsg)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, testCase.reclaimPolicy, *classBuilder.Definition.ReclaimPolicy)
		}
	}
}

func TestClassWithVolumeBindingMode(t *testing.T) {
	testCases := []struct {
		volumeBindingMode storageV1.VolumeBindingMode
		expectedErrorText string
	}{
		{
			volumeBindingMode: storageV1.VolumeBindingImmediate,
			expectedErrorText: "",
		},
		{
			volumeBindingMode: "",
			expectedErrorText: "storageclass 'bindingMode' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})
		classBuilder := buildValidClassTestBuilder(testSettings).WithVolumeBindingMode(testCase.volumeBindingMode)
		assert.Equal(t, testCase.expectedErrorText, classBuilder.errorMsg)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, testCase.volumeBindingMode, *classBuilder.Definition.VolumeBindingMode)
		}
	}
}

func TestClassWithParameter(t *testing.T) {
	testCases := []struct {
		parameterKey      string
		parameterValue    string
		expectedErrorText string
	}{
		{
			parameterKey:      defaultParameterKey,
			parameterValue:    defaultParameterValue,
			expectedErrorText: "",
		},
		{
			parameterKey:      "",
			parameterValue:    defaultParameterValue,
			expectedErrorText: "storageclass parameter key cannot be empty",
		},
		{
			parameterKey:      defaultParameterKey,
			parameterValue:    "",
			expectedErrorText: "storageclass parameter value cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})
		classBuilder := buildValidClassTestBuilder(testSettings).WithParameter(testCase.parameterKey, testCase.parameterValue)
		assert.Equal(t, testCase.expectedErrorText, classBuilder.errorMsg)

		if testCase.expectedErrorText == "" {
			assert.Contains(t, classBuilder.Definition.Parameters, testCase.parameterKey)
			assert.Equal(t, classBuilder.Definition.Parameters[testCase.parameterKey], testCase.parameterValue)
		}
	}
}

func TestClassWithOptions(t *testing.T) {
	testCases := []struct {
		testBuilder       *ClassBuilder
		option            AdditionalOptions
		expectedErrorText string
	}{
		{
			testBuilder: buildValidClassTestBuilder(buildTestClientWithDummyStorageClass()),
			option: func(builder *ClassBuilder) (*ClassBuilder, error) {
				builder.Definition.MountOptions = []string{"ro"}

				return builder, nil
			},
			expectedErrorText: "",
		},
		{
			testBuilder: buildValidClassTestBuilder(buildTestClientWithDummyStorageClass()),
			option: func(builder *ClassBuilder) (*ClassBuilder, error) {
				return builder, fmt.Errorf("error in mutation function")
			},
			expectedErrorText: "error in mutation function",
		},
		{
			testBuilder:       buildInvalidClassTestBuilder(buildTestClientWithDummyStorageClass()),
			option:            nil,
			expectedErrorText: "storageclass 'provisioner' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testCase.testBuilder.WithOptions(testCase.option)
		assert.Equal(t, testCase.expectedErrorText, testCase.testBuilder.errorMsg)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, []string{"ro"}, testCase.testBuilder.Definition.MountOptions)
		}
	}
}

func TestPullClass(t *testing.T) {
	testCases := []struct {
		storageClassName    string
		addToRuntimeObjects bool
		client              bool
		expectedErrorText   string
	}{
		{
			storageClassName:    defaultStorageClassName,
			addToRuntimeObjects: true,
			client:              true,
			expectedErrorText:   "",
		},
		{
			storageClassName:    "",
			addToRuntimeObjects: true,
			client:              true,
			expectedErrorText:   "storageclass 'name' cannot be empty",
		},
		{
			storageClassName:    defaultStorageClassName,
			addToRuntimeObjects: false,
			client:              true,
			expectedErrorText:   fmt.Sprintf("storageclass object %s does not exist", defaultStorageClassName),
		},
		{
			storageClassName:    defaultStorageClassName,
			addToRuntimeObjects: true,
			client:              false,
			expectedErrorText:   "storageclass 'apiClient' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testStorageClass := buildDummyStorageClass(testCase.storageClassName, defaultStorageClassProvisioner)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testStorageClass)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})
		}

		classBuilder, err := PullClass(testSettings, testCase.storageClassName)

		if testCase.expectedErrorText != "" {
			assert.EqualError(t, err, testCase.expectedErrorText)
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testStorageClass.Name, classBuilder.Definition.Name)
		}
	}
}

func TestClassExists(t *testing.T) {
	testCases := []struct {
		testBuilder *ClassBuilder
		exists      bool
	}{
		{
			testBuilder: buildValidClassTestBuilder(buildTestClientWithDummyStorageClass()),
			exists:      true,
		},
		{
			testBuilder: buildInvalidClassTestBuilder(buildTestClientWithDummyStorageClass()),
			exists:      false,
		},
		{
			testBuilder: buildValidClassTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

func TestClassCreate(t *testing.T) {
	testCases := []struct {
		testBuilder   *ClassBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidClassTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidClassTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("storageclass 'provisioner' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		classBuilder, err := testCase.testBuilder.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, classBuilder.Definition, classBuilder.Object)
		}
	}
}

func TestClassDelete(t *testing.T) {
	testCases := []struct {
		testBuilder       *ClassBuilder
		expectedErrorText string
	}{
		{
			testBuilder:       buildValidClassTestBuilder(buildTestClientWithDummyStorageClass()),
			expectedErrorText: "",
		},
		{
			testBuilder:       buildValidClassTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedErrorText: "",
		},
		{
			testBuilder:       buildInvalidClassTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedErrorText: "storageclass 'provisioner' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		err := testCase.testBuilder.Delete()

		if testCase.expectedErrorText == "" {
			assert.Nil(t, err)
			assert.Nil(t, testCase.testBuilder.Object)
		}
	}
}

func TestClassDeleteAndWait(t *testing.T) {
	testCases := []struct {
		testBuilder       *ClassBuilder
		expectedErrorText string
	}{
		{
			testBuilder:       buildValidClassTestBuilder(buildTestClientWithDummyStorageClass()),
			expectedErrorText: "",
		},
		{
			testBuilder:       buildValidClassTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedErrorText: "",
		},
		{
			testBuilder:       buildInvalidClassTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedErrorText: "storageclass 'provisioner' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		err := testCase.testBuilder.DeleteAndWait(time.Second)

		if testCase.expectedErrorText == "" {
			assert.Nil(t, err)
			assert.Nil(t, testCase.testBuilder.Object)
		}
	}
}

func TestClassWaitUntilDeleted(t *testing.T) {
	testCases := []struct {
		testBuilder   *ClassBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidClassTestBuilder(buildTestClientWithDummyStorageClass()),
			expectedError: context.DeadlineExceeded,
		},
		{
			testBuilder:   buildValidClassTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidClassTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("storageclass 'provisioner' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		err := testCase.testBuilder.WaitUntilDeleted(time.Second)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testBuilder.Object)
		}
	}
}

func TestClassUpdate(t *testing.T) {
	testCases := []struct {
		alreadyExists     bool
		force             bool
		expectedErrorText string
	}{
		{
			alreadyExists:     false,
			force:             false,
			expectedErrorText: "cannot update non-existent storageclass",
		},
		{
			alreadyExists:     true,
			force:             false,
			expectedErrorText: "",
		},
		{
			alreadyExists:     false,
			force:             true,
			expectedErrorText: "cannot update non-existent storageclass",
		},
		{
			alreadyExists:     true,
			force:             true,
			expectedErrorText: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidClassTestBuilder(clients.GetTestClients(clients.TestClientParams{}))

		// Create the builder rather than just adding it to the client so that the proper metadata is added and
		// the update will not fail.
		if testCase.alreadyExists {
			var err error

			testBuilder = buildValidClassTestBuilder(clients.GetTestClients(clients.TestClientParams{}))
			testBuilder, err = testBuilder.Create()
			assert.Nil(t, err)
			assert.True(t, testBuilder.Exists())
		}

		assert.NotNil(t, testBuilder.Definition)
		assert.Empty(t, testBuilder.Definition.MountOptions)

		testBuilder.Definition.MountOptions = []string{"ro"}

		classBuilder, err := testBuilder.Update(testCase.force)
		assert.NotNil(t, testBuilder.Definition)

		if testCase.expectedErrorText == "" {
			assert.Nil(t, err)
			assert.Equal(t, testBuilder.Definition.Name, classBuilder.Definition.Name)
			assert.Equal(t, testBuilder.Definition.MountOptions, classBuilder.Definition.MountOptions)
		} else {
			assert.EqualError(t, err, testCase.expectedErrorText)
		}
	}
}

func TestClassValidate(t *testing.T) {
	testCases := []struct {
		builderNil      bool
		definitionNil   bool
		apiClientNil    bool
		builderErrorMsg string
		expectedError   error
	}{
		{
			builderNil:      false,
			definitionNil:   false,
			apiClientNil:    false,
			builderErrorMsg: "",
			expectedError:   nil,
		},
		{
			builderNil:      true,
			definitionNil:   false,
			apiClientNil:    false,
			builderErrorMsg: "",
			expectedError:   fmt.Errorf("error: received nil storageClass builder"),
		},
		{
			builderNil:      false,
			definitionNil:   true,
			apiClientNil:    false,
			builderErrorMsg: "",
			expectedError:   fmt.Errorf("can not redefine the undefined storageClass"),
		},
		{
			builderNil:      false,
			definitionNil:   false,
			apiClientNil:    true,
			builderErrorMsg: "",
			expectedError:   fmt.Errorf("storageClass builder cannot have nil apiClient"),
		},
		{
			builderNil:      false,
			definitionNil:   false,
			apiClientNil:    false,
			builderErrorMsg: "test error",
			expectedError:   fmt.Errorf("test error"),
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidClassTestBuilder(clients.GetTestClients(clients.TestClientParams{}))

		if testCase.builderNil {
			testBuilder = nil
		}

		if testCase.definitionNil {
			testBuilder.Definition = nil
		}

		if testCase.apiClientNil {
			testBuilder.apiClient = nil
		}

		if testCase.builderErrorMsg != "" {
			testBuilder.errorMsg = testCase.builderErrorMsg
		}

		valid, err := testBuilder.validate()
		if testCase.expectedError != nil {
			assert.False(t, valid)
			assert.Equal(t, testCase.expectedError, err)
		} else {
			assert.True(t, valid)
			assert.Nil(t, err)
		}
	}
}

// buildDummyStorageClass returns a StorageClass with the provided name and provisioner.
func buildDummyStorageClass(name, provisioner string) *storageV1.StorageClass {
	return &storageV1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Provisioner: provisioner,
	}
}

// buildTestClientWithDummyStorageClass returns a client with a mock StorageClass.
func buildTestClientWithDummyStorageClass() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyStorageClass(defaultStorageClassName, defaultStorageClassProvisioner),
		},
	})
}

// buildValidClassTestBuilder returns a valid ClassBuilder for testing.
func buildValidClassTestBuilder(apiClient *clients.Settings) *ClassBuilder {
	return NewClassBuilder(apiClient, defaultStorageClassName, defaultStorageClassProvisioner)
}

// buildInvalidClassTestBuilder returns an invalid ClassBuilder for testing.
func buildInvalidClassTestBuilder(apiClient *clients.Settings) *ClassBuilder {
	return NewClassBuilder(apiClient, defaultStorageClassName, "")
}
