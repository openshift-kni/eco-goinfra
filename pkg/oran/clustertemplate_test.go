package oran

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	provisioningv1alpha1 "github.com/openshift-kni/oran-o2ims/api/provisioning/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	defaultClusterTemplateName      = "test-cluster-template"
	defaultClusterTemplateNamespace = "test-namespace"
)

var defaultClusterTemplateCondition = metav1.Condition{
	Type:   string(provisioningv1alpha1.CTconditionTypes.Validated),
	Reason: string(provisioningv1alpha1.CTconditionReasons.Completed),
	Status: metav1.ConditionTrue,
}

func TestPullClusterTemplate(t *testing.T) {
	testCases := []struct {
		name                string
		nsname              string
		addToRuntimeObjects bool
		client              bool
		expectedError       error
	}{
		{
			name:                defaultClusterTemplateName,
			nsname:              defaultClusterTemplateNamespace,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       nil,
		},
		{
			name:                "",
			nsname:              defaultClusterTemplateNamespace,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("clusterTemplate 'name' cannot be empty"),
		},
		{
			name:                defaultClusterTemplateName,
			nsname:              "",
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("clusterTemplate 'nsname' cannot be empty"),
		},
		{
			name:                defaultClusterTemplateName,
			nsname:              defaultClusterTemplateNamespace,
			addToRuntimeObjects: false,
			client:              true,
			expectedError: fmt.Errorf("clusterTemplate object %s does not exist in namespace %s",
				defaultClusterTemplateName, defaultClusterTemplateNamespace),
		},
		{
			name:                defaultClusterTemplateName,
			nsname:              defaultClusterTemplateNamespace,
			addToRuntimeObjects: true,
			client:              false,
			expectedError:       fmt.Errorf("clusterTemplate 'apiClient' cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects,
				buildDummyClusterTemplate(defaultClusterTemplateName, defaultClusterTemplateNamespace))
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: provisioningTestSchemes,
			})
		}

		testBuilder, err := PullClusterTemplate(testSettings, testCase.name, testCase.nsname)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, testBuilder.Definition.Name)
			assert.Equal(t, testCase.nsname, testBuilder.Definition.Namespace)
		}
	}
}

func TestClusterTemplateGet(t *testing.T) {
	testCases := []struct {
		testBuilder   *ClusterTemplateBuilder
		expectedError string
	}{
		{
			testBuilder:   buildValidClusterTemplateTestBuilder(buildTestClientWithDummyClusterTemplate()),
			expectedError: "",
		},
		{
			testBuilder: buildValidClusterTemplateTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Sprintf(
				"clustertemplates.o2ims.provisioning.oran.org \"%s\" not found", defaultClusterTemplateName),
		},
	}

	for _, testCase := range testCases {
		clusterTemplate, err := testCase.testBuilder.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, testCase.testBuilder.Definition.Name, clusterTemplate.Name)
			assert.Equal(t, testCase.testBuilder.Definition.Namespace, clusterTemplate.Namespace)
		} else {
			assert.EqualError(t, err, testCase.expectedError)
		}
	}
}

func TestClusterTemplateExists(t *testing.T) {
	testCases := []struct {
		testBuilder *ClusterTemplateBuilder
		exists      bool
	}{
		{
			testBuilder: buildValidClusterTemplateTestBuilder(buildTestClientWithDummyClusterTemplate()),
			exists:      true,
		},
		{
			testBuilder: buildValidClusterTemplateTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

func TestClusterTemplateWaitForCondition(t *testing.T) {
	testCases := []struct {
		conditionMet  bool
		exists        bool
		expectedError error
	}{
		{
			conditionMet:  true,
			exists:        true,
			expectedError: nil,
		},
		{
			conditionMet:  false,
			exists:        true,
			expectedError: context.DeadlineExceeded,
		},
		{
			conditionMet:  true,
			exists:        false,
			expectedError: fmt.Errorf("cannot wait for non-existent ClusterTemplate"),
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		if testCase.exists {
			clusterTemplate := buildDummyClusterTemplate(defaultClusterTemplateName, defaultClusterTemplateNamespace)

			if testCase.conditionMet {
				clusterTemplate.Status.Conditions = append(clusterTemplate.Status.Conditions, defaultClusterTemplateCondition)
			}

			runtimeObjects = append(runtimeObjects, clusterTemplate)
		}

		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  runtimeObjects,
			SchemeAttachers: provisioningTestSchemes,
		})
		testBuilder := buildValidClusterTemplateTestBuilder(testSettings)

		_, err := testBuilder.WaitForCondition(defaultClusterTemplateCondition, time.Second)
		assert.Equal(t, testCase.expectedError, err)
	}
}

// buildDummyClusterTemplate returns a ClusterTemplate with the provided name and nsname.
//
//nolint:unparam
func buildDummyClusterTemplate(name, nsname string) *provisioningv1alpha1.ClusterTemplate {
	return &provisioningv1alpha1.ClusterTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsname,
		},
	}
}

// buildTestClientWithDummyClusterTemplate returns an apiClient with the correct schemes and a ClusterTemplate with
// default name and namespace.
func buildTestClientWithDummyClusterTemplate() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyClusterTemplate(defaultClusterTemplateName, defaultClusterTemplateNamespace),
		},
		SchemeAttachers: provisioningTestSchemes,
	})
}

// buildValidClusterTemplateTestBuilder returns a valid ClusterTemplateBuilder with all defaults and the provided
// apiClient.
func buildValidClusterTemplateTestBuilder(apiClient *clients.Settings) *ClusterTemplateBuilder {
	_ = apiClient.AttachScheme(provisioningv1alpha1.AddToScheme)

	return &ClusterTemplateBuilder{
		Definition: buildDummyClusterTemplate(defaultClusterTemplateName, defaultClusterTemplateNamespace),
		apiClient:  apiClient,
	}
}
