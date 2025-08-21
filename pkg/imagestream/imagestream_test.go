package imagestream

import (
	"fmt"
	"testing"

	goclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/golang/glog"
	imagev1 "github.com/openshift/api/image/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	imageStreamName      = "network-tools"
	imageStreamNamespace = "openshift"
	imageTag             = "latest"
	testSchemes          = []clients.SchemeAttacher{
		imagev1.AddToScheme,
	}
)

func TestImageStreamPull(t *testing.T) {
	testCases := []struct {
		name                string
		namespace           string
		addToRuntimeObjects bool
		expectedError       error
		client              bool
	}{
		{
			name:                imageStreamName,
			namespace:           imageStreamNamespace,
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           imageStreamNamespace,
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("imageStream 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                imageStreamName,
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("imageStream 'nsname' cannot be empty"),
			client:              true,
		},
		{
			name:                imageStreamName,
			namespace:           imageStreamNamespace,
			addToRuntimeObjects: false,
			expectedError: fmt.Errorf("imageStream object %s does not exist in namespace %s",
				imageStreamName, imageStreamNamespace),
			client: true,
		},
		{
			name:                imageStreamName,
			namespace:           imageStreamNamespace,
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("imageStream 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testImageStream := buildDummyImageStream(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testImageStream)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		builderResult, err := Pull(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testImageStream.Name, builderResult.Object.Name)
			assert.Equal(t, testImageStream.Namespace, builderResult.Object.Namespace)
		}
	}
}

func TestImageStreamGet(t *testing.T) {
	testCases := []struct {
		testImageStream *Builder
		expectedError   string
	}{
		{
			testImageStream: buildValidImageStreamBuilder(buildImageStreamClientWithDummyObject()),
			expectedError:   "",
		},
		{
			testImageStream: buildInValidImageStreamBuilder(buildImageStreamClientWithDummyObject()),
			expectedError:   "the imageStream 'name' cannot be empty",
		},
		{
			testImageStream: buildValidImageStreamBuilder(buildTestClientWithImageStreamScheme()),
			expectedError:   "imagestreams.image.openshift.io \"network-tools\" not found",
		},
	}

	for _, testCase := range testCases {
		imageStreamObj, err := testCase.testImageStream.Get()

		if testCase.expectedError == "" {
			assert.Equal(t, imageStreamObj.Name, testCase.testImageStream.Definition.Name)
			assert.Equal(t, imageStreamObj.Namespace, testCase.testImageStream.Definition.Namespace)
			assert.Nil(t, err)
		} else {
			assert.EqualError(t, err, testCase.expectedError)
		}
	}
}

func TestImageStreamExists(t *testing.T) {
	testCases := []struct {
		testImageStream *Builder
		expectedStatus  bool
	}{
		{
			testImageStream: buildValidImageStreamBuilder(buildImageStreamClientWithDummyObject()),
			expectedStatus:  true,
		},
		{
			testImageStream: buildInValidImageStreamBuilder(buildImageStreamClientWithDummyObject()),
			expectedStatus:  false,
		},
		{
			testImageStream: buildValidImageStreamBuilder(buildTestClientWithImageStreamScheme()),
			expectedStatus:  false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testImageStream.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

//nolint:funlen
func TestImageStreamGetDockerImage(t *testing.T) {
	testCases := []struct {
		testImageStreamTags []imagev1.TagReference
		testImageTag        string
		expectedError       error
	}{
		{
			testImageStreamTags: []imagev1.TagReference{{
				Name: "latest",
				From: &corev1.ObjectReference{
					Kind: "DockerImage",
					Name: "quay.io/dummy-server/ocp-v4.0-art-dev@sha256:c8bc1d7bdf77538653ff1cd40d9bfa00f0",
				},
			}},
			testImageTag:  imageTag,
			expectedError: nil,
		},
		{
			testImageStreamTags: []imagev1.TagReference{{
				Name: "7.4.0",
				From: &corev1.ObjectReference{
					Kind: "DockerImage",
					Name: "registry.dummy.io/jboss-eap-7/eap74-openjdk8-openshift-rhel7:latest",
				}}, {
				Name: "latest",
				From: &corev1.ObjectReference{
					Kind: "DockerImage",
					Name: "registry.dummy.io/jboss-eap-7/eap74-openjdk8-openshift-rhel7:latest",
				},
			}},
			testImageTag:  imageTag,
			expectedError: nil,
		},
		{
			testImageStreamTags: []imagev1.TagReference{},
			testImageTag:        imageTag,
			expectedError:       fmt.Errorf("imageStream object network-tools in namespace openshift has no tags"),
		},
		{
			testImageStreamTags: []imagev1.TagReference{{
				Name: "latest",
				From: &corev1.ObjectReference{
					Kind: "DockerImage",
					Name: "quay.io/dummy-server/ocp-v4.0-art-dev@sha256:c8bc1d7bdf77538653ff1cd40d9bfa00f0",
				},
			}},
			testImageTag: "4.16",
			expectedError: fmt.Errorf("image tag 4.16 not found for imageStream object network-tools " +
				"in namespace openshift"),
		},
		{
			testImageStreamTags: []imagev1.TagReference{{
				Name: "",
				From: &corev1.ObjectReference{
					Kind: "DockerImage",
					Name: "quay.io/dummy-server/ocp-v4.0-art-dev@sha256:c8bc1d7bdf77538653ff1cd40d9bfa00f0",
				},
			}},
			testImageTag:  imageTag,
			expectedError: fmt.Errorf("imageStream object network-tools in namespace openshift has no DockerImage tag value"),
		},
		{
			testImageStreamTags: []imagev1.TagReference{{
				Name: "latest",
				From: nil,
			}},
			testImageTag:  imageTag,
			expectedError: fmt.Errorf("imageStream object network-tools in namespace openshift has no DockerImage value"),
		},
		{
			testImageStreamTags: []imagev1.TagReference{{
				Name: "latest",
				From: &corev1.ObjectReference{
					Kind: "DockerImage",
					Name: "quay.io/dummy-server/ocp-v4.0-art-dev@sha256:c8bc1d7bdf77538653ff1cd40d9bfa00f0",
				},
			}},
			testImageTag:  "",
			expectedError: fmt.Errorf("imageStream 'imageTag' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		testImageStreamClientWithDummyObject := buildDummyImageStreamWithTag(testCase.testImageStreamTags)

		runtimeObjects = append(runtimeObjects, testImageStreamClientWithDummyObject)

		dummyImageStream := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  runtimeObjects,
			SchemeAttachers: testSchemes,
		})

		testImageStreamBuilder := buildValidImageStreamBuilder(dummyImageStream)

		dockerImage, err := testImageStreamBuilder.GetDockerImage(testCase.testImageTag)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.NotEqual(t, "", dockerImage)
		} else {
			assert.Equal(t, "", dockerImage)
		}
	}
}

func buildValidImageStreamBuilder(apiClient *clients.Settings) *Builder {
	imageStreamBuilder := newBuilder(
		apiClient, imageStreamName, imageStreamNamespace)

	return imageStreamBuilder
}

func buildInValidImageStreamBuilder(apiClient *clients.Settings) *Builder {
	imageStreamBuilder := newBuilder(
		apiClient, "", imageStreamNamespace)

	return imageStreamBuilder
}

func buildImageStreamClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyImageStream(imageStreamName, imageStreamNamespace),
		},
		SchemeAttachers: testSchemes,
	})
}

// buildImageStreamClientWithScheme returns a client with no objects but the ImageStream scheme attached.
func buildTestClientWithImageStreamScheme() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: testSchemes,
	})
}

func buildDummyImageStream(name, namespace string) *imagev1.ImageStream {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: imagev1.ImageStreamSpec{
			LookupPolicy: imagev1.ImageLookupPolicy{
				Local: false,
			},
			Tags: []imagev1.TagReference{{
				Name: "latest",
				From: &corev1.ObjectReference{
					Kind: "DockerImage",
					Name: "quay.io/dummy-server/ocp-v4.0-art-dev@sha256:c8bc1d7bdf77538653ff1cd40d9bfa00f0",
				},
			}},
		},
	}
}

func buildDummyImageStreamWithTag(imagestreamTags []imagev1.TagReference) *imagev1.ImageStream {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name:      imageStreamName,
			Namespace: imageStreamNamespace,
		},
		Spec: imagev1.ImageStreamSpec{
			LookupPolicy: imagev1.ImageLookupPolicy{
				Local: false,
			},
			Tags: imagestreamTags,
		},
	}
}

// newBuilder method creates new instance of builder (for the unit test propose only).
func newBuilder(apiClient *clients.Settings, name, namespace string) *Builder {
	glog.V(100).Infof("Initializing new Builder structure with the name %s in namespace %s",
		name, namespace)

	var client goclient.Client

	if apiClient != nil {
		client = apiClient.Client
	}

	builder := &Builder{
		apiClient: client,
		Definition: &imagev1.ImageStream{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: imagev1.ImageStreamSpec{
				LookupPolicy: imagev1.ImageLookupPolicy{
					Local: false,
				},
				Tags: []imagev1.TagReference{{
					Name: "latest",
					From: &corev1.ObjectReference{
						Kind: "DockerImage",
						Name: "quay.io/dummy-server/ocp-v4.0-art-dev@sha256:c8bc1d7bdf77538653ff1cd40d9bfa00f0",
					},
				}},
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the imageStream is empty")

		builder.errorMsg = "the imageStream 'name' cannot be empty"

		return builder
	}

	if namespace == "" {
		glog.V(100).Infof("The namespace of the imageStream is empty")

		builder.errorMsg = "the imageStream 'name' cannot be empty"

		return builder
	}

	return builder
}
