package api

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/openshift-kni/eco-goinfra/pkg/oran/api/internal/common"
	"github.com/openshift-kni/eco-goinfra/pkg/oran/api/internal/provisioning"
	provisioningv1alpha1 "github.com/openshift-kni/oran-o2ims/api/provisioning/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	// dummyProvisioningRequestData is a dummy provisioning request data and the basis for the
	// dummyProvisioningRequest.
	dummyProvisioningRequestData = provisioning.ProvisioningRequestData{
		ProvisioningRequestId: uuid.New(),
		Description:           "mock description",
		Name:                  "mock name",
		TemplateName:          "mock template",
		TemplateVersion:       "v4.19.0",
		TemplateParameters:    make(map[string]any),
	}

	// defaultPRID is the string representation of the dummyProvisioningRequestData's ProvisioningRequestId. It is
	// used as the default name for the provisioning request.
	defaultPRID = dummyProvisioningRequestData.ProvisioningRequestId.String()

	// dummyProvisioningRequest uses the dummyProvisioningRequestData to create a dummy provisioning request,
	// complete with TypeMeta like from provisioningRequestFromInfo, but without the status.
	dummyProvisioningRequest = provisioningv1alpha1.ProvisioningRequest{
		TypeMeta: provisioningRequestTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name: defaultPRID,
		},
		Spec: provisioningv1alpha1.ProvisioningRequestSpec{
			Name:               dummyProvisioningRequestData.Name,
			Description:        dummyProvisioningRequestData.Description,
			TemplateName:       dummyProvisioningRequestData.TemplateName,
			TemplateVersion:    dummyProvisioningRequestData.TemplateVersion,
			TemplateParameters: runtime.RawExtension{Raw: []byte("{}")},
		},
	}
)

func TestProvisioningClientGet(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		key           runtimeclient.ObjectKey
		object        runtimeclient.Object
		validateError func(t *testing.T, err error)
	}{
		{
			name:          "success",
			key:           runtimeclient.ObjectKey{Name: defaultPRID},
			object:        &provisioningv1alpha1.ProvisioningRequest{},
			validateError: validateNoError,
		},
		{
			name:          "wrong object type",
			key:           runtimeclient.ObjectKey{Name: defaultPRID},
			object:        &provisioningv1alpha1.ClusterTemplate{},
			validateError: validatePRWrongObjectType,
		},
		{
			name:          "invalid uuid",
			key:           runtimeclient.ObjectKey{Name: "not-a-uuid"},
			object:        &provisioningv1alpha1.ProvisioningRequest{},
			validateError: validateInvalidUUID,
		},
		{
			name:          "nonexistent object",
			key:           runtimeclient.ObjectKey{Name: uuid.New().String()},
			object:        &provisioningv1alpha1.ProvisioningRequest{},
			validateError: validateNotFoundError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockClient := buildMockProvisioningClientWithDummy()
			client := ProvisioningClient{mockClient}

			err := client.Get(t.Context(), testCase.key, testCase.object)
			testCase.validateError(t, err)
		})
	}
}

func TestProvisioningClientList(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		object        runtimeclient.ObjectList
		opts          []runtimeclient.ListOption
		validateError func(t *testing.T, err error)
	}{
		{
			name:          "success",
			object:        &provisioningv1alpha1.ProvisioningRequestList{},
			validateError: validateNoError,
		},
		{
			name:          "wrong object type",
			object:        &provisioningv1alpha1.ClusterTemplateList{},
			validateError: validatePRWrongObjectListType,
		},
		{
			name:   "wrong list options type",
			object: &provisioningv1alpha1.ProvisioningRequestList{},
			opts:   []runtimeclient.ListOption{&runtimeclient.ListOptions{}},
			validateError: func(t *testing.T, err error) {
				t.Helper()
				assert.Error(t, err)
				assert.Contains(t, err.Error(),
					"options must be pointer to ProvisioningRequestListOptions, not *client.ListOptions")
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			mockClient := buildMockProvisioningClientWithDummy()
			client := ProvisioningClient{mockClient}

			err := client.List(t.Context(), testCase.object, testCase.opts...)
			testCase.validateError(t, err)
		})
	}
}

func TestProvisioningClientCreate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		client        provisioning.ClientWithResponsesInterface
		object        runtimeclient.Object
		validateError func(t *testing.T, err error)
	}{
		{
			name:          "success",
			client:        buildMockProvisioningClient(),
			object:        &dummyProvisioningRequest,
			validateError: validateNoError,
		},
		{
			name:          "wrong object type",
			client:        buildMockProvisioningClient(),
			object:        &provisioningv1alpha1.ClusterTemplate{},
			validateError: validatePRWrongObjectType,
		},
		{
			name:   "invalid uuid",
			client: buildMockProvisioningClient(),
			object: &provisioningv1alpha1.ProvisioningRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name: "not-a-uuid",
				},
			},
			validateError: validateInvalidUUID,
		},
		{
			name:          "already exists",
			client:        buildMockProvisioningClientWithDummy(),
			object:        &dummyProvisioningRequest,
			validateError: validateConflictError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			client := ProvisioningClient{testCase.client}

			err := client.Create(t.Context(), testCase.object)
			testCase.validateError(t, err)
		})
	}
}

func TestProvisioningClientDelete(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		client        provisioning.ClientWithResponsesInterface
		object        runtimeclient.Object
		validateError func(t *testing.T, err error)
	}{
		{
			name:          "success",
			client:        buildMockProvisioningClientWithDummy(),
			object:        &dummyProvisioningRequest,
			validateError: validateNoError,
		},
		{
			name:          "wrong object type",
			client:        buildMockProvisioningClientWithDummy(),
			object:        &provisioningv1alpha1.ClusterTemplate{},
			validateError: validatePRWrongObjectType,
		},
		{
			name:   "invalid uuid",
			client: buildMockProvisioningClientWithDummy(),
			object: &provisioningv1alpha1.ProvisioningRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name: "not-a-uuid",
				},
			},
			validateError: validateInvalidUUID,
		},
		{
			name:          "not found",
			client:        buildMockProvisioningClient(),
			object:        &dummyProvisioningRequest,
			validateError: validateNotFoundError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			client := ProvisioningClient{testCase.client}

			err := client.Delete(t.Context(), testCase.object)
			testCase.validateError(t, err)
		})
	}
}

func TestProvisioningClientUpdate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		client        provisioning.ClientWithResponsesInterface
		object        runtimeclient.Object
		validateError func(t *testing.T, err error)
	}{
		{
			name:          "success",
			client:        buildMockProvisioningClientWithDummy(),
			object:        &dummyProvisioningRequest,
			validateError: validateNoError,
		},
		{
			name:          "wrong object type",
			client:        buildMockProvisioningClientWithDummy(),
			object:        &provisioningv1alpha1.ClusterTemplate{},
			validateError: validatePRWrongObjectType,
		},
		{
			name:   "invalid uuid",
			client: buildMockProvisioningClientWithDummy(),
			object: &provisioningv1alpha1.ProvisioningRequest{
				ObjectMeta: metav1.ObjectMeta{
					Name: "not-a-uuid",
				},
			},
			validateError: validateInvalidUUID,
		},
		{
			name:          "not found",
			client:        buildMockProvisioningClient(),
			object:        &dummyProvisioningRequest,
			validateError: validateNotFoundError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			client := ProvisioningClient{testCase.client}

			err := client.Update(t.Context(), testCase.object)
			testCase.validateError(t, err)
		})
	}
}

func TestProvisioningClientPatch(t *testing.T) {
	t.Parallel()

	client := ProvisioningClient{}
	err := client.Patch(t.Context(), nil, nil)
	assert.Error(t, err)

	target := &unimplementedError{}
	assert.ErrorAs(t, err, &target)
	assert.Equal(t, ProvisioningClientType, target.clientType)
	assert.Equal(t, "Patch", target.method)
}

func TestProvisioningClientDeleteAllOf(t *testing.T) {
	t.Parallel()

	client := ProvisioningClient{}
	err := client.DeleteAllOf(t.Context(), nil)
	assert.Error(t, err)

	target := &unimplementedError{}
	assert.ErrorAs(t, err, &target)
	assert.Equal(t, ProvisioningClientType, target.clientType)
	assert.Equal(t, "DeleteAllOf", target.method)
}

func TestProvisioningClientStatus(t *testing.T) {
	t.Parallel()

	client := ProvisioningClient{}
	status := client.Status()
	assert.Nil(t, status)
}

func TestProvisioningClientSubResource(t *testing.T) {
	t.Parallel()

	client := ProvisioningClient{}
	subResource := client.SubResource("")
	assert.Nil(t, subResource)
}

func TestProvisioningClientScheme(t *testing.T) {
	t.Parallel()

	client := ProvisioningClient{}
	scheme := client.Scheme()
	assert.NotNil(t, scheme)
}

func TestProvisioningClientRESTMapper(t *testing.T) {
	t.Parallel()

	client := ProvisioningClient{}
	restMapper := client.RESTMapper()
	assert.Nil(t, restMapper)
}

func TestProvisioningClientGroupVersionKindFor(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		obj           runtime.Object
		validateError func(t *testing.T, err error)
	}{
		{
			name:          "ProvisioningRequest",
			obj:           &provisioningv1alpha1.ProvisioningRequest{},
			validateError: validateNoError,
		},
		{
			name: "nil",
			obj:  nil,
			validateError: func(t *testing.T, err error) {
				t.Helper()
				assert.Error(t, err)
				target := &unimplementedError{}
				assert.ErrorAs(t, err, &target)
				assert.Equal(t, ProvisioningClientType, target.clientType)
				assert.Equal(t, "GroupVersionKindFor", target.method)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			client := ProvisioningClient{}
			_, err := client.GroupVersionKindFor(testCase.obj)
			testCase.validateError(t, err)
		})
	}
}

func TestProvisioningClientIsObjectNamespaced(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		obj           runtime.Object
		validateError func(t *testing.T, err error)
	}{
		{
			name:          "ProvisioningRequest",
			obj:           &provisioningv1alpha1.ProvisioningRequest{},
			validateError: validateNoError,
		},
		{
			name: "nil",
			obj:  nil,
			validateError: func(t *testing.T, err error) {
				t.Helper()
				assert.Error(t, err)
				target := &unimplementedError{}
				assert.ErrorAs(t, err, &target)
				assert.Equal(t, ProvisioningClientType, target.clientType)
				assert.Equal(t, "IsObjectNamespaced", target.method)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			client := ProvisioningClient{}
			_, err := client.IsObjectNamespaced(testCase.obj)
			testCase.validateError(t, err)
		})
	}
}

func buildMockProvisioningClient() *mockProvisioningClientWithResponses {
	return &mockProvisioningClientWithResponses{
		provisioningRequests: make(map[uuid.UUID]provisioning.ProvisioningRequestInfo),
	}
}

func buildMockProvisioningClientWithDummy() *mockProvisioningClientWithResponses {
	return &mockProvisioningClientWithResponses{
		provisioningRequests: map[uuid.UUID]provisioning.ProvisioningRequestInfo{
			dummyProvisioningRequestData.ProvisioningRequestId: {
				ProvisioningRequestData: dummyProvisioningRequestData,
			},
		},
	}
}

func validateNoError(t *testing.T, err error) {
	t.Helper()
	assert.NoError(t, err)
}

func validateInvalidUUID(t *testing.T, err error) {
	t.Helper()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid UUID length")
}

func validatePRWrongObjectType(t *testing.T, err error) {
	t.Helper()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "object must be pointer to ProvisioningRequest, not *v1alpha1.ClusterTemplate")
}

func validatePRWrongObjectListType(t *testing.T, err error) {
	t.Helper()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "object must be pointer to ProvisioningRequestList, not *v1alpha1.ClusterTemplateList")
}

func validateNotFoundError(t *testing.T, err error) {
	t.Helper()
	assert.Error(t, err)

	target := &Error{}
	assert.ErrorAs(t, err, &target)
	assert.Equal(t, 404, target.Status)
}

func validateConflictError(t *testing.T, err error) {
	t.Helper()
	assert.Error(t, err)

	target := &Error{}
	assert.ErrorAs(t, err, &target)
	assert.Equal(t, 409, target.Status)
}

var (
	// mockNotSupported is the response returned by the mock client when a method is not supported.
	mockNotSupported = common.ProblemDetails{
		Status: 500,
		Detail: "mock does not support this method",
	}

	// mockNotFound is the response returned by the mock client when a resource is not found.
	mockNotFound = common.ProblemDetails{
		Status: 404,
		Detail: "The specified provisioning request was not found.",
	}

	// mockConflict is the response returned by the mock client when a resource already exists.
	mockConflict = common.ProblemDetails{
		Status: 409,
		Detail: "Conflict",
	}
)

// mockProvisioningClientWithResponses is a mock implementation of the provisioning.ClientWithResponsesInterface.
type mockProvisioningClientWithResponses struct {
	provisioningRequests map[uuid.UUID]provisioning.ProvisioningRequestInfo
}

// Enforce at compile time that mockProvisioningClientWithResponses implements the
// provisioning.ClientWithResponsesInterface.
var _ provisioning.ClientWithResponsesInterface = (*mockProvisioningClientWithResponses)(nil)

// GetProvisioningRequestWithResponse returns a single provisioning request by ID, or a 404 error if not found.
func (mock *mockProvisioningClientWithResponses) GetProvisioningRequestWithResponse(
	_ context.Context, provisioningRequestID uuid.UUID, _ ...provisioning.RequestEditorFn,
) (*provisioning.GetProvisioningRequestResponse, error) {
	if request, ok := mock.provisioningRequests[provisioningRequestID]; ok {
		return &provisioning.GetProvisioningRequestResponse{
			HTTPResponse: &http.Response{
				StatusCode: 200,
			},
			JSON200: &request,
		}, nil
	}

	return &provisioning.GetProvisioningRequestResponse{
		HTTPResponse: &http.Response{
			StatusCode: 404,
		},
		ApplicationProblemJSON404: &mockNotFound,
	}, nil
}

// GetProvisioningRequestsWithResponse always returns all the provisioning requests. Parameters are ignored, so
// filtering is not supported.
func (mock *mockProvisioningClientWithResponses) GetProvisioningRequestsWithResponse(
	_ context.Context, _ *provisioning.GetProvisioningRequestsParams, _ ...provisioning.RequestEditorFn,
) (*provisioning.GetProvisioningRequestsResponse, error) {
	requests := make([]provisioning.ProvisioningRequestInfo, 0, len(mock.provisioningRequests))
	for _, request := range mock.provisioningRequests {
		requests = append(requests, request)
	}

	return &provisioning.GetProvisioningRequestsResponse{
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
		JSON200: &requests,
	}, nil
}

// CreateProvisioningRequestWithResponse creates a new provisioning request. If the request already exists, a 409
// conflict error is returned.
func (mock *mockProvisioningClientWithResponses) CreateProvisioningRequestWithResponse(
	_ context.Context, body provisioning.CreateProvisioningRequestJSONRequestBody, _ ...provisioning.RequestEditorFn,
) (*provisioning.CreateProvisioningRequestResponse, error) {
	if _, ok := mock.provisioningRequests[body.ProvisioningRequestId]; ok {
		return &provisioning.CreateProvisioningRequestResponse{
			HTTPResponse: &http.Response{
				StatusCode: 409,
			},
			ApplicationProblemJSON409: &mockConflict,
		}, nil
	}

	request := provisioning.ProvisioningRequestInfo{
		ProvisioningRequestData: body,
	}
	mock.provisioningRequests[body.ProvisioningRequestId] = request

	return &provisioning.CreateProvisioningRequestResponse{
		HTTPResponse: &http.Response{
			StatusCode: 201,
		},
		JSON201: &request,
	}, nil
}

// UpdateProvisioningRequestWithResponse updates an existing provisioning request. If the request does not exist, a 404
// not found error is returned.
func (mock *mockProvisioningClientWithResponses) UpdateProvisioningRequestWithResponse(
	_ context.Context,
	provisioningRequestID uuid.UUID,
	body provisioning.UpdateProvisioningRequestJSONRequestBody,
	_ ...provisioning.RequestEditorFn,
) (*provisioning.UpdateProvisioningRequestResponse, error) {
	if _, ok := mock.provisioningRequests[provisioningRequestID]; !ok {
		return &provisioning.UpdateProvisioningRequestResponse{
			HTTPResponse: &http.Response{
				StatusCode: 404,
			},
			ApplicationProblemJSON404: &mockNotFound,
		}, nil
	}

	request := provisioning.ProvisioningRequestInfo{
		ProvisioningRequestData: body,
	}
	mock.provisioningRequests[provisioningRequestID] = request

	return &provisioning.UpdateProvisioningRequestResponse{
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
		JSON200: &request,
	}, nil
}

// DeleteProvisioningRequestWithResponse deletes an existing provisioning request. If the request does not exist, a 404
// not found error is returned.
func (mock *mockProvisioningClientWithResponses) DeleteProvisioningRequestWithResponse(
	_ context.Context, provisioningRequestID uuid.UUID, _ ...provisioning.RequestEditorFn,
) (*provisioning.DeleteProvisioningRequestResponse, error) {
	if _, ok := mock.provisioningRequests[provisioningRequestID]; !ok {
		return &provisioning.DeleteProvisioningRequestResponse{
			HTTPResponse: &http.Response{
				StatusCode: 404,
			},
			ApplicationProblemJSON404: &mockNotFound,
		}, nil
	}

	delete(mock.provisioningRequests, provisioningRequestID)

	return &provisioning.DeleteProvisioningRequestResponse{
		HTTPResponse: &http.Response{
			StatusCode: 200,
		},
	}, nil
}

// GetAllVersionsWithResponse returns a 500 error to indicate that the method is not supported by the mock.
func (mock *mockProvisioningClientWithResponses) GetAllVersionsWithResponse(
	_ context.Context, _ ...provisioning.RequestEditorFn) (*provisioning.GetAllVersionsResponse, error) {
	return &provisioning.GetAllVersionsResponse{
		HTTPResponse: &http.Response{
			StatusCode: 500,
		},
		ApplicationProblemJSON500: &mockNotSupported,
	}, nil
}

// GetMinorVersionsWithResponse returns a 500 error to indicate that the method is not supported by the mock.
func (mock *mockProvisioningClientWithResponses) GetMinorVersionsWithResponse(
	_ context.Context, _ ...provisioning.RequestEditorFn) (*provisioning.GetMinorVersionsResponse, error) {
	return &provisioning.GetMinorVersionsResponse{
		HTTPResponse: &http.Response{
			StatusCode: 500,
		},
		ApplicationProblemJSON500: &mockNotSupported,
	}, nil
}

// CreateProvisioningRequestWithBodyWithResponse returns a 500 error to indicate that the method is not supported by the
// mock. The CreateProvisioningRequestWithResponse method should be used instead.
func (mock *mockProvisioningClientWithResponses) CreateProvisioningRequestWithBodyWithResponse(
	_ context.Context, _ string, _ io.Reader, _ ...provisioning.RequestEditorFn,
) (*provisioning.CreateProvisioningRequestResponse, error) {
	return &provisioning.CreateProvisioningRequestResponse{
		HTTPResponse: &http.Response{
			StatusCode: 500,
		},
		ApplicationProblemJSON500: &mockNotSupported,
	}, nil
}

// UpdateProvisioningRequestWithBodyWithResponse returns a 500 error to indicate that the method is not supported by
// the mock. The UpdateProvisioningRequestWithResponse method should be used instead.
func (mock *mockProvisioningClientWithResponses) UpdateProvisioningRequestWithBodyWithResponse(
	_ context.Context, _ uuid.UUID, _ string, _ io.Reader, _ ...provisioning.RequestEditorFn,
) (*provisioning.UpdateProvisioningRequestResponse, error) {
	return &provisioning.UpdateProvisioningRequestResponse{
		HTTPResponse: &http.Response{
			StatusCode: 500,
		},
		ApplicationProblemJSON500: &mockNotSupported,
	}, nil
}
