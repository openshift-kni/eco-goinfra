package replicaset

import (
	"fmt"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

var (
	defaultReplicaSetName      = "test-name"
	defaultReplicaSetNamespace = "test-namespace"
	defaultReplicaSetLabel     = map[string]string{"testLabels": "testLabelValue"}
	defaultReplicaSetContainer = []corev1.Container{{Name: "test-container"}}
)

func TestPullReplicaSet(t *testing.T) {
	generateReplicaSet := func(name, namespace string) *appsv1.ReplicaSet {
		return &appsv1.ReplicaSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				Labels:    defaultReplicaSetLabel,
			},
			Spec: appsv1.ReplicaSetSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: defaultReplicaSetContainer,
					},
				},
			},
		}
	}

	testCases := []struct {
		name                string
		namespace           string
		addToRuntimeObjects bool
		expectedError       error
		apiClient           bool
	}{
		{
			name:                defaultReplicaSetName,
			namespace:           defaultReplicaSetNamespace,
			addToRuntimeObjects: true,
			expectedError:       nil,
			apiClient:           true,
		},
		{
			name:                "",
			namespace:           defaultReplicaSetNamespace,
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("replicaset 'name' cannot be empty"),
			apiClient:           true,
		},
		{
			name:                defaultReplicaSetName,
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("replicaset 'nsname' cannot be empty"),
			apiClient:           true,
		},
		{
			name:                "replicaset-test",
			namespace:           defaultReplicaSetNamespace,
			addToRuntimeObjects: false,
			expectedError: fmt.Errorf("replicaset object replicaset-test does not exist " +
				"in namespace test-namespace"),
			apiClient: true,
		},
		{
			name:                defaultReplicaSetName,
			namespace:           defaultReplicaSetNamespace,
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("replicaset 'apiClient' cannot be empty"),
			apiClient:           false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testReplicaSet := generateReplicaSet(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testReplicaSet)
		}

		if testCase.apiClient {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})
		}

		builderResult, err := Pull(testSettings, testCase.name, testCase.namespace)

		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		} else {
			assert.Equal(t, testReplicaSet.Name, builderResult.Object.Name)
			assert.Nil(t, err)
		}
	}
}

func TestNewReplicaSetBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		labels        map[string]string
		container     []corev1.Container
		expectedError string
		apiClient     bool
	}{
		{
			name:          defaultReplicaSetName,
			namespace:     defaultReplicaSetNamespace,
			labels:        defaultReplicaSetLabel,
			container:     defaultReplicaSetContainer,
			expectedError: "",
			apiClient:     true,
		},
		{
			name:          "",
			namespace:     defaultReplicaSetNamespace,
			labels:        defaultReplicaSetLabel,
			container:     defaultReplicaSetContainer,
			expectedError: "replicaset 'name' cannot be empty",
			apiClient:     true,
		},
		{
			name:          defaultReplicaSetName,
			namespace:     "",
			labels:        defaultReplicaSetLabel,
			container:     defaultReplicaSetContainer,
			expectedError: "replicaset 'nsname' cannot be empty",
			apiClient:     true,
		},
		{
			name:          defaultReplicaSetName,
			namespace:     defaultReplicaSetNamespace,
			labels:        map[string]string{},
			container:     defaultReplicaSetContainer,
			expectedError: "replicaset 'labels' cannot be empty",
			apiClient:     true,
		},
		{
			name:          defaultReplicaSetName,
			namespace:     defaultReplicaSetNamespace,
			labels:        defaultReplicaSetLabel,
			container:     []corev1.Container{},
			expectedError: "replicaset 'containerSpec' cannot be empty",
			apiClient:     true,
		},
		{
			name:          defaultReplicaSetName,
			namespace:     defaultReplicaSetNamespace,
			labels:        defaultReplicaSetLabel,
			container:     defaultReplicaSetContainer,
			expectedError: "",
			apiClient:     false,
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.apiClient {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		testReplicaSetBuilder := NewBuilder(testSettings,
			testCase.name,
			testCase.namespace,
			testCase.labels,
			testCase.container)

		if testCase.expectedError == "" {
			if testCase.apiClient {
				assert.NotNil(t, testReplicaSetBuilder.Definition)
				assert.Equal(t, testCase.name, testReplicaSetBuilder.Definition.Name)
				assert.Equal(t, testCase.namespace, testReplicaSetBuilder.Definition.Namespace)
			} else {
				assert.Nil(t, testReplicaSetBuilder)
			}
		} else {
			assert.Equal(t, testCase.expectedError, testReplicaSetBuilder.errorMsg)
		}
	}
}

func TestReplicaSetExists(t *testing.T) {
	testCases := []struct {
		testReplicaSet *Builder
		expectedStatus bool
	}{
		{
			testReplicaSet: buildValidReplicaSetBuilder(buildReplicaSetClientWithDummyObject()),
			expectedStatus: true,
		},
		{
			testReplicaSet: buildInValidReplicaSetBuilder(buildReplicaSetClientWithDummyObject()),
			expectedStatus: false,
		},
		{
			testReplicaSet: buildValidReplicaSetBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testReplicaSet.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestReplicaSetCreate(t *testing.T) {
	testCases := []struct {
		testReplicaSet *Builder
		expectedError  string
	}{
		{
			testReplicaSet: buildValidReplicaSetBuilder(buildReplicaSetClientWithDummyObject()),
			expectedError:  "",
		},
		{
			testReplicaSet: buildInValidReplicaSetBuilder(buildReplicaSetClientWithDummyObject()),
			expectedError:  "replicaset 'name' cannot be empty",
		},
		{
			testReplicaSet: buildValidReplicaSetBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:  "",
		},
	}

	for _, testCase := range testCases {
		testReplicaSetBuilder, err := testCase.testReplicaSet.Create()

		if testCase.expectedError == "" {
			assert.Equal(t, testReplicaSetBuilder.Definition.Name, testReplicaSetBuilder.Object.Name)
			assert.Equal(t, testReplicaSetBuilder.Definition.Namespace, testReplicaSetBuilder.Object.Namespace)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError, err.Error())
		}
	}
}

func TestReplicaSetDelete(t *testing.T) {
	testCases := []struct {
		testReplicaSet *Builder
		expectedError  error
	}{
		{
			testReplicaSet: buildValidReplicaSetBuilder(buildReplicaSetClientWithDummyObject()),
			expectedError:  nil,
		},
		{
			testReplicaSet: buildInValidReplicaSetBuilder(buildReplicaSetClientWithDummyObject()),
			expectedError:  fmt.Errorf("replicaset 'name' cannot be empty"),
		},
		{
			testReplicaSet: buildValidReplicaSetBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:  nil,
		},
	}

	for _, testCase := range testCases {
		err := testCase.testReplicaSet.Delete()

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testReplicaSet.Object)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestReplicaSetUpdate(t *testing.T) {
	testCases := []struct {
		testReplicaSet *Builder
		testLabels     map[string]string
		expectedError  string
	}{
		{
			testReplicaSet: buildValidReplicaSetBuilder(buildReplicaSetClientWithDummyObject()),
			testLabels:     map[string]string{"test-node-selector-key": "test-node-selector-value"},
			expectedError:  "",
		},
		{
			testReplicaSet: buildValidReplicaSetBuilder(buildReplicaSetClientWithDummyObject()),
			testLabels:     defaultReplicaSetLabel,
			expectedError:  "",
		},
		{
			testReplicaSet: buildValidReplicaSetBuilder(buildReplicaSetClientWithDummyObject()),
			testLabels: map[string]string{"test-node-selector-key1": "test-node-selector-value1",
				"test-node-selector-key2": "test-node-selector-value2"},
			expectedError: "",
		},
		{
			testReplicaSet: buildValidReplicaSetBuilder(buildReplicaSetClientWithDummyObject()),
			testLabels:     map[string]string{"test-node-selector-key": ""},
			expectedError:  "",
		},
		{
			testReplicaSet: buildValidReplicaSetBuilder(buildReplicaSetClientWithDummyObject()),
			testLabels:     map[string]string{"": "test-node-selector-value"},
			expectedError:  "can not apply labels with an empty labelKey value",
		},
		{
			testReplicaSet: buildValidReplicaSetBuilder(buildReplicaSetClientWithDummyObject()),
			testLabels:     map[string]string{},
			expectedError:  "can not apply empty labels",
		},
		{
			testReplicaSet: buildInValidReplicaSetBuilder(buildReplicaSetClientWithDummyObject()),
			expectedError:  "replicaset 'name' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, defaultReplicaSetLabel, testCase.testReplicaSet.Definition.Labels)
		assert.Nil(t, nil, testCase.testReplicaSet.Object)
		testCase.testReplicaSet.WithLabel(testCase.testLabels)
		_, err := testCase.testReplicaSet.Update()

		if testCase.expectedError != "" {
			assert.Equal(t, testCase.expectedError, err.Error())
		} else {
			assert.Equal(t, testCase.testLabels, testCase.testReplicaSet.Definition.Labels)
		}
	}
}

func TestReplicaSetWithLabel(t *testing.T) {
	testCases := []struct {
		labels         map[string]string
		expectedErrMsg string
		emptyLabels    bool
		originalLabels map[string]string
	}{
		{
			labels:         map[string]string{"test-label-key": "test-label-value"},
			expectedErrMsg: "",
			emptyLabels:    true,
			originalLabels: map[string]string{},
		},
		{
			labels:         map[string]string{"test-label-key": ""},
			expectedErrMsg: "",
			emptyLabels:    true,
			originalLabels: map[string]string{},
		},
		{
			labels:         map[string]string{"": "test-label-value"},
			expectedErrMsg: "can not apply labels with an empty labelKey value",
			emptyLabels:    true,
			originalLabels: map[string]string{},
		},
		{
			labels:         map[string]string{},
			expectedErrMsg: "can not apply empty labels",
			emptyLabels:    true,
			originalLabels: map[string]string{},
		},
		{
			labels:         map[string]string{"test-label-key": "test-label-value"},
			expectedErrMsg: "",
			emptyLabels:    false,
			originalLabels: map[string]string{"other-labels": "other-label-value"},
		},
		{
			labels:         map[string]string{"test-label-key": "test-label-value"},
			expectedErrMsg: "",
			emptyLabels:    false,
			originalLabels: map[string]string{"test-label-key": "", "other-labels": "other-label-value"},
		},
		{
			labels:         map[string]string{"test-label-key": ""},
			expectedErrMsg: "",
			emptyLabels:    false,
			originalLabels: map[string]string{"test-label-key": "test-label-value",
				"other-labels": "other-label-value"},
		},
		{
			labels:         map[string]string{"test-label-key": "test-label-value"},
			expectedErrMsg: "",
			emptyLabels:    false,
			originalLabels: map[string]string{"test-label-key": "test-label-value"},
		},
		{
			labels: map[string]string{"test-label-key1": "test-label-value1",
				"test-label-key2": "test-label-value2"},
			expectedErrMsg: "",
			emptyLabels:    false,
			originalLabels: map[string]string{"test-label-key": "test-label-value"},
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidReplicaSetBuilder(buildReplicaSetClientWithDummyObject())

		if !testCase.emptyLabels {
			testBuilder.Definition.Labels = testCase.originalLabels
		}

		testBuilder.WithLabel(testCase.labels)

		assert.Equal(t, testCase.expectedErrMsg, testBuilder.errorMsg)

		if testCase.expectedErrMsg == "" {
			assert.Equal(t, testCase.labels, testBuilder.Definition.Labels)
		}
	}
}

func TestReplicaSetWithNodeSelector(t *testing.T) {
	testCases := []struct {
		nodeSelector         map[string]string
		expectedErrMsg       string
		emptyLabels          bool
		originalNodeSelector map[string]string
	}{
		{
			nodeSelector:         map[string]string{"test-node-selector-key": "test-node-selector-value"},
			expectedErrMsg:       "",
			emptyLabels:          true,
			originalNodeSelector: map[string]string{},
		},
		{
			nodeSelector:         map[string]string{"test-node-selector-key": ""},
			expectedErrMsg:       "",
			emptyLabels:          true,
			originalNodeSelector: map[string]string{},
		},
		{
			nodeSelector:         map[string]string{"": "test-node-selector-value"},
			expectedErrMsg:       "can not apply a nodeSelector with an empty key value",
			emptyLabels:          true,
			originalNodeSelector: map[string]string{},
		},
		{
			nodeSelector:         map[string]string{},
			expectedErrMsg:       "can not apply empty nodeSelector",
			emptyLabels:          true,
			originalNodeSelector: map[string]string{},
		},
		{
			nodeSelector:         map[string]string{"test-node-selector-key": "test-node-selector-value"},
			expectedErrMsg:       "",
			emptyLabels:          false,
			originalNodeSelector: map[string]string{"other-node-selector-key": "other-node-selector-value"},
		},
		{
			nodeSelector:   map[string]string{"test-node-selector-key": "test-node-selector-value"},
			expectedErrMsg: "",
			emptyLabels:    false,
			originalNodeSelector: map[string]string{"test-node-selector-key": "",
				"other-node-selector-key": "other-node-selector-value"},
		},
		{
			nodeSelector:   map[string]string{"test-node-selector-key": ""},
			expectedErrMsg: "",
			emptyLabels:    false,
			originalNodeSelector: map[string]string{"test-node-selector-key": "test-node-selector-value",
				"other-node-selector-key": "other-node-selector-value"},
		},
		{
			nodeSelector:   map[string]string{"test-node-selector-key": "test-node-selector-value"},
			expectedErrMsg: "",
			emptyLabels:    false,
			originalNodeSelector: map[string]string{"test-node-selector-key": "test-node-selector-value",
				"other-node-selector-key": "other-node-selector-value"},
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidReplicaSetBuilder(buildReplicaSetClientWithDummyObject())

		if !testCase.emptyLabels {
			testBuilder.Definition.Spec.Template.Spec.NodeSelector = testCase.originalNodeSelector
		}

		testBuilder.WithNodeSelector(testCase.nodeSelector)

		assert.Equal(t, testCase.expectedErrMsg, testBuilder.errorMsg)

		if testCase.expectedErrMsg == "" {
			assert.Equal(t, testCase.nodeSelector, testBuilder.Definition.Spec.Template.Spec.NodeSelector)
		}
	}
}

func TestReplicaSetWithVolume(t *testing.T) {
	testCases := []struct {
		testVolume        corev1.Volume
		expectedError     bool
		expectedErrorText string
	}{
		{
			testVolume: corev1.Volume{
				Name: "test-volume",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			testVolume: corev1.Volume{
				Name: "",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
			expectedError:     true,
			expectedErrorText: "volume name parameter is empty",
		},
		{
			testVolume:        corev1.Volume{},
			expectedError:     true,
			expectedErrorText: "volume name parameter is empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidReplicaSetBuilder(buildReplicaSetClientWithDummyObject())

		result := testBuilder.WithVolume(testCase.testVolume)

		if testCase.expectedError {
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
			}
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testVolume.Name, result.Definition.Spec.Template.Spec.Volumes[0].Name)
		}
	}
}

func TestReplicaSetWithAdditionalContainerSpecs(t *testing.T) {
	testCases := []struct {
		testSpecs         []corev1.Container
		expectedError     bool
		expectedErrorText string
	}{
		{
			testSpecs: []corev1.Container{
				{
					Name: "test-additional-container",
				},
			},
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			testSpecs:         []corev1.Container{},
			expectedError:     true,
			expectedErrorText: "cannot accept empty list as container specs",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidReplicaSetBuilder(buildReplicaSetClientWithDummyObject())

		result := testBuilder.WithAdditionalContainerSpecs(testCase.testSpecs)

		if testCase.expectedError {
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
			}
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testSpecs[0].Name, result.Definition.Spec.Template.Spec.Containers[1].Name)
		}
	}
}

func buildValidReplicaSetBuilder(apiClient *clients.Settings) *Builder {
	replicaSetBuilder := NewBuilder(
		apiClient, defaultReplicaSetName, defaultReplicaSetNamespace,
		defaultReplicaSetLabel, defaultReplicaSetContainer)

	return replicaSetBuilder
}

func buildInValidReplicaSetBuilder(apiClient *clients.Settings) *Builder {
	replicaSetBuilder := NewBuilder(
		apiClient, "", defaultReplicaSetNamespace,
		defaultReplicaSetLabel, defaultReplicaSetContainer)

	return replicaSetBuilder
}

func buildReplicaSetClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: buildDummyReplicaSet(),
	})
}

func buildDummyReplicaSet() []runtime.Object {
	return append([]runtime.Object{}, &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultReplicaSetName,
			Namespace: defaultReplicaSetNamespace,
			Labels:    defaultReplicaSetLabel,
		},
		Spec: appsv1.ReplicaSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: defaultReplicaSetContainer,
				},
			},
		},
	})
}
