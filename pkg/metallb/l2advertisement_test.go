package metallb

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/schemes/metallb/mlbtypes"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	defaultL2AdvertisementName   = "default-l2advertisement"
	defaultL2AdvertisementNsName = "test-namespace"
)

func TestPullL2Advertisement(t *testing.T) {
	generatel2Advertisement := func(name, namespace string) *mlbtypes.L2Advertisement {
		return &mlbtypes.L2Advertisement{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: mlbtypes.L2AdvertisementSpec{},
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
			name:                "l2advertisement",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("l2advertisement 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                "l2advertisement",
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("l2advertisement 'namespace' cannot be empty"),
			client:              true,
		},
		{
			name:                "l2advertisement",
			namespace:           "test-namespace",
			addToRuntimeObjects: false,
			expectedError:       fmt.Errorf("l2advertisement object l2advertisement does not exist in namespace test-namespace"),
			client:              true,
		},
		{
			name:                "l2advertisement",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("l2Advertisement 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testL2Advertisement := generatel2Advertisement(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testL2Advertisement)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		builderResult, err := PullL2Advertisement(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, builderResult.Object.Name)
			assert.Equal(t, testCase.namespace, builderResult.Object.Namespace)
		}
	}
}

func TestNewL2AdvertisementBuilder(t *testing.T) {
	generateL2AdvertisementBuilder := NewL2AdvertisementBuilder

	testCases := []struct {
		name          string
		namespace     string
		expectedError string
	}{
		{
			name:          "l2Advertisement",
			namespace:     "test-namespace",
			expectedError: "",
		},

		{
			name:          "",
			namespace:     "test-namespace",
			expectedError: "L2Advertisement 'name' cannot be empty",
		},
		{
			name:          "l2Advertisement",
			namespace:     "",
			expectedError: "L2Advertisement 'nsname' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{
			SchemeAttachers: testSchemes})
		testL2AdvertisementBuilder := generateL2AdvertisementBuilder(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, testL2AdvertisementBuilder.errorMsg)
		assert.NotNil(t, testL2AdvertisementBuilder.Definition)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.name, testL2AdvertisementBuilder.Definition.Name)
			assert.Equal(t, testCase.namespace, testL2AdvertisementBuilder.Definition.Namespace)
		}
	}
}

func TestL2AdvertisementExist(t *testing.T) {
	testCases := []struct {
		testL2Advertisement *L2AdvertisementBuilder
		expectedStatus      bool
	}{
		{
			testL2Advertisement: buildValidL2AdvertisementBuilder(buildL2AdvertisementTestClientWithDummyObject()),
			expectedStatus:      true,
		},
		{
			testL2Advertisement: buildInValidL2AdvertisementBuilder(buildL2AdvertisementTestClientWithDummyObject()),
			expectedStatus:      false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testL2Advertisement.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestL2AdvertisementGet(t *testing.T) {
	testCases := []struct {
		testL2Advertisement *L2AdvertisementBuilder
		expectedError       error
	}{
		{
			testL2Advertisement: buildValidL2AdvertisementBuilder(buildL2AdvertisementTestClientWithDummyObject()),
			expectedError:       nil,
		},
		{
			testL2Advertisement: buildInValidL2AdvertisementBuilder(buildL2AdvertisementTestClientWithDummyObject()),
			expectedError:       fmt.Errorf("L2Advertisement 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		l2Advertisement, err := testCase.testL2Advertisement.Get()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, l2Advertisement.Name, testCase.testL2Advertisement.Definition.Name)
		}
	}
}

func TestL2AdvertisementCreate(t *testing.T) {
	testCases := []struct {
		testL2Advertisement *L2AdvertisementBuilder
		expectedError       error
	}{
		{
			testL2Advertisement: buildValidL2AdvertisementBuilder(buildL2AdvertisementTestClientWithDummyObject()),
			expectedError:       nil,
		},
		{
			testL2Advertisement: buildInValidL2AdvertisementBuilder(buildL2AdvertisementTestClientWithDummyObject()),
			expectedError:       fmt.Errorf("L2Advertisement 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		ipAddressPoolBuilder, err := testCase.testL2Advertisement.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, ipAddressPoolBuilder.Definition.Name, ipAddressPoolBuilder.Object.Name)
		}
	}
}

func TestL2AdvertisementDelete(t *testing.T) {
	testCases := []struct {
		testL2Advertisement *L2AdvertisementBuilder
		expectedError       error
	}{
		{
			testL2Advertisement: buildValidL2AdvertisementBuilder(buildL2AdvertisementTestClientWithDummyObject()),
			expectedError:       nil,
		},
		{
			testL2Advertisement: buildInValidL2AdvertisementBuilder(buildL2AdvertisementTestClientWithDummyObject()),
			expectedError:       fmt.Errorf("L2Advertisement 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		_, err := testCase.testL2Advertisement.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testL2Advertisement.Object)
		}
	}
}

func TestL2AdvertisementUpdate(t *testing.T) {
	testCases := []struct {
		testL2Advertisement *L2AdvertisementBuilder
		expectedError       error
		addressPool         []string
	}{
		{
			testL2Advertisement: buildValidL2AdvertisementBuilder(buildL2AdvertisementTestClientWithDummyObject()),
			expectedError:       nil,
			addressPool:         []string{"1.1.1.1-1.1.1.2"},
		},
		{
			testL2Advertisement: buildValidL2AdvertisementBuilder(buildL2AdvertisementTestClientWithDummyObject()),
			expectedError: fmt.Errorf("error: IPAddressPools setting is empty list, the list should " +
				"contain at least one element"),
			addressPool: []string{},
		},
		{
			testL2Advertisement: buildInValidL2AdvertisementBuilder(buildL2AdvertisementTestClientWithDummyObject()),
			expectedError:       fmt.Errorf("L2Advertisement 'nsname' cannot be empty"),
			addressPool:         []string{"1.1.1.1-1.1.1.2"},
		},
	}

	for _, testCase := range testCases {
		assert.Nil(t, testCase.testL2Advertisement.Definition.Spec.IPAddressPools)
		assert.Nil(t, nil, testCase.testL2Advertisement.Object)
		testCase.testL2Advertisement.WithIPAddressPools(testCase.addressPool)
		testCase.testL2Advertisement.Definition.ResourceVersion = "999"
		_, err := testCase.testL2Advertisement.Update(false)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.addressPool, testCase.testL2Advertisement.Definition.Spec.IPAddressPools)
		}
	}
}

func TestL2AdvertisementWithNodeSelector(t *testing.T) {
	testCases := []struct {
		testL2Advertisement *L2AdvertisementBuilder
		nodeSelector        []metav1.LabelSelector
		expectedError       string
	}{
		{
			testL2Advertisement: buildValidL2AdvertisementBuilder(buildL2AdvertisementTestClientWithDummyObject()),
			nodeSelector:        []metav1.LabelSelector{{MatchLabels: map[string]string{"test": "test1"}}},
		},
		{
			testL2Advertisement: buildInValidL2AdvertisementBuilder(buildL2AdvertisementTestClientWithDummyObject()),
			nodeSelector:        []metav1.LabelSelector{{MatchLabels: map[string]string{"test": "test1"}}},
			expectedError:       "L2Advertisement 'nsname' cannot be empty",
		},
		{
			testL2Advertisement: buildValidL2AdvertisementBuilder(buildL2AdvertisementTestClientWithDummyObject()),
			nodeSelector:        []metav1.LabelSelector{},
			expectedError:       "error: nodeSelectors setting is empty list, the list should contain at least one element",
		},
	}

	for _, testCase := range testCases {
		l2AdvertisementBuilder := testCase.testL2Advertisement.WithNodeSelector(testCase.nodeSelector)
		assert.Equal(t, testCase.expectedError, l2AdvertisementBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.nodeSelector, l2AdvertisementBuilder.Definition.Spec.NodeSelectors)
		}
	}
}

func TestL2AdvertisementWithIPAddressPools(t *testing.T) {
	testCases := []struct {
		testL2Advertisement *L2AdvertisementBuilder
		expectedError       string
		addressPool         []string
	}{
		{
			testL2Advertisement: buildValidL2AdvertisementBuilder(buildL2AdvertisementTestClientWithDummyObject()),
			expectedError:       "",
			addressPool:         []string{"1.1.1.1-1.1.1.2"},
		},
		{
			testL2Advertisement: buildValidL2AdvertisementBuilder(buildL2AdvertisementTestClientWithDummyObject()),
			expectedError:       "error: IPAddressPools setting is empty list, the list should contain at least one element",
			addressPool:         []string{},
		},
		{
			testL2Advertisement: buildInValidL2AdvertisementBuilder(buildL2AdvertisementTestClientWithDummyObject()),
			expectedError:       "L2Advertisement 'nsname' cannot be empty",
			addressPool:         []string{"1.1.1.1-1.1.1.2"},
		},
	}

	for _, testCase := range testCases {
		l2AdvertisementBuilder := testCase.testL2Advertisement.WithIPAddressPools(testCase.addressPool)
		assert.Equal(t, testCase.expectedError, l2AdvertisementBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.addressPool, l2AdvertisementBuilder.Definition.Spec.IPAddressPools)
		}
	}
}

func TestL2AdvertisementWithIPAddressPoolsSelectors(t *testing.T) {
	testCases := []struct {
		testL2Advertisement *L2AdvertisementBuilder
		poolSelector        []metav1.LabelSelector
		expectedError       string
	}{
		{
			testL2Advertisement: buildValidL2AdvertisementBuilder(buildL2AdvertisementTestClientWithDummyObject()),
			poolSelector:        []metav1.LabelSelector{{MatchLabels: map[string]string{"test": "test1"}}},
		},
		{
			testL2Advertisement: buildInValidL2AdvertisementBuilder(buildL2AdvertisementTestClientWithDummyObject()),
			poolSelector:        []metav1.LabelSelector{{MatchLabels: map[string]string{"test": "test1"}}},
			expectedError:       "L2Advertisement 'nsname' cannot be empty",
		},
		{
			testL2Advertisement: buildValidL2AdvertisementBuilder(buildL2AdvertisementTestClientWithDummyObject()),
			poolSelector:        []metav1.LabelSelector{},
			expectedError: "error: IPAddressPoolSelectors setting is empty list, " +
				"the list should contain at least one element",
		},
	}

	for _, testCase := range testCases {
		l2AdvertisementBuilder := testCase.testL2Advertisement.WithIPAddressPoolsSelectors(testCase.poolSelector)
		assert.Equal(t, testCase.expectedError, l2AdvertisementBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.poolSelector, l2AdvertisementBuilder.Definition.Spec.IPAddressPoolSelectors)
		}
	}
}

func TestL2AdvertisementWithInterfaces(t *testing.T) {
	testCases := []struct {
		testL2Advertisement *L2AdvertisementBuilder
		interfaces          []string
		expectedError       string
	}{
		{
			testL2Advertisement: buildValidL2AdvertisementBuilder(buildL2AdvertisementTestClientWithDummyObject()),
			interfaces:          []string{"eno1"},
		},
		{
			testL2Advertisement: buildInValidL2AdvertisementBuilder(buildL2AdvertisementTestClientWithDummyObject()),
			interfaces:          []string{"eno1"},
			expectedError:       "L2Advertisement 'nsname' cannot be empty",
		},
		{
			testL2Advertisement: buildValidL2AdvertisementBuilder(buildL2AdvertisementTestClientWithDummyObject()),
			interfaces:          []string{},
			expectedError:       "error: Interfaces setting is empty list, the list should contain at least one element",
		},
	}

	for _, testCase := range testCases {
		l2AdvertisementBuilder := testCase.testL2Advertisement.WithInterfaces(testCase.interfaces)
		assert.Equal(t, testCase.expectedError, l2AdvertisementBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.interfaces, l2AdvertisementBuilder.Definition.Spec.Interfaces)
		}
	}
}

func TestL2AdvertisementWithOptions(t *testing.T) {
	testSettings := buildL2AdvertisementTestClientWithDummyObject()
	testBuilder := buildValidL2AdvertisementBuilder(testSettings).WithOptions(
		func(builder *L2AdvertisementBuilder) (*L2AdvertisementBuilder, error) {
			return builder, nil
		})

	assert.Equal(t, "", testBuilder.errorMsg)
	testBuilder = buildValidL2AdvertisementBuilder(testSettings).WithOptions(
		func(builder *L2AdvertisementBuilder) (*L2AdvertisementBuilder, error) {
			return builder, fmt.Errorf("error")
		})
	assert.Equal(t, "error", testBuilder.errorMsg)
}

func TestL2AdvertisementGVR(t *testing.T) {
	assert.Equal(t, GetIPAddressPoolGVR(),
		schema.GroupVersionResource{
			Group: APIGroup, Version: APIVersion, Resource: "ipaddresspools",
		})
}

func buildValidL2AdvertisementBuilder(apiClient *clients.Settings) *L2AdvertisementBuilder {
	return NewL2AdvertisementBuilder(
		apiClient, defaultL2AdvertisementName, defaultL2AdvertisementNsName)
}

func buildInValidL2AdvertisementBuilder(apiClient *clients.Settings) *L2AdvertisementBuilder {
	return NewL2AdvertisementBuilder(
		apiClient, defaultL2AdvertisementName, "")
}

func buildL2AdvertisementTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyL2Advertisement(),
		SchemeAttachers: testSchemes,
	})
}

func buildDummyL2Advertisement() []runtime.Object {
	return append([]runtime.Object{}, &mlbtypes.L2Advertisement{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultL2AdvertisementName,
			Namespace: defaultL2AdvertisementNsName,
		},
		Spec: mlbtypes.L2AdvertisementSpec{},
	})
}
