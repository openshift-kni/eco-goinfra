package egressip

import (
	"fmt"
	"testing"

	egressipv1 "github.com/ovn-org/ovn-kubernetes/go-controller/pkg/crd/egressip/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultEgressIPName            = "egress-test"
	defaultEgressIPs               = []string{"1.1.1.2", "1.1.1.3"}
	defaultEgressNamespaceSelector = metav1.LabelSelector{
		MatchLabels: map[string]string{"env": "qa"},
	}
	defaultEgressPodSelector = metav1.LabelSelector{
		MatchLabels: map[string]string{"env": "qa"},
	}
	testSchemes = []clients.SchemeAttacher{
		egressipv1.AddToScheme,
	}
)

func TestNewEgressIPBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		expectedError string
	}{
		{
			name:          "",
			expectedError: "the name parameter of the EgressIP is empty",
		},
		{
			name:          defaultEgressIPName,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})

		testEgressServiceBuilder := NewEgressIPBuilder(testSettings,
			testCase.name)

		assert.NotNil(t, testEgressServiceBuilder.Definition)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.name, testEgressServiceBuilder.Definition.Name)
			assert.Equal(t, "", testEgressServiceBuilder.errorMsg)
		} else {
			assert.Equal(t, testCase.expectedError, testEgressServiceBuilder.errorMsg)
		}
	}
}

func TestEgressIpWithEgressIPs(t *testing.T) {
	testCases := []struct {
		name          string
		egressIPs     []string
		expectedError string
	}{
		{
			name:          defaultEgressIPName,
			egressIPs:     []string{},
			expectedError: "cannot accept empty list as egressIPs value",
		},
		{
			name:          defaultEgressIPName,
			egressIPs:     []string{"10.10.11.2"},
			expectedError: "",
		},
		{
			name:          defaultEgressIPName,
			egressIPs:     []string{"10.10.11.2", "10.10.11.3", "10.10.11.4"},
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})

		testEgressIPBuilder := NewEgressIPBuilder(testSettings, testCase.name)

		assert.NotNil(t, testEgressIPBuilder.Definition)

		testEgressIPBuilder = testEgressIPBuilder.WithEgressIPs(testCase.egressIPs)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.name, testEgressIPBuilder.Definition.Name)
			assert.Equal(t, "", testEgressIPBuilder.errorMsg)
			assert.Equal(t, testCase.egressIPs, testEgressIPBuilder.Definition.Spec.EgressIPs)
		} else {
			assert.Equal(t, testCase.expectedError, testEgressIPBuilder.errorMsg)
		}
	}
}

func TestEgressIPWithNamespaceSelector(t *testing.T) {
	testCases := []struct {
		name              string
		namespaceSelector metav1.LabelSelector
		expectedError     string
	}{
		{
			name:              defaultEgressIPName,
			namespaceSelector: defaultEgressNamespaceSelector,
			expectedError:     "",
		},
		{
			name:              defaultEgressIPName,
			namespaceSelector: metav1.LabelSelector{},
			expectedError:     "EgressIP 'namespaceSelector' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})

		testEgressIPBuilder := NewEgressIPBuilder(testSettings, testCase.name)

		assert.NotNil(t, testEgressIPBuilder.Definition)

		testEgressIPBuilder = testEgressIPBuilder.WithNamespaceSelector(testCase.namespaceSelector)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.name, testEgressIPBuilder.Definition.Name)
			assert.Equal(t, testCase.namespaceSelector, testEgressIPBuilder.Definition.Spec.NamespaceSelector)
			assert.Equal(t, "", testEgressIPBuilder.errorMsg)
		} else {
			assert.Equal(t, testCase.expectedError, testEgressIPBuilder.errorMsg)
		}
	}
}

func TestEgressIPWithPodSelector(t *testing.T) {
	testCases := []struct {
		name          string
		podSelector   metav1.LabelSelector
		expectedError string
	}{
		{
			name:          defaultEgressIPName,
			podSelector:   defaultEgressPodSelector,
			expectedError: "",
		},
		{
			name:          defaultEgressIPName,
			podSelector:   metav1.LabelSelector{},
			expectedError: "EgressIP 'podSelector' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})

		testEgressIPBuilder := NewEgressIPBuilder(testSettings, testCase.name)

		assert.NotNil(t, testEgressIPBuilder.Definition)

		testEgressIPBuilder = testEgressIPBuilder.WithPodSelector(testCase.podSelector)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.name, testEgressIPBuilder.Definition.Name)
			assert.Equal(t, testCase.podSelector, testEgressIPBuilder.Definition.Spec.PodSelector)
			assert.Equal(t, "", testEgressIPBuilder.errorMsg)
		} else {
			assert.Equal(t, testCase.expectedError, testEgressIPBuilder.errorMsg)
		}
	}
}

func TestEgressIPPull(t *testing.T) {
	testCases := []struct {
		name                string
		addToRuntimeObjects bool
		client              bool
		expectedError       error
	}{
		{
			name:                defaultEgressIPName,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       nil,
		},
		{
			name:                "",
			addToRuntimeObjects: false,
			client:              true,
			expectedError:       fmt.Errorf("egressIP's name cannot be empty"),
		},
		{
			name:                defaultEgressIPName,
			addToRuntimeObjects: false,
			client:              false,
			expectedError:       fmt.Errorf("egressIP's 'apiClient' cannot be empty"),
		},
		{
			name:                defaultEgressIPName,
			addToRuntimeObjects: false,
			client:              true,
			expectedError:       fmt.Errorf("egressIP object %q does not exist", defaultEgressIPName),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testEgressIP := buildDummyEgressIP(testCase.name)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testEgressIP)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		testEgressIPBuilder, err := Pull(testSettings, testCase.name)

		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, testEgressIPBuilder.Definition.Name)
		}
	}
}

func TestEgressIPExist(t *testing.T) {
	testCases := []struct {
		egressIP       *EgressIPBuilder
		expectedStatus bool
	}{
		{
			egressIP:       buildDummyEgressIPBuilder(buildTestClientWithDummyEgressIP()),
			expectedStatus: true,
		},
		{
			egressIP:       buildInvalidDummyEgressIPBuilder(buildTestClientWithDummyEgressIP()),
			expectedStatus: false,
		},
		{
			egressIP: buildDummyEgressIPBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.egressIP.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestEgressIPDelete(t *testing.T) {
	testCases := []struct {
		egressIP      *EgressIPBuilder
		expectedError error
	}{
		{
			egressIP:      buildDummyEgressIPBuilder(buildTestClientWithDummyEgressIP()),
			expectedError: nil,
		},
		{
			egressIP:      buildInvalidDummyEgressIPBuilder(buildTestClientWithDummyEgressIP()),
			expectedError: fmt.Errorf("the name parameter of the EgressIP is empty"),
		},
		{
			egressIP: buildDummyEgressIPBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: testSchemes})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		egressIPBuilder, err := testCase.egressIP.Delete()

		if testCase.expectedError == nil {
			assert.Nil(t, egressIPBuilder.Object)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError, err)
		}
	}
}

func TestEgressIPGet(t *testing.T) {
	testCases := []struct {
		addToRuntimeObjects bool
		expectedError       error
	}{
		{
			addToRuntimeObjects: true,
			expectedError:       nil,
		},
		{
			addToRuntimeObjects: false,
			expectedError:       fmt.Errorf("the name parameter of the EgressIP is empty"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects      []runtime.Object
			testSettings        *clients.Settings
			testEgressIPBuilder *EgressIPBuilder
		)

		testEgressIP := buildDummyEgressIP(defaultEgressIPName)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testEgressIP)
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  runtimeObjects,
			SchemeAttachers: testSchemes,
		})

		if testCase.expectedError != nil {
			testEgressIPBuilder = buildInvalidDummyEgressIPBuilder(testSettings)
		} else {
			testEgressIPBuilder = buildDummyEgressIPBuilder(testSettings)
		}

		testEgressIP, err := testEgressIPBuilder.Get()

		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
			assert.Empty(t, testEgressIP)
		} else {
			assert.Equal(t, defaultEgressIPName, testEgressIP.ObjectMeta.Name)
		}
	}
}

func TestEgressIPUpdate(t *testing.T) {
	testCases := []struct {
		name          string
		egressIPs     []string
		expectedError error
	}{
		{
			name:          defaultEgressIPName,
			egressIPs:     []string{"10.11.12.13"},
			expectedError: nil,
		},
		{
			name:          defaultEgressIPName,
			egressIPs:     defaultEgressIPs,
			expectedError: nil,
		},
		{
			name:          defaultEgressIPName,
			egressIPs:     []string{},
			expectedError: fmt.Errorf("cannot accept empty list as egressIPs value"),
		},
	}

	for _, testCase := range testCases {
		testEgressIPBuilder, err := Pull(buildTestClientWithDummyEgressIP(), testCase.name)

		assert.Nil(t, err)

		testEgressIPBuilder = testEgressIPBuilder.WithEgressIPs(testCase.egressIPs)

		testEgressIPBuilder, err = testEgressIPBuilder.Update()

		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError, err)
		} else {
			assert.Equal(t, testCase.name, testEgressIPBuilder.Object.Name)

			assert.Equal(t, testCase.egressIPs, testEgressIPBuilder.Object.Spec.EgressIPs)
		}
	}
}

func TestEgressIPCreate(t *testing.T) {
	testCases := []struct {
		name          string
		egressIPs     []string
		expectedError error
	}{
		{
			name:          defaultEgressIPName,
			egressIPs:     defaultEgressIPs,
			expectedError: nil,
		},
		{
			name:          defaultEgressIPName,
			egressIPs:     []string{},
			expectedError: fmt.Errorf("cannot accept empty list as egressIPs value"),
		},
	}

	for _, testCase := range testCases {
		testEgressIPBuilder := buildDummyEgressIPBuilder(buildTestClientWithCustomizedEgressIP(testCase.name))

		createdEgressIP, err := testEgressIPBuilder.WithEgressIPs(testCase.egressIPs).Create()

		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError, err)
		} else {
			assert.Equal(t, testCase.name, createdEgressIP.Definition.Name)
			assert.Equal(t, testCase.egressIPs, createdEgressIP.Definition.Spec.EgressIPs)
		}
	}
}

//nolint:funlen
func TestEgressIPGetAssignedEgressIPMap(t *testing.T) {
	testCases := []struct {
		testEgressIPClientWithDummyObject *egressipv1.EgressIP
		items                             []egressipv1.EgressIPStatusItem
		itemsMap                          map[string]string
		addToRuntimeObjects               bool
		expectedError                     error
	}{
		{
			testEgressIPClientWithDummyObject: &egressipv1.EgressIP{
				ObjectMeta: metav1.ObjectMeta{
					Name: defaultEgressIPName,
				},
				Spec: egressipv1.EgressIPSpec{
					EgressIPs:         defaultEgressIPs,
					NamespaceSelector: defaultEgressNamespaceSelector,
				},
				Status: egressipv1.EgressIPStatus{
					Items: []egressipv1.EgressIPStatusItem{{
						Node:     "node-1",
						EgressIP: "1.1.1.2",
					}},
				},
			},
			itemsMap:            map[string]string{"node-1": "1.1.1.2"},
			addToRuntimeObjects: true,
			expectedError:       nil,
		},
		{
			testEgressIPClientWithDummyObject: &egressipv1.EgressIP{
				ObjectMeta: metav1.ObjectMeta{
					Name: defaultEgressIPName,
				},
				Spec: egressipv1.EgressIPSpec{
					EgressIPs:         defaultEgressIPs,
					NamespaceSelector: defaultEgressNamespaceSelector,
				},
				Status: egressipv1.EgressIPStatus{
					Items: []egressipv1.EgressIPStatusItem{{
						Node:     "node-1",
						EgressIP: "1.1.1.2",
					}, {
						Node:     "node-2",
						EgressIP: "1.1.1.3",
					}},
				},
			},
			itemsMap:            map[string]string{"node-1": "1.1.1.2", "node-2": "1.1.1.3"},
			addToRuntimeObjects: true,
			expectedError:       nil,
		},
		{
			testEgressIPClientWithDummyObject: &egressipv1.EgressIP{
				ObjectMeta: metav1.ObjectMeta{
					Name: defaultEgressIPName,
				},
				Spec: egressipv1.EgressIPSpec{
					EgressIPs:         defaultEgressIPs,
					NamespaceSelector: defaultEgressNamespaceSelector,
				},
				Status: egressipv1.EgressIPStatus{
					Items: nil,
				},
			},
			itemsMap:            nil,
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("egressIP \"egress-test\" nodes assignment does not exist"),
		},
		{
			testEgressIPClientWithDummyObject: &egressipv1.EgressIP{
				ObjectMeta: metav1.ObjectMeta{
					Name: defaultEgressIPName,
				},
				Spec: egressipv1.EgressIPSpec{
					EgressIPs:         defaultEgressIPs,
					NamespaceSelector: defaultEgressNamespaceSelector,
				},
				Status: egressipv1.EgressIPStatus{
					Items: []egressipv1.EgressIPStatusItem{{
						Node:     "node-1",
						EgressIP: "1.1.1.2",
					}, {
						Node:     "node-2",
						EgressIP: "1.1.1.3",
					}},
				},
			},
			itemsMap:            map[string]string{"node-1": "1.1.1.2", "node-2": "1.1.1.3"},
			addToRuntimeObjects: false,
			expectedError:       fmt.Errorf("egressIP \"egress-test\" object does not exist"),
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testCase.testEgressIPClientWithDummyObject)
		}

		dummyEgressIPClient := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  runtimeObjects,
			SchemeAttachers: testSchemes,
		})

		testEgressIPBuilder := buildDummyEgressIPBuilder(dummyEgressIPClient)

		testEgressIPAssignmentMap, err := testEgressIPBuilder.GetAssignedEgressIPMap()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.itemsMap, testEgressIPAssignmentMap)
		}
	}
}

// buildDummyEgressIP generates EgressIP definition.
func buildDummyEgressIP(name string) *egressipv1.EgressIP {
	return &egressipv1.EgressIP{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func buildDummyEgressIPBuilder(
	apiClient *clients.Settings) *EgressIPBuilder {
	return NewEgressIPBuilder(apiClient, defaultEgressIPName)
}

func buildInvalidDummyEgressIPBuilder(
	apiClient *clients.Settings) *EgressIPBuilder {
	return NewEgressIPBuilder(apiClient, "")
}

// buildTestClientWithDummyPod returns a client with a dummy Pod.
func buildTestClientWithDummyEgressIP() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyEgressIP(defaultEgressIPName),
		},
		SchemeAttachers: testSchemes,
	})
}

func buildTestClientWithCustomizedEgressIP(name string) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyEgressIP(name),
		},
		SchemeAttachers: testSchemes,
	})
}
