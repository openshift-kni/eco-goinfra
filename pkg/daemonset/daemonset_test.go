package daemonset

import (
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func buildValidTestBuilderWithClient(objects []runtime.Object) *Builder {
	fakeClient := k8sfake.NewSimpleClientset(objects...)

	return NewBuilder(&clients.Settings{
		K8sClient:       fakeClient,
		CoreV1Interface: fakeClient.CoreV1(),
		AppsV1Interface: fakeClient.AppsV1(),
	}, "test-name", "test-namespace", map[string]string{
		"test-key": "test-value",
	}, corev1.Container{
		Name: "test-container",
	})
}

func TestWithNodeSelector(t *testing.T) {
	testBuilder := buildValidTestBuilderWithClient([]runtime.Object{})
	testBuilder.WithNodeSelector(map[string]string{
		"test-node-selector-key": "test-node-selector-value",
	})

	assert.Equal(t, "test-node-selector-value",
		testBuilder.Definition.Spec.Template.Spec.NodeSelector["test-node-selector-key"])

	testBuilder.WithNodeSelector(map[string]string{})
	assert.Equal(t, "cannot accept empty map as nodeselector", testBuilder.errorMsg)
}

func TestWithAdditionalContainerSpecs(t *testing.T) {
	testBuilder := buildValidTestBuilderWithClient([]runtime.Object{})
	testBuilder.WithAdditionalContainerSpecs([]corev1.Container{
		{
			Name: "test-additional-container",
		},
	})

	assert.Equal(t, "test-additional-container",
		testBuilder.Definition.Spec.Template.Spec.Containers[1].Name)

	testBuilder.WithAdditionalContainerSpecs([]corev1.Container{})
	assert.Equal(t, "cannot accept empty list as container specs", testBuilder.errorMsg)
}

func TestWithOptions(t *testing.T) {
	testBuilder := buildValidTestBuilderWithClient([]runtime.Object{})
	testBuilder.WithOptions(func(builder *Builder) (*Builder, error) {
		builder.Definition.Spec.Template.Spec.Containers[0].Name = "test-container-name"

		return builder, nil
	})

	assert.Equal(t, "test-container-name",
		testBuilder.Definition.Spec.Template.Spec.Containers[0].Name)
}

func TestNewDaemonsetBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		nodeLabels    map[string]string
		containerSpec corev1.Container
		expectedError string
		apiClientNil  bool
	}{
		{ // Test case 1 - API client is nil
			name:      "test-name",
			namespace: "test-namespace",
			nodeLabels: map[string]string{
				"test-key": "test-value",
			},
			containerSpec: corev1.Container{
				Name: "test-container",
			},
			expectedError: "",
			apiClientNil:  true,
		},
		{ // Test case 2 - name is empty
			name:          "",
			namespace:     "test-namespace",
			nodeLabels:    map[string]string{},
			containerSpec: corev1.Container{},
			expectedError: "daemonset 'name' cannot be empty",
			apiClientNil:  false,
		},
		{ // Test case 3 - namespace is empty
			name:          "test-name",
			namespace:     "",
			nodeLabels:    map[string]string{},
			containerSpec: corev1.Container{},
			expectedError: "daemonset 'namespace' cannot be empty",
			apiClientNil:  false,
		},
		{ // Test case 4 - API client is not nil
			name:      "test-name",
			namespace: "test-namespace",
			nodeLabels: map[string]string{
				"test-key": "test-value",
			},
			containerSpec: corev1.Container{
				Name: "test-container",
			},
			expectedError: "",
			apiClientNil:  false,
		},
		{ // Test case 5 - labels are empty
			name:       "test-name",
			namespace:  "test-namespace",
			nodeLabels: map[string]string{},
			containerSpec: corev1.Container{
				Name: "test-container",
			},
			expectedError: "daemonset 'labels' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		var testClients *clients.Settings

		if testCase.apiClientNil {
			testClients = nil
		} else {
			testClients = clients.GetTestClients(clients.TestClientParams{})
		}

		testBuilder := NewBuilder(testClients, testCase.name, testCase.namespace, testCase.nodeLabels, testCase.containerSpec)

		if testCase.apiClientNil {
			assert.Nil(t, testBuilder)
		} else {
			if testCase.expectedError != "" {
				assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)
			} else {
				assert.NotNil(t, testBuilder)
				assert.Equal(t, testCase.name, testBuilder.Definition.Name)
				assert.Equal(t, testCase.namespace, testBuilder.Definition.Namespace)
				assert.Equal(t, testCase.nodeLabels, testBuilder.Definition.Spec.Selector.MatchLabels)
				assert.Equal(t, testCase.containerSpec.Name, testBuilder.Definition.Spec.Template.Spec.Containers[0].Name)
			}
		}
	}
}

func TestDaemonsetPull(t *testing.T) {
	generateDaemonset := func(name, namespace string) *appsv1.DaemonSet {
		return &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
	}

	testCases := []struct {
		name                string
		namespace           string
		expectedError       bool
		addToRuntimeObjects bool
		expectedErrorText   string
	}{
		{ // Test Case 1 - happy path
			name:                "test-name",
			namespace:           "test-namespace",
			expectedError:       false,
			addToRuntimeObjects: true,
			expectedErrorText:   "",
		},
		{ // Test Case 2 - daemonset not found
			name:                "test-name",
			namespace:           "test-namespace",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "daemonset object test-name does not exist in namespace test-namespace",
		},
		{ // Test Case 3 - daemonset name is empty
			name:                "",
			namespace:           "test-namespace",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "daemonset name cannot be empty",
		},
		{ // Test Case 4 - daemonset namespace is empty
			name:                "test-name",
			namespace:           "",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "daemonset namespace cannot be empty",
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testDaemonset := generateDaemonset(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testDaemonset)
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		builderResult, err := Pull(testSettings, testCase.name, testCase.namespace)

		if testCase.expectedError {
			assert.NotNil(t, err)

			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.NotNil(t, builderResult)
			assert.Equal(t, testCase.name, builderResult.Definition.Name)
			assert.Equal(t, testCase.namespace, builderResult.Definition.Namespace)
		}
	}
}

func TestDaemonsetWithHostNetwork(t *testing.T) {
	testBuilder := buildValidTestBuilderWithClient([]runtime.Object{})
	testBuilder.WithHostNetwork()

	assert.Equal(t, true, testBuilder.Definition.Spec.Template.Spec.HostNetwork)
}

func TestDaemonsetWithVolume(t *testing.T) {
	testBuilder := buildValidTestBuilderWithClient([]runtime.Object{})
	testBuilder.WithVolume(corev1.Volume{
		Name: "test-volume",
	})

	assert.Equal(t, "test-volume", testBuilder.Definition.Spec.Template.Spec.Volumes[0].Name)

	testBuilder.WithVolume(corev1.Volume{})
	assert.Equal(t, "Volume name parameter is empty", testBuilder.errorMsg)
}

func TestDaemonsetCreate(t *testing.T) {
	testCases := []struct {
		existsAlready bool
	}{
		{ // Test Case 1 - daemonset does not exist
			existsAlready: false,
		},
		{ // Test Case 2 - daemonset exists
			existsAlready: true,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		if testCase.existsAlready {
			runtimeObjects = append(runtimeObjects, &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test1",
					Namespace: "testNamespace",
				},
			})
		}

		testBuilder := buildValidTestBuilderWithClient(runtimeObjects)

		builder, err := testBuilder.Create()
		assert.Nil(t, err)
		assert.NotNil(t, builder)
		assert.Equal(t, "test-name", builder.Definition.Name)
		assert.Equal(t, "test-namespace", builder.Definition.Namespace)
	}
}

func TestDaemonsetUpdate(t *testing.T) {
	testCases := []struct {
		existsAlready bool
	}{
		{ // Test Case 1 - daemonset does not exist
			existsAlready: false,
		},
		{ // Test Case 2 - daemonset exists
			existsAlready: true,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		if testCase.existsAlready {
			runtimeObjects = append(runtimeObjects, &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-name",
					Namespace: "test-namespace",
				},
				Spec: appsv1.DaemonSetSpec{
					UpdateStrategy: appsv1.DaemonSetUpdateStrategy{
						Type: appsv1.OnDeleteDaemonSetStrategyType,
					},
				},
			})
		}

		testBuilder := buildValidTestBuilderWithClient(runtimeObjects)

		testBuilder.Definition.Spec.UpdateStrategy = appsv1.DaemonSetUpdateStrategy{
			Type: appsv1.RollingUpdateDaemonSetStrategyType,
		}

		builder, err := testBuilder.Update()

		if testCase.existsAlready {
			assert.Nil(t, err)
			assert.NotNil(t, builder)
			assert.Equal(t, "test-name", builder.Definition.Name)
			assert.Equal(t, "test-namespace", builder.Definition.Namespace)
			assert.Equal(t, appsv1.RollingUpdateDaemonSetStrategyType, builder.Definition.Spec.UpdateStrategy.Type)
		} else {
			assert.NotNil(t, err)
			assert.Equal(t, "daemonsets.apps \"test-name\" not found", err.Error())
		}
	}
}

func TestDaemonsetDelete(t *testing.T) {
	testCases := []struct {
		existsAlready bool
	}{
		{ // Test Case 1 - daemonset does not exist
			existsAlready: false,
		},
		{ // Test Case 2 - daemonset exists
			existsAlready: true,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		if testCase.existsAlready {
			runtimeObjects = append(runtimeObjects, &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-name",
					Namespace: "test-namespace",
				},
			})
		}

		testBuilder := buildValidTestBuilderWithClient(runtimeObjects)

		err := testBuilder.Delete()
		assert.Nil(t, err)
		assert.Nil(t, testBuilder.Object)
	}
}

func TestDaemonsetValidate(t *testing.T) {
	testCases := []struct {
		builderNil    bool
		definitionNil bool
		apiClientNil  bool
		expectedError string
		builderErrMsg string
	}{
		{
			builderNil:    true,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: "error: received nil DaemonSet builder",
			builderErrMsg: "",
		},
		{
			builderNil:    false,
			definitionNil: true,
			apiClientNil:  false,
			expectedError: "can not redefine the undefined DaemonSet",
			builderErrMsg: "",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  true,
			expectedError: "DaemonSet builder cannot have nil apiClient",
			builderErrMsg: "",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: "",
			builderErrMsg: "",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: "",
			builderErrMsg: "test error",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestBuilderWithClient([]runtime.Object{})

		if testCase.builderNil {
			testBuilder = nil
		}

		if testCase.definitionNil {
			testBuilder.Definition = nil
		}

		if testCase.apiClientNil {
			testBuilder.apiClient = nil
		}

		if testCase.builderErrMsg != "" {
			testBuilder.errorMsg = testCase.builderErrMsg
		}

		result, err := testBuilder.validate()
		if testCase.expectedError != "" {
			assert.NotNil(t, err)
			assert.Equal(t, testCase.expectedError, err.Error())
			assert.False(t, result)
		} else {
			if testCase.builderErrMsg != "" {
				assert.Equal(t, testCase.builderErrMsg, testBuilder.errorMsg)
			} else {
				assert.Nil(t, err)
				assert.True(t, result)
			}
		}
	}
}
