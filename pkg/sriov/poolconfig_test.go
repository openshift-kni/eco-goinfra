package sriov

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"

	srIovV1 "github.com/k8snetworkplumbingwg/sriov-network-operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	defaultPoolConfigName   = "poolconfig"
	defaultPoolConfigNsName = "testnamespace"
)

func TestNewPoolConfigBuilder(t *testing.T) {
	testCases := []struct {
		poolConfigName      string
		poolConfigNamespace string
		expectedErrorText   string
		client              bool
	}{
		{
			poolConfigName:      defaultPoolConfigName,
			poolConfigNamespace: defaultPoolConfigNsName,
			client:              true,
		},
		{
			poolConfigName:      "",
			poolConfigNamespace: defaultPoolConfigNsName,
			expectedErrorText:   "SriovNetworkPoolConfig 'name' cannot be empty",
			client:              true,
		},
		{
			poolConfigName:      defaultPoolConfigName,
			poolConfigNamespace: "",
			expectedErrorText:   "SriovNetworkPoolConfig 'nsname' cannot be empty",
			client:              true,
		},
	}
	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})
		testPoolconfigStructure := NewPoolConfigBuilder(
			testSettings, testCase.poolConfigName, testCase.poolConfigNamespace)
		assert.NotNil(t, testPoolconfigStructure)

		if len(testCase.expectedErrorText) > 0 {
			assert.Equal(t, testPoolconfigStructure.errorMsg, testCase.expectedErrorText)
		}
	}
}

func TestPooConfigCreate(t *testing.T) {
	testCases := []struct {
		testPoolConfig *PoolConfigBuilder
		expectedError  error
	}{
		{
			testPoolConfig: buildValidPoolConfigTestBuilder(buildTestPoolConfigClientWithDummyObject()),
			expectedError:  nil,
		},
		{
			testPoolConfig: buildInvalidPoolConfigTestBuilder(buildTestPoolConfigClientWithDummyObject()),
			expectedError:  fmt.Errorf("SriovNetworkPoolConfig 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		testBuilder, err := testCase.testPoolConfig.Create()
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil {
			assert.Equal(t, testBuilder.Definition, testBuilder.Object)
		}
	}
}

func TestPooConfigDelete(t *testing.T) {
	testCases := []struct {
		testPoolConfig *PoolConfigBuilder
		expectedError  error
	}{
		{
			testPoolConfig: buildValidPoolConfigTestBuilder(buildTestPoolConfigClientWithDummyObject()),
			expectedError:  nil,
		},
		{
			testPoolConfig: buildInvalidPoolConfigTestBuilder(buildTestPoolConfigClientWithDummyObject()),
			expectedError:  fmt.Errorf("SriovNetworkPoolConfig 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		err := testCase.testPoolConfig.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testPoolConfig.Object)
		}
	}
}

func TestPooConfigExist(t *testing.T) {
	testCases := []struct {
		testPoolConfig *PoolConfigBuilder
		expectedStatus bool
	}{
		{
			testPoolConfig: buildValidPoolConfigTestBuilder(buildTestPoolConfigClientWithDummyObject()),
			expectedStatus: true,
		},
		{
			testPoolConfig: buildInvalidPoolConfigTestBuilder(buildTestPoolConfigClientWithDummyObject()),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		_, _ = testCase.testPoolConfig.Create()
		assert.Equal(t, testCase.expectedStatus, testCase.testPoolConfig.Exists())
	}
}

func TestTestPooConfigGet(t *testing.T) {
	testCases := []struct {
		testPoolConfig *PoolConfigBuilder
		expectedError  error
	}{
		{
			testPoolConfig: buildValidPoolConfigTestBuilder(buildTestPoolConfigClientWithDummyObject()),
			expectedError:  nil,
		},
		{
			testPoolConfig: buildInvalidPoolConfigTestBuilder(buildTestPoolConfigClientWithDummyObject()),
			expectedError:  fmt.Errorf("SriovNetworkPoolConfig 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		_, _ = testCase.testPoolConfig.Create()
		testBuilder, err := testCase.testPoolConfig.Get()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.NotNil(t, testBuilder)
		}
	}
}

func TestPoolConfigUpdate(t *testing.T) {
	testCases := []struct {
		testPoolConfig *PoolConfigBuilder
		expectedError  error
	}{
		{
			testPoolConfig: buildValidPoolConfigTestBuilder(buildTestPoolConfigClientWithDummyObject()),
			expectedError:  nil,
		},
	}
	for _, testCase := range testCases {
		poolConfigBuilder, err := testCase.testPoolConfig.WithMaxUnavailable(intstr.FromInt32(2)).Create()
		assert.Nil(t, err)
		assert.Equal(t, int32(2), poolConfigBuilder.Definition.Spec.MaxUnavailable.IntVal)
		testCase.testPoolConfig.WithMaxUnavailable(intstr.FromString("100%"))

		poolConfigBuilder.Definition.ObjectMeta.ResourceVersion = "999"
		poolConfigBuilder, err = poolConfigBuilder.Update()
		assert.Nil(t, err)
		assert.Equal(t, "100%", poolConfigBuilder.Object.Spec.MaxUnavailable.StrVal)
		assert.Equal(t, poolConfigBuilder.Definition, poolConfigBuilder.Object)
	}
}

func TestWithNodeSelector(t *testing.T) {
	testCases := []struct {
		nodeSelector      map[string]string
		expectedErrorText string
	}{
		{
			nodeSelector:      map[string]string{"test": "test"},
			expectedErrorText: "",
		},
		{
			nodeSelector:      map[string]string{},
			expectedErrorText: "SriovNetworkPoolConfig 'nodeSelector' cannot be empty map",
		},
	}

	for _, testCase := range testCases {
		testSettings := buildTestPoolConfigClientWithDummyObject()
		testBuilder := buildValidPoolConfigTestBuilder(testSettings).WithNodeSelector(testCase.nodeSelector)
		assert.Equal(t, testBuilder.errorMsg, testCase.expectedErrorText)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, testBuilder.Definition.Spec.NodeSelector.MatchLabels, testCase.nodeSelector)
		}
	}
}

func TestWithMaxUnavailable(t *testing.T) {
	testCases := []struct {
		maxUnavailable    intstr.IntOrString
		expectedErrorText string
	}{
		{
			maxUnavailable:    intstr.FromInt32(3),
			expectedErrorText: "",
		},
		{
			maxUnavailable:    intstr.FromString("wrongValue"),
			expectedErrorText: "invalid type: strings needs to be a percentage: {1 0 wrongValue}",
		},
		{
			maxUnavailable:    intstr.FromString("99s%"),
			expectedErrorText: "invalid value \"99s%\": strconv.Atoi: parsing \"99s\": invalid syntax",
		},
		{
			maxUnavailable:    intstr.FromString("101%"),
			expectedErrorText: "invalid value: percentage needs to be between 1 and 100: {1 0 101%}",
		},
		{
			maxUnavailable:    intstr.FromInt32(-2),
			expectedErrorText: "negative number is not allowed: {0 -2 }",
		},
	}

	for _, testCase := range testCases {
		testSettings := buildTestPoolConfigClientWithDummyObject()
		testBuilder := buildValidPoolConfigTestBuilder(testSettings).WithMaxUnavailable(testCase.maxUnavailable)
		assert.Equal(t, testBuilder.errorMsg, testCase.expectedErrorText)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, testBuilder.Definition.Spec.MaxUnavailable.IntVal, testCase.maxUnavailable.IntVal)
		}
	}
}

func TestPullPoolConfig(t *testing.T) {
	testCases := []struct {
		poolConfigName      string
		poolConfigNamespace string
		expectedErrorText   string
		expectedError       bool
		client              bool
	}{
		{
			poolConfigName:      defaultPoolConfigName,
			poolConfigNamespace: defaultPoolConfigNsName,
			expectedError:       false,
			client:              true,
		},
		{
			poolConfigName:      "",
			poolConfigNamespace: defaultPoolConfigNsName,
			expectedErrorText:   "SriovNetworkPoolConfig 'name' cannot be empty",
			expectedError:       true,
			client:              true,
		},
		{
			poolConfigName:      defaultPoolConfigName,
			poolConfigNamespace: "",
			expectedErrorText:   "SriovNetworkPoolConfig 'namespace' cannot be empty",
			expectedError:       true,
			client:              true,
		},
		{
			poolConfigName:      defaultPoolConfigName,
			poolConfigNamespace: defaultPoolConfigNsName,
			expectedErrorText:   "SriovNetworkPoolConfig 'apiClient' cannot be empty",
			expectedError:       true,
			client:              false,
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  buildDummyPoolConfigObject(),
				SchemeAttachers: testSchemes,
			})
			_, _ = buildValidPoolConfigTestBuilder(testSettings).Create()
		}

		testPoolConfig, err := PullPoolConfig(testSettings, testCase.poolConfigName, testCase.poolConfigNamespace)

		if testCase.expectedError {
			assert.NotNil(t, err)

			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testCase.poolConfigName, testPoolConfig.Object.Name)
			assert.Equal(t, testCase.poolConfigNamespace, testPoolConfig.Object.Namespace)
		}
	}
}

// buildValidPoolConfigTestBuilder returns a valid Builder for testing purposes.
func buildValidPoolConfigTestBuilder(apiClient *clients.Settings) *PoolConfigBuilder {
	return NewPoolConfigBuilder(apiClient, defaultPoolConfigName, defaultPoolConfigNsName)
}

// buildInvalidPoolConfigTestBuilder returns an invalid Builder for testing purposes.
func buildInvalidPoolConfigTestBuilder(apiClient *clients.Settings) *PoolConfigBuilder {
	return NewPoolConfigBuilder(apiClient, defaultPoolConfigName, "")
}

func buildTestPoolConfigClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyPoolConfigObject(),
		SchemeAttachers: testSchemes,
	})
}

func buildDummyPoolConfigObject() []runtime.Object {
	return append([]runtime.Object{}, &srIovV1.SriovNetworkPoolConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultPoolConfigName,
			Namespace: defaultPoolConfigNsName,
		},
		Spec: srIovV1.SriovNetworkPoolConfigSpec{},
	})
}
