package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/oran/api/filter"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/oran/api/internal/artifacts"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/oran/api/internal/common"
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
		ParameterSchema:    map[string]interface{}{"param1": "string"},
		Extensions:         &map[string]string{"key1": "value1"},
	}

	// dummyManagedInfrastructureTemplateDefaults is test defaults for use in tests.
	dummyManagedInfrastructureTemplateDefaults = artifacts.ManagedInfrastructureTemplateDefaults{
		ClusterInstanceDefaults: &map[string]interface{}{"cluster": "default"},
		PolicyTemplateDefaults:  &map[string]interface{}{"policy": "default"},
	}

	// dummyProblemDetails is a test problem details object for use in tests where the API returns a 500 error.
	dummyProblemDetails = common.ProblemDetails{
		Status: 500,
		Title:  ptr.To("Internal Server Error"),
		Detail: "Internal server error occurred",
	}
)

func TestListManagedInfrastructureTemplates(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		filter          []filter.Filter
		handler         http.HandlerFunc
		expectedError   string
		validateRequest func(t *testing.T, req *http.Request)
	}{
		{
			name:            "success without filter",
			filter:          nil,
			handler:         listTemplatesHandler([]artifacts.ManagedInfrastructureTemplate{dummyManagedInfrastructureTemplate}),
			validateRequest: validateListTemplatesRequest(""),
		},
		{
			name:            "success with filter",
			filter:          []filter.Filter{filter.Equals("name", "test-template")},
			handler:         listTemplatesHandler([]artifacts.ManagedInfrastructureTemplate{dummyManagedInfrastructureTemplate}),
			validateRequest: validateListTemplatesRequest("(eq,name,test-template)"),
		},
		{
			name:            "success with multiple filters - only first used",
			filter:          []filter.Filter{filter.Equals("name", "test1"), filter.Equals("version", "v1.0")},
			handler:         listTemplatesHandler([]artifacts.ManagedInfrastructureTemplate{dummyManagedInfrastructureTemplate}),
			validateRequest: validateListTemplatesRequest("(eq,name,test1)"),
		},
		{
			name:            "server error 500",
			filter:          nil,
			handler:         listTemplatesProblemDetailsHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError:   "failed to list ManagedInfrastructureTemplates: received error from api:",
			validateRequest: validateListTemplatesRequest(""),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var capturedRequest *http.Request

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedRequest = r
				testCase.handler(w, r)
			}))
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
				assert.Equal(t, dummyManagedInfrastructureTemplate.Name, result[0].Name)
			}

			testCase.validateRequest(t, capturedRequest)
		})
	}
}

func TestGetManagedInfrastructureTemplate(t *testing.T) {
	t.Parallel()

	testTemplateID := "test-template-id"

	testCases := []struct {
		name            string
		templateID      string
		handler         http.HandlerFunc
		expectedError   string
		validateRequest func(t *testing.T, req *http.Request)
		validateResult  func(t *testing.T, result *artifacts.ManagedInfrastructureTemplate)
	}{
		{
			name:            "success",
			templateID:      testTemplateID,
			handler:         getTemplateHandler(dummyManagedInfrastructureTemplate, http.StatusOK),
			validateRequest: validateGetTemplateRequest(testTemplateID),
			validateResult: func(t *testing.T, result *artifacts.ManagedInfrastructureTemplate) {
				t.Helper()
				validateGetTemplateResult(t, result, dummyManagedInfrastructureTemplate)
			},
		},
		{
			name:            "server error 500",
			templateID:      testTemplateID,
			handler:         getTemplateProblemDetailsHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError:   "failed to get ManagedInfrastructureTemplate: received error from api:",
			validateRequest: validateGetTemplateRequest(testTemplateID),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var capturedRequest *http.Request

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedRequest = r
				testCase.handler(w, r)
			}))
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
				testCase.validateResult(t, result)
			}

			testCase.validateRequest(t, capturedRequest)
		})
	}
}

func TestGetManagedInfrastructureTemplateDefaults(t *testing.T) {
	t.Parallel()

	testTemplateID := "test-template-id"

	testCases := []struct {
		name            string
		templateID      string
		handler         http.HandlerFunc
		expectedError   string
		validateRequest func(t *testing.T, req *http.Request)
		validateResult  func(t *testing.T, result *artifacts.ManagedInfrastructureTemplateDefaults)
	}{
		{
			name:            "success",
			templateID:      testTemplateID,
			handler:         getTemplateDefaultsHandler(dummyManagedInfrastructureTemplateDefaults),
			validateRequest: validateGetTemplateDefaultsRequest(testTemplateID),
			validateResult:  validateGetTemplateDefaultsResult(dummyManagedInfrastructureTemplateDefaults),
		},
		{
			name:       "success with nil defaults",
			templateID: testTemplateID,
			handler: getTemplateDefaultsHandler(artifacts.ManagedInfrastructureTemplateDefaults{
				ClusterInstanceDefaults: nil,
				PolicyTemplateDefaults:  nil,
			}),
			validateRequest: validateGetTemplateDefaultsRequest(testTemplateID),
			validateResult:  validateGetTemplateDefaultsResultNil(),
		},
		{
			name:            "server error 500",
			templateID:      testTemplateID,
			handler:         getTemplateDefaultsProblemDetailsHandler(dummyProblemDetails, http.StatusInternalServerError),
			expectedError:   "failed to get ManagedInfrastructureTemplateDefaults: received error from api:",
			validateRequest: validateGetTemplateDefaultsRequest(testTemplateID),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var capturedRequest *http.Request

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedRequest = r
				testCase.handler(w, r)
			}))
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
				testCase.validateResult(t, result)
			}

			testCase.validateRequest(t, capturedRequest)
		})
	}
}

func TestNetworkError(t *testing.T) {
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

// These next few functions are handlers that effectively mock the API server. However, they mock it at the HTTP level
// rather than being a mock generated client. Since artifacts focuses more on thin abstractions over the client, mocking
// the HTTP server gives better coverage.

// listTemplatesHandler returns an http.HandlerFunc that serves a list of ManagedInfrastructureTemplate objects.
func listTemplatesHandler(templates []artifacts.ManagedInfrastructureTemplate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(templates)
	}
}

// listTemplatesProblemDetailsHandler returns an http.HandlerFunc that serves a ProblemDetails error.
func listTemplatesProblemDetailsHandler(problemDetails common.ProblemDetails, statusCode int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(problemDetails)
	}
}

// getTemplateHandler returns an http.HandlerFunc that serves a single ManagedInfrastructureTemplate object.
func getTemplateHandler(template artifacts.ManagedInfrastructureTemplate, statusCode int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(template)
	}
}

// getTemplateProblemDetailsHandler returns an http.HandlerFunc that serves a ProblemDetails error.
func getTemplateProblemDetailsHandler(problemDetails common.ProblemDetails, statusCode int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(problemDetails)
	}
}

// getTemplateDefaultsHandler returns an http.HandlerFunc that serves a ManagedInfrastructureTemplateDefaults object.
func getTemplateDefaultsHandler(defaults artifacts.ManagedInfrastructureTemplateDefaults) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(defaults)
	}
}

// getTemplateDefaultsProblemDetailsHandler returns an http.HandlerFunc that serves a ProblemDetails error.
func getTemplateDefaultsProblemDetailsHandler(problemDetails common.ProblemDetails, statusCode int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(problemDetails)
	}
}

// validateListTemplatesRequest validates the request for ListManagedInfrastructureTemplates.
func validateListTemplatesRequest(expectedFilter string) func(t *testing.T, req *http.Request) {
	return func(t *testing.T, req *http.Request) {
		t.Helper()

		const expectedURL = "/o2ims-infrastructureArtifacts/v1/managedInfrastructureTemplates"

		assert.Equal(t, "GET", req.Method)
		assert.Equal(t, expectedURL, req.URL.Path)
		assert.Equal(t, expectedFilter, req.URL.Query().Get("filter"))
	}
}

// validateGetTemplateRequest validates the request for GetManagedInfrastructureTemplate.
func validateGetTemplateRequest(templateID string) func(t *testing.T, req *http.Request) {
	return func(t *testing.T, req *http.Request) {
		t.Helper()

		expectedURL := fmt.Sprintf("/o2ims-infrastructureArtifacts/v1/managedInfrastructureTemplates/%s", templateID)

		assert.Equal(t, "GET", req.Method)
		assert.Equal(t, expectedURL, req.URL.Path)
	}
}

// validateGetTemplateResult validates the result for GetManagedInfrastructureTemplate.
func validateGetTemplateResult(t *testing.T, result *artifacts.ManagedInfrastructureTemplate,
	expectedTemplate artifacts.ManagedInfrastructureTemplate) {
	t.Helper()
	assert.NotNil(t, result)
	assert.Equal(t, expectedTemplate.Name, result.Name)
	assert.Equal(t, expectedTemplate.Description, result.Description)
}

// validateGetTemplateDefaultsRequest validates the request for GetManagedInfrastructureTemplateDefaults.
func validateGetTemplateDefaultsRequest(templateID string) func(t *testing.T, req *http.Request) {
	return func(t *testing.T, req *http.Request) {
		t.Helper()

		expectedURL := fmt.Sprintf("/o2ims-infrastructureArtifacts/v1/managedInfrastructureTemplates/%s/defaults", templateID)

		assert.Equal(t, "GET", req.Method)
		assert.Equal(t, expectedURL, req.URL.Path)
	}
}

// templateDefaultsResultValidator is a function that validates the result for GetManagedInfrastructureTemplateDefaults.
type templateDefaultsResultValidator = func(t *testing.T, result *artifacts.ManagedInfrastructureTemplateDefaults)

// validateGetTemplateDefaultsResult validates the result for GetManagedInfrastructureTemplateDefaults.
func validateGetTemplateDefaultsResult(
	expectedDefaults artifacts.ManagedInfrastructureTemplateDefaults) templateDefaultsResultValidator {
	return func(t *testing.T, result *artifacts.ManagedInfrastructureTemplateDefaults) {
		t.Helper()
		assert.NotNil(t, result)
		assert.NotNil(t, result.ClusterInstanceDefaults)
		assert.NotNil(t, result.PolicyTemplateDefaults)
		assert.Equal(t, (*expectedDefaults.ClusterInstanceDefaults)["cluster"], (*result.ClusterInstanceDefaults)["cluster"])
		assert.Equal(t, (*expectedDefaults.PolicyTemplateDefaults)["policy"], (*result.PolicyTemplateDefaults)["policy"])
	}
}

// validateGetTemplateDefaultsResultNil validates that the result for GetManagedInfrastructureTemplateDefaults has nil
// defaults.
func validateGetTemplateDefaultsResultNil() templateDefaultsResultValidator {
	return func(t *testing.T, result *artifacts.ManagedInfrastructureTemplateDefaults) {
		t.Helper()
		assert.NotNil(t, result)
		assert.Nil(t, result.ClusterInstanceDefaults)
		assert.Nil(t, result.PolicyTemplateDefaults)
	}
}
