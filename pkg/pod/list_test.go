package pod

import (
	"context"
	"testing"
	"time"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestWaitForAllPodsInNamespacesHealthy(t *testing.T) {
	generateTestPod := func(namespace string, conditionType corev1.PodConditionType) *corev1.Pod {
		pod := buildDummyPodWithPhaseAndCondition(corev1.PodRunning, conditionType, false)
		pod.Namespace = namespace

		return pod
	}

	testCases := []struct {
		namespaces               []string
		includeSucceeded         bool
		skipRedinessCheck        bool
		ignoreRestartPolicyNever bool
		ignoreNamespaces         []string
		pods                     []runtime.Object
		expectedError            error
	}{
		{
			namespaces:               []string{"ns1"},
			includeSucceeded:         true,
			skipRedinessCheck:        true,
			ignoreRestartPolicyNever: true,
			ignoreNamespaces:         []string{},
			pods: []runtime.Object{
				generateTestPod("ns1", corev1.PodReady), generateTestPod("ns2", corev1.PodInitialized)},
			expectedError: nil,
		},
		{
			namespaces:               []string{"ns1"},
			includeSucceeded:         true,
			skipRedinessCheck:        false,
			ignoreRestartPolicyNever: true,
			ignoreNamespaces:         []string{},
			pods:                     []runtime.Object{generateTestPod("ns1", corev1.PodReady)},
			expectedError:            nil,
		},
		{
			namespaces:               []string{"ns2"},
			includeSucceeded:         true,
			skipRedinessCheck:        false,
			ignoreRestartPolicyNever: true,
			ignoreNamespaces:         []string{},
			pods:                     []runtime.Object{generateTestPod("ns2", corev1.PodInitialized)},
			expectedError:            context.DeadlineExceeded,
		},
		{
			namespaces:               []string{},
			includeSucceeded:         true,
			skipRedinessCheck:        false,
			ignoreRestartPolicyNever: true,
			ignoreNamespaces:         []string{},
			pods: []runtime.Object{
				generateTestPod("ns1", corev1.PodReady), generateTestPod("ns2", corev1.PodInitialized)},
			expectedError: context.DeadlineExceeded,
		},
		{
			namespaces:               []string{},
			includeSucceeded:         true,
			skipRedinessCheck:        false,
			ignoreRestartPolicyNever: true,
			ignoreNamespaces:         []string{"ns2"},
			pods:                     []runtime.Object{generateTestPod("ns1", corev1.PodReady)},
			expectedError:            nil,
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: testCase.pods,
		})

		err := WaitForAllPodsInNamespacesHealthy(
			testSettings,
			testCase.namespaces,
			time.Second,
			testCase.includeSucceeded,
			testCase.skipRedinessCheck,
			testCase.ignoreRestartPolicyNever,
			testCase.ignoreNamespaces)

		assert.Equal(t, testCase.expectedError, err)
	}
}
