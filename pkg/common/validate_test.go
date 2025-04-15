package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockBuilder struct {
	definitionFunc   func() interface{}
	errorMsgFunc     func() string
	apiClientFunc    func() interface{}
	resourceTypeFunc func() string
}

func (b *mockBuilder) GetDefinition() interface{} {
	if b.definitionFunc != nil {
		return b.definitionFunc()
	}

	return nil
}

func (b *mockBuilder) GetErrorMsg() string {
	if b.errorMsgFunc != nil {
		return b.errorMsgFunc()
	}

	return ""
}

func (b *mockBuilder) GetAPIClient() interface{} {
	if b.apiClientFunc != nil {
		return b.apiClientFunc()
	}

	return nil
}

func (b *mockBuilder) GetResourceType() string {
	if b.resourceTypeFunc != nil {
		return b.resourceTypeFunc()
	}

	return "default"
}

func TestValidateBuilder(t *testing.T) {
	testCases := []struct {
		definitionReturnValue   interface{}
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
			definitionReturnValue:   struct{}{}, // non-nil definition
			errorMsgReturnValue:     "",
			apiClientReturnValue:    nil,
			resourceTypeReturnValue: "testResource",
			expectedValid:           false,
			expectedError:           "testResource builder cannot have nil apiClient",
		},
		{
			definitionReturnValue:   struct{}{}, // non-nil definition
			errorMsgReturnValue:     "some error message",
			apiClientReturnValue:    struct{}{}, // non-nil apiClient
			resourceTypeReturnValue: "testResource",
			expectedValid:           false,
			expectedError:           "some error message",
		},
		{ // Happy path case
			definitionReturnValue:   struct{}{}, // non-nil definition
			errorMsgReturnValue:     "",
			apiClientReturnValue:    struct{}{}, // non-nil apiClient
			resourceTypeReturnValue: "testResource",
			expectedValid:           true,
			expectedError:           "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := &mockBuilder{
			definitionFunc:   func() interface{} { return testCase.definitionReturnValue },
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
	valid, err := ValidateBuilder(nil)
	assert.False(t, valid)
	assert.NotNil(t, err)
	assert.Equal(t, "error: received nil builder", err.Error())
}
