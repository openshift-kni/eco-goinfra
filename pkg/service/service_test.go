package service

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	defaultServiceName      = "test-service-name"
	defaultServiceNamespace = "test-service-namespace"
	defaultServiceSelector  = map[string]string{"testLabel": "testLabelValue"}
	defaultServicePort      = corev1.ServicePort{
		Name:     "http",
		Protocol: "TCP",
		Port:     80,
		TargetPort: intstr.IntOrString{
			Type:   intstr.Int,
			IntVal: int32(8080),
		},
	}
	defaultServiceAnnotation = map[string]string{"service-test/annotation": "true"}
)

func TestPullService(t *testing.T) {
	generateServiceObject := func(name, namespace string) *corev1.Service {
		return &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: corev1.ServiceSpec{
				Selector: defaultServiceSelector,
				Ports:    []corev1.ServicePort{defaultServicePort},
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
			name:                defaultServiceName,
			namespace:           defaultServiceNamespace,
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           defaultServiceNamespace,
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("service 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                defaultServiceName,
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("service 'namespace' cannot be empty"),
			client:              true,
		},
		{
			name:                "servicetesttest",
			namespace:           defaultServiceNamespace,
			addToRuntimeObjects: false,
			expectedError: fmt.Errorf("service object servicetesttest does not exist " +
				"in namespace test-service-namespace"),
			client: true,
		},
		{
			name:                "servicetesttest",
			namespace:           defaultServiceNamespace,
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("service 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testService := generateServiceObject(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testService)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})
		}

		builderResult, err := Pull(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		} else {
			assert.Equal(t, testService.Name, builderResult.Object.Name)
			assert.Equal(t, testService.Namespace, builderResult.Object.Namespace)
			assert.Nil(t, err)
		}
	}
}

func TestNewServiceBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		selectors     map[string]string
		servicePort   corev1.ServicePort
		expectedError string
		apiClient     bool
	}{
		{
			name:          defaultServiceName,
			namespace:     defaultServiceNamespace,
			selectors:     defaultServiceSelector,
			servicePort:   defaultServicePort,
			expectedError: "",
			apiClient:     true,
		},
		{
			name:          "",
			namespace:     defaultServiceNamespace,
			selectors:     defaultServiceSelector,
			servicePort:   defaultServicePort,
			expectedError: "Service 'name' cannot be empty",
			apiClient:     true,
		},
		{
			name:          defaultServiceName,
			namespace:     "",
			selectors:     defaultServiceSelector,
			servicePort:   defaultServicePort,
			expectedError: "Service 'nsname' cannot be empty",
			apiClient:     true,
		},
		{
			name:          defaultServiceName,
			namespace:     defaultServiceNamespace,
			selectors:     map[string]string{},
			servicePort:   defaultServicePort,
			expectedError: "",
			apiClient:     true,
		},
		{
			name:          defaultServiceName,
			namespace:     defaultServiceNamespace,
			selectors:     defaultServiceSelector,
			servicePort:   corev1.ServicePort{},
			expectedError: "",
			apiClient:     true,
		},
		{
			name:          defaultServiceName,
			namespace:     defaultServiceNamespace,
			selectors:     defaultServiceSelector,
			servicePort:   defaultServicePort,
			expectedError: "",
			apiClient:     false,
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.apiClient {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		testReplicaSetBuilder := NewBuilder(testSettings,
			testCase.name,
			testCase.namespace,
			testCase.selectors,
			testCase.servicePort)

		if testCase.expectedError == "" {
			if testCase.apiClient {
				assert.NotNil(t, testReplicaSetBuilder.Definition)
				assert.Equal(t, testCase.name, testReplicaSetBuilder.Definition.Name)
				assert.Equal(t, testCase.namespace, testReplicaSetBuilder.Definition.Namespace)
			}
		} else {
			assert.Equal(t, testCase.expectedError, testReplicaSetBuilder.errorMsg)
		}
	}
}

func TestServiceExists(t *testing.T) {
	testCases := []struct {
		testService    *Builder
		expectedStatus bool
	}{
		{
			testService:    buildValidServiceBuilder(buildServiceClientWithDummyObject()),
			expectedStatus: true,
		},
		{
			testService:    buildInValidServiceBuilder(buildServiceClientWithDummyObject()),
			expectedStatus: false,
		},
		{
			testService:    buildValidServiceBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testService.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestServiceCreate(t *testing.T) {
	testCases := []struct {
		testService   *Builder
		expectedError string
	}{
		{
			testService:   buildValidServiceBuilder(buildServiceClientWithDummyObject()),
			expectedError: "",
		},
		{
			testService:   buildInValidServiceBuilder(buildServiceClientWithDummyObject()),
			expectedError: "Service 'name' cannot be empty",
		},
		{
			testService:   buildValidServiceBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testServiceBuilder, err := testCase.testService.Create()

		if testCase.expectedError == "" {
			assert.Equal(t, testServiceBuilder.Definition.Name, testServiceBuilder.Object.Name)
			assert.Equal(t, testServiceBuilder.Definition.Namespace, testServiceBuilder.Object.Namespace)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError, err.Error())
		}
	}
}

func TestServiceDelete(t *testing.T) {
	testCases := []struct {
		testService   *Builder
		expectedError error
	}{
		{
			testService:   buildValidServiceBuilder(buildServiceClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testService:   buildInValidServiceBuilder(buildServiceClientWithDummyObject()),
			expectedError: fmt.Errorf("Service 'name' cannot be empty"),
		},
		{
			testService:   buildValidServiceBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		err := testCase.testService.Delete()

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testService.Object)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestServiceUpdate(t *testing.T) {
	testCases := []struct {
		testService    *Builder
		expectedError  string
		testAnnotation map[string]string
	}{
		{
			testService:    buildValidServiceBuilder(buildServiceClientWithDummyObject()),
			expectedError:  "",
			testAnnotation: defaultServiceAnnotation,
		},
		{
			testService:    buildValidServiceBuilder(buildServiceClientWithDummyObject()),
			expectedError:  "annotation can not be empty map",
			testAnnotation: map[string]string{},
		},
	}

	for _, testCase := range testCases {
		assert.Nil(t, map[string]string(nil), testCase.testService.Object)
		testCase.testService.WithAnnotation(testCase.testAnnotation)
		_, err := testCase.testService.Update()

		if testCase.expectedError != "" {
			assert.Equal(t, testCase.expectedError, err.Error())
		} else {
			assert.Equal(t, testCase.testAnnotation,
				testCase.testService.Definition.Annotations)
		}
	}
}

func TestServiceWithNodePort(t *testing.T) {
	testCases := []struct {
		expectedErrorText string
	}{
		{
			expectedErrorText: "",
		},
		{
			expectedErrorText: "service does not have the available ports",
		},
	}

	for _, testCase := range testCases {
		if testCase.expectedErrorText != "" {
			testBuilder := buildInValidPortServiceBuilder(buildServiceClientWithDummyObject())

			result := testBuilder.WithNodePort()

			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
			}
		} else {
			testBuilder := buildValidServiceBuilder(buildServiceClientWithDummyObject())

			result := testBuilder.WithNodePort()

			assert.NotNil(t, result)
			assert.Equal(t, defaultServicePort.Port, result.Definition.Spec.Ports[0].NodePort)
		}
	}
}

func TestServiceWithExternalTrafficPolicy(t *testing.T) {
	testCases := []struct {
		policyType        corev1.ServiceExternalTrafficPolicyType
		expectedErrorText string
	}{
		{
			policyType:        corev1.ServiceExternalTrafficPolicyTypeLocal,
			expectedErrorText: "",
		},
		{
			policyType:        corev1.ServiceExternalTrafficPolicyTypeCluster,
			expectedErrorText: "",
		},
		{
			policyType:        "",
			expectedErrorText: "ExternalTrafficPolicy can not be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidServiceBuilder(buildServiceClientWithDummyObject())

		result := testBuilder.WithExternalTrafficPolicy(testCase.policyType)

		if testCase.expectedErrorText != "" {
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
			}
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, corev1.ServiceType("LoadBalancer"), result.Definition.Spec.Type)
			assert.Equal(t, testCase.policyType, result.Definition.Spec.ExternalTrafficPolicy)
		}
	}
}

func TestServiceWithSelector(t *testing.T) {
	testCases := []struct {
		testSelector   map[string]string
		expectedErrMsg string
	}{
		{
			testSelector:   map[string]string{"test-selector-key": "test-selector-value"},
			expectedErrMsg: "",
		},
		{
			testSelector:   map[string]string{"test-selector-key": ""},
			expectedErrMsg: "",
		},
		{
			testSelector:   map[string]string{"": "test-selector-value"},
			expectedErrMsg: "can not apply a selector with an empty key",
		},
		{
			testSelector:   map[string]string{},
			expectedErrMsg: "selector can not be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidServiceBuilder(buildServiceClientWithDummyObject())

		testBuilder.WithSelector(testCase.testSelector)

		assert.Equal(t, testCase.expectedErrMsg, testBuilder.errorMsg)

		if testCase.expectedErrMsg == "" {
			assert.Equal(t, testCase.testSelector, testBuilder.Definition.Spec.Selector)
		}
	}
}

func TestServiceWithLabel(t *testing.T) {
	testCases := []struct {
		testLabel      map[string]string
		expectedErrMsg string
	}{
		{
			testLabel:      map[string]string{"test-label-key": "test-label-value"},
			expectedErrMsg: "",
		},
		{
			testLabel: map[string]string{"test-label1-key": "test-label1-value",
				"test-label2-key": "test-label2-value",
				"test-label3-key": "test-label3-value",
			},
			expectedErrMsg: "",
		},
		{
			testLabel:      map[string]string{"test-label-key": ""},
			expectedErrMsg: "",
		},
		{
			testLabel:      map[string]string{"": "test-label-value"},
			expectedErrMsg: "can not apply a labels with an empty key",
		},
		{
			testLabel:      map[string]string{},
			expectedErrMsg: "labels can not be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidServiceBuilder(buildServiceClientWithDummyObject())

		testBuilder.WithLabels(testCase.testLabel)

		assert.Equal(t, testCase.expectedErrMsg, testBuilder.errorMsg)

		if testCase.expectedErrMsg == "" {
			assert.Equal(t, testCase.testLabel, testBuilder.Definition.Labels)
		}
	}
}

func TestServiceWithAnnotation(t *testing.T) {
	testCases := []struct {
		testAnnotation map[string]string
		expectedErrMsg string
	}{
		{
			testAnnotation: map[string]string{"test-annotation-key": "test-annotation-value"},
			expectedErrMsg: "",
		},
		{
			testAnnotation: map[string]string{"test-annotation1-key": "test-annotation1-value",
				"test-annotation2-key": "test-annotation2-value",
				"test-annotation3-key": "test-annotation3-value",
			},
			expectedErrMsg: "",
		},
		{
			testAnnotation: map[string]string{"test-annotation-key": ""},
			expectedErrMsg: "",
		},
		{
			testAnnotation: map[string]string{"": "test-annotation-value"},
			expectedErrMsg: "can not apply a annotation with an empty key",
		},
		{
			testAnnotation: map[string]string{},
			expectedErrMsg: "annotation can not be empty map",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidServiceBuilder(buildServiceClientWithDummyObject())

		testBuilder.WithAnnotation(testCase.testAnnotation)

		assert.Equal(t, testCase.expectedErrMsg, testBuilder.errorMsg)

		if testCase.expectedErrMsg == "" {
			assert.Equal(t, testCase.testAnnotation, testBuilder.Definition.Annotations)
		}
	}
}

func TestServiceWithIPFamily(t *testing.T) {
	testCases := []struct {
		testIPFamily      []corev1.IPFamily
		testIPStackPolicy corev1.IPFamilyPolicyType
		expectedErrorText string
	}{
		{
			testIPFamily:      []corev1.IPFamily{corev1.IPv4Protocol},
			testIPStackPolicy: corev1.IPFamilyPolicySingleStack,
			expectedErrorText: "",
		},
		{
			testIPFamily:      []corev1.IPFamily{corev1.IPv6Protocol},
			testIPStackPolicy: corev1.IPFamilyPolicySingleStack,
			expectedErrorText: "",
		},
		{
			testIPFamily:      []corev1.IPFamily{corev1.IPv6Protocol, corev1.IPv4Protocol},
			testIPStackPolicy: corev1.IPFamilyPolicySingleStack,
			expectedErrorText: "",
		},
		{
			testIPFamily:      []corev1.IPFamily{corev1.IPv6Protocol, corev1.IPv4Protocol},
			testIPStackPolicy: corev1.IPFamilyPolicyPreferDualStack,
			expectedErrorText: "",
		},
		{
			testIPFamily:      []corev1.IPFamily{corev1.IPv6Protocol, corev1.IPv4Protocol},
			testIPStackPolicy: corev1.IPFamilyPolicyRequireDualStack,
			expectedErrorText: "",
		},
		{
			testIPFamily:      []corev1.IPFamily{corev1.IPv6Protocol},
			testIPStackPolicy: corev1.IPFamilyPolicyRequireDualStack,
			expectedErrorText: "",
		},
		{
			testIPFamily:      []corev1.IPFamily{corev1.IPv4Protocol},
			testIPStackPolicy: corev1.IPFamilyPolicyRequireDualStack,
			expectedErrorText: "",
		},
		{
			testIPFamily:      []corev1.IPFamily{corev1.IPFamilyUnknown},
			testIPStackPolicy: "",
			expectedErrorText: "failed to set empty ipStackPolicy",
		},
		{
			testIPFamily:      []corev1.IPFamily{},
			testIPStackPolicy: "",
			expectedErrorText: "failed to set empty ipStackPolicy",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidServiceBuilder(buildServiceClientWithDummyObject())

		result := testBuilder.WithIPFamily(testCase.testIPFamily, testCase.testIPStackPolicy)

		if testCase.expectedErrorText != "" {
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
			}
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testIPFamily, result.Definition.Spec.IPFamilies)
			assert.Equal(t, &testCase.testIPStackPolicy, result.Definition.Spec.IPFamilyPolicy)
		}
	}
}

func TestServiceDefineServicePort(t *testing.T) {
	testCases := []struct {
		testPort       int32
		testTargetPort int32
		testProtocol   corev1.Protocol
		expectedError  string
	}{
		{
			testPort:       int32(80),
			testTargetPort: int32(8080),
			testProtocol:   corev1.ProtocolTCP,
			expectedError:  "",
		},
		{
			testPort:       int32(80),
			testTargetPort: int32(8080),
			testProtocol:   corev1.ProtocolUDP,
			expectedError:  "",
		},
		{
			testPort:       int32(80),
			testTargetPort: int32(8080),
			testProtocol:   corev1.ProtocolSCTP,
			expectedError:  "",
		},
		{
			testPort:       int32(0),
			testTargetPort: int32(8080),
			testProtocol:   corev1.ProtocolSCTP,
			expectedError:  "invalid port number",
		},
		{
			testPort:       int32(655350),
			testTargetPort: int32(8080),
			testProtocol:   corev1.ProtocolSCTP,
			expectedError:  "invalid port number",
		},
		{
			testPort:       int32(80),
			testTargetPort: int32(0),
			testProtocol:   corev1.ProtocolSCTP,
			expectedError:  "invalid target port number",
		},
		{
			testPort:       int32(80),
			testTargetPort: int32(655350),
			testProtocol:   corev1.ProtocolSCTP,
			expectedError:  "invalid target port number",
		},
		{
			testPort:       int32(80),
			testTargetPort: int32(8080),
			testProtocol:   "",
			expectedError:  "",
		},
	}

	for _, testCase := range testCases {
		testServicePort, err := DefineServicePort(testCase.testPort, testCase.testTargetPort, testCase.testProtocol)
		if testCase.expectedError != "" {
			assert.Equal(t, testCase.expectedError, err.Error())
		} else {
			assert.Equal(t, testCase.testPort, testServicePort.Port)
			assert.Equal(t, testCase.testTargetPort, testServicePort.TargetPort.IntVal)
			assert.Equal(t, testCase.testProtocol, testServicePort.Protocol)
			assert.Nil(t, err)
		}
	}
}

func buildValidServiceBuilder(apiClient *clients.Settings) *Builder {
	serviceBuilder := NewBuilder(
		apiClient,
		defaultServiceName,
		defaultServiceNamespace,
		defaultServiceSelector,
		defaultServicePort)

	return serviceBuilder
}

func buildInValidServiceBuilder(apiClient *clients.Settings) *Builder {
	serviceBuilder := NewBuilder(
		apiClient,
		"",
		defaultServiceNamespace,
		defaultServiceSelector,
		corev1.ServicePort{})

	return serviceBuilder
}

func buildInValidPortServiceBuilder(apiClient *clients.Settings) *Builder {
	serviceBuilder := &Builder{
		apiClient: apiClient,
		Definition: &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      defaultServiceName,
				Namespace: defaultServiceNamespace,
			},
		},
	}

	return serviceBuilder
}

func buildServiceClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: buildDummyService(),
	})
}

func buildDummyService() []runtime.Object {
	return append([]runtime.Object{}, &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultServiceName,
			Namespace: defaultServiceNamespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: defaultServiceSelector,
			Ports:    []corev1.ServicePort{defaultServicePort},
		},
	})
}
