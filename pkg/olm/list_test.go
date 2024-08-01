package olm

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListCatalogSources(t *testing.T) {
	testCases := []struct {
		catalogSource []*CatalogSourceBuilder
		nsName        string
		listOptions   []client.ListOptions
		expectedError error
		client        bool
	}{
		{
			catalogSource: []*CatalogSourceBuilder{buildValidCatalogSourceBuilder(buildTestClientWithDummyObject())},
			nsName:        "test-namespace",
			expectedError: nil,
			client:        true,
		},
		{
			catalogSource: []*CatalogSourceBuilder{buildValidCatalogSourceBuilder(buildTestClientWithDummyObject())},
			nsName:        "",
			expectedError: fmt.Errorf("failed to list catalogsource, 'nsname' parameter is empty"),
			client:        true,
		},
		{
			catalogSource: []*CatalogSourceBuilder{buildValidCatalogSourceBuilder(buildTestClientWithDummyObject())},
			nsName:        "test-namespace",
			expectedError: nil,
			listOptions:   []client.ListOptions{{Continue: "true"}},
			client:        true,
		},
		{
			catalogSource: []*CatalogSourceBuilder{buildValidCatalogSourceBuilder(buildTestClientWithDummyObject())},
			nsName:        "test-namespace",
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
			listOptions:   []client.ListOptions{{Continue: "true"}, {Limit: 100}},
			client:        true,
		},
		{
			catalogSource: []*CatalogSourceBuilder{buildValidCatalogSourceBuilder(buildTestClientWithDummyObject())},
			nsName:        "test-namespace",
			expectedError: fmt.Errorf("failed to list catalogSource, 'apiClient' parameter is empty"),
			listOptions:   []client.ListOptions{},
			client:        false,
		},
	}
	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  buildDummyCatalogSource(),
				SchemeAttachers: testSchemes,
			})
		}

		netBuilders, err := ListCatalogSources(testSettings, testCase.nsName, testCase.listOptions...)
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(netBuilders), len(testCase.catalogSource))
		}
	}
}

func TestListClusterServiceVersion(t *testing.T) {
	testCases := []struct {
		clusterVersion []*ClusterServiceVersionBuilder
		nsName         string
		listOptions    []client.ListOptions
		expectedError  error
		client         bool
	}{
		{
			clusterVersion: []*ClusterServiceVersionBuilder{
				buildValidClusterServiceBuilder(buildTestClientWithDummyClusterServiceObject())},
			nsName:        "test-namespace",
			expectedError: nil,
			client:        true,
		},
		{
			clusterVersion: []*ClusterServiceVersionBuilder{
				buildValidClusterServiceBuilder(buildTestClientWithDummyClusterServiceObject())},
			nsName:        "",
			expectedError: fmt.Errorf("failed to list clusterserviceversion, 'nsname' parameter is empty"),
			client:        true,
		},
		{
			clusterVersion: []*ClusterServiceVersionBuilder{
				buildValidClusterServiceBuilder(buildTestClientWithDummyClusterServiceObject())},
			nsName:        "test-namespace",
			expectedError: nil,
			listOptions:   []client.ListOptions{{Continue: "true"}},
			client:        true,
		},
		{
			clusterVersion: []*ClusterServiceVersionBuilder{
				buildValidClusterServiceBuilder(buildTestClientWithDummyClusterServiceObject())},
			nsName:        "test-namespace",
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
			listOptions:   []client.ListOptions{{Continue: "true"}, {Limit: 100}},
			client:        true,
		},
		{
			clusterVersion: []*ClusterServiceVersionBuilder{
				buildValidClusterServiceBuilder(buildTestClientWithDummyClusterServiceObject())},
			nsName:        "test-namespace",
			expectedError: fmt.Errorf("clusterserviceversion 'apiClient' cannot be empty"),
			listOptions:   []client.ListOptions{},
			client:        false,
		},
	}
	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  buildDummyClusterService(nil, ""),
				SchemeAttachers: testSchemes,
			})
		}

		netBuilders, err := ListClusterServiceVersion(testSettings, testCase.nsName, testCase.listOptions...)
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(netBuilders), len(testCase.clusterVersion))
		}
	}
}

func TestListClusterServiceVersionWithNamePattern(t *testing.T) {
	testCases := []struct {
		clusterVersion []*ClusterServiceVersionBuilder
		nsName         string
		namePattern    string
		listOptions    []client.ListOptions
		expectedError  error
		client         bool
	}{
		{
			clusterVersion: []*ClusterServiceVersionBuilder{
				buildValidClusterServiceBuilder(buildTestClientWithDummyClusterServiceObject())},
			nsName:        "test-namespace",
			namePattern:   "cluster",
			expectedError: nil,
			client:        true,
		},
		{
			clusterVersion: []*ClusterServiceVersionBuilder{
				buildValidClusterServiceBuilder(buildTestClientWithDummyClusterServiceObject())},
			nsName:        "",
			namePattern:   "cluster",
			expectedError: fmt.Errorf("failed to list clusterserviceversion, 'nsname' parameter is empty"),
			client:        true,
		},
		{
			clusterVersion: []*ClusterServiceVersionBuilder{
				buildValidClusterServiceBuilder(buildTestClientWithDummyClusterServiceObject())},
			nsName:        "test-namespace",
			namePattern:   "",
			expectedError: fmt.Errorf("the namePattern field to filter out all relevant clusterserviceversion cannot be empty"),
			client:        true,
		},
		{
			clusterVersion: []*ClusterServiceVersionBuilder{
				buildValidClusterServiceBuilder(buildTestClientWithDummyClusterServiceObject())},
			nsName:        "test-namespace",
			namePattern:   "cluster",
			expectedError: nil,
			listOptions:   []client.ListOptions{{Continue: "true"}},
			client:        true,
		},
		{
			clusterVersion: []*ClusterServiceVersionBuilder{
				buildValidClusterServiceBuilder(buildTestClientWithDummyClusterServiceObject())},
			nsName:        "test-namespace",
			namePattern:   "cluster",
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
			listOptions:   []client.ListOptions{{Continue: "true"}, {Limit: 100}},
			client:        true,
		},
		{
			clusterVersion: []*ClusterServiceVersionBuilder{
				buildValidClusterServiceBuilder(buildTestClientWithDummyClusterServiceObject())},
			nsName:        "test-namespace",
			namePattern:   "cluster",
			expectedError: fmt.Errorf("clusterserviceversion 'apiClient' cannot be empty"),
			listOptions:   []client.ListOptions{},
			client:        false,
		},
	}
	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  buildDummyClusterService(nil, ""),
				SchemeAttachers: testSchemes,
			})
		}

		netBuilders, err := ListClusterServiceVersionWithNamePattern(
			testSettings, testCase.namePattern, testCase.nsName, testCase.listOptions...)
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(netBuilders), len(testCase.clusterVersion))
		}
	}
}

func TestListClusterServiceVersionInAllNamespaces(t *testing.T) {
	testCases := []struct {
		clusterVersion []*ClusterServiceVersionBuilder
		listOptions    []client.ListOptions
		expectedError  error
		client         bool
	}{
		{
			clusterVersion: []*ClusterServiceVersionBuilder{
				buildValidClusterServiceBuilder(buildTestClientWithDummyClusterServiceObject())},
			expectedError: nil,
			client:        true,
		},
		{
			clusterVersion: []*ClusterServiceVersionBuilder{
				buildValidClusterServiceBuilder(buildTestClientWithDummyClusterServiceObject())},
			expectedError: nil,
			listOptions:   []client.ListOptions{{Continue: "true"}},
			client:        true,
		},
		{
			clusterVersion: []*ClusterServiceVersionBuilder{
				buildValidClusterServiceBuilder(buildTestClientWithDummyClusterServiceObject())},
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
			listOptions:   []client.ListOptions{{Continue: "true"}, {Limit: 100}},
			client:        true,
		},
		{
			clusterVersion: []*ClusterServiceVersionBuilder{
				buildValidClusterServiceBuilder(buildTestClientWithDummyClusterServiceObject())},
			expectedError: fmt.Errorf("clusterserviceversion 'apiClient' cannot be empty"),
			listOptions:   []client.ListOptions{},
			client:        false,
		},
	}
	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  buildDummyClusterService(nil, ""),
				SchemeAttachers: testSchemes,
			})
		}

		netBuilders, err := ListClusterServiceVersionInAllNamespaces(testSettings, testCase.listOptions...)
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil && len(testCase.listOptions) == 0 {
			assert.Equal(t, len(netBuilders), len(testCase.clusterVersion))
		}
	}
}
