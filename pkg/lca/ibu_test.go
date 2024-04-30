package lca

import (
	"fmt"
	"testing"

	lcav1alpha1 "github.com/openshift-kni/lifecycle-agent/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
)

func TestImageBasedUpgradeWithOptions(t *testing.T) {
	testSettings := buildTestClientWithDummyObject()
	testBuilder, _ := PullImageBasedUpgrade(testSettings)
	testBuilder = testBuilder.WithOptions(
		func(builder *ImageBasedUpgradeBuilder) (*ImageBasedUpgradeBuilder, error) {
			return builder, nil
		})

	assert.Equal(t, "", testBuilder.errorMsg)

	testBuilder = testBuilder.WithOptions(
		func(builder *ImageBasedUpgradeBuilder) (*ImageBasedUpgradeBuilder, error) {
			return builder, fmt.Errorf("error")
		})
	assert.Equal(t, "error", testBuilder.errorMsg)
}

func TestImageBasedUpgradePull(t *testing.T) {
	testCases := []struct {
		expectedError       error
		addToRuntimeObjects bool
		expectedErrorText   string
	}{
		{
			expectedError:       nil,
			addToRuntimeObjects: true,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testIBU := generateImageBasedUpgrade()

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testIBU)
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		// Test the PullImageBasedUpgrade function
		builderResult, err := PullImageBasedUpgrade(testSettings)

		// Check the error
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil {
			assert.NotNil(t, builderResult)
		}
	}
}

func TestImageBasedUpgradeUpdate(t *testing.T) {
	testCases := []struct {
		expectedError       error
		addToRuntimeObjects bool
		expectedErrorText   string
	}{
		{
			expectedError:       nil,
			addToRuntimeObjects: true,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testIBU := generateImageBasedUpgrade()

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testIBU)
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		ibuBuilder, err := PullImageBasedUpgrade(testSettings)
		assert.Nil(t, err)

		// Test the PullImageBasedUpgrade function
		builderResult, err := ibuBuilder.WithSeedImage("quay.io/no-image").Update()

		// Check the error
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil {
			assert.NotNil(t, builderResult)
			assert.Equal(t, builderResult.Object.Spec.SeedImageRef.Image, "quay.io/no-image")
		}
	}
}

func TestImageBasedUpgradeDelete(t *testing.T) {
	testCases := []struct {
		expectedError       error
		addToRuntimeObjects bool
	}{
		{
			expectedError:       nil,
			addToRuntimeObjects: true,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testIBU := generateImageBasedUpgrade()

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testIBU)
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		ibuBuilder, err := PullImageBasedUpgrade(testSettings)
		assert.Nil(t, err)

		// Test the Delete function
		builderResult, err := ibuBuilder.Delete()

		// Check the error
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil {
			assert.Nil(t, builderResult.Object)
		}
	}
}

func TestImageBasedUpgradeGet(t *testing.T) {
	testCases := []struct {
		expectedError       error
		addToRuntimeObjects bool
	}{
		{
			expectedError:       nil,
			addToRuntimeObjects: true,
		},
		{
			expectedError:       fmt.Errorf("error: received nil ImageBasedUpgrade builder"),
			addToRuntimeObjects: false,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testIBU := generateImageBasedUpgrade()

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testIBU)
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		ibuBuilder, err := PullImageBasedUpgrade(testSettings)
		if testCase.expectedError == nil {
			assert.Nil(t, err)
		}

		// Test the Get function
		builderResult, err := ibuBuilder.Get()

		// Check the error
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil {
			assert.NotNil(t, builderResult)
		}
	}
}

func TestImageBasedUpgradeExists(t *testing.T) {
	testCases := []struct {
		expectedError       error
		addToRuntimeObjects bool
	}{
		{
			expectedError:       nil,
			addToRuntimeObjects: true,
		},
		{
			expectedError:       fmt.Errorf("imagebasedupgrade object upgrade doesn't exist"),
			addToRuntimeObjects: false,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testIBU := generateImageBasedUpgrade()

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testIBU)
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		ibuBuilder, err := PullImageBasedUpgrade(testSettings)

		if testCase.expectedError == nil {
			assert.Nil(t, err)
		}

		// Test the Exists function
		builderResult := ibuBuilder.Exists()

		// Check the error
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil {
			assert.NotNil(t, builderResult)
		}
	}
}

func TestImageBasedUpgradeWithSeedImage(t *testing.T) {
	testCases := []struct {
		expectedError       error
		addToRuntimeObjects bool
		seedImage           string
	}{
		{
			expectedError:       nil,
			addToRuntimeObjects: true,
			seedImage:           "quay.io/foo",
		},
		{
			expectedError:       nil,
			addToRuntimeObjects: true,
			seedImage:           "",
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testIBU := generateImageBasedUpgrade()

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testIBU)
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		ibuBuilder, err := PullImageBasedUpgrade(testSettings)
		assert.Nil(t, err)

		// Test the WithSeedImage function
		builderResult := ibuBuilder.WithSeedImage(testCase.seedImage)

		// Check the error
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil {
			assert.NotNil(t, builderResult)
		}
	}
}

func TestImageBasedUpgradeWithAdditionalImages(t *testing.T) {
	testCases := []struct {
		expectedError                      error
		addToRuntimeObjects                bool
		additionalImagesConfigMapName      string
		additionalImagesConfigMapNamespace string
	}{
		{
			expectedError:                      nil,
			addToRuntimeObjects:                true,
			additionalImagesConfigMapName:      "cmName",
			additionalImagesConfigMapNamespace: "nsName",
		},
		{
			expectedError:                      nil,
			addToRuntimeObjects:                true,
			additionalImagesConfigMapName:      "",
			additionalImagesConfigMapNamespace: "",
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testIBU := generateImageBasedUpgrade()

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testIBU)
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		ibuBuilder, err := PullImageBasedUpgrade(testSettings)
		assert.Nil(t, err)

		// Test the WithAdditionalImages function
		builderResult := ibuBuilder.WithAdditionalImages(
			testCase.additionalImagesConfigMapName, testCase.additionalImagesConfigMapNamespace)

		// Check the error
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil {
			assert.NotNil(t, builderResult)
		}
	}
}

func TestImageBasedUpgradeWithExtraManifests(t *testing.T) {
	testCases := []struct {
		expectedError                    error
		addToRuntimeObjects              bool
		extraManifestsConfigMapName      string
		extraManifestsConfigMapNamespace string
	}{
		{
			expectedError:                    nil,
			addToRuntimeObjects:              true,
			extraManifestsConfigMapName:      "cmName",
			extraManifestsConfigMapNamespace: "nsName",
		},
		{
			expectedError:                    nil,
			addToRuntimeObjects:              true,
			extraManifestsConfigMapName:      "",
			extraManifestsConfigMapNamespace: "",
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testIBU := generateImageBasedUpgrade()

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testIBU)
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		ibuBuilder, err := PullImageBasedUpgrade(testSettings)
		assert.Nil(t, err)

		// Test the WithExtraManifests function
		builderResult := ibuBuilder.WithExtraManifests(
			testCase.extraManifestsConfigMapName, testCase.extraManifestsConfigMapNamespace)

		// Check the error
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil {
			assert.NotNil(t, builderResult)
		}
	}
}

func TestImageBasedUpgradeWithOadpContent(t *testing.T) {
	testCases := []struct {
		expectedError                 error
		addToRuntimeObjects           bool
		oadpContentConfigMapName      string
		oadpContentConfigMapNamespace string
	}{
		{
			expectedError:                 nil,
			addToRuntimeObjects:           true,
			oadpContentConfigMapName:      "cmName",
			oadpContentConfigMapNamespace: "nsName",
		},
		{
			expectedError:                 nil,
			addToRuntimeObjects:           true,
			oadpContentConfigMapName:      "",
			oadpContentConfigMapNamespace: "",
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testIBU := generateImageBasedUpgrade()

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testIBU)
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		ibuBuilder, err := PullImageBasedUpgrade(testSettings)
		assert.Nil(t, err)

		// Test the WithExtraManifests function
		builderResult := ibuBuilder.WithExtraManifests(
			testCase.oadpContentConfigMapName, testCase.oadpContentConfigMapNamespace)

		// Check the error
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil {
			assert.NotNil(t, builderResult)
		}
	}
}

func TestImageBasedUpgradeWithSeedImageVersion(t *testing.T) {
	testCases := []struct {
		expectedError       error
		addToRuntimeObjects bool
		seedImageVersion    string
	}{
		{
			expectedError:       nil,
			addToRuntimeObjects: true,
			seedImageVersion:    "seedversion",
		},
		{
			expectedError:       nil,
			addToRuntimeObjects: true,
			seedImageVersion:    "",
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testIBU := generateImageBasedUpgrade()

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testIBU)
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		ibuBuilder, err := PullImageBasedUpgrade(testSettings)
		assert.Nil(t, err)

		// Test the WithExtraManifests function
		builderResult := ibuBuilder.WithSeedImageVersion(
			testCase.seedImageVersion)

		// Check the error
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil {
			assert.NotNil(t, builderResult)
		}
	}
}

func TestImageBasedUpgradeWithSeedImagePullSecretRef(t *testing.T) {
	testCases := []struct {
		expectedError       error
		addToRuntimeObjects bool
		pullSecretName      string
	}{
		{
			expectedError:       nil,
			addToRuntimeObjects: true,
			pullSecretName:      "pull-secret",
		},
		{
			expectedError:       nil,
			addToRuntimeObjects: true,
			pullSecretName:      "",
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testIBU := generateImageBasedUpgrade()

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testIBU)
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		ibuBuilder, err := PullImageBasedUpgrade(testSettings)
		assert.Nil(t, err)

		// Test the WithExtraManifests function
		builderResult := ibuBuilder.WithSeedImagePullSecretRef(
			testCase.pullSecretName)

		// Check the error
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil {
			assert.NotNil(t, builderResult)
		}
	}
}

func TestImageBasedUpgradeWaitUntilStageComplete(t *testing.T) {
	testCases := []struct {
		expectedError       error
		addToRuntimeObjects bool
		stage               string
	}{
		{
			expectedError:       fmt.Errorf("wrong stage selected for imagebasedupgrade"),
			addToRuntimeObjects: true,
			stage:               "wrong_stage",
		},
		{
			expectedError:       nil,
			addToRuntimeObjects: true,
			stage:               "Idle",
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.addToRuntimeObjects {
			testSettings = buildTestClientWithDummyObject()
		} else {
			testSettings = buildTestClientWithoutDummyObject()
		}

		ibuBuilder, err := PullImageBasedUpgrade(testSettings)
		if testCase.expectedError == nil {
			assert.Nil(t, err)
		}

		// Test the WithExtraManifests function
		builderResult, err := ibuBuilder.WaitUntilStageComplete(
			testCase.stage)

		// Check the error
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil {
			assert.NotNil(t, builderResult)
		}
	}
}

func TestImageBasedUpgradeWithStage(t *testing.T) {
	testCases := []struct {
		expectedError       error
		addToRuntimeObjects bool
		stage               string
	}{
		{
			expectedError:       nil,
			addToRuntimeObjects: true,
			stage:               "Idle",
		},
		{
			expectedError:       nil,
			addToRuntimeObjects: true,
			stage:               "Wrong",
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testIBU := generateImageBasedUpgrade()

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testIBU)
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		ibuBuilder, err := PullImageBasedUpgrade(testSettings)
		assert.Nil(t, err)

		// Test the WithExtraManifests function
		builderResult := ibuBuilder.WithStage(
			testCase.stage)

		// Check the error
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil {
			assert.NotNil(t, builderResult)
		}
	}
}

func generateImageBasedUpgrade() *lcav1alpha1.ImageBasedUpgrade {
	return &lcav1alpha1.ImageBasedUpgrade{
		ObjectMeta: metav1.ObjectMeta{
			Name: "upgrade",
		},
	}
}

func buildTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: buildDummyIBU(),
	})
}

func buildTestClientWithoutDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{})
}

func buildDummyIBU() []runtime.Object {
	return append([]runtime.Object{}, &lcav1alpha1.ImageBasedUpgrade{
		ObjectMeta: metav1.ObjectMeta{
			Name: "upgrade",
		},
		Spec: lcav1alpha1.ImageBasedUpgradeSpec{
			Stage: "Idle",
		},
		Status: lcav1alpha1.ImageBasedUpgradeStatus{
			Conditions: []metav1.Condition{
				{
					Type:   idle,
					Status: "True",
				},
			},
		},
	})
}
