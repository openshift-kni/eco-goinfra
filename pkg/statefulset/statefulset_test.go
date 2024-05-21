package statefulset

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

//nolint:funlen
func TestListInAllNamespaces(t *testing.T) {
	int32Ptr := func(i int32) *int32 { return &i }

	generateStatefulSet := func(name, namespace string) *appsv1.StatefulSet {
		return &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				Labels:    map[string]string{"demo": name},
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: int32Ptr(1),
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": name,
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": name,
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "test-container",
								Image: "test-image",
							},
						},
					},
				},
			},
		}
	}

	testCases := []struct {
		statefulsetMap         map[string][]string
		addToRuntimeObjects    bool
		expectedCount          int
		statefulsetListOptions []metav1.ListOptions
		expectedError          bool
		expectedErrorMsg       string
		debug                  bool
	}{
		{
			statefulsetMap: map[string][]string{
				"one":   {"one-1", "one-2", "one-3"},
				"two":   {"two-1", "two-2"},
				"three": {"three-1"},
			},
			addToRuntimeObjects:    true,
			expectedCount:          int(6),
			statefulsetListOptions: []metav1.ListOptions{},
			expectedError:          false,
			expectedErrorMsg:       "",
			debug:                  false,
		},
		{
			statefulsetMap: map[string][]string{
				"one": {"one-1"},
			},
			addToRuntimeObjects:    true,
			expectedCount:          int(1),
			statefulsetListOptions: []metav1.ListOptions{},
			expectedError:          false,
			expectedErrorMsg:       "",
			debug:                  false,
		},
		{
			statefulsetMap: map[string][]string{
				"one":   {"one-1", "one-2", "one-3"},
				"two":   {"one-1", "two-2"},
				"three": {"one-1"},
			},
			addToRuntimeObjects: true,
			expectedCount:       int(3),
			statefulsetListOptions: []metav1.ListOptions{
				{
					LabelSelector: "demo=one-1",
				},
			},
			expectedError:    false,
			expectedErrorMsg: "",
			debug:            true,
		},
		{
			statefulsetMap: map[string][]string{
				"one":   {"one-1", "one-2", "one-3"},
				"two":   {"one-1", "two-2"},
				"three": {"one-1"},
			},
			addToRuntimeObjects: true,
			expectedCount:       int(0),
			statefulsetListOptions: []metav1.ListOptions{
				{
					LabelSelector: "fake=fake",
				},
			},
			expectedError:    false,
			expectedErrorMsg: "",
			debug:            false,
		},
		{
			statefulsetMap: map[string][]string{
				"one":   {"one-1", "one-2", "one-3"},
				"two":   {"one-1", "two-2"},
				"three": {"one-1"},
			},
			addToRuntimeObjects: true,
			expectedCount:       int(0),
			statefulsetListOptions: []metav1.ListOptions{
				{
					LabelSelector: "fake=fake",
				},
				{
					FieldSelector: "testfield=testvalue",
				},
			},
			expectedError:    true,
			expectedErrorMsg: "error: more than one ListOptions was passed",
			debug:            false,
		},
		{
			statefulsetMap: map[string][]string{
				"one":   {"one-1", "one-2", "one-3"},
				"two":   {"two-1", "two-2"},
				"three": {"three-1"},
			},
			addToRuntimeObjects:    false,
			expectedCount:          int(0),
			statefulsetListOptions: []metav1.ListOptions{},
			expectedError:          false,
			expectedErrorMsg:       "",
			debug:                  false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		for nsName, value := range testCase.statefulsetMap {
			for _, stName := range value {
				testStatefulset := generateStatefulSet(stName, nsName)

				if testCase.addToRuntimeObjects {
					runtimeObjects = append(runtimeObjects, testStatefulset)
				}
			}
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		resStatefulSet, err := ListInAllNamespaces(testSettings, testCase.statefulsetListOptions...)

		if testCase.expectedError {
			assert.NotNil(t, err)
			assert.Equal(t, testCase.expectedErrorMsg, err.Error())
		} else {
			if testCase.debug {
				for _, k := range resStatefulSet {
					fmt.Printf("Namespace: %q\tName: %q\n", k.Definition.Namespace, k.Definition.Name)
				}
			}

			assert.Nil(t, err)
			assert.Equal(t, testCase.expectedCount, len(resStatefulSet))
		}
	}
}
