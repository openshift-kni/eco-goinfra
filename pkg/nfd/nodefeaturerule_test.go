package nfd_test

import (
	"fmt"
	"testing"

	. "github.com/openshift-kni/eco-goinfra/pkg/nfd"
	"k8s.io/apimachinery/pkg/runtime"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	nfdv1 "github.com/openshift/node-feature-discovery/api/nfd/v1alpha1"
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
			expectedErrorText: "NodeFeatureRule definition is nil",
		},
		{
			name:              "Invalid ALM Example",
			almString:         "{invalid}",
			client:            true,
			expectedErrorText: "NodeFeatureRule definition is nil",
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

			if testCase.client {
				assert.Equal(t, testCase.expectedErrorText, builder.GetErrorMessage())

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
			expectedError: fmt.Errorf("can not redefine the undefined NodeFeatureRule"),
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
			builder:        buildValidNFDRuleTestBuilder(buildTestClientWithNFDRuleSchemeOnly()),
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
				buildTestClientWithNFDRuleSchemeOnly(),
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

func buildTestClientWithNFDRuleSchemeOnly() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: nfdRuleTestSchemes,
	})
}
