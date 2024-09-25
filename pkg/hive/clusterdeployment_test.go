package hive

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	hivev1 "github.com/openshift-kni/eco-goinfra/pkg/schemes/hive/api/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	testSchemes = []clients.SchemeAttacher{
		hivev1.AddToScheme,
	}
	defaultClusterDeploymentName      = "clusterdeploymentname"
	defaultClusterDeploymentNamespace = "clusterdeploymentnamespace"
	defaultClusterName                = "clustername"
	defaultBaseDomain                 = "clusterbasedomain"
	defaultClusterInstallRef          = "clusterInstallRef"
	defaultAgentSelector              = metav1.LabelSelector{MatchLabels: map[string]string{"test": "test"}}
)

func TestNewABMClusterDeploymentBuilder(t *testing.T) {
	testCases := []struct {
		name              string
		nsName            string
		clusterName       string
		baseDomain        string
		clusterInstallRef string
		agentSelector     metav1.LabelSelector
		expectedError     string
	}{
		{
			name:              "testdeployment",
			nsName:            "test-namespace",
			clusterName:       "clustertest",
			baseDomain:        "domaintest",
			clusterInstallRef: "installreftest",
			agentSelector:     metav1.LabelSelector{MatchLabels: map[string]string{"test": "test"}},
			expectedError:     "",
		},
		{
			name:              "",
			nsName:            "test-namespace",
			clusterName:       "clustertest",
			baseDomain:        "domaintest",
			clusterInstallRef: "installreftest",
			agentSelector:     metav1.LabelSelector{MatchLabels: map[string]string{"test": "test"}},
			expectedError:     "clusterdeployment 'name' cannot be empty",
		},
		{
			name:              "testdeployment",
			nsName:            "",
			clusterName:       "clustertest",
			baseDomain:        "domaintest",
			clusterInstallRef: "installreftest",
			agentSelector:     metav1.LabelSelector{MatchLabels: map[string]string{"test": "test"}},
			expectedError:     "clusterdeployment 'namespace' cannot be empty",
		},
		{
			name:              "testdeployment",
			nsName:            "test-namespace",
			clusterName:       "",
			baseDomain:        "domaintest",
			clusterInstallRef: "installreftest",
			agentSelector:     metav1.LabelSelector{MatchLabels: map[string]string{"test": "test"}},
			expectedError:     "clusterdeployment 'clusterName' cannot be empty",
		},
		{
			name:              "testdeployment",
			nsName:            "test-namespace",
			clusterName:       "clustertest",
			baseDomain:        "",
			clusterInstallRef: "installreftest",
			agentSelector:     metav1.LabelSelector{MatchLabels: map[string]string{"test": "test"}},
			expectedError:     "clusterdeployment 'baseDomain' cannot be empty",
		},
		{
			name:              "testdeployment",
			nsName:            "test-namespace",
			clusterName:       "clustertest",
			baseDomain:        "domaintest",
			clusterInstallRef: "",
			agentSelector:     metav1.LabelSelector{MatchLabels: map[string]string{"test": "test"}},
			expectedError:     "clusterdeployment 'clusterInstallRef' cannot be empty",
		},
		{
			name:              "testdeployment",
			nsName:            "test-namespace",
			clusterName:       "clustertest",
			baseDomain:        "domaintest",
			clusterInstallRef: "installreftest",
			agentSelector:     metav1.LabelSelector{},
			expectedError:     "",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})
		testClusterDeployment := NewABMClusterDeploymentBuilder(
			testSettings, testCase.name, testCase.nsName, testCase.clusterName,
			testCase.baseDomain, testCase.clusterInstallRef, testCase.agentSelector)
		assert.Equal(t, testCase.expectedError, testClusterDeployment.errorMsg)
		assert.NotNil(t, testClusterDeployment.Definition)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.name, testClusterDeployment.Definition.Name)
		}
	}
}

func TestPullClusterDeployment(t *testing.T) {
	generateClusterDeployment := func(name, nsName string) *hivev1.ClusterDeployment {
		return &hivev1.ClusterDeployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsName,
			},
			Spec: hivev1.ClusterDeploymentSpec{
				//ReleaseImage: "release-1-1",
			},
		}
	}

	testCases := []struct {
		name                string
		namespace           string
		addToRuntimeObjects bool
		expectedError       error
		client              bool
	}{
		{
			name:                "imageset",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("clusterdeployment 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                "imageset",
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("clusterdeployment 'namespace' cannot be empty"),
			client:              true,
		},
		{
			name:                "imageset",
			namespace:           "test-namespace",
			addToRuntimeObjects: false,
			expectedError:       fmt.Errorf("clusterdeployment object imageset does not exist in namespace test-namespace"),
			client:              true,
		},
		{
			name:                "imageset",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("the apiClient cannot be nil"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testClusterDeployment := generateClusterDeployment(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testClusterDeployment)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		builderResult, err := PullClusterDeployment(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, builderResult.Object.Name)
		}
	}
}

func TestClusterDeploymentGet(t *testing.T) {
	testCases := []struct {
		testClusterDeployment *ClusterDeploymentBuilder
		expectedError         error
	}{
		{
			testClusterDeployment: buildValidClusterDeploymentBuilder(buildClusterDeploymentClientWithDummyObject()),
			expectedError:         nil,
		},
		{
			testClusterDeployment: buildInValidClusterDeploymentBuilder(buildClusterImageSetClientWithDummyObject()),
			expectedError:         fmt.Errorf("clusterdeployment 'namespace' cannot be empty"),
		},
		{
			testClusterDeployment: buildValidClusterDeploymentBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:         fmt.Errorf("clusterdeployments.hive.openshift.io \"clusterdeploymentname\" not found"),
		},
	}

	for _, testCase := range testCases {
		clusterClusterDeployment, err := testCase.testClusterDeployment.Get()
		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}

		if testCase.expectedError == nil {
			assert.Equal(t, clusterClusterDeployment.Name, testCase.testClusterDeployment.Definition.Name)
		}
	}
}

func TestClusterDeploymentCreate(t *testing.T) {
	testCases := []struct {
		testClusterDeployment *ClusterDeploymentBuilder
		expectedError         error
	}{
		{
			testClusterDeployment: buildValidClusterDeploymentBuilder(buildClusterDeploymentClientWithDummyObject()),
			expectedError:         nil,
		},
		{
			testClusterDeployment: buildInValidClusterDeploymentBuilder(buildClusterDeploymentClientWithDummyObject()),
			expectedError:         fmt.Errorf("clusterdeployment 'namespace' cannot be empty"),
		},
		{
			testClusterDeployment: buildValidClusterDeploymentBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:         nil,
		},
	}

	for _, testCase := range testCases {
		clusterClusterDeployment, err := testCase.testClusterDeployment.Create()
		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.testClusterDeployment.Definition.Name, clusterClusterDeployment.Object.Name)
		}
	}
}

func TestClusterDeploymentUpdate(t *testing.T) {
	testCases := []struct {
		testClusterDeployment *ClusterDeploymentBuilder
		expectedError         error
	}{
		{
			testClusterDeployment: buildValidClusterDeploymentBuilder(buildClusterDeploymentClientWithDummyObject()),
			expectedError:         nil,
		},
		{
			testClusterDeployment: buildInValidClusterDeploymentBuilder(buildClusterDeploymentClientWithDummyObject()),
			expectedError:         fmt.Errorf("clusterdeployment 'namespace' cannot be empty"),
		},
		{
			testClusterDeployment: buildValidClusterDeploymentBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:         fmt.Errorf("clusterdeployments.hive.openshift.io \"clusterdeploymentname\" not found"),
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, testCase.testClusterDeployment.Definition.Spec.ClusterName, defaultClusterName)
		testCase.testClusterDeployment.Definition.ResourceVersion = "999"
		testCase.testClusterDeployment.Definition.Spec.ClusterName = "test"
		clusterDeployment, err := testCase.testClusterDeployment.Update(false)

		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedError, err)
			assert.Equal(t, clusterDeployment.Object.Spec.ClusterName, "test")
		}
	}
}

func TestClusterDeploymentDelete(t *testing.T) {
	testCases := []struct {
		testClusterDeployment *ClusterDeploymentBuilder
		expectedError         error
	}{
		{
			testClusterDeployment: buildValidClusterDeploymentBuilder(buildClusterDeploymentClientWithDummyObject()),
			expectedError:         nil,
		},
		{
			testClusterDeployment: buildInValidClusterDeploymentBuilder(buildClusterDeploymentClientWithDummyObject()),
			expectedError:         fmt.Errorf("clusterdeployment 'namespace' cannot be empty"),
		},
		{
			testClusterDeployment: buildValidClusterDeploymentBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:         nil,
		},
	}

	for _, testCase := range testCases {
		err := testCase.testClusterDeployment.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testClusterDeployment.Object)
		}
	}
}

func TestClusterDeploymentExists(t *testing.T) {
	testCases := []struct {
		testClusterDeployment *ClusterDeploymentBuilder
		expectedStatus        bool
	}{
		{
			testClusterDeployment: buildValidClusterDeploymentBuilder(buildClusterDeploymentClientWithDummyObject()),
			expectedStatus:        true,
		},
		{
			testClusterDeployment: buildInValidClusterDeploymentBuilder(buildClusterDeploymentClientWithDummyObject()),
			expectedStatus:        false,
		},
		{
			testClusterDeployment: buildValidClusterDeploymentBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus:        false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testClusterDeployment.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestClusterDeploymentWithOptions(t *testing.T) {
	testSettings := buildClusterDeploymentClientWithDummyObject()
	testBuilder := buildValidClusterDeploymentBuilder(testSettings).WithOptions(
		func(builder *ClusterDeploymentBuilder) (*ClusterDeploymentBuilder, error) {
			return builder, nil
		})

	assert.Equal(t, "", testBuilder.errorMsg)
	testBuilder = buildValidClusterDeploymentBuilder(testSettings).WithOptions(
		func(builder *ClusterDeploymentBuilder) (*ClusterDeploymentBuilder, error) {
			return builder, fmt.Errorf("error")
		})
	assert.Equal(t, "error", testBuilder.errorMsg)
}

func TestClusterDeploymentWithAdditionalAgentSelectorLabels(t *testing.T) {
	testCases := []struct {
		additionalAgentSelector map[string]string
		expectedErrorText       string
	}{
		{
			additionalAgentSelector: map[string]string{"test": "test"},
			expectedErrorText:       "",
		},
		{
			additionalAgentSelector: map[string]string{},
			expectedErrorText:       "agentSelector cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := buildClusterDeploymentClientWithDummyObject()
		clusterDeployment := buildValidClusterDeploymentBuilder(testSettings).WithAdditionalAgentSelectorLabels(
			testCase.additionalAgentSelector)
		assert.Equal(t, testCase.expectedErrorText, clusterDeployment.errorMsg)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, clusterDeployment.Definition.Spec.Platform.AgentBareMetal.AgentSelector.MatchLabels,
				testCase.additionalAgentSelector)
		}
	}
}

func TestClusterDeploymentWithPullSecret(t *testing.T) {
	testCases := []struct {
		psName            string
		expectedErrorText string
	}{
		{
			psName:            "test",
			expectedErrorText: "",
		},
		{
			psName:            "",
			expectedErrorText: "",
		},
	}

	for _, testCase := range testCases {
		testSettings := buildClusterDeploymentClientWithDummyObject()
		clusterDeployment := buildValidClusterDeploymentBuilder(testSettings).WithPullSecret(testCase.psName)
		assert.Equal(t, testCase.expectedErrorText, clusterDeployment.errorMsg)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, clusterDeployment.Definition.Spec.PullSecretRef.Name, testCase.psName)
		}
	}
}

func buildValidClusterDeploymentBuilder(apiClient *clients.Settings) *ClusterDeploymentBuilder {
	return NewABMClusterDeploymentBuilder(
		apiClient,
		defaultClusterDeploymentName,
		defaultClusterDeploymentNamespace,
		defaultClusterName,
		defaultBaseDomain,
		defaultClusterInstallRef,
		defaultAgentSelector)
}

func buildInValidClusterDeploymentBuilder(apiClient *clients.Settings) *ClusterDeploymentBuilder {
	return NewABMClusterDeploymentBuilder(
		apiClient,
		defaultClusterDeploymentName,
		"",
		defaultClusterName,
		defaultBaseDomain,
		defaultClusterInstallRef,
		defaultAgentSelector)
}

func buildClusterDeploymentClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyClusterDeployment(),
		SchemeAttachers: testSchemes,
	})
}

func buildDummyClusterDeployment() []runtime.Object {
	return append([]runtime.Object{}, &hivev1.ClusterDeployment{
		ObjectMeta: metav1.ObjectMeta{
			ResourceVersion: "999",
			Name:            defaultClusterDeploymentName,
			Namespace:       defaultClusterDeploymentNamespace,
		},
		Spec: hivev1.ClusterDeploymentSpec{},
	})
}
