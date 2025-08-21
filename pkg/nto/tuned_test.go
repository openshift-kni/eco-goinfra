package nto //nolint:misspell

import (
	"fmt"
	"testing"

	tunedv1 "github.com/openshift/cluster-node-tuning-operator/pkg/apis/tuned/v1"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	tunedAPIGroup           = "tuned.openshift.io"
	tunedAPIVersion         = "v1"
	tunedKind               = "Tuned"
	defaultTunedName        = "default"
	defaultTunedNamespace   = "openshift-cluster-node-tuning-operator"
	defaultTunedProfileName = "openshift"
	defaultTunedProfileData = "[main]\nsummary=Optimize systems running OpenShift (provider specific parent profile)" +
		"\ninclude=-provider-${f:exec:cat:/var/lib/ocp-tuned/provider},openshift\n"
)

func TestPullTuned(t *testing.T) {
	generateTuned := func(name, namespace string) *tunedv1.Tuned {
		return &tunedv1.Tuned{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
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
			name:                "test",
			namespace:           "openshift-cluster-node-tuning-operator",
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           "openshift-cluster-node-tuning-operator",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("tuned 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                "test",
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("tuned 'nsname' cannot be empty"),
			client:              true,
		},
		{
			name:                "tunedtest",
			namespace:           "openshift-cluster-node-tuning-operator",
			addToRuntimeObjects: false,
			expectedError: fmt.Errorf("tuned object tunedtest does not exist in " +
				"namespace openshift-cluster-node-tuning-operator"),
			client: true,
		},
		{
			name:                "tunedtest",
			namespace:           "openshift-cluster-node-tuning-operator",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("tuned 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testTuned := generateTuned(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testTuned)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})
		}

		builderResult, err := PullTuned(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		} else {
			assert.Equal(t, testTuned.Name, builderResult.Object.Name)
			assert.Nil(t, err)
		}
	}
}

func TestNewTunedBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		expectedError string
	}{
		{
			name:          defaultTunedName,
			namespace:     defaultTunedNamespace,
			expectedError: "",
		},
		{
			name:          "",
			namespace:     defaultTunedNamespace,
			expectedError: "tuned 'name' cannot be empty",
		},
		{
			name:          defaultTunedName,
			namespace:     "",
			expectedError: "tuned 'nsname' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})
		testTunedBuilder := NewTunedBuilder(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, testTunedBuilder.errorMsg)
		assert.NotNil(t, testTunedBuilder.Definition)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.name, testTunedBuilder.Definition.Name)
			assert.Equal(t, testCase.namespace, testTunedBuilder.Definition.Namespace)
		}
	}
}

func TestTunedExists(t *testing.T) {
	testCases := []struct {
		testTuned      *TunedBuilder
		expectedStatus bool
	}{
		{
			testTuned:      buildValidTunedBuilder(buildTunedClientWithDummyObject()),
			expectedStatus: true,
		},
		{
			testTuned:      buildInValidTunedBuilder(buildTunedClientWithDummyObject()),
			expectedStatus: false,
		},
		{
			testTuned:      buildValidTunedBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testTuned.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestTunedGet(t *testing.T) {
	testCases := []struct {
		testTuned     *TunedBuilder
		expectedError error
	}{
		{
			testTuned:     buildValidTunedBuilder(buildTunedClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testTuned:     buildInValidTunedBuilder(buildTunedClientWithDummyObject()),
			expectedError: fmt.Errorf("tuneds.tuned.openshift.io \"\" not found"),
		},
		{
			testTuned:     buildValidTunedBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("tuneds.tuned.openshift.io \"default\" not found"),
		},
	}

	for _, testCase := range testCases {
		tunedObj, err := testCase.testTuned.Get()

		if testCase.expectedError == nil {
			assert.Equal(t, tunedObj.Name, testCase.testTuned.Definition.Name)
			assert.Equal(t, tunedObj.Namespace, testCase.testTuned.Definition.Namespace)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestTunedCreate(t *testing.T) {
	testCases := []struct {
		testTuned     *TunedBuilder
		expectedError string
	}{
		{
			testTuned:     buildValidTunedBuilder(buildTunedClientWithDummyObject()),
			expectedError: "",
		},
		{
			testTuned:     buildInValidTunedBuilder(buildTunedClientWithDummyObject()),
			expectedError: "Tuned.tuned.openshift.io \"\" is invalid: metadata.name: Required value: name is required",
		},
		{
			testTuned:     buildValidTunedBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testTunedBuilder, err := testCase.testTuned.Create()

		if testCase.expectedError == "" {
			assert.Equal(t, testTunedBuilder.Definition.Name, testTunedBuilder.Object.Name)
			assert.Equal(t, testTunedBuilder.Definition.Namespace, testTunedBuilder.Object.Namespace)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError, err.Error())
		}
	}
}

func TestTunedDelete(t *testing.T) {
	testCases := []struct {
		testTuned     *TunedBuilder
		expectedError error
	}{
		{
			testTuned:     buildValidTunedBuilder(buildTunedClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testTuned:     buildInValidTunedBuilder(buildTunedClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testTuned:     buildValidTunedBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		_, err := testCase.testTuned.Delete()

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testTuned.Object)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestTunedUpdate(t *testing.T) {
	testCases := []struct {
		testTuned     *TunedBuilder
		expectedError string
		profile       tunedv1.TunedProfile
	}{
		{
			testTuned:     buildValidTunedBuilder(buildTunedClientWithDummyObject()),
			expectedError: "",
			profile: tunedv1.TunedProfile{
				Name: &defaultTunedProfileName,
				Data: &defaultTunedProfileData,
			},
		},
		{
			testTuned: buildInValidTunedBuilder(buildTunedClientWithDummyObject()),
			expectedError: "Tuned.tuned.openshift.io \"\" is invalid: metadata.name: " +
				"Required value: name is required",
			profile: tunedv1.TunedProfile{
				Name: &defaultTunedProfileName,
				Data: &defaultTunedProfileData,
			},
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, []tunedv1.TunedProfile(nil), testCase.testTuned.Definition.Spec.Profile)
		assert.Nil(t, nil, testCase.testTuned.Object)
		testCase.testTuned.WithProfile(testCase.profile)
		_, err := testCase.testTuned.Update()

		if testCase.expectedError != "" {
			assert.Equal(t, testCase.expectedError, err.Error())
		} else {
			assert.Equal(t, []tunedv1.TunedProfile{testCase.profile}, testCase.testTuned.Definition.Spec.Profile)
		}
	}
}

func TestTunedWithProfile(t *testing.T) {
	testCases := []struct {
		testProfile       tunedv1.TunedProfile
		expectedError     bool
		expectedErrorText string
	}{
		{
			testProfile: tunedv1.TunedProfile{
				Name: &defaultTunedName,
				Data: &defaultTunedProfileData,
			},
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			testProfile:       tunedv1.TunedProfile{},
			expectedError:     false,
			expectedErrorText: "'profile' argument cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidTunedBuilder(buildTunedClientWithDummyObject())

		result := testBuilder.WithProfile(testCase.testProfile)

		if testCase.expectedError {
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
			}
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, []tunedv1.TunedProfile{testCase.testProfile}, result.Definition.Spec.Profile)
		}
	}
}

func buildValidTunedBuilder(apiClient *clients.Settings) *TunedBuilder {
	tunedBuilder := NewTunedBuilder(
		apiClient, defaultTunedName, defaultTunedNamespace)
	tunedBuilder.Definition.TypeMeta = metav1.TypeMeta{
		Kind:       tunedKind,
		APIVersion: fmt.Sprintf("%s/%s", tunedAPIGroup, tunedAPIVersion),
	}

	return tunedBuilder
}

func buildInValidTunedBuilder(apiClient *clients.Settings) *TunedBuilder {
	tunedBuilder := NewTunedBuilder(
		apiClient, "", defaultTunedNamespace)
	tunedBuilder.Definition.TypeMeta = metav1.TypeMeta{
		Kind:       tunedKind,
		APIVersion: fmt.Sprintf("%s/%s", tunedAPIGroup, tunedAPIVersion),
	}

	return tunedBuilder
}

func buildTunedClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: buildDummyTuned(),
	})
}

func buildDummyTuned() []runtime.Object {
	return append([]runtime.Object{}, &tunedv1.Tuned{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultTunedName,
			Namespace: defaultTunedNamespace,
		},
	})
}
