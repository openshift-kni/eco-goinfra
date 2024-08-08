package pod

import (
	"testing"
	"time"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestWaitForAllPodsInNamespacesHealthy(t *testing.T) {
	testCases := []struct {
		namespaces               []string
		includeSucceeded         bool
		skipRedinessCheck        bool
		ignoreRestartPolicyNever bool
		ignoreNamespaces         []string
		expectedErrMsg           string
	}{
		{
			namespaces:               []string{"ns1"},
			includeSucceeded:         true,
			skipRedinessCheck:        true,
			ignoreRestartPolicyNever: true,
			ignoreNamespaces:         []string{},
			expectedErrMsg:           "",
		},
		{
			namespaces:               []string{"ns1"},
			includeSucceeded:         true,
			skipRedinessCheck:        false,
			ignoreRestartPolicyNever: true,
			ignoreNamespaces:         []string{},
			expectedErrMsg:           "",
		},
		{
			namespaces:               []string{"ns2"},
			includeSucceeded:         true,
			skipRedinessCheck:        false,
			ignoreRestartPolicyNever: true,
			ignoreNamespaces:         []string{},
			expectedErrMsg:           "context deadline exceeded",
		},
		{
			namespaces:               []string{},
			includeSucceeded:         true,
			skipRedinessCheck:        false,
			ignoreRestartPolicyNever: true,
			ignoreNamespaces:         []string{},
			expectedErrMsg:           "context deadline exceeded",
		},
		{
			namespaces:               []string{},
			includeSucceeded:         true,
			skipRedinessCheck:        false,
			ignoreRestartPolicyNever: true,
			ignoreNamespaces:         []string{"ns2"},
			expectedErrMsg:           "context deadline exceeded",
		},
	}

	var runtimeObjects []runtime.Object
	runtimeObjects = append(runtimeObjects, generateTestPod("test1", "ns1", corev1.PodRunning, corev1.PodReady, false))
	runtimeObjects = append(runtimeObjects, generateTestPod("test2", "ns1", corev1.PodRunning, corev1.PodReady, false))
	runtimeObjects = append(runtimeObjects, generateTestPod("test3", "ns1", corev1.PodRunning, corev1.PodReady, false))
	runtimeObjects = append(runtimeObjects, generateTestPod("test4", "ns1", corev1.PodRunning, corev1.PodReady, false))
	runtimeObjects = append(runtimeObjects, generateTestPod("test5", "ns1", corev1.PodRunning, corev1.PodReady, false))
	runtimeObjects = append(runtimeObjects, generateTestPod("test1", "ns2", corev1.PodRunning, corev1.PodReady, false))
	runtimeObjects = append(runtimeObjects, generateTestPod("test2", "ns2", corev1.PodRunning, corev1.PodInitialized,
		false))

	testSettings := clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: runtimeObjects,
	})

	for _, testCase := range testCases {
		err := WaitForAllPodsInNamespacesHealthy(testSettings, testCase.namespaces, 2*time.Second, testCase.includeSucceeded,
			testCase.skipRedinessCheck,
			testCase.ignoreRestartPolicyNever, testCase.ignoreNamespaces)

		assert.Equal(t, testCase.expectedErrMsg, getErrorString(err))
	}
}
