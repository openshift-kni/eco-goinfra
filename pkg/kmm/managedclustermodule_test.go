package kmm

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	mcmV1Beta1 "github.com/rh-ecosystem-edge/kernel-module-management/api-hub/v1beta1"
	"github.com/rh-ecosystem-edge/kernel-module-management/api/v1beta1"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	testSchemesMcmV1beta1 = []clients.SchemeAttacher{
		mcmV1Beta1.AddToScheme,
	}
	defaultManagedClusterModuleName      = "managedclustermoduletest"
	defaultManagedClusterModuleNamespace = "managedclustermoduletestns"
)

func TestNewManagedClusterModuleBuilder(t *testing.T) {
	testCases := []struct {
		name        string
		namespace   string
		expectedErr string
		client      bool
	}{
		{
			name:        defaultManagedClusterModuleName,
			namespace:   defaultManagedClusterModuleNamespace,
			expectedErr: "",
			client:      true,
		},
		{
			name:        defaultManagedClusterModuleName,
			namespace:   defaultManagedClusterModuleNamespace,
			expectedErr: "",
			client:      false,
		},
		{
			name:        defaultManagedClusterModuleName,
			namespace:   "",
			expectedErr: "managedClusterModule 'nsname' cannot be empty",
			client:      true,
		},
		{
			name:        "",
			namespace:   defaultManagedClusterModuleNamespace,
			expectedErr: "managedClusterModule 'name' cannot be empty",
			client:      true,
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemesV1beta1})
		}

		testBuilder := NewManagedClusterModuleBuilder(testSettings, testCase.name, testCase.namespace)

		if testCase.expectedErr == "" {
			if testCase.client {
				assert.NotNil(t, testBuilder)
				assert.Equal(t, testCase.name, testBuilder.Definition.Name)
				assert.Equal(t, testCase.namespace, testBuilder.Definition.Namespace)
			} else {
				assert.Nil(t, testBuilder)
			}
		} else {
			assert.Equal(t, testCase.expectedErr, testBuilder.errorMsg)
		}
	}
}

func TestPullManagedClusterModule(t *testing.T) {
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
			expectedError:       fmt.Errorf("managedclustermodule 'name' cannot be empty"),
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			name:                "test",
			namespace:           "",
			expectedError:       fmt.Errorf("managedclustermodule 'namespace' cannot be empty"),
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			name:                "test",
			namespace:           "testns",
			expectedError:       fmt.Errorf("managedclustermodule object test does not exist in namespace testns"),
			addToRuntimeObjects: false,
			client:              true,
		},
		{
			name:                "test",
			namespace:           "testns",
			expectedError:       fmt.Errorf("managedclustermodule 'apiClient' cannot be empty"),
			addToRuntimeObjects: true,
			client:              false,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testManagedClusterModule := generateManagedClusterModule(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testManagedClusterModule)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemesMcmV1beta1,
			})
		}

		builderResult, err := PullManagedClusterModule(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, builderResult.Definition.Name)
			assert.Equal(t, testCase.namespace, builderResult.Definition.Namespace)
		}
	}
}

func TestManagedClusterModuleBuilderGet(t *testing.T) {
	testCases := []struct {
		testManagedClusterModule *ManagedClusterModuleBuilder
		expectedError            error
	}{
		{
			testManagedClusterModule: buildValidTestManagedClusterModule(
				buildManagedClusterModuleTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testManagedClusterModule: buildInValidTestManagedClusterModule(
				buildManagedClusterModuleTestClientWithDummyObject()),
			expectedError: fmt.Errorf("managedClusterModule 'nsname' cannot be empty"),
		},
		{
			testManagedClusterModule: buildValidTestManagedClusterModule(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf(
				"managedclustermodules.hub.kmm.sigs.x-k8s.io \"managedclustermoduletest\" not found"),
		},
	}

	for _, testCase := range testCases {
		managedClusterModule, err := testCase.testManagedClusterModule.Get()

		if testCase.expectedError == nil {
			assert.Equal(t, managedClusterModule.Name, testCase.testManagedClusterModule.Definition.Name)
			assert.Equal(t, managedClusterModule.Namespace, testCase.testManagedClusterModule.Definition.Namespace)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestManagedClusterModuleExists(t *testing.T) {
	testCases := []struct {
		testManagedClusterModule *ManagedClusterModuleBuilder
		expectedStatus           bool
	}{
		{
			testManagedClusterModule: buildValidTestManagedClusterModule(
				buildManagedClusterModuleTestClientWithDummyObject()),
			expectedStatus: true,
		},
		{
			testManagedClusterModule: buildInValidTestManagedClusterModule(
				buildManagedClusterModuleTestClientWithDummyObject()),
			expectedStatus: false,
		},
		{
			testManagedClusterModule: buildValidTestManagedClusterModule(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus:           false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testManagedClusterModule.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestManagedClusterModuleCreate(t *testing.T) {
	testCases := []struct {
		testManagedClusterModule *ManagedClusterModuleBuilder
		expectedError            string
	}{
		{
			testManagedClusterModule: buildValidTestManagedClusterModule(
				buildManagedClusterModuleTestClientWithDummyObject()),
			expectedError: "",
		},
		{
			testManagedClusterModule: buildInValidTestManagedClusterModule(
				buildManagedClusterModuleTestClientWithDummyObject()),
			expectedError: "managedClusterModule 'nsname' cannot be empty",
		},
		{
			testManagedClusterModule: buildValidTestManagedClusterModule(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:            "",
		},
	}

	for _, testCase := range testCases {
		testModuleBuilder, err := testCase.testManagedClusterModule.Create()
		if testCase.expectedError == "" {
			assert.Equal(t, testModuleBuilder.Definition.Name, testModuleBuilder.Object.Name)
		} else {
			assert.Equal(t, testCase.expectedError, err.Error())
		}
	}
}

func TestManagedClusterModuleDelete(t *testing.T) {
	testCases := []struct {
		testManagedClusterModule *ManagedClusterModuleBuilder
		expectedError            error
	}{
		{
			testManagedClusterModule: buildValidTestManagedClusterModule(
				buildManagedClusterModuleTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testManagedClusterModule: buildInValidTestManagedClusterModule(
				buildManagedClusterModuleTestClientWithDummyObject()),
			expectedError: fmt.Errorf("managedClusterModule 'nsname' cannot be empty"),
		},
		{
			testManagedClusterModule: buildValidTestManagedClusterModule(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:            nil,
		},
	}

	for _, testCase := range testCases {
		_, err := testCase.testManagedClusterModule.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testManagedClusterModule.Object)
		}
	}
}

func TestManagedClusterModuleUpdate(t *testing.T) {
	testCases := []struct {
		testManagedClusterModule *ManagedClusterModuleBuilder
		expectedError            error
	}{
		{
			testManagedClusterModule: buildValidTestManagedClusterModule(
				buildManagedClusterModuleTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testManagedClusterModule: buildInValidTestManagedClusterModule(
				buildManagedClusterModuleTestClientWithDummyObject()),
			expectedError: fmt.Errorf("managedClusterModule 'nsname' cannot be empty"),
		},
		{
			testManagedClusterModule: buildValidTestManagedClusterModule(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf(
				"managedclustermodules.hub.kmm.sigs.x-k8s.io \"managedclustermoduletest\" not found"),
		},
	}

	for _, testCase := range testCases {
		assert.Empty(t, testCase.testManagedClusterModule.Definition.Spec.Selector)
		testCase.testManagedClusterModule.Definition.ResourceVersion = "999"
		testCase.testManagedClusterModule.Definition.Spec.Selector = map[string]string{"test": "test"}
		_, err := testCase.testManagedClusterModule.Update()

		if errors.IsNotFound(err) {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedError, err)
		}

		if testCase.expectedError == nil {
			assert.Equal(t, map[string]string{"test": "test"}, testCase.testManagedClusterModule.Object.Spec.Selector)
		}
	}
}

func TestManagedClusterModuleWithOptions(t *testing.T) {
	testSettings := buildManagedClusterModuleTestClientWithDummyObject()
	testBuilder := buildValidTestManagedClusterModule(testSettings).WithOptions(
		func(builder *ManagedClusterModuleBuilder) (*ManagedClusterModuleBuilder, error) {
			return builder, nil
		})

	assert.Equal(t, "", testBuilder.errorMsg)
	testBuilder = buildValidTestManagedClusterModule(testSettings).WithOptions(
		func(builder *ManagedClusterModuleBuilder) (*ManagedClusterModuleBuilder, error) {
			return builder, fmt.Errorf("error")
		})
	assert.Equal(t, "error", testBuilder.errorMsg)
}

func TestManagedClusterModuleWithModuleSpec(t *testing.T) {
	testCases := []struct {
		moduleSpec  *v1beta1.ModuleSpec
		expectedErr string
	}{
		{
			moduleSpec:  &v1beta1.ModuleSpec{Selector: map[string]string{"test": "test"}},
			expectedErr: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestManagedClusterModule(buildManagedClusterModuleTestClientWithDummyObject())
		testBuilder.WithModuleSpec(*testCase.moduleSpec)

		if testCase.expectedErr == "" {
			assert.Equal(t, testBuilder.Definition.Spec.ModuleSpec, *testCase.moduleSpec)
		} else {
			assert.Equal(t, testCase.expectedErr, testBuilder.errorMsg)
		}
	}
}

func TestManagedClusterModuleWithSpokeNamespace(t *testing.T) {
	testCases := []struct {
		spokeNamespace string
		expectedErr    string
	}{
		{
			spokeNamespace: "test",
			expectedErr:    "",
		},
		{
			spokeNamespace: "",
			expectedErr:    "invalid 'spokeNamespace' argument cannot be nil",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestManagedClusterModule(buildManagedClusterModuleTestClientWithDummyObject())
		testBuilder.WithSpokeNamespace(testCase.spokeNamespace)

		if testCase.expectedErr == "" {
			assert.Equal(t, testBuilder.Definition.Spec.SpokeNamespace, testCase.spokeNamespace)
		} else {
			assert.Equal(t, testCase.expectedErr, testBuilder.errorMsg)
		}
	}
}

func TestManagedClusterModuleWithSelector(t *testing.T) {
	testCases := []struct {
		selector    map[string]string
		expectedErr string
	}{
		{
			selector:    map[string]string{"test": "test"},
			expectedErr: "",
		},
		{
			selector:    nil,
			expectedErr: "invalid 'selector' argument cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTestManagedClusterModule(buildManagedClusterModuleTestClientWithDummyObject())
		testBuilder.WithSelector(testCase.selector)

		if testCase.expectedErr == "" {
			assert.Equal(t, testBuilder.Definition.Spec.Selector, testCase.selector)
		} else {
			assert.Equal(t, testCase.expectedErr, testBuilder.errorMsg)
		}
	}
}

func buildValidTestManagedClusterModule(apiClient *clients.Settings) *ManagedClusterModuleBuilder {
	return NewManagedClusterModuleBuilder(
		apiClient, defaultManagedClusterModuleName, defaultManagedClusterModuleNamespace)
}

func buildInValidTestManagedClusterModule(apiClient *clients.Settings) *ManagedClusterModuleBuilder {
	return NewManagedClusterModuleBuilder(apiClient, defaultManagedClusterModuleName, "")
}

func buildManagedClusterModuleTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyManagedClusterModule(),
		SchemeAttachers: testSchemesMcmV1beta1,
	})
}

func buildDummyManagedClusterModule() []runtime.Object {
	return append([]runtime.Object{}, &mcmV1Beta1.ManagedClusterModule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultManagedClusterModuleName,
			Namespace: defaultManagedClusterModuleNamespace,
		},

		Spec: mcmV1Beta1.ManagedClusterModuleSpec{},
	})
}

func generateManagedClusterModule(name, nsname string) *mcmV1Beta1.ManagedClusterModule {
	return &mcmV1Beta1.ManagedClusterModule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsname,
		},
	}
}
