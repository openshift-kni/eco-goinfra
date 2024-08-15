package cgu

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/cluster-group-upgrades-operator/pkg/api/clustergroupupgrades/v1alpha1"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultPreCachingConfigName   = "precachingconfig-test"
	defaultPreCachingConfigNsName = "test-ns"
)

func TestNewPreCachingConfigBuilder(t *testing.T) {
	testCases := []struct {
		preCachingConfigName      string
		preCachingConfigNamespace string
		expectedErrorText         string
	}{
		{
			preCachingConfigName:      defaultPreCachingConfigName,
			preCachingConfigNamespace: defaultPreCachingConfigNsName,
			expectedErrorText:         "",
		},
		{
			preCachingConfigName:      "",
			preCachingConfigNamespace: defaultPreCachingConfigNsName,
			expectedErrorText:         "preCachingConfig 'name' cannot be empty",
		},
		{
			preCachingConfigName:      defaultPreCachingConfigName,
			preCachingConfigNamespace: "",
			expectedErrorText:         "preCachingConfig 'nsname' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})
		preCachingConfigBuilder := NewPreCachingConfigBuilder(
			testSettings, testCase.preCachingConfigName, testCase.preCachingConfigNamespace)
		assert.NotNil(t, preCachingConfigBuilder)
		assert.Equal(t, testCase.expectedErrorText, preCachingConfigBuilder.errorMsg)
	}
}

func TestPullPreCachingConfig(t *testing.T) {
	testCases := []struct {
		preCachingConfigName      string
		preCachingConfigNamespace string
		addToRuntimeObjects       bool
		client                    bool
		expectedErrorText         string
	}{
		{
			preCachingConfigName:      defaultPreCachingConfigName,
			preCachingConfigNamespace: defaultPreCachingConfigNsName,
			addToRuntimeObjects:       true,
			client:                    true,
			expectedErrorText:         "",
		},
		{
			preCachingConfigName:      defaultPreCachingConfigName,
			preCachingConfigNamespace: defaultPreCachingConfigNsName,
			addToRuntimeObjects:       false,
			client:                    true,
			expectedErrorText: fmt.Sprintf(
				"preCachingConfig object %s does not exist in namespace %s",
				defaultPreCachingConfigName, defaultPreCachingConfigNsName),
		},
		{
			preCachingConfigName:      "",
			preCachingConfigNamespace: defaultPreCachingConfigNsName,
			addToRuntimeObjects:       false,
			client:                    true,
			expectedErrorText:         "preCachingConfig 'name' cannot be empty",
		},
		{
			preCachingConfigName:      defaultPreCachingConfigName,
			preCachingConfigNamespace: "",
			addToRuntimeObjects:       false,
			client:                    true,
			expectedErrorText:         "preCachingConfig 'nsname' cannot be empty",
		},
		{
			preCachingConfigName:      defaultPreCachingConfigName,
			preCachingConfigNamespace: defaultPreCachingConfigNsName,
			addToRuntimeObjects:       false,
			client:                    false,
			expectedErrorText:         "preCachingConfig 'apiClient' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testPreCachingConfig := buildDummyPreCachingConfig(testCase.preCachingConfigName, testCase.preCachingConfigNamespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testPreCachingConfig)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		preCachingConfigBuilder, err := PullPreCachingConfig(
			testSettings, testPreCachingConfig.Name, testPreCachingConfig.Namespace)

		if testCase.expectedErrorText != "" {
			assert.NotNil(t, err)
			assert.Equal(t, testCase.expectedErrorText, err.Error())
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testPreCachingConfig.Name, preCachingConfigBuilder.Object.Name)
			assert.Equal(t, testPreCachingConfig.Namespace, preCachingConfigBuilder.Object.Namespace)
		}
	}
}

func TestPreCachingConfigExists(t *testing.T) {
	testCases := []struct {
		testBuilder *PreCachingConfigBuilder
		exists      bool
	}{
		{
			testBuilder: buildValidPreCachingConfigTestBuilder(buildTestClientWithDummyPreCachingConfig()),
			exists:      true,
		},
		{
			testBuilder: buildInvalidPreCachingConfigTestBuilder(buildTestClientWithDummyPreCachingConfig()),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

func TestPreCachingConfigGet(t *testing.T) {
	testCases := []struct {
		testBuilder              *PreCachingConfigBuilder
		expectedPreCachingConfig *v1alpha1.PreCachingConfig
	}{
		{
			testBuilder:              buildValidPreCachingConfigTestBuilder(buildTestClientWithDummyPreCachingConfig()),
			expectedPreCachingConfig: buildDummyPreCachingConfig(defaultPreCachingConfigName, defaultPreCachingConfigNsName),
		},
		{
			testBuilder:              buildValidPreCachingConfigTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedPreCachingConfig: nil,
		},
	}

	for _, testCase := range testCases {
		preCachingConfig, err := testCase.testBuilder.Get()

		if testCase.expectedPreCachingConfig == nil {
			assert.Nil(t, preCachingConfig)
			assert.True(t, k8serrors.IsNotFound(err))
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testCase.expectedPreCachingConfig.Name, preCachingConfig.Name)
			assert.Equal(t, testCase.expectedPreCachingConfig.Namespace, preCachingConfig.Namespace)
		}
	}
}

func TestPreCachingConfigCreate(t *testing.T) {
	testCases := []struct {
		testBuilder   *PreCachingConfigBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidPreCachingConfigTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidPreCachingConfigTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("preCachingConfig 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		preCachingConfigBuilder, err := testCase.testBuilder.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, preCachingConfigBuilder.Definition, preCachingConfigBuilder.Object)
		}
	}
}

func TestPreCachingConfigDelete(t *testing.T) {
	testCases := []struct {
		testBuilder   *PreCachingConfigBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidPreCachingConfigTestBuilder(buildTestClientWithDummyPreCachingConfig()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidPreCachingConfigTestBuilder(buildTestClientWithDummyPreCachingConfig()),
			expectedError: fmt.Errorf("preCachingConfig 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		err := testCase.testBuilder.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testBuilder.Object)
		}
	}
}

func TestPreCachingConfigUpdate(t *testing.T) {
	testCases := []struct {
		alreadyExists bool
		force         bool
	}{
		{
			alreadyExists: false,
			force:         false,
		},
		{
			alreadyExists: true,
			force:         false,
		},
		{
			alreadyExists: false,
			force:         true,
		},
		{
			alreadyExists: true,
			force:         true,
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidPreCachingConfigTestBuilder(clients.GetTestClients(clients.TestClientParams{}))

		// Create the builder rather than just adding it to the client so that the proper metadata is added and
		// the update will not fail.
		if testCase.alreadyExists {
			var err error

			testBuilder = buildValidPreCachingConfigTestBuilder(clients.GetTestClients(clients.TestClientParams{}))
			testBuilder, err = testBuilder.Create()
			assert.Nil(t, err)
		}

		assert.NotNil(t, testBuilder.Definition)
		assert.Empty(t, testBuilder.Definition.Spec.SpaceRequired)

		testBuilder.Definition.Spec.SpaceRequired = "10 GiB"

		preCachingConfigBuilder, err := testBuilder.Update(testCase.force)
		assert.NotNil(t, testBuilder.Definition)

		if testCase.alreadyExists || testCase.force {
			assert.Nil(t, err)
			assert.Equal(t, testBuilder.Definition.Name, preCachingConfigBuilder.Definition.Name)
			assert.Equal(t, testBuilder.Definition.Spec.SpaceRequired, preCachingConfigBuilder.Definition.Spec.SpaceRequired)
		} else {
			assert.NotNil(t, err)
		}
	}
}

func TestPreCachingConfigValidate(t *testing.T) {
	testCases := []struct {
		builderNil    bool
		definitionNil bool
		apiClientNil  bool
		expectedError error
		builderErrMsg string
	}{
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: nil,
			builderErrMsg: "",
		},
		{
			builderNil:    true,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: fmt.Errorf("error received nil preCachingConfig builder"),
			builderErrMsg: "",
		},
		{
			builderNil:    false,
			definitionNil: true,
			apiClientNil:  false,
			expectedError: fmt.Errorf("can not redefine the undefined preCachingConfig"),
			builderErrMsg: "",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  true,
			expectedError: fmt.Errorf("preCachingConfig builder cannot have nil apiClient"),
			builderErrMsg: "",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  false,
			builderErrMsg: "test error",
			expectedError: fmt.Errorf("test error"),
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidPreCachingConfigTestBuilder(clients.GetTestClients(clients.TestClientParams{}))

		if testCase.builderNil {
			testBuilder = nil
		}

		if testCase.definitionNil {
			testBuilder.Definition = nil
		}

		if testCase.apiClientNil {
			testBuilder.apiClient = nil
		}

		if testCase.builderErrMsg != "" {
			testBuilder.errorMsg = testCase.builderErrMsg
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

// buildDummyPreCachingConfig returns a PreCachingConfig with the provided name and namespace.
func buildDummyPreCachingConfig(name, nsname string) *v1alpha1.PreCachingConfig {
	return &v1alpha1.PreCachingConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsname,
		},
	}
}

// buildTestClientWithDummyPreCachingConfig returns a client with a mock dummy PreCachingConfig.
func buildTestClientWithDummyPreCachingConfig() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyPreCachingConfig(defaultPreCachingConfigName, defaultPreCachingConfigNsName),
		},
		SchemeAttachers: testSchemes,
	})
}

// buildValidPreCachingConfigTestBuilder returns a valid PreCachingConfigBuilder for testing.
func buildValidPreCachingConfigTestBuilder(apiClient *clients.Settings) *PreCachingConfigBuilder {
	return NewPreCachingConfigBuilder(apiClient, defaultPreCachingConfigName, defaultPreCachingConfigNsName)
}

// buildInvalidPreCachingConfigTestBuilder returns an invalid PreCachingConfigBuilder for testing.
func buildInvalidPreCachingConfigTestBuilder(apiClient *clients.Settings) *PreCachingConfigBuilder {
	return NewPreCachingConfigBuilder(apiClient, defaultPreCachingConfigName, "")
}
