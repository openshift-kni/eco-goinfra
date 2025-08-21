package certificate

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	certificatesv1 "k8s.io/api/certificates/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var testSchemes = []clients.SchemeAttacher{
	certificatesv1.AddToScheme,
}

const (
	defaultSigningRequestName = "test-signing-request"
)

func TestPullSigningRequest(t *testing.T) {
	testCases := []struct {
		name                string
		addToRuntimeObjects bool
		client              bool
		expectedError       error
	}{
		{
			name:                defaultSigningRequestName,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       nil,
		},
		{
			name:                "",
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("certificateSigningRequest 'name' cannot be empty"),
		},
		{
			name:                defaultSigningRequestName,
			addToRuntimeObjects: false,
			client:              true,
			expectedError:       fmt.Errorf("certificateSigningRequest %s does not exist", defaultSigningRequestName),
		},
		{
			name:                defaultSigningRequestName,
			addToRuntimeObjects: true,
			client:              false,
			expectedError:       fmt.Errorf("certificateSigniingRequest apiClient cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, buildDummySigningRequest(testCase.name))
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		signingRequestBuilder, err := PullSigningRequest(testSettings, testCase.name)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, signingRequestBuilder.Object.Name)
		}
	}
}

func TestSigningRequestGet(t *testing.T) {
	testCases := []struct {
		testBuilder   *SigningRequestBuilder
		expectedError string
	}{
		{
			testBuilder:   newSigningRequestBuilder(buildTestClientWithDummySigningRequest()),
			expectedError: "",
		},
		{
			testBuilder: newSigningRequestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Sprintf("certificatesigningrequests.certificates.k8s.io \"%s\" not found",
				defaultSigningRequestName),
		},
	}

	for _, testCase := range testCases {
		signingRequest, err := testCase.testBuilder.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, testCase.testBuilder.Definition.Name, signingRequest.Name)
		} else {
			assert.EqualError(t, err, testCase.expectedError)
		}
	}
}

func TestSigningRequestExists(t *testing.T) {
	testCases := []struct {
		testBuilder *SigningRequestBuilder
		exists      bool
	}{
		{
			testBuilder: newSigningRequestBuilder(buildTestClientWithDummySigningRequest()),
			exists:      true,
		},
		{
			testBuilder: newSigningRequestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

func TestSigningRequestCreate(t *testing.T) {
	testCases := []struct {
		testBuilder   *SigningRequestBuilder
		expectedError error
	}{
		{
			testBuilder:   newSigningRequestBuilder(buildTestClientWithDummySigningRequest()),
			expectedError: nil,
		},
		{
			testBuilder:   newSigningRequestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		signingRequestBuilder, err := testCase.testBuilder.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, signingRequestBuilder.Definition.Name, signingRequestBuilder.Object.Name)
		}
	}
}

func TestSigningRequestDelete(t *testing.T) {
	testCases := []struct {
		testBuilder   *SigningRequestBuilder
		expectedError error
	}{
		{
			testBuilder:   newSigningRequestBuilder(buildTestClientWithDummySigningRequest()),
			expectedError: nil,
		},
		{
			testBuilder:   newSigningRequestBuilder(clients.GetTestClients(clients.TestClientParams{})),
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

func TestSigningRequestValidate(t *testing.T) {
	testCases := []struct {
		builderNil    bool
		definitionNil bool
		apiClientNil  bool
		expectedError error
	}{
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: nil,
		},
		{
			builderNil:    true,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: fmt.Errorf("error: received nil certificateSigningRequest builder"),
		},
		{
			builderNil:    false,
			definitionNil: true,
			apiClientNil:  false,
			expectedError: fmt.Errorf("can not redefine the undefined certificateSigningRequest"),
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  true,
			expectedError: fmt.Errorf("certificateSigningRequest builder cannot have nil apiClient"),
		},
	}

	for _, testCase := range testCases {
		signingRequestBuilder := newSigningRequestBuilder(buildTestClientWithDummySigningRequest())

		if testCase.builderNil {
			signingRequestBuilder = nil
		}

		if testCase.definitionNil {
			signingRequestBuilder.Definition = nil
		}

		if testCase.apiClientNil {
			signingRequestBuilder.apiClient = nil
		}

		valid, err := signingRequestBuilder.validate()
		assert.Equal(t, testCase.expectedError, err)
		assert.Equal(t, testCase.expectedError == nil, valid)
	}
}

// buildDummySigningRequest returns a dummy CertificateSigningRequest object with the given name.
func buildDummySigningRequest(name string) *certificatesv1.CertificateSigningRequest {
	return &certificatesv1.CertificateSigningRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

// buildTestClientWithDummySigningRequest returns a clients.Settings object with a dummy CertificateSigningRequest
// object using the default name.
func buildTestClientWithDummySigningRequest() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummySigningRequest(defaultSigningRequestName),
		},
		SchemeAttachers: testSchemes,
	})
}

func newSigningRequestBuilder(apiClient *clients.Settings) *SigningRequestBuilder {
	return &SigningRequestBuilder{
		Definition: buildDummySigningRequest(defaultSigningRequestName),
		apiClient:  apiClient.Client,
	}
}
