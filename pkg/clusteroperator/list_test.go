package clusteroperator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifyClusterOperatorsVersion(t *testing.T) {
	testCases := []struct {
		desiredVersion      string
		clusterOperatorList []*Builder
		expectedOutput      bool
		expectedError       error
	}{
		{
			desiredVersion:      "",
			expectedOutput:      false,
			clusterOperatorList: buildFakeClusterOperatorListWithDesiredVersion(buildClusterOperatorClientWithDummyObject()),
			expectedError:       fmt.Errorf("desiredVersion can't be empty"),
		},
		{
			desiredVersion:      "4.14.0",
			expectedOutput:      false,
			clusterOperatorList: buildFakeClusterOperatorListWithoutDesiredVersion(buildClusterOperatorClientWithDummyObject()),
			expectedError: fmt.Errorf("the clusterOperator %s doesn't have the desired version 4.14.0",
				defaultClusterOperatorName),
		},
		{
			desiredVersion:      "4.14.0",
			expectedOutput:      true,
			clusterOperatorList: buildFakeClusterOperatorListWithDesiredVersion(buildClusterOperatorClientWithDummyObject()),
			expectedError:       nil,
		},
		{
			desiredVersion:      "4.14.0",
			expectedOutput:      false,
			clusterOperatorList: nil,
			expectedError:       fmt.Errorf("clusterOperatorList is invalid"),
		},
		{
			desiredVersion:      "4.14.0",
			expectedOutput:      false,
			clusterOperatorList: []*Builder{},
			expectedError:       fmt.Errorf("clusterOperatorList can't be empty"),
		},
	}

	for _, testCase := range testCases {
		result, err := VerifyClusterOperatorsVersion(testCase.desiredVersion, testCase.clusterOperatorList)
		assert.Equal(t, testCase.expectedOutput, result)
		assert.Equal(t, testCase.expectedError, err)
	}
}
