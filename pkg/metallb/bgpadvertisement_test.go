package metallb

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/metallb/mlbtypes"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	defaultBGPAdvertisementName   = "default-bgpadvert"
	defaultBGPAdvertisementNsName = "test-namespace"
)

func TestPullBGPAdvertisement(t *testing.T) {
	generateBGPAdvertisement := func(name, namespace string) *mlbtypes.BGPAdvertisement {
		return &mlbtypes.BGPAdvertisement{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: mlbtypes.BGPAdvertisementSpec{},
		}
	}

	testCases := []struct {
		name                string
		namespace           string
		addToRuntimeObjects bool
		expectedError       error
		client              bool
	}{
		{
			name:                "bgpadvertisement",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("bgpadvertisement 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                "bgpadvertisement",
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("bgpadvertisement 'namespace' cannot be empty"),
			client:              true,
		},
		{
			name:                "bgpadvertisement",
			namespace:           "test-namespace",
			addToRuntimeObjects: false,
			expectedError: fmt.Errorf(
				"bgpadvertisement object bgpadvertisement does not exist in namespace test-namespace"),
			client: true,
		},
		{
			name:                "bgpadvertisement",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("bgpadvertisement 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testBGPAdvertisement := generateBGPAdvertisement(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testBGPAdvertisement)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		builderResult, err := PullBGPAdvertisement(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, builderResult.Object.Name)
			assert.Equal(t, testCase.namespace, builderResult.Object.Namespace)
		}
	}
}

func TestNewBGPAdvertisementBuilder(t *testing.T) {
	generateBGPAdvertisement := NewBGPAdvertisementBuilder

	testCases := []struct {
		name          string
		namespace     string
		expectedError string
	}{
		{
			name:          "bgpadvertisement",
			namespace:     "test-namespace",
			expectedError: "",
		},
		{
			name:          "",
			namespace:     "test-namespace",
			expectedError: "BGPAdvertisement 'name' cannot be empty",
		},
		{
			name:          "bgpadvertisement",
			namespace:     "",
			expectedError: "BGPAdvertisement 'nsname' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{
			SchemeAttachers: testSchemes,
		})
		testBGPAdvertisementBuilder := generateBGPAdvertisement(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, testBGPAdvertisementBuilder.errorMsg)
		assert.NotNil(t, testBGPAdvertisementBuilder.Definition)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.name, testBGPAdvertisementBuilder.Definition.Name)
			assert.Equal(t, testCase.namespace, testBGPAdvertisementBuilder.Definition.Namespace)
		}
	}
}

func TestBGPAdvertisementExist(t *testing.T) {
	testCases := []struct {
		testBGPAdvertisement *BGPAdvertisementBuilder
		expectedStatus       bool
	}{
		{
			testBGPAdvertisement: buildValidBGPAdvertisementBuilder(buildBGPAdvertisementTestClientWithDummyObject()),
			expectedStatus:       true,
		},
		{
			testBGPAdvertisement: buildInValidBGPAdvertisementBuilder(buildBGPAdvertisementTestClientWithDummyObject()),
			expectedStatus:       false,
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, testCase.testBGPAdvertisement.Exists(), testCase.expectedStatus)
	}
}

func TestBGPAdvertisementGet(t *testing.T) {
	testCases := []struct {
		testBGPAdvertisement *BGPAdvertisementBuilder
		expectedError        error
	}{
		{
			testBGPAdvertisement: buildValidBGPAdvertisementBuilder(buildBGPAdvertisementTestClientWithDummyObject()),
			expectedError:        nil,
		},
		{
			testBGPAdvertisement: buildInValidBGPAdvertisementBuilder(buildBGPAdvertisementTestClientWithDummyObject()),
			expectedError:        fmt.Errorf("BGPAdvertisement 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		bgpAdvertisement, err := testCase.testBGPAdvertisement.Get()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, bgpAdvertisement.Name, testCase.testBGPAdvertisement.Definition.Name)
		}
	}
}

func TestBGPAdvertisementCreate(t *testing.T) {
	testCases := []struct {
		testBGPAdvertisement *BGPAdvertisementBuilder
		expectedError        error
	}{
		{
			testBGPAdvertisement: buildValidBGPAdvertisementBuilder(buildBGPAdvertisementTestClientWithDummyObject()),
			expectedError:        nil,
		},
		{
			testBGPAdvertisement: buildInValidBGPAdvertisementBuilder(buildBGPAdvertisementTestClientWithDummyObject()),
			expectedError:        fmt.Errorf("BGPAdvertisement 'nsname' cannot be empty"),
		},
	}
	for _, testCase := range testCases {
		bgpAdvertisement, err := testCase.testBGPAdvertisement.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, bgpAdvertisement.Definition, bgpAdvertisement.Object)
		}
	}
}

func TestBGPAdvertisementDelete(t *testing.T) {
	testCases := []struct {
		testBGPAdvertisement *BGPAdvertisementBuilder
		expectedError        error
	}{
		{
			testBGPAdvertisement: buildValidBGPAdvertisementBuilder(buildBGPAdvertisementTestClientWithDummyObject()),
			expectedError:        nil,
		},
		{
			testBGPAdvertisement: buildInValidBGPAdvertisementBuilder(buildBGPAdvertisementTestClientWithDummyObject()),
			expectedError:        fmt.Errorf("BGPAdvertisement 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		_, err := testCase.testBGPAdvertisement.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testBGPAdvertisement.Object)
		}
	}
}

func TestBGPAdvertisementUpdate(t *testing.T) {
	testCases := []struct {
		testBGPAdvertisement *BGPAdvertisementBuilder
		expectedError        error
		ipAddressPool        []string
	}{
		{
			testBGPAdvertisement: buildValidBGPAdvertisementBuilder(buildBGPAdvertisementTestClientWithDummyObject()),
			expectedError:        nil,
			ipAddressPool:        []string{"1.1.1.1."},
		},
		{
			testBGPAdvertisement: buildInValidBGPAdvertisementBuilder(buildBGPAdvertisementTestClientWithDummyObject()),
			expectedError:        fmt.Errorf("BGPAdvertisement 'nsname' cannot be empty"),
			ipAddressPool:        []string{"1.1.1.1."},
		},
	}

	for _, testCase := range testCases {
		assert.Nil(t, testCase.testBGPAdvertisement.Definition.Spec.IPAddressPools)
		assert.Nil(t, nil, testCase.testBGPAdvertisement.Object)
		testCase.testBGPAdvertisement.WithIPAddressPools(testCase.ipAddressPool)
		testCase.testBGPAdvertisement.Definition.ResourceVersion = "999"
		_, err := testCase.testBGPAdvertisement.Update(false)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.ipAddressPool, testCase.testBGPAdvertisement.Definition.Spec.IPAddressPools)
		}
	}
}

func TestBGPAdvertisementWithAggregationLength4(t *testing.T) {
	testCases := []struct {
		testBGPAdvertisement *BGPAdvertisementBuilder
		expectedError        string
		aggregationLength    int32
	}{
		{
			testBGPAdvertisement: buildValidBGPAdvertisementBuilder(buildBGPAdvertisementTestClientWithDummyObject()),
			expectedError:        "",
			aggregationLength:    32,
		},
		{
			testBGPAdvertisement: buildValidBGPAdvertisementBuilder(buildBGPAdvertisementTestClientWithDummyObject()),
			expectedError:        "AggregationLength 64 is invalid, the value shoud be in range 0...32",
			aggregationLength:    64,
		},
		{
			testBGPAdvertisement: buildInValidBGPAdvertisementBuilder(buildBGPAdvertisementTestClientWithDummyObject()),
			expectedError:        "BGPAdvertisement 'nsname' cannot be empty",
			aggregationLength:    64,
		},
	}

	for _, testCase := range testCases {
		bGPAdvertisement := testCase.testBGPAdvertisement.WithAggregationLength4(testCase.aggregationLength)
		assert.Equal(t, testCase.expectedError, bGPAdvertisement.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, &testCase.aggregationLength, bGPAdvertisement.Definition.Spec.AggregationLength)
		}
	}
}

func TestBGPAdvertisementWithAggregationLength6(t *testing.T) {
	testCases := []struct {
		testBGPAdvertisement *BGPAdvertisementBuilder
		expectedError        string
		aggregationLength    int32
	}{
		{
			testBGPAdvertisement: buildValidBGPAdvertisementBuilder(buildBGPAdvertisementTestClientWithDummyObject()),
			expectedError:        "",
			aggregationLength:    32,
		},
		{
			testBGPAdvertisement: buildValidBGPAdvertisementBuilder(buildBGPAdvertisementTestClientWithDummyObject()),
			expectedError:        "AggregationLength 200 is invalid, the value shoud be in range 0...128",
			aggregationLength:    200,
		},
		{
			testBGPAdvertisement: buildInValidBGPAdvertisementBuilder(buildBGPAdvertisementTestClientWithDummyObject()),
			expectedError:        "BGPAdvertisement 'nsname' cannot be empty",
			aggregationLength:    64,
		},
	}

	for _, testCase := range testCases {
		bGPAdvertisement := testCase.testBGPAdvertisement.WithAggregationLength6(testCase.aggregationLength)
		assert.Equal(t, testCase.expectedError, bGPAdvertisement.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, &testCase.aggregationLength, bGPAdvertisement.Definition.Spec.AggregationLengthV6)
		}
	}
}

func TestBGPAdvertisementWithLocalPref(t *testing.T) {
	testCases := []struct {
		testBGPAdvertisement *BGPAdvertisementBuilder
		expectedError        string
		localPref            uint32
	}{
		{
			testBGPAdvertisement: buildValidBGPAdvertisementBuilder(buildBGPAdvertisementTestClientWithDummyObject()),
			expectedError:        "",
			localPref:            32,
		},
		{
			testBGPAdvertisement: buildInValidBGPAdvertisementBuilder(buildBGPAdvertisementTestClientWithDummyObject()),
			expectedError:        "BGPAdvertisement 'nsname' cannot be empty",
			localPref:            64,
		},
	}

	for _, testCase := range testCases {
		bGPAdvertisement := testCase.testBGPAdvertisement.WithLocalPref(testCase.localPref)
		assert.Equal(t, testCase.expectedError, bGPAdvertisement.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.localPref, bGPAdvertisement.Definition.Spec.LocalPref)
		}
	}
}

func TestBGPAdvertisementWithCommunities(t *testing.T) {
	testCases := []struct {
		testBGPAdvertisement *BGPAdvertisementBuilder
		expectedError        string
		community            []string
	}{
		{
			testBGPAdvertisement: buildValidBGPAdvertisementBuilder(buildBGPAdvertisementTestClientWithDummyObject()),
			expectedError:        "",
			community:            []string{"5252"},
		},
		{
			testBGPAdvertisement: buildValidBGPAdvertisementBuilder(buildBGPAdvertisementTestClientWithDummyObject()),
			expectedError:        "error: community setting is empty list, the list should contain at least one element",
			community:            []string{},
		},
		{
			testBGPAdvertisement: buildInValidBGPAdvertisementBuilder(buildBGPAdvertisementTestClientWithDummyObject()),
			expectedError:        "BGPAdvertisement 'nsname' cannot be empty",
			community:            []string{"5252"},
		},
	}

	for _, testCase := range testCases {
		bGPAdvertisement := testCase.testBGPAdvertisement.WithCommunities(testCase.community)
		assert.Equal(t, testCase.expectedError, bGPAdvertisement.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.community, bGPAdvertisement.Definition.Spec.Communities)
		}
	}
}

func TestBGPAdvertisementWithIPAddressPools(t *testing.T) {
	testCases := []struct {
		testBGPAdvertisement *BGPAdvertisementBuilder
		expectedError        string
		addressPool          []string
	}{
		{
			testBGPAdvertisement: buildValidBGPAdvertisementBuilder(buildBGPAdvertisementTestClientWithDummyObject()),
			expectedError:        "",
			addressPool:          []string{"1.1.1.1-1.1.1.2"},
		},
		{
			testBGPAdvertisement: buildValidBGPAdvertisementBuilder(buildBGPAdvertisementTestClientWithDummyObject()),
			expectedError:        "error: IPAddressPools setting is empty list, the list should contain at least one element",
			addressPool:          []string{},
		},
		{
			testBGPAdvertisement: buildInValidBGPAdvertisementBuilder(buildBGPAdvertisementTestClientWithDummyObject()),
			expectedError:        "BGPAdvertisement 'nsname' cannot be empty",
			addressPool:          []string{"5252"},
		},
	}

	for _, testCase := range testCases {
		bGPAdvertisement := testCase.testBGPAdvertisement.WithIPAddressPools(testCase.addressPool)
		assert.Equal(t, testCase.expectedError, bGPAdvertisement.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.addressPool, bGPAdvertisement.Definition.Spec.IPAddressPools)
		}
	}
}

func TestBGPAdvertisementWithIPAddressPoolsSelectors(t *testing.T) {
	testCases := []struct {
		testBGPAdvertisement *BGPAdvertisementBuilder
		expectedError        string
		poolSelector         []metav1.LabelSelector
	}{
		{
			testBGPAdvertisement: buildValidBGPAdvertisementBuilder(buildBGPAdvertisementTestClientWithDummyObject()),
			expectedError:        "",
			poolSelector:         []metav1.LabelSelector{{MatchLabels: map[string]string{"test": "test1"}}},
		},
		{
			testBGPAdvertisement: buildValidBGPAdvertisementBuilder(buildBGPAdvertisementTestClientWithDummyObject()),
			expectedError: "error: IPAddressPoolSelectors setting is empty list, the list should contain at " +
				"least one element",
			poolSelector: []metav1.LabelSelector{},
		},
		{
			testBGPAdvertisement: buildInValidBGPAdvertisementBuilder(buildBGPAdvertisementTestClientWithDummyObject()),
			expectedError:        "BGPAdvertisement 'nsname' cannot be empty",
			poolSelector:         []metav1.LabelSelector{{MatchLabels: map[string]string{"test": "test1"}}},
		},
	}

	for _, testCase := range testCases {
		bGPAdvertisement := testCase.testBGPAdvertisement.WithIPAddressPoolsSelectors(testCase.poolSelector)
		assert.Equal(t, testCase.expectedError, bGPAdvertisement.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.poolSelector, bGPAdvertisement.Definition.Spec.IPAddressPoolSelectors)
		}
	}
}

func TestBGPAdvertisementWithNodeSelector(t *testing.T) {
	testCases := []struct {
		testBGPAdvertisement *BGPAdvertisementBuilder
		expectedError        string
		nodeSelector         []metav1.LabelSelector
	}{
		{
			testBGPAdvertisement: buildValidBGPAdvertisementBuilder(buildBGPAdvertisementTestClientWithDummyObject()),
			expectedError:        "",
			nodeSelector:         []metav1.LabelSelector{{MatchLabels: map[string]string{"test": "test1"}}},
		},
		{
			testBGPAdvertisement: buildValidBGPAdvertisementBuilder(buildBGPAdvertisementTestClientWithDummyObject()),
			expectedError:        "error: nodeSelectors setting is empty list, the list should contain at least one element",
			nodeSelector:         []metav1.LabelSelector{},
		},
		{
			testBGPAdvertisement: buildInValidBGPAdvertisementBuilder(buildBGPAdvertisementTestClientWithDummyObject()),
			expectedError:        "BGPAdvertisement 'nsname' cannot be empty",
			nodeSelector:         []metav1.LabelSelector{{MatchLabels: map[string]string{"test": "test1"}}},
		},
	}

	for _, testCase := range testCases {
		bGPAdvertisement := testCase.testBGPAdvertisement.WithNodeSelector(testCase.nodeSelector)
		assert.Equal(t, testCase.expectedError, bGPAdvertisement.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.nodeSelector, bGPAdvertisement.Definition.Spec.NodeSelectors)
		}
	}
}

func TestBGPAdvertisementWithPeers(t *testing.T) {
	testCases := []struct {
		testBGPAdvertisement *BGPAdvertisementBuilder
		expectedError        string
		peers                []string
	}{
		{
			testBGPAdvertisement: buildValidBGPAdvertisementBuilder(buildBGPAdvertisementTestClientWithDummyObject()),
			expectedError:        "",
			peers:                []string{"test", "test1"},
		},
		{
			testBGPAdvertisement: buildValidBGPAdvertisementBuilder(buildBGPAdvertisementTestClientWithDummyObject()),
			expectedError:        "error: peers setting is empty list, the list should contain at least one element",
			peers:                []string{},
		},
		{
			testBGPAdvertisement: buildInValidBGPAdvertisementBuilder(buildBGPAdvertisementTestClientWithDummyObject()),
			expectedError:        "BGPAdvertisement 'nsname' cannot be empty",
			peers:                []string{"test", "test1"},
		},
	}

	for _, testCase := range testCases {
		bGPAdvertisement := testCase.testBGPAdvertisement.WithPeers(testCase.peers)
		assert.Equal(t, testCase.expectedError, bGPAdvertisement.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.peers, bGPAdvertisement.Definition.Spec.Peers)
		}
	}
}

func TestBGPAdvertisementWithOptions(t *testing.T) {
	testSettings := buildBGPAdvertisementTestClientWithDummyObject()
	testBuilder := buildValidBGPAdvertisementBuilder(testSettings).WithOptions(
		func(builder *BGPAdvertisementBuilder) (*BGPAdvertisementBuilder, error) {
			return builder, nil
		})

	assert.Equal(t, "", testBuilder.errorMsg)
	testBuilder = buildValidBGPAdvertisementBuilder(testSettings).WithOptions(
		func(builder *BGPAdvertisementBuilder) (*BGPAdvertisementBuilder, error) {
			return builder, fmt.Errorf("error")
		})
	assert.Equal(t, "error", testBuilder.errorMsg)
}

func TestBGPAdvertisementGVR(t *testing.T) {
	assert.Equal(t, GetBGPAdvertisementGVR(),
		schema.GroupVersionResource{
			Group: APIGroup, Version: APIVersion, Resource: "bgpadvertisements",
		})
}

func buildValidBGPAdvertisementBuilder(apiClient *clients.Settings) *BGPAdvertisementBuilder {
	return NewBGPAdvertisementBuilder(
		apiClient, defaultBGPAdvertisementName, defaultBGPAdvertisementNsName)
}

func buildInValidBGPAdvertisementBuilder(apiClient *clients.Settings) *BGPAdvertisementBuilder {
	return NewBGPAdvertisementBuilder(
		apiClient, defaultBGPAdvertisementName, "")
}

func buildBGPAdvertisementTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyBGPAdvertisement(),
		SchemeAttachers: testSchemes,
	})
}

func buildDummyBGPAdvertisement() []runtime.Object {
	return append([]runtime.Object{}, &mlbtypes.BGPAdvertisement{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultBGPAdvertisementName,
			Namespace: defaultBGPAdvertisementNsName,
		},
		Spec: mlbtypes.BGPAdvertisementSpec{},
	})
}
