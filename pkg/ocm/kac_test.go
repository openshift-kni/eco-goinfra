package ocm

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	kacv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	defaultKACName      = "klusterletaddonconfig-test"
	defaultKACNamespace = "test-ns"
)

func TestNewKACBuilder(t *testing.T) {
	testCases := []struct {
		kacName           string
		kacNamespace      string
		client            bool
		expectedErrorText string
	}{
		{
			kacName:           defaultKACName,
			kacNamespace:      defaultKACNamespace,
			client:            true,
			expectedErrorText: "",
		},
		{
			kacName:           "",
			kacNamespace:      defaultKACNamespace,
			client:            true,
			expectedErrorText: "klusterletAddonConfig 'name' cannot be empty",
		},
		{
			kacName:           defaultKACName,
			kacNamespace:      "",
			client:            true,
			expectedErrorText: "klusterletAddonConfig 'nsname' cannot be empty",
		},
		{
			kacName:           defaultKACName,
			kacNamespace:      defaultKACNamespace,
			client:            false,
			expectedErrorText: "klusterletAddonConfig 'apiClient' cannot be nil",
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		kacBuilder := NewKACBuilder(testSettings, testCase.kacName, testCase.kacNamespace)
		assert.Equal(t, testCase.expectedErrorText, kacBuilder.errorMsg)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, testCase.kacName, kacBuilder.Definition.Name)
			assert.Equal(t, testCase.kacNamespace, kacBuilder.Definition.Namespace)
		}
	}
}

func TestPullKAC(t *testing.T) {
	testCases := []struct {
		kacName             string
		kacNamespace        string
		addToRuntimeObjects bool
		client              bool
		expectedErrorText   string
	}{
		{
			kacName:             defaultKACName,
			kacNamespace:        defaultKACNamespace,
			addToRuntimeObjects: true,
			client:              true,
			expectedErrorText:   "",
		},
		{
			kacName:             "",
			kacNamespace:        defaultKACNamespace,
			addToRuntimeObjects: true,
			client:              true,
			expectedErrorText:   "klusterletAddonConfig 'name' cannot be empty",
		},
		{
			kacName:             defaultKACName,
			kacNamespace:        "",
			addToRuntimeObjects: true,
			client:              true,
			expectedErrorText:   "klusterletAddonConfig 'nsname' cannot be empty",
		},
		{
			kacName:             defaultKACName,
			kacNamespace:        defaultKACNamespace,
			addToRuntimeObjects: false,
			client:              true,
			expectedErrorText: fmt.Sprintf(
				"klusterletAddonConfig object %s does not exist in namespace %s", defaultKACName, defaultKACNamespace),
		},
		{
			kacName:             defaultKACName,
			kacNamespace:        defaultKACNamespace,
			addToRuntimeObjects: true,
			client:              false,
			expectedErrorText:   "klusterletAddonConfig 'apiClient' cannot be nil",
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testKAC := buildDummyKAC(testCase.kacName, testCase.kacNamespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testKAC)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})
		}

		kacBuilder, err := PullKAC(testSettings, testCase.kacName, testCase.kacNamespace)

		if testCase.expectedErrorText == "" {
			assert.Nil(t, err)
			assert.Equal(t, testKAC.Name, kacBuilder.Definition.Name)
			assert.Equal(t, testKAC.Namespace, kacBuilder.Definition.Namespace)
		} else {
			assert.EqualError(t, err, testCase.expectedErrorText)
		}
	}
}

func TestKACExists(t *testing.T) {
	testCases := []struct {
		testBuilder *KACBuilder
		exists      bool
	}{
		{
			testBuilder: buildValidKACTestBuilder(buildTestClientWithDummyKAC()),
			exists:      true,
		},
		{
			testBuilder: buildInvalidKACTestBuilder(buildTestClientWithDummyKAC()),
			exists:      false,
		},
		{
			testBuilder: buildValidKACTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

func TestKACCreate(t *testing.T) {
	testCases := []struct {
		testBuilder   *KACBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidKACTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
		{
			testBuilder:   buildValidKACTestBuilder(buildTestClientWithDummyKAC()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidKACTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("klusterletAddonConfig 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		kacBuilder, err := testCase.testBuilder.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, kacBuilder.Definition.Name, kacBuilder.Object.Name)
			assert.Equal(t, kacBuilder.Definition.Namespace, kacBuilder.Object.Namespace)
		}
	}
}

func TestKACUpdate(t *testing.T) {
	testCases := []struct {
		alreadyExists     bool
		force             bool
		expectedErrorText string
	}{
		{
			alreadyExists:     false,
			force:             false,
			expectedErrorText: "cannot update non-existent klusterletAddonConfig",
		},
		{
			alreadyExists:     true,
			force:             false,
			expectedErrorText: "",
		},
		{
			alreadyExists:     false,
			force:             true,
			expectedErrorText: "cannot update non-existent klusterletAddonConfig",
		},
		{
			alreadyExists:     true,
			force:             true,
			expectedErrorText: "",
		},
	}

	for _, testCase := range testCases {
		kacBuilder := buildValidKACTestBuilder(clients.GetTestClients(clients.TestClientParams{}))

		// Create the builder rather than just adding it to the client so that the proper metadata is added and
		// the update will not fail.
		if testCase.alreadyExists {
			var err error
			kacBuilder, err = kacBuilder.Create()

			assert.Nil(t, err)
			assert.True(t, kacBuilder.Exists())
		}

		assert.NotNil(t, kacBuilder.Definition)
		assert.False(t, kacBuilder.Definition.Spec.SearchCollectorConfig.Enabled)

		kacBuilder.Definition.Spec.SearchCollectorConfig.Enabled = true

		kacBuilder, err := kacBuilder.Update(testCase.force)

		if testCase.expectedErrorText == "" {
			assert.Nil(t, err)
			assert.True(t, kacBuilder.Object.Spec.SearchCollectorConfig.Enabled)
		} else {
			assert.EqualError(t, err, testCase.expectedErrorText)
		}
	}
}

func TestKACDelete(t *testing.T) {
	testCases := []struct {
		testBuilder       *KACBuilder
		expectedErrorText string
	}{
		{
			testBuilder:       buildValidKACTestBuilder(buildTestClientWithDummyKAC()),
			expectedErrorText: "",
		},
		{
			testBuilder:       buildValidKACTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedErrorText: "",
		},
		{
			testBuilder:       buildInvalidKACTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedErrorText: "klusterletAddonConfig 'nsname' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		err := testCase.testBuilder.Delete()

		if testCase.expectedErrorText == "" {
			assert.Nil(t, err)
			assert.Nil(t, testCase.testBuilder.Object)
		} else {
			assert.EqualError(t, err, testCase.expectedErrorText)
		}
	}
}

func TestKACValidate(t *testing.T) {
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
			expectedError:   fmt.Errorf("error: received nil klusterletAddonConfig builder"),
		},
		{
			builderNil:      false,
			definitionNil:   true,
			apiClientNil:    false,
			builderErrorMsg: "",
			expectedError:   fmt.Errorf("can not redefine the undefined klusterletAddonConfig"),
		},
		{
			builderNil:      false,
			definitionNil:   false,
			apiClientNil:    true,
			builderErrorMsg: "",
			expectedError:   fmt.Errorf("klusterletAddonConfig builder cannot have nil apiClient"),
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
		kacBuilder := buildValidKACTestBuilder(clients.GetTestClients(clients.TestClientParams{}))

		if testCase.builderNil {
			kacBuilder = nil
		}

		if testCase.definitionNil {
			kacBuilder.Definition = nil
		}

		if testCase.apiClientNil {
			kacBuilder.apiClient = nil
		}

		if testCase.builderErrorMsg != "" {
			kacBuilder.errorMsg = testCase.builderErrorMsg
		}

		valid, err := kacBuilder.validate()

		if testCase.expectedError != nil {
			assert.False(t, valid)
			assert.Equal(t, testCase.expectedError, err)
		} else {
			assert.True(t, valid)
			assert.Nil(t, err)
		}
	}
}

// buildDummyKAC returns a KlusterletAddonConfig with the provided name and namespace.
func buildDummyKAC(name, namespace string) *kacv1.KlusterletAddonConfig {
	return &kacv1.KlusterletAddonConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}

// buildTestClientWithDummyKAC returns a client with a mock KlusterletAddonConfig.
func buildTestClientWithDummyKAC() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyKAC(defaultKACName, defaultKACNamespace),
		},
	})
}

// buildValidKACTestBuilder returns a valid Builder for testing.
func buildValidKACTestBuilder(apiClient *clients.Settings) *KACBuilder {
	return NewKACBuilder(apiClient, defaultKACName, defaultKACNamespace)
}

// buildInvalidKACTestBuilder returns an invalid Builder for testing.
func buildInvalidKACTestBuilder(apiClient *clients.Settings) *KACBuilder {
	return NewKACBuilder(apiClient, defaultKACName, "")
}
