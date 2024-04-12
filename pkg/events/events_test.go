package events

import (
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func TestPull(t *testing.T) {
	generateEvent := func(name, nsname string) *corev1.Event {
		return &corev1.Event{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
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
		{
			name:                "test-event",
			namespace:           "test-namespace",
			expectedError:       false,
			addToRuntimeObjects: true,
			expectedErrorText:   "",
		},
		{
			name:                "test-event",
			namespace:           "test-namespace",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "event object test-event doesn't exist in namespace test-namespace",
		},
		{
			name:                "",
			namespace:           "test-namespace",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "event 'name' cannot be empty",
		},
		{
			name:                "test-event",
			namespace:           "",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "event 'nsname' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testEvent := generateEvent(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testEvent)
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		result, err := Pull(testSettings, testCase.name, testCase.namespace)

		if testCase.expectedError {
			assert.NotNil(t, err)

			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testEvent.Name, result.Object.Name)
			assert.Equal(t, testEvent.Namespace, result.Object.Namespace)
		}
	}
}

func TestValidate(t *testing.T) {
	testCases := []struct {
		builderNil    bool
		apiClientNil  bool
		expectedError string
	}{
		{
			builderNil:    true,
			apiClientNil:  false,
			expectedError: "error: received nil Event builder",
		},
		{
			builderNil:    false,
			apiClientNil:  true,
			expectedError: "Event builder cannot have nil apiClient",
		},
		{
			builderNil:    false,
			apiClientNil:  false,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestBuilder()

		if testCase.builderNil {
			testBuilder = nil
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

func buildValidTestBuilder() *Builder {
	return &Builder{
		apiClient: k8sfake.NewSimpleClientset().CoreV1().Events("test-namespace"),
		Object: &corev1.Event{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-event",
				Namespace: "test-namespace",
			},
		},
	}
}
