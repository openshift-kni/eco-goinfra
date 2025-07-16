package nrop

import (
	"fmt"
	"testing"

	nropv1 "github.com/openshift-kni/numaresources-operator/api/v1"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultNROSchedulerName      = "numaresourcesscheduler"
	defaultNROSchedulerNamespace = "openshift-numaresources"
	defaultImageSpec             = "test-registry.com/noderesourcetopology-scheduler-image:latest"
	defaultSchedulerName         = "topo-aware-scheduler"
)

func TestNROSchedulerPull(t *testing.T) {
	generateNROScheduler := func(name, nsname string) *nropv1.NUMAResourcesScheduler {
		return &nropv1.NUMAResourcesScheduler{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		}
	}

	testCases := []struct {
		name                string
		nsname              string
		addToRuntimeObjects bool
		expectedError       error
		client              bool
	}{
		{
			name:                defaultNROSchedulerName,
			nsname:              defaultNROSchedulerNamespace,
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			nsname:              defaultNROSchedulerNamespace,
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("NUMAResourcesScheduler 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                defaultNROSchedulerName,
			nsname:              "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("NUMAResourcesScheduler 'nsname' cannot be empty"),
			client:              true,
		},
		{
			name:                "nrostest",
			nsname:              defaultNROSchedulerNamespace,
			addToRuntimeObjects: false,
			expectedError:       fmt.Errorf("NUMAResourcesScheduler object nrostest does not exist"),
			client:              true,
		},
		{
			name:                "nrostest",
			nsname:              defaultNROSchedulerNamespace,
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("NUMAResourcesScheduler 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testNROS := generateNROScheduler(testCase.name, testCase.nsname)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testNROS)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		builderResult, err := PullScheduler(testSettings, testCase.name, testCase.nsname)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		} else {
			assert.Equal(t, testNROS.Name, builderResult.Object.Name)
			assert.Equal(t, testNROS.Namespace, builderResult.Object.Namespace)
			assert.Nil(t, err)
		}
	}
}

func TestNewNROSchedulerBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		nsname        string
		expectedError string
	}{
		{
			name:          defaultNROSchedulerName,
			nsname:        defaultNROSchedulerNamespace,
			expectedError: "",
		},
		{
			name:          "",
			nsname:        defaultNROSchedulerNamespace,
			expectedError: "NUMAResourcesScheduler 'name' cannot be empty",
		},
		{
			name:          defaultNROSchedulerName,
			nsname:        "",
			expectedError: "NUMAResourcesScheduler 'nsname' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})
		testNROSBuilder := NewSchedulerBuilder(testSettings, testCase.name, testCase.nsname)
		assert.NotNil(t, testNROSBuilder.Definition)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.name, testNROSBuilder.Definition.Name)
			assert.Equal(t, testCase.nsname, testNROSBuilder.Definition.Namespace)
			assert.Equal(t, "", testNROSBuilder.errorMsg)
		} else {
			assert.Equal(t, testCase.expectedError, testNROSBuilder.errorMsg)
		}
	}
}

func TestNROSchedulerExists(t *testing.T) {
	testCases := []struct {
		testNROS       *SchedulerBuilder
		expectedStatus bool
	}{
		{
			testNROS:       buildValidNROSchedulerBuilder(buildNROSchedulerClientWithDummyObject()),
			expectedStatus: true,
		},
		{
			testNROS:       buildInValidNROSchedulerBuilder(buildNROSchedulerClientWithDummyObject()),
			expectedStatus: false,
		},
		{
			testNROS:       buildValidNROSchedulerBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testNROS.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestNROSchedulerGet(t *testing.T) {
	testCases := []struct {
		testNROS      *SchedulerBuilder
		expectedError error
	}{
		{
			testNROS:      buildValidNROSchedulerBuilder(buildNROSchedulerClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testNROS:      buildInValidNROSchedulerBuilder(buildNROSchedulerClientWithDummyObject()),
			expectedError: fmt.Errorf("NUMAResourcesScheduler 'name' cannot be empty"),
		},
		{
			testNROS: buildValidNROSchedulerBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("numaresourcesschedulers.nodetopology.openshift.io " +
				"\"numaresourcesscheduler\" not found"),
		},
	}

	for _, testCase := range testCases {
		nrosObj, err := testCase.testNROS.Get()

		if testCase.expectedError == nil {
			assert.Equal(t, nrosObj.Name, testCase.testNROS.Definition.Name)
			assert.Equal(t, nrosObj.Namespace, testCase.testNROS.Definition.Namespace)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestNROSchedulerCreate(t *testing.T) {
	testCases := []struct {
		testNROS      *SchedulerBuilder
		expectedError string
	}{
		{
			testNROS:      buildValidNROSchedulerBuilder(buildNROSchedulerClientWithDummyObject()),
			expectedError: "",
		},
		{
			testNROS:      buildInValidNROSchedulerBuilder(buildNROSchedulerClientWithDummyObject()),
			expectedError: "NUMAResourcesScheduler 'name' cannot be empty",
		},
		{
			testNROS:      buildValidNROSchedulerBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testNROSBuilder, err := testCase.testNROS.Create()

		if testCase.expectedError == "" {
			assert.Equal(t, testNROSBuilder.Definition.Name, testNROSBuilder.Object.Name)
			assert.Equal(t, testNROSBuilder.Definition.Namespace, testNROSBuilder.Object.Namespace)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError, err.Error())
		}
	}
}

func TestNROSchedulerDelete(t *testing.T) {
	testCases := []struct {
		testNROS      *SchedulerBuilder
		expectedError error
	}{
		{
			testNROS:      buildValidNROSchedulerBuilder(buildNROSchedulerClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testNROS:      buildInValidNROSchedulerBuilder(buildNROSchedulerClientWithDummyObject()),
			expectedError: fmt.Errorf("NUMAResourcesScheduler 'name' cannot be empty"),
		},
		{
			testNROS:      buildValidNROSchedulerBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		_, err := testCase.testNROS.Delete()

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testNROS.Object)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestNROSchedulerUpdate(t *testing.T) {
	testCases := []struct {
		testNROS      *SchedulerBuilder
		expectedError string
		imageSpec     string
	}{
		{
			testNROS:      buildValidNROSchedulerBuilder(buildNROSchedulerClientWithDummyObject()),
			expectedError: "",
			imageSpec:     defaultImageSpec,
		},
		{
			testNROS:      buildInValidNROSchedulerBuilder(buildNROSchedulerClientWithDummyObject()),
			expectedError: "NUMAResourcesScheduler 'name' cannot be empty",
			imageSpec:     "",
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, "", testCase.testNROS.Definition.Spec.SchedulerImage)
		assert.Nil(t, nil, testCase.testNROS.Object)
		testCase.testNROS.WithImageSpec(testCase.imageSpec)
		_, err := testCase.testNROS.Update()

		if testCase.expectedError != "" {
			assert.Equal(t, testCase.expectedError, err.Error())
		} else {
			assert.Equal(t, testCase.imageSpec, testCase.testNROS.Definition.Spec.SchedulerImage)
		}
	}
}

func TestNROSchedulerWithImageSpec(t *testing.T) {
	testCases := []struct {
		imageSpec      string
		expectedErrMsg string
	}{
		{
			imageSpec:      defaultImageSpec,
			expectedErrMsg: "",
		},
		{
			imageSpec:      "",
			expectedErrMsg: "can not apply a NUMAResourcesScheduler with an empty imageSpec",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidNROSchedulerBuilder(buildNROSchedulerClientWithDummyObject())

		testBuilder.WithImageSpec(testCase.imageSpec)

		assert.Equal(t, testCase.expectedErrMsg, testBuilder.errorMsg)

		if testCase.expectedErrMsg == "" {
			assert.Equal(t, testCase.imageSpec, testBuilder.Definition.Spec.SchedulerImage)
		}
	}
}

func TestNROSchedulerWithSchedulerName(t *testing.T) {
	testCases := []struct {
		schedulerName  string
		expectedErrMsg string
	}{
		{
			schedulerName:  defaultSchedulerName,
			expectedErrMsg: "",
		},
		{
			schedulerName:  "",
			expectedErrMsg: "can not apply a NUMAResourcesScheduler with an empty schedulerName",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidNROSchedulerBuilder(buildNROSchedulerClientWithDummyObject())

		testBuilder.WithSchedulerName(testCase.schedulerName)

		assert.Equal(t, testCase.expectedErrMsg, testBuilder.errorMsg)

		if testCase.expectedErrMsg == "" {
			assert.Equal(t, testCase.schedulerName, testBuilder.Definition.Spec.SchedulerName)
		}
	}
}

func buildValidNROSchedulerBuilder(apiClient *clients.Settings) *SchedulerBuilder {
	nrosBuilder := NewSchedulerBuilder(apiClient, defaultNROSchedulerName, defaultNROSchedulerNamespace)

	return nrosBuilder
}

func buildInValidNROSchedulerBuilder(apiClient *clients.Settings) *SchedulerBuilder {
	nrosBuilder := NewSchedulerBuilder(apiClient, "", defaultNROSchedulerNamespace)

	return nrosBuilder
}

func buildNROSchedulerClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyNROScheduler(),
		SchemeAttachers: testSchemes,
	})
}

func buildDummyNROScheduler() []runtime.Object {
	return append([]runtime.Object{}, &nropv1.NUMAResourcesScheduler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultNROSchedulerName,
			Namespace: defaultNROSchedulerNamespace,
		},
	})
}
