package pod

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var testUser = int64(1000)

func TestNewContainerBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		cmd           []string
		expectedError string
	}{
		{
			name:          "container",
			namespace:     "test-namespace",
			cmd:           []string{"/bin/bash", "-c", "sleep"},
			expectedError: "",
		},
		{
			name:          "",
			namespace:     "test-namespace",
			cmd:           []string{"/bin/bash", "-c", "sleep"},
			expectedError: "container's name is empty",
		},
		{
			name:          "container",
			namespace:     "",
			cmd:           []string{"/bin/bash", "-c", "sleep"},
			expectedError: "container's image is empty",
		},
		{
			name:          "container",
			namespace:     "test-namespace",
			cmd:           []string{},
			expectedError: "container's cmd is empty",
		},
	}
	for _, testCase := range testCases {
		container := NewContainerBuilder(testCase.name, testCase.namespace, testCase.cmd)
		assert.Equal(t, container.errorMsg, testCase.expectedError)

		if testCase.expectedError == "" {
			assert.NotNil(t, container.definition)
		}
	}
}

func TestPodContainerWithSecurityCapabilities(t *testing.T) {
	testCases := []struct {
		securityCapabilities []string
		redefine             bool
		expectedError        string
	}{
		{
			securityCapabilities: []string{"NET_RAW", "NET_ADMIN"},
			redefine:             false,
			expectedError:        "can not modify pre-existing security context",
		},
		{
			securityCapabilities: []string{"NET_RAW", "NET_ADMIN"},
			redefine:             true,
			expectedError:        "",
		},
		{
			securityCapabilities: []string{"NET_RAW", "NET_ADMIN", "invalid"},
			redefine:             true,
			expectedError: "one of the give securityCapabilities is invalid. " +
				"Please extend allowed list or fix parameter",
		},
	}

	for _, testCase := range testCases {
		container := NewContainerBuilder("container", "test", []string{"/bin/bash", "-c", "sleep"})
		container = container.WithSecurityCapabilities(testCase.securityCapabilities, testCase.redefine)
		assert.Equal(t, testCase.expectedError, container.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, container.definition)
		}
	}
}

func TestPodContainerWithDropSecurityCapabilities(t *testing.T) {
	testCases := []struct {
		securityCapabilities []string
		redefine             bool
		expectedError        string
	}{
		{
			securityCapabilities: []string{"NET_RAW", "NET_ADMIN"},
			redefine:             false,
			expectedError:        "",
		},
		{
			securityCapabilities: []string{},
			redefine:             true,
			expectedError:        "",
		},
		{
			securityCapabilities: []string{"NET_RAW", "NET_ADMIN", "invalid"},
			redefine:             true,
			expectedError: "one of the provided securityCapabilities is invalid. " +
				"Please extend the allowed list or fix parameter",
		},
	}

	for _, testCase := range testCases {
		container := NewContainerBuilder("container", "test", []string{"/bin/bash", "-c", "sleep"})
		container = container.WithDropSecurityCapabilities(testCase.securityCapabilities, testCase.redefine)
		assert.Equal(t, testCase.expectedError, container.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, container.definition)
		}
	}
}

func TestPodContainerWithSecurityContext(t *testing.T) {
	testCases := []struct {
		securityContext *corev1.SecurityContext
		expectedError   string
	}{
		{
			securityContext: &corev1.SecurityContext{RunAsUser: &testUser},
			expectedError:   "",
		},
		{
			securityContext: nil,
			expectedError:   "can not modify container config with empty securityContext",
		},
	}

	for _, testCase := range testCases {
		container := NewContainerBuilder("container", "test", []string{"/bin/bash", "-c", "sleep"})
		container = container.WithSecurityContext(testCase.securityContext)
		assert.Equal(t, testCase.expectedError, container.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, container.definition)
		}
	}
}

func TestPodContainerWithResourceLimit(t *testing.T) {
	testCases := []struct {
		hugepages     string
		memory        string
		cpu           int64
		expectedError string
	}{
		{
			hugepages:     "16",
			memory:        "24G",
			cpu:           10,
			expectedError: "",
		},
		{
			hugepages:     "",
			memory:        "24G",
			cpu:           10,
			expectedError: "container's resource limit 'hugePages' is empty",
		},
		{
			hugepages:     "16",
			memory:        "",
			cpu:           10,
			expectedError: "container's resource limit 'memory' is empty",
		},
		{
			hugepages:     "16",
			memory:        "24G",
			cpu:           0,
			expectedError: "container's resource limit 'cpu' is invalid",
		},
	}

	for _, testCase := range testCases {
		container := NewContainerBuilder("container", "test", []string{"/bin/bash", "-c", "sleep"})
		container = container.WithResourceLimit(testCase.hugepages, testCase.memory, testCase.cpu)
		assert.Equal(t, testCase.expectedError, container.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, container.definition)
		}
	}
}

func TestPodContainerWithResourceRequest(t *testing.T) {
	testCases := []struct {
		hugepages     string
		memory        string
		cpu           int64
		expectedError string
	}{
		{
			hugepages:     "16",
			memory:        "24G",
			cpu:           10,
			expectedError: "",
		},
		{
			hugepages:     "",
			memory:        "24G",
			cpu:           10,
			expectedError: "container's resource request 'hugePages' is empty",
		},
		{
			hugepages:     "16",
			memory:        "",
			cpu:           10,
			expectedError: "container's resource request 'memory' is empty",
		},
		{
			hugepages:     "16",
			memory:        "24G",
			cpu:           0,
			expectedError: "container's resource request 'cpu' is invalid",
		},
	}

	for _, testCase := range testCases {
		container := NewContainerBuilder("container", "test", []string{"/bin/bash", "-c", "sleep"})
		container = container.WithResourceRequest(testCase.hugepages, testCase.memory, testCase.cpu)
		assert.Equal(t, testCase.expectedError, container.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, container.definition)
		}
	}
}

func TestPodContainerWithCustomResourcesRequests(t *testing.T) {
	testCases := []struct {
		customResourcesRequests corev1.ResourceList
		expectedError           string
	}{
		{
			customResourcesRequests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("10")},
			expectedError:           "",
		},
		{
			customResourcesRequests: corev1.ResourceList{},
			expectedError:           "container's resource requests var 'resourceList' is empty",
		},
	}

	for _, testCase := range testCases {
		container := NewContainerBuilder("container", "test", []string{"/bin/bash", "-c", "sleep"})
		container = container.WithCustomResourcesRequests(testCase.customResourcesRequests)
		assert.Equal(t, testCase.expectedError, container.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, container.definition)
		}
	}
}

func TestPodContainerWithCustomResourcesLimits(t *testing.T) {
	testCases := []struct {
		customResourcesRequests corev1.ResourceList
		expectedError           string
	}{
		{
			customResourcesRequests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("10")},
			expectedError:           "",
		},
		{
			customResourcesRequests: corev1.ResourceList{},
			expectedError:           "container's resource limit var 'resourceList' is empty",
		},
	}

	for _, testCase := range testCases {
		container := NewContainerBuilder("container", "test", []string{"/bin/bash", "-c", "sleep"})
		container = container.WithCustomResourcesLimits(testCase.customResourcesRequests)
		assert.Equal(t, testCase.expectedError, container.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, container.definition)
		}
	}
}

func TestPodContainerWithImagePullPolicy(t *testing.T) {
	testCases := []struct {
		pullPolicy    corev1.PullPolicy
		expectedError string
	}{
		{
			pullPolicy:    corev1.PullAlways,
			expectedError: "",
		},
		{
			pullPolicy:    "container's pull policy var 'pullPolicy' is empty",
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		container := NewContainerBuilder("container", "test", []string{"/bin/bash", "-c", "sleep"})
		container = container.WithImagePullPolicy(testCase.pullPolicy)
		assert.Equal(t, testCase.expectedError, container.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, container.definition)
		}
	}
}

func TestPodContainerWithEnvVar(t *testing.T) {
	testCases := []struct {
		name          string
		value         string
		expectedError string
	}{
		{
			name:          "test",
			value:         "test",
			expectedError: "",
		},
		{
			name:          "",
			value:         "test",
			expectedError: "container's environment var 'name' is empty",
		},
		{
			name:          "test",
			value:         "",
			expectedError: "container's environment var 'value' is empty",
		},
	}

	for _, testCase := range testCases {
		container := NewContainerBuilder("container", "test", []string{"/bin/bash", "-c", "sleep"})
		container = container.WithEnvVar(testCase.name, testCase.value)
		assert.Equal(t, testCase.expectedError, container.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, container.definition)
		}
	}
}

func TestPodContainerWithVolumeMount(t *testing.T) {
	testCases := []struct {
		mount         corev1.VolumeMount
		expectedError string
	}{
		{
			mount:         corev1.VolumeMount{Name: "test", MountPath: "/tmp"},
			expectedError: "",
		},
		{
			mount:         corev1.VolumeMount{},
			expectedError: "container's volume mount path is empty",
		},
	}

	for _, testCase := range testCases {
		container := NewContainerBuilder("container", "test", []string{"/bin/bash", "-c", "sleep"})
		container = container.WithVolumeMount(testCase.mount)
		assert.Equal(t, testCase.expectedError, container.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, container.definition)
		}
	}
}

func TestPodContainerWithPorts(t *testing.T) {
	testCases := []struct {
		ports         []corev1.ContainerPort
		expectedError string
	}{
		{
			ports:         []corev1.ContainerPort{{Name: "test", Protocol: "TCP", ContainerPort: int32(5000)}},
			expectedError: "",
		},
		{
			ports:         []corev1.ContainerPort{},
			expectedError: "can not modify container config without any port",
		},
	}

	for _, testCase := range testCases {
		container := NewContainerBuilder("container", "test", []string{"/bin/bash", "-c", "sleep"})
		container = container.WithPorts(testCase.ports)
		assert.Equal(t, testCase.expectedError, container.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, container.definition)
		}
	}
}

func TestPodContainerWithReadinessProbe(t *testing.T) {
	testCases := []struct {
		readinessProbe *corev1.Probe
		expectedError  string
	}{
		{
			readinessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					Exec: &corev1.ExecAction{
						Command: []string{
							"echo",
							"ready",
						},
					},
				},
			},
			expectedError: "",
		},
		{
			readinessProbe: nil,
			expectedError:  "container's readinessProbe is empty",
		},
	}

	for _, testCase := range testCases {
		container := NewContainerBuilder("container", "test", []string{"/bin/bash", "-c", "sleep"})
		container = container.WithReadinessProbe(testCase.readinessProbe)
		assert.Equal(t, testCase.expectedError, container.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, container.definition)
			assert.Equal(t, testCase.readinessProbe, container.definition.ReadinessProbe)
		}
	}
}

func TestPodContainerWithTTY(t *testing.T) {
	testCases := []struct {
		enableTty     bool
		expectedError string
	}{
		{
			enableTty:     true,
			expectedError: "",
		},
		{
			enableTty:     false,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		container := NewContainerBuilder("container", "test", []string{"/bin/bash", "-c", "sleep"})
		container = container.WithTTY(testCase.enableTty)
		assert.Equal(t, testCase.expectedError, container.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, container.definition)
			assert.Equal(t, testCase.enableTty, container.definition.TTY)
		}
	}
}

func TestPodContainerWithStdin(t *testing.T) {
	testCases := []struct {
		enableStdin   bool
		expectedError string
	}{
		{
			enableStdin:   true,
			expectedError: "",
		},
		{
			enableStdin:   false,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		container := NewContainerBuilder("container", "test", []string{"/bin/bash", "-c", "sleep"})
		container = container.WithStdin(testCase.enableStdin)
		assert.Equal(t, testCase.expectedError, container.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, container.definition)
			assert.Equal(t, testCase.enableStdin, container.definition.Stdin)
		}
	}
}

func TestPodContainerGetContainerCfg(t *testing.T) {
	testCases := []struct {
		builder       *ContainerBuilder
		expectedError error
	}{
		{
			builder:       NewContainerBuilder("container", "test", []string{"/bin/bash", "-c", "sleep"}),
			expectedError: nil,
		},
		{
			builder:       NewContainerBuilder("", "test", []string{"/bin/bash", "-c", "sleep"}),
			expectedError: fmt.Errorf("container's name is empty"),
		},
		{
			builder:       NewContainerBuilder("container", "", []string{"/bin/bash", "-c", "sleep"}),
			expectedError: fmt.Errorf("container's image is empty"),
		},
		{
			builder: NewContainerBuilder("container", "test", []string{"/bin/bash", "-c", "sleep"}).
				WithEnvVar("", ""),
			expectedError: fmt.Errorf("container's environment var 'value' is empty"),
		},
	}

	for _, testCase := range testCases {
		container, err := testCase.builder.GetContainerCfg()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.NotNil(t, container)
		}
	}
}
