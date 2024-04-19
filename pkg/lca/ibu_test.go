package lca

import (
	"testing"

	lcav1alpha1 "github.com/openshift-kni/lifecycle-agent/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
)

func TestPullImageBasedUpgrade(t *testing.T) {
	testCases := []struct {
		expectedError       bool
		addToRuntimeObjects bool
		expectedErrorText   string
	}{
		{
			expectedError:       false,
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
		if testCase.expectedError {
			assert.NotNil(t, err)

			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.NotNil(t, builderResult)
		}
	}
}

func TestImageBasedUpgradeUpdate(t *testing.T) {
	testCases := []struct {
		expectedError       bool
		addToRuntimeObjects bool
		expectedErrorText   string
	}{
		{
			expectedError:       false,
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
		if testCase.expectedError {
			assert.NotNil(t, err)

			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.NotNil(t, builderResult)
		}
	}
}

func TestImageBasedUpgradeDelete(t *testing.T) {
	testCases := []struct {
		expectedError       bool
		addToRuntimeObjects bool
		expectedErrorText   string
	}{
		{
			expectedError:       false,
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
		if testCase.expectedError {
			assert.NotNil(t, err)

			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.NotNil(t, builderResult)
		}
	}
}

func TestImageBasedUpgradeGet(t *testing.T) {
	testCases := []struct {
		expectedError       bool
		addToRuntimeObjects bool
		expectedErrorText   string
	}{
		{
			expectedError:       false,
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

		// Test the Get function
		builderResult, err := ibuBuilder.Get()

		// Check the error
		if testCase.expectedError {
			assert.NotNil(t, err)

			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.NotNil(t, builderResult)
		}
	}
}

func TestImageBasedUpgradeExists(t *testing.T) {
	testCases := []struct {
		expectedError       bool
		addToRuntimeObjects bool
		expectedErrorText   string
	}{
		{
			expectedError:       false,
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

		// Test the Exists function
		builderResult := ibuBuilder.Exists()

		// Check the error
		if testCase.expectedError {
			assert.NotNil(t, err)

			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.NotNil(t, builderResult)
		}
	}
}

func TestImageBasedUpgradeWithSeedImage(t *testing.T) {
	testCases := []struct {
		expectedError       bool
		addToRuntimeObjects bool
		expectedErrorText   string
		seedImage           string
	}{
		{
			expectedError:       false,
			addToRuntimeObjects: true,
			seedImage:           "quay.io/foo",
		},
		{
			expectedError:       false,
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
		if testCase.expectedError {
			assert.NotNil(t, err)

			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.NotNil(t, builderResult)
		}
	}
}

func TestImageBasedUpgradeWithAdditionalImages(t *testing.T) {
	testCases := []struct {
		expectedError                      bool
		addToRuntimeObjects                bool
		expectedErrorText                  string
		additionalImagesConfigMapName      string
		additionalImagesConfigMapNamespace string
	}{
		{
			expectedError:                      false,
			addToRuntimeObjects:                true,
			additionalImagesConfigMapName:      "cmName",
			additionalImagesConfigMapNamespace: "nsName",
		},
		{
			expectedError:                      false,
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
		if testCase.expectedError {
			assert.NotNil(t, err)

			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.NotNil(t, builderResult)
		}
	}
}

func TestImageBasedUpgradeWithExtraManifests(t *testing.T) {
	testCases := []struct {
		expectedError                    bool
		addToRuntimeObjects              bool
		expectedErrorText                string
		extraManifestsConfigMapName      string
		extraManifestsConfigMapNamespace string
	}{
		{
			expectedError:                    false,
			addToRuntimeObjects:              true,
			extraManifestsConfigMapName:      "cmName",
			extraManifestsConfigMapNamespace: "nsName",
		},
		{
			expectedError:                    false,
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
		if testCase.expectedError {
			assert.NotNil(t, err)

			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.NotNil(t, builderResult)
		}
	}
}

func TestImageBasedUpgradeWithOadpContent(t *testing.T) {
	testCases := []struct {
		expectedError                 bool
		addToRuntimeObjects           bool
		expectedErrorText             string
		oadpContentConfigMapName      string
		oadpContentConfigMapNamespace string
	}{
		{
			expectedError:                 false,
			addToRuntimeObjects:           true,
			oadpContentConfigMapName:      "cmName",
			oadpContentConfigMapNamespace: "nsName",
		},
		{
			expectedError:                 false,
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
		if testCase.expectedError {
			assert.NotNil(t, err)

			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.NotNil(t, builderResult)
		}
	}
}

func TestImageBasedUpgradeWithSeedImageVersion(t *testing.T) {
	testCases := []struct {
		expectedError       bool
		addToRuntimeObjects bool
		expectedErrorText   string
		seedImageVersion    string
	}{
		{
			expectedError:       false,
			addToRuntimeObjects: true,
			seedImageVersion:    "seedversion",
		},
		{
			expectedError:       false,
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
		if testCase.expectedError {
			assert.NotNil(t, err)

			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.NotNil(t, builderResult)
		}
	}
}

func TestImageBasedUpgradeWithSeedImagePullSecretRef(t *testing.T) {
	testCases := []struct {
		expectedError       bool
		addToRuntimeObjects bool
		expectedErrorText   string
		pullSecretName      string
	}{
		{
			expectedError:       false,
			addToRuntimeObjects: true,
			pullSecretName:      "pull-secret",
		},
		{
			expectedError:       false,
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
		if testCase.expectedError {
			assert.NotNil(t, err)

			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.NotNil(t, builderResult)
		}
	}
}

func TestImageBasedUpgradeWaitUntilStageComplete(t *testing.T) {
	testCases := []struct {
		expectedError       bool
		addToRuntimeObjects bool
		expectedErrorText   string
		stage               string
	}{
		{
			expectedError:       false,
			addToRuntimeObjects: true,
			stage:               "Idle",
		},
		{
			expectedError:       false,
			addToRuntimeObjects: true,
			stage:               "",
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
		builderResult, err := ibuBuilder.WaitUntilStageComplete(
			testCase.stage)

		// Check the error
		if testCase.expectedError {
			assert.NotNil(t, err)

			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
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
