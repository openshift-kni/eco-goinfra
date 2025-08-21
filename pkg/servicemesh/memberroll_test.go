package servicemesh

import (
	"fmt"
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	istiov1 "maistra.io/api/core/v1"
)

var (
	defaultMemberRollName       = "default"
	defaultServiceMeshNamespace = "istio-system"
	defaultMembersList          = []string(nil)
	newMembersList              = []string{"test-ns1", "test-ns2"}
	readyCondition              = istiov1.ServiceMeshMemberRollCondition{
		Type:               "Ready",
		Status:             "True",
		LastTransitionTime: metav1.Time{Time: time.Now()},
		Reason:             "Configured",
		Message:            "All namespaces have been configured successfully",
	}
	notReadyCondition = istiov1.ServiceMeshMemberRollCondition{
		Type:               "Ready",
		Status:             "False",
		LastTransitionTime: metav1.Time{Time: time.Now()},
		Reason:             "ErrSMCPMissing",
		Message:            "No ServiceMeshControlPlane exists in the namespace",
	}
	istiov1TestSchemes = []clients.SchemeAttacher{
		istiov1.AddToScheme,
	}
)

func TestPullMemberRoll(t *testing.T) {
	generateMemberRoll := func(name, namespace string) *istiov1.ServiceMeshMemberRoll {
		return &istiov1.ServiceMeshMemberRoll{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: istiov1.ServiceMeshMemberRollSpec{
				Members: defaultMembersList,
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
			name:                "test",
			namespace:           "istio-system",
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           "istio-system",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("serviceMeshMemberRoll 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                "test",
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("serviceMeshMemberRoll 'nsname' cannot be empty"),
			client:              true,
		},
		{
			name:                "smomrtest",
			namespace:           "istio-system",
			addToRuntimeObjects: false,
			expectedError:       fmt.Errorf("serviceMeshMemberRoll object smomrtest does not exist in namespace istio-system"),
			client:              true,
		},
		{
			name:                "smomrtest",
			namespace:           "istio-system",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("serviceMeshMemberRoll 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testMemberRoll := generateMemberRoll(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testMemberRoll)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: istiov1TestSchemes,
			})
		}

		builderResult, err := PullMemberRoll(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		} else {
			assert.Equal(t, testMemberRoll.Name, builderResult.Object.Name)
			assert.Equal(t, testMemberRoll.Namespace, builderResult.Object.Namespace)
		}
	}
}

func TestNewMemberRollBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		expectedError string
	}{
		{
			name:          defaultMemberRollName,
			namespace:     defaultServiceMeshNamespace,
			expectedError: "",
		},
		{
			name:          "",
			namespace:     defaultServiceMeshNamespace,
			expectedError: "serviceMeshMemberRoll 'name' cannot be empty",
		},
		{
			name:          defaultMemberRollName,
			namespace:     "",
			expectedError: "serviceMeshMemberRoll 'nsname' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})
		testMemberRollBuilder := NewMemberRollBuilder(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, testMemberRollBuilder.errorMsg)
		assert.NotNil(t, testMemberRollBuilder.Definition)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.name, testMemberRollBuilder.Definition.Name)
			assert.Equal(t, testCase.namespace, testMemberRollBuilder.Definition.Namespace)
		}
	}
}

func TestMemberRollExists(t *testing.T) {
	testCases := []struct {
		testMemberRoll *MemberRollBuilder
		expectedStatus bool
	}{
		{
			testMemberRoll: buildValidMemberRollBuilder(buildMemberRollClientWithDummyObject()),
			expectedStatus: true,
		},
		{
			testMemberRoll: buildInValidMemberRollBuilder(buildMemberRollClientWithDummyObject()),
			expectedStatus: false,
		},
		{
			testMemberRoll: buildValidMemberRollBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testMemberRoll.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestMemberRollGet(t *testing.T) {
	testCases := []struct {
		testMemberRoll *MemberRollBuilder
		expectedError  error
	}{
		{
			testMemberRoll: buildValidMemberRollBuilder(buildMemberRollClientWithDummyObject()),
			expectedError:  nil,
		},
		{
			testMemberRoll: buildInValidMemberRollBuilder(buildMemberRollClientWithDummyObject()),
			expectedError:  fmt.Errorf("serviceMeshMemberRoll 'name' cannot be empty"),
		},
		{
			testMemberRoll: buildValidMemberRollBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:  fmt.Errorf("servicemeshmemberrolls.maistra.io \"default\" not found"),
		},
	}

	for _, testCase := range testCases {
		memberRollObj, err := testCase.testMemberRoll.Get()

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.testMemberRoll.Definition, memberRollObj)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestMemberRollCreate(t *testing.T) {
	testCases := []struct {
		testMemberRoll *MemberRollBuilder
		expectedError  string
	}{
		{
			testMemberRoll: buildValidMemberRollBuilder(buildMemberRollClientWithDummyObject()),
			expectedError:  "",
		},
		{
			testMemberRoll: buildInValidMemberRollBuilder(buildMemberRollClientWithDummyObject()),
			expectedError:  "serviceMeshMemberRoll 'name' cannot be empty",
		},
		{
			testMemberRoll: buildValidMemberRollBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:  "resourceVersion can not be set for Create requests",
		},
	}

	for _, testCase := range testCases {
		testMemberRollBuilder, err := testCase.testMemberRoll.Create()

		if testCase.expectedError == "" {
			assert.Equal(t, testMemberRollBuilder.Definition, testMemberRollBuilder.Object)
		} else {
			assert.Equal(t, testCase.expectedError, err.Error())
		}
	}
}

func TestMemberRollDelete(t *testing.T) {
	testCases := []struct {
		testMemberRoll *MemberRollBuilder
		expectedError  error
	}{
		{
			testMemberRoll: buildValidMemberRollBuilder(buildMemberRollClientWithDummyObject()),
			expectedError:  nil,
		},
		{
			testMemberRoll: buildInValidMemberRollBuilder(buildMemberRollClientWithDummyObject()),
			expectedError:  fmt.Errorf("serviceMeshMemberRoll 'name' cannot be empty"),
		},
		{
			testMemberRoll: buildValidMemberRollBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:  nil,
		},
	}

	for _, testCase := range testCases {
		err := testCase.testMemberRoll.Delete()

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testMemberRoll.Object)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestMemberRollUpdate(t *testing.T) {
	testCases := []struct {
		testMemberRoll *MemberRollBuilder
		expectedError  string
		members        []string
	}{
		{
			testMemberRoll: buildValidMemberRollBuilder(buildMemberRollClientWithDummyObject()),
			expectedError:  "",
			members:        newMembersList,
		},
		{
			testMemberRoll: buildInValidMemberRollBuilder(buildMemberRollClientWithDummyObject()),
			expectedError:  "serviceMeshMemberRoll 'name' cannot be empty",
			members:        newMembersList,
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, defaultMembersList, testCase.testMemberRoll.Definition.Spec.Members)
		assert.Nil(t, nil, testCase.testMemberRoll.Object)
		testCase.testMemberRoll.WithMembersList(testCase.members)
		_, err := testCase.testMemberRoll.Update(true)

		if testCase.expectedError != "" {
			assert.Equal(t, testCase.expectedError, err.Error())
		} else {
			assert.Equal(t, testCase.members,
				testCase.testMemberRoll.Definition.Spec.Members)
		}
	}
}

func TestMemberRollWithMembersList(t *testing.T) {
	testCases := []struct {
		testMembers       []string
		expectedError     bool
		expectedErrorText string
	}{
		{
			testMembers:       newMembersList,
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			testMembers:       nil,
			expectedError:     true,
			expectedErrorText: "can not modify memberRoll config with empty membersList",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidMemberRollBuilder(buildMemberRollClientWithDummyObject())

		result := testBuilder.WithMembersList(testCase.testMembers)

		if testCase.expectedError {
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
			}
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testMembers, result.Definition.Spec.Members)
		}
	}
}

func TestMemberRollGetMembersList(t *testing.T) {
	testCases := []struct {
		testMemberRoll *MemberRollBuilder
		expectedError  error
	}{
		{
			testMemberRoll: buildValidMemberRollBuilder(buildMemberRollClientWithDummyObject()),
			expectedError:  nil,
		},
		{
			testMemberRoll: buildValidMemberRollBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:  fmt.Errorf("memberRoll object default does not exist in namespace istio-system"),
		},
	}

	for _, testCase := range testCases {
		currentMemberRollMembersList, err := testCase.testMemberRoll.GetMembersList()

		if testCase.expectedError == nil {
			assert.Equal(t, *currentMemberRollMembersList, testCase.testMemberRoll.Object.Spec.Members)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestMemberRollIsReady(t *testing.T) {
	testCases := []struct {
		testMemberRoll *MemberRollBuilder
		testCondition  bool
		expectedError  error
	}{
		{
			testMemberRoll: buildValidMemberRollBuilderWithCondition(buildMemberRollClientWithDummyObject(),
				readyCondition),
			expectedError: nil,
		},
		{
			testMemberRoll: buildValidMemberRollBuilderWithCondition(buildMemberRollClientWithDummyObject(),
				notReadyCondition),
			expectedError: fmt.Errorf("the Ready condition did not reached for the Service Mesh MemberRoll " +
				"default in namespace istio-system during 2s; context deadline exceeded"),
		},
		{
			testMemberRoll: buildValidMemberRollBuilderWithCondition(clients.GetTestClients(clients.TestClientParams{}),
				readyCondition),
			expectedError: fmt.Errorf("the Ready condition did not reached for the Service Mesh MemberRoll " +
				"default in namespace istio-system during 2s; context deadline exceeded"),
		},
	}

	for _, testCase := range testCases {
		isReadyResult, err := testCase.testMemberRoll.IsReady(2 * time.Second)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.testCondition, isReadyResult)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func buildValidMemberRollBuilderWithCondition(apiClient *clients.Settings,
	condition istiov1.ServiceMeshMemberRollCondition) *MemberRollBuilder {
	memberRollBuilder := buildValidMemberRollBuilder(apiClient)
	memberRollBuilder.Definition.Status.Conditions = []istiov1.ServiceMeshMemberRollCondition{condition}

	return memberRollBuilder
}

func buildValidMemberRollBuilder(apiClient *clients.Settings) *MemberRollBuilder {
	memberRollBuilder := NewMemberRollBuilder(
		apiClient, defaultMemberRollName, defaultServiceMeshNamespace)
	memberRollBuilder.Definition.ResourceVersion = "999"

	return memberRollBuilder
}

func buildInValidMemberRollBuilder(apiClient *clients.Settings) *MemberRollBuilder {
	memberRollBuilder := NewMemberRollBuilder(
		apiClient, "", defaultServiceMeshNamespace)
	memberRollBuilder.Definition.ResourceVersion = "999"

	return memberRollBuilder
}

func buildMemberRollClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyMemberRoll(),
		SchemeAttachers: istiov1TestSchemes,
	})
}

func buildDummyMemberRoll() []runtime.Object {
	return append([]runtime.Object{}, &istiov1.ServiceMeshMemberRoll{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultMemberRollName,
			Namespace: defaultServiceMeshNamespace,
		},
		Spec: istiov1.ServiceMeshMemberRollSpec{
			Members: []string{},
		},
	})
}
