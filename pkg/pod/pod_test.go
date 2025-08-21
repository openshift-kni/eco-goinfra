package pod

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	defaultPodName   = "test-pod"
	defaultPodNsName = "test-ns"
	defaultPodImage  = "test-image"
)

func TestPodNewBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		nsname        string
		image         string
		client        bool
		expectedError string
	}{
		{
			name:          defaultPodName,
			nsname:        defaultPodNsName,
			image:         defaultPodImage,
			client:        true,
			expectedError: "",
		},
		{
			name:          "",
			nsname:        defaultPodNsName,
			image:         defaultPodImage,
			client:        true,
			expectedError: "pod 'name' cannot be empty",
		},
		{
			name:          defaultPodName,
			nsname:        "",
			image:         defaultPodImage,
			client:        true,
			expectedError: "pod 'namespace' cannot be empty",
		},
		{
			name:          defaultPodName,
			nsname:        defaultPodNsName,
			image:         "",
			client:        true,
			expectedError: "pod 'image' cannot be empty",
		},
		{
			name:          defaultPodName,
			nsname:        defaultPodNsName,
			image:         defaultPodImage,
			client:        false,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		testBuilder := NewBuilder(testSettings, testCase.name, testCase.nsname, testCase.image)

		if testCase.client {
			assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

			if testCase.expectedError == "" {
				assert.Equal(t, testCase.name, testBuilder.Definition.Name)
				assert.Equal(t, testCase.nsname, testBuilder.Definition.Namespace)
				assert.NotEmpty(t, testBuilder.Definition.Spec.Containers)
			}
		} else {
			assert.Nil(t, testBuilder)
		}
	}
}

func TestPodPull(t *testing.T) {
	testCases := []struct {
		name                string
		nsname              string
		addToRuntimeObjects bool
		client              bool
		expectedError       error
	}{
		{
			name:                defaultPodName,
			nsname:              defaultPodNsName,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       nil,
		},
		{
			name:                "",
			nsname:              defaultPodNsName,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("pod 'name' cannot be empty"),
		},
		{
			name:                defaultPodName,
			nsname:              "",
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("pod 'namespace' cannot be empty"),
		},
		{
			name:                defaultPodName,
			nsname:              defaultPodNsName,
			addToRuntimeObjects: false,
			client:              true,
			expectedError:       fmt.Errorf("pod object %s does not exist in namespace %s", defaultPodName, defaultPodNsName),
		},
		{
			name:                defaultPodName,
			nsname:              defaultPodNsName,
			addToRuntimeObjects: true,
			client:              false,
			expectedError:       fmt.Errorf("pod 'apiClient' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testPod := buildDummyPod(testCase.name, testCase.nsname, defaultPodImage)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testPod)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})
		}

		testBuilder, err := Pull(testSettings, testCase.name, testCase.nsname)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, testBuilder.Definition.Name)
			assert.Equal(t, testCase.nsname, testBuilder.Definition.Namespace)
			assert.NotEmpty(t, testBuilder.Definition.Spec.Containers)
		}
	}
}

func TestPodCreate(t *testing.T) {
	testCases := []struct {
		testBuilder   *Builder
		expectedError error
	}{
		{
			testBuilder:   buildValidPodTestBuilder(buildTestClientWithDummyPod()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidPodTestBuilder(buildTestClientWithDummyPod()),
			expectedError: fmt.Errorf("pod 'namespace' cannot be empty"),
		},
		{
			testBuilder:   buildValidPodTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		testBuilder, err := testCase.testBuilder.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testBuilder.Definition.Name, testBuilder.Object.Name)
			assert.Equal(t, testBuilder.Definition.Namespace, testBuilder.Object.Namespace)
			assert.NotEmpty(t, testBuilder.Object.Spec.Containers)
		}
	}
}

func TestPodDelete(t *testing.T) {
	testCases := []struct {
		testBuilder   *Builder
		expectedError error
	}{
		{
			testBuilder:   buildValidPodTestBuilder(buildTestClientWithDummyPod()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidPodTestBuilder(buildTestClientWithDummyPod()),
			expectedError: fmt.Errorf("pod 'namespace' cannot be empty"),
		},
		{
			testBuilder:   buildValidPodTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		testBuilder, err := testCase.testBuilder.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testBuilder.Object)
		}
	}
}

func TestPodExists(t *testing.T) {
	testCases := []struct {
		testBuilder *Builder
		exists      bool
	}{
		{
			testBuilder: buildValidPodTestBuilder(buildTestClientWithDummyPod()),
			exists:      true,
		},
		{
			testBuilder: buildInvalidPodTestBuilder(buildTestClientWithDummyPod()),
			exists:      false,
		},
		{
			testBuilder: buildValidPodTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

func TestPodWaitUntilInStatuses(t *testing.T) {
	testCases := []struct {
		pod              *corev1.Pod
		checkedPhases    []corev1.PodPhase
		expectedPodPhase corev1.PodPhase
		expectedError    error
	}{
		{
			pod:              buildDummyPodWithPhaseAndCondition(corev1.PodRunning, corev1.PodReady, false),
			checkedPhases:    []corev1.PodPhase{corev1.PodRunning},
			expectedPodPhase: corev1.PodRunning,
			expectedError:    nil,
		},
		{
			pod:              buildDummyPodWithPhaseAndCondition(corev1.PodRunning, corev1.PodReady, false),
			checkedPhases:    []corev1.PodPhase{corev1.PodRunning, corev1.PodSucceeded},
			expectedPodPhase: corev1.PodRunning,
			expectedError:    nil,
		},
		{
			pod:              buildDummyPodWithPhaseAndCondition(corev1.PodSucceeded, corev1.PodReady, false),
			checkedPhases:    []corev1.PodPhase{corev1.PodRunning, corev1.PodSucceeded},
			expectedPodPhase: corev1.PodSucceeded,
			expectedError:    nil,
		},
		{
			pod:              buildDummyPodWithPhaseAndCondition(corev1.PodFailed, corev1.PodReady, false),
			checkedPhases:    []corev1.PodPhase{corev1.PodRunning, corev1.PodSucceeded},
			expectedPodPhase: "",
			expectedError:    context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: []runtime.Object{testCase.pod},
		})
		testBuilder := buildValidPodTestBuilder(testSettings)

		phase, err := testBuilder.WaitUntilInOneOfStatuses(testCase.checkedPhases, 2*time.Second)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.expectedPodPhase, *phase)
		}
	}
}

func TestPodWaitUntilHealthy(t *testing.T) {
	testCases := []struct {
		pod              *corev1.Pod
		includeSucceeded bool
		skipReadiness    bool
		ignoreFailedPods bool
		expectedError    error
	}{
		{
			pod:              buildDummyPodWithPhaseAndCondition(corev1.PodRunning, corev1.PodReady, false),
			includeSucceeded: true,
			skipReadiness:    false,
			ignoreFailedPods: true,
			expectedError:    nil,
		},
		{
			pod:              buildDummyPodWithPhaseAndCondition(corev1.PodRunning, corev1.PodInitialized, false),
			includeSucceeded: true,
			skipReadiness:    false,
			ignoreFailedPods: true,
			expectedError:    context.DeadlineExceeded,
		},
		{
			pod:              buildDummyPodWithPhaseAndCondition(corev1.PodSucceeded, corev1.PodScheduled, false),
			includeSucceeded: false,
			skipReadiness:    false,
			ignoreFailedPods: true,
			expectedError:    context.DeadlineExceeded,
		},
		{
			pod:              buildDummyPodWithPhaseAndCondition(corev1.PodSucceeded, corev1.PodScheduled, false),
			includeSucceeded: true,
			skipReadiness:    false,
			ignoreFailedPods: true,
			expectedError:    nil,
		},
		{
			pod:              buildDummyPodWithPhaseAndCondition(corev1.PodRunning, corev1.PodScheduled, false),
			includeSucceeded: true,
			skipReadiness:    true,
			ignoreFailedPods: true,
			expectedError:    nil,
		},
		{
			pod:              buildDummyPodWithPhaseAndCondition(corev1.PodFailed, corev1.PodScheduled, false),
			includeSucceeded: true,
			skipReadiness:    false,
			ignoreFailedPods: false,
			expectedError:    context.DeadlineExceeded,
		},
		{
			pod:              buildDummyPodWithPhaseAndCondition(corev1.PodFailed, corev1.PodScheduled, true),
			includeSucceeded: true,
			skipReadiness:    false,
			ignoreFailedPods: true,
			expectedError:    nil,
		},
		{
			pod:              buildDummyPodWithPhaseAndCondition(corev1.PodFailed, corev1.PodScheduled, false),
			includeSucceeded: true,
			skipReadiness:    false,
			ignoreFailedPods: true,
			expectedError:    context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: []runtime.Object{testCase.pod},
		})
		testBuilder := buildValidPodTestBuilder(testSettings)
		testBuilder.Object = testCase.pod

		err := testBuilder.WaitUntilHealthy(
			time.Second, testCase.includeSucceeded, testCase.skipReadiness, testCase.ignoreFailedPods)

		assert.Equal(t, testCase.expectedError, err)
	}
}

// buildDummyPod returns a Pod with the provided name, nsname, and container image.
func buildDummyPod(name, nsname, image string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsname,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "test",
					Image:   image,
					Command: []string{"/bin/bash", "-c", "sleep INF"},
				},
			},
		},
	}
}

// buildTestClientWithDummyPod returns a client with a dummy Pod.
func buildTestClientWithDummyPod() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyPod(defaultPodName, defaultPodNsName, defaultPodImage),
		},
	})
}

// buildValidPodTestBuilder returns a valid Pod builder for testing.
func buildValidPodTestBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, defaultPodName, defaultPodNsName, defaultPodImage)
}

// buildInvalidPodTestBuilder returns an invalid Pod builder for testing.
func buildInvalidPodTestBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, defaultPodName, "", defaultPodImage)
}

// buildDummyPodWithPhaseAndCondition returns a Pod object with the specified phase, condition, and restart policy.
func buildDummyPodWithPhaseAndCondition(
	phase corev1.PodPhase, conditionType corev1.PodConditionType, neverRestart bool) *corev1.Pod {
	pod := buildDummyPod(defaultPodName, defaultPodNsName, defaultPodImage)
	pod.Status.Phase = phase
	pod.Status.Conditions = append(pod.Status.Conditions, corev1.PodCondition{
		Type:   conditionType,
		Status: corev1.ConditionTrue,
	})

	if neverRestart {
		pod.Spec.RestartPolicy = corev1.RestartPolicyNever
	}

	return pod
}
