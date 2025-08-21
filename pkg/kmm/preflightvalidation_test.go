package kmm

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	kmmv1beta2 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/kmm/v1beta2"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	testSchemes = []clients.SchemeAttacher{
		kmmv1beta2.AddToScheme,
	}
	defaultPreflightName      = "preflight"
	defaultPreflightNamespace = "preflightns"
)

func TestNewPreflightValidationBuilder(t *testing.T) {
	testCases := []struct {
		name              string
		namespace         string
		expectedPreflight *kmmv1beta2.PreflightValidation
		expectedErr       string
	}{
		{
			name:      defaultPreflightName,
			namespace: defaultPreflightNamespace,
			expectedPreflight: &kmmv1beta2.PreflightValidation{
				ObjectMeta: metav1.ObjectMeta{
					Name:      defaultPreflightName,
					Namespace: defaultPreflightNamespace,
				},
			},
			expectedErr: "",
		},
		{
			name:              defaultPreflightName,
			namespace:         "",
			expectedPreflight: nil,
			expectedErr:       "PreflightValidation 'nsname' cannot be empty",
		},
		{
			name:              "",
			namespace:         defaultPreflightNamespace,
			expectedPreflight: nil,
			expectedErr:       "PreflightValidation 'name' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})

		testBuilder := NewPreflightValidationBuilder(testSettings, testCase.name, testCase.namespace)

		if testCase.expectedErr == "" {
			assert.NotNil(t, testBuilder)
			assert.Equal(t, testCase.expectedPreflight.Name, testBuilder.Definition.Name)
			assert.Equal(t, testCase.expectedPreflight.Namespace, testBuilder.Definition.Namespace)
		} else {
			assert.Equal(t, testCase.expectedErr, testBuilder.errorMsg)
		}
	}
}

func TestPreflightValidationWithKernelVersion(t *testing.T) {
	testCases := []struct {
		kernelVersion string
		expectedErr   string
	}{
		{
			kernelVersion: "some-kernel",
			expectedErr:   "",
		},
		{
			kernelVersion: "",
			expectedErr:   "invalid 'kernelVersion' argument can not be nil",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestPreflight(buildPreflightTestClientWithDummyObject())
		testBuilder.WithKernelVersion(testCase.kernelVersion)

		if testCase.expectedErr == "" {
			assert.Equal(t, testBuilder.Definition.Spec.KernelVersion, testCase.kernelVersion)
		} else {
			assert.Equal(t, testCase.expectedErr, testBuilder.errorMsg)
		}
	}
}

func TestPreflightValidationWithPushBuiltImage(t *testing.T) {
	testCases := []struct {
		push bool
	}{
		{
			push: true,
		},
		{
			push: false,
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestPreflight(buildPreflightTestClientWithDummyObject())
		testBuilder.WithPushBuiltImage(testCase.push)
		assert.Equal(t, testCase.push, testBuilder.Definition.Spec.PushBuiltImage)
	}
}

func TestPreflightValidationWithOptions(t *testing.T) {
	testBuilder := buildValidTestPreflight(buildPreflightTestClientWithDummyObject())

	testBuilder.WithOptions(func(builder *PreflightValidationBuilder) (*PreflightValidationBuilder, error) {
		return builder, nil
	})

	assert.Equal(t, "", testBuilder.errorMsg)

	testBuilder.WithOptions(func(builder *PreflightValidationBuilder) (*PreflightValidationBuilder, error) {
		return builder, fmt.Errorf("error")
	})

	assert.Equal(t, "error", testBuilder.errorMsg)
}

func TestPullPreflightValidation(t *testing.T) {
	testCases := []struct {
		name                string
		namespace           string
		expectedError       error
		addToRuntimeObjects bool
		client              bool
	}{
		{
			name:                "test",
			namespace:           "testns",
			expectedError:       nil,
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			name:                "",
			namespace:           "testns",
			expectedError:       fmt.Errorf("preflightvalidation 'name' cannot be empty"),
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			name:                "test",
			namespace:           "",
			expectedError:       fmt.Errorf("preflightvalidation 'nsname' cannot be empty"),
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			name:                "test",
			namespace:           "testns",
			expectedError:       fmt.Errorf("preflightvalidation object test doesn't exist in namespace testns"),
			addToRuntimeObjects: false,
			client:              true,
		},
		{
			name:                "test",
			namespace:           "testns",
			expectedError:       fmt.Errorf("preflightvalidation 'apiClient' cannot be empty"),
			addToRuntimeObjects: true,
			client:              false,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testPreflight := generatePreflight(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testPreflight)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		builderResult, err := PullPreflightValidation(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, builderResult.Definition.Name)
			assert.Equal(t, testCase.namespace, builderResult.Definition.Namespace)
		}
	}
}

func TestPreflightValidationCreate(t *testing.T) {
	testCases := []struct {
		testPreflight *PreflightValidationBuilder
		expectedError string
	}{
		{
			testPreflight: buildValidTestPreflight(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: "resourceVersion can not be set for Create requests",
		},
		{
			testPreflight: buildInValidTestPreflight(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: "PreflightValidation 'nsname' cannot be empty",
		},
		{
			testPreflight: buildValidTestPreflight(buildPreflightTestClientWithDummyObject()),
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testPreflightBuilder, err := testCase.testPreflight.Create()

		if testCase.expectedError == "" {
			assert.Equal(t, testPreflightBuilder.Definition, testPreflightBuilder.Object)
		} else {
			assert.Equal(t, testCase.expectedError, err.Error())
		}
	}
}

func TestPreflightValidationUpdate(t *testing.T) {
	testCases := []struct {
		testPreflight *PreflightValidationBuilder
		expectedError error
		kernelVersion string
	}{
		{
			testPreflight: buildInValidTestPreflight(buildPreflightTestClientWithDummyObject()),
			expectedError: fmt.Errorf("PreflightValidation 'nsname' cannot be empty"),
			kernelVersion: "testkernel",
		},
		{
			testPreflight: buildValidTestPreflight(buildPreflightTestClientWithDummyObject()),
			expectedError: nil,
			kernelVersion: "testkernel",
		},
		{
			testPreflight: buildValidTestPreflight(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("preflightvalidations.kmm.sigs.x-k8s.io \"preflight\" not found"),
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, "", testCase.testPreflight.Definition.Spec.KernelVersion)
		testCase.testPreflight.Definition.ResourceVersion = "999"
		assert.Equal(t, "", testCase.testPreflight.Definition.Spec.KernelVersion)
		testCase.testPreflight.Definition.Spec.KernelVersion = testCase.kernelVersion
		_, err := testCase.testPreflight.Update()

		if errors.IsNotFound(err) {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedError, err)
		}

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.kernelVersion, testCase.testPreflight.Object.Spec.KernelVersion)
			assert.Equal(t, testCase.testPreflight.Object, testCase.testPreflight.Definition)
		}
	}
}

func TestPreflightValidationExists(t *testing.T) {
	testCases := []struct {
		testPreflight  *PreflightValidationBuilder
		expectedStatus bool
	}{
		{
			testPreflight:  buildValidTestPreflight(buildPreflightTestClientWithDummyObject()),
			expectedStatus: true,
		},
		{
			testPreflight:  buildInValidTestPreflight(buildPreflightTestClientWithDummyObject()),
			expectedStatus: false,
		},
		{
			testPreflight:  buildInValidTestPreflight(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testPreflight.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestPreflightValidationDelete(t *testing.T) {
	testCases := []struct {
		testPreflight *PreflightValidationBuilder
		expectedError error
	}{
		{
			testPreflight: buildValidTestPreflight(buildPreflightTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testPreflight: buildInValidTestPreflight(buildPreflightTestClientWithDummyObject()),
			expectedError: fmt.Errorf("PreflightValidation 'nsname' cannot be empty"),
		},
		{
			testPreflight: buildValidTestPreflight(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		_, err := testCase.testPreflight.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testPreflight.Object)
		}
	}
}

func TestPreflightValidationGet(t *testing.T) {
	testCases := []struct {
		testPreflight *PreflightValidationBuilder
		expectedError error
	}{
		{
			testPreflight: buildValidTestPreflight(buildPreflightTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testPreflight: buildInValidTestPreflight(buildPreflightTestClientWithDummyObject()),
			expectedError: fmt.Errorf("PreflightValidation 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		preflight, err := testCase.testPreflight.Get()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, preflight.Name, testCase.testPreflight.Definition.Name)
			assert.Equal(t, preflight.Namespace, testCase.testPreflight.Definition.Namespace)
		}
	}
}

func buildValidTestPreflight(apiClient *clients.Settings) *PreflightValidationBuilder {
	preflightBuilder := NewPreflightValidationBuilder(apiClient, defaultPreflightName, defaultPreflightNamespace)
	preflightBuilder.Definition.ResourceVersion = "999"

	return preflightBuilder
}

func buildInValidTestPreflight(apiClient *clients.Settings) *PreflightValidationBuilder {
	preflightBuilder := NewPreflightValidationBuilder(apiClient, defaultPreflightName, "")
	preflightBuilder.Definition.ResourceVersion = "999"

	return preflightBuilder
}

func buildPreflightTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyPreflight(),
		SchemeAttachers: testSchemes,
	})
}

func buildDummyPreflight() []runtime.Object {
	return append([]runtime.Object{}, &kmmv1beta2.PreflightValidation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultPreflightName,
			Namespace: defaultPreflightNamespace,
		},

		Spec: kmmv1beta2.PreflightValidationSpec{
			KernelVersion: "",
		},
	})
}

func generatePreflight(name, nsname string) *kmmv1beta2.PreflightValidation {
	return &kmmv1beta2.PreflightValidation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsname,
		},
	}
}
