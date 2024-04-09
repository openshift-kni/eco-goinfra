package deployment

import (
	"testing"
	"time"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	multus "gopkg.in/k8snetworkplumbingwg/multus-cni.v4/pkg/types"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "k8s.io/client-go/kubernetes/fake"
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
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		// Test the Pull method
		builderResult, err := Pull(testSettings, testDeployment.Name, testDeployment.Namespace)

		// Check the error
		if testCase.expectedError {
			assert.NotNil(t, err)

			// Check the error message
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testDeployment.Name, builderResult.Object.Name)
			assert.Equal(t, testDeployment.Namespace, builderResult.Object.Namespace)
		}
	}
}

// buildValidTestBuilder returns a valid Builder for testing purposes.
func buildValidTestBuilder() *Builder {
	return NewBuilder(&clients.Settings{
		Client:          nil,
		AppsV1Interface: k8sfake.NewSimpleClientset().AppsV1(),
	}, "test-name", "test-namespace", map[string]string{
		"test-key": "test-value",
	}, &corev1.Container{
		Name: "test-container",
	})
}

func buildTestBuilderWithFakeObjects(objects []runtime.Object) *Builder {
	fakeClient := k8sfake.NewSimpleClientset(objects...)

	return NewBuilder(&clients.Settings{
		K8sClient:       fakeClient,
		CoreV1Interface: fakeClient.CoreV1(),
		AppsV1Interface: fakeClient.AppsV1(),
	}, "test-name", "test-namespace", map[string]string{
		"test-key": "test-value",
	}, &corev1.Container{
		Name: "test-container",
	})
}

func TestWithNodeSelector(t *testing.T) {
	testBuilder := buildValidTestBuilder()

	testBuilder.WithNodeSelector(map[string]string{
		"test-node-selector-key": "test-node-selector-value",
	})

	assert.Empty(t, testBuilder.errorMsg)

	assert.Equal(t, "test-node-selector-value",
		testBuilder.Definition.Spec.Template.Spec.NodeSelector["test-node-selector-key"])
}

func TestWithReplicas(t *testing.T) {
	testBuilder := buildValidTestBuilder()

	testBuilder.WithReplicas(3)

	assert.Equal(t, int32(3), *testBuilder.Definition.Spec.Replicas)
}

func TestWithAdditionalContainerSpecs(t *testing.T) {
	testCases := []struct {
		specsAvailable bool
		expectedErrMsg string
	}{
		{
			specsAvailable: true,
			expectedErrMsg: "",
		},
		{
			specsAvailable: false,
			expectedErrMsg: "cannot accept empty list as container specs",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestBuilder()

		if testCase.specsAvailable {
			testBuilder.WithAdditionalContainerSpecs([]corev1.Container{
				{
					Name: "test-additional-container",
				},
			})
		} else {
			testBuilder.WithAdditionalContainerSpecs([]corev1.Container{})
		}

		assert.Equal(t, testCase.expectedErrMsg, testBuilder.errorMsg)
	}
}

func TestWithSecondaryNetwork(t *testing.T) {
	for _, testCase := range []struct {
		secondaryNetworkAvailable bool
		expectedErrMsg            string
	}{
		{
			secondaryNetworkAvailable: true,
			expectedErrMsg:            "",
		},
		{
			secondaryNetworkAvailable: false,
			expectedErrMsg:            "can not apply empty networks list",
		},
	} {
		testBuilder := buildValidTestBuilder()

		if testCase.secondaryNetworkAvailable {
			testBuilder.WithSecondaryNetwork([]*multus.NetworkSelectionElement{
				{
					Name:      "test-secondary-network",
					Namespace: "test-secondary-network-namespace",
				},
			})
		} else {
			testBuilder.WithSecondaryNetwork(
				[]*multus.NetworkSelectionElement{},
			)
		}

		assert.Equal(t, testCase.expectedErrMsg, testBuilder.errorMsg)

		if testCase.secondaryNetworkAvailable {
			assert.Equal(t,
				"[{\"name\":\"test-secondary-network\",\"namespace\":\"test-secondary-network-namespace\",\"cni-args\":null}]",
				testBuilder.Definition.Spec.Template.Annotations["k8s.v1.cni.cncf.io/networks"])
		}
	}
}

func TestWithHugePages(t *testing.T) {
	testBuilder := buildValidTestBuilder()

	testBuilder.WithHugePages()

	// Assert the volumes are added to the spec
	assert.Equal(t, "hugepages", testBuilder.Definition.Spec.Template.Spec.Volumes[0].Name)
	assert.Equal(t, corev1.StorageMedium("HugePages"),
		testBuilder.Definition.Spec.Template.Spec.Volumes[0].VolumeSource.EmptyDir.Medium)

	// Assert the container is updated with the volume mount
	assert.Equal(t, "hugepages", testBuilder.Definition.Spec.Template.Spec.Containers[0].VolumeMounts[0].Name)
	assert.Equal(t, "/mnt/huge",
		testBuilder.Definition.Spec.Template.Spec.Containers[0].VolumeMounts[0].MountPath)
}

func TestWithSecurityContext(t *testing.T) {
	testCases := []struct {
		securityContextAvailable bool
		expectedErrMsg           string
	}{
		{
			securityContextAvailable: true,
			expectedErrMsg:           "",
		},
		{
			securityContextAvailable: false,
			expectedErrMsg:           "'securityContext' parameter is empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestBuilder()

		if testCase.securityContextAvailable {
			boolVar := true
			testBuilder.WithSecurityContext(&corev1.PodSecurityContext{
				RunAsNonRoot: &boolVar,
			})
		} else {
			testBuilder.WithSecurityContext(nil)
		}

		assert.Equal(t, testCase.expectedErrMsg, testBuilder.errorMsg)

		if testCase.securityContextAvailable {
			assert.Equal(t, true, *testBuilder.Definition.Spec.Template.Spec.SecurityContext.RunAsNonRoot)
		}
	}
}

func TestWithLabel(t *testing.T) {
	testCases := []struct {
		labelKey       string
		labelValue     string
		expectedErrMsg string
		emptyLabels    bool
	}{
		{
			labelKey:    "test-label-key",
			labelValue:  "test-label-value",
			emptyLabels: false,
		},
		{
			labelKey:       "",
			expectedErrMsg: "can not apply empty labelKey",
			emptyLabels:    false,
		},
		{
			labelKey:    "test-label-key",
			labelValue:  "test-label-value",
			emptyLabels: true,
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestBuilder()

		if testCase.emptyLabels {
			testBuilder.Definition.Spec.Template.Labels = nil
		}

		testBuilder.WithLabel(testCase.labelKey, testCase.labelValue)

		assert.Equal(t, testCase.expectedErrMsg, testBuilder.errorMsg)

		if testCase.expectedErrMsg == "" {
			assert.Equal(t, testCase.labelValue, testBuilder.Definition.Spec.Template.Labels[testCase.labelKey])
		}
	}
}

func TestWithServiceAccountName(t *testing.T) {
	testCases := []struct {
		serviceAccountName string
		expectedErrMsg     string
	}{
		{
			serviceAccountName: "test-service-account",
		},
		{
			serviceAccountName: "",
			expectedErrMsg:     "can not apply empty serviceAccount",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestBuilder()

		testBuilder.WithServiceAccountName(testCase.serviceAccountName)

		assert.Equal(t, testCase.expectedErrMsg, testBuilder.errorMsg)

		if testCase.expectedErrMsg == "" {
			assert.Equal(t, testCase.serviceAccountName, testBuilder.Definition.Spec.Template.Spec.ServiceAccountName)
		}
	}
}

func TestWithVolume(t *testing.T) {
	testCases := []struct {
		volumeName     string
		expectedErrMsg string
	}{
		{
			volumeName: "test-volume",
		},
		{
			volumeName:     "",
			expectedErrMsg: "The volume's name cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestBuilder()

		testBuilder.WithVolume(corev1.Volume{
			Name: testCase.volumeName,
		})

		assert.Equal(t, testCase.expectedErrMsg, testBuilder.errorMsg)

		if testCase.expectedErrMsg == "" {
			assert.Equal(t, testCase.volumeName, testBuilder.Definition.Spec.Template.Spec.Volumes[0].Name)
		}
	}
}

func TestWithSchedulerName(t *testing.T) {
	testCases := []struct {
		schedulerName  string
		expectedErrMsg string
	}{
		{
			schedulerName: "test-scheduler",
		},
		{
			schedulerName:  "",
			expectedErrMsg: "Scheduler's name cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestBuilder()

		testBuilder.WithSchedulerName(testCase.schedulerName)

		assert.Equal(t, testCase.expectedErrMsg, testBuilder.errorMsg)

		if testCase.expectedErrMsg == "" {
			assert.Equal(t, testCase.schedulerName, testBuilder.Definition.Spec.Template.Spec.SchedulerName)
		}
	}
}

func TestWithOptions(t *testing.T) {
	testBuilder := buildValidTestBuilder()

	testBuilder.WithOptions(func(builder *Builder) (*Builder, error) {
		return builder, nil
	})

	assert.Equal(t, "", testBuilder.errorMsg)
}

func TestWithToleration(t *testing.T) {
	testCases := []struct {
		toleration     corev1.Toleration
		expectedErrMsg string
	}{
		{
			toleration: corev1.Toleration{
				Key:      "test-toleration-key",
				Operator: "test-toleration-operator",
				Value:    "test-toleration-value",
			},
		},
		{
			toleration:     corev1.Toleration{},
			expectedErrMsg: "The toleration cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestBuilder()

		testBuilder.WithToleration(testCase.toleration)

		assert.Equal(t, testCase.expectedErrMsg, testBuilder.errorMsg)

		if testCase.expectedErrMsg == "" {
			assert.Equal(t, testCase.toleration, testBuilder.Definition.Spec.Template.Spec.Tolerations[0])
		}
	}
}

func TestCreate(t *testing.T) {
	generateTestDeployment := func() *appsv1.Deployment {
		return &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-name",
				Namespace: "test-namespace",
			},
		}
	}

	testCases := []struct {
		deploymentExistsAlready bool
	}{
		{
			deploymentExistsAlready: false,
		},
		{
			deploymentExistsAlready: true,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		if testCase.deploymentExistsAlready {
			runtimeObjects = append(runtimeObjects, generateTestDeployment())
		}

		testBuilder := buildTestBuilderWithFakeObjects(runtimeObjects)
		result, err := testBuilder.Create()
		assert.Nil(t, err)
		assert.Equal(t, testBuilder.Definition.Name, result.Definition.Name)
	}
}

func TestUpdate(t *testing.T) {
	generateTestDeployment := func() *appsv1.Deployment {
		return &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-name",
				Namespace: "test-namespace",
			},
		}
	}

	int32Ptr := func(i int32) *int32 { return &i }

	testCases := []struct {
		deploymentExistsAlready bool
	}{
		{
			deploymentExistsAlready: false,
		},
		{
			deploymentExistsAlready: true,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		if testCase.deploymentExistsAlready {
			runtimeObjects = append(runtimeObjects, generateTestDeployment())
		}

		testBuilder := buildTestBuilderWithFakeObjects(runtimeObjects)

		// Assert the deployment before the update
		assert.NotNil(t, testBuilder.Definition)
		assert.Nil(t, testBuilder.Definition.Spec.Replicas)

		// Set a value in the definition to test the update
		testBuilder.Definition.Spec.Replicas = int32Ptr(3)

		// Perform the update
		result, err := testBuilder.Update()

		// Assert the result
		assert.NotNil(t, testBuilder.Definition)

		if !testCase.deploymentExistsAlready {
			assert.NotNil(t, err)
			assert.Nil(t, result.Object)
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testBuilder.Definition.Name, result.Definition.Name)
			assert.Equal(t, testBuilder.Definition.Spec.Replicas, result.Definition.Spec.Replicas)
		}
	}
}

func TestDelete(t *testing.T) {
	generateTestDeployment := func() *appsv1.Deployment {
		return &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-name",
				Namespace: "test-namespace",
			},
		}
	}

	testCases := []struct {
		deploymentExistsAlready bool
	}{
		{
			deploymentExistsAlready: false,
		},
		{
			deploymentExistsAlready: true,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		if testCase.deploymentExistsAlready {
			runtimeObjects = append(runtimeObjects, generateTestDeployment())
		}

		testBuilder := buildTestBuilderWithFakeObjects(runtimeObjects)
		err := testBuilder.Delete()

		assert.Nil(t, err)
		assert.Nil(t, testBuilder.Object)
	}
}

func TestCreateAndWaitUntilReady(t *testing.T) {
	generateTestDeployment := func() *appsv1.Deployment {
		return &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-name",
				Namespace: "test-namespace",
			},
			Status: appsv1.DeploymentStatus{
				Replicas:      1,
				ReadyReplicas: 1,
			},
		}
	}

	var runtimeObjects []runtime.Object

	runtimeObjects = append(runtimeObjects, generateTestDeployment())

	testBuilder := buildTestBuilderWithFakeObjects(runtimeObjects)

	_, err := testBuilder.CreateAndWaitUntilReady(time.Second * 5)
	assert.Nil(t, err)
}

func TestDeleteAndWait(t *testing.T) {
	generateTestDeployment := func() *appsv1.Deployment {
		return &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-name",
				Namespace: "test-namespace",
			},
		}
	}

	var runtimeObjects []runtime.Object

	runtimeObjects = append(runtimeObjects, generateTestDeployment())

	testBuilder := buildTestBuilderWithFakeObjects(runtimeObjects)

	err := testBuilder.DeleteAndWait(time.Second * 5)
	assert.Nil(t, err)
}

func TestWaitUntilCondition(t *testing.T) {
	generateTestDeployment := func() *appsv1.Deployment {
		return &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-name",
				Namespace: "test-namespace",
			},
			Status: appsv1.DeploymentStatus{
				Replicas:      1,
				ReadyReplicas: 1,
				Conditions: []appsv1.DeploymentCondition{
					{
						Type:   appsv1.DeploymentAvailable,
						Status: corev1.ConditionTrue,
					},
				},
			},
		}
	}

	var runtimeObjects []runtime.Object

	runtimeObjects = append(runtimeObjects, generateTestDeployment())

	testBuilder := buildTestBuilderWithFakeObjects(runtimeObjects)

	err := testBuilder.WaitUntilCondition(appsv1.DeploymentAvailable, time.Second*5)

	assert.Nil(t, err)
}

func TestValidate(t *testing.T) {
	testCases := []struct {
		builderNil    bool
		definitionNil bool
		apiClientNil  bool
		expectedError string
	}{
		{
			builderNil:    true,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: "error: received nil ClusterDeployment builder",
		},
		{
			builderNil:    false,
			definitionNil: true,
			apiClientNil:  false,
			expectedError: "can not redefine the undefined ClusterDeployment",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  true,
			expectedError: "ClusterDeployment builder cannot have nil apiClient",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestBuilder()

		if testCase.builderNil {
			testBuilder = nil
		}

		if testCase.definitionNil {
			testBuilder.Definition = nil
		}

		if testCase.apiClientNil {
			testBuilder.apiClient = nil
		}

		result, err := testBuilder.validate()
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
