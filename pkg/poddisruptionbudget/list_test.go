package poddisruptionbudget

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	testSchemes = []clients.SchemeAttacher{
		policyv1.AddToScheme,
	}
	defaultPDBName   = "pdbtest"
	defaultPDBNsName = "pdbnamespace"
)

func TestPDBList(t *testing.T) {
	testCases := []struct {
		pdb           []*Builder
		nsName        string
		listOptions   []metav1.ListOptions
		expectedError error
		client        bool
	}{
		{
			pdb: []*Builder{
				buildValidPDBTestBuilder(buildTestClientWithDummyObject())},
			nsName:        defaultPDBNsName,
			expectedError: nil,
			client:        true,
		},
		{
			pdb: []*Builder{
				buildValidPDBTestBuilder(buildTestClientWithDummyObject())},
			nsName:        defaultPDBNsName,
			listOptions:   []metav1.ListOptions{{LabelSelector: "test1"}},
			expectedError: nil,
			client:        true,
		},
		{
			pdb: []*Builder{
				buildValidPDBTestBuilder(buildTestClientWithDummyObject())},
			nsName:        defaultPDBNsName,
			listOptions:   []metav1.ListOptions{{LabelSelector: ""}},
			expectedError: nil,
			client:        true,
		},
		{
			pdb: []*Builder{
				buildValidPDBTestBuilder(buildTestClientWithDummyObject())},
			nsName:        defaultPDBNsName,
			listOptions:   []metav1.ListOptions{{LabelSelector: "test1"}, {LabelSelector: "test2"}},
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
			client:        true,
		},
		{
			pdb: []*Builder{
				buildValidPDBTestBuilder(buildTestClientWithDummyObject())},
			nsName:        "",
			expectedError: fmt.Errorf("failed to list podDisruptionBudgets, 'nsname' parameter is empty"),
			client:        true,
		},
		{
			pdb: []*Builder{
				buildValidPDBTestBuilder(buildTestClientWithDummyObject())},
			nsName:        defaultPDBNsName,
			expectedError: fmt.Errorf("podDisruptionBudget 'apiClient' cannot be empty"),
			client:        false,
		},
	}
	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  buildDummyPDBObject(),
				SchemeAttachers: testSchemes,
			})
		}

		pdbBuilders, err := List(testSettings, testCase.nsName, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(pdbBuilders), len(testCase.pdb))
		}
	}
}

func TestPDBListInAllNamespaces(t *testing.T) {
	testCases := []struct {
		pdb           []*Builder
		listOptions   []metav1.ListOptions
		expectedError error
		client        bool
	}{
		{
			pdb: []*Builder{
				buildValidPDBTestBuilder(buildTestClientWithDummyObject())},
			expectedError: nil,
			client:        true,
		},
		{
			pdb: []*Builder{
				buildValidPDBTestBuilder(buildTestClientWithDummyObject())},
			listOptions:   []metav1.ListOptions{{LabelSelector: "test1"}},
			expectedError: nil,
			client:        true,
		},
		{
			pdb: []*Builder{
				buildValidPDBTestBuilder(buildTestClientWithDummyObject())},
			listOptions:   []metav1.ListOptions{{LabelSelector: ""}},
			expectedError: nil,
			client:        true,
		},
		{
			pdb: []*Builder{
				buildValidPDBTestBuilder(buildTestClientWithDummyObject())},
			listOptions:   []metav1.ListOptions{{LabelSelector: "test1"}, {LabelSelector: "test2"}},
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
			client:        true,
		},
		{
			pdb: []*Builder{
				buildValidPDBTestBuilder(buildTestClientWithDummyObject())},
			expectedError: fmt.Errorf("podDisruptionBudget 'apiClient' cannot be empty"),
			client:        false,
		},
	}
	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  buildDummyPDBObject(),
				SchemeAttachers: testSchemes,
			})
		}

		pdbBuilders, err := ListInAllNamespaces(testSettings, testCase.listOptions...)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(pdbBuilders), len(testCase.pdb))
		}
	}
}

// buildValidTestBuilder returns a valid Builder for testing purposes.
func buildValidPDBTestBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, defaultPDBName, defaultPDBNsName)
}

func buildTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyPDBObject(),
		SchemeAttachers: testSchemes,
	})
}

func buildDummyPDBObject() []runtime.Object {
	return append([]runtime.Object{}, &policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultPDBName,
			Namespace: defaultPDBNsName,
		},
		Spec: policyv1.PodDisruptionBudgetSpec{},
	})
}
