package ocm

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	operatorv1 "open-cluster-management.io/api/operator/v1"
)

var operatorTestSchemes = []clients.SchemeAttacher{
	operatorv1.Install,
}

func TestNewKlusterletBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		client        bool
		expectedError string
	}{
		{
			name:          KlusterletName,
			client:        true,
			expectedError: "",
		},
		{
			name:          "",
			client:        true,
			expectedError: "klusterlet 'name' cannot be empty",
		},
		{
			name:          KlusterletName,
			client:        false,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		testBuilder := NewKlusterletBuilder(testSettings, testCase.name)

		if testCase.client {
			assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

			if testCase.expectedError == "" {
				assert.Equal(t, testCase.name, testBuilder.Definition.Name)
			}
		} else {
			assert.Nil(t, testBuilder)
		}
	}
}

func TestPullKlusterlet(t *testing.T) {
	testCases := []struct {
		name                string
		addToRuntimeObjects bool
		client              bool
		expectedError       error
	}{
		{
			name:                KlusterletName,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       nil,
		},
		{
			name:                "",
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("klusterlet 'name' cannot be empty"),
		},
		{
			name:                KlusterletName,
			addToRuntimeObjects: false,
			client:              true,
			expectedError:       fmt.Errorf("klusterlet object %s does not exist", KlusterletName),
		},
		{
			name:                KlusterletName,
			addToRuntimeObjects: true,
			client:              false,
			expectedError:       fmt.Errorf("klusterlet 'apiClient' cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, buildDummyKlusterlet(KlusterletName))
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: operatorTestSchemes,
			})
		}

		testBuilder, err := PullKlusterlet(testSettings, testCase.name)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, testBuilder.Definition.Name)
		}
	}
}

func TestKlusterletGet(t *testing.T) {
	testCases := []struct {
		testBuilder   *KlusterletBuilder
		expectedError string
	}{
		{
			testBuilder:   buildValidKlusterletTestBuilder(buildTestClientWithDummyKlusterlet()),
			expectedError: "",
		},
		{
			testBuilder:   buildInvalidKlusterlettestbuilder(buildTestClientWithDummyKlusterlet()),
			expectedError: "klusterlet 'name' cannot be empty",
		},
		{
			testBuilder:   buildValidKlusterletTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Sprintf("klusterlets.operator.open-cluster-management.io \"%s\" not found", KlusterletName),
		},
	}

	for _, testCase := range testCases {
		klusterlet, err := testCase.testBuilder.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, testCase.testBuilder.Definition.Name, klusterlet.Name)
		} else {
			assert.EqualError(t, err, testCase.expectedError)
		}
	}
}

func TestKlusterletExists(t *testing.T) {
	testCases := []struct {
		testBuilder *KlusterletBuilder
		exists      bool
	}{
		{
			testBuilder: buildValidKlusterletTestBuilder(buildTestClientWithDummyKlusterlet()),
			exists:      true,
		},
		{
			testBuilder: buildInvalidKlusterlettestbuilder(buildTestClientWithDummyKlusterlet()),
			exists:      false,
		},
		{
			testBuilder: buildValidKlusterletTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

func TestKlusterletCreate(t *testing.T) {
	testCases := []struct {
		testBuilder   *KlusterletBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidKlusterletTestBuilder(buildTestClientWithDummyKlusterlet()),
			expectedError: nil,
		},
		{
			testBuilder:   buildValidKlusterletTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidKlusterlettestbuilder(buildTestClientWithDummyKlusterlet()),
			expectedError: fmt.Errorf("klusterlet 'name' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		testBuilder, err := testCase.testBuilder.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testBuilder.Definition.Name, testBuilder.Object.Name)
		}
	}
}

func TestKlusterletUpdate(t *testing.T) {
	testCases := []struct {
		testBuilder   *KlusterletBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidKlusterletTestBuilder(buildTestClientWithDummyKlusterlet()),
			expectedError: nil,
		},
		{
			testBuilder:   buildValidKlusterletTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("cannot update non-existent klusterlet"),
		},
		{
			testBuilder:   buildInvalidKlusterlettestbuilder(buildTestClientWithDummyKlusterlet()),
			expectedError: fmt.Errorf("klusterlet 'name' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		assert.Empty(t, testCase.testBuilder.Definition.Spec.ClusterName)

		testCase.testBuilder.Definition.Spec.ClusterName = "test"
		testCase.testBuilder.Definition.ResourceVersion = "999"

		testBuilder, err := testCase.testBuilder.Update()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, "test", testBuilder.Object.Spec.ClusterName)
		}
	}
}

func TestKlusterletDelete(t *testing.T) {
	testCases := []struct {
		testBuilder   *KlusterletBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidKlusterletTestBuilder(buildTestClientWithDummyKlusterlet()),
			expectedError: nil,
		},
		{
			testBuilder:   buildValidKlusterletTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidKlusterlettestbuilder(buildTestClientWithDummyKlusterlet()),
			expectedError: fmt.Errorf("klusterlet 'name' cannot be empty"),
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

func TestKlusterletValidate(t *testing.T) {
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
			expectedError:   fmt.Errorf("error: received nil klusterlet builder"),
		},
		{
			builderNil:      false,
			definitionNil:   true,
			apiClientNil:    false,
			builderErrorMsg: "",
			expectedError:   fmt.Errorf("can not redefine the undefined klusterlet"),
		},
		{
			builderNil:      false,
			definitionNil:   false,
			apiClientNil:    true,
			builderErrorMsg: "",
			expectedError:   fmt.Errorf("klusterlet builder cannot have nil apiClient"),
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
		testBuilder := buildValidKlusterletTestBuilder(clients.GetTestClients(clients.TestClientParams{}))

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
		assert.Equal(t, testCase.expectedError, err)
		assert.Equal(t, testCase.expectedError == nil, valid)
	}
}

// buildDummyKlusterlet returns a Klusterlet with the provided name.
func buildDummyKlusterlet(name string) *operatorv1.Klusterlet {
	return &operatorv1.Klusterlet{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

// buildTestClientWithDummyKlusterlet returns a test client with a Klusterlet using KlusterletName.
func buildTestClientWithDummyKlusterlet() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyKlusterlet(KlusterletName),
		},
		SchemeAttachers: operatorTestSchemes,
	})
}

// buildValidKlusterletTestBuilder returns a valid KlusterletBuilder using KlusterletName.
func buildValidKlusterletTestBuilder(apiClient *clients.Settings) *KlusterletBuilder {
	return NewKlusterletBuilder(apiClient, KlusterletName)
}

// buildInvalidKlusterlettestbuilder returns an invalid KlusterletBuilder with an empty name.
func buildInvalidKlusterlettestbuilder(apiClient *clients.Settings) *KlusterletBuilder {
	return NewKlusterletBuilder(apiClient, "")
}
