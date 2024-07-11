package storage

import (
	"fmt"
	"testing"

	noobaav1alpha1 "github.com/kube-object-storage/lib-bucket-provisioner/pkg/apis/objectbucket.io/v1alpha1"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultObjectBucketClaimName      = "obc-test"
	defaultObjectBucketClaimNamespace = "obc-space"
)

func TestPullObjectBucketClaim(t *testing.T) {
	generateObjectBucketClaim := func(name, namespace string) *noobaav1alpha1.ObjectBucketClaim {
		return &noobaav1alpha1.ObjectBucketClaim{
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
			name:                defaultObjectBucketClaimName,
			namespace:           defaultObjectBucketClaimNamespace,
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           defaultObjectBucketClaimNamespace,
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("objectBucketClaim 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                defaultObjectBucketClaimName,
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("objectBucketClaim 'nsname' cannot be empty"),
			client:              true,
		},
		{
			name:                "objectBucketClaim-test",
			namespace:           defaultObjectBucketClaimNamespace,
			addToRuntimeObjects: false,
			expectedError: fmt.Errorf("objectBucketClaim object objectBucketClaim-test does not exist in " +
				"namespace obc-space"),
			client: true,
		},
		{
			name:                "objectBucketClaim-test",
			namespace:           defaultObjectBucketClaimNamespace,
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("objectBucketClaim 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testObjectBucketClaim := generateObjectBucketClaim(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testObjectBucketClaim)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})
		}

		builderResult, err := PullObjectBucketClaim(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		} else {
			assert.Equal(t, testObjectBucketClaim.Name, builderResult.Object.Name)
			assert.Equal(t, testObjectBucketClaim.Namespace, builderResult.Object.Namespace)
			assert.Nil(t, err)
		}
	}
}

func TestNewObjectBucketClaimBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		expectedError string
		client        bool
	}{
		{
			name:          defaultObjectBucketClaimName,
			namespace:     defaultObjectBucketClaimNamespace,
			expectedError: "",
			client:        true,
		},
		{
			name:          "",
			namespace:     defaultObjectBucketClaimNamespace,
			expectedError: "objectBucketClaim 'name' cannot be empty",
			client:        true,
		},
		{
			name:          defaultObjectBucketClaimName,
			namespace:     "",
			expectedError: "objectBucketClaim 'nsname' cannot be empty",
			client:        true,
		},
		{
			name:          defaultObjectBucketClaimName,
			namespace:     defaultObjectBucketClaimNamespace,
			expectedError: "",
			client:        false,
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings
		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		testObjectBucketClaimBuilder := NewObjectBucketClaimBuilder(testSettings, testCase.name, testCase.namespace)

		if testCase.expectedError == "" {
			if testCase.client {
				assert.Equal(t, testCase.name, testObjectBucketClaimBuilder.Definition.Name)
				assert.Equal(t, testCase.namespace, testObjectBucketClaimBuilder.Definition.Namespace)
			} else {
				assert.Nil(t, testObjectBucketClaimBuilder)
			}
		} else {
			assert.Equal(t, testCase.expectedError, testObjectBucketClaimBuilder.errorMsg)
			assert.NotNil(t, testObjectBucketClaimBuilder.Definition)
		}
	}
}

func TestObjectBucketClaimExists(t *testing.T) {
	testCases := []struct {
		testObjectBucketClaim *ObjectBucketClaimBuilder
		expectedStatus        bool
	}{
		{
			testObjectBucketClaim: buildValidObjectBucketClaimBuilder(buildObjectBucketClaimClientWithDummyObject()),
			expectedStatus:        true,
		},
		{
			testObjectBucketClaim: buildInValidObjectBucketClaimBuilder(buildObjectBucketClaimClientWithDummyObject()),
			expectedStatus:        false,
		},
		{
			testObjectBucketClaim: buildValidObjectBucketClaimBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus:        false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testObjectBucketClaim.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestObjectBucketClaimGet(t *testing.T) {
	testCases := []struct {
		testObjectBucketClaim *ObjectBucketClaimBuilder
		expectedError         error
	}{
		{
			testObjectBucketClaim: buildValidObjectBucketClaimBuilder(buildObjectBucketClaimClientWithDummyObject()),
			expectedError:         nil,
		},
		{
			testObjectBucketClaim: buildInValidObjectBucketClaimBuilder(buildObjectBucketClaimClientWithDummyObject()),
			expectedError:         fmt.Errorf("objectBucketClaim 'name' cannot be empty"),
		},
		{
			testObjectBucketClaim: buildValidObjectBucketClaimBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:         fmt.Errorf("objectbucketclaims.objectbucket.io \"obc-test\" not found"),
		},
	}

	for _, testCase := range testCases {
		objectBucketClaimObj, err := testCase.testObjectBucketClaim.Get()

		if testCase.expectedError == nil {
			assert.Equal(t, objectBucketClaimObj.Name, testCase.testObjectBucketClaim.Definition.Name)
			assert.Equal(t, objectBucketClaimObj.Namespace, testCase.testObjectBucketClaim.Definition.Namespace)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestObjectBucketClaimCreate(t *testing.T) {
	testCases := []struct {
		testObjectBucketClaim *ObjectBucketClaimBuilder
		expectedError         error
	}{
		{
			testObjectBucketClaim: buildValidObjectBucketClaimBuilder(buildObjectBucketClaimClientWithDummyObject()),
			expectedError:         nil,
		},
		{
			testObjectBucketClaim: buildInValidObjectBucketClaimBuilder(buildObjectBucketClaimClientWithDummyObject()),
			expectedError:         fmt.Errorf("objectBucketClaim 'name' cannot be empty"),
		},
		{
			testObjectBucketClaim: buildValidObjectBucketClaimBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:         nil,
		},
	}

	for _, testCase := range testCases {
		testObjectBucketClaimBuilder, err := testCase.testObjectBucketClaim.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testObjectBucketClaimBuilder.Definition.Name, testObjectBucketClaimBuilder.Object.Name)
			assert.Equal(t, testObjectBucketClaimBuilder.Definition.Namespace, testObjectBucketClaimBuilder.Object.Namespace)
			assert.Nil(t, err)
		}
	}
}

func TestObjectBucketClaimDelete(t *testing.T) {
	testCases := []struct {
		testObjectBucketClaim *ObjectBucketClaimBuilder
		expectedError         error
	}{
		{
			testObjectBucketClaim: buildValidObjectBucketClaimBuilder(buildObjectBucketClaimClientWithDummyObject()),
			expectedError:         nil,
		},
		{
			testObjectBucketClaim: buildValidObjectBucketClaimBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError:         nil,
		},
	}

	for _, testCase := range testCases {
		err := testCase.testObjectBucketClaim.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testObjectBucketClaim.Object)
			assert.Nil(t, err)
		}
	}
}

func TestObjectBucketClaimUpdate(t *testing.T) {
	testCases := []struct {
		testObjectBucketClaim *ObjectBucketClaimBuilder
		testStorageClassName  string
		expectedError         error
	}{
		{
			testObjectBucketClaim: buildValidObjectBucketClaimBuilder(buildObjectBucketClaimClientWithDummyObject()),
			testStorageClassName:  "gp2",
			expectedError:         nil,
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, "", testCase.testObjectBucketClaim.Definition.Spec.StorageClassName)
		assert.Nil(t, nil, testCase.testObjectBucketClaim.Object)
		testCase.testObjectBucketClaim.WithStorageClassName(testCase.testStorageClassName)
		_, err := testCase.testObjectBucketClaim.Update()

		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError, err.Error())
		} else {
			assert.Equal(t, testCase.testStorageClassName,
				testCase.testObjectBucketClaim.Definition.Spec.StorageClassName)
		}
	}
}

func TestObjectBucketClaimWithStorageClassName(t *testing.T) {
	testCases := []struct {
		testStorageClassName string
		expectedError        string
	}{
		{
			testStorageClassName: "gp2",
			expectedError:        "",
		},
		{
			testStorageClassName: "",
			expectedError:        "'storageClassName' argument cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidObjectBucketClaimBuilder(buildObjectBucketClaimClientWithDummyObject())

		result := testBuilder.WithStorageClassName(testCase.testStorageClassName)
		assert.Equal(t, testCase.expectedError, result.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testStorageClassName, result.Definition.Spec.StorageClassName)
		}
	}
}

func TestObjectBucketClaimWithGenerateBucketName(t *testing.T) {
	testCases := []struct {
		testGenerateBucketName string
		expectedError          string
	}{
		{
			testGenerateBucketName: "test-bucket-odf",
			expectedError:          "",
		},
		{
			testGenerateBucketName: "",
			expectedError:          "'generateBucketName' argument cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidObjectBucketClaimBuilder(buildObjectBucketClaimClientWithDummyObject())

		result := testBuilder.WithGenerateBucketName(testCase.testGenerateBucketName)
		assert.Equal(t, testCase.expectedError, result.errorMsg)

		if testCase.expectedError == "" {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testGenerateBucketName, result.Definition.Spec.GenerateBucketName)
		}
	}
}

func buildValidObjectBucketClaimBuilder(apiClient *clients.Settings) *ObjectBucketClaimBuilder {
	objectBucketClaimBuilder := NewObjectBucketClaimBuilder(
		apiClient, defaultObjectBucketClaimName, defaultObjectBucketClaimNamespace)

	return objectBucketClaimBuilder
}

func buildInValidObjectBucketClaimBuilder(apiClient *clients.Settings) *ObjectBucketClaimBuilder {
	objectBucketClaimBuilder := NewObjectBucketClaimBuilder(
		apiClient, "", defaultObjectBucketClaimNamespace)

	return objectBucketClaimBuilder
}

func buildObjectBucketClaimClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: buildDummyObjectBucketClaim(),
	})
}

func buildDummyObjectBucketClaim() []runtime.Object {
	return append([]runtime.Object{}, &noobaav1alpha1.ObjectBucketClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultObjectBucketClaimName,
			Namespace: defaultObjectBucketClaimNamespace,
		},
	})
}
