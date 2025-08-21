package nfd

import (
	"fmt"

	"testing"

	"k8s.io/apimachinery/pkg/runtime"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	nfdv1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/nfd/v1alpha1"
	"github.com/stretchr/testify/assert"
)

var (
	nodeFeatureRuleExampleName = "test-node-feature-rule"
	nodeFeatureRuleNamespace   = "test-namespace"
	nodeFeatureRuleAlmExample  = fmt.Sprintf(`[
	{
		"apiVersion": "nfd.openshift.io/v1alpha1",
		"kind": "NodeFeatureRule",
		"metadata": {
			"name": "%s",
			"namespace": "%s"
		}
	}]`, nodeFeatureRuleExampleName, nodeFeatureRuleNamespace)

	nfdRuleTestSchemes = []clients.SchemeAttacher{
		nfdv1.AddToScheme,
	}
)

func TestNewnodeFeatureRuleBuilderFromObjectString(t *testing.T) {
	testCases := []struct {
		name              string
		almString         string
		client            bool
		expectedErrorText string
	}{
		{
			name:              "Valid ALM Example with Client",
			almString:         nodeFeatureRuleAlmExample,
			client:            true,
			expectedErrorText: "",
		},
		{
			name:              "Empty ALM Example",
			almString:         "",
			client:            true,
			expectedErrorText: "error initializing NodeFeatureRule from alm-examples: almExample is an empty string",
		},
		{
			name:      "Invalid ALM Example",
			almString: "{invalid}",
			client:    true,
			expectedErrorText: "error initializing NodeFeatureRule from alm-examples:" +
				" invalid character 'i' looking for beginning of object key string",
		},
		{
			name:              "No Client Provided",
			almString:         nodeFeatureRuleAlmExample,
			client:            false,
			expectedErrorText: "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var client *clients.Settings
			if testCase.client {
				client = buildTestClientWithNFDRuleScheme()
			}

			builder := NewNodeFeatureRuleBuilderFromObjectString(client, testCase.almString)

			errormessage := ""

			if builder != nil {
				errormessage = builder.errorMsg
			}

			if testCase.client {
				assert.Equal(t, testCase.expectedErrorText, errormessage)

				if testCase.expectedErrorText == "" {
					assert.Equal(t, nodeFeatureRuleExampleName, builder.Definition.Name)
				}
			} else {
				assert.Nil(t, builder)
			}
		})
	}
}

func TestNodeFeatureRuleBuilderCreate(t *testing.T) {
	testCases := []struct {
		name          string
		builder       *NodeFeatureRuleBuilder
		expectedError error
	}{
		{
			name:          "Valid Create",
			builder:       buildValidNFDRuleTestBuilder(buildTestClientWithDummyNFDRule()),
			expectedError: nil,
		},
		{
			name:          "Invalid Builder",
			builder:       buildInvalidNFDRuleTestBuilder(buildTestClientWithDummyNFDRule()),
			expectedError: fmt.Errorf("can not redefine the undefined nodeFeatureRule"),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			builder, err := testCase.builder.Create()
			assert.Equal(t, testCase.expectedError, err)

			if testCase.expectedError == nil {
				assert.NotNil(t, builder.Object)
				assert.Equal(t, builder.Definition.Name, builder.Object.Name)
			}
		})
	}
}

func TestNodeFeatureRuleBuilderExists(t *testing.T) {
	testCases := []struct {
		name           string
		builder        *NodeFeatureRuleBuilder
		expectedStatus bool
	}{
		{
			name:           "Existing Object",
			builder:        buildValidNFDRuleTestBuilder(buildTestClientWithDummyNFDRule()),
			expectedStatus: true,
		},
		{
			name:           "Non-Existent Object",
			builder:        buildValidNFDRuleTestBuilder(buildTestClientWithNFDRuleScheme()),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			exists := testCase.builder.Exists()
			assert.Equal(t, testCase.expectedStatus, exists)
		})
	}
}

func TestNodeFeatureRuleBuilderGet(t *testing.T) {
	testCases := []struct {
		name          string
		builder       *NodeFeatureRuleBuilder
		expectedError error
	}{
		{
			name:          "Valid Get",
			builder:       buildValidNFDRuleTestBuilder(buildTestClientWithDummyNFDRule()),
			expectedError: nil,
		},
		{
			name: "Invalid Get - Missing Object",
			builder: buildValidNFDRuleTestBuilder(
				buildTestClientWithNFDRuleScheme(),
			),
			expectedError: fmt.Errorf("nodefeaturerules.nfd.openshift.io \"%s\" not found", nodeFeatureRuleExampleName),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			obj, err := testCase.builder.Get()
			if testCase.expectedError == nil {
				assert.NotNil(t, obj)
			} else {
				assert.Equal(t, testCase.expectedError.Error(), err.Error())
			}

			if testCase.expectedError == nil {
				assert.NotNil(t, obj)
				assert.Equal(t, testCase.builder.Definition.Name, obj.Name)
			}
		})
	}
}

// Helper Functions

func buildTestClientWithDummyNFDRule() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyNFDRule(nodeFeatureRuleExampleName, nodeFeatureRuleNamespace),
		},
		SchemeAttachers: nfdRuleTestSchemes,
	})
}

func buildDummyNFDRule(name, namespace string) *nfdv1.NodeFeatureRule {
	return &nfdv1.NodeFeatureRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: nfdv1.NodeFeatureRuleSpec{},
	}
}

func buildValidNFDRuleTestBuilder(apiClient *clients.Settings) *NodeFeatureRuleBuilder {
	return NewNodeFeatureRuleBuilderFromObjectString(apiClient, nodeFeatureRuleAlmExample)
}

func buildInvalidNFDRuleTestBuilder(apiClient *clients.Settings) *NodeFeatureRuleBuilder {
	return NewNodeFeatureRuleBuilderFromObjectString(apiClient, "{invalid}")
}

func buildTestClientWithNFDRuleScheme() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: nfdRuleTestSchemes,
	})
}
