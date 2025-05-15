package argocd

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
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

func TestApplicationWithGitPathAppended(t *testing.T) {
	const testPath = "test/path"

	testCases := []struct {
		name                   string
		testApplicationBuilder *ApplicationBuilder
		hasSource              bool
		elements               []string
		expectedPath           string
		expectedError          string
	}{
		{
			name:                   "valid-builder-with-source",
			testApplicationBuilder: buildValidApplicationBuilder(buildApplicationTestClientWithScheme()),
			hasSource:              true,
			elements:               []string{"element1", "element2"},
			expectedPath:           fmt.Sprintf("%s/%s/%s", testPath, "element1", "element2"),
			expectedError:          "",
		},
		{
			name:                   "no-source",
			testApplicationBuilder: buildValidApplicationBuilder(buildApplicationTestClientWithScheme()),
			hasSource:              false,
			elements:               []string{"element"},
			expectedPath:           "",
			expectedError:          "cannot append to git path because the source is nil",
		},
		{
			name:                   "no-elements",
			testApplicationBuilder: buildValidApplicationBuilder(buildApplicationTestClientWithScheme()),
			hasSource:              true,
			elements:               []string{},
			expectedPath:           testPath,
			expectedError:          "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if testCase.hasSource {
				testCase.testApplicationBuilder.Definition.Spec.Source = &argocdtypes.ApplicationSource{
					Path: testPath,
				}
			} else {
				// buildDummyAppplication already sets the source so we must reset it to nil.
				testCase.testApplicationBuilder.Definition.Spec.Source = nil
			}

			applicationBuilder := testCase.testApplicationBuilder.WithGitPathAppended(testCase.elements...)
			assert.Equal(t, testCase.expectedError, applicationBuilder.errorMsg)

			if testCase.expectedError == "" {
				assert.Equal(t, testCase.expectedPath, applicationBuilder.Definition.Spec.Source.Path)
			}
		})
	}
}

func TestApplicationWaitForCondition(t *testing.T) {
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

func TestApplicationDoesGitPathExist(t *testing.T) {
	testCases := []struct {
		name      string
		hasSource bool
		validURL  bool
		exists    bool
	}{
		{
			name:      "exists",
			hasSource: true,
			validURL:  true,
			exists:    true,
		},
		{
			name:      "no-source",
			hasSource: false,
			validURL:  true,
			exists:    false,
		},
		{
			name:      "invalid-url",
			hasSource: true,
			validURL:  false,
			exists:    false,
		},
		{
			name:      "does-not-exist",
			hasSource: true,
			validURL:  true,
			exists:    false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				requestedPath   string
				requestedMethod string
			)

			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				requestedPath = request.URL.Path
				requestedMethod = request.Method

				if testCase.exists {
					writer.WriteHeader(http.StatusOK)

					return
				}

				writer.WriteHeader(http.StatusNotFound)
			}))
			defer server.Close()

			var serverURL string

			if testCase.validURL {
				serverURL = server.URL
			} else {
				serverURL = "invalid-url"
			}

			testBuilder := buildValidApplicationBuilder(buildApplicationTestClientWithScheme())

			if testCase.hasSource {
				testBuilder.Definition.Spec.Source = &argocdtypes.ApplicationSource{
					RepoURL:        serverURL,
					Path:           "some/path",
					TargetRevision: "main",
				}
			}

			exists := testBuilder.DoesGitPathExist("test")
			assert.Equal(t, testCase.exists, exists)

			if requestedMethod != "" {
				assert.Equal(t, http.MethodHead, requestedMethod)
			}

			if requestedPath != "" {
				assert.Equal(t, "/raw/main/some/path/test/kustomization.yaml", requestedPath)
			}
		})
	}
}

func TestApplicationWaitForSourceUpdate(t *testing.T) {
	var expectedSource = argocdtypes.ApplicationSource{
		TargetRevision: "main",
	}

	testCases := []struct {
		name          string
		sourceExists  bool
		sourceUpdated bool
		synced        bool
		expectSynced  bool
		expectedError error
	}{
		{
			name:          "source-synced",
			sourceExists:  true,
			sourceUpdated: true,
			synced:        true,
			expectSynced:  true,
			expectedError: nil,
		},
		{
			name:          "source-not-updated",
			sourceExists:  true,
			sourceUpdated: false,
			synced:        true,
			expectSynced:  true,
			expectedError: context.DeadlineExceeded,
		},
		{
			name:          "source-updated-not-synced",
			sourceExists:  true,
			sourceUpdated: true,
			synced:        false,
			expectSynced:  true,
			expectedError: context.DeadlineExceeded,
		},
		{
			name:          "source-not-synced-expect-synced-false",
			sourceExists:  true,
			sourceUpdated: true,
			synced:        false,
			expectSynced:  false,
			expectedError: nil,
		},
		{
			name:          "source-does-not-exist",
			sourceExists:  false,
			sourceUpdated: true,
			synced:        true,
			expectSynced:  true,
			expectedError: context.DeadlineExceeded,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			testApp := buildDummyApplication(defaultApplicationName, defaultApplicationNsName)

			if !testCase.sourceExists {
				testApp.Spec.Source = nil
			} else {
				testApp.Spec.Source = &expectedSource
			}

			if testCase.sourceUpdated {
				testApp.Status.Sync.ComparedTo.Source = expectedSource
			}

			if testCase.synced {
				testApp.Status.Sync.Status = argocdtypes.SyncStatusCodeSynced
			}

			testBuilder := buildValidApplicationBuilder(clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  []runtime.Object{testApp},
				SchemeAttachers: appsTestSchemes,
			}))

			err := testBuilder.WaitForSourceUpdate(testCase.expectSynced, time.Second)
			assert.Equal(t, testCase.expectedError, err)
		})
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
