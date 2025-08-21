package oadp

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/oadp/oadptypes"
	"github.com/stretchr/testify/assert"
	v1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
)

var (
	DPAGVK = schema.GroupVersionKind{
		Group:   APIGroup,
		Version: V1Alpha1Version,
		Kind:    DPAKind,
	}
)

const (
	testDataProtectionApplication = "test-dataprotectionapplication"
)

//nolint:funlen
func TestNewDPABuilder(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		config        oadptypes.ApplicationConfig
		client        bool
		expectedError string
	}{
		{
			name:      testDataProtectionApplication,
			namespace: testDataProtectionApplication,
			config: oadptypes.ApplicationConfig{
				Velero: &oadptypes.VeleroConfig{
					DefaultPlugins: []oadptypes.DefaultPlugin{
						oadptypes.DefaultPluginAWS,
						oadptypes.DefaultPluginOpenShift,
					},
					ResourceTimeout: "10m",
				},
			},
			client:        true,
			expectedError: "",
		},
		{
			name:      "",
			namespace: testDataProtectionApplication,
			config: oadptypes.ApplicationConfig{
				Velero: &oadptypes.VeleroConfig{
					DefaultPlugins: []oadptypes.DefaultPlugin{
						oadptypes.DefaultPluginAWS,
						oadptypes.DefaultPluginOpenShift,
					},
					ResourceTimeout: "10m",
				},
			},
			client:        true,
			expectedError: "dataprotectionapplication 'name' cannot be empty",
		},
		{
			name:      testDataProtectionApplication,
			namespace: "",
			config: oadptypes.ApplicationConfig{
				Velero: &oadptypes.VeleroConfig{
					DefaultPlugins: []oadptypes.DefaultPlugin{
						oadptypes.DefaultPluginAWS,
						oadptypes.DefaultPluginOpenShift,
					},
					ResourceTimeout: "10m",
				},
			},
			client:        true,
			expectedError: "dataprotectionapplication 'namespace' cannot be empty",
		},
		{
			name:          testDataProtectionApplication,
			namespace:     testDataProtectionApplication,
			config:        oadptypes.ApplicationConfig{},
			client:        true,
			expectedError: "dataprotectionapplication velero config cannot be empty",
		},
		{
			name:      testDataProtectionApplication,
			namespace: testDataProtectionApplication,
			config: oadptypes.ApplicationConfig{
				Velero: &oadptypes.VeleroConfig{
					DefaultPlugins: []oadptypes.DefaultPlugin{
						oadptypes.DefaultPluginAWS,
						oadptypes.DefaultPluginOpenShift,
					},
					ResourceTimeout: "10m",
				},
			},
			client:        false,
			expectedError: "",
		},
	}

	for _, testcase := range testCases {
		var testSettings *clients.Settings
		if testcase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		testBuilder := NewDPABuilder(
			testSettings, testcase.name, testcase.namespace, testcase.config)

		if testcase.client {
			assert.NotNil(t, testBuilder)
			assert.Equal(t, testcase.expectedError, testBuilder.errorMsg)
			assert.Equal(t, testcase.name, testBuilder.Definition.Name)
			assert.Equal(t, testcase.namespace, testBuilder.Definition.Namespace)
			assert.Equal(t, testcase.config, *testBuilder.Definition.Spec.Configuration)
		} else {
			assert.Nil(t, testBuilder)
		}
	}
}

//nolint:lll
func TestPullDPA(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		client        bool
		exists        bool
		expectedError error
	}{
		{
			name:          testDataProtectionApplication,
			namespace:     testDataProtectionApplication,
			client:        true,
			exists:        true,
			expectedError: nil,
		},
		{
			name:          "",
			namespace:     testDataProtectionApplication,
			client:        true,
			exists:        true,
			expectedError: fmt.Errorf("dataprotectionapplication 'name' cannot be empty"),
		},
		{
			name:          testDataProtectionApplication,
			namespace:     "",
			client:        true,
			exists:        true,
			expectedError: fmt.Errorf("dataprotectionapplication 'namespace' cannot be empty"),
		},
		{
			name:          testDataProtectionApplication,
			namespace:     testDataProtectionApplication,
			client:        false,
			exists:        true,
			expectedError: fmt.Errorf("the apiClient is nil"),
		},
		{
			name:          testDataProtectionApplication,
			namespace:     testDataProtectionApplication,
			client:        true,
			exists:        false,
			expectedError: fmt.Errorf("dataprotectionapplication object test-dataprotectionapplication does not exist in namespace test-dataprotectionapplication"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateDataProtectionApplication())
		}

		if testCase.client {
			testSettings = clients.GetTestClients(
				clients.TestClientParams{GVK: []schema.GroupVersionKind{DPAGVK}, K8sMockObjects: runtimeObjects})
		}

		testBuilder, err := PullDPA(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError != nil {
			assert.Nil(t, testBuilder)
		} else {
			assert.Equal(t, testCase.name, testBuilder.Object.Name)
			assert.Equal(t, testCase.name, testBuilder.Definition.Name)
			assert.Equal(t, testCase.namespace, testBuilder.Object.Namespace)
			assert.Equal(t, testCase.namespace, testBuilder.Definition.Namespace)
		}
	}
}

func TestListDataProtectionApplication(t *testing.T) {
	testCases := []struct {
		namespace     string
		exists        bool
		client        bool
		count         int
		expectedError error
	}{
		{
			namespace:     testDataProtectionApplication,
			exists:        true,
			client:        true,
			count:         1,
			expectedError: nil,
		},
		{
			namespace:     "",
			exists:        true,
			client:        true,
			count:         1,
			expectedError: fmt.Errorf("failed to list dataprotectionapplications, 'nsname' parameter is empty"),
		},
		{
			namespace:     testDataProtectionApplication,
			exists:        false,
			client:        true,
			count:         0,
			expectedError: nil,
		},
		{
			namespace:     testDataProtectionApplication,
			exists:        true,
			client:        false,
			count:         1,
			expectedError: fmt.Errorf("the apiClient is nil"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjectLists []runtime.Object
			testClient         *clients.Settings
		)

		dpas := generateDataProtectionApplicationList()

		if testCase.exists {
			runtimeObjectLists = append(runtimeObjectLists, dpas)
		}

		if testCase.client {
			testClient = clients.GetTestClients(clients.TestClientParams{})

			testScheme := runtime.NewScheme()
			testScheme.AddKnownTypes(oadptypes.DPAGroupVersion, runtimeObjectLists...)
			testScheme.AddKnownTypes(oadptypes.DPAGroupVersion, []runtime.Object{&dpas.Items[0]}...)

			testClient.Interface = fake.NewSimpleDynamicClientWithCustomListKinds(
				testScheme, map[schema.GroupVersionResource]string{
					GetDataProtectionApplicationGVR(): "DataProtectionApplicationList",
				}, runtimeObjectLists...)
		}

		testBuilders, err := ListDataProtectionApplication(testClient, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.count, len(testBuilders))
		}
	}
}

func TestDPAWithBackupLocation(t *testing.T) {
	testCases := []struct {
		backupLocation oadptypes.BackupLocation
		expectError    string
	}{
		{
			backupLocation: oadptypes.BackupLocation{
				Velero: &v1.BackupStorageLocationSpec{
					Provider: "aws",
					StorageType: v1.StorageType{
						ObjectStorage: &v1.ObjectStorageLocation{
							Bucket: "test-bucket",
							Prefix: "backup",
						},
					},
					Config: map[string]string{
						"insecureSkipTLSVerify": "true",
						"s3Url":                 "http://example.com/",
					},
				},
			},
			expectError: "",
		},
		{
			backupLocation: oadptypes.BackupLocation{},
			expectError:    "dataprotectionapplication backuplocation cannot have empty velero config",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateDPABuilder()
		testBuilder.WithBackupLocation(testCase.backupLocation)
		assert.Equal(t, testCase.expectError, testBuilder.errorMsg)

		if testCase.expectError == "" {
			assert.Equal(t, testCase.backupLocation, testBuilder.Definition.Spec.BackupLocations[0])
		}
	}
}

func TestDPAGet(t *testing.T) {
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
			expectedError: fmt.Sprintf("dataprotectionapplications.oadp.openshift.io \"%s\" not found",
				testDataProtectionApplication),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateDataProtectionApplication())
		}

		testBuilder := buildTestDPABuilderWithFakeObjects(runtimeObjects)

		aci, err := testBuilder.Get()
		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.NotNil(t, aci)
		} else {
			assert.Equal(t, testCase.expectedError, err.Error())
			assert.Nil(t, aci)
		}
	}
}

func TestDPAExists(t *testing.T) {
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
			runtimeObjects = append(runtimeObjects, generateDataProtectionApplication())
		}

		testBuilder := buildTestDPABuilderWithFakeObjects(runtimeObjects)

		assert.Equal(t, testCase.exists, testBuilder.Exists())
	}
}

func TestDPACreate(t *testing.T) {
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
			runtimeObjects = append(runtimeObjects, generateDataProtectionApplication())
		}

		testBuilder := buildTestDPABuilderWithFakeObjects(runtimeObjects)

		result, err := testBuilder.Create()
		assert.Nil(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, testDataProtectionApplication, result.Definition.Name)
		assert.Equal(t, testDataProtectionApplication, result.Definition.Namespace)
	}
}

func TestDPAUpdate(t *testing.T) {
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
			expectedError: fmt.Errorf("failed to update dataprotectionapplication, object does not exist on cluster"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateDataProtectionApplication())
		}

		testBuilder := buildTestDPABuilderWithFakeObjects(runtimeObjects)

		True := true
		testBuilder.Definition.Spec.BackupImages = &True

		dpa, err := testBuilder.Update(true)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, &True, dpa.Object.Spec.BackupImages)
		}
	}
}

func TestDPADelete(t *testing.T) {
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
			runtimeObjects = append(runtimeObjects, generateDataProtectionApplication())
		}

		testBuilder := buildTestDPABuilderWithFakeObjects(runtimeObjects)

		err := testBuilder.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testBuilder.Object)
		}
	}
}

func TestDPAValidate(t *testing.T) {
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
			expectedError: "error: received nil DataProtectionApplication builder",
		},
		{
			builderNil:    false,
			definitionNil: true,
			apiClientNil:  false,
			expectedError: "can not redefine the undefined DataProtectionApplication",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  true,
			expectedError: "DataProtectionApplication builder cannot have nil apiClient",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildTestDPABuilderWithFakeObjects([]runtime.Object{})

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

func buildTestDPABuilderWithFakeObjects(objects []runtime.Object) *DPABuilder {
	testClient := clients.GetTestClients(
		clients.TestClientParams{GVK: []schema.GroupVersionKind{DPAGVK}, K8sMockObjects: objects})

	testBuilder := NewDPABuilder(
		testClient, testDataProtectionApplication, testDataProtectionApplication, oadptypes.ApplicationConfig{
			Velero: &oadptypes.VeleroConfig{
				DefaultPlugins: []oadptypes.DefaultPlugin{
					oadptypes.DefaultPluginAWS,
					oadptypes.DefaultPluginOpenShift,
				},
				ResourceTimeout: "10m",
			},
		})

	return testBuilder
}

func generateDPABuilder() *DPABuilder {
	return &DPABuilder{
		apiClient:  clients.GetTestClients(clients.TestClientParams{GVK: []schema.GroupVersionKind{DPAGVK}}),
		Definition: generateDataProtectionApplication(),
	}
}

func generateDataProtectionApplicationList() *oadptypes.DataProtectionApplicationList {
	return &oadptypes.DataProtectionApplicationList{
		TypeMeta: metav1.TypeMeta{
			Kind:       APIGroup,
			APIVersion: V1Alpha1Version,
		},
		Items: []oadptypes.DataProtectionApplication{
			*generateDataProtectionApplication(),
		},
	}
}

func generateDataProtectionApplication() *oadptypes.DataProtectionApplication {
	return &oadptypes.DataProtectionApplication{
		TypeMeta: metav1.TypeMeta{
			Kind:       APIGroup,
			APIVersion: V1Alpha1Version,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      testDataProtectionApplication,
			Namespace: testDataProtectionApplication,
		},
		Spec: oadptypes.DataProtectionApplicationSpec{
			Configuration: &oadptypes.ApplicationConfig{
				Velero: &oadptypes.VeleroConfig{
					DefaultPlugins: []oadptypes.DefaultPlugin{
						oadptypes.DefaultPluginAWS,
						oadptypes.DefaultPluginOpenShift,
					},
					ResourceTimeout: "10m",
				},
			},
			BackupLocations: []oadptypes.BackupLocation{
				{
					Velero: &v1.BackupStorageLocationSpec{
						Provider: "aws",
						StorageType: v1.StorageType{
							ObjectStorage: &v1.ObjectStorageLocation{
								Bucket: "test-bucket",
								Prefix: "backup",
							},
						},
						Config: map[string]string{
							"insecureSkipTLSVerify": "true",
							"s3Url":                 "http://example.com/",
						},
					},
				},
			},
		},
	}
}
