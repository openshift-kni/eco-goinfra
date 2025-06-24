package kmm

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	kmmv1beta2 "github.com/openshift-kni/eco-goinfra/pkg/schemes/kmm/v1beta2"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultPreflightOCPName      = "preflightocp"
	defaultPreflightOCPNamespace = "preflightocpns"

	testSchemev1beta2 = []clients.SchemeAttacher{
		kmmv1beta2.AddToScheme,
	}
)

func TestNewPreflightValidationOCPBuilder(t *testing.T) {
	testCases := []struct {
		name              string
		namespace         string
		expectedPreflight *kmmv1beta2.PreflightValidationOCP
		expectedErr       string
	}{
		{
			name:      defaultPreflightOCPName,
			namespace: defaultPreflightOCPNamespace,
			expectedPreflight: &kmmv1beta2.PreflightValidationOCP{
				ObjectMeta: metav1.ObjectMeta{
					Name:      defaultPreflightOCPName,
					Namespace: defaultPreflightOCPNamespace,
				},
			},
			expectedErr: "",
		},
		{
			name:              defaultPreflightOCPName,
			namespace:         "",
			expectedPreflight: nil,
			expectedErr:       "PreflightValidationOCP 'nsname' cannot be empty",
		},
		{
			name:              "",
			namespace:         defaultPreflightOCPNamespace,
			expectedPreflight: nil,
			expectedErr:       "PreflightValidationOCP 'name' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})

		testBuilder := NewPreflightValidationOCPBuilder(testSettings, testCase.name, testCase.namespace)

		if testCase.expectedErr == "" {
			assert.NotNil(t, testBuilder)
			assert.Equal(t, testCase.expectedPreflight.Name, testBuilder.Definition.Name)
			assert.Equal(t, testCase.expectedPreflight.Namespace, testBuilder.Definition.Namespace)
		} else {
			assert.Equal(t, testCase.expectedErr, testBuilder.errorMsg)
		}
	}
}

func TestPreflightValidationOcpWithDtkImage(t *testing.T) {
	testCases := []struct {
		dtkImage    string
		expectedErr string
	}{
		{
			dtkImage:    "some-image",
			expectedErr: "",
		},
		{
			dtkImage:    "",
			expectedErr: "invalid 'dtkImage' argument can not be nil",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestPreflightOCP(buildPreflightOCPTestClientWithDummyObject())
		testBuilder.WithDtkImage(testCase.dtkImage)

		if testCase.expectedErr == "" {
			assert.Equal(t, testBuilder.Definition.Spec.DTKImage, testCase.dtkImage)
		} else {
			assert.Equal(t, testCase.expectedErr, testBuilder.errorMsg)
		}
	}
}

func TestPreflightValidationOcpWithKernelVersion(t *testing.T) {
	testCases := []struct {
		kernelVersion string
		expectedErr   string
	}{
		{
			kernelVersion: "some-image",
			expectedErr:   "",
		},
		{
			kernelVersion: "",
			expectedErr:   "invalid 'kernelVersion' argument can not be nil",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestPreflightOCP(buildPreflightOCPTestClientWithDummyObject())
		testBuilder.WithKernelVersion(testCase.kernelVersion)

		if testCase.expectedErr == "" {
			assert.Equal(t, testBuilder.Definition.Spec.KernelVersion, testCase.kernelVersion)
		} else {
			assert.Equal(t, testCase.expectedErr, testBuilder.errorMsg)
		}
	}
}

func TestPreflightValidationOCPWithPushBuiltImage(t *testing.T) {
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
		testBuilder := buildValidTestPreflightOCP(buildPreflightOCPTestClientWithDummyObject())
		testBuilder.WithPushBuiltImage(testCase.push)
		assert.Equal(t, testCase.push, testBuilder.Definition.Spec.PushBuiltImage)
	}
}

func TestPreflightValidationOCPWithOptions(t *testing.T) {
	testBuilder := buildValidTestPreflightOCP(buildPreflightOCPTestClientWithDummyObject())

	testBuilder.WithOptions(func(builder *PreflightValidationOCPBuilder) (*PreflightValidationOCPBuilder, error) {
		return builder, nil
	})

	assert.Equal(t, "", testBuilder.errorMsg)

	testBuilder.WithOptions(func(builder *PreflightValidationOCPBuilder) (*PreflightValidationOCPBuilder, error) {
		return builder, fmt.Errorf("error")
	})

	assert.Equal(t, "error", testBuilder.errorMsg)
}

func TestPullPreflightValidationOCP(t *testing.T) {
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
			expectedError:       fmt.Errorf("preflightvalidationocp 'name' cannot be empty"),
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			name:                "test",
			namespace:           "",
			expectedError:       fmt.Errorf("preflightvalidationocp 'nsname' cannot be empty"),
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			name:                "test",
			namespace:           "testns",
			expectedError:       fmt.Errorf("preflightvalidationocp object test doesn't exist in namespace testns"),
			addToRuntimeObjects: false,
			client:              true,
		},
		{
			name:                "test",
			namespace:           "testns",
			expectedError:       fmt.Errorf("preflightvalidationocp 'apiClient' cannot be empty"),
			addToRuntimeObjects: true,
			client:              false,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testPreflight := generatePreflightOCP(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testPreflight)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemev1beta2,
			})
		}

		builderResult, err := PullPreflightValidationOCP(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, builderResult.Definition.Name)
			assert.Equal(t, testCase.namespace, builderResult.Definition.Namespace)
		}
	}
}

func TestPreflightValidationOCPCreate(t *testing.T) {
	testCases := []struct {
		testPreflight *PreflightValidationOCPBuilder
		expectedError string
	}{
		{
			testPreflight: buildValidTestPreflightOCP(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: "resourceVersion can not be set for Create requests",
		},
		{
			testPreflight: buildInValidTestPreflightOCP(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: "PreflightValidationOCP 'nsname' cannot be empty",
		},
		{
			testPreflight: buildValidTestPreflightOCP(buildPreflightOCPTestClientWithDummyObject()),
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

func TestPreflightValidationOCPUpdate(t *testing.T) {
	testCases := []struct {
		testPreflightOCP *PreflightValidationOCPBuilder
		expectedError    error
		kernelVersion    string
		dtkImage         string
	}{
		{
			testPreflightOCP: buildInValidTestPreflightOCP(buildPreflightOCPTestClientWithDummyObject()),
			expectedError:    fmt.Errorf("PreflightValidationOCP 'nsname' cannot be empty"),
			kernelVersion:    "testKernel",
			dtkImage:         "",
		},
		{
			testPreflightOCP: buildValidTestPreflightOCP(buildPreflightOCPTestClientWithDummyObject()),
			expectedError:    nil,
			kernelVersion:    "testKernel",
			dtkImage:         "",
		},
		{
			testPreflightOCP: buildValidTestPreflightOCP(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:    fmt.Errorf("preflightvalidationocps.kmm.sigs.x-k8s.io \"preflightocp\" not found"),
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, "", testCase.testPreflightOCP.Definition.Spec.KernelVersion)
		testCase.testPreflightOCP.Definition.ResourceVersion = "999"
		assert.Equal(t, "", testCase.testPreflightOCP.Definition.Spec.KernelVersion)
		testCase.testPreflightOCP.Definition.Spec.DTKImage = testCase.dtkImage
		testCase.testPreflightOCP.Definition.Spec.KernelVersion = testCase.kernelVersion
		_, err := testCase.testPreflightOCP.Update()

		if errors.IsNotFound(err) {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedError, err)
		}

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.kernelVersion, testCase.testPreflightOCP.Object.Spec.KernelVersion)
			assert.Equal(t, testCase.testPreflightOCP.Object, testCase.testPreflightOCP.Definition)
		}
	}
}

func TestPreflightValidationOCPExists(t *testing.T) {
	testCases := []struct {
		testPreflight  *PreflightValidationOCPBuilder
		expectedStatus bool
	}{
		{
			testPreflight:  buildValidTestPreflightOCP(buildPreflightOCPTestClientWithDummyObject()),
			expectedStatus: true,
		},
		{
			testPreflight:  buildInValidTestPreflightOCP(buildPreflightOCPTestClientWithDummyObject()),
			expectedStatus: false,
		},
		{
			testPreflight:  buildInValidTestPreflightOCP(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testPreflight.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestPreflightValidationOCPDelete(t *testing.T) {
	testCases := []struct {
		testPreflight *PreflightValidationOCPBuilder
		expectedError error
	}{
		{
			testPreflight: buildValidTestPreflightOCP(buildPreflightOCPTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testPreflight: buildInValidTestPreflightOCP(buildPreflightOCPTestClientWithDummyObject()),
			expectedError: fmt.Errorf("PreflightValidationOCP 'nsname' cannot be empty"),
		},
		{
			testPreflight: buildValidTestPreflightOCP(clients.GetTestClients(clients.TestClientParams{})),
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

func TestPreflightValidationOCPGet(t *testing.T) {
	testCases := []struct {
		testPreflight *PreflightValidationOCPBuilder
		expectedError error
	}{
		{
			testPreflight: buildValidTestPreflightOCP(buildPreflightOCPTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testPreflight: buildInValidTestPreflightOCP(buildPreflightOCPTestClientWithDummyObject()),
			expectedError: fmt.Errorf("PreflightValidationOCP 'nsname' cannot be empty"),
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

func buildValidTestPreflightOCP(apiClient *clients.Settings) *PreflightValidationOCPBuilder {
	preflightBuilder := NewPreflightValidationOCPBuilder(apiClient, defaultPreflightOCPName, defaultPreflightOCPNamespace)
	preflightBuilder.Definition.ResourceVersion = "999"

	return preflightBuilder
}

func buildInValidTestPreflightOCP(apiClient *clients.Settings) *PreflightValidationOCPBuilder {
	preflightBuilder := NewPreflightValidationOCPBuilder(apiClient, defaultPreflightOCPName, "")
	preflightBuilder.Definition.ResourceVersion = "999"

	return preflightBuilder
}

func buildPreflightOCPTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyPreflightOCP(),
		SchemeAttachers: testSchemev1beta2,
	})
}

func buildDummyPreflightOCP() []runtime.Object {
	return append([]runtime.Object{}, &kmmv1beta2.PreflightValidationOCP{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultPreflightOCPName,
			Namespace: defaultPreflightOCPNamespace,
		},

		Spec: kmmv1beta2.PreflightValidationOCPSpec{
			KernelVersion: "",
		},
	})
}

func generatePreflightOCP(name, nsname string) *kmmv1beta2.PreflightValidationOCP {
	return &kmmv1beta2.PreflightValidationOCP{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsname,
		},
	}
}
