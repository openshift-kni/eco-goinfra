package clusterversion

import (
	"testing"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	corev1 "github.com/openshift/api/config/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Common variables.
var (
	updateChannel         = "test-channel"
	updateVersion         = "4.9.0"
	updateImage           = "4.9.0-image"
	stream                = "Z"
	availableCondition    = "Available"
	desiredVersion        = "4.8.0"
	defaultClusterVersion = "4.8.0"
	defaultImage          = "4.8.0-image"
)

// Helper function to create a Builder with a clientfunc.
func NewBuilder(apiClient *clients.Settings, name string) *Builder {
	glog.V(100).Infof(
		"Initializing new clusterversion structure with the following params: %s",
		name)

	if apiClient == nil {
		glog.V(100).Infof("clusterversion 'apiClient' cannot be empty")

		return nil
	}

	builder := Builder{
		apiClient: apiClient,
		Definition: &corev1.ClusterVersion{
			ObjectMeta: metav1.ObjectMeta{
				Name: defaultClusterVersion,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the clusterversion is empty")

		builder.errorMsg = "clusterversion 'name' cannot be empty"
	}

	return &builder
}

func TestClusterVersionPull(t *testing.T) {
	testCases := []struct {
		clusterVersionName  string
		expectedError       bool
		addToRuntimeObjects bool
		client              bool
		expectedErrorText   string
	}{
		{
			clusterVersionName:  defaultClusterVersion,
			expectedError:       false,
			addToRuntimeObjects: true,
			client:              true,
			expectedErrorText:   "",
		},
		{
			clusterVersionName:  "",
			expectedError:       true,
			addToRuntimeObjects: false,
			client:              true,
			expectedErrorText:   "clusterversion name cannot be empty",
		},
		{
			clusterVersionName:  defaultClusterVersion,
			expectedError:       true,
			addToRuntimeObjects: false,
			client:              false,
			expectedErrorText:   "clusterversion client cannot be empty",
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object
		var testSettings *clients.Settings

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, &corev1.ClusterVersion{
				ObjectMeta: metav1.ObjectMeta{
					Name: testCase.clusterVersionName,
				},
			})
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})
		}

		builderResult, err := Pull(testSettings)

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

func TestClusterVersionExists(t *testing.T) {
	testCases := []struct {
		clusterVersionExistsAlready bool
		clusterVersionName          string
	}{
		{
			clusterVersionExistsAlready: true,
			clusterVersionName:          defaultClusterVersion,
		},
		{
			clusterVersionExistsAlready: false,
			clusterVersionName:          defaultClusterVersion,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object
		var testSettings *clients.Settings

		if testCase.clusterVersionExistsAlready {
			runtimeObjects = append(runtimeObjects, &corev1.ClusterVersion{
				ObjectMeta: metav1.ObjectMeta{
					Name: testCase.clusterVersionName,
				},
			})
		}

		testBuilder := NewBuilder(testSettings, testCase.clusterVersionName)

		result := testBuilder.Exists()
		if testCase.clusterVersionExistsAlready {
			assert.True(t, result)
		} else {
			assert.False(t, result)
		}
	}
}

func TestClusterVersionWithDesiredUpdateImage(t *testing.T) {
	generateTestclusterVersion := func() *corev1.ClusterVersion {
		return &corev1.ClusterVersion{
			ObjectMeta: metav1.ObjectMeta{
				Name: defaultClusterVersion,
			},
		}
	}

	testCases := []struct {
		clusterVersionChannelExistsAlready bool
		clusterVersionName                 string
		clusterVersionImage                string
	}{
		{
			clusterVersionChannelExistsAlready: false,
			clusterVersionName:                 defaultClusterVersion,
			clusterVersionImage:                "",
		},
		{
			clusterVersionChannelExistsAlready: true,
			clusterVersionName:                 defaultClusterVersion,
			clusterVersionImage:                defaultImage,
		},
		{
			clusterVersionChannelExistsAlready: true,
			clusterVersionName:                 defaultClusterVersion,
			clusterVersionImage:                updateImage,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object
		var testSettings *clients.Settings

		if testCase.clusterVersionChannelExistsAlready {
			runtimeObjects = append(runtimeObjects, generateTestclusterVersion())
		}
		testBuilder := NewBuilder(testSettings, testCase.clusterVersionName)

		// Assert the clusterVersion before adding the image
		assert.NotNil(t, testBuilder.Definition)

		assert.Equal(t, testCase.clusterVersionName, testBuilder.Definition.Name)

		// Add the image by update
		result, err := testBuilder.Update()

		// Assert the result
		assert.NotNil(t, testBuilder.Definition)

		if !testCase.clusterVersionChannelExistsAlready {
			assert.NotNil(t, err)
			assert.Nil(t, result.Object)
		} else {
			assert.Nil(t, err)

			// test clusterverison name
			assert.Equal(t, testBuilder.Definition.Name, result.Definition.Name)

			// Test standard update function
			// assert.Equal(t, defaultImage, result.Definition.Spec.Image)

			// test desired update function
			result := testBuilder.WithDesiredUpdateChannel(updateImage)
			assert.Equal(t, testBuilder.Definition.Spec.DesiredUpdate, result.Definition.Spec.DesiredUpdate)
		}
	}
}

func TestClusterVersionWithDesiredUpdateChannel(t *testing.T) {
	generateTestclusterVersion := func() *corev1.ClusterVersion {
		return &corev1.ClusterVersion{
			ObjectMeta: metav1.ObjectMeta{
				Name: defaultClusterVersion,
			},
		}
	}

	testCases := []struct {
		clusterVersionChannelExistsAlready bool
		clusterVersionName                 string
		clusterVersionChannel              string
	}{
		{
			clusterVersionChannelExistsAlready: false,
			clusterVersionName:                 defaultClusterVersion,
			clusterVersionChannel:              "",
		},
		{
			clusterVersionChannelExistsAlready: true,
			clusterVersionName:                 defaultClusterVersion,
			clusterVersionChannel:              updateChannel,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object
		var testSettings *clients.Settings

		if testCase.clusterVersionChannelExistsAlready {
			runtimeObjects = append(runtimeObjects, generateTestclusterVersion())
		}

		testBuilder := NewBuilder(testSettings, testCase.clusterVersionName)

		// Assert the clusterVersion before adding the channel
		assert.NotNil(t, testBuilder.Definition)

		assert.Equal(t, testCase.clusterVersionName, testBuilder.Definition.Name)

		// Add the channel
		result := testBuilder.WithDesiredUpdateChannel(updateChannel)

		// Assert the result
		assert.NotNil(t, testBuilder.Definition)

		if !testCase.clusterVersionChannelExistsAlready {
			assert.Nil(t, result.Object)
		} else {
			assert.Equal(t, testBuilder.Definition.Name, result.Definition.Name)
			assert.Equal(t, testBuilder.Definition.Spec.Channel, result.Definition.Spec.Channel)
		}
	}
}

// func TestClusterVersionWaitUntilConditionTrue(t *testing.T) {
//	testCases := []struct {
//		testClusterVersionBuilder *Builder
//		condition                 string
//
//		expectedError error
//	}{
//		{
//			condition:                 "NodeInstallerProgressing",
//			testClusterVersionBuilder: buildValidClusterVersionBuilder(buildClusterVersionWithDummyObject()),
//			expectedError:             nil,
//		},
//		{
//			condition:                 "unavailable",
//			testClusterVersionBuilder: buildValidClusterVersionBuilder(buildClusterVersionWithDummyObject()),
//			expectedError:             fmt.Errorf("the unavailable condition not found exists: context deadline exceeded"),
//		},
//		{
//			condition:                 "",
//			testClusterVersionBuilder: buildValidClusterVersionBuilder(buildClusterVersionWithDummyObject()),
//			expectedError:             fmt.Errorf("ClusterVersion 'conditionType' cannot be empty"),
//		},
//		{
//			condition:                 "NodeInstallerProgressing",
//			testClusterVersionBuilder: buildValidClusterVersionBuilder(clients.GetTestClients(clients.TestClientParams{})),
//			expectedError:             fmt.Errorf("cluster ClusterVersion not found"),
//		},
//	}
//
//	for _, testCase := range testCases {
//		err := testCase.testClusterVersionBuilder.WaitUntilConditionTrue(testCase.condition, 1*time.Second)
//		if err != nil {
//			assert.Equal(t, testCase.expectedError.Error(), err.Error())
//		}
//	}
//}

// func TestClusterVersionGetNextUpdateVersionImage(t *testing.T) {
//	cv := &corev1.ClusterVersion{
//		ObjectMeta: metav1.ObjectMeta{Name: clusterVersionName},
//		Status: corev1.ClusterVersionStatus{
//			Desired: corev1.Update{Version: desiredVersion},
//			AvailableUpdates: []corev1.Update{
//				{Version: updateVersion, Image: updateImage},
//			},
//			ConditionalUpdates: []corev1.ConditionalUpdate{
//				{Release: corev1.Release{Version: "4.9.1", Image: "4.9.1-image"}},
//			},
//		},
//	}
//	builder := NewBuilder(cv, nil)
//
//	// Test with acceptConditionalVersions set to false
//	image, err := builder.GetNextUpdateVersionImage(stream, false)
//	assert.NoError(t, err)
//	assert.Equal(t, updateImage, image)
//
//	// Test with acceptConditionalVersions set to true
//	image, err = builder.GetNextUpdateVersionImage(stream, true)
//	assert.NoError(t, err)
//	assert.Equal(t, updateImage, image)
//
//	// Test when conditional update is required
//	builder.Definition.Status.AvailableUpdates = []corev1.Update{}
//	image, err = builder.GetNextUpdateVersionImage(stream, true)
//	assert.NoError(t, err)
//	assert.Equal(t, "4.9.1-image", image)
//
//	// Test when no updates are available
//	builder.Definition.Status.ConditionalUpdates = []corev1.ConditionalUpdate{}
//	image, err = builder.GetNextUpdateVersionImage(stream, false)
//	assert.Error(t, err)
//	assert.Equal(t, "", image)
//}
