package configmap

import (
	"errors"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func TestNewBuilder(t *testing.T) {
	testCases := []struct {
		name        string
		nsname      string
		expectedCM  *corev1.ConfigMap
		expectedErr string
	}{
		{
			name:   "test",
			nsname: "testns",
			expectedCM: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "testns",
				},
			},
			expectedErr: "",
		},
		{
			name:        "",
			nsname:      "testns",
			expectedCM:  nil,
			expectedErr: "configmap 'name' cannot be empty",
		},
		{
			name:        "test",
			nsname:      "",
			expectedCM:  nil,
			expectedErr: "configmap 'nsname' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})

		testBuilder := NewBuilder(testSettings, testCase.name, testCase.nsname)

		if testCase.expectedErr == "" {
			assert.NotNil(t, testBuilder)
			assert.Equal(t, testCase.expectedCM.Name, testBuilder.Definition.Name)
			assert.Equal(t, testCase.expectedCM.Namespace, testBuilder.Definition.Namespace)
		} else {
			assert.Equal(t, testCase.expectedErr, testBuilder.errorMsg)
		}
	}
}

func TestPull(t *testing.T) {
	testCases := []struct {
		name                string
		nsname              string
		expectedError       bool
		addToRuntimeObjects bool
		expectedErrorText   string
	}{
		{
			name:                "test",
			nsname:              "testns",
			expectedError:       false,
			addToRuntimeObjects: true,
		},
		{
			name:                "",
			nsname:              "testns",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "configmap 'name' cannot be empty",
		},
		{
			name:                "test",
			nsname:              "",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "configmap 'nsname' cannot be empty",
		},
		{
			name:                "test",
			nsname:              "testns",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "configmap object test does not exist in namespace testns",
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testCM := generateConfigMap(testCase.name, testCase.nsname)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testCM)
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		// Test the Pull function
		builderResult, err := Pull(testSettings, testCase.name, testCase.nsname)

		// Check the error
		if testCase.expectedError {
			assert.NotNil(t, err)

			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.NotNil(t, builderResult)
			assert.Equal(t, testCase.name, builderResult.Definition.Name)
			assert.Equal(t, testCase.nsname, builderResult.Definition.Namespace)
		}
	}
}

func TestCreate(t *testing.T) {
	testCases := []struct {
		addToRuntimeObjects bool
	}{
		{
			addToRuntimeObjects: true,
		},
		{
			addToRuntimeObjects: false,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		testCM := generateConfigMap("test-name", "test-namespace")

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testCM)
		}

		testBuilder := buildTestBuilderWithFakeObjects(runtimeObjects)

		// Test the Create function
		builderResult, err := testBuilder.Create()
		assert.Nil(t, err)
		assert.NotNil(t, builderResult)
		assert.Equal(t, "test-name", builderResult.Definition.Name)
		assert.Equal(t, "test-namespace", builderResult.Definition.Namespace)
	}
}

func TestDelete(t *testing.T) {
	testCases := []struct {
		addToRuntimeObjects bool
	}{
		{
			addToRuntimeObjects: true,
		},
		{
			addToRuntimeObjects: false,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		testCM := generateConfigMap("test-name", "test-namespace")

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testCM)
		}

		testBuilder := buildTestBuilderWithFakeObjects(runtimeObjects)

		// Test the Delete function
		err := testBuilder.Delete()
		assert.Nil(t, err)
	}
}

func TestUpdate(t *testing.T) {
	generateTestConfigMap := func() *corev1.ConfigMap {
		return &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-name",
				Namespace: "test-namespace",
			},
		}
	}

	testCases := []struct {
		configMapExistsAlready bool
	}{
		{
			configMapExistsAlready: false,
		},
		{
			configMapExistsAlready: true,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		if testCase.configMapExistsAlready {
			runtimeObjects = append(runtimeObjects, generateTestConfigMap())
		}

		testBuilder := buildTestBuilderWithFakeObjects(runtimeObjects)

		// Assert the deployment before the update
		assert.NotNil(t, testBuilder.Definition)
		assert.Nil(t, testBuilder.Definition.Data)

		// Set a value in the definition to test the update
		testBuilder.Definition.Data = map[string]string{"key1": "value1", "key2": "value2"}

		// Perform the update
		result, err := testBuilder.Update()

		// Assert the result
		assert.NotNil(t, testBuilder.Definition)

		if !testCase.configMapExistsAlready {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testBuilder.Definition.Name, result.Definition.Name)
			assert.Equal(t, testBuilder.Definition.Data, result.Definition.Data)
		}
	}
}

func TestGetGVR(t *testing.T) {
	testGVR := GetGVR()
	assert.Equal(t, "configmaps", testGVR.Resource)
	assert.Equal(t, "v1", testGVR.Version)
	assert.Equal(t, "", testGVR.Group)
}

func TestValidate(t *testing.T) {
	testCases := []struct {
		builderNil    bool
		definitionNil bool
		apiClientNil  bool
		expectedError string
	}{
		{
			builderNil:    true,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: "error: received nil ConfigMap builder",
		},
		{
			builderNil:    false,
			definitionNil: true,
			apiClientNil:  false,
			expectedError: "can not redefine the undefined ConfigMap",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  true,
			expectedError: "ConfigMap builder cannot have nil apiClient",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildTestBuilderWithFakeObjects([]runtime.Object{})

		if testCase.builderNil {
			testBuilder = nil
		}

		if testCase.definitionNil {
			testBuilder.Definition = nil
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

func TestWithOptions(t *testing.T) {
	testBuilder := buildTestBuilderWithFakeObjects([]runtime.Object{})

	testBuilder.WithOptions(func(builder *Builder) (*Builder, error) {
		return builder, nil
	})

	assert.Equal(t, "", testBuilder.errorMsg)

	testBuilder.WithOptions(func(builder *Builder) (*Builder, error) {
		return builder, errors.New("error")
	})

	assert.Equal(t, "error", testBuilder.errorMsg)
}

func TestWithData(t *testing.T) {
	testCases := []struct {
		key         string
		value       string
		expectedErr string
	}{
		{
			key:         "key",
			value:       "value",
			expectedErr: "",
		},
		{
			key:         "",
			value:       "",
			expectedErr: "'data' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildTestBuilderWithFakeObjects([]runtime.Object{})

		if testCase.expectedErr == "" {
			testBuilder.WithData(map[string]string{testCase.key: testCase.value})

			assert.Equal(t, testCase.value, testBuilder.Definition.Data[testCase.key])
		} else {
			testBuilder.WithData(map[string]string{})

			assert.Equal(t, testCase.expectedErr, testBuilder.errorMsg)
		}
	}
}

func buildTestBuilderWithFakeObjects(objects []runtime.Object) *Builder {
	fakeClient := k8sfake.NewSimpleClientset(objects...)

	return NewBuilder(&clients.Settings{
		CoreV1Interface: fakeClient.CoreV1(),
		K8sClient:       fakeClient,
	}, "test-name", "test-namespace")
}

func generateConfigMap(name, nsname string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsname,
		},
	}
}
