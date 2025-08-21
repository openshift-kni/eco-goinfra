package keda

import (
	"fmt"
	"testing"

	kedav2v1alpha1 "github.com/kedacore/keda/v2/apis/keda/v1alpha1"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultScaledObjectName      = "prometheus-scaledobject"
	defaultScaledObjectNamespace = "test-appspace"
	pollingInterval              = int32(5)
	cooldownPeriod               = int32(10)
	minReplicaCount              = int32(1)
	maxReplicaCount              = int32(8)
	zeroValue                    = int32(0)
	kedav2v1alpha1TestSchemes    = []clients.SchemeAttacher{
		kedav2v1alpha1.AddToScheme,
	}
)

func TestPullScaledObject(t *testing.T) {
	generateScaleObject := func(name, namespace string) *kedav2v1alpha1.ScaledObject {
		return &kedav2v1alpha1.ScaledObject{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
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
			name:                defaultScaledObjectName,
			namespace:           defaultScaledObjectNamespace,
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           defaultScaledObjectNamespace,
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("scaledObject 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                defaultScaledObjectName,
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("scaledObject 'nsname' cannot be empty"),
			client:              true,
		},
		{
			name:                "sotest",
			namespace:           defaultScaledObjectNamespace,
			addToRuntimeObjects: false,
			expectedError: fmt.Errorf("scaledObject object sotest does not exist " +
				"in namespace test-appspace"),
			client: true,
		},
		{
			name:                "sotest",
			namespace:           defaultScaledObjectNamespace,
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("scaledObject 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testTriggerAuth := generateScaleObject(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testTriggerAuth)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: kedav2v1alpha1TestSchemes,
			})
		}

		builderResult, err := PullScaledObject(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		} else {
			assert.Equal(t, testTriggerAuth.Name, builderResult.Object.Name)
			assert.Equal(t, testTriggerAuth.Namespace, builderResult.Object.Namespace)
			assert.Nil(t, err)
		}
	}
}

func TestNewScaledObjectBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		expectedError string
		client        bool
	}{
		{
			name:          defaultScaledObjectName,
			namespace:     defaultScaledObjectNamespace,
			expectedError: "",
			client:        true,
		},
		{
			name:          "",
			namespace:     defaultScaledObjectNamespace,
			expectedError: "scaledObject 'name' cannot be empty",
			client:        true,
		},
		{
			name:          defaultScaledObjectName,
			namespace:     "",
			expectedError: "scaledObject 'nsname' cannot be empty",
			client:        true,
		},
		{
			name:          defaultScaledObjectName,
			namespace:     defaultScaledObjectNamespace,
			expectedError: "",
			client:        false,
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings
		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		testScaledObjectBuilder := NewScaledObjectBuilder(testSettings, testCase.name, testCase.namespace)

		if testCase.expectedError == "" {
			if testCase.client {
				assert.Equal(t, testCase.name, testScaledObjectBuilder.Definition.Name)
				assert.Equal(t, testCase.namespace, testScaledObjectBuilder.Definition.Namespace)
			} else {
				assert.Nil(t, testScaledObjectBuilder)
			}
		} else {
			assert.Equal(t, testCase.expectedError, testScaledObjectBuilder.errorMsg)
			assert.NotNil(t, testScaledObjectBuilder.Definition)
		}
	}
}

func TestScaledObjectExists(t *testing.T) {
	testCases := []struct {
		testScaledObject *ScaledObjectBuilder
		expectedStatus   bool
	}{
		{
			testScaledObject: buildValidScaledObjectBuilder(buildScaledObjectClientWithDummyObject()),
			expectedStatus:   true,
		},
		{
			testScaledObject: buildInValidScaledObjectBuilder(buildScaledObjectClientWithDummyObject()),
			expectedStatus:   false,
		},
		{
			testScaledObject: buildValidScaledObjectBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus:   false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testScaledObject.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestScaledObjectGet(t *testing.T) {
	testCases := []struct {
		testScaledObject *ScaledObjectBuilder
		expectedError    error
	}{
		{
			testScaledObject: buildValidScaledObjectBuilder(buildScaledObjectClientWithDummyObject()),
			expectedError:    nil,
		},
		{
			testScaledObject: buildInValidScaledObjectBuilder(buildScaledObjectClientWithDummyObject()),
			expectedError:    fmt.Errorf("scaledObject 'name' cannot be empty"),
		},
		{
			testScaledObject: buildValidScaledObjectBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:    fmt.Errorf("scaledobjects.keda.sh \"prometheus-scaledobject\" not found"),
		},
	}

	for _, testCase := range testCases {
		scaledObjectObj, err := testCase.testScaledObject.Get()

		if testCase.expectedError == nil {
			assert.Equal(t, scaledObjectObj.Name, testCase.testScaledObject.Definition.Name)
			assert.Equal(t, scaledObjectObj.Namespace, testCase.testScaledObject.Definition.Namespace)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestScaledObjectCreate(t *testing.T) {
	testCases := []struct {
		testScaledObject *ScaledObjectBuilder
		expectedError    string
	}{
		{
			testScaledObject: buildValidScaledObjectBuilder(buildScaledObjectClientWithDummyObject()),
			expectedError:    "",
		},
		{
			testScaledObject: buildInValidScaledObjectBuilder(buildScaledObjectClientWithDummyObject()),
			expectedError:    "scaledObject 'name' cannot be empty",
		},
		{
			testScaledObject: buildValidScaledObjectBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:    "",
		},
	}

	for _, testCase := range testCases {
		testScaledObjectBuilder, err := testCase.testScaledObject.Create()

		if testCase.expectedError == "" {
			assert.Equal(t, testScaledObjectBuilder.Definition.Name, testScaledObjectBuilder.Object.Name)
			assert.Equal(t, testScaledObjectBuilder.Definition.Namespace, testScaledObjectBuilder.Object.Namespace)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError, err.Error())
		}
	}
}

func TestScaledObjectDelete(t *testing.T) {
	testCases := []struct {
		testScaledObject *ScaledObjectBuilder
		expectedError    error
	}{
		{
			testScaledObject: buildValidScaledObjectBuilder(buildScaledObjectClientWithDummyObject()),
			expectedError:    nil,
		},
		{
			testScaledObject: buildInValidScaledObjectBuilder(buildScaledObjectClientWithDummyObject()),
			expectedError:    fmt.Errorf("scaledObject 'name' cannot be empty"),
		},
		{
			testScaledObject: buildValidScaledObjectBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:    nil,
		},
	}

	for _, testCase := range testCases {
		_, err := testCase.testScaledObject.Delete()

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testScaledObject.Object)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestScaledObjectUpdate(t *testing.T) {
	testCases := []struct {
		testScaleObject    *ScaledObjectBuilder
		expectedError      string
		testScaleTargetRef kedav2v1alpha1.ScaleTarget
	}{
		{
			testScaleObject: buildValidScaledObjectBuilder(buildScaledObjectClientWithDummyObject()),
			expectedError:   "",
			testScaleTargetRef: kedav2v1alpha1.ScaleTarget{
				Name: "test-app",
			},
		},
		{
			testScaleObject:    buildInValidScaledObjectBuilder(buildScaledObjectClientWithDummyObject()),
			expectedError:      "scaledObject 'name' cannot be empty",
			testScaleTargetRef: kedav2v1alpha1.ScaleTarget{},
		},
	}

	for _, testCase := range testCases {
		assert.Nil(t, nil, testCase.testScaleObject.Object)
		testCase.testScaleObject.WithScaleTargetRef(testCase.testScaleTargetRef)
		_, err := testCase.testScaleObject.Update()

		if testCase.expectedError != "" {
			assert.Equal(t, testCase.expectedError, err.Error())
		} else {
			assert.Equal(t, testCase.testScaleTargetRef.Name,
				testCase.testScaleObject.Definition.Spec.ScaleTargetRef.Name)
		}
	}
}

func TestScaleObjectWithScaleTargetRef(t *testing.T) {
	testCases := []struct {
		testScaleTargetRef kedav2v1alpha1.ScaleTarget
		expectedError      bool
		expectedErrorText  string
	}{
		{
			testScaleTargetRef: kedav2v1alpha1.ScaleTarget{
				Name: "test-app",
			},
			expectedError:     false,
			expectedErrorText: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidScaledObjectBuilder(buildScaledObjectClientWithDummyObject())

		result := testBuilder.WithScaleTargetRef(testCase.testScaleTargetRef)

		if testCase.expectedError {
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
			}
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testScaleTargetRef, *result.Definition.Spec.ScaleTargetRef)
		}
	}
}

func TestScaleObjectWithPollingInterval(t *testing.T) {
	testCases := []struct {
		testPollingInterval int32
		expectedError       bool
		expectedErrorText   string
	}{
		{
			testPollingInterval: pollingInterval,
			expectedError:       false,
			expectedErrorText:   "",
		},
		{
			testPollingInterval: zeroValue,
			expectedError:       false,
			expectedErrorText:   "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidScaledObjectBuilder(buildScaledObjectClientWithDummyObject())

		result := testBuilder.WithPollingInterval(testCase.testPollingInterval)

		if testCase.expectedError {
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
			}
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testPollingInterval, *result.Definition.Spec.PollingInterval)
		}
	}
}

func TestScaleObjectWithCooldownPeriod(t *testing.T) {
	testCases := []struct {
		testCooldownPeriod int32
		expectedError      bool
		expectedErrorText  string
	}{
		{
			testCooldownPeriod: cooldownPeriod,
			expectedError:      false,
			expectedErrorText:  "",
		},
		{
			testCooldownPeriod: zeroValue,
			expectedError:      false,
			expectedErrorText:  "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidScaledObjectBuilder(buildScaledObjectClientWithDummyObject())

		result := testBuilder.WithCooldownPeriod(testCase.testCooldownPeriod)

		if testCase.expectedError {
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
			}
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testCooldownPeriod, *result.Definition.Spec.CooldownPeriod)
		}
	}
}

func TestScaleObjectWithMinReplicaCount(t *testing.T) {
	testCases := []struct {
		testMinReplicaCount int32
		expectedErrorText   string
	}{
		{
			testMinReplicaCount: minReplicaCount,
			expectedErrorText:   "",
		},
		{
			testMinReplicaCount: zeroValue,
			expectedErrorText:   "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidScaledObjectBuilder(buildScaledObjectClientWithDummyObject())

		result := testBuilder.WithMinReplicaCount(testCase.testMinReplicaCount)

		if testCase.expectedErrorText != "" {
			assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testMinReplicaCount, *result.Definition.Spec.MinReplicaCount)
		}
	}
}

func TestScaleObjectWithMaxReplicaCount(t *testing.T) {
	testCases := []struct {
		testMaxReplicaCount int32
		expectedErrorText   string
	}{
		{
			testMaxReplicaCount: maxReplicaCount,
			expectedErrorText:   "",
		},
		{
			testMaxReplicaCount: zeroValue,
			expectedErrorText:   "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidScaledObjectBuilder(buildScaledObjectClientWithDummyObject())

		result := testBuilder.WithMaxReplicaCount(testCase.testMaxReplicaCount)

		if testCase.expectedErrorText != "" {
			assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testMaxReplicaCount, *result.Definition.Spec.MaxReplicaCount)
		}
	}
}

func TestScaleObjectWithTriggers(t *testing.T) {
	testCases := []struct {
		testTriggers      []kedav2v1alpha1.ScaleTriggers
		expectedError     bool
		expectedErrorText string
	}{
		{
			testTriggers: []kedav2v1alpha1.ScaleTriggers{{
				Type: "prometheus",
				Metadata: map[string]string{
					"serverAddress": "https://thanos-querier.openshift-monitoring.svc.cluster.local:9092",
					"namespace":     defaultScaledObjectNamespace,
					"metricName":    "http_requests_total",
					"threshold":     "5",
					"query":         "sum(rate(http_requests_total{job=\"test-app\"}[1m]))",
					"authModes":     "bearer",
				},
				AuthenticationRef: &kedav2v1alpha1.AuthenticationRef{
					Name: defaultScaledObjectName,
				},
			}},
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			testTriggers: []kedav2v1alpha1.ScaleTriggers{{
				Type: "prometheus",
				Metadata: map[string]string{
					"serverAddress": "https://thanos-querier.openshift-monitoring.svc.cluster.local:9092",
					"namespace":     defaultScaledObjectNamespace,
					"metricName":    "http_requests_total",
					"threshold":     "5",
					"query":         "sum(rate(http_requests_total{job=\"test-app\"}[1m]))",
					"authModes":     "bearer",
				},
			}},
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			testTriggers: []kedav2v1alpha1.ScaleTriggers{{
				Type: "prometheus",
				AuthenticationRef: &kedav2v1alpha1.AuthenticationRef{
					Name: defaultScaledObjectName,
				},
			}},
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			testTriggers: []kedav2v1alpha1.ScaleTriggers{{
				Type: "prometheus",
			}},
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			testTriggers:      []kedav2v1alpha1.ScaleTriggers{},
			expectedError:     true,
			expectedErrorText: "'triggers' argument cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidScaledObjectBuilder(buildScaledObjectClientWithDummyObject())

		result := testBuilder.WithTriggers(testCase.testTriggers)

		if testCase.expectedError {
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
			}
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testTriggers, result.Definition.Spec.Triggers)
		}
	}
}

func buildValidScaledObjectBuilder(apiClient *clients.Settings) *ScaledObjectBuilder {
	scaleObjectBuilder := NewScaledObjectBuilder(
		apiClient, defaultScaledObjectName, defaultScaledObjectNamespace)

	return scaleObjectBuilder
}

func buildInValidScaledObjectBuilder(apiClient *clients.Settings) *ScaledObjectBuilder {
	scaleObjectBuilder := NewScaledObjectBuilder(
		apiClient, "", defaultScaledObjectNamespace)

	return scaleObjectBuilder
}

func buildScaledObjectClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyScaledObject(),
		SchemeAttachers: kedav2v1alpha1TestSchemes,
	})
}

func buildDummyScaledObject() []runtime.Object {
	return append([]runtime.Object{}, &kedav2v1alpha1.ScaledObject{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultScaledObjectName,
			Namespace: defaultScaledObjectNamespace,
		},
	})
}
