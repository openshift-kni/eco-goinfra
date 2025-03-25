package velero

import (
	"fmt"
	"testing"
	"time"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	testBackupStorageLocation = "test-bsl"
)

//nolint:funlen
func TestNewBackupStorageLocationBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		provider      string
		objectStorage velerov1.ObjectStorageLocation
		client        bool
		expectedError string
	}{
		{
			name:      testBackupStorageLocation,
			namespace: testBackupStorageLocation,
			provider:  "aws",
			objectStorage: velerov1.ObjectStorageLocation{
				Bucket: "test-bucket",
				Prefix: "backups",
			},
			client:        true,
			expectedError: "",
		},
		{
			name:      "",
			namespace: testBackupStorageLocation,
			provider:  "aws",
			objectStorage: velerov1.ObjectStorageLocation{
				Bucket: "test-bucket",
				Prefix: "backups",
			},
			client:        true,
			expectedError: "backupstoragelocation name cannot be empty",
		},
		{
			name:      testBackupStorageLocation,
			namespace: "",
			provider:  "aws",
			objectStorage: velerov1.ObjectStorageLocation{
				Bucket: "test-bucket",
				Prefix: "backups",
			},
			client:        true,
			expectedError: "backupstoragelocation namespace cannot be empty",
		},
		{
			name:      testBackupStorageLocation,
			namespace: testBackupStorageLocation,
			provider:  "",
			objectStorage: velerov1.ObjectStorageLocation{
				Bucket: "test-bucket",
				Prefix: "backups",
			},
			client:        true,
			expectedError: "backupstoragelocation provider cannot be empty",
		},
		{
			name:      testBackupStorageLocation,
			namespace: testBackupStorageLocation,
			provider:  "aws",
			objectStorage: velerov1.ObjectStorageLocation{
				Bucket: "",
				Prefix: "backups",
			},
			client:        true,
			expectedError: "backupstoragelocation objectstorage bucket cannot be empty",
		},
		{
			name:      testBackupStorageLocation,
			namespace: testBackupStorageLocation,
			provider:  "aws",
			objectStorage: velerov1.ObjectStorageLocation{
				Bucket: "test-bucket",
				Prefix: "backups",
			},
			client:        false,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		var (
			client *clients.Settings
		)

		if testCase.client {
			client = clients.GetTestClients(clients.TestClientParams{})
		}

		testBuilder := NewBackupStorageLocationBuilder(
			client, testCase.name, testCase.namespace, testCase.provider, testCase.objectStorage)

		if testCase.client {
			assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

			if testCase.expectedError == "" {
				assert.Equal(t, testCase.name, testBuilder.Definition.Name)
				assert.Equal(t, testCase.namespace, testBuilder.Definition.Namespace)
				assert.Equal(t, testCase.provider, testBuilder.Definition.Spec.Provider)
				assert.Equal(t, testCase.objectStorage, *testBuilder.Definition.Spec.ObjectStorage)
			}
		} else {
			assert.Nil(t, testBuilder)
		}
	}
}

func TestPullBackupStorageLocation(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		client        bool
		exists        bool
		expectedError error
	}{
		{
			name:          testBackupStorageLocation,
			namespace:     testBackupStorageLocation,
			client:        true,
			exists:        true,
			expectedError: nil,
		},
		{
			name:          "",
			namespace:     testBackupStorageLocation,
			client:        true,
			exists:        true,
			expectedError: fmt.Errorf("backupstoragelocation name cannot be empty"),
		},
		{
			name:          testBackupStorageLocation,
			namespace:     "",
			client:        true,
			exists:        true,
			expectedError: fmt.Errorf("backupstoragelocation namespace cannot be empty"),
		},
		{
			name:          testBackupStorageLocation,
			namespace:     testBackupStorageLocation,
			client:        false,
			exists:        true,
			expectedError: fmt.Errorf("the apiClient is nil"),
		},
		{
			name:      testBackupStorageLocation,
			namespace: testBackupStorageLocation,
			client:    true,
			exists:    false,
			expectedError: fmt.Errorf(
				"backupstoragelocation object %s does not exist in namespace %s",
				testBackupStorageLocation, testBackupStorageLocation),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testClient     *clients.Settings
		)

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateBackupStorageLocation())
		}

		if testCase.client {
			testClient = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: v1TestSchemes,
			})
		}

		testBuilder, err := PullBackupStorageLocationBuilder(testClient, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, testBuilder.Definition.Name)
			assert.Equal(t, testCase.namespace, testBuilder.Definition.Namespace)
		}
	}
}

func TestListBackupStorageLocationBuilder(t *testing.T) {
	testCases := []struct {
		namespace     string
		exists        bool
		client        bool
		count         int
		expectedError error
	}{
		{
			namespace:     testBackupStorageLocation,
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
			expectedError: fmt.Errorf("failed to list backupstoragelocations, 'nsname' parameter is empty"),
		},
		{
			namespace:     testBackupStorageLocation,
			exists:        false,
			client:        true,
			count:         0,
			expectedError: nil,
		},
		{
			namespace:     testBackupStorageLocation,
			exists:        true,
			client:        false,
			count:         1,
			expectedError: fmt.Errorf("the apiClient is nil"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObject []runtime.Object
			testClient    *clients.Settings
		)

		if testCase.exists {
			runtimeObject = append(runtimeObject, generateBackupStorageLocation())
		}

		if testCase.client {
			testClient = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObject, SchemeAttachers: v1TestSchemes})
		}

		testBuilders, err := ListBackupStorageLocationBuilder(testClient, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.count, len(testBuilders))
		}
	}
}

func TestBackupStorageLocationWithConfig(t *testing.T) {
	testCases := []struct {
		config        map[string]string
		expectedError string
	}{
		{
			config: map[string]string{
				"testConfig": "testData",
			},
			expectedError: "",
		},
		{
			config:        map[string]string{},
			expectedError: "backupstoragelocation cannot have empty config",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateBackupStorageLocationBuilder()
		testBuilder.WithConfig(testCase.config)

		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.config, testBuilder.Definition.Spec.Config)
		}
	}
}

func TestBackupStorageLocationWaitUntilAvailable(t *testing.T) {
	testCases := []struct {
		status        velerov1.BackupStorageLocationStatus
		expectedError error
	}{
		{
			status: velerov1.BackupStorageLocationStatus{
				Phase: velerov1.BackupStorageLocationPhaseAvailable,
			},
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObject []runtime.Object
		)

		testBSL := generateBackupStorageLocation()
		testBSL.Status = testCase.status
		runtimeObject = append(runtimeObject, testBSL)

		testBuilder := generateBackupStorageLocationBuilderWithFakeObjects(runtimeObject)

		_, err := testBuilder.WaitUntilAvailable(time.Second * 2)
		assert.Equal(t, testCase.expectedError, err)
	}
}

func TestBackupStorageLocationWaitUntilUnvailable(t *testing.T) {
	testCases := []struct {
		status        velerov1.BackupStorageLocationStatus
		expectedError error
	}{
		{
			status: velerov1.BackupStorageLocationStatus{
				Phase: velerov1.BackupStorageLocationPhaseUnavailable,
			},
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObject []runtime.Object
		)

		testBSL := generateBackupStorageLocation()
		testBSL.Status = testCase.status
		runtimeObject = append(runtimeObject, testBSL)

		testBuilder := generateBackupStorageLocationBuilderWithFakeObjects(runtimeObject)

		_, err := testBuilder.WaitUntilUnavailable(time.Second * 2)
		assert.Equal(t, testCase.expectedError, err)
	}
}

func TestBackupStorageLocationExists(t *testing.T) {
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
			runtimeObjects = append(runtimeObjects, generateBackupStorageLocation())
		}

		testBuilder := generateBackupStorageLocationBuilderWithFakeObjects(runtimeObjects)

		assert.Equal(t, testCase.exists, testBuilder.Exists())
	}
}

func TestBackupStorageLocationCreate(t *testing.T) {
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
			runtimeObjects = append(runtimeObjects, generateBackupStorageLocation())
		}

		testBuilder := generateBackupStorageLocationBuilderWithFakeObjects(runtimeObjects)

		result, err := testBuilder.Create()
		assert.Nil(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, testBackupStorageLocation, result.Definition.Name)
		assert.Equal(t, testBackupStorageLocation, result.Definition.Namespace)
	}
}

func TestBackupStorageLocationUpdate(t *testing.T) {
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
			expectedError: fmt.Errorf("cannot update non-existent backupstoragelocation"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateBackupStorageLocation())
		}

		testBuilder := generateBackupStorageLocationBuilderWithFakeObjects(runtimeObjects)

		testBuilder.Definition.Spec.Provider = "testProvider"

		testBuilder.Definition.ResourceVersion = "999"
		bsl, err := testBuilder.Update()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, bsl.Object.Spec.Provider, "testProvider")
		}
	}
}

func TestBackupStorageLocationDelete(t *testing.T) {
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
			runtimeObjects = append(runtimeObjects, generateBackupStorageLocation())
		}

		testBuilder := generateBackupStorageLocationBuilderWithFakeObjects(runtimeObjects)

		err := testBuilder.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testBuilder.Object)
		}
	}
}

func TestBackupStorageLocationValidate(t *testing.T) {
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
			expectedError: "error: received nil BackupStorageLocation builder",
		},
		{
			builderNil:    false,
			definitionNil: true,
			apiClientNil:  false,
			expectedError: "can not redefine the undefined BackupStorageLocation",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  true,
			expectedError: "BackupStorageLocation builder cannot have nil apiClient",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateBackupStorageLocationBuilderWithFakeObjects([]runtime.Object{})

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

func generateBackupStorageLocationBuilderWithFakeObjects(objects []runtime.Object) *BackupStorageLocationBuilder {
	return &BackupStorageLocationBuilder{
		apiClient: clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: objects, SchemeAttachers: v1TestSchemes}),
		Definition: generateBackupStorageLocation(),
	}
}

func generateBackupStorageLocationBuilder() *BackupStorageLocationBuilder {
	var runtimeObjects []runtime.Object
	runtimeObjects = append(runtimeObjects, generateBackupStorageLocation())

	return &BackupStorageLocationBuilder{
		apiClient: clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects: runtimeObjects, SchemeAttachers: v1TestSchemes}),
		Definition: generateBackupStorageLocation(),
	}
}

func generateBackupStorageLocation() *velerov1.BackupStorageLocation {
	return &velerov1.BackupStorageLocation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testBackupStorageLocation,
			Namespace: testBackupStorageLocation,
		},
		Spec: velerov1.BackupStorageLocationSpec{
			Provider: "aws",
			StorageType: velerov1.StorageType{
				ObjectStorage: &velerov1.ObjectStorageLocation{
					Bucket: "test-bucket",
					Prefix: "backup",
				},
			},
		},
	}
}
