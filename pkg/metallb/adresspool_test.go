package metallb

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/metallb/mlbtypes"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	addressPoolGvk = schema.GroupVersionKind{
		Group:   APIGroup,
		Version: APIVersion,
		Kind:    ipAddressPoolKind,
	}
	defaultIPAddressPoolName = "default-pool"
	defaultNsName            = "test-namespace"
	defaultIPPoolRange       = []string{"1.1.1.1", "1.1.1.20"}
)

func TestPullAddressPool(t *testing.T) {
	generateIPAddressPool := func(name, namespace string) *mlbtypes.IPAddressPool {
		return &mlbtypes.IPAddressPool{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: mlbtypes.IPAddressPoolSpec{
				AvoidBuggyIPs: true,
			},
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
			name:                "addresspool",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("addresspool 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                "addresspool",
			namespace:           "test-namespace",
			addToRuntimeObjects: false,
			expectedError:       fmt.Errorf("addresspool object addresspool doesn't exist in namespace test-namespace"),
			client:              true,
		},
		{
			name:                "addresspool",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("addresspool 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testIPAddressPool := generateIPAddressPool(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testIPAddressPool)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
				GVK:            []schema.GroupVersionKind{addressPoolGvk},
			})
		}

		builderResult, err := PullAddressPool(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, builderResult.Object.Name)
			assert.Equal(t, testCase.namespace, builderResult.Object.Namespace)
		}
	}
}

func TestNewIPAddressPoolBuilder(t *testing.T) {
	generateIPAddressPoolBuilder := NewIPAddressPoolBuilder

	testCases := []struct {
		name          string
		namespace     string
		addrPool      []string
		expectedError string
	}{
		{
			name:          "addresspool",
			namespace:     "test-namespace",
			addrPool:      []string{"1.1.1.1", "1.1.1.20"},
			expectedError: "",
		},
		{
			name:          "",
			namespace:     "test-namespace",
			addrPool:      []string{"1.1.1.1", "1.1.1.20"},
			expectedError: "IPAddressPool 'name' cannot be empty",
		},
		{
			name:          "addresspool",
			namespace:     "",
			addrPool:      []string{"1.1.1.1", "1.1.1.20"},
			expectedError: "IPAddressPool 'nsname' cannot be empty",
		},
		{
			name:          "addresspool",
			namespace:     "test-namespace",
			addrPool:      []string{},
			expectedError: "IPAddressPool 'addrPool' cannot be empty list",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{
			GVK: []schema.GroupVersionKind{addressPoolGvk},
		})
		testIPAddressPoolBuilder := generateIPAddressPoolBuilder(
			testSettings, testCase.name, testCase.namespace, testCase.addrPool)
		assert.Equal(t, testCase.expectedError, testIPAddressPoolBuilder.errorMsg)
		assert.NotNil(t, testIPAddressPoolBuilder.Definition)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.name, testIPAddressPoolBuilder.Definition.Name)
			assert.Equal(t, testCase.namespace, testIPAddressPoolBuilder.Definition.Namespace)
		}
	}
}

func TestIPAddressPoolGet(t *testing.T) {
	testCases := []struct {
		testIPAddressPool *IPAddressPoolBuilder
		expectedError     error
	}{
		{
			testIPAddressPool: buildValidIPAddressPoolBuilder(buildTestClientWithDummyObject()),
			expectedError:     nil,
		},
		{
			testIPAddressPool: buildInValidIPAddressPoolBuilder(buildTestClientWithDummyObject()),
			expectedError:     fmt.Errorf("IPAddressPool 'addrPool' cannot be empty list"),
		},
	}

	for _, testCase := range testCases {
		ipAddressPool, err := testCase.testIPAddressPool.Get()
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil {
			assert.Equal(t, ipAddressPool, testCase.testIPAddressPool.Definition)
		}
	}
}

func TestIPAddressPoolExist(t *testing.T) {
	testCases := []struct {
		testIPAddressPool *IPAddressPoolBuilder
		expectedStatus    bool
	}{
		{
			testIPAddressPool: buildValidIPAddressPoolBuilder(buildTestClientWithDummyObject()),
			expectedStatus:    true,
		},
		{
			testIPAddressPool: buildInValidIPAddressPoolBuilder(buildTestClientWithDummyObject()),
			expectedStatus:    false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testIPAddressPool.Exists()
		assert.Equal(t, exist, testCase.expectedStatus)
	}
}

func TestIPAddressPoolCreate(t *testing.T) {
	testCases := []struct {
		testIPAddressPool *IPAddressPoolBuilder
		expectedError     error
	}{
		{
			testIPAddressPool: buildValidIPAddressPoolBuilder(buildTestClientWithDummyObject()),
			expectedError:     nil,
		},
		{
			testIPAddressPool: buildInValidIPAddressPoolBuilder(buildTestClientWithDummyObject()),
			expectedError:     fmt.Errorf("IPAddressPool 'addrPool' cannot be empty list"),
		},
	}

	for _, testCase := range testCases {
		ipAddressPoolBuilder, err := testCase.testIPAddressPool.Create()
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil {
			assert.Equal(t, ipAddressPoolBuilder.Definition, ipAddressPoolBuilder.Object)
		}
	}
}

func TestIPAddressPoolDelete(t *testing.T) {
	testCases := []struct {
		testIPAddressPool *IPAddressPoolBuilder
		expectedError     error
	}{
		{
			testIPAddressPool: buildValidIPAddressPoolBuilder(buildTestClientWithDummyObject()),
			expectedError:     nil,
		},
		{
			testIPAddressPool: buildInValidIPAddressPoolBuilder(buildTestClientWithDummyObject()),
			expectedError:     fmt.Errorf("IPAddressPool 'addrPool' cannot be empty list"),
		},
	}

	for _, testCase := range testCases {
		_, err := testCase.testIPAddressPool.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testIPAddressPool.Object)
		}
	}
}

func TestIPAddressPoolUpdate(t *testing.T) {
	testCases := []struct {
		testIPAddressPool *IPAddressPoolBuilder
		expectedError     error
		autoAssign        bool
	}{
		{
			testIPAddressPool: buildValidIPAddressPoolBuilder(buildTestClientWithDummyObject()),
			expectedError:     nil,
			autoAssign:        true,
		},
		{
			testIPAddressPool: buildInValidIPAddressPoolBuilder(buildTestClientWithDummyObject()),
			expectedError:     fmt.Errorf("IPAddressPool 'addrPool' cannot be empty list"),
			autoAssign:        false,
		},
	}

	for _, testCase := range testCases {
		assert.Nil(t, testCase.testIPAddressPool.Definition.Spec.AutoAssign)
		assert.Nil(t, nil, testCase.testIPAddressPool.Object)
		testCase.testIPAddressPool.WithAutoAssign(true)
		_, err := testCase.testIPAddressPool.Update(false)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, &testCase.autoAssign, testCase.testIPAddressPool.Definition.Spec.AutoAssign)
		}
	}
}

func TestIPAddressPoolWithAutoAssign(t *testing.T) {
	testCases := []struct {
		testIPAddressPool *IPAddressPoolBuilder
		autoAssign        bool
	}{
		{
			testIPAddressPool: buildValidIPAddressPoolBuilder(buildTestClientWithDummyObject()),
			autoAssign:        true,
		},
		{
			testIPAddressPool: buildValidIPAddressPoolBuilder(buildTestClientWithDummyObject()),
			autoAssign:        false,
		},
	}

	for _, testCase := range testCases {
		ipAddressPoolBuilder := testCase.testIPAddressPool.WithAutoAssign(testCase.autoAssign)
		assert.Equal(t, testCase.autoAssign, *ipAddressPoolBuilder.Definition.Spec.AutoAssign)
	}
}

func TestIPAddressPoolWithAvoidBuggyIPs(t *testing.T) {
	testCases := []struct {
		testIPAddressPool *IPAddressPoolBuilder
		avoidBuggyIPs     bool
	}{
		{
			testIPAddressPool: buildValidIPAddressPoolBuilder(buildTestClientWithDummyObject()),
			avoidBuggyIPs:     true,
		},
		{
			testIPAddressPool: buildValidIPAddressPoolBuilder(buildTestClientWithDummyObject()),
			avoidBuggyIPs:     false,
		},
	}

	for _, testCase := range testCases {
		ipAddressPoolBuilder := testCase.testIPAddressPool.WithAvoidBuggyIPs(testCase.avoidBuggyIPs)
		assert.Equal(t, testCase.avoidBuggyIPs, ipAddressPoolBuilder.Definition.Spec.AvoidBuggyIPs)
	}
}

func TestIPAddressPoolWithOptions(t *testing.T) {
	testSettings := buildTestClientWithDummyObject()
	testBuilder := buildValidIPAddressPoolBuilder(testSettings).WithOptions(
		func(builder *IPAddressPoolBuilder) (*IPAddressPoolBuilder, error) {
			return builder, nil
		})

	assert.Equal(t, "", testBuilder.errorMsg)
	testBuilder = buildValidIPAddressPoolBuilder(testSettings).WithOptions(
		func(builder *IPAddressPoolBuilder) (*IPAddressPoolBuilder, error) {
			return builder, fmt.Errorf("error")
		})
	assert.Equal(t, "error", testBuilder.errorMsg)
}

func TestGetIPAddressPoolGVR(t *testing.T) {
	assert.Equal(t, GetIPAddressPoolGVR(),
		schema.GroupVersionResource{
			Group: APIGroup, Version: APIVersion, Resource: "ipaddresspools",
		})
}

func buildValidIPAddressPoolBuilder(apiClient *clients.Settings) *IPAddressPoolBuilder {
	return NewIPAddressPoolBuilder(
		apiClient, defaultIPAddressPoolName, defaultNsName, defaultIPPoolRange)
}

func buildInValidIPAddressPoolBuilder(apiClient *clients.Settings) *IPAddressPoolBuilder {
	return NewIPAddressPoolBuilder(
		apiClient, defaultIPAddressPoolName, defaultNsName, []string{})
}

func buildTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: buildDummyIPAddressPool(),
		GVK:            []schema.GroupVersionKind{addressPoolGvk},
	})
}

func buildDummyIPAddressPool() []runtime.Object {
	return append([]runtime.Object{}, &mlbtypes.IPAddressPool{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultIPAddressPoolName,
			Namespace: defaultNsName,
		},
		Spec: mlbtypes.IPAddressPoolSpec{
			Addresses: defaultIPPoolRange,
		},
	})
}
