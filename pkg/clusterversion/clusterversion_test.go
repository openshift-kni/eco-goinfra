package clusterversion

import (
	"context"
	"fmt"
	"testing"
	"time"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultRelease = configv1.Release{
		Version: "4.17.1",
		Image:   "test-image",
	}
	testSchemes = []clients.SchemeAttacher{
		configv1.Install,
	}
)

func TestPullClusterVersion(t *testing.T) {
	testCases := []struct {
		addToRuntimeObjects bool
		client              bool
		expectedError       error
	}{
		{
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       nil,
		},
		{
			addToRuntimeObjects: false,
			client:              true,
			expectedError:       fmt.Errorf("clusterversion object %s does not exist", clusterVersionName),
		},
		{
			addToRuntimeObjects: true,
			client:              false,
			expectedError:       fmt.Errorf("clusterversion 'apiClient' cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testClusterVersion := buildDummyClusterVersion()

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testClusterVersion)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		testBuilder, err := Pull(testSettings)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, clusterVersionName, testBuilder.Definition.Name)
		}
	}
}

func TestClusterVersionGet(t *testing.T) {
	testCases := []struct {
		testBuilder   *Builder
		expectedError string
	}{
		{
			testBuilder:   newClusterVersionBuilder(buildTestClientWithDummyClusterVersion()),
			expectedError: "",
		},
		{
			testBuilder:   newClusterVersionBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: "clusterversions.config.openshift.io \"version\" not found",
		},
	}

	for _, testCase := range testCases {
		clusterVersion, err := testCase.testBuilder.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, testCase.testBuilder.Definition.Name, clusterVersion.Name)
		} else {
			assert.EqualError(t, err, testCase.expectedError)
		}
	}
}

func TestClusterVersionExists(t *testing.T) {
	testCases := []struct {
		testBuilder *Builder
		exists      bool
	}{
		{
			testBuilder: newClusterVersionBuilder(buildTestClientWithDummyClusterVersion()),
			exists:      true,
		},
		{
			testBuilder: newClusterVersionBuilder(clients.GetTestClients(clients.TestClientParams{})),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

func TestClusterVersionWithDesiredUpdateImage(t *testing.T) {
	testCases := []struct {
		desiredUpdateImage string
		expectedError      string
	}{
		{
			desiredUpdateImage: "test-image",
			expectedError:      "",
		},
		{
			desiredUpdateImage: "",
			expectedError:      "clusterversion 'desiredUpdateImage' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := newClusterVersionBuilder(clients.GetTestClients(clients.TestClientParams{}))
		testBuilder = testBuilder.WithDesiredUpdateImage(testCase.desiredUpdateImage, true)

		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.desiredUpdateImage, testBuilder.Definition.Spec.DesiredUpdate.Image)
			assert.True(t, testBuilder.Definition.Spec.DesiredUpdate.Force)
		}
	}
}

func TestClusterVersionWithDesiredUpdateChannel(t *testing.T) {
	testCases := []struct {
		desiredUpdateChannel string
		expectedError        string
	}{
		{
			desiredUpdateChannel: "test-channel",
			expectedError:        "",
		},
		{
			desiredUpdateChannel: "",
			expectedError:        "clusterversion 'updateChannel' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := newClusterVersionBuilder(clients.GetTestClients(clients.TestClientParams{}))
		testBuilder = testBuilder.WithDesiredUpdateChannel(testCase.desiredUpdateChannel)

		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.desiredUpdateChannel, testBuilder.Definition.Spec.Channel)
		}
	}
}

func TestClusterVersionUpdate(t *testing.T) {
	testCases := []struct {
		testBuilder   *Builder
		expectedError error
	}{
		{
			testBuilder:   newClusterVersionBuilder(buildTestClientWithDummyClusterVersion()),
			expectedError: nil,
		},
		{
			testBuilder:   newClusterVersionBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("clusterversion object %s does not exist", clusterVersionName),
		},
	}

	for _, testCase := range testCases {
		assert.Empty(t, testCase.testBuilder.Definition.Spec.Channel)

		testCase.testBuilder.Definition.ResourceVersion = "999"
		testCase.testBuilder.Definition.Spec.Channel = "stable"

		testBuilder, err := testCase.testBuilder.Update()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, "stable", testBuilder.Object.Spec.Channel)
		}
	}
}

func TestClusterVersionWaitUntilProgressing(t *testing.T) {
	testWaitUntilConditionTrueHelper(t, configv1.OperatorProgressing, func(builder *Builder) error {
		return builder.WaitUntilProgressing(time.Second)
	})
}

func TestClusterVersionWaitUntilAvailable(t *testing.T) {
	testWaitUntilConditionTrueHelper(t, configv1.OperatorAvailable, func(builder *Builder) error {
		return builder.WaitUntilAvailable(time.Second)
	})
}

func TestClusterVersionWaitUntilConditionTrue(t *testing.T) {
	testWaitUntilConditionTrueHelper(t, configv1.OperatorAvailable, func(builder *Builder) error {
		return builder.WaitUntilConditionTrue(configv1.OperatorAvailable, time.Second)
	})
}

func TestClusterVersionWaitUntilUpdateIsStarted(t *testing.T) {
	testWaitUntilUpdateHistoryStateHelper(t, configv1.PartialUpdate, func(builder *Builder) error {
		return builder.WaitUntilUpdateIsStarted(time.Second)
	})
}

func TestClusterVersionWaitUntilUpdateIsCompleted(t *testing.T) {
	testWaitUntilUpdateHistoryStateHelper(t, configv1.CompletedUpdate, func(builder *Builder) error {
		return builder.WaitUntilUpdateIsCompleted(time.Second)
	})
}

func TestClusterVersionWaitUntilUpdateHistoryStateTrue(t *testing.T) {
	testWaitUntilUpdateHistoryStateHelper(t, configv1.CompletedUpdate, func(builder *Builder) error {
		return builder.WaitUntilUpdateHistoryStateTrue(configv1.CompletedUpdate, time.Second)
	})
}

func TestClusterVersionGetNextUpdateVersionImage(t *testing.T) {
	testCases := []struct {
		stream            string
		acceptConditional bool
		exists            bool
		hasAvailable      bool
		hasConditional    bool
		expectedError     error
	}{
		{
			stream:            Z,
			acceptConditional: false,
			exists:            true,
			hasAvailable:      true,
			hasConditional:    false,
			expectedError:     nil,
		},
		{
			stream:            Z,
			acceptConditional: true,
			exists:            true,
			hasAvailable:      false,
			hasConditional:    true,
			expectedError:     nil,
		},
		{
			stream:            "",
			acceptConditional: false,
			exists:            true,
			hasAvailable:      true,
			hasConditional:    false,
			expectedError:     fmt.Errorf("stream can not be empty"),
		},
		{
			stream:            Z,
			acceptConditional: false,
			exists:            false,
			hasAvailable:      false,
			hasConditional:    false,
			expectedError:     fmt.Errorf("clusterversion object %s does not exist", clusterVersionName),
		},
		{
			stream:            Z,
			acceptConditional: false,
			exists:            true,
			hasAvailable:      false,
			hasConditional:    false,
			expectedError:     fmt.Errorf("update version in Z stream not found"),
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		if testCase.exists {
			clusterVersion := buildDummyClusterVersion()
			clusterVersion.Status.Desired.Version = "4.17.0"

			if testCase.hasAvailable {
				clusterVersion.Status.AvailableUpdates = []configv1.Release{defaultRelease}
			}

			if testCase.hasConditional {
				clusterVersion.Status.ConditionalUpdates = []configv1.ConditionalUpdate{{Release: defaultRelease}}
			}

			runtimeObjects = append(runtimeObjects, clusterVersion)
		}

		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  runtimeObjects,
			SchemeAttachers: testSchemes,
		})
		testBuilder := newClusterVersionBuilder(testSettings)

		nextUpdate, err := testBuilder.GetNextUpdateVersionImage(testCase.stream, testCase.acceptConditional)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, "test-image", nextUpdate)
		}
	}
}

// testWaitUntilConditionTrueHelper is a helper for functions that call WaitUntilConditionTrue. It will run the test
// cases and ensure the provided condition type is set then call testFunc and compare errors.
func testWaitUntilConditionTrueHelper(
	t *testing.T, condType configv1.ClusterStatusConditionType, testFunc func(builder *Builder) error) {
	t.Helper()

	testCases := []struct {
		exists        bool
		hasCondType   bool
		expectedError error
	}{
		{
			exists:        true,
			hasCondType:   true,
			expectedError: nil,
		},
		{
			exists:        false,
			hasCondType:   true,
			expectedError: fmt.Errorf("clusterversion object %s does not exist", clusterVersionName),
		},
		{
			exists:        true,
			hasCondType:   false,
			expectedError: context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		if testCase.exists {
			clusterVersion := buildDummyClusterVersion()

			if testCase.hasCondType {
				clusterVersion.Status.Conditions = []configv1.ClusterOperatorStatusCondition{{
					Type:   condType,
					Status: configv1.ConditionTrue,
				}}
			}

			runtimeObjects = append(runtimeObjects, clusterVersion)
		}

		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  runtimeObjects,
			SchemeAttachers: testSchemes,
		})
		testBuilder := newClusterVersionBuilder(testSettings)

		err := testFunc(testBuilder)
		assert.Equal(t, testCase.expectedError, err)
	}
}

// testWaitUntilUpdateHistoryStateHelper is a helper for functions that call WaitUntilUpdateHistoryStateTrue. It will
// run the test cases and set the provided state then call testFunc and compare errors.
func testWaitUntilUpdateHistoryStateHelper(
	t *testing.T, state configv1.UpdateState, testFunc func(builder *Builder) error) {
	t.Helper()

	testCases := []struct {
		exists        bool
		inState       bool
		expectedError error
	}{
		{
			exists:        true,
			inState:       true,
			expectedError: nil,
		},
		{
			exists:        false,
			inState:       true,
			expectedError: fmt.Errorf("clusterversion object %s does not exist", clusterVersionName),
		},
		{
			exists:        true,
			inState:       false,
			expectedError: context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		if testCase.exists {
			clusterVersion := buildDummyClusterVersion()
			clusterVersion.Status.Desired.Image = "test-image"

			if testCase.inState {
				clusterVersion.Status.History = []configv1.UpdateHistory{{
					Image: "test-image",
					State: state,
				}}
			}

			runtimeObjects = append(runtimeObjects, clusterVersion)
		}

		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  runtimeObjects,
			SchemeAttachers: testSchemes,
		})
		testBuilder := newClusterVersionBuilder(testSettings)

		err := testFunc(testBuilder)
		assert.Equal(t, testCase.expectedError, err)
	}
}

// buildDummyClusterVersion returns a ClusterVersion with the clusterVersionName.
func buildDummyClusterVersion() *configv1.ClusterVersion {
	return &configv1.ClusterVersion{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterVersionName,
		},
	}
}

func buildTestClientWithDummyClusterVersion() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  []runtime.Object{buildDummyClusterVersion()},
		SchemeAttachers: testSchemes,
	})
}

// newClusterVersionBuilder returns a Builder with the provided apiClient and the clusterVersionName.
func newClusterVersionBuilder(apiClient *clients.Settings) *Builder {
	return &Builder{
		apiClient:  apiClient.Client,
		Definition: buildDummyClusterVersion(),
	}
}
