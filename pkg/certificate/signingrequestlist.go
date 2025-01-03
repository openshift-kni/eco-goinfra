package certificate

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	certificatesv1 "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ListSigningRequests returns a list of all CertificateSigningRequest objects in the cluster with the provided options.
func ListSigningRequests(
	apiClient *clients.Settings, options ...runtimeclient.ListOptions) ([]*SigningRequestBuilder, error) {
	if apiClient == nil {
		glog.V(100).Infof("CertificateSigningRequest 'apiClient' cannot be nil")

		return nil, fmt.Errorf("certificateSigningRequest 'apiClient' cannot be nil")
	}

	err := apiClient.AttachScheme(certificatesv1.AddToScheme)
	if err != nil {
		glog.V(100).Infof("Failed to add certificates v1 scheme to client schemes")

		return nil, err
	}

	if len(options) > 1 {
		glog.V(100).Infof("Only one ListOptions object can be provided to ListSigningRequests")

		return nil, fmt.Errorf("error: more than one ListOptions was passed")
	}

	logMessage := "Listing all CertificateSigningRequests"
	passedOptions := runtimeclient.ListOptions{}

	if len(options) == 1 {
		passedOptions = options[0]
		logMessage += fmt.Sprintf(" with options: %v", passedOptions)
	}

	glog.V(100).Info(logMessage)

	csrList := new(certificatesv1.CertificateSigningRequestList)
	err = apiClient.Client.List(context.TODO(), csrList, &passedOptions)

	if err != nil {
		glog.V(100).Infof("Failed to list CertificateSigningRequests: %v", err)

		return nil, err
	}

	var signingRequestBuilders []*SigningRequestBuilder

	for _, csr := range csrList.Items {
		copiedCSR := csr
		signingRequestBuilder := &SigningRequestBuilder{
			apiClient:  apiClient.Client,
			Definition: &copiedCSR,
			Object:     &copiedCSR,
		}

		signingRequestBuilders = append(signingRequestBuilders, signingRequestBuilder)
	}

	return signingRequestBuilders, nil
}

// WaitUntilSigningRequestsApproved polls the cluster for all CertificateSigningRequests with the provided options every
// 3 seconds for up to the timeout duration or until all CertificateSigningRequests are approved.
func WaitUntilSigningRequestsApproved(
	apiClient *clients.Settings, timeout time.Duration, options ...runtimeclient.ListOptions) error {
	if apiClient == nil {
		glog.V(100).Infof("CertificateSigningRequest 'apiClient' cannot be nil")

		return fmt.Errorf("certificateSigningRequest 'apiClient' cannot be nil")
	}

	if len(options) > 1 {
		glog.V(100).Infof("Only one ListOptions object can be provided to WaitUntilSigningRequestsApproved")

		return fmt.Errorf("error: more than one ListOptions was passed")
	}

	logMessage := "Waiting for all CertificateSigningRequests to be approved"
	passedOptions := runtimeclient.ListOptions{}

	if len(options) == 1 {
		passedOptions = options[0]
		logMessage += fmt.Sprintf(" with options: %v", passedOptions)
	}

	glog.V(100).Info(logMessage)

	return wait.PollUntilContextTimeout(
		context.TODO(), 3*time.Second, timeout, true, func(ctx context.Context) (bool, error) {
			signingRequests, err := ListSigningRequests(apiClient, passedOptions)
			if err != nil {
				glog.V(100).Infof("Failed to list CertificateSigningRequests: %v", err)

				return false, nil
			}

			for _, signingRequest := range signingRequests {
				if !slices.ContainsFunc(signingRequest.Object.Status.Conditions, approvedCondition) {
					glog.V(100).Infof("CertificateSigningRequest %s is not approved yet", signingRequest.Object.Name)

					return false, nil
				}
			}

			return true, nil
		})
}

func approvedCondition(cond certificatesv1.CertificateSigningRequestCondition) bool {
	return cond.Type == certificatesv1.CertificateApproved && cond.Status == corev1.ConditionTrue
}
