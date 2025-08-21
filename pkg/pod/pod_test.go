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
	defaultPodName     = "test-pod"
	defaultPodNsName   = "test-ns"
	defaultPodImage    = "test-image"
	defaultPodNodeName = "test-node"
	defaultVolumeName  = "test-volume"
	defaultMountPath   = "/test"
)

var podRunningErrorMsg = fmt.Sprintf("can not redefine running pod. pod already running on node %s", defaultPodNodeName)

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

func TestPodDefineOnNode(t *testing.T) {
	testCases := []struct {
		nodeName      string
		hasObject     bool
		expectedError string
	}{
		{
			nodeName:      defaultPodNodeName,
			hasObject:     false,
			expectedError: "",
		},
		{
			nodeName:      "",
			hasObject:     false,
			expectedError: "can not define pod on empty node",
		},
		{
			nodeName:      defaultPodNodeName,
			hasObject:     true,
			expectedError: podRunningErrorMsg,
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidPodTestBuilder(buildTestClientWithDummyPod())

		if testCase.hasObject {
			testBuilder.Object = testBuilder.Definition
			testBuilder.Object.Spec.NodeName = defaultPodNodeName
		}

		testBuilder = testBuilder.DefineOnNode(testCase.nodeName)
		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.nodeName, testBuilder.Definition.Spec.NodeName)
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
	testPodDeleteHelper(t, func(builder *Builder) (*Builder, error) {
		return builder.Delete()
	})
}

func TestPodDeleteAndWait(t *testing.T) {
	testPodDeleteHelper(t, func(builder *Builder) (*Builder, error) {
		return builder.DeleteAndWait(5 * time.Second)
	})
}

func TestPodDeleteImmediate(t *testing.T) {
	testPodDeleteHelper(t, func(builder *Builder) (*Builder, error) {
		return builder.DeleteImmediate()
	})
}

func TestPodWaitUntilDeleted(t *testing.T) {
	testCases := []struct {
		testBuilder   *Builder
		expectedError error
	}{
		{
			testBuilder:   buildValidPodTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidPodTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("pod 'namespace' cannot be empty"),
		},
		{
			testBuilder:   buildValidPodTestBuilder(buildTestClientWithDummyPod()),
			expectedError: context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		err := testCase.testBuilder.WaitUntilDeleted(2 * time.Second)
		assert.Equal(t, testCase.expectedError, err)
	}
}

func TestPodWaitUntilReady(t *testing.T) {
	testPodWaitUntilConditionHelper(t, func(builder *Builder) error {
		return builder.WaitUntilReady(time.Second)
	})
}

func TestPodWaitUntilCondition(t *testing.T) {
	testPodWaitUntilConditionHelper(t, func(builder *Builder) error {
		return builder.WaitUntilCondition(corev1.PodReady, time.Second)
	})
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

func TestPodRedefineDefaultCMD(t *testing.T) {
	testCases := []struct {
		command       []string
		hasObject     bool
		expectedError string
	}{
		{
			command:       []string{"test"},
			hasObject:     false,
			expectedError: "",
		},
		{
			command:       []string{},
			hasObject:     false,
			expectedError: "",
		},
		{
			command:       []string{"test"},
			hasObject:     true,
			expectedError: podRunningErrorMsg,
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidPodTestBuilder(buildTestClientWithDummyPod())

		if testCase.hasObject {
			testBuilder.Object = testBuilder.Definition
			testBuilder.Object.Spec.NodeName = defaultPodNodeName
		}

		testBuilder = testBuilder.RedefineDefaultCMD(testCase.command)
		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.command, testBuilder.Definition.Spec.Containers[0].Command)
		}
	}
}

func TestPodWithRestartPolicy(t *testing.T) {
	testCases := []struct {
		restartPolicy corev1.RestartPolicy
		hasObject     bool
		expectedError string
	}{
		{
			restartPolicy: corev1.RestartPolicyAlways,
			hasObject:     false,
			expectedError: "",
		},
		{
			restartPolicy: "",
			hasObject:     false,
			expectedError: "can not define pod with empty restart policy",
		},
		{
			restartPolicy: corev1.RestartPolicyAlways,
			hasObject:     true,
			expectedError: podRunningErrorMsg,
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidPodTestBuilder(buildTestClientWithDummyPod())

		if testCase.hasObject {
			testBuilder.Object = testBuilder.Definition
			testBuilder.Object.Spec.NodeName = defaultPodNodeName
		}

		testBuilder = testBuilder.WithRestartPolicy(testCase.restartPolicy)
		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.restartPolicy, testBuilder.Definition.Spec.RestartPolicy)
		}
	}
}

func TestPodWithTolerationToMaster(t *testing.T) {
	toleration := corev1.Toleration{
		Key:    "node-role.kubernetes.io/master",
		Effect: "NoSchedule",
	}

	testPodWithTolerationHelper(t, toleration, func(builder *Builder, toleration corev1.Toleration) *Builder {
		return builder.WithTolerationToMaster()
	})
}

func TestPodWithTolerationToControlPlane(t *testing.T) {
	toleration := corev1.Toleration{
		Key:    "node-role.kubernetes.io/control-plane",
		Effect: "NoSchedule",
	}

	testPodWithTolerationHelper(t, toleration, func(builder *Builder, toleration corev1.Toleration) *Builder {
		return builder.WithTolerationToControlPlane()
	})
}

func TestPodWithToleration(t *testing.T) {
	toleration := corev1.Toleration{
		Key:    "node-role.kubernetes.io/control-plane",
		Effect: "NoSchedule",
	}

	testPodWithTolerationHelper(t, toleration, func(builder *Builder, toleration corev1.Toleration) *Builder {
		return builder.WithToleration(toleration)
	})
}

func TestPodWithNodeSelector(t *testing.T) {
	testCases := []struct {
		nodeSelector  map[string]string
		hasObject     bool
		expectedError string
	}{
		{
			nodeSelector:  map[string]string{"test": "test"},
			hasObject:     false,
			expectedError: "",
		},
		{
			nodeSelector:  map[string]string{},
			hasObject:     false,
			expectedError: "can not define pod with empty nodeSelector",
		},
		{
			nodeSelector:  map[string]string{"test": "test"},
			hasObject:     true,
			expectedError: podRunningErrorMsg,
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidPodTestBuilder(buildTestClientWithDummyPod())

		if testCase.hasObject {
			testBuilder.Object = testBuilder.Definition
			testBuilder.Object.Spec.NodeName = defaultPodNodeName
		}

		testBuilder = testBuilder.WithNodeSelector(testCase.nodeSelector)
		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.nodeSelector, testBuilder.Definition.Spec.NodeSelector)
		}
	}
}

func TestPodWithPrivilegedFlag(t *testing.T) {
	testCases := []struct {
		hasObject     bool
		expectedError string
	}{
		{
			hasObject:     false,
			expectedError: "",
		},
		{
			hasObject:     true,
			expectedError: podRunningErrorMsg,
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidPodTestBuilder(buildTestClientWithDummyPod())

		if testCase.hasObject {
			testBuilder.Object = testBuilder.Definition
			testBuilder.Object.Spec.NodeName = defaultPodNodeName
		}

		testBuilder = testBuilder.WithPrivilegedFlag()
		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.True(t, *testBuilder.Definition.Spec.Containers[0].SecurityContext.Privileged)
		}
	}
}

func TestPodWithVolume(t *testing.T) {
	testCases := []struct {
		volume        corev1.Volume
		expectedError string
	}{
		{
			volume:        corev1.Volume{Name: defaultVolumeName},
			expectedError: "",
		},
		{
			volume:        corev1.Volume{},
			expectedError: "the volume's name cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidPodTestBuilder(buildTestClientWithDummyPod())
		testBuilder = testBuilder.WithVolume(testCase.volume)
		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, []corev1.Volume{testCase.volume}, testBuilder.Definition.Spec.Volumes)
		}
	}
}

func TestPodWithLocalVolume(t *testing.T) {
	testCases := []struct {
		volumeName    string
		mountPath     string
		alreadyUsed   bool
		hasObject     bool
		expectedError string
	}{
		{
			volumeName:    defaultVolumeName,
			mountPath:     defaultMountPath,
			alreadyUsed:   false,
			hasObject:     false,
			expectedError: "",
		},
		{
			volumeName:    "",
			mountPath:     defaultMountPath,
			alreadyUsed:   false,
			hasObject:     false,
			expectedError: "'volumeName' parameter is empty",
		},
		{
			volumeName:    defaultVolumeName,
			mountPath:     "",
			alreadyUsed:   false,
			hasObject:     false,
			expectedError: "'mountPath' parameter is empty",
		},
		{
			volumeName:    defaultVolumeName,
			mountPath:     defaultMountPath,
			alreadyUsed:   true,
			hasObject:     false,
			expectedError: fmt.Sprintf("given mount %s already mounted to pod's container test", defaultVolumeName),
		},
		{
			volumeName:    defaultVolumeName,
			mountPath:     defaultMountPath,
			alreadyUsed:   false,
			hasObject:     true,
			expectedError: podRunningErrorMsg,
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidPodTestBuilder(buildTestClientWithDummyPod())

		if testCase.hasObject {
			testBuilder.Object = testBuilder.Definition
			testBuilder.Object.Spec.NodeName = defaultPodNodeName
		}

		if testCase.alreadyUsed {
			testBuilder.Definition.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{{
				Name:      testCase.volumeName,
				MountPath: testCase.mountPath,
			}}
		}

		testBuilder = testBuilder.WithLocalVolume(testCase.volumeName, testCase.mountPath)
		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.volumeName, testBuilder.Definition.Spec.Containers[0].VolumeMounts[0].Name)
			assert.Equal(t, testCase.mountPath, testBuilder.Definition.Spec.Containers[0].VolumeMounts[0].MountPath)
			assert.Equal(t, testCase.volumeName, testBuilder.Definition.Spec.Volumes[0].Name)
		}
	}
}

func TestPodIsHealthy(t *testing.T) {
	testCases := []struct {
		exists          bool
		phase           corev1.PodPhase
		condition       corev1.PodConditionType
		expectedHealthy bool
	}{
		{
			exists:          true,
			phase:           corev1.PodSucceeded,
			condition:       "",
			expectedHealthy: true,
		},
		{
			exists:          true,
			phase:           corev1.PodRunning,
			condition:       corev1.PodReady,
			expectedHealthy: true,
		},
		{
			exists:          false,
			phase:           "",
			condition:       "",
			expectedHealthy: false,
		},
		{
			exists:          true,
			phase:           corev1.PodRunning,
			condition:       "",
			expectedHealthy: false,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		if testCase.exists {
			testPod := buildDummyPodWithPhaseAndCondition(testCase.phase, testCase.condition, false)
			runtimeObjects = append(runtimeObjects, testPod)
		}

		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})
		testBuilder := buildValidPodTestBuilder(testSettings)

		assert.Equal(t, testCase.expectedHealthy, testBuilder.IsHealthy())
	}
}

func testPodDeleteHelper(t *testing.T, deleteFunc func(builder *Builder) (*Builder, error)) {
	t.Helper()

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
		testBuilder, err := deleteFunc(testCase.testBuilder)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testBuilder.Object)
		}
	}
}

// testPodWaitUntilConditionHelper handles the test cases where a function waits for a condition. This helper uses
// corev1.PodReady so that it can also be used for testing WaitUntilReady.
func testPodWaitUntilConditionHelper(t *testing.T, waitFunc func(builder *Builder) error) {
	t.Helper()

	testCases := []struct {
		valid         bool
		ready         bool
		expectedError error
	}{
		{
			valid:         true,
			ready:         true,
			expectedError: nil,
		},
		{
			valid:         false,
			ready:         true,
			expectedError: fmt.Errorf("pod 'namespace' cannot be empty"),
		},
		{
			valid:         true,
			ready:         false,
			expectedError: context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		var testBuilder *Builder

		if testCase.valid {
			pod := buildDummyPodWithPhaseAndCondition(corev1.PodRunning, corev1.PodReady, false)

			if !testCase.ready {
				pod.Status.Conditions[0].Status = corev1.ConditionFalse
			}

			testBuilder = buildValidPodTestBuilder(clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: []runtime.Object{pod},
			}))
		} else {
			testBuilder = buildInvalidPodTestBuilder(clients.GetTestClients(clients.TestClientParams{}))
		}

		err := waitFunc(testBuilder)
		assert.Equal(t, testCase.expectedError, err)
	}
}

// testPodWithTolerationHelper handles the test cases where a function applies the specified toleration to a pod.
func testPodWithTolerationHelper(
	t *testing.T, toleration corev1.Toleration, testFunc func(builder *Builder, toleration corev1.Toleration) *Builder) {
	t.Helper()

	testCases := []struct {
		hasObject     bool
		expectedError string
	}{
		{
			hasObject:     false,
			expectedError: "",
		},
		{
			hasObject:     true,
			expectedError: podRunningErrorMsg,
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidPodTestBuilder(buildTestClientWithDummyPod())

		if testCase.hasObject {
			testBuilder.Object = testBuilder.Definition
			testBuilder.Object.Spec.NodeName = defaultPodNodeName
		}

		testBuilder = testFunc(testBuilder, toleration)
		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, []corev1.Toleration{toleration}, testBuilder.Definition.Spec.Tolerations)
		}
	}
}

// buildDummyPod returns a Pod with the provided name, nsname, and container image.
//
//nolint:unparam
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
