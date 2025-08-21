package idms

import (
	"fmt"
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
)

const (
	TestIDMS = "test-image-digest-mirror-set"
)

var testSchemes = []clients.SchemeAttacher{
	configv1.Install,
}

func TestNewIDMSBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		mirror        configv1.ImageDigestMirrors
		client        bool
		expectedError string
	}{
		{
			name: TestIDMS,
			mirror: configv1.ImageDigestMirrors{
				Mirrors: []configv1.ImageMirror{
					"cloned.registry.org",
				},
				Source: "registry.org",
			},
			client:        true,
			expectedError: "",
		},
		{
			name: "",
			mirror: configv1.ImageDigestMirrors{
				Mirrors: []configv1.ImageMirror{
					"cloned.registry.org",
				},
				Source: "registry.org",
			},
			client:        true,
			expectedError: "imagedigestmirrorset 'name' cannot be empty",
		},
		{
			name: TestIDMS,
			mirror: configv1.ImageDigestMirrors{
				Mirrors: []configv1.ImageMirror{
					"cloned.registry.org",
				},
				Source: "registry.org",
			},
			client:        false,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		var (
			client *clients.Settings
		)

		if testCase.client {
			client = clients.GetTestClients(clients.TestClientParams{})
		}

		testBuilder := NewBuilder(
			client, testCase.name, testCase.mirror)

		if testCase.client {
			assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

			if testCase.expectedError == "" {
				assert.Equal(t, testCase.name, testBuilder.Definition.Name)
				assert.Equal(t, testCase.mirror, testBuilder.Definition.Spec.ImageDigestMirrors[0])
			}
		} else {
			assert.Nil(t, testBuilder)
		}
	}
}

func TestIDMSPull(t *testing.T) {
	testCases := []struct {
		name          string
		client        bool
		exists        bool
		expectedError error
	}{
		{
			name:          TestIDMS,
			client:        true,
			exists:        true,
			expectedError: nil,
		},
		{
			name:          "",
			client:        true,
			exists:        true,
			expectedError: fmt.Errorf("imagedigestmirrorset 'name' cannot be empty"),
		},
		{
			name:          TestIDMS,
			client:        false,
			exists:        true,
			expectedError: fmt.Errorf("apiClient cannot be nil"),
		},
		{
			name:   TestIDMS,
			client: true,
			exists: false,
			expectedError: fmt.Errorf(
				"imagedigestmirrorset object %s does not exist",
				TestIDMS),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testClient     *clients.Settings
		)

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateImagDigestMirrorSet())
		}

		if testCase.client {
			testClient = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		testBuilder, err := Pull(testClient, testCase.name)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, testBuilder.Definition.Name)
		}
	}
}

func TestIDMSWithMirror(t *testing.T) {
	testCases := []struct {
		mirror           configv1.ImageDigestMirrors
		expectedErrorMsg string
	}{
		{
			mirror: configv1.ImageDigestMirrors{
				Mirrors: []configv1.ImageMirror{
					"cloned.test.org",
				},
				Source: "test.org",
			},
			expectedErrorMsg: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateImagDigestMirrorSetBuilder()

		testBuilder.WithMirror(testCase.mirror)
		assert.Equal(t, testCase.expectedErrorMsg, testBuilder.errorMsg)

		if testCase.expectedErrorMsg == "" {
			assert.Equal(t, testCase.mirror, testBuilder.Definition.Spec.ImageDigestMirrors[1])
		}
	}
}

func TestIDMSGet(t *testing.T) {
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
		var (
			runtimeObjects []runtime.Object
		)

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateImagDigestMirrorSet())
		}

		testBuilder := generateIDMSBuilderWithFakeObjects(runtimeObjects)

		aci, err := testBuilder.Get()
		if testCase.exists {
			assert.Nil(t, err)
			assert.NotNil(t, aci)
		} else {
			assert.NotNil(t, err)
			assert.Nil(t, aci)
		}
	}
}

func TestIDMSCreate(t *testing.T) {
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
		var (
			runtimeObjects []runtime.Object
		)

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateImagDigestMirrorSet())
		}

		testBuilder := generateIDMSBuilderWithFakeObjects(runtimeObjects)

		result, err := testBuilder.Create()
		assert.Nil(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, TestIDMS, result.Definition.Name)
	}
}
func TestIDMSUpdate(t *testing.T) {
	testCases := []struct {
		exists        bool
		expectedError error
	}{
		{
			exists:        true,
			expectedError: nil,
		},
		{
			exists:        false,
			expectedError: fmt.Errorf("cannot update non-existent imagedigestmirrorset"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateImagDigestMirrorSet())
		}

		testBuilder := generateIDMSBuilderWithFakeObjects(runtimeObjects)

		testBuilder.Definition.Spec.ImageDigestMirrors[0].Source = "new.registry.org"

		idms, err := testBuilder.Update(true)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, idms.Object.Spec.ImageDigestMirrors[0].Source, "new.registry.org")
		}
	}
}
func TestIDMSDelete(t *testing.T) {
	testCases := []struct {
		exists        bool
		expectedError error
	}{
		{
			exists:        true,
			expectedError: nil,
		},
		{
			exists:        false,
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateImagDigestMirrorSet())
		}

		testBuilder := generateIDMSBuilderWithFakeObjects(runtimeObjects)

		err := testBuilder.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testBuilder.Object)
		}
	}
}
func TestIDMSExists(t *testing.T) {
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
		var (
			runtimeObjects []runtime.Object
		)

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateImagDigestMirrorSet())
		}

		testBuilder := generateIDMSBuilderWithFakeObjects(runtimeObjects)

		assert.Equal(t, testCase.exists, testBuilder.Exists())
	}
}

func TestIDMSValidate(t *testing.T) {
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
			expectedError: "error: received nil ImageDigestMirrorSet builder",
		},
		{
			builderNil:    false,
			definitionNil: true,
			apiClientNil:  false,
			expectedError: "can not redefine the undefined ImageDigestMirrorSet",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  true,
			expectedError: "ImageDigestMirrorSet builder cannot have nil apiClient",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateIDMSBuilderWithFakeObjects([]runtime.Object{})

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

func generateIDMSBuilderWithFakeObjects(objects []runtime.Object) *Builder {
	return &Builder{
		apiClient: clients.GetTestClients(
			clients.TestClientParams{K8sMockObjects: objects, SchemeAttachers: testSchemes}).Client,
		Definition: generateImagDigestMirrorSet(),
	}
}

func generateImagDigestMirrorSetBuilder() *Builder {
	return &Builder{
		apiClient:  clients.GetTestClients(clients.TestClientParams{}).Client,
		Definition: generateImagDigestMirrorSet(),
	}
}

func generateImagDigestMirrorSet() *configv1.ImageDigestMirrorSet {
	return &configv1.ImageDigestMirrorSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: TestIDMS,
		},
		Spec: configv1.ImageDigestMirrorSetSpec{
			ImageDigestMirrors: []configv1.ImageDigestMirrors{
				{
					Mirrors: []configv1.ImageMirror{
						"cloned.registry.org",
					},
					Source: "registry.org",
				},
			},
		},
	}
}
