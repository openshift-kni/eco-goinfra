package pod

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"

	corev1 "k8s.io/api/core/v1"
)

func buildValidContainterBuilder() *ContainerBuilder {
	return NewContainerBuilder(
		"test-container",
		"registry.example.com/test/image:latest",
		[]string{"/bin/sh", "-c", "sleep infinity"})
}

func TestWithCustomResourcesRequests(t *testing.T) {
	testCases := []struct {
		testRequests   corev1.ResourceList
		expectedErrMsg string
		expectedValues map[string]string
	}{
		{
			testRequests:   corev1.ResourceList{},
			expectedErrMsg: "container's resource limit var 'resourceList' is empty",
		},
		{
			testRequests: corev1.ResourceList{
				corev1.ResourceName("openshift.io/sriov1"): resource.MustParse("1"),
			},
			expectedValues: map[string]string{
				"openshift.io/sriov1": "1",
			},
		},
		{
			testRequests: corev1.ResourceList{
				corev1.ResourceName("openshift.io/sriov1"): resource.MustParse("1"),
				corev1.ResourceName("openshift.io/fake2"):  resource.MustParse("2"),
			},
			expectedValues: map[string]string{
				"openshift.io/sriov1": "1",
				"openshift.io/fake2":  "2",
			},
		},
		{
			testRequests: corev1.ResourceList{
				corev1.ResourceName("openshift.io/sriov1"): resource.MustParse("1"),
				corev1.ResourceName("cpu"):                 resource.MustParse("2.5m"),
				corev1.ResourceName("memory"):              resource.MustParse("0.5Gi"),
			},
			expectedValues: map[string]string{
				"openshift.io/sriov1": "1",
				"cpu":                 "2.5m",
				"memory":              "0.5Gi",
			},
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidContainterBuilder()
		testBuilder = testBuilder.WithCustomResourcesRequests(testCase.testRequests)

		if testCase.expectedErrMsg != "" {
			assert.Equal(t, testCase.expectedErrMsg, testBuilder.errorMsg)
		} else {
			assert.Empty(t, testBuilder.errorMsg)
		}

		if len(testCase.testRequests) != 0 && len(testCase.expectedValues) != 0 {
			for k, v := range testCase.expectedValues {
				assert.Equal(t,
					resource.MustParse(v), testBuilder.definition.Resources.Requests[corev1.ResourceName(k)])
			}
		}
	}
}
