package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
)

type mockBuilder[T any] struct {
	definitionFunc   func() *T
	errorMsgFunc     func() string
	apiClientFunc    func() interface{}
	resourceTypeFunc func() string
}

func (b *mockBuilder[T]) GetDefinition() *T {
	if b.definitionFunc != nil {
		return b.definitionFunc()
	}

	return nil
}

func (b *mockBuilder[T]) GetErrorMsg() string {
	if b.errorMsgFunc != nil {
		return b.errorMsgFunc()
	}

	return ""
}

func (b *mockBuilder[T]) GetAPIClient() interface{} {
	if b.apiClientFunc != nil {
		return b.apiClientFunc()
	}

	return nil
}

func (b *mockBuilder[T]) GetResourceType() string {
	if b.resourceTypeFunc != nil {
		return b.resourceTypeFunc()
	}

	return "default"
}

func TestValidateBuilder(t *testing.T) {
	testCases := []struct {
		definitionReturnValue   *appsv1.Deployment
		errorMsgReturnValue     string
		apiClientReturnValue    interface{}
		resourceTypeReturnValue string
		expectedValid           bool
		expectedError           string
	}{
		{
			definitionReturnValue:   nil,
			errorMsgReturnValue:     "",
			apiClientReturnValue:    nil,
			resourceTypeReturnValue: "testResource",
			expectedValid:           false,
			expectedError:           "can not redefine the undefined testResource",
		},
		{
			definitionReturnValue:   &appsv1.Deployment{}, // non-nil definition
			errorMsgReturnValue:     "",
			apiClientReturnValue:    nil,
			resourceTypeReturnValue: "testResource",
			expectedValid:           false,
			expectedError:           "testResource builder cannot have nil apiClient",
		},
		{
			definitionReturnValue:   &appsv1.Deployment{}, // non-nil definition
			errorMsgReturnValue:     "some error message",
			apiClientReturnValue:    struct{}{}, // non-nil apiClient
			resourceTypeReturnValue: "testResource",
			expectedValid:           false,
			expectedError:           "some error message",
		},
		{ // Happy path case
			definitionReturnValue:   &appsv1.Deployment{}, // non-nil definition
			errorMsgReturnValue:     "",
			apiClientReturnValue:    struct{}{}, // non-nil apiClient
			resourceTypeReturnValue: "testResource",
			expectedValid:           true,
			expectedError:           "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := &mockBuilder[appsv1.Deployment]{
			definitionFunc:   func() *appsv1.Deployment { return testCase.definitionReturnValue },
			errorMsgFunc:     func() string { return testCase.errorMsgReturnValue },
			apiClientFunc:    func() interface{} { return testCase.apiClientReturnValue },
			resourceTypeFunc: func() string { return testCase.resourceTypeReturnValue },
		}

		valid, err := ValidateBuilder(testBuilder)
		if testCase.expectedValid {
			assert.True(t, valid)
			assert.Nil(t, err)
		} else {
			assert.False(t, valid)
			assert.NotNil(t, err)
			assert.Equal(t, testCase.expectedError, err.Error())
		}
	}

	// One additional test case for nil builder
	valid, err := ValidateBuilder[appsv1.Deployment](nil)
	assert.False(t, valid)
	assert.NotNil(t, err)
	assert.Equal(t, "error: received nil builder", err.Error())
}
