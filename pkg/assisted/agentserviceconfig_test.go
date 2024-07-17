package assisted

import (
	"fmt"
	"testing"
	"time"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fakeRuntimeClient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	agentInstallV1Beta1 "github.com/openshift-kni/eco-goinfra/pkg/schemes/assisted/api/v1beta1"
)

func TestNewAgentServiceConfigBuilder(t *testing.T) {
	testCases := []struct {
		databaseStorageSize   resource.Quantity
		filesystemStorageSize resource.Quantity
		databaseStorage       corev1.PersistentVolumeClaimSpec
		filesystemStorage     corev1.PersistentVolumeClaimSpec
		client                bool
	}{
		{
			databaseStorageSize:   resource.MustParse("50Gi"),
			filesystemStorageSize: resource.MustParse("10Gi"),
			databaseStorage: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					"ReadWriteOnce",
				},
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("50Gi"),
					},
				},
			},
			filesystemStorage: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					"ReadWriteOnce",
				},
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("10Gi"),
					},
				},
			},
			client: true,
		},
		{
			databaseStorageSize:   resource.MustParse("50Gi"),
			filesystemStorageSize: resource.MustParse("10Gi"),
			databaseStorage: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					"ReadWriteOnce",
				},
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("50Gi"),
					},
				},
			},
			filesystemStorage: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					"ReadWriteOnce",
				},
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("10Gi"),
					},
				},
			},
			client: false,
		},
	}

	for _, testcase := range testCases {
		var testSettings *clients.Settings
		if testcase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		testBuilder := NewAgentServiceConfigBuilder(
			testSettings, testcase.databaseStorage, testcase.filesystemStorage)

		if testcase.client {
			assert.Equal(t, testcase.databaseStorageSize,
				*testBuilder.Definition.Spec.DatabaseStorage.Resources.Requests.Storage())
			assert.Equal(t, testcase.filesystemStorageSize,
				*testBuilder.Definition.Spec.FileSystemStorage.Resources.Requests.Storage())
			assert.Empty(t, testBuilder.errorMsg)
		} else {
			assert.Nil(t, testBuilder)
		}
	}
}

func TestNewDefaultAgentServiceConfigBuilder(t *testing.T) {
	testCases := []struct {
		client            bool
		databaseStorage   resource.Quantity
		filesystemStorage resource.Quantity
	}{
		{
			client:            true,
			databaseStorage:   resource.MustParse(defaultDatabaseStorageSize),
			filesystemStorage: resource.MustParse(defaultFilesystemStorageSize),
		},
		{
			client:            false,
			databaseStorage:   resource.MustParse(defaultDatabaseStorageSize),
			filesystemStorage: resource.MustParse(defaultFilesystemStorageSize),
		},
	}

	for _, testcase := range testCases {
		var testSettings *clients.Settings
		if testcase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		testBuilder := NewDefaultAgentServiceConfigBuilder(testSettings)

		if testcase.client {
			assert.Equal(t, testcase.databaseStorage,
				*testBuilder.Definition.Spec.DatabaseStorage.Resources.Requests.Storage())
			assert.Equal(t, testcase.filesystemStorage,
				*testBuilder.Definition.Spec.FileSystemStorage.Resources.Requests.Storage())
			assert.Empty(t, testBuilder.errorMsg)
		} else {
			assert.Nil(t, testBuilder)
		}
	}
}

func TestPullAgentServiceConfig(t *testing.T) {
	testCases := []struct {
		client        bool
		exists        bool
		expectedError error
	}{
		{
			client:        true,
			exists:        true,
			expectedError: nil,
		},
		{
			client:        false,
			exists:        true,
			expectedError: fmt.Errorf("the apiClient is nil"),
		},
		{
			client:        true,
			exists:        false,
			expectedError: fmt.Errorf("agentserviceconfig object agent does not exist"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateAgentServiceConfig())
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{K8sMockObjects: runtimeObjects})
		}

		testBuilder, err := PullAgentServiceConfig(testSettings)
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError != nil {
			assert.Nil(t, testBuilder)
		} else {
			assert.NotNil(t, testBuilder)
			assert.Equal(t, agentServiceConfigName, testBuilder.Object.Name)
			assert.Empty(t, testBuilder.errorMsg)
		}
	}
}

func TestAgentServiceConfigGet(t *testing.T) {
	testCases := []struct {
		exists        bool
		expectedError string
	}{
		{
			exists:        true,
			expectedError: "",
		},
		{
			exists: false,
			expectedError: fmt.Sprintf("agentserviceconfigs.agent-install.openshift.io \"%s\" not found",
				agentServiceConfigName),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateAgentServiceConfig())
		}

		testBuilder := buildTestASCBuilderWithFakeObjects(runtimeObjects)

		aci, err := testBuilder.Get()
		if testCase.exists {
			assert.Nil(t, err)
			assert.NotNil(t, aci)
		} else {
			assert.NotNil(t, err)
			assert.Equal(t, testCase.expectedError, err.Error())
			assert.Nil(t, aci)
		}
	}
}
func TestAgentServiceConfigCreate(t *testing.T) {
	testCases := []struct {
		exists bool
	}{
		{
			exists: true,
		},
		{
			exists: false,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateAgentServiceConfig())
		}

		testBuilder := buildTestASCBuilderWithFakeObjects(runtimeObjects)

		result, err := testBuilder.Create()
		assert.Nil(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, agentServiceConfigName, result.Definition.Name)
	}
}
func TestAgentServiceConfigUpdate(t *testing.T) {
	testCases := []struct {
		exists        bool
		expectedError error
	}{
		{
			exists:        true,
			expectedError: nil,
		},
		{
			exists:        false,
			expectedError: fmt.Errorf("cannot update non-existent agentserviceconfig"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateAgentServiceConfig())
		}

		testBuilder := buildTestASCBuilderWithFakeObjects(runtimeObjects)

		testBuilder.Definition.Spec.IPXEHTTPRoute = "enabled"

		aci, err := testBuilder.Update(true)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, aci.Object.Spec.IPXEHTTPRoute, "enabled")
		}
	}
}
func TestAgentServiceConfigDelete(t *testing.T) {
	testCases := []struct {
		exists        bool
		expectedError error
	}{
		{
			exists:        true,
			expectedError: nil,
		},
		{
			exists:        false,
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateAgentServiceConfig())
		}

		testBuilder := buildTestASCBuilderWithFakeObjects(runtimeObjects)

		err := testBuilder.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testBuilder.Object)
		}
	}
}
func TestAgentServiceConfigDeleteAndWait(t *testing.T) {
	testCases := []struct {
		exists        bool
		expectedError error
	}{
		{
			exists:        true,
			expectedError: nil,
		},
		{
			exists:        false,
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateAgentServiceConfig())
		}

		testBuilder := buildTestASCBuilderWithFakeObjects(runtimeObjects)

		err := testBuilder.DeleteAndWait(time.Second * 1)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testBuilder.Object)
		}
	}
}
func TestAgentServiceConfigExists(t *testing.T) {
	testCases := []struct {
		exists bool
	}{
		{
			exists: true,
		},
		{
			exists: false,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateAgentServiceConfig())
		}

		testBuilder := buildTestASCBuilderWithFakeObjects(runtimeObjects)

		assert.Equal(t, testCase.exists, testBuilder.Exists())
	}
}

func TestAgentServiceConfigWithImageStorage(t *testing.T) {
	testCases := []struct {
		pvc           corev1.PersistentVolumeClaimSpec
		expectedError string
	}{
		{
			pvc: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					"ReadWriteOnce",
				},
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("30Gi"),
					},
				},
			},
			expectedError: "",
		},
		{
			pvc: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					"ReadWriteMany",
				},
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("50Gi"),
					},
				},
			},
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateTestAgentServiceConfigBuilder()
		testBuilder.WithImageStorage(testCase.pvc)
		assert.Equal(t, testCase.pvc, *testBuilder.Definition.Spec.ImageStorage)
		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)
	}
}

func TestAgentServiceConfigWithMirrorRegistryRef(t *testing.T) {
	testCases := []struct {
		mirrorRegistryConfig string
		expectedError        string
	}{
		{
			mirrorRegistryConfig: "mirror-config",
			expectedError:        "",
		},
		{
			mirrorRegistryConfig: "",
			expectedError:        "cannot add agentserviceconfig mirrorRegistryRef with empty configmap name",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateTestAgentServiceConfigBuilder()
		testBuilder.WithMirrorRegistryRef(testCase.mirrorRegistryConfig)
		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.mirrorRegistryConfig, testBuilder.Definition.Spec.MirrorRegistryRef.Name)
		}
	}
}

//nolint:lll
func TestAgentServiceConfigWithOSImage(t *testing.T) {
	testCases := []struct {
		osImages      []agentInstallV1Beta1.OSImage
		expectedError string
	}{
		{
			osImages: []agentInstallV1Beta1.OSImage{
				{
					OpenshiftVersion: "4.12",
					Version:          "412.86.202402272018-0",
					Url:              "https://mirror.openshift.com/pub/openshift-v4/x86_64/dependencies/rhcos/4.12/4.12.30/rhcos-live.x86_64.iso",
					CPUArchitecture:  "x86_64",
				},
			},
			expectedError: "",
		},
		{
			osImages: []agentInstallV1Beta1.OSImage{
				{
					OpenshiftVersion: "4.12",
					Version:          "412.86.202402272018-0",
					Url:              "https://mirror.openshift.com/pub/openshift-v4/x86_64/dependencies/rhcos/4.12/4.12.30/rhcos-live.x86_64.iso",
					CPUArchitecture:  "x86_64",
				},
				{
					OpenshiftVersion: "4.13",
					Version:          "413.92.202307260246-0",
					Url:              "https://mirror.openshift.com/pub/openshift-v4/x86_64/dependencies/rhcos/4.13/4.13.10/rhcos-live.x86_64.iso",
					CPUArchitecture:  "x86_64",
				},
			},
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateTestAgentServiceConfigBuilder()

		for _, image := range testCase.osImages {
			testBuilder.WithOSImage(image)
		}

		assert.Equal(t, testCase.osImages, testBuilder.Definition.Spec.OSImages)
		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)
	}
}
func TestAgentServiceConfigWithUnauthenticatedRegistry(t *testing.T) {
	testCases := []struct {
		resgistries   []string
		expectedError string
	}{
		{
			resgistries: []string{
				"docker.io",
			},
			expectedError: "",
		},
		{
			resgistries: []string{
				"docker.io",
				"quay.io",
			},
			expectedError: "",
		},
		{
			resgistries: []string{
				"quay.io",
				"",
			},
			expectedError: "agentserviceconfig cannot have empty unauthenticated registry",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateTestAgentServiceConfigBuilder()

		for _, registry := range testCase.resgistries {
			testBuilder.WithUnauthenticatedRegistry(registry)
		}

		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.resgistries, testBuilder.Definition.Spec.UnauthenticatedRegistries)
		}
	}
}

func TestAgentServiceConfigWithIPXEHTTPRoute(t *testing.T) {
	testCases := []struct {
		route         string
		expectedError string
	}{
		{
			route:         "enabled",
			expectedError: "",
		},
		{
			route:         "disabled",
			expectedError: "",
		},
		{
			route:         "",
			expectedError: fmt.Sprintf("agentserviceconfig passed invalid ipxeroute: , valid options: %v", validIPXEOptions),
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateTestAgentServiceConfigBuilder()

		testBuilder.WithIPXEHTTPRoute(testCase.route)
		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.route, testBuilder.Definition.Spec.IPXEHTTPRoute)
		}
	}
}
func TestAgentServiceConfigWithOptions(t *testing.T) {
	testCases := []struct {
		option        AgentServiceConfigAdditionalOptions
		validator     func(*AgentServiceConfigBuilder) bool
		expectedError string
	}{
		{
			option: func(builder *AgentServiceConfigBuilder) (*AgentServiceConfigBuilder, error) {
				builder.Definition.Spec.MustGatherImages = []agentInstallV1Beta1.MustGatherImage{{
					Name:             "acm-must-gather",
					OpenshiftVersion: "4.16",
				}}

				return builder, nil
			},
			validator: func(acib *AgentServiceConfigBuilder) bool {
				return acib.Definition.Spec.MustGatherImages[0].Name == "acm-must-gather" &&
					acib.Definition.Spec.MustGatherImages[0].OpenshiftVersion == "4.16"
			},
			expectedError: "",
		},
		{
			option: func(builder *AgentServiceConfigBuilder) (*AgentServiceConfigBuilder, error) {
				builder.Definition.Annotations = map[string]string{
					"additional-options-annotation": "true",
				}

				return builder, nil
			},
			validator: func(acib *AgentServiceConfigBuilder) bool {
				return acib.Definition.Annotations["additional-options-annotation"] == "true"
			},
			expectedError: "",
		},
		{
			option: func(builder *AgentServiceConfigBuilder) (*AgentServiceConfigBuilder, error) {
				return builder, fmt.Errorf("error: got agentserviceconfig with error")
			},
			validator: func(acib *AgentServiceConfigBuilder) bool {
				return acib.errorMsg != ""
			},
			expectedError: "error: got agentserviceconfig with error",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateTestAgentServiceConfigBuilder()

		testBuilder.WithOptions(testCase.option)
		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)
		assert.True(t, testCase.validator(testBuilder))
	}
}
func TestAgentServiceConfigWaitUntilDeployed(t *testing.T) {
	testCases := []struct {
		status        agentInstallV1Beta1.AgentServiceConfigStatus
		exists        bool
		expectedErorr error
	}{
		{
			status: agentInstallV1Beta1.AgentServiceConfigStatus{
				Conditions: []conditionsv1.Condition{
					{
						Type:   agentInstallV1Beta1.ConditionDeploymentsHealthy,
						Status: corev1.ConditionTrue,
					},
				},
			},
			exists:        true,
			expectedErorr: nil,
		},
		{
			status: agentInstallV1Beta1.AgentServiceConfigStatus{
				Conditions: []conditionsv1.Condition{
					{
						Type:   agentInstallV1Beta1.ConditionDeploymentsHealthy,
						Status: corev1.ConditionTrue,
					},
				},
			},
			exists:        false,
			expectedErorr: fmt.Errorf("cannot wait for non-existent agentserviceconfig to be deployed"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		testAgentServiceConfig := generateAgentServiceConfig()
		testAgentServiceConfig.Status = testCase.status

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, testAgentServiceConfig)
		}

		testBuilder := buildTestASCBuilderWithFakeObjects(runtimeObjects)
		_, err := testBuilder.WaitUntilDeployed(time.Second * 2)

		assert.Equal(t, testCase.expectedErorr, err)
	}
}

func TestAgentServiceConfigGetDefaultStorageSpec(t *testing.T) {
	testCases := []struct {
		size          string
		expectedError error
	}{
		{
			size:          defaultDatabaseStorageSize,
			expectedError: nil,
		},
		{
			size:          defaultImageStoreStorageSize,
			expectedError: nil,
		},
		{
			size:          defaultFilesystemStorageSize,
			expectedError: nil,
		},
		{
			size:          "30GiB",
			expectedError: fmt.Errorf("the storage size is in wrong format"),
		},
		{
			size:          "",
			expectedError: fmt.Errorf("the storage size is in wrong format"),
		},
	}

	for _, testCase := range testCases {
		pvc, err := GetDefaultStorageSpec(testCase.size)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.NotNil(t, pvc)
		}
	}
}

func TestAgentServiceConfigValidate(t *testing.T) {
	testCases := []struct {
		builderNil    bool
		definitionNil bool
		apiClientNil  bool
		expectedError string
	}{
		{
			builderNil:    true,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: "error: received nil AgentServiceConfig builder",
		},
		{
			builderNil:    false,
			definitionNil: true,
			apiClientNil:  false,
			expectedError: "can not redefine the undefined AgentServiceConfig",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  true,
			expectedError: "AgentServiceConfig builder cannot have nil apiClient",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildTestASCBuilderWithFakeObjects([]runtime.Object{})

		if testCase.builderNil {
			testBuilder = nil
		}

		if testCase.definitionNil {
			testBuilder.Definition = nil
		}

		if testCase.apiClientNil {
			testBuilder.apiClient = nil
		}

		result, err := testBuilder.validate()
		if testCase.expectedError != "" {
			assert.NotNil(t, err)
			assert.Equal(t, testCase.expectedError, err.Error())
			assert.False(t, result)
		} else {
			assert.Nil(t, err)
			assert.True(t, result)
		}
	}
}

func generateAgentServiceConfig() *agentInstallV1Beta1.AgentServiceConfig {
	return &agentInstallV1Beta1.AgentServiceConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: agentServiceConfigName,
		},
		Spec: agentInstallV1Beta1.AgentServiceConfigSpec{
			DatabaseStorage: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					"ReadWriteOnce",
				},
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse(defaultDatabaseStorageSize),
					},
				},
			},
			FileSystemStorage: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					"ReadWriteOnce",
				},
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse(defaultFilesystemStorageSize),
					},
				},
			},
		},
	}
}

func generateTestAgentServiceConfigBuilder() *AgentServiceConfigBuilder {
	return &AgentServiceConfigBuilder{
		apiClient:  clients.GetTestClients(clients.TestClientParams{}).Client,
		Definition: generateAgentServiceConfig(),
	}
}

func buildTestASCBuilderWithFakeObjects(objects []runtime.Object) *AgentServiceConfigBuilder {
	fakeClientScheme := runtime.NewScheme()

	err := clients.SetScheme(fakeClientScheme)
	if err != nil {
		return nil
	}

	testBuilder := NewDefaultAgentServiceConfigBuilder(&clients.Settings{
		Client: fakeRuntimeClient.NewClientBuilder().WithScheme(fakeClientScheme).WithRuntimeObjects(objects...).Build(),
	})

	return testBuilder
}
