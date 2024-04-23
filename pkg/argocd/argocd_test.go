package argocd

import (
	"fmt"
	"testing"

	argocdoperatorv1alpha1 "github.com/argoproj-labs/argocd-operator/api/v1alpha1"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultArgoCdName   = "argocd"
	defaultArgoCdNSName = "test-namespace"
)

func TestArgoCdPull(t *testing.T) {
	generateArgoCd := func(name, namespace string) *argocdoperatorv1alpha1.ArgoCD {
		return &argocdoperatorv1alpha1.ArgoCD{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: argocdoperatorv1alpha1.ArgoCDSpec{},
		}
	}

	testCases := []struct {
		name                string
		namespace           string
		addToRuntimeObjects bool
		expectedError       error
		client              bool
	}{
		{
			name:                "argocdtest",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("argocd 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                "argocdtest",
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("argocd 'namespace' cannot be empty"),
			client:              true,
		},
		{
			name:                "argocdtest",
			namespace:           "test-namespace",
			addToRuntimeObjects: false,
			expectedError:       fmt.Errorf("argocd object argocdtest doesn't exist in namespace test-namespace"),
			client:              true,
		},
		{
			name:                "argocdtest",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("argocd 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testArgoCd := generateArgoCd(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testArgoCd)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})
		}

		builderResult, err := Pull(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, builderResult.Object.Name)
			assert.Equal(t, testCase.namespace, builderResult.Object.Namespace)
		}
	}
}

func TestArgoCdNewBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		expectedError string
	}{
		{
			name:          "argocdtest",
			namespace:     "test-namespace",
			expectedError: "",
		},
		{
			name:          "",
			namespace:     "test-namespace",
			expectedError: "argocd 'name' cannot be empty",
		},
		{
			name:          "argocd",
			namespace:     "",
			expectedError: "argocd 'nsname' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})
		testArgoCdBuilder := NewBuilder(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, testArgoCdBuilder.errorMsg)
		assert.NotNil(t, testArgoCdBuilder.Definition)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.name, testArgoCdBuilder.Definition.Name)
			assert.Equal(t, testCase.namespace, testArgoCdBuilder.Definition.Namespace)
		}
	}
}

func TestArgoCdGet(t *testing.T) {
	testCases := []struct {
		testArgoCd    *Builder
		expectedError error
	}{
		{
			testArgoCd:    buildValidArgoCdBuilder(buildArgoCdTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testArgoCd:    buildInValidArgoCdBuilder(buildArgoCdTestClientWithDummyObject()),
			expectedError: fmt.Errorf("argocd 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		argoCd, err := testCase.testArgoCd.Get()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, argoCd.Name, testCase.testArgoCd.Definition.Name)
			assert.Equal(t, argoCd.Namespace, testCase.testArgoCd.Definition.Namespace)
		}
	}
}

func TestArgoCdExist(t *testing.T) {
	testCases := []struct {
		testArgoCd     *Builder
		expectedStatus bool
	}{
		{
			testArgoCd:     buildValidArgoCdBuilder(buildArgoCdTestClientWithDummyObject()),
			expectedStatus: true,
		},
		{
			testArgoCd:     buildInValidArgoCdBuilder(buildArgoCdTestClientWithDummyObject()),
			expectedStatus: false,
		},
		{
			testArgoCd:     buildInValidArgoCdBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testArgoCd.Exists()
		assert.Equal(t, exist, testCase.expectedStatus)
	}
}

func TestArgoCdCreate(t *testing.T) {
	testCases := []struct {
		testArgoCd    *Builder
		expectedError error
	}{
		{
			testArgoCd:    buildValidArgoCdBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
		{
			testArgoCd:    buildInValidArgoCdBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("argocd 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		testArgoCdBuilder, err := testCase.testArgoCd.Create()
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil {
			assert.Equal(t, testArgoCdBuilder.Definition, testArgoCdBuilder.Object)
		}
	}
}

func TestArgoCdDelete(t *testing.T) {
	testCases := []struct {
		testArgoCd    *Builder
		expectedError error
	}{
		{
			testArgoCd:    buildValidArgoCdBuilder(buildArgoCdTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testArgoCd:    buildInValidArgoCdBuilder(buildArgoCdTestClientWithDummyObject()),
			expectedError: fmt.Errorf("argocd 'nsname' cannot be empty"),
		},
		{
			testArgoCd:    buildValidArgoCdBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		_, err := testCase.testArgoCd.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testArgoCd.Object)
		}
	}
}

func TestArgoCdUpdate(t *testing.T) {
	testCases := []struct {
		testArgoCd    *Builder
		expectedError error
		image         string
	}{
		{
			testArgoCd:    buildValidArgoCdBuilder(buildArgoCdTestClientWithDummyObject()),
			expectedError: nil,
			image:         "testimage",
		},
		{
			testArgoCd:    buildInValidArgoCdBuilder(buildArgoCdTestClientWithDummyObject()),
			expectedError: fmt.Errorf("argocd 'nsname' cannot be empty"),
			image:         "testimage",
		},
		{
			testArgoCd:    buildValidArgoCdBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("argocds.argoproj.io \"argocd\" not found"),
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, "", testCase.testArgoCd.Definition.Spec.Image)
		testCase.testArgoCd.Definition.ResourceVersion = "999"
		assert.Equal(t, "", testCase.testArgoCd.Definition.Spec.Image)
		testCase.testArgoCd.Definition.Spec.Image = testCase.image
		_, err := testCase.testArgoCd.Update(false)

		if errors.IsNotFound(err) {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedError, err)
		}

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.image, testCase.testArgoCd.Object.Spec.Image)
			assert.Equal(t, testCase.testArgoCd.Object, testCase.testArgoCd.Definition)
		}
	}
}

func buildValidArgoCdBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, defaultArgoCdName, defaultArgoCdNSName)
}

func buildInValidArgoCdBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, defaultArgoCdName, "")
}

func buildArgoCdTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: buildDummyArgoCd(),
	})
}

func buildDummyArgoCd() []runtime.Object {
	return append([]runtime.Object{}, &argocdoperatorv1alpha1.ArgoCD{
		ObjectMeta: metav1.ObjectMeta{
			Name:            defaultArgoCdName,
			Namespace:       defaultArgoCdNSName,
			ResourceVersion: "999",
		},

		Spec: argocdoperatorv1alpha1.ArgoCDSpec{
			Image: "",
		},
	})
}
