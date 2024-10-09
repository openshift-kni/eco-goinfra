package icsp

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	v1alpha1 "github.com/openshift/api/operator/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	defaultICSPName   = "test-icsp"
	defaultICSPSource = "test-source"
)

var (
	defaultICSPMirrors = []string{"test-mirror"}
	testSchemes        = []clients.SchemeAttacher{
		v1alpha1.Install,
	}
)

func TestNewICSPBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		source        string
		mirrors       []string
		client        bool
		expectedError string
	}{
		{
			name:          defaultICSPName,
			source:        defaultICSPSource,
			mirrors:       defaultICSPMirrors,
			client:        true,
			expectedError: "",
		},
		{
			name:          "",
			source:        defaultICSPSource,
			mirrors:       defaultICSPMirrors,
			client:        true,
			expectedError: "imageContentSourcePolicy 'name' cannot be empty",
		},
		{
			name:          defaultICSPName,
			source:        "",
			mirrors:       defaultICSPMirrors,
			client:        true,
			expectedError: "imageContentSourcePolicy 'source' cannot be empty",
		},
		{
			name:          defaultICSPName,
			source:        defaultICSPSource,
			mirrors:       nil,
			client:        true,
			expectedError: "imageContentSourcePolicy 'mirrors' cannot be empty",
		},
		{
			name:          defaultICSPName,
			source:        defaultICSPSource,
			mirrors:       defaultICSPMirrors,
			client:        false,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		testBuilder := NewICSPBuilder(testSettings, testCase.name, testCase.source, testCase.mirrors)

		if testCase.client {
			assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

			if testCase.expectedError == "" {
				assert.Equal(t, testCase.name, testBuilder.Definition.Name)
				assert.Equal(t, []v1alpha1.RepositoryDigestMirrors{{
					Source:  testCase.source,
					Mirrors: testCase.mirrors,
				}}, testBuilder.Definition.Spec.RepositoryDigestMirrors)
			}
		} else {
			assert.Nil(t, testBuilder)
		}
	}
}

func TestPullICSP(t *testing.T) {
	testCases := []struct {
		name                string
		addToRuntimeObjects bool
		client              bool
		expectedError       error
	}{
		{
			name:                defaultICSPName,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       nil,
		},
		{
			name:                "",
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("imageContentSourcePolicy 'name' cannot be empty"),
		},
		{
			name:                defaultICSPName,
			addToRuntimeObjects: false,
			client:              true,
			expectedError:       fmt.Errorf("imageContentSourcePolicy object %s does not exist", defaultICSPName),
		},
		{
			name:                defaultICSPName,
			addToRuntimeObjects: true,
			client:              false,
			expectedError:       fmt.Errorf("imageContentSourcePolicy 'apiClient' cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testMC := buildDummyICSP(testCase.name, defaultICSPSource, defaultICSPMirrors)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testMC)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		testBuilder, err := Pull(testSettings, testCase.name)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testMC.Name, testBuilder.Definition.Name)
		}
	}
}

func TestICSPGet(t *testing.T) {
	testCases := []struct {
		testBuilder   *ICSPBuilder
		expectedError string
	}{
		{
			testBuilder:   buildValidICSPBuilder(buildTestClientWithDummyICSP()),
			expectedError: "",
		},
		{
			testBuilder:   buildInvalidICSPBuilder(buildTestClientWithDummyICSP()),
			expectedError: "imageContentSourcePolicy 'mirrors' cannot be empty",
		},
		{
			testBuilder:   buildValidICSPBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: "imagecontentsourcepolicies.operator.openshift.io \"test-icsp\" not found",
		},
	}

	for _, testCase := range testCases {
		icsp, err := testCase.testBuilder.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, testCase.testBuilder.Definition.Name, icsp.Name)
		} else {
			assert.EqualError(t, err, testCase.expectedError)
		}
	}
}

func TestICSPExists(t *testing.T) {
	testCases := []struct {
		testBuilder *ICSPBuilder
		exists      bool
	}{
		{
			testBuilder: buildValidICSPBuilder(buildTestClientWithDummyICSP()),
			exists:      true,
		},
		{
			testBuilder: buildInvalidICSPBuilder(buildTestClientWithDummyICSP()),
			exists:      false,
		},
		{
			testBuilder: buildValidICSPBuilder(clients.GetTestClients(clients.TestClientParams{})),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

func TestICSPCreate(t *testing.T) {
	testCases := []struct {
		testBuilder   *ICSPBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidICSPBuilder(buildTestClientWithDummyICSP()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidICSPBuilder(buildTestClientWithDummyICSP()),
			expectedError: fmt.Errorf("imageContentSourcePolicy 'mirrors' cannot be empty"),
		},
		{
			testBuilder:   buildValidICSPBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		testBuilder, err := testCase.testBuilder.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testBuilder.Definition.Name, testBuilder.Object.Name)
		}
	}
}

func TestICSPDelete(t *testing.T) {
	testCases := []struct {
		testBuilder   *ICSPBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidICSPBuilder(buildTestClientWithDummyICSP()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidICSPBuilder(buildTestClientWithDummyICSP()),
			expectedError: fmt.Errorf("imageContentSourcePolicy 'mirrors' cannot be empty"),
		},
		{
			testBuilder:   buildValidICSPBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		err := testCase.testBuilder.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testBuilder.Object)
		}
	}
}

func TestICSPUpdate(t *testing.T) {
	testCases := []struct {
		testBuilder   *ICSPBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidICSPBuilder(buildTestClientWithDummyICSP()),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidICSPBuilder(buildTestClientWithDummyICSP()),
			expectedError: fmt.Errorf("imageContentSourcePolicy 'mirrors' cannot be empty"),
		},
		{
			testBuilder:   buildValidICSPBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("cannot update non-existent ImageContentSourcePolicy"),
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, 1, len(testCase.testBuilder.Definition.Spec.RepositoryDigestMirrors))
		testBuilder := testCase.testBuilder.WithRepositoryDigestMirror(defaultICSPSource, defaultICSPMirrors)

		testBuilder, err := testBuilder.Update()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, 2, len(testBuilder.Object.Spec.RepositoryDigestMirrors))
		}
	}
}

func TestICSPWithRepositoryDigestMirror(t *testing.T) {
	testCases := []struct {
		testBuilder   *ICSPBuilder
		source        string
		mirrors       []string
		expectedError string
	}{
		{
			testBuilder:   buildValidICSPBuilder(clients.GetTestClients(clients.TestClientParams{})),
			source:        defaultICSPSource,
			mirrors:       defaultICSPMirrors,
			expectedError: "",
		},
		{
			testBuilder:   buildInvalidICSPBuilder(clients.GetTestClients(clients.TestClientParams{})),
			source:        defaultICSPSource,
			mirrors:       defaultICSPMirrors,
			expectedError: "imageContentSourcePolicy 'mirrors' cannot be empty",
		},
		{
			testBuilder:   buildValidICSPBuilder(clients.GetTestClients(clients.TestClientParams{})),
			source:        "",
			mirrors:       defaultICSPMirrors,
			expectedError: "imageContentSourcePolicy 'source' cannot be empty",
		},
		{
			testBuilder:   buildValidICSPBuilder(clients.GetTestClients(clients.TestClientParams{})),
			source:        defaultICSPSource,
			mirrors:       nil,
			expectedError: "imageContentSourcePolicy 'mirrors' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := testCase.testBuilder.WithRepositoryDigestMirror(testCase.source, testCase.mirrors)
		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, v1alpha1.RepositoryDigestMirrors{
				Source:  testCase.source,
				Mirrors: testCase.mirrors,
			}, testBuilder.Definition.Spec.RepositoryDigestMirrors[1])
		}
	}
}

func TestICSPWithOptions(t *testing.T) {
	testCases := []struct {
		testBuilder   *ICSPBuilder
		options       AdditionalOptions
		expectedError string
	}{
		{
			testBuilder: buildValidICSPBuilder(clients.GetTestClients(clients.TestClientParams{})),
			options: func(builder *ICSPBuilder) (*ICSPBuilder, error) {
				builder.Definition.Spec.RepositoryDigestMirrors = nil

				return builder, nil
			},
			expectedError: "",
		},
		{
			testBuilder: buildInvalidICSPBuilder(clients.GetTestClients(clients.TestClientParams{})),
			options: func(builder *ICSPBuilder) (*ICSPBuilder, error) {
				return builder, nil
			},
			expectedError: "imageContentSourcePolicy 'mirrors' cannot be empty",
		},
		{
			testBuilder: buildValidICSPBuilder(clients.GetTestClients(clients.TestClientParams{})),
			options: func(builder *ICSPBuilder) (*ICSPBuilder, error) {
				return builder, fmt.Errorf("error adding additional options")
			},
			expectedError: "error adding additional options",
		},
	}

	for _, testCase := range testCases {
		testBuilder := testCase.testBuilder.WithOptions(testCase.options)
		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Empty(t, testBuilder.Definition.Spec.RepositoryDigestMirrors)
		}
	}
}

// buildDummyICSP returns an ImageContentSourcePolicy with the provided name and a single RepositoryDigestMirrors with
// the provided source and mirrors.
func buildDummyICSP(name, source string, mirrors []string) *v1alpha1.ImageContentSourcePolicy {
	return &v1alpha1.ImageContentSourcePolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.ImageContentSourcePolicySpec{
			RepositoryDigestMirrors: []v1alpha1.RepositoryDigestMirrors{{
				Source:  source,
				Mirrors: mirrors,
			}},
		},
	}
}

// buildTestClientWithDummyICSP returns a client with a dummy ImageContentSourcePolicy with the default name, source,
// and mirrors.
func buildTestClientWithDummyICSP() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyICSP(defaultICSPName, defaultICSPSource, defaultICSPMirrors),
		},
		SchemeAttachers: testSchemes,
	})
}

// buildValidICSPBuilder returns a valid ICSPBuilder for testing.
func buildValidICSPBuilder(apiClient *clients.Settings) *ICSPBuilder {
	return NewICSPBuilder(apiClient, defaultICSPName, defaultICSPSource, defaultICSPMirrors)
}

// buildInvalidICSPBuilder returns an invalid ICSPBuilder for testing, missing the mirrors.
func buildInvalidICSPBuilder(apiClient *clients.Settings) *ICSPBuilder {
	return NewICSPBuilder(apiClient, defaultICSPName, defaultICSPSource, nil)
}
