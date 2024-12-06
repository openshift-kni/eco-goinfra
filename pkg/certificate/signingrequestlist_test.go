package certificate

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	certificatesv1 "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListSigningRequests(t *testing.T) {
	testCases := []struct {
		signingRequests []*SigningRequestBuilder
		listOptions     []runtimeclient.ListOptions
		client          bool
		expectedError   error
	}{
		{
			signingRequests: []*SigningRequestBuilder{newSigningRequestBuilder(buildTestClientWithDummySigningRequest())},
			listOptions:     nil,
			client:          true,
			expectedError:   nil,
		},
		{
			signingRequests: []*SigningRequestBuilder{newSigningRequestBuilder(buildTestClientWithDummySigningRequest())},
			listOptions:     []runtimeclient.ListOptions{{LabelSelector: labels.NewSelector()}},
			client:          true,
			expectedError:   nil,
		},
		{
			signingRequests: []*SigningRequestBuilder{newSigningRequestBuilder(buildTestClientWithDummySigningRequest())},
			listOptions:     []runtimeclient.ListOptions{{}, {}},
			client:          true,
			expectedError:   fmt.Errorf("error: more than one ListOptions was passed"),
		},
		{
			signingRequests: []*SigningRequestBuilder{newSigningRequestBuilder(buildTestClientWithDummySigningRequest())},
			listOptions:     nil,
			client:          false,
			expectedError:   fmt.Errorf("certificateSigningRequest 'apiClient' cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = buildTestClientWithDummySigningRequest()
		}

		builders, err := ListSigningRequests(testSettings, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(testCase.signingRequests), len(builders))
		}
	}
}

func TestWaitUntilSigningRequestsApproved(t *testing.T) {
	testcases := []struct {
		listOptions   []runtimeclient.ListOptions
		client        bool
		approved      bool
		expectedError error
	}{
		{
			listOptions:   nil,
			client:        true,
			approved:      true,
			expectedError: nil,
		},
		{
			listOptions:   []runtimeclient.ListOptions{{}, {}},
			client:        true,
			approved:      true,
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
		},
		{
			listOptions:   nil,
			client:        false,
			approved:      true,
			expectedError: fmt.Errorf("certificateSigningRequest 'apiClient' cannot be nil"),
		},
		{
			listOptions:   nil,
			client:        true,
			approved:      false,
			expectedError: context.DeadlineExceeded,
		},
	}

	for _, testCase := range testcases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		if testCase.approved {
			runtimeObjects = append(runtimeObjects, buildDummyApprovedSigningRequest())
		} else {
			runtimeObjects = append(runtimeObjects, buildDummySigningRequest(defaultSigningRequestName))
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		err := WaitUntilSigningRequestsApproved(testSettings, time.Second, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)
	}
}

func buildDummyApprovedSigningRequest() *certificatesv1.CertificateSigningRequest {
	return &certificatesv1.CertificateSigningRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: defaultSigningRequestName,
		},
		Status: certificatesv1.CertificateSigningRequestStatus{
			Conditions: []certificatesv1.CertificateSigningRequestCondition{{
				Type:   certificatesv1.CertificateApproved,
				Status: corev1.ConditionTrue,
			}},
		},
	}
}
