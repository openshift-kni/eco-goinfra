package commonbuilder

import (
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	siteconfigv1alpha1 "github.com/openshift-kni/eco-goinfra/pkg/schemes/siteconfig/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestCommonBuilderValidate(t *testing.T) {
	testCases := []struct {
		builderNil    bool
		definitionNil bool
		apiClientNil  bool
		expectedError string
	}{
		{
			definitionNil: true,
			apiClientNil:  false,
			expectedError: "can not redefine the undefined clusterinstance",
		},
		{
			definitionNil: false,
			apiClientNil:  true,
			expectedError: "clusterinstance builder cannot have nil apiClient",
		},
		{
			definitionNil: false,
			apiClientNil:  false,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateTestBuilder()

		if testCase.definitionNil {
			testBuilder.Definition = nil
		}

		if testCase.apiClientNil {
			testBuilder.apiClient = nil
		}

		testCommonBuilder := New(testBuilder)

		result, err := testCommonBuilder.Validate()
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

func generateTestBuilder() *testBuilder {
	return &testBuilder{
		apiClient: clients.GetTestClients(
			clients.TestClientParams{}).Client,
		Definition: generateClusterInstance(),
	}
}

type testBuilder struct {
	apiClient  goclient.Client
	Definition *siteconfigv1alpha1.ClusterInstance
	errorMsg   string
}

func (t *testBuilder) GetClient() goclient.Client {
	return t.apiClient
}

func (t *testBuilder) GetDefinition() goclient.Object {
	return t.Definition
}

func (t *testBuilder) GetErrorMsg() string {
	return t.errorMsg
}

func (t *testBuilder) GetKind() string {
	return siteconfigv1alpha1.ClusterInstanceKind
}

func generateClusterInstance() *siteconfigv1alpha1.ClusterInstance {
	return &siteconfigv1alpha1.ClusterInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-common-builder",
			Namespace: "test-common-builder",
		},
	}
}
