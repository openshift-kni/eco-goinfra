package deployment

import (
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

//nolint:funlen
func TestPull(t *testing.T) {
	int32Ptr := func(i int32) *int32 { return &i }
	generateDeployment := func(name, namespace string) *appsv1.Deployment {
		return &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: int32Ptr(1),
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": name,
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": name,
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "test-container",
								Image: "test-image",
							},
						},
					},
				},
			},
		}
	}

	testCases := []struct {
		deploymentName      string
		deploymentNamespace string
		expectedError       bool
		addToRuntimeObjects bool
		expectedErrorText   string
	}{
		{
			deploymentName:      "test1",
			deploymentNamespace: "test-namespace",
			expectedError:       false,
			addToRuntimeObjects: true,
		},
		{
			deploymentName:      "test2",
			deploymentNamespace: "test-namespace",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "deployment object test2 doesn't exist in namespace test-namespace",
		},
		{
			deploymentName:      "",
			deploymentNamespace: "test-namespace",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "deployment 'name' cannot be empty",
		},
		{
			deploymentName:      "test3",
			deploymentNamespace: "",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "deployment 'namespace' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testDeployment := generateDeployment(testCase.deploymentName, testCase.deploymentNamespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testDeployment)
			testSettings = clients.GetTestClients(runtimeObjects)
		}

		// Test the Pull method
		builderResult, err := Pull(testSettings, testDeployment.Name, testDeployment.Namespace)

		// Check the error
		if testCase.expectedError {
			assert.NotNil(t, err)

			// Check the error message
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, builderResult.errorMsg)
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testDeployment.Name, builderResult.Object.Name)
		}
	}
}
