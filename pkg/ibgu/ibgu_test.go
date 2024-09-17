package ibgu

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/schemes/imagebasedgroupupgrades/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	testIbguName      = "test-ibgu"
	testIbguNamespace = "test-namespace"
)

var (
	defaultIbguName   = "ibgu-test"
	defaultIbguNsName = "test-ns"
)

var testSchemes = []clients.SchemeAttacher{
	v1alpha1.AddToScheme,
}

func TestNewIbguBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		client        bool
		expectedError string
	}{
		{
			name:          testIbguName,
			namespace:     testIbguNamespace,
			client:        true,
			expectedError: "",
		},
		{
			name:          "",
			namespace:     testIbguNamespace,
			client:        true,
			expectedError: "ibgu 'name' cannot be empty",
		},
		{
			name:          testIbguName,
			namespace:     "",
			client:        true,
			expectedError: "ibgu 'nsname' cannot be empty",
		},
		{
			name:          testIbguName,
			namespace:     testIbguNamespace,
			client:        false,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		var client *clients.Settings

		if testCase.client {
			client = clients.GetTestClients(clients.TestClientParams{})
		}

		testBuilder := NewIbguBuilder(client, testCase.name, testCase.namespace)

		if testCase.client {
			assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

			if testCase.expectedError == "" {
				assert.Equal(t, testCase.name, testBuilder.Definition.Name)
				assert.Equal(t, testCase.namespace, testBuilder.Definition.Namespace)
			}
		} else {
			assert.Nil(t, testBuilder)
		}
	}
}

func TestIbguWithClusterLabelSelectors(t *testing.T) {
	testCases := []struct {
		labels        map[string]string
		expectedError string
	}{
		{
			labels:        map[string]string{"key": "value"},
			expectedError: "",
		},
		{
			labels:        map[string]string{},
			expectedError: "can not apply empty cluster label selectors to the IBGU",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateIbguBuilder()

		testBuilder.WithClusterLabelSelectors(testCase.labels)
		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.labels, testBuilder.Definition.Spec.ClusterLabelSelectors[0].MatchLabels)
		}
	}
}

func TestIbguWithSeedImageRef(t *testing.T) {
	testCases := []struct {
		seedImage     string
		seedVersion   string
		expectedError string
	}{
		{
			seedImage:     "test-image",
			seedVersion:   "v1.0",
			expectedError: "",
		},
		{
			seedImage:     "",
			seedVersion:   "v1.0",
			expectedError: "seedImage cannot be empty",
		},
		{
			seedImage:     "test-image",
			seedVersion:   "",
			expectedError: "seedVersion cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateIbguBuilder()

		testBuilder.WithSeedImageRef(testCase.seedImage, testCase.seedVersion)
		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.seedImage, testBuilder.Definition.Spec.IBUSpec.SeedImageRef.Image)
			assert.Equal(t, testCase.seedVersion, testBuilder.Definition.Spec.IBUSpec.SeedImageRef.Version)
		}
	}
}

func TestIbguWithOadpContent(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		expectedError string
	}{
		{
			name:          "test-oadp",
			namespace:     "test-ns",
			expectedError: "",
		},
		{
			name:          "",
			namespace:     "test-ns",
			expectedError: "oadp content name cannot be empty",
		},
		{
			name:          "test-oadp",
			namespace:     "",
			expectedError: "oadp content namespace cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateIbguBuilder()

		testBuilder.WithOadpContent(testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.name, testBuilder.Definition.Spec.IBUSpec.OADPContent[0].Name)
			assert.Equal(t, testCase.namespace, testBuilder.Definition.Spec.IBUSpec.OADPContent[0].Namespace)
		}
	}
}

func TestIbguWithPlan(t *testing.T) {
	testCases := []struct {
		actions        []string
		maxConcurrency int
		timeout        int
		expectedError  string
	}{
		{
			actions:        []string{"action1", "action2"},
			maxConcurrency: 2,
			timeout:        300,
			expectedError:  "",
		},
		{
			actions:        []string{},
			maxConcurrency: 2,
			timeout:        300,
			expectedError:  "plan actions cannot be empty",
		},
		{
			actions:        []string{"action1"},
			maxConcurrency: 0,
			timeout:        300,
			expectedError:  "maxConcurrency must be greater than 0",
		},
		{
			actions:        []string{"action1"},
			maxConcurrency: 2,
			timeout:        0,
			expectedError:  "timeout must be greater than 0",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateIbguBuilder()

		testBuilder.WithPlan(testCase.actions, testCase.maxConcurrency, testCase.timeout)
		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.actions, testBuilder.Definition.Spec.Plan[0].Actions)
			assert.Equal(t, testCase.maxConcurrency, testBuilder.Definition.Spec.Plan[0].RolloutStrategy.MaxConcurrency)
			assert.Equal(t, testCase.timeout, testBuilder.Definition.Spec.Plan[0].RolloutStrategy.Timeout)
		}
	}
}

func TestIbguGet(t *testing.T) {
	testCases := []struct {
		exists bool
	}{
		{
			exists: true,
		},
		{
			exists: false,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateIbgu())
		}

		testBuilder := generateIbguBuilderWithFakeObjects(runtimeObjects)

		ibgu, err := testBuilder.Get()
		if testCase.exists {
			assert.Nil(t, err)
			assert.NotNil(t, ibgu)
		} else {
			assert.NotNil(t, err)
			assert.Nil(t, ibgu)
		}
	}
}

func TestIbguExists(t *testing.T) {
	testCases := []struct {
		exists bool
	}{
		{
			exists: true,
		},
		{
			exists: false,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateIbgu())
		}

		testBuilder := generateIbguBuilderWithFakeObjects(runtimeObjects)

		assert.Equal(t, testCase.exists, testBuilder.Exists())
	}
}

func TestIbguCreate(t *testing.T) {
	testCases := []struct {
		exists bool
	}{
		{
			exists: true,
		},
		{
			exists: false,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateIbgu())
		}

		testBuilder := generateIbguBuilderWithFakeObjects(runtimeObjects)

		result, err := testBuilder.Create()
		assert.Nil(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, testIbguName, result.Definition.Name)
		assert.Equal(t, testIbguNamespace, result.Definition.Namespace)
	}
}

func TestIbguDelete(t *testing.T) {
	testCases := []struct {
		name          string
		exists        bool
		expectedError bool
	}{
		{
			name:          "Delete existing IBGU",
			exists:        true,
			expectedError: false,
		},
		{
			name:          "Delete non-existing IBGU",
			exists:        false,
			expectedError: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var runtimeObjects []runtime.Object

			if testCase.exists {
				runtimeObjects = append(runtimeObjects, generateIbgu())
			}

			testBuilder := generateIbguBuilderWithFakeObjects(runtimeObjects)

			err := testBuilder.Delete()

			if testCase.expectedError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

			// Verify that the object no longer exists
			assert.False(t, testBuilder.Exists())
		})
	}
}

func TestIbguValidate(t *testing.T) {
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
			expectedError: "error: received nil ibgu builder",
		},
		{
			builderNil:    false,
			definitionNil: true,
			apiClientNil:  false,
			expectedError: "can not redefine the undefined ibgu",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  true,
			expectedError: "ibgu builder cannot have nil apiClient",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateIbguBuilderWithFakeObjects([]runtime.Object{})

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

func generateIbguBuilderWithFakeObjects(objects []runtime.Object) *IbguBuilder {
	return &IbguBuilder{
		apiClient: clients.GetTestClients(
			clients.TestClientParams{K8sMockObjects: objects, SchemeAttachers: testSchemes}).Client,
		Definition: generateIbgu(),
	}
}

func generateIbguBuilder() *IbguBuilder {
	return &IbguBuilder{
		apiClient:  clients.GetTestClients(clients.TestClientParams{}).Client,
		Definition: generateIbgu(),
	}
}

func generateIbgu() *v1alpha1.ImageBasedGroupUpgrade {
	return &v1alpha1.ImageBasedGroupUpgrade{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testIbguName,
			Namespace: testIbguNamespace,
		},
		Spec: v1alpha1.ImageBasedGroupUpgradeSpec{},
	}
}

//nolint:funlen
func TestPullIgu(t *testing.T) {
	generateibgu := func(name, namespace string) *v1alpha1.ImageBasedGroupUpgrade {
		return &v1alpha1.ImageBasedGroupUpgrade{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: v1alpha1.ImageBasedGroupUpgradeSpec{},
		}
	}

	testCases := []struct {
		ibguName            string
		ibguNamespace       string
		expectedError       bool
		addToRuntimeObjects bool
		expectedErrorText   string
		client              bool
	}{
		{
			ibguName:            "test1",
			ibguNamespace:       "test-namespace",
			expectedError:       false,
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			ibguName:            "test2",
			ibguNamespace:       "test-namespace",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "ibgu object test2 does not exist in namespace test-namespace",
			client:              true,
		},
		{
			ibguName:            "",
			ibguNamespace:       "test-namespace",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "ibgu 'name' cannot be empty",
			client:              true,
		},
		{
			ibguName:            "test3",
			ibguNamespace:       "",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "ibgu 'namespace' cannot be empty",
			client:              true,
		},
		{
			ibguName:            "test3",
			ibguNamespace:       "test-namespace",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "ibgu 'apiClient' cannot be empty",
			client:              false,
		},
	}
	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testIbgu := generateibgu(testCase.ibguName, testCase.ibguNamespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testIbgu)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		// Test the Pull method
		builderResult, err := PullIbgu(testSettings, testIbgu.Name, testIbgu.Namespace)

		// Check the error
		if testCase.expectedError {
			assert.NotNil(t, err)

			// Check the error message
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testIbgu.Name, builderResult.Object.Name)
			assert.Equal(t, testIbgu.Namespace, builderResult.Object.Namespace)
		}
	}
}

func TestIbguDeleteAndWait(t *testing.T) {
	testCases := []struct {
		testibgu      *IbguBuilder
		expectedError error
	}{
		{
			testibgu:      buildValidIbguTestBuilder(buildTestClientWithDummyIbguObject()),
			expectedError: nil,
		},
		{
			testibgu:      buildValidIbguTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("ibgu cannot be deleted because it does not exist"),
		},
		{
			testibgu:      buildInvalidibguTestBuilder(buildTestClientWithDummyIbguObject()),
			expectedError: fmt.Errorf("ibgu 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		_, err := testCase.testibgu.DeleteAndWait(time.Second)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testibgu.Object)
			assert.Nil(t, testCase.testibgu.Object)
		}
	}
}

func TestIbguWaitUntilDeleted(t *testing.T) {
	testCases := []struct {
		testibgu      *IbguBuilder
		expectedError error
	}{
		{
			testibgu:      buildValidIbguTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
		{
			testibgu:      buildValidIbguTestBuilder(buildTestClientWithDummyIbguObject()),
			expectedError: context.DeadlineExceeded,
		},
		{
			testibgu:      buildInvalidibguTestBuilder(buildTestClientWithDummyIbguObject()),
			expectedError: fmt.Errorf("ibgu 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		err := testCase.testibgu.WaitUntilDeleted(time.Second)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testibgu.Object)
		}
	}
}

func buildTestClientWithDummyIbguObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyibguObject(),
		SchemeAttachers: testSchemes,
	})
}

func buildDummyibguObject() []runtime.Object {
	return append([]runtime.Object{}, buildDummyIbgu(defaultIbguName, defaultIbguNsName))
}

func buildDummyIbgu(name, namespace string) *v1alpha1.ImageBasedGroupUpgrade {
	return &v1alpha1.ImageBasedGroupUpgrade{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.ImageBasedGroupUpgradeSpec{
			ClusterLabelSelectors: make([]metav1.LabelSelector, 0),
		},
	}
}

// buildValidIbguTestBuilder returns a valid IbguBuilder for testing purposes.
func buildValidIbguTestBuilder(apiClient *clients.Settings) *IbguBuilder {
	return NewIbguBuilder(
		apiClient,
		defaultIbguName,
		defaultIbguNsName,
	)
}

// buildinInvalidibguTestBuilder returns an invalid ibguBuilder for testing purposes.
func buildInvalidibguTestBuilder(apiClient *clients.Settings) *IbguBuilder {
	return NewIbguBuilder(
		apiClient,
		defaultIbguName,
		"")
}
