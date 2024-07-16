package hive

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	hivev1 "github.com/openshift-kni/eco-goinfra/pkg/schemes/hive/api/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultClusterImageSetName    = "imageset"
	defaultClusterImageSetRelease = "imageset-release-1-1"
)

func TestNewClusterImageSetBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		releaseImage  string
		expectedError string
	}{
		{
			name:          "imageset",
			releaseImage:  "release-1-1",
			expectedError: "",
		},
		{
			name:          "",
			releaseImage:  "release-1-1",
			expectedError: "clusterimageset 'name' cannot be empty",
		},
		{
			name:          "imageset",
			releaseImage:  "",
			expectedError: "clusterimageset 'releaseImage' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})
		testImageSet := NewClusterImageSetBuilder(testSettings, testCase.name, testCase.releaseImage)
		assert.Equal(t, testCase.expectedError, testImageSet.errorMsg)
		assert.NotNil(t, testImageSet.Definition)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.name, testImageSet.Definition.Name)
		}
	}
}

func TestPullClusterImageSet(t *testing.T) {
	generateClusterImageSet := func(name string) *hivev1.ClusterImageSet {
		return &hivev1.ClusterImageSet{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: hivev1.ClusterImageSetSpec{
				ReleaseImage: "release-1-1",
			},
		}
	}

	testCases := []struct {
		name                string
		addToRuntimeObjects bool
		expectedError       error
		client              bool
	}{
		{
			name:                "imageset",
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("clusterimageset 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                "imageset",
			addToRuntimeObjects: false,
			expectedError:       fmt.Errorf("clusterimageset object imageset does not exist"),
			client:              true,
		},
		{
			name:                "imageset",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("clusterImageSet 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testImageSet := generateClusterImageSet(testCase.name)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testImageSet)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{K8sMockObjects: runtimeObjects})
		}

		builderResult, err := PullClusterImageSet(testSettings, testCase.name)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, builderResult.Object.Name)
		}
	}
}

func TestClusterImageSetGet(t *testing.T) {
	testCases := []struct {
		testClusterImageSet *ClusterImageSetBuilder
		expectedError       error
	}{
		{
			testClusterImageSet: buildValidClusterImageSetBuilder(buildClusterImageSetClientWithDummyObject()),
			expectedError:       nil,
		},
		{
			testClusterImageSet: buildInValidClusterImageSetBuilder(buildClusterImageSetClientWithDummyObject()),
			expectedError:       fmt.Errorf("clusterimageset 'name' cannot be empty"),
		},
		{
			testClusterImageSet: buildValidClusterImageSetBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:       fmt.Errorf("clusterimagesets.hive.openshift.io \"imageset\" not found"),
		},
	}

	for _, testCase := range testCases {
		clusterImageSet, err := testCase.testClusterImageSet.Get()
		if testCase.expectedError != nil {
			assert.Equal(t, err.Error(), testCase.expectedError.Error())
		}

		if testCase.expectedError == nil {
			assert.Equal(t, clusterImageSet.Name, testCase.testClusterImageSet.Definition.Name)
		}
	}
}

func TestClusterImageSetCreate(t *testing.T) {
	testCases := []struct {
		testClusterImageSet *ClusterImageSetBuilder
		expectedError       error
	}{
		{
			testClusterImageSet: buildValidClusterImageSetBuilder(buildClusterImageSetClientWithDummyObject()),
			expectedError:       nil,
		},
		{
			testClusterImageSet: buildInValidClusterImageSetBuilder(buildClusterImageSetClientWithDummyObject()),
			expectedError:       fmt.Errorf("clusterimageset 'name' cannot be empty"),
		},
		{
			testClusterImageSet: buildValidClusterImageSetBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:       nil,
		},
	}

	for _, testCase := range testCases {
		clusterImageSet, err := testCase.testClusterImageSet.Create()
		if testCase.expectedError != nil {
			assert.Equal(t, err.Error(), testCase.expectedError.Error())
		}

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.testClusterImageSet.Definition.Name, clusterImageSet.Object.Name)
		}
	}
}

func TestClusterImageSetUpdate(t *testing.T) {
	testCases := []struct {
		testClusterImageSet *ClusterImageSetBuilder
		expectedError       error
	}{
		{
			testClusterImageSet: buildValidClusterImageSetBuilder(buildClusterImageSetClientWithDummyObject()),
			expectedError:       nil,
		},
		{
			testClusterImageSet: buildInValidClusterImageSetBuilder(buildClusterImageSetClientWithDummyObject()),
			expectedError:       fmt.Errorf("clusterimageset 'name' cannot be empty"),
		},
		{
			testClusterImageSet: buildValidClusterImageSetBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:       fmt.Errorf("clusterimagesets.hive.openshift.io \"imageset\" not found"),
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, testCase.testClusterImageSet.Definition.Spec.ReleaseImage, "imageset-release-1-1")
		testCase.testClusterImageSet.Definition.ResourceVersion = "999"
		testCase.testClusterImageSet.Definition.Spec.ReleaseImage = "test"
		application, err := testCase.testClusterImageSet.Update(false)

		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedError, err)
			assert.Equal(t, application.Object.Spec.ReleaseImage, "test")
		}
	}
}

func TestClusterImageSetDelete(t *testing.T) {
	testCases := []struct {
		testClusterImageSet *ClusterImageSetBuilder
		expectedError       error
	}{
		{
			testClusterImageSet: buildValidClusterImageSetBuilder(buildClusterImageSetClientWithDummyObject()),
			expectedError:       nil,
		},
		{
			testClusterImageSet: buildInValidClusterImageSetBuilder(buildClusterImageSetClientWithDummyObject()),
			expectedError:       fmt.Errorf("clusterimageset 'name' cannot be empty"),
		},
		{
			testClusterImageSet: buildValidClusterImageSetBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:       nil,
		},
	}

	for _, testCase := range testCases {
		err := testCase.testClusterImageSet.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testClusterImageSet.Object)
		}
	}
}

func TestClusterImageSetExists(t *testing.T) {
	testCases := []struct {
		testClusterImageSet *ClusterImageSetBuilder
		expectedStatus      bool
	}{
		{
			testClusterImageSet: buildValidClusterImageSetBuilder(buildClusterImageSetClientWithDummyObject()),
			expectedStatus:      true,
		},
		{
			testClusterImageSet: buildInValidClusterImageSetBuilder(buildClusterImageSetClientWithDummyObject()),
			expectedStatus:      false,
		},
		{
			testClusterImageSet: buildValidClusterImageSetBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus:      false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testClusterImageSet.Exists()
		assert.Equal(t, exist, testCase.expectedStatus)
	}
}

func TestClusterImageSetWithOptions(t *testing.T) {
	testSettings := buildClusterImageSetClientWithDummyObject()
	testBuilder := buildValidClusterImageSetBuilder(testSettings).WithOptions(
		func(builder *ClusterImageSetBuilder) (*ClusterImageSetBuilder, error) {
			return builder, nil
		})

	assert.Equal(t, "", testBuilder.errorMsg)
	testBuilder = buildValidClusterImageSetBuilder(testSettings).WithOptions(
		func(builder *ClusterImageSetBuilder) (*ClusterImageSetBuilder, error) {
			return builder, fmt.Errorf("error")
		})
	assert.Equal(t, "error", testBuilder.errorMsg)
}

func TestClusterImageSetWithReleaseImage(t *testing.T) {
	testCases := []struct {
		testClusterImageSet *ClusterImageSetBuilder
		releaseImage        string
		expectedError       string
	}{
		{
			testClusterImageSet: buildValidClusterImageSetBuilder(buildClusterImageSetClientWithDummyObject()),
			releaseImage:        "release-1-1",
			expectedError:       "",
		},
		{
			testClusterImageSet: buildValidClusterImageSetBuilder(buildClusterImageSetClientWithDummyObject()),
			releaseImage:        "",
			expectedError:       "cannot set releaseImage to empty string",
		},
	}

	for _, testCase := range testCases {
		testClusterImageSetBuilder := testCase.testClusterImageSet.WithReleaseImage(testCase.releaseImage)
		assert.Equal(t, testCase.expectedError, testClusterImageSetBuilder.errorMsg)

		if testCase.releaseImage != "" {
			assert.Equal(t, testCase.releaseImage, testClusterImageSetBuilder.Definition.Spec.ReleaseImage)
		} else {
			assert.Equal(t, defaultClusterImageSetRelease, testClusterImageSetBuilder.Definition.Spec.ReleaseImage)
		}
	}
}

func buildValidClusterImageSetBuilder(apiClient *clients.Settings) *ClusterImageSetBuilder {
	builder := NewClusterImageSetBuilder(apiClient, defaultClusterImageSetName, defaultClusterImageSetRelease)

	return builder
}

func buildInValidClusterImageSetBuilder(apiClient *clients.Settings) *ClusterImageSetBuilder {
	return NewClusterImageSetBuilder(apiClient, "", defaultClusterImageSetRelease)
}

func buildClusterImageSetClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: buildDummyClusterImageSetConfig(),
	})
}

func buildDummyClusterImageSetConfig() []runtime.Object {
	return append([]runtime.Object{}, &hivev1.ClusterImageSet{
		ObjectMeta: metav1.ObjectMeta{
			ResourceVersion: "999",
			Name:            defaultClusterImageSetName,
		},
		Spec: hivev1.ClusterImageSetSpec{},
	})
}
