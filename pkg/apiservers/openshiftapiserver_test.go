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

func TestPullOpenshiftAPIServer(t *testing.T) {
	generateOpenShiftAPIServer := func() *operatorv1.OpenShiftAPIServer {
		return &operatorv1.OpenShiftAPIServer{
			ObjectMeta: metav1.ObjectMeta{
				Name: openshiftAPIServerObjName,
			},
			Spec: operatorv1.OpenShiftAPIServerSpec{},
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
			expectedError:       fmt.Errorf("openshiftApiServer 'apiClient' cannot be empty"),
			client:              false,
		},
		{
			addToRuntimeObjects: false,
			expectedError:       fmt.Errorf("openshiftAPIServer object cluster doesn't found"),
			client:              true,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		openShiftAPIServer := generateOpenShiftAPIServer()

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, openShiftAPIServer)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{K8sMockObjects: runtimeObjects})
		}

		builderResult, err := PullOpenshiftAPIServer(testSettings)

		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, "cluster", builderResult.Object.Name)
		}
	}
}

func TestOpenshiftAPIServerGet(t *testing.T) {
	testCases := []struct {
		testOpenshiftAPIServerBuilder *OpenshiftAPIServerBuilder
		expectedError                 error
	}{
		{
			testOpenshiftAPIServerBuilder: buildValidOpenshiftAPIServerBuilder(buildOpenshiftAPIServerBuilderWithDummyObject()),
			expectedError:                 nil,
		},
		{
			testOpenshiftAPIServerBuilder: buildValidOpenshiftAPIServerBuilder(
				clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("openshiftapiservers.operator.openshift.io \"cluster\" not found"),
		},
	}

	for _, testCase := range testCases {
		testOpenshiftAPIServer, err := testCase.testOpenshiftAPIServerBuilder.Get()

		if testCase.expectedError == nil {
			assert.Equal(t, testOpenshiftAPIServer.Name, testCase.testOpenshiftAPIServerBuilder.Definition.Name)
		} else {
			assert.Equal(t, err.Error(), testCase.expectedError.Error())
		}
	}
}

func TestOpenshiftAPIServerExists(t *testing.T) {
	testCases := []struct {
		testOpenshiftAPIServerBuilder *OpenshiftAPIServerBuilder
		expectedStatus                bool
	}{
		{
			testOpenshiftAPIServerBuilder: buildValidOpenshiftAPIServerBuilder(
				buildOpenshiftAPIServerBuilderWithDummyObject()),
			expectedStatus: true,
		},
		{
			testOpenshiftAPIServerBuilder: buildValidOpenshiftAPIServerBuilder(
				clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		status := testCase.testOpenshiftAPIServerBuilder.Exists()
		assert.Equal(t, status, testCase.expectedStatus)
	}
}

func TestOpenshiftAPIServerGetCondition(t *testing.T) {
	testCases := []struct {
		testOpenshiftAPIServerBuilder *OpenshiftAPIServerBuilder
		condition                     string
		conditionStatus               operatorv1.ConditionStatus
		expectedError                 error
	}{
		{
			condition:                     "APIServerDeploymentProgressing",
			conditionStatus:               "True",
			testOpenshiftAPIServerBuilder: buildValidOpenshiftAPIServerBuilder(buildOpenshiftAPIServerBuilderWithDummyObject()),
			expectedError:                 nil,
		},
		{
			condition:                     "Unavailable",
			conditionStatus:               "",
			testOpenshiftAPIServerBuilder: buildValidOpenshiftAPIServerBuilder(buildOpenshiftAPIServerBuilderWithDummyObject()),
			expectedError:                 fmt.Errorf("the cluster openshiftAPIServer Unavailable condition not found"),
		},
		{
			condition:                     "",
			conditionStatus:               "",
			testOpenshiftAPIServerBuilder: buildValidOpenshiftAPIServerBuilder(buildOpenshiftAPIServerBuilderWithDummyObject()),
			expectedError:                 fmt.Errorf("openshiftAPIServer 'conditionType' cannot be empty"),
		},
		{
			condition:       "APIServerDeploymentProgressing",
			conditionStatus: "True",
			testOpenshiftAPIServerBuilder: buildValidOpenshiftAPIServerBuilder(
				clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("cluster openshiftAPIServer not found"),
		},
	}

	for _, testCase := range testCases {
		status, msg, err := testCase.testOpenshiftAPIServerBuilder.GetCondition(testCase.condition)
		assert.Equal(t, testCase.expectedError, err)

		if err == nil {
			assert.Equal(t, *status, testCase.conditionStatus)
		} else {
			assert.Nil(t, status)
			assert.Equal(t, "", msg)
		}
	}
}

func TestOpenshiftAPIServerWaitUntilConditionTrue(t *testing.T) {
	testCases := []struct {
		testOpenshiftAPIServerBuilder *OpenshiftAPIServerBuilder
		condition                     string
		expectedError                 error
	}{
		{
			condition: "APIServerDeploymentProgressing",
			testOpenshiftAPIServerBuilder: buildValidOpenshiftAPIServerBuilder(
				buildOpenshiftAPIServerBuilderWithDummyObject()),
			expectedError: nil,
		},
		{
			condition: "Unavailable",
			testOpenshiftAPIServerBuilder: buildValidOpenshiftAPIServerBuilder(
				buildOpenshiftAPIServerBuilderWithDummyObject()),
			expectedError: fmt.Errorf("the Unavailable condition not found exists: context deadline exceeded"),
		},
		{
			condition: "",
			testOpenshiftAPIServerBuilder: buildValidOpenshiftAPIServerBuilder(
				buildOpenshiftAPIServerBuilderWithDummyObject()),
			expectedError: fmt.Errorf("openshiftAPIServer 'conditionType' cannot be empty"),
		},
		{
			condition: "APIServerDeploymentProgressing",
			testOpenshiftAPIServerBuilder: buildValidOpenshiftAPIServerBuilder(
				clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("cluster openshiftAPIServer not found"),
		},
	}

	for _, testCase := range testCases {
		err := testCase.testOpenshiftAPIServerBuilder.WaitUntilConditionTrue(testCase.condition, 1*time.Second)
		if err != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestOpenshiftAPIServerWaitAllPodsAtTheLatestGeneration(t *testing.T) {
	testCases := []struct {
		testOpenshiftAPIServerBuilder *OpenshiftAPIServerBuilder
		condition                     string
		conditionStatus               operatorv1.ConditionStatus
		expectedError                 error
	}{
		{
			condition:                     "APIServerDeploymentProgressing",
			conditionStatus:               "True",
			testOpenshiftAPIServerBuilder: buildValidOpenshiftAPIServerBuilder(buildOpenshiftAPIServerBuilderWithDummyObject()),
			expectedError:                 nil,
		},
		{
			condition:                     "Unavailable",
			conditionStatus:               "",
			testOpenshiftAPIServerBuilder: buildValidOpenshiftAPIServerBuilder(buildOpenshiftAPIServerBuilderWithDummyObject()),
			expectedError:                 fmt.Errorf("the Unavailable condition not found exists: context deadline exceeded"),
		},
		{
			condition:       "",
			conditionStatus: "",
			testOpenshiftAPIServerBuilder: buildValidOpenshiftAPIServerBuilder(
				buildOpenshiftAPIServerBuilderWithDummyObject()),
			expectedError: fmt.Errorf("openshiftAPIServer 'conditionType' cannot be empty"),
		},
		{
			condition:       "NodeInstallerProgressing",
			conditionStatus: "True",
			testOpenshiftAPIServerBuilder: buildValidOpenshiftAPIServerBuilder(
				clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("cluster openshiftAPIServer not found"),
		},
	}

	for _, testCase := range testCases {
		err := testCase.testOpenshiftAPIServerBuilder.WaitAllPodsAtTheLatestGeneration(1 * time.Second)
		if err != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func buildValidOpenshiftAPIServerBuilder(apiClient *clients.Settings) *OpenshiftAPIServerBuilder {
	return &OpenshiftAPIServerBuilder{
		apiClient: apiClient.Client,
		Definition: &operatorv1.OpenShiftAPIServer{
			ObjectMeta: metav1.ObjectMeta{
				Name: openshiftAPIServerObjName,
			},
			Spec: operatorv1.OpenShiftAPIServerSpec{},
			Status: operatorv1.OpenShiftAPIServerStatus{
				OperatorStatus: operatorv1.OperatorStatus{
					Conditions: []operatorv1.OperatorCondition{
						{Type: "APIServerDeploymentProgressing", Status: "True", Reason: "AsExpected"}},
				},
			},
		},
	}
}

func buildOpenshiftAPIServerBuilderWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: buildDummyOpenshiftAPIServerBuilderObject()})
}

func buildDummyOpenshiftAPIServerBuilderObject() []runtime.Object {
	return append([]runtime.Object{}, &operatorv1.OpenShiftAPIServer{
		ObjectMeta: metav1.ObjectMeta{
			Name: openshiftAPIServerObjName,
		},
		Spec: operatorv1.OpenShiftAPIServerSpec{},
		Status: operatorv1.OpenShiftAPIServerStatus{
			OperatorStatus: operatorv1.OperatorStatus{
				Conditions: []operatorv1.OperatorCondition{
					{Type: "APIServerDeploymentProgressing", Status: "True", Reason: "AsExpected"}},
			},
		},
	})
}
