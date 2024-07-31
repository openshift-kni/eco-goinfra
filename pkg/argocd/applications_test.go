package argocd

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	appsv1alpha1 "github.com/openshift-kni/eco-goinfra/pkg/schemes/argocd/argocdtypes/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	appsTestSchemes = []clients.SchemeAttacher{
		appsv1alpha1.AddToScheme,
	}
	defaultApplicationName   = "application-name"
	defaultApplicationNsName = "application-ns-name"
)

func TestPullApplication(t *testing.T) {
	testCases := []struct {
		name                string
		namespace           string
		addToRuntimeObjects bool
		expectedError       error
		client              bool
	}{
		{
			name:                "applicationdtest",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("application 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                "applicationtest",
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("application 'namespace' cannot be empty"),
			client:              true,
		},
		{
			name:                "applicationtest",
			namespace:           "test-namespace",
			addToRuntimeObjects: false,
			expectedError:       fmt.Errorf("application object applicationtest does not exist in namespace test-namespace"),
			client:              true,
		},
		{
			name:                "applicationtest",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("application 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testApplication := buildDummyApplication(testCase.name, testCase.namespace)
		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testApplication)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: appsTestSchemes,
			})
		}

		builderResult, err := PullApplication(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, builderResult.Object.Name)
			assert.Equal(t, testCase.namespace, builderResult.Object.Namespace)
		}
	}
}

func TestApplicationExist(t *testing.T) {
	testCases := []struct {
		testApplicationBuilder *ApplicationBuilder
		expectedStatus         bool
	}{
		{
			testApplicationBuilder: buildValidApplicationBuilder(buildApplicationTestClientWithDummyObject()),
			expectedStatus:         true,
		},
		{
			testApplicationBuilder: buildValidApplicationBuilder(buildApplicationTestClientWithScheme()),
			expectedStatus:         false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testApplicationBuilder.Exists()
		assert.Equal(t, exist, testCase.expectedStatus)
	}
}

func TestApplicationGet(t *testing.T) {
	testCases := []struct {
		testApplicationBuilder *ApplicationBuilder
		expectedError          error
	}{
		{
			testApplicationBuilder: buildValidApplicationBuilder(buildApplicationTestClientWithDummyObject()),
			expectedError:          nil,
		},
		{
			testApplicationBuilder: buildValidApplicationBuilder(buildApplicationTestClientWithScheme()),
			expectedError:          fmt.Errorf("applications.argoproj.io \"application-name\" not found"),
		},
	}

	for _, testCase := range testCases {
		application, err := testCase.testApplicationBuilder.Get()
		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}

		if testCase.expectedError == nil {
			assert.Equal(t, application.Name, testCase.testApplicationBuilder.Definition.Name)
			assert.Equal(t, application.Namespace, testCase.testApplicationBuilder.Definition.Namespace)
		}
	}
}

func TestApplicationUpdate(t *testing.T) {
	testCases := []struct {
		alreadyExists     bool
		force             bool
		expectedErrorText string
	}{
		{
			alreadyExists:     false,
			force:             false,
			expectedErrorText: "cannot update non-existent application",
		},
		{
			alreadyExists:     true,
			force:             false,
			expectedErrorText: "",
		},
		{
			alreadyExists:     false,
			force:             true,
			expectedErrorText: "cannot update non-existent application",
		},
		{
			alreadyExists:     true,
			force:             true,
			expectedErrorText: "",
		},
	}

	for _, testCase := range testCases {
		kacBuilder := buildValidApplicationBuilder(buildApplicationTestClientWithScheme())

		// Create the builder rather than just adding it to the client so that the proper metadata is added and
		// the update will not fail.
		if testCase.alreadyExists {
			var err error
			kacBuilder, err = kacBuilder.Create()

			assert.Nil(t, err)
			assert.True(t, kacBuilder.Exists())
		}

		assert.NotNil(t, kacBuilder.Definition)
		assert.Empty(t, kacBuilder.Definition.Spec.Project)

		kacBuilder.Definition.Spec.Project = "test"

		kacBuilder, err := kacBuilder.Update(testCase.force)

		if testCase.expectedErrorText == "" {
			assert.Nil(t, err)
			assert.Equal(t, "test", kacBuilder.Object.Spec.Project)
		} else {
			assert.EqualError(t, err, testCase.expectedErrorText)
		}
	}
}

func TestApplicationDelete(t *testing.T) {
	testCases := []struct {
		testApplicationBuilder *ApplicationBuilder
		expectedError          error
	}{
		{
			testApplicationBuilder: buildValidApplicationBuilder(buildApplicationTestClientWithDummyObject()),
			expectedError:          nil,
		},
		{
			testApplicationBuilder: buildValidApplicationBuilder(buildApplicationTestClientWithScheme()),
			expectedError:          nil,
		},
	}

	for _, testCase := range testCases {
		_, err := testCase.testApplicationBuilder.Delete()

		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedError, err)
		}

		assert.Nil(t, testCase.testApplicationBuilder.Object)
	}
}

func TestApplicationCreate(t *testing.T) {
	testCases := []struct {
		testApplicationBuilder *ApplicationBuilder
		expectedError          error
	}{
		{
			testApplicationBuilder: buildValidApplicationBuilder(buildApplicationTestClientWithDummyObject()),
			expectedError:          nil,
		},
		{
			testApplicationBuilder: buildValidApplicationBuilder(buildApplicationTestClientWithScheme()),
			expectedError:          nil,
		},
	}

	for _, testCase := range testCases {
		_, err := testCase.testApplicationBuilder.Create()

		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedError, err)
		}

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.testApplicationBuilder.Definition.Name, testCase.testApplicationBuilder.Object.Name)
			assert.Equal(
				t, testCase.testApplicationBuilder.Definition.Namespace, testCase.testApplicationBuilder.Object.Namespace)
		}
	}
}

func TestApplicationWithGitDetails(t *testing.T) {
	testCases := []struct {
		testApplicationBuilder *ApplicationBuilder
		gitRepo                string
		gitBranch              string
		gitPath                string
		expectedError          string
	}{
		{
			testApplicationBuilder: buildValidApplicationBuilder(buildApplicationTestClientWithDummyObject()),
			gitRepo:                "http://test.git",
			gitBranch:              "main",
			gitPath:                "./dir/www/repo",
			expectedError:          "",
		},
		{
			testApplicationBuilder: buildValidApplicationBuilder(buildApplicationTestClientWithDummyObject()),
			gitRepo:                "",
			gitBranch:              "main",
			gitPath:                "./dir/www/repo",
			expectedError:          "'gitRepo' parameter is empty",
		},
		{
			testApplicationBuilder: buildValidApplicationBuilder(buildApplicationTestClientWithDummyObject()),
			gitRepo:                "http://test.git",
			gitBranch:              "",
			gitPath:                "./dir/www/repo",
			expectedError:          "'gitBranch' parameter is empty",
		},
		{
			testApplicationBuilder: buildValidApplicationBuilder(buildApplicationTestClientWithDummyObject()),
			gitRepo:                "http://test.git",
			gitBranch:              "main",
			gitPath:                "",
			expectedError:          "'gitPath' parameter is empty",
		},
	}

	for _, testCase := range testCases {
		applicationBuilder := testCase.testApplicationBuilder.WithGitDetails(
			testCase.gitRepo, testCase.gitBranch, testCase.gitPath)
		assert.Equal(t, testCase.expectedError, applicationBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, applicationBuilder.Definition.Spec.Source.Path, testCase.gitPath)
			assert.Equal(t, applicationBuilder.Definition.Spec.Source.RepoURL, testCase.gitRepo)
			assert.Equal(t, applicationBuilder.Definition.Spec.Source.Path, testCase.gitPath)
		}
	}
}

func TestApplicationGVR(t *testing.T) {
	assert.Equal(t, GetApplicationsGVR(),
		schema.GroupVersionResource{
			Group:    appsv1alpha1.SchemeGroupVersion.Group,
			Version:  appsv1alpha1.SchemeGroupVersion.Version,
			Resource: "applications",
		})
}

func buildValidApplicationBuilder(apiClient *clients.Settings) *ApplicationBuilder {
	return &ApplicationBuilder{
		apiClient:  apiClient.Client,
		Definition: buildDummyApplication(defaultApplicationName, defaultApplicationNsName),
	}
}

func buildApplicationTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyApplicationRuntime(),
		SchemeAttachers: appsTestSchemes,
	})
}

func buildApplicationTestClientWithScheme() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		SchemeAttachers: appsTestSchemes,
	})
}

func buildDummyApplicationRuntime() []runtime.Object {
	return append([]runtime.Object{}, &appsv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultApplicationName,
			Namespace: defaultApplicationNsName,
		},

		Spec: appsv1alpha1.ApplicationSpec{},
	})
}

func buildDummyApplication(name, namespace string) *appsv1alpha1.Application {
	return &appsv1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},

		Spec: appsv1alpha1.ApplicationSpec{
			Source: &appsv1alpha1.ApplicationSource{},
		},
	}
}
