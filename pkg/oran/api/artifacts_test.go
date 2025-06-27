package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/openshift-kni/eco-goinfra/pkg/oran/api/filter"
	"github.com/openshift-kni/eco-goinfra/pkg/oran/api/internal/artifacts"
	"github.com/openshift-kni/eco-goinfra/pkg/oran/api/internal/common"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"
)

var (
	// dummyManagedInfrastructureTemplate is a test template for use in tests.
	dummyManagedInfrastructureTemplate = artifacts.ManagedInfrastructureTemplate{
		ArtifactResourceId: uuid.New(),
		Name:               "test-template",
		Description:        "Test template description",
		Version:            "v1.0.0",
		ParameterSchema:    map[string]any{"param1": "string"},
		Extensions:         &map[string]string{"key1": "value1"},
	}

	// dummyManagedInfrastructureTemplateDefaults is test defaults for use in tests.
	dummyManagedInfrastructureTemplateDefaults = artifacts.ManagedInfrastructureTemplateDefaults{
		ClusterInstanceDefaults: &map[string]any{"cluster": "default"},
		PolicyTemplateDefaults:  &map[string]any{"policy": "default"},
	}

	// dummyProblemDetails is a test problem details object for use in tests where the API returns a 500 error.
	dummyProblemDetails = common.ProblemDetails{
		Status: 500,
		Title:  ptr.To("Internal Server Error"),
		Detail: "Internal server error occurred",
	}

	// testTemplateID is a common template ID for use in tests.
	testTemplateID = "test-template-id"
)

func TestArtifactsListManagedInfrastructureTemplates(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		filter        []filter.Filter
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:   "success without filter",
			filter: nil,
			handler: successHandler([]artifacts.ManagedInfrastructureTemplate{dummyManagedInfrastructureTemplate},
				http.StatusOK),
		},
		{
			name:   "success with filter",
			filter: []filter.Filter{filter.Equals("name", "test-template")},
			handler: filterSuccessHandler([]artifacts.ManagedInfrastructureTemplate{dummyManagedInfrastructureTemplate},
				"(eq,name,test-template)"),
		},
		{
			name:   "success with multiple filters - only first used",
			filter: []filter.Filter{filter.Equals("name", "test1"), filter.Equals("version", "v1.0")},
			handler: filterSuccessHandler([]artifacts.ManagedInfrastructureTemplate{dummyManagedInfrastructureTemplate},
				"(eq,name,test1)"),
		},
		{
			name:          "server error 500",
			filter:        nil,
			handler:       problemDetailsHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError: "failed to list ManagedInfrastructureTemplates: received error from api:",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(testCase.handler)
			defer server.Close()

			client, err := artifacts.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			artifactsClient := &ArtifactsClient{ClientWithResponsesInterface: client}
			result, err := artifactsClient.ListManagedInfrastructureTemplates(testCase.filter...)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), testCase.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestArtifactsGetManagedInfrastructureTemplate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		templateID    string
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:       "success",
			templateID: testTemplateID,
			handler:    successHandler(dummyManagedInfrastructureTemplate, http.StatusOK),
		},
		{
			name:          "server error 500",
			templateID:    testTemplateID,
			handler:       problemDetailsHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError: "failed to get ManagedInfrastructureTemplate: received error from api:",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(testCase.handler)
			defer server.Close()

			client, err := artifacts.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			artifactsClient := &ArtifactsClient{ClientWithResponsesInterface: client}
			result, err := artifactsClient.GetManagedInfrastructureTemplate(testCase.templateID)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestArtifactsGetManagedInfrastructureTemplateDefaults(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		templateID    string
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name:       "success",
			templateID: testTemplateID,
			handler:    successHandler(dummyManagedInfrastructureTemplateDefaults, http.StatusOK),
		},
		{
			name:       "success with nil defaults",
			templateID: testTemplateID,
			handler: successHandler(artifacts.ManagedInfrastructureTemplateDefaults{
				ClusterInstanceDefaults: nil,
				PolicyTemplateDefaults:  nil,
			}, http.StatusOK),
		},
		{
			name:          "server error 500",
			templateID:    testTemplateID,
			handler:       problemDetailsHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError: "failed to get ManagedInfrastructureTemplateDefaults: received error from api:",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(testCase.handler)
			defer server.Close()

			client, err := artifacts.NewClientWithResponses(server.URL)
			assert.NoError(t, err)

			artifactsClient := &ArtifactsClient{ClientWithResponsesInterface: client}
			result, err := artifactsClient.GetManagedInfrastructureTemplateDefaults(testCase.templateID)

			if testCase.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestArtifactsNetworkError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		testFunc func(client *ArtifactsClient) error
	}{
		{
			name: "ListManagedInfrastructureTemplates network error",
			testFunc: func(client *ArtifactsClient) error {
				_, err := client.ListManagedInfrastructureTemplates()

				return err
			},
		},
		{
			name: "GetManagedInfrastructureTemplate network error",
			testFunc: func(client *ArtifactsClient) error {
				_, err := client.GetManagedInfrastructureTemplate("test-id")

				return err
			},
		},
		{
			name: "GetManagedInfrastructureTemplateDefaults network error",
			testFunc: func(client *ArtifactsClient) error {
				_, err := client.GetManagedInfrastructureTemplateDefaults("test-id")

				return err
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// 192.0.2.0 is a reserved test address so we never accidentally use a valid IP. Still, we set a
			// timeout to ensure that we do not timeout the test.
			client, err := artifacts.NewClientWithResponses("http://192.0.2.0:8080",
				artifacts.WithHTTPClient(&http.Client{Timeout: time.Second * 1}))
			assert.NoError(t, err)

			artifactsClient := &ArtifactsClient{ClientWithResponsesInterface: client}
			err = testCase.testFunc(artifactsClient)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "error contacting api")
		})
	}
}

// successHandler returns an http.HandlerFunc that serves a successful response with the given data and status code.
//
//nolint:unparam // The status code will be used by other APIs.
func successHandler(data any, statusCode int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(data)
	}
}

// problemDetailsHandler returns an http.HandlerFunc that serves a ProblemDetails error.
func problemDetailsHandler(problemDetails common.ProblemDetails, statusCode int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(problemDetails)
	}
}

// filterSuccessHandler returns an http.HandlerFunc that serves a successful response with the given data, but only if
// the filter query parameter matches the expected filter. If the filter doesn't match, it returns a 400 error.
func filterSuccessHandler(data any, expectedFilter string) http.HandlerFunc {
	return func(writer http.ResponseWriter, r *http.Request) {
		actualFilter := r.URL.Query().Get("filter")
		if actualFilter != expectedFilter {
			// Since the test case will be expecting a successful response, by returning a 400 error we are
			// causing the test to fail. This is somewhat roundabout but avoids using the testing T inside
			// the handler.
			writer.Header().Set("Content-Type", "application/problem+json")
			writer.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(writer).Encode(common.ProblemDetails{
				Status: 400,
				Title:  ptr.To("Bad Request"),
				Detail: "Filter parameter does not match expected value",
			})

			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(writer).Encode(data)
	}
}
