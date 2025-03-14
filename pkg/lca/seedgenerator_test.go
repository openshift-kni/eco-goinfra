package lca

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	lcasgv1 "github.com/openshift-kni/lifecycle-agent/api/seedgenerator/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	lcasgv1TestSchemes = []clients.SchemeAttacher{
		lcasgv1.AddToScheme,
	}
)

func TestNewSeedGeneratorBuilder(t *testing.T) {
	testCases := []struct {
		name        string
		expectedErr string
	}{
		{
			name:        seedImageName,
			expectedErr: "",
		},
		{
			name:        "",
			expectedErr: "SeedGenerator name must be seedimage",
		},
		{
			name:        "seedgen2",
			expectedErr: "SeedGenerator name must be seedimage",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})

		testBuilder := NewSeedGeneratorBuilder(testSettings, testCase.name)

		assert.Equal(t, testCase.expectedErr, testBuilder.errorMsg)

		if testCase.expectedErr == "" {
			assert.NotNil(t, testBuilder)
			assert.Equal(t, testCase.name, testBuilder.Definition.Name)
		}
	}
}

func TestSeedGeneratorWithOptions(t *testing.T) {
	testBuilder := buildTestBuilderWithFakeObjects()

	testBuilder.WithOptions(func(builder *SeedGeneratorBuilder) (*SeedGeneratorBuilder, error) {
		return builder, nil
	})

	assert.Equal(t, "", testBuilder.errorMsg)

	testBuilder.WithOptions(func(builder *SeedGeneratorBuilder) (*SeedGeneratorBuilder, error) {
		return builder, fmt.Errorf("error")
	})

	assert.Equal(t, "error", testBuilder.errorMsg)
}

func TestSeedGeneratorCreate(t *testing.T) {
	testCases := []struct {
		seedGenerator *SeedGeneratorBuilder
		expectedError error
	}{
		{
			seedGenerator: buildValidSeedGeneratorBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
		{
			seedGenerator: buildInValidSeedGeneratorBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("SeedGenerator name must be seedimage"),
		},
	}

	for _, testCase := range testCases {
		testSeedGeneratorBuilder, err := testCase.seedGenerator.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testSeedGeneratorBuilder.Definition.Name, testSeedGeneratorBuilder.Object.Name)
		}
	}
}

func TestSeedGeneratorPull(t *testing.T) {
	testCases := []struct {
		name                string
		expectedError       error
		addToRuntimeObjects bool
		client              bool
	}{
		{
			name:                "",
			expectedError:       fmt.Errorf("seedgenerator 'name' cannot be empty"),
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			name:                "notseedimage",
			expectedError:       fmt.Errorf("seedgenerator object notseedimage does not exist"),
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			name:                seedImageName,
			expectedError:       nil,
			addToRuntimeObjects: true,
			client:              true,
		},
		{
			name:                seedImageName,
			expectedError:       fmt.Errorf("the apiClient is nil"),
			addToRuntimeObjects: true,
			client:              false,
		},
		{
			name:                seedImageName,
			expectedError:       fmt.Errorf("seedgenerator object seedimage does not exist"),
			addToRuntimeObjects: false,
			client:              true,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		testSG := generateSeedGenerator(seedImageName)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testSG)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: lcasgv1TestSchemes,
			})
		}

		// Test the PullSeedGenerator function
		builderResult, err := PullSeedGenerator(testSettings, testCase.name)

		// Check the error
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.NotNil(t, builderResult)
		}
	}
}

func TestSeedGeneratorDelete(t *testing.T) {
	testCases := []struct {
		seedGenerator *SeedGeneratorBuilder
		expectedError error
	}{
		{
			seedGenerator: buildValidSeedGeneratorBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
		{
			seedGenerator: buildInValidSeedGeneratorBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("SeedGenerator name must be seedimage"),
		},
		{
			seedGenerator: buildValidSeedGeneratorBuilder(
				buildSeedGeneratorTestClientWithDummyObject(buildDummySeedGeneratorRuntime())),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		testSeedGeneratorBuilder, err := testCase.seedGenerator.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testSeedGeneratorBuilder.Object)
		}
	}
}

func TestSeedGeneratorGet(t *testing.T) {
	testCases := []struct {
		seedGenerator *SeedGeneratorBuilder
		expectedError error
	}{
		{
			seedGenerator: buildValidSeedGeneratorBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("seedgenerators.lca.openshift.io \"seedimage\" not found"),
		},
		{
			seedGenerator: buildInValidSeedGeneratorBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("SeedGenerator name must be seedimage"),
		},
		{
			seedGenerator: buildValidSeedGeneratorBuilder(
				buildSeedGeneratorTestClientWithDummyObject(buildDummySeedGeneratorRuntime())),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		testSeedGenerator, err := testCase.seedGenerator.Get()
		if err != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}

		if testCase.expectedError == nil {
			assert.Equal(t, testSeedGenerator.Name, testCase.seedGenerator.Definition.Name)
		}
	}
}

func TestSeedGeneratorExists(t *testing.T) {
	testCases := []struct {
		seedGenerator  *SeedGeneratorBuilder
		expectedStatus bool
	}{
		{
			seedGenerator:  buildValidSeedGeneratorBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus: false,
		},
		{
			seedGenerator:  buildInValidSeedGeneratorBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus: false,
		},
		{
			seedGenerator: buildValidSeedGeneratorBuilder(
				buildSeedGeneratorTestClientWithDummyObject(buildDummySeedGeneratorRuntime())),
			expectedStatus: true,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.seedGenerator.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestSeedGeneratorWithSeedImage(t *testing.T) {
	testCases := []struct {
		seedGenerator *SeedGeneratorBuilder
		expectedError string
		seedImage     string
	}{
		{
			seedGenerator: buildValidSeedGeneratorBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: "",
			seedImage:     "testSeedImage",
		},
		{
			seedGenerator: buildValidSeedGeneratorBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: "seedImage 'name' cannot be empty",
			seedImage:     "",
		},
	}

	for _, testCase := range testCases {
		testSeedGeneratorBuilder := testCase.seedGenerator.WithSeedImage(testCase.seedImage)

		assert.Nil(t, testSeedGeneratorBuilder.Object)
		assert.Equal(t, testCase.expectedError, testSeedGeneratorBuilder.errorMsg)
	}
}

func TestSeedGeneratorWithRecertImage(t *testing.T) {
	testCases := []struct {
		seedGenerator *SeedGeneratorBuilder
		expectedError string
		recertImage   string
	}{
		{
			seedGenerator: buildValidSeedGeneratorBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: "",
			recertImage:   "testSeedImage",
		},
		{
			seedGenerator: buildValidSeedGeneratorBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: "recertImage 'name' cannot be empty",
			recertImage:   "",
		},
	}

	for _, testCase := range testCases {
		testSeedGeneratorBuilder := testCase.seedGenerator.WithRecertImage(testCase.recertImage)

		assert.Nil(t, testSeedGeneratorBuilder.Object)

		assert.Equal(t, testCase.expectedError, testSeedGeneratorBuilder.errorMsg)
	}
}

func TestSeedGeneratorWaitUntilComplete(t *testing.T) {
	testCases := []struct {
		expectedError error
		status        lcasgv1.SeedGeneratorStatus
	}{
		{
			expectedError: context.DeadlineExceeded,
			status: lcasgv1.SeedGeneratorStatus{
				Conditions: []metav1.Condition{{Status: "True1", Type: "SeedGenCompleted", Reason: "Completed"}},
			},
		},
		{
			expectedError: nil,
			status: lcasgv1.SeedGeneratorStatus{
				Conditions: []metav1.Condition{{Status: "True", Type: "SeedGenCompleted", Reason: "Completed"}},
			},
		},
	}

	for _, testCase := range testCases {
		testSeedGenetator := generateSeedGenerator(seedImageName)
		testSeedGenetator.Status = testCase.status

		var runtimeObjects []runtime.Object
		runtimeObjects = append(runtimeObjects, testSeedGenetator)

		testSeedGeneratorBuilder := buildValidSeedGeneratorBuilder(
			buildSeedGeneratorTestClientWithDummyObject(runtimeObjects))
		_, err := testSeedGeneratorBuilder.WaitUntilComplete(time.Second * 1)

		// Check the error
		assert.Equal(t, testCase.expectedError, err)
	}
}

func buildTestBuilderWithFakeObjects() *SeedGeneratorBuilder {
	return NewSeedGeneratorBuilder(
		buildSeedGeneratorTestClientWithDummyObject(buildDummySeedGeneratorRuntime()), seedImageName)
}

func generateSeedGenerator(name string) *lcasgv1.SeedGenerator {
	return &lcasgv1.SeedGenerator{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: lcasgv1.SeedGeneratorSpec{},
	}
}

func buildValidSeedGeneratorBuilder(apiClient *clients.Settings) *SeedGeneratorBuilder {
	return NewSeedGeneratorBuilder(apiClient, seedImageName)
}

func buildInValidSeedGeneratorBuilder(apiClient *clients.Settings) *SeedGeneratorBuilder {
	return NewSeedGeneratorBuilder(apiClient, "")
}

func buildSeedGeneratorTestClientWithDummyObject(objects []runtime.Object) *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  objects,
		SchemeAttachers: lcasgv1TestSchemes,
	})
}

func buildDummySeedGeneratorRuntime() []runtime.Object {
	return append([]runtime.Object{}, &lcasgv1.SeedGenerator{
		ObjectMeta: metav1.ObjectMeta{
			Name: seedImageName,
		},
		Spec: lcasgv1.SeedGeneratorSpec{},
	})
}
