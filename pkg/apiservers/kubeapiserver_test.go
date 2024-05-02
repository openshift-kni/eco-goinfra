package apiservers

import (
	"fmt"
	"testing"
	"time"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	operatorv1 "github.com/openshift/api/operator/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestPullKubeAPIServer(t *testing.T) {
	generateKubeAPIServer := func() *operatorv1.KubeAPIServer {
		return &operatorv1.KubeAPIServer{
			ObjectMeta: metav1.ObjectMeta{
				Name: kubeAPIServerObjName,
			},
			Spec: operatorv1.KubeAPIServerSpec{},
		}
	}

	testCases := []struct {
		addToRuntimeObjects bool
		expectedError       error
		client              bool
	}{
		{
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("kubeApiServer 'apiClient' cannot be empty"),
			client:              false,
		},
		{
			addToRuntimeObjects: false,
			expectedError:       fmt.Errorf("kubeAPIServer object cluster doesn't found"),
			client:              true,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		kubeAPIServer := generateKubeAPIServer()

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, kubeAPIServer)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{K8sMockObjects: runtimeObjects})
		}

		builderResult, err := PullKubeAPIServerBuilder(testSettings)

		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, "cluster", builderResult.Object.Name)
		}
	}
}

func TestKubeAPIServerGet(t *testing.T) {
	testCases := []struct {
		testKubeAPIServerBuilder *KubeAPIServerBuilder
		expectedError            error
	}{
		{
			testKubeAPIServerBuilder: buildValidKubeAPIServerBuilder(buildKubeAPIServerWithDummyObject()),
			expectedError:            nil,
		},
		{
			testKubeAPIServerBuilder: buildValidKubeAPIServerBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:            fmt.Errorf("kubeapiservers.operator.openshift.io \"cluster\" not found"),
		},
	}

	for _, testCase := range testCases {
		testKubeAPIServer, err := testCase.testKubeAPIServerBuilder.Get()

		if testCase.expectedError == nil {
			assert.Equal(t, testKubeAPIServer.Name, testCase.testKubeAPIServerBuilder.Definition.Name)
		} else {
			assert.Equal(t, err.Error(), testCase.expectedError.Error())
		}
	}
}

func TestKubeAPIServerExist(t *testing.T) {
	testCases := []struct {
		testKubeAPIServerBuilder *KubeAPIServerBuilder
		expectedStatus           bool
	}{
		{
			testKubeAPIServerBuilder: buildValidKubeAPIServerBuilder(buildKubeAPIServerWithDummyObject()),
			expectedStatus:           true,
		},
		{
			testKubeAPIServerBuilder: buildValidKubeAPIServerBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus:           false,
		},
	}

	for _, testCase := range testCases {
		status := testCase.testKubeAPIServerBuilder.Exists()
		assert.Equal(t, status, testCase.expectedStatus)
	}
}

func TestKubeAPIServerGetCondition(t *testing.T) {
	testCases := []struct {
		testKubeAPIServerBuilder *KubeAPIServerBuilder
		condition                string
		conditionStatus          operatorv1.ConditionStatus
		expectedError            error
	}{
		{
			condition:                "NodeInstallerProgressing",
			conditionStatus:          "True",
			testKubeAPIServerBuilder: buildValidKubeAPIServerBuilder(buildKubeAPIServerWithDummyObject()),
			expectedError:            nil,
		},
		{
			condition:                "Unavailable",
			conditionStatus:          "",
			testKubeAPIServerBuilder: buildValidKubeAPIServerBuilder(buildKubeAPIServerWithDummyObject()),
			expectedError:            fmt.Errorf("the cluster kubeAPIServer Unavailable condition not found"),
		},
		{
			condition:                "",
			conditionStatus:          "",
			testKubeAPIServerBuilder: buildValidKubeAPIServerBuilder(buildKubeAPIServerWithDummyObject()),
			expectedError:            fmt.Errorf("kubeAPIServer 'conditionType' cannot be empty"),
		},
		{
			condition:                "NodeInstallerProgressing",
			conditionStatus:          "True",
			testKubeAPIServerBuilder: buildValidKubeAPIServerBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:            fmt.Errorf("cluster kubeAPIServer not found"),
		},
	}

	for _, testCase := range testCases {
		status, msg, err := testCase.testKubeAPIServerBuilder.GetCondition(testCase.condition)
		assert.Equal(t, testCase.expectedError, err)

		if err == nil {
			assert.Equal(t, *status, testCase.conditionStatus)
		} else {
			assert.Nil(t, status)
			assert.Equal(t, "", msg)
		}
	}
}

func TestKubeAPIServerWaitUntilConditionTrue(t *testing.T) {
	testCases := []struct {
		testKubeAPIServerBuilder *KubeAPIServerBuilder
		condition                string
		expectedError            error
	}{
		{
			condition:                "NodeInstallerProgressing",
			testKubeAPIServerBuilder: buildValidKubeAPIServerBuilder(buildKubeAPIServerWithDummyObject()),
			expectedError:            nil,
		},
		{
			condition:                "unavailable",
			testKubeAPIServerBuilder: buildValidKubeAPIServerBuilder(buildKubeAPIServerWithDummyObject()),
			expectedError:            fmt.Errorf("the unavailable condition not found exists: context deadline exceeded"),
		},
		{
			condition:                "",
			testKubeAPIServerBuilder: buildValidKubeAPIServerBuilder(buildKubeAPIServerWithDummyObject()),
			expectedError:            fmt.Errorf("kubeAPIServer 'conditionType' cannot be empty"),
		},
		{
			condition:                "NodeInstallerProgressing",
			testKubeAPIServerBuilder: buildValidKubeAPIServerBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:            fmt.Errorf("cluster kubeAPIServer not found"),
		},
	}

	for _, testCase := range testCases {
		err := testCase.testKubeAPIServerBuilder.WaitUntilConditionTrue(testCase.condition, 1*time.Second)
		if err != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestKubeAPIServerWaitAllNodesAtTheLatestRevision(t *testing.T) {
	testCases := []struct {
		testKubeAPIServerBuilder *KubeAPIServerBuilder
		expectedError            error
	}{
		{
			testKubeAPIServerBuilder: buildValidKubeAPIServerBuilder(buildKubeAPIServerWithDummyObject()),
			expectedError:            nil,
		},
		{
			testKubeAPIServerBuilder: buildValidKubeAPIServerBuilder(buildKubeAPIServerWithDummyObject()),
			expectedError:            fmt.Errorf("the unavailable condition not found exists: context deadline exceeded"),
		},
		{
			testKubeAPIServerBuilder: buildValidKubeAPIServerBuilder(buildKubeAPIServerWithDummyObject()),
			expectedError:            fmt.Errorf("kubeAPIServer 'conditionType' cannot be empty"),
		},
		{
			testKubeAPIServerBuilder: buildValidKubeAPIServerBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:            fmt.Errorf("cluster kubeAPIServer not found"),
		},
	}

	for _, testCase := range testCases {
		err := testCase.testKubeAPIServerBuilder.WaitAllNodesAtTheLatestRevision(1 * time.Second)
		if err != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func buildValidKubeAPIServerBuilder(apiClient *clients.Settings) *KubeAPIServerBuilder {
	return &KubeAPIServerBuilder{
		apiClient: apiClient.Client,
		Definition: &operatorv1.KubeAPIServer{
			ObjectMeta: metav1.ObjectMeta{
				Name:            kubeAPIServerObjName,
				ResourceVersion: "999",
			},
			Spec: operatorv1.KubeAPIServerSpec{},
			Status: operatorv1.KubeAPIServerStatus{
				StaticPodOperatorStatus: operatorv1.StaticPodOperatorStatus{
					OperatorStatus: operatorv1.OperatorStatus{
						Conditions: []operatorv1.OperatorCondition{
							{Type: "NodeInstallerProgressing", Status: "True", Reason: "AllNodesAtLatestRevision"}},
					},
				},
			},
		},
	}
}

func buildKubeAPIServerWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: buildDummyKubeAPIServerObject()})
}

func buildDummyKubeAPIServerObject() []runtime.Object {
	return append([]runtime.Object{}, &operatorv1.KubeAPIServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:            kubeAPIServerObjName,
			ResourceVersion: "999",
		},
		Spec: operatorv1.KubeAPIServerSpec{},
		Status: operatorv1.KubeAPIServerStatus{
			StaticPodOperatorStatus: operatorv1.StaticPodOperatorStatus{
				OperatorStatus: operatorv1.OperatorStatus{
					Conditions: []operatorv1.OperatorCondition{
						{Type: "NodeInstallerProgressing", Status: "True", Reason: "AllNodesAtLatestRevision"}},
				},
			},
		},
	})
}
