package service

import (
	"errors"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestList(t *testing.T) {
	generateService := func(name, namespace string) *corev1.Service {
		return &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				Labels:    map[string]string{"demo": name},
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{
					"app": name,
				},
			},
		}
	}

	testCases := []struct {
		serviceExists       bool
		testNamespace       string
		listOptions         []metav1.ListOptions
		expetedError        error
		expectedNumServices int
	}{
		{ // Service does exist
			serviceExists:       true,
			testNamespace:       "test-namespace",
			expetedError:        nil,
			listOptions:         []metav1.ListOptions{},
			expectedNumServices: 1,
		},
		{ // Service does not exist
			serviceExists:       false,
			testNamespace:       "test-namespace",
			expetedError:        nil,
			listOptions:         []metav1.ListOptions{},
			expectedNumServices: 0,
		},
		{ // Missing namespace parameter
			serviceExists:       true,
			testNamespace:       "",
			expetedError:        errors.New("failed to list services, 'nsname' parameter is empty"),
			listOptions:         []metav1.ListOptions{},
			expectedNumServices: 0,
		},
		{ // More than one ListOptions was passed
			serviceExists:       true,
			testNamespace:       "test-namespace",
			expetedError:        errors.New("error: more than one ListOptions was passed"),
			listOptions:         []metav1.ListOptions{{}, {}},
			expectedNumServices: 0,
		},
		{ // Valid number of list options
			serviceExists:       true,
			testNamespace:       "test-namespace",
			expetedError:        nil,
			listOptions:         []metav1.ListOptions{{}},
			expectedNumServices: 1,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		if testCase.serviceExists {
			runtimeObjects = append(runtimeObjects, generateService("test-service", "test-namespace"))
		}

		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		networkPolicyList, err := List(testSettings, testCase.testNamespace, testCase.listOptions...)
		assert.Equal(t, testCase.expetedError, err)
		assert.Equal(t, testCase.expectedNumServices, len(networkPolicyList))
	}
}
