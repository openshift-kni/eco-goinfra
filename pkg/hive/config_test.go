package hive

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	hiveV1 "github.com/openshift-kni/eco-goinfra/pkg/schemes/hive/api/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultHiveConfigName = "hiveconfig"
)

func TestNewConfigBuilder(t *testing.T) {
	generateConfig := NewConfigBuilder

	testCases := []struct {
		name          string
		expectedError string
	}{
		{
			name:          "hiveconfig",
			expectedError: "",
		},
		{
			name:          "",
			expectedError: "hiveconfig 'name' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})
		testConfig := generateConfig(testSettings, testCase.name)
		assert.Equal(t, testCase.expectedError, testConfig.errorMsg)
		assert.NotNil(t, testConfig.Definition)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.name, testConfig.Definition.Name)
		}
	}
}

func TestPullConfig(t *testing.T) {
	generateConfig := func(name string) *hiveV1.HiveConfig {
		return &hiveV1.HiveConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: hiveV1.HiveConfigSpec{},
		}
	}

	testCases := []struct {
		name                string
		addToRuntimeObjects bool
		expectedError       error
		client              bool
	}{
		{
			name:                "hiveconfig",
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("hiveconfig 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                "hiveconfig",
			addToRuntimeObjects: false,
			expectedError:       fmt.Errorf("hiveconfig object hiveconfig does not exist"),
			client:              true,
		},
		{
			name:                "hiveconfig",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("hiveconfig 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testHiveConfig := generateConfig(testCase.name)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testHiveConfig)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(
				clients.TestClientParams{K8sMockObjects: runtimeObjects, SchemeAttachers: testSchemes})
		}

		builderResult, err := PullConfig(testSettings, testCase.name)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, builderResult.Object.Name)
		}
	}
}

func TestConfigGet(t *testing.T) {
	testCases := []struct {
		testConfig    *ConfigBuilder
		expectedError error
	}{
		{
			testConfig:    buildValidConfigBuilder(buildHiveConfigTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testConfig:    buildInValidConfigBuilder(buildHiveConfigTestClientWithDummyObject()),
			expectedError: fmt.Errorf("hiveconfig 'name' cannot be empty"),
		},
		{
			testConfig:    buildValidConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("hiveconfigs.hive.openshift.io \"hiveconfig\" not found"),
		},
	}

	for _, testCase := range testCases {
		hiveConfig, err := testCase.testConfig.Get()
		if testCase.expectedError != nil {
			assert.Equal(t, err.Error(), testCase.expectedError.Error())
		}

		if testCase.expectedError == nil {
			assert.Equal(t, hiveConfig.Name, testCase.testConfig.Definition.Name)
		}
	}
}

func TestConfigExists(t *testing.T) {
	testCases := []struct {
		testConfig     *ConfigBuilder
		expectedStatus bool
	}{
		{
			testConfig:     buildValidConfigBuilder(buildHiveConfigTestClientWithDummyObject()),
			expectedStatus: true,
		},
		{
			testConfig:     buildInValidConfigBuilder(buildHiveConfigTestClientWithDummyObject()),
			expectedStatus: false,
		},
		{
			testConfig:     buildValidConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testConfig.Exists()
		assert.Equal(t, exist, testCase.expectedStatus)
	}
}

func TestConfigDelete(t *testing.T) {
	testCases := []struct {
		testConfig    *ConfigBuilder
		expectedError error
	}{
		{
			testConfig:    buildValidConfigBuilder(buildHiveConfigTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testConfig:    buildInValidConfigBuilder(buildHiveConfigTestClientWithDummyObject()),
			expectedError: fmt.Errorf("hiveconfig 'name' cannot be empty"),
		},
		{
			testConfig:    buildValidConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		err := testCase.testConfig.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testConfig.Object)
			assert.Nil(t, err)
		}
	}
}

func TestConfigUpdate(t *testing.T) {
	testCases := []struct {
		testConfig    *ConfigBuilder
		expectedError error
	}{
		{
			testConfig:    buildValidConfigBuilder(buildHiveConfigTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testConfig:    buildInValidConfigBuilder(buildHiveConfigTestClientWithDummyObject()),
			expectedError: fmt.Errorf("hiveconfig 'name' cannot be empty"),
		},
		{
			testConfig:    buildValidConfigBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("hiveconfigs.hive.openshift.io \"hiveconfig\" not found"),
		},
	}

	for _, testCase := range testCases {
		assert.Empty(t, testCase.testConfig.Definition.Spec.LogLevel)

		assert.Nil(t, nil, testCase.testConfig.Object)
		testCase.testConfig.Definition.Spec.LogLevel = "100"
		configBuilder, err := testCase.testConfig.Update()

		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		} else {
			assert.Nil(t, err)
		}

		if testCase.expectedError == nil {
			assert.NotNil(t, configBuilder.Object)
			assert.Equal(t, testCase.testConfig.Definition.Spec.LogLevel, configBuilder.Object.Spec.LogLevel)
		}
	}
}

func TestConfigWithOptions(t *testing.T) {
	testSettings := buildHiveConfigTestClientWithDummyObject()
	testBuilder := buildValidConfigBuilder(testSettings).WithOptions(
		func(builder *ConfigBuilder) (*ConfigBuilder, error) {
			return builder, nil
		})

	assert.Equal(t, "", testBuilder.errorMsg)
	testBuilder = buildValidConfigBuilder(testSettings).WithOptions(
		func(builder *ConfigBuilder) (*ConfigBuilder, error) {
			return builder, fmt.Errorf("error")
		})
	assert.Equal(t, "error", testBuilder.errorMsg)
}

func buildValidConfigBuilder(apiClient *clients.Settings) *ConfigBuilder {
	builder := NewConfigBuilder(apiClient, defaultHiveConfigName)
	builder.Definition.ResourceVersion = "999"

	return builder
}

func buildInValidConfigBuilder(apiClient *clients.Settings) *ConfigBuilder {
	return NewConfigBuilder(apiClient, "")
}

func buildHiveConfigTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyHiveConfig(),
		SchemeAttachers: testSchemes,
	})
}

func buildDummyHiveConfig() []runtime.Object {
	return append([]runtime.Object{}, &hiveV1.HiveConfig{
		ObjectMeta: metav1.ObjectMeta{
			ResourceVersion: "999",
			Name:            defaultHiveConfigName,
		},
		Spec: hiveV1.HiveConfigSpec{},
	})
}
