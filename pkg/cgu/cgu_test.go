package cgu

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/openshift-kni/cluster-group-upgrades-operator/pkg/api/clustergroupupgrades/v1alpha1"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultCguName           = "cgu-test"
	defaultCguNsName         = "test-ns"
	defaultCguMaxConcurrency = 1
	defaultCguClusterName    = "test-cluster"
	defaultCguClusterState   = v1alpha1.NotStarted
	defaultCguCondition      = conditionComplete
)

var (
	testSchemes = []clients.SchemeAttacher{
		v1alpha1.AddToScheme,
	}
)

//nolint:funlen
func TestPullCgu(t *testing.T) {
	generateCgu := func(name, namespace string) *v1alpha1.ClusterGroupUpgrade {
		return &v1alpha1.ClusterGroupUpgrade{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: v1alpha1.ClusterGroupUpgradeSpec{},
		}
	}

	testCases := []struct {
		cguName             string
		cguNamespace        string
		expectedError       bool
		addToRuntimeObjects bool
		expectedErrorText   string
		client              bool
	}{
		{
			cguName:             "test1",
			cguNamespace:        "test-namespace",
			expectedError:       false,
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			cguName:             "test2",
			cguNamespace:        "test-namespace",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "cgu object test2 does not exist in namespace test-namespace",
			client:              true,
		},
		{
			cguName:             "",
			cguNamespace:        "test-namespace",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "cgu 'name' cannot be empty",
			client:              true,
		},
		{
			cguName:             "test3",
			cguNamespace:        "",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "cgu 'namespace' cannot be empty",
			client:              true,
		},
		{
			cguName:             "test3",
			cguNamespace:        "test-namespace",
			expectedError:       true,
			addToRuntimeObjects: false,
			expectedErrorText:   "cgu 'apiClient' cannot be empty",
			client:              false,
		},
	}
	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testCgu := generateCgu(testCase.cguName, testCase.cguNamespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testCgu)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		// Test the Pull method
		builderResult, err := Pull(testSettings, testCgu.Name, testCgu.Namespace)

		// Check the error
		if testCase.expectedError {
			assert.NotNil(t, err)

			// Check the error message
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testCgu.Name, builderResult.Object.Name)
			assert.Equal(t, testCgu.Namespace, builderResult.Object.Namespace)
		}
	}
}

func TestNewCguBuilder(t *testing.T) {
	generateCguBuilder := NewCguBuilder

	testCases := []struct {
		cguName           string
		cguNamespace      string
		cguMaxConcurrency int
		client            bool
		expectedErrorText string
	}{
		{
			cguName:           "test1",
			cguNamespace:      "test-namespace",
			cguMaxConcurrency: 1,
			client:            true,
			expectedErrorText: "",
		},
		{
			cguName:           "",
			cguNamespace:      "test-namespace",
			cguMaxConcurrency: 1,
			client:            true,
			expectedErrorText: "CGU 'name' cannot be empty",
		},
		{
			cguName:           "test1",
			cguNamespace:      "",
			cguMaxConcurrency: 1,
			client:            true,
			expectedErrorText: "CGU 'nsname' cannot be empty",
		},
		{
			cguName:           "test1",
			cguNamespace:      "test-namespace",
			cguMaxConcurrency: 0,
			client:            true,
			expectedErrorText: "CGU 'maxConcurrency' cannot be less than 1",
		},
		{
			cguName:           "test1",
			cguNamespace:      "test-namespace",
			cguMaxConcurrency: 1,
			client:            false,
			expectedErrorText: "CGU 'apiClient' cannot be nil",
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		testCguStructure := generateCguBuilder(
			testSettings,
			testCase.cguName,
			testCase.cguNamespace,
			testCase.cguMaxConcurrency)

		if testCase.client {
			assert.NotNil(t, testCguStructure)
			assert.Equal(t, testCase.expectedErrorText, testCguStructure.errorMsg)
		} else {
			assert.Nil(t, testCguStructure)
		}
	}
}

func TestCguWithCluster(t *testing.T) {
	testCases := []struct {
		cluster           string
		expectedErrorText string
	}{
		{
			cluster:           "test-cluster",
			expectedErrorText: "",
		},
		{
			cluster:           "",
			expectedErrorText: "cluster in CGU cluster spec cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := buildTestClientWithDummyCguObject()
		cguBuilder := buildValidCguTestBuilder(testSettings).WithCluster(testCase.cluster)
		assert.Equal(t, testCase.expectedErrorText, cguBuilder.errorMsg)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, cguBuilder.Definition.Spec.Clusters, []string{testCase.cluster})
		}
	}
}

func TestCguWithManagedPolicy(t *testing.T) {
	testCases := []struct {
		policy            string
		expectedErrorText string
	}{
		{
			policy:            "test-policy",
			expectedErrorText: "",
		},
		{
			policy:            "",
			expectedErrorText: "policy in CGU managedpolicies spec cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := buildTestClientWithDummyCguObject()
		cguBuilder := buildValidCguTestBuilder(testSettings).WithManagedPolicy(testCase.policy)
		assert.Equal(t, testCase.expectedErrorText, cguBuilder.errorMsg)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, cguBuilder.Definition.Spec.ManagedPolicies, []string{testCase.policy})
		}
	}
}

func TestCguWithCanary(t *testing.T) {
	testCases := []struct {
		canary            string
		expectedErrorText string
	}{
		{
			canary:            "test-canary",
			expectedErrorText: "",
		},
		{
			canary:            "",
			expectedErrorText: "canary in CGU remediationstrategy spec cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := buildTestClientWithDummyCguObject()
		cguBuilder := buildValidCguTestBuilder(testSettings).WithCanary(testCase.canary)
		assert.Equal(t, testCase.expectedErrorText, cguBuilder.errorMsg)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, cguBuilder.Definition.Spec.RemediationStrategy.Canaries, []string{testCase.canary})
		}
	}
}

func TestCguCreate(t *testing.T) {
	testCases := []struct {
		testCgu       *CguBuilder
		expectedError error
	}{
		{
			testCgu:       buildValidCguTestBuilder(buildTestClientWithDummyCguObject()),
			expectedError: nil,
		},
		{
			testCgu:       buildInvalidCguTestBuilder(buildTestClientWithDummyCguObject()),
			expectedError: fmt.Errorf("CGU 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		cguBuilder, err := testCase.testCgu.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, cguBuilder.Definition, cguBuilder.Object)
		}
	}
}

func TestCguDelete(t *testing.T) {
	testCases := []struct {
		testCgu       *CguBuilder
		expectedError error
	}{
		{
			testCgu:       buildValidCguTestBuilder(buildTestClientWithDummyCguObject()),
			expectedError: nil,
		},
		{
			testCgu:       buildValidCguTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
		{
			testCgu:       buildInvalidCguTestBuilder(buildTestClientWithDummyCguObject()),
			expectedError: fmt.Errorf("CGU 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		_, err := testCase.testCgu.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testCgu.Object)
			assert.Nil(t, testCase.testCgu.Object)
		}
	}
}

func TestCguExist(t *testing.T) {
	testCases := []struct {
		testCgu        *CguBuilder
		expectedStatus bool
	}{
		{
			testCgu:        buildValidCguTestBuilder(buildTestClientWithDummyCguObject()),
			expectedStatus: true,
		},
		{
			testCgu:        buildInvalidCguTestBuilder(buildTestClientWithDummyCguObject()),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testCgu.Exists()
		assert.Equal(t, testCase.expectedStatus, exists)
	}
}

func TestCguUpdate(t *testing.T) {
	testCases := []struct {
		force bool
	}{
		{
			force: true,
		},
		{
			force: false,
		},
	}

	for _, testCase := range testCases {
		// Create the builder rather than just adding it to the client so that the proper metadata is added and
		// the update will not fail.
		var err error

		testBuilder := buildValidCguTestBuilder(clients.GetTestClients(clients.TestClientParams{}))
		testBuilder, err = testBuilder.Create()
		assert.Nil(t, err)

		assert.NotNil(t, testBuilder.Definition)
		assert.False(t, testBuilder.Definition.Spec.Backup)

		testBuilder.Definition.Spec.Backup = true

		cguBuilder, err := testBuilder.Update(testCase.force)
		assert.NotNil(t, testBuilder.Definition)

		assert.Nil(t, err)
		assert.Equal(t, testBuilder.Definition.Name, cguBuilder.Definition.Name)
		assert.Equal(t, testBuilder.Definition.Spec.Backup, cguBuilder.Definition.Spec.Backup)
	}
}

func TestCguDeleteAndWait(t *testing.T) {
	testCases := []struct {
		testCgu       *CguBuilder
		expectedError error
	}{
		{
			testCgu:       buildValidCguTestBuilder(buildTestClientWithDummyCguObject()),
			expectedError: nil,
		},
		{
			testCgu:       buildValidCguTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
		{
			testCgu:       buildInvalidCguTestBuilder(buildTestClientWithDummyCguObject()),
			expectedError: fmt.Errorf("CGU 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		_, err := testCase.testCgu.DeleteAndWait(time.Second)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testCgu.Object)
			assert.Nil(t, testCase.testCgu.Object)
		}
	}
}

func TestCguWaitUntilDeleted(t *testing.T) {
	testCases := []struct {
		testCgu       *CguBuilder
		expectedError error
	}{
		{
			testCgu:       buildValidCguTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
		{
			testCgu:       buildValidCguTestBuilder(buildTestClientWithDummyCguObject()),
			expectedError: context.DeadlineExceeded,
		},
		{
			testCgu:       buildInvalidCguTestBuilder(buildTestClientWithDummyCguObject()),
			expectedError: fmt.Errorf("CGU 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		err := testCase.testCgu.WaitUntilDeleted(time.Second)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testCgu.Object)
		}
	}
}

func TestCguWaitForCondition(t *testing.T) {
	testCases := []struct {
		condition     metav1.Condition
		exists        bool
		conditionMet  bool
		valid         bool
		expectedError error
	}{
		{
			condition:     defaultCguCondition,
			exists:        true,
			conditionMet:  true,
			valid:         true,
			expectedError: nil,
		},
		{
			condition:     defaultCguCondition,
			exists:        false,
			conditionMet:  true,
			valid:         true,
			expectedError: fmt.Errorf("cgu object %s does not exist in namespace %s", defaultCguName, defaultCguNsName),
		},
		{
			condition:     defaultCguCondition,
			exists:        true,
			conditionMet:  false,
			valid:         true,
			expectedError: context.DeadlineExceeded,
		},
		{
			condition:     defaultCguCondition,
			exists:        true,
			conditionMet:  true,
			valid:         false,
			expectedError: fmt.Errorf("CGU 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			cguBuilder     *CguBuilder
		)

		if testCase.exists {
			cgu := buildDummyCgu(defaultCguName, defaultCguNsName, defaultCguMaxConcurrency)

			if testCase.conditionMet {
				cgu.Status.Conditions = append(cgu.Status.Conditions, testCase.condition)
			}

			runtimeObjects = append(runtimeObjects, cgu)
		}

		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  runtimeObjects,
			SchemeAttachers: testSchemes,
		})

		if testCase.valid {
			cguBuilder = buildValidCguTestBuilder(testSettings)
		} else {
			cguBuilder = buildInvalidCguTestBuilder(testSettings)
		}

		_, err := cguBuilder.WaitForCondition(testCase.condition, time.Second)
		assert.Equal(t, testCase.expectedError, err)
	}
}

func TestCguWaitUntilComplete(t *testing.T) {
	testCases := []struct {
		complete      bool
		expectedError error
	}{
		{
			complete:      true,
			expectedError: nil,
		},
		{
			complete:      false,
			expectedError: context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		cgu := buildDummyCgu(defaultCguName, defaultCguNsName, defaultCguMaxConcurrency)

		if testCase.complete {
			cgu.Status.Conditions = append(cgu.Status.Conditions, conditionComplete)
		}

		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  []runtime.Object{cgu},
			SchemeAttachers: testSchemes,
		})

		cguBuilder := buildValidCguTestBuilder(testSettings)
		_, err := cguBuilder.WaitUntilComplete(time.Second)

		assert.Equal(t, testCase.expectedError, err)
	}
}

func TestCguWaitUntilClusterInState(t *testing.T) {
	testCases := []struct {
		cluster       string
		state         string
		exists        bool
		inState       bool
		valid         bool
		expectedError error
	}{
		{
			cluster:       defaultCguClusterName,
			state:         defaultCguClusterState,
			exists:        true,
			inState:       true,
			valid:         true,
			expectedError: nil,
		},
		{
			cluster:       "",
			state:         defaultCguClusterState,
			exists:        true,
			inState:       true,
			valid:         true,
			expectedError: fmt.Errorf("cluster name cannot be empty"),
		},
		{
			cluster:       defaultCguClusterName,
			state:         "",
			exists:        true,
			inState:       true,
			valid:         true,
			expectedError: fmt.Errorf("state cannot be empty"),
		},
		{
			cluster:       defaultCguClusterName,
			state:         defaultCguClusterState,
			exists:        false,
			inState:       true,
			valid:         true,
			expectedError: fmt.Errorf("cgu object %s does not exist in namespace %s", defaultCguName, defaultCguNsName),
		},
		{
			cluster:       defaultCguClusterName,
			state:         defaultCguClusterState,
			exists:        true,
			inState:       false,
			valid:         true,
			expectedError: context.DeadlineExceeded,
		},
		{
			cluster:       defaultCguClusterName,
			state:         defaultCguClusterState,
			exists:        true,
			inState:       true,
			valid:         false,
			expectedError: fmt.Errorf("CGU 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			cguBuilder     *CguBuilder
		)

		if testCase.exists {
			cgu := buildDummyCgu(defaultCguName, defaultCguNsName, defaultCguMaxConcurrency)

			if testCase.inState {
				cgu.Status.Status.CurrentBatchRemediationProgress = map[string]*v1alpha1.ClusterRemediationProgress{
					testCase.cluster: {State: testCase.state},
				}
			}

			runtimeObjects = append(runtimeObjects, cgu)
		}

		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  runtimeObjects,
			SchemeAttachers: testSchemes,
		})

		if testCase.valid {
			cguBuilder = buildValidCguTestBuilder(testSettings)
		} else {
			cguBuilder = buildInvalidCguTestBuilder(testSettings)
		}

		_, err := cguBuilder.WaitUntilClusterInState(testCase.cluster, testCase.state, time.Second)
		assert.Equal(t, testCase.expectedError, err)
	}
}

func TestCguWaitUntilClusterComplete(t *testing.T) {
	testCases := []struct {
		complete      bool
		expectedError error
	}{
		{
			complete:      true,
			expectedError: nil,
		},
		{
			complete:      false,
			expectedError: context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		cgu := buildDummyCgu(defaultCguName, defaultCguNsName, defaultCguMaxConcurrency)

		if testCase.complete {
			cgu.Status.Status.CurrentBatchRemediationProgress = map[string]*v1alpha1.ClusterRemediationProgress{
				defaultCguClusterName: {State: v1alpha1.Completed},
			}
		}

		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  []runtime.Object{cgu},
			SchemeAttachers: testSchemes,
		})

		cguBuilder := buildValidCguTestBuilder(testSettings)
		_, err := cguBuilder.WaitUntilClusterComplete(defaultCguClusterName, time.Second)

		assert.Equal(t, testCase.expectedError, err)
	}
}

func TestCguWaitUntilClusterInProgress(t *testing.T) {
	testCases := []struct {
		inProgress    bool
		expectedError error
	}{
		{
			inProgress:    true,
			expectedError: nil,
		},
		{
			inProgress:    false,
			expectedError: context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		cgu := buildDummyCgu(defaultCguName, defaultCguNsName, defaultCguMaxConcurrency)

		if testCase.inProgress {
			cgu.Status.Status.CurrentBatchRemediationProgress = map[string]*v1alpha1.ClusterRemediationProgress{
				defaultCguClusterName: {State: v1alpha1.InProgress},
			}
		}

		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  []runtime.Object{cgu},
			SchemeAttachers: testSchemes,
		})

		cguBuilder := buildValidCguTestBuilder(testSettings)
		_, err := cguBuilder.WaitUntilClusterInProgress(defaultCguClusterName, time.Second)

		assert.Equal(t, testCase.expectedError, err)
	}
}

func TestWaitUntilBackupStarts(t *testing.T) {
	cguObject := buildDummyCgu(defaultCguName, defaultCguNsName, defaultCguMaxConcurrency)
	cguObject.Status.Backup = &v1alpha1.BackupStatus{}

	cguBuilder := buildValidCguTestBuilder(clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  []runtime.Object{cguObject},
		SchemeAttachers: testSchemes,
	}))
	cguBuilder, err := cguBuilder.WaitUntilBackupStarts(5 * time.Second)

	assert.Nil(t, err)
	assert.Equal(t, cguBuilder.Object.Name, defaultCguName)
	assert.Equal(t, cguBuilder.Object.Namespace, defaultCguNsName)
}

func TestCguBuilderValidate(t *testing.T) {
	testCases := []struct {
		builderNil    bool
		definitionNil bool
		apiClientNil  bool
		expectedError error
		builderErrMsg string
	}{
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: nil,
			builderErrMsg: "",
		},
		{
			builderNil:    true,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: fmt.Errorf("error: received nil cgu builder"),
			builderErrMsg: "",
		},
		{
			builderNil:    false,
			definitionNil: true,
			apiClientNil:  false,
			expectedError: fmt.Errorf("can not redefine the undefined cgu"),
			builderErrMsg: "",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  true,
			expectedError: fmt.Errorf("cgu builder cannot have nil apiClient"),
			builderErrMsg: "",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  false,
			builderErrMsg: "test error",
			expectedError: fmt.Errorf("test error"),
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidCguTestBuilder(clients.GetTestClients(clients.TestClientParams{}))

		if testCase.builderNil {
			testBuilder = nil
		}

		if testCase.definitionNil {
			testBuilder.Definition = nil
		}

		if testCase.apiClientNil {
			testBuilder.apiClient = nil
		}

		if testCase.builderErrMsg != "" {
			testBuilder.errorMsg = testCase.builderErrMsg
		}

		valid, err := testBuilder.validate()
		if testCase.expectedError != nil {
			assert.False(t, valid)
			assert.Equal(t, testCase.expectedError, err)
		} else {
			assert.True(t, valid)
			assert.Nil(t, err)
		}
	}
}

func buildTestClientWithDummyCguObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyCguObject(),
		SchemeAttachers: testSchemes,
	})
}

func buildDummyCguObject() []runtime.Object {
	return append([]runtime.Object{}, buildDummyCgu(defaultCguName, defaultCguNsName, defaultCguMaxConcurrency))
}

func buildDummyCgu(name, namespace string, maxConcurrency int) *v1alpha1.ClusterGroupUpgrade {
	return &v1alpha1.ClusterGroupUpgrade{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.ClusterGroupUpgradeSpec{
			RemediationStrategy: &v1alpha1.RemediationStrategySpec{
				MaxConcurrency: maxConcurrency,
			},
		},
	}
}

// buildValidCguTestBuilder returns a valid CguBuilder for testing purposes.
func buildValidCguTestBuilder(apiClient *clients.Settings) *CguBuilder {
	return NewCguBuilder(
		apiClient,
		defaultCguName,
		defaultCguNsName,
		defaultCguMaxConcurrency)
}

// buildinInvalidCguTestBuilder returns an invalid CguBuilder for testing purposes.
func buildInvalidCguTestBuilder(apiClient *clients.Settings) *CguBuilder {
	return NewCguBuilder(
		apiClient,
		defaultCguName,
		"",
		defaultCguMaxConcurrency)
}
