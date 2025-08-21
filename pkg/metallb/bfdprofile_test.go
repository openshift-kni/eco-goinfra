package metallb

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/metallb/mlbtypes"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	defaultBFDProfileName   = "default-bfd"
	defaultBFDProfileNsName = "test-namespace"
)

func TestPullBFDProfile(t *testing.T) {
	generateBFDProfile := func(name, namespace string) *mlbtypes.BFDProfile {
		return &mlbtypes.BFDProfile{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: mlbtypes.BFDProfileSpec{},
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
			name:                "bfdprofile",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("bfdprofile 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                "bfdprofile",
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("bfdprofile 'namespace' cannot be empty"),
			client:              true,
		},
		{
			name:                "bfdprofile",
			namespace:           "test-namespace",
			addToRuntimeObjects: false,
			expectedError:       fmt.Errorf("bfdprofile object bfdprofile does not exist in namespace test-namespace"),
			client:              true,
		},
		{
			name:                "bfdprofile",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("bfdprofile 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testBFDProfile := generateBFDProfile(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testBFDProfile)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		builderResult, err := PullBFDProfile(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, builderResult.Object.Name)
			assert.Equal(t, testCase.namespace, builderResult.Object.Namespace)
		}
	}
}

func TestNewBFDBuilder(t *testing.T) {
	generateBFDProfile := NewBFDBuilder

	testCases := []struct {
		name          string
		namespace     string
		expectedError string
	}{
		{
			name:          "bfdprofile",
			namespace:     "test-namespace",
			expectedError: "",
		},
		{
			name:          "",
			namespace:     "test-namespace",
			expectedError: "BFDProfile 'name' cannot be empty",
		},
		{
			name:          "bfdprofile",
			namespace:     "",
			expectedError: "BFDProfile 'nsname' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{
			SchemeAttachers: testSchemes,
		})
		testBFDProfileBuilder := generateBFDProfile(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, testBFDProfileBuilder.errorMsg)
		assert.NotNil(t, testBFDProfileBuilder.Definition)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.name, testBFDProfileBuilder.Definition.Name)
			assert.Equal(t, testCase.namespace, testBFDProfileBuilder.Definition.Namespace)
		}
	}
}

func TestBFDProfileGet(t *testing.T) {
	testCases := []struct {
		testBFDProfile *BFDBuilder
		expectedError  error
	}{
		{
			testBFDProfile: buildValidBFDProfileBuilder(buildBFDProfileTestClientWithDummyObject()),
			expectedError:  nil,
		},
		{
			testBFDProfile: buildInValidBFDProfileBuilder(buildTestClientWithDummyObject()),
			expectedError:  fmt.Errorf("BFDProfile 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		bfdProfile, err := testCase.testBFDProfile.Get()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, bfdProfile.Name, testCase.testBFDProfile.Definition.Name)
		}
	}
}

func TestBFDProfileExist(t *testing.T) {
	testCases := []struct {
		testBFDProfile *BFDBuilder
		expectedStatus bool
	}{
		{
			testBFDProfile: buildValidBFDProfileBuilder(buildBFDProfileTestClientWithDummyObject()),
			expectedStatus: true,
		},
		{
			testBFDProfile: buildInValidBFDProfileBuilder(buildBFDProfileTestClientWithDummyObject()),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testBFDProfile.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestBFDProfileCreate(t *testing.T) {
	testCases := []struct {
		testBFDProfile *BFDBuilder
		expectedError  error
	}{
		{
			testBFDProfile: buildValidBFDProfileBuilder(buildBFDProfileTestClientWithDummyObject()),
			expectedError:  nil,
		},
		{
			testBFDProfile: buildInValidBFDProfileBuilder(buildBFDProfileTestClientWithDummyObject()),
			expectedError:  fmt.Errorf("BFDProfile 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		BFDProfileBuilder, err := testCase.testBFDProfile.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, BFDProfileBuilder.Definition, BFDProfileBuilder.Object)
		}
	}
}

func TestBFDProfileDelete(t *testing.T) {
	testCases := []struct {
		testBFDProfile *BFDBuilder
		expectedError  error
	}{
		{
			testBFDProfile: buildValidBFDProfileBuilder(buildBFDProfileTestClientWithDummyObject()),
			expectedError:  nil,
		},
		{
			testBFDProfile: buildInValidBFDProfileBuilder(buildBFDProfileTestClientWithDummyObject()),
			expectedError:  fmt.Errorf("BFDProfile 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		_, err := testCase.testBFDProfile.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testBFDProfile.Object)
		}
	}
}

func TestBFDProfileUpdate(t *testing.T) {
	testCases := []struct {
		testBFDProfile *BFDBuilder
		expectedError  error
		echoMode       bool
	}{
		{
			testBFDProfile: buildValidBFDProfileBuilder(buildBFDProfileTestClientWithDummyObject()),
			expectedError:  nil,
			echoMode:       true,
		},
		{
			testBFDProfile: buildInValidBFDProfileBuilder(buildBFDProfileTestClientWithDummyObject()),
			expectedError:  fmt.Errorf("BFDProfile 'nsname' cannot be empty"),
			echoMode:       false,
		},
	}

	for _, testCase := range testCases {
		assert.Nil(t, testCase.testBFDProfile.Definition.Spec.EchoMode)
		assert.Nil(t, nil, testCase.testBFDProfile.Object)
		testCase.testBFDProfile.WithEchoMode(true)
		testCase.testBFDProfile.Definition.ObjectMeta.ResourceVersion = "999"
		_, err := testCase.testBFDProfile.Update(false)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, &testCase.echoMode, testCase.testBFDProfile.Definition.Spec.EchoMode)
		}
	}
}

func TestBFDProfileWithRcvInterval(t *testing.T) {
	testCases := []struct {
		testBFDProfile *BFDBuilder
		interval       uint32
	}{
		{
			testBFDProfile: buildValidBFDProfileBuilder(buildBFDProfileTestClientWithDummyObject()),
			interval:       10,
		},
		{
			testBFDProfile: buildValidBFDProfileBuilder(buildBFDProfileTestClientWithDummyObject()),
			interval:       20,
		},
	}

	for _, testCase := range testCases {
		bFDProfileBuilder := testCase.testBFDProfile.WithRcvInterval(testCase.interval)
		assert.Equal(t, testCase.interval, *bFDProfileBuilder.Definition.Spec.ReceiveInterval)
	}
}

func TestBFDProfileWithTransmitInterval(t *testing.T) {
	testCases := []struct {
		testBFDProfile *BFDBuilder
		interval       uint32
	}{
		{
			testBFDProfile: buildValidBFDProfileBuilder(buildBFDProfileTestClientWithDummyObject()),
			interval:       10,
		},
		{
			testBFDProfile: buildValidBFDProfileBuilder(buildBFDProfileTestClientWithDummyObject()),
			interval:       20,
		},
	}

	for _, testCase := range testCases {
		bFDProfileBuilder := testCase.testBFDProfile.WithTransmitInterval(testCase.interval)
		assert.Equal(t, testCase.interval, *bFDProfileBuilder.Definition.Spec.TransmitInterval)
	}
}

func TestBFDProfileWithEchoInterval(t *testing.T) {
	testCases := []struct {
		testBFDProfile *BFDBuilder
		interval       uint32
	}{
		{
			testBFDProfile: buildValidBFDProfileBuilder(buildBFDProfileTestClientWithDummyObject()),
			interval:       10,
		},
		{
			testBFDProfile: buildValidBFDProfileBuilder(buildBFDProfileTestClientWithDummyObject()),
			interval:       20,
		},
	}

	for _, testCase := range testCases {
		bFDProfileBuilder := testCase.testBFDProfile.WithEchoInterval(testCase.interval)
		assert.Equal(t, testCase.interval, *bFDProfileBuilder.Definition.Spec.EchoInterval)
	}
}

func TestBFDProfileWithMultiplier(t *testing.T) {
	testCases := []struct {
		testBFDProfile *BFDBuilder
		interval       uint32
	}{
		{
			testBFDProfile: buildValidBFDProfileBuilder(buildBFDProfileTestClientWithDummyObject()),
			interval:       10,
		},
		{
			testBFDProfile: buildValidBFDProfileBuilder(buildBFDProfileTestClientWithDummyObject()),
			interval:       20,
		},
	}

	for _, testCase := range testCases {
		bFDProfileBuilder := testCase.testBFDProfile.WithMultiplier(testCase.interval)
		assert.Equal(t, testCase.interval, *bFDProfileBuilder.Definition.Spec.DetectMultiplier)
	}
}

func TestBFDProfileWithEchoMode(t *testing.T) {
	testCases := []struct {
		testBFDProfile *BFDBuilder
		echoMode       bool
	}{
		{
			testBFDProfile: buildValidBFDProfileBuilder(buildBFDProfileTestClientWithDummyObject()),
			echoMode:       true,
		},
		{
			testBFDProfile: buildValidBFDProfileBuilder(buildBFDProfileTestClientWithDummyObject()),
			echoMode:       false,
		},
	}

	for _, testCase := range testCases {
		bfdProfileBuilder := testCase.testBFDProfile.WithEchoMode(testCase.echoMode)
		assert.Equal(t, testCase.echoMode, *bfdProfileBuilder.Definition.Spec.EchoMode)
	}
}

func TestBFDProfileWithPassiveMode(t *testing.T) {
	testCases := []struct {
		testBFDProfile *BFDBuilder
		passiveMode    bool
	}{
		{
			testBFDProfile: buildValidBFDProfileBuilder(buildBFDProfileTestClientWithDummyObject()),
			passiveMode:    true,
		},
		{
			testBFDProfile: buildValidBFDProfileBuilder(buildBFDProfileTestClientWithDummyObject()),
			passiveMode:    false,
		},
	}

	for _, testCase := range testCases {
		bfdProfileBuilder := testCase.testBFDProfile.WithPassiveMode(testCase.passiveMode)
		assert.Equal(t, testCase.passiveMode, *bfdProfileBuilder.Definition.Spec.PassiveMode)
	}
}

func TestBFDProfileWithWithMinimumTTL(t *testing.T) {
	testCases := []struct {
		testBFDProfile *BFDBuilder
		interval       uint32
	}{
		{
			testBFDProfile: buildValidBFDProfileBuilder(buildBFDProfileTestClientWithDummyObject()),
			interval:       10,
		},
		{
			testBFDProfile: buildValidBFDProfileBuilder(buildBFDProfileTestClientWithDummyObject()),
			interval:       20,
		},
	}

	for _, testCase := range testCases {
		bFDProfileBuilder := testCase.testBFDProfile.WithMinimumTTL(testCase.interval)
		assert.Equal(t, testCase.interval, *bFDProfileBuilder.Definition.Spec.MinimumTTL)
	}
}

func TestBFDProfileWithOptions(t *testing.T) {
	testSettings := buildBFDProfileTestClientWithDummyObject()
	testBuilder := buildValidBFDProfileBuilder(testSettings).WithOptions(
		func(builder *BFDBuilder) (*BFDBuilder, error) {
			return builder, nil
		})

	assert.Equal(t, "", testBuilder.errorMsg)
	testBuilder = buildValidBFDProfileBuilder(testSettings).WithOptions(
		func(builder *BFDBuilder) (*BFDBuilder, error) {
			return builder, fmt.Errorf("error")
		})

	assert.Equal(t, "error", testBuilder.errorMsg)
}

func TestBFDProfileGVR(t *testing.T) {
	assert.Equal(t, GetBFDProfileGVR(),
		schema.GroupVersionResource{
			Group: APIGroup, Version: APIVersion, Resource: "bfdprofiles",
		})
}

func buildValidBFDProfileBuilder(apiClient *clients.Settings) *BFDBuilder {
	return NewBFDBuilder(
		apiClient, defaultBFDProfileName, defaultBFDProfileNsName)
}

func buildInValidBFDProfileBuilder(apiClient *clients.Settings) *BFDBuilder {
	return NewBFDBuilder(
		apiClient, defaultBFDProfileName, "")
}

func buildBFDProfileTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyBFDProfile(),
		SchemeAttachers: testSchemes,
	})
}

func buildDummyBFDProfile() []runtime.Object {
	return append([]runtime.Object{}, &mlbtypes.BFDProfile{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultBFDProfileName,
			Namespace: defaultBFDProfileNsName,
		},
		Spec: mlbtypes.BFDProfileSpec{},
	})
}
