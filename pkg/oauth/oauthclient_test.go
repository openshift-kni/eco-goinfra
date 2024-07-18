package oauth

import (
	"fmt"
	"testing"

	oauthv1 "github.com/openshift/api/oauth/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
)

var (
	oauthClientName = "oauth"
)

func TestOAuthClientPull(t *testing.T) {
	testCases := []struct {
		expectedError       error
		addToRuntimeObjects bool
		oauthTestClientName string
	}{
		{
			expectedError:       nil,
			addToRuntimeObjects: true,
			oauthTestClientName: oauthClientName,
		},

		{
			expectedError:       fmt.Errorf("error: OAuthClient 'name' cannot be empty"),
			oauthTestClientName: "",
			addToRuntimeObjects: true,
		},
		{
			expectedError:       fmt.Errorf("error: OAuthClient object %s not found", oauthClientName),
			oauthTestClientName: oauthClientName,
			addToRuntimeObjects: false,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testOAuthClient := generateOAuthClient(testCase.oauthTestClientName)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testOAuthClient)
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		// Test the PullOAuthClient function
		builderResult, err := PullOAuthClient(testSettings, testCase.oauthTestClientName)

		// Check the error
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil {
			assert.NotNil(t, builderResult)
		}
	}
}

func TestOAuthClientUpdate(t *testing.T) {
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

		testOAuthClient := generateOAuthClient(oauthClientName)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testOAuthClient)
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		oauthClientBuilder, err := PullOAuthClient(testSettings, oauthClientName)
		assert.Nil(t, err)

		// Create a change in the builder
		annotation := map[string]string{"test": "test"}
		oauthClientBuilder.Definition.Annotations = annotation

		// Test the Update function
		builderResult, err := oauthClientBuilder.Update()

		assert.Equal(t, err, testCase.expectedError)

		// Validate that the resource was updated
		if testCase.expectedError == nil {
			assert.NotNil(t, builderResult)
			assert.Equal(t, builderResult.Object.Annotations, annotation)
		}
	}
}

func TestOAuthClientCreate(t *testing.T) {
	testCases := []struct {
		addToRuntimeObjects bool
	}{
		{
			addToRuntimeObjects: true,
		},
		{
			addToRuntimeObjects: false,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, &oauthv1.OAuthClient{
				ObjectMeta: metav1.ObjectMeta{
					Name: oauthClientName,
				},
			})
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		testBuilder, client := generateOAuthClientBuilder(testSettings, oauthClientName)

		// Test the Create function
		_, err := testBuilder.Create()
		assert.Nil(t, err)

		// Assert that the object actually exists
		_, err = PullOAuthClient(client, oauthClientName)
		assert.Nil(t, err)
	}
}

func TestOAuthClientDelete(t *testing.T) {
	testCases := []struct {
		addToRuntimeObjects bool
	}{
		{
			addToRuntimeObjects: true,
		},
		{
			addToRuntimeObjects: false,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, &oauthv1.OAuthClient{
				ObjectMeta: metav1.ObjectMeta{
					Name: oauthClientName,
				},
			})
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		testBuilder, client := generateOAuthClientBuilder(testSettings, oauthClientName)

		// Testing the Delete function
		err := testBuilder.Delete()
		assert.Nil(t, err)

		// Assert that the object actually does not exist
		_, err = PullOAuthClient(client, oauthClientName)
		assert.NotNil(t, err)
	}
}

func TestOAuthClientGet(t *testing.T) {
	testCases := []struct {
		expectedError       error
		addToRuntimeObjects bool
	}{
		{
			expectedError:       nil,
			addToRuntimeObjects: true,
		},
		{
			expectedError:       fmt.Errorf("error: received nil OAuthClient builder"),
			addToRuntimeObjects: false,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testOAuthClient := generateOAuthClient(oauthClientName)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testOAuthClient)
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		oauthClientBuilder, err := PullOAuthClient(testSettings, oauthClientName)
		if testCase.expectedError == nil {
			assert.Nil(t, err)
		}

		// Test the Get function
		builderResult, err := oauthClientBuilder.Get()

		// Check the error
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil {
			assert.NotNil(t, builderResult)
		}
	}
}

func TestOAuthClientExists(t *testing.T) {
	testCases := []struct {
		expectedError       error
		addToRuntimeObjects bool
	}{
		{
			addToRuntimeObjects: true,
		},
		{
			addToRuntimeObjects: false,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, &oauthv1.OAuthClient{
				ObjectMeta: metav1.ObjectMeta{
					Name: oauthClientName,
				},
			})
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		testOAuthClientBuilder, _ := generateOAuthClientBuilder(testSettings, oauthClientName)

		// Test the Exists function
		exists := testOAuthClientBuilder.Exists()
		assert.Equal(t, testCase.addToRuntimeObjects, exists)
	}
}

func TestMultiClusterHubValidate(t *testing.T) {
	testCases := []struct {
		builderNil    bool
		definitionNil bool
		apiClientNil  bool
		expectedError error
	}{
		{
			builderNil:    true,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: fmt.Errorf("error: received nil OAuthClient builder"),
		},
		{
			builderNil:    false,
			definitionNil: true,
			apiClientNil:  false,
			expectedError: fmt.Errorf("can not redefine the undefined OAuthClient"),
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  true,
			expectedError: fmt.Errorf("error: OAuthClient builder cannot have nil apiClient"),
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testMultiClusterHub := generateOAuthClient(oauthClientName)

		runtimeObjects = append(runtimeObjects, testMultiClusterHub)

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects,
		})

		testBuilder, err := PullOAuthClient(testSettings, oauthClientName)

		if testCase.expectedError == nil {
			assert.Nil(t, err)
		}

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
		if testCase.expectedError != nil {
			assert.NotNil(t, err)
			assert.Equal(t, testCase.expectedError, err)
			assert.False(t, result)
		} else {
			assert.Nil(t, err)
			assert.True(t, result)
		}
	}
}

func generateOAuthClient(name string) *oauthv1.OAuthClient {
	return &oauthv1.OAuthClient{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func generateOAuthClientBuilder(testSettings *clients.Settings, name string) (*OAuthClientBuilder,
	*clients.Settings) {
	return &OAuthClientBuilder{
		apiClient: testSettings.Client,
		Definition: &oauthv1.OAuthClient{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		},
	}, testSettings
}
