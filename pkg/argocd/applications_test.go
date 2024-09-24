package argocd

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	argocdtypes "github.com/openshift-kni/eco-goinfra/pkg/schemes/argocd/argocdtypes/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	defaultApplicationName   = "application-name"
	defaultApplicationNsName = "application-ns-name"
)

var (
	defaultApplicationCondition = argocdtypes.ApplicationCondition{
		Type:    argocdtypes.ApplicationConditionSyncError,
		Message: "test-message",
	}
	appsTestSchemes = []clients.SchemeAttacher{
		argocdtypes.AddToScheme,
	}
)

func TestPullApplication(t *testing.T) {
	generateApplication := func(name, namespace string) *argocdtypes.Application {
		return &argocdtypes.Application{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: argocdtypes.ApplicationSpec{},
		}
	}

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

		testApplication := generateApplication(testCase.name, testCase.namespace)
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
		assert.Equal(t, testCase.expectedStatus, exist)
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
		testApplicationBuilder *ApplicationBuilder
		expectedError          error
	}{
		{
			testApplicationBuilder: buildValidApplicationBuilder(buildApplicationTestClientWithDummyObject()),
			expectedError:          nil,
		},
		{
			testApplicationBuilder: buildValidApplicationBuilder(buildApplicationTestClientWithScheme()),
			expectedError:          fmt.Errorf("cannot update non-existent Application"),
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, testCase.testApplicationBuilder.Definition.Spec.Project, "")
		testCase.testApplicationBuilder.Definition.Spec.Project = "test"

		application, err := testCase.testApplicationBuilder.Update(false)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, application.Object.Spec.Project, "test")
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

func TestImageRegistryWaitForCondition(t *testing.T) {
	testCases := []struct {
		exists        bool
		conditionMet  bool
		expectedError error
	}{
		{
			exists:        true,
			conditionMet:  true,
			expectedError: nil,
		},
		{
			exists:       false,
			conditionMet: true,
			expectedError: fmt.Errorf(
				"application object %s in namespace %s does not exist", defaultApplicationName, defaultApplicationNsName),
		},
		{
			exists:        true,
			conditionMet:  false,
			expectedError: context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		if testCase.exists {
			application := buildDummyApplication(defaultApplicationName, defaultApplicationNsName)

			if testCase.conditionMet {
				application.Status.Conditions = append(application.Status.Conditions, defaultApplicationCondition)
			}

			runtimeObjects = append(runtimeObjects, application)
		}

		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  runtimeObjects,
			SchemeAttachers: appsTestSchemes,
		})

		testBuilder := buildValidApplicationBuilder(testSettings)

		_, err := testBuilder.WaitForCondition(defaultApplicationCondition, time.Second)
		assert.Equal(t, testCase.expectedError, err)
	}
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
	return append([]runtime.Object{}, &argocdtypes.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultApplicationName,
			Namespace: defaultApplicationNsName,
		},

		Spec: argocdtypes.ApplicationSpec{},
	})
}

func buildDummyApplication(name, namespace string) *argocdtypes.Application {
	return &argocdtypes.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},

		Spec: argocdtypes.ApplicationSpec{
			Source: &argocdtypes.ApplicationSource{},
		},
	}
}
