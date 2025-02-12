package pod

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func TestWaitForPodsInNamespacesHealthy(t *testing.T) {
	testCases := []struct {
		namespaces    []string
		listOptions   []metav1.ListOptions
		healthy       bool
		failed        bool
		client        bool
		expectedError error
	}{
		{
			namespaces:    nil,
			listOptions:   nil,
			healthy:       true,
			failed:        false,
			client:        true,
			expectedError: nil,
		},
		{
			namespaces:    nil,
			listOptions:   nil,
			healthy:       false,
			failed:        true,
			client:        true,
			expectedError: nil,
		},
		{
			namespaces:    nil,
			listOptions:   []metav1.ListOptions{{}, {}},
			healthy:       true,
			failed:        false,
			client:        true,
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
		},
		{
			namespaces:    nil,
			listOptions:   nil,
			healthy:       true,
			failed:        false,
			client:        false,
			expectedError: fmt.Errorf("podList 'apiClient' cannot be empty"),
		},
		{
			namespaces:    nil,
			listOptions:   nil,
			healthy:       false,
			failed:        false,
			client:        true,
			expectedError: context.DeadlineExceeded,
		},
		{
			namespaces:    []string{defaultPodNsName + "-has-no-pods"},
			listOptions:   nil,
			healthy:       false,
			failed:        false,
			client:        true,
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testPod := buildDummyPod(defaultPodName, defaultPodNsName, defaultPodImage)
			if testCase.healthy {
				testPod = buildDummyPodWithPhaseAndCondition(corev1.PodSucceeded, corev1.PodReady, false)
			} else if testCase.failed {
				testPod = buildDummyPodWithPhaseAndCondition(corev1.PodFailed, corev1.PodReady, true)
			}

			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: []runtime.Object{testPod},
			})
		}

		err := WaitForPodsInNamespacesHealthy(testSettings, testCase.namespaces, time.Second, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)
	}
}
