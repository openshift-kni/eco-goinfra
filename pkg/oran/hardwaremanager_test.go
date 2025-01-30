package oran

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	pluginv1alpha1 "github.com/openshift-kni/oran-hwmgr-plugin/api/hwmgr-plugin/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	defaultHwmgrName      = "test-hwmgr"
	defaultHwmgrNamespace = "test-ns"
)

var (
	defaultHwmgrAdaptorID = pluginv1alpha1.SupportedAdaptors.Loopback
	defaultHwmgrCondition = metav1.Condition{
		Type:   string(pluginv1alpha1.ConditionTypes.Validation),
		Status: metav1.ConditionTrue,
	}

	pluginTestSchemes = []clients.SchemeAttacher{
		pluginv1alpha1.AddToScheme,
	}
)

func TestNewHwmgrBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		nsname        string
		adaptorID     pluginv1alpha1.HardwareManagerAdaptorID
		client        bool
		expectedError string
	}{
		{
			name:          defaultHwmgrName,
			nsname:        defaultHwmgrNamespace,
			adaptorID:     pluginv1alpha1.SupportedAdaptors.Loopback,
			client:        true,
			expectedError: "",
		},
		{
			name:          "",
			nsname:        defaultHwmgrNamespace,
			adaptorID:     pluginv1alpha1.SupportedAdaptors.Loopback,
			client:        true,
			expectedError: "hardwareManager 'name' cannot be empty",
		},
		{
			name:          defaultHwmgrName,
			nsname:        "",
			adaptorID:     pluginv1alpha1.SupportedAdaptors.Loopback,
			client:        true,
			expectedError: "hardwareManager 'nsname' cannot be empty",
		},
		{
			name:          defaultHwmgrName,
			nsname:        defaultHwmgrNamespace,
			adaptorID:     "invalid-adaptor-id",
			client:        true,
			expectedError: "hardwareManager 'adaptorID' must be loopback or dell-hwmgr",
		},
		{
			name:          defaultHwmgrName,
			nsname:        defaultHwmgrNamespace,
			adaptorID:     pluginv1alpha1.SupportedAdaptors.Loopback,
			client:        false,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		testBuilder := NewHwmgrBuilder(testSettings, testCase.name, testCase.nsname, testCase.adaptorID)

		if testCase.client {
			assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

			if testCase.expectedError == "" {
				assert.Equal(t, testCase.name, testBuilder.Definition.Name)
				assert.Equal(t, testCase.nsname, testBuilder.Definition.Namespace)
				assert.Equal(t, testCase.adaptorID, testBuilder.Definition.Spec.AdaptorID)
			}
		} else {
			assert.Nil(t, testBuilder)
		}
	}
}

func TestHwmgrWithLoopbackData(t *testing.T) {
	testCases := []struct {
		adaptorID     pluginv1alpha1.HardwareManagerAdaptorID
		expectedError string
	}{
		{
			adaptorID:     pluginv1alpha1.SupportedAdaptors.Loopback,
			expectedError: "",
		},
		{
			adaptorID:     pluginv1alpha1.SupportedAdaptors.Dell,
			expectedError: "cannot set LoopbackData unless AdaptorID is loopback",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidHwmgrTestBuilder(clients.GetTestClients(clients.TestClientParams{}))
		testBuilder.Definition.Spec.AdaptorID = testCase.adaptorID

		testBuilder = testBuilder.WithLoopbackData(pluginv1alpha1.LoopbackData{})
		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)
	}
}

func TestHwmgrWithDellData(t *testing.T) {
	testCases := []struct {
		adaptorID     pluginv1alpha1.HardwareManagerAdaptorID
		authSecret    string
		apiURL        string
		expectedError string
	}{
		{
			adaptorID:     pluginv1alpha1.SupportedAdaptors.Dell,
			authSecret:    "test",
			apiURL:        "test",
			expectedError: "",
		},
		{
			adaptorID:     pluginv1alpha1.SupportedAdaptors.Loopback,
			authSecret:    "test",
			apiURL:        "test",
			expectedError: "cannot set DellData unless AdaptorID is dell-hwmgr",
		},
		{
			adaptorID:     pluginv1alpha1.SupportedAdaptors.Dell,
			authSecret:    "",
			apiURL:        "test",
			expectedError: "hardwareManager 'AuthSecret' cannot be empty",
		},
		{
			adaptorID:     pluginv1alpha1.SupportedAdaptors.Dell,
			authSecret:    "test",
			apiURL:        "",
			expectedError: "hardwareManager 'ApiUrl' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidHwmgrTestBuilder(clients.GetTestClients(clients.TestClientParams{}))
		testBuilder.Definition.Spec.AdaptorID = testCase.adaptorID

		testBuilder = testBuilder.WithDellData(pluginv1alpha1.DellData{
			AuthSecret: testCase.authSecret,
			ApiUrl:     testCase.apiURL,
		})
		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)
	}
}

func TestPullHwmgr(t *testing.T) {
	testCases := []struct {
		name                string
		nsname              string
		addToRuntimeObjects bool
		client              bool
		expectedError       error
	}{
		{
			name:                defaultHwmgrName,
			nsname:              defaultHwmgrNamespace,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       nil,
		},
		{
			name:                "",
			nsname:              defaultHwmgrNamespace,
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("hardwareManager 'name' cannot be empty"),
		},
		{
			name:                defaultHwmgrName,
			nsname:              "",
			addToRuntimeObjects: true,
			client:              true,
			expectedError:       fmt.Errorf("hardwareManager 'nsname' cannot be empty"),
		},
		{
			name:                defaultHwmgrName,
			nsname:              defaultHwmgrNamespace,
			addToRuntimeObjects: false,
			client:              true,
			expectedError: fmt.Errorf(
				"hardwareManager object %s does not exist in namespace %s", defaultHwmgrName, defaultHwmgrNamespace),
		},
		{
			name:                defaultHwmgrName,
			nsname:              defaultHwmgrNamespace,
			addToRuntimeObjects: true,
			client:              false,
			expectedError:       fmt.Errorf("hardwareManager 'apiClient' cannot be nil"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects,
				buildDummyHwmgr(defaultHwmgrName, defaultHwmgrNamespace, defaultHwmgrAdaptorID))
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: pluginTestSchemes,
			})
		}

		testBuilder, err := PullHwmgr(testSettings, testCase.name, testCase.nsname)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, testBuilder.Definition.Name)
			assert.Equal(t, testCase.nsname, testBuilder.Definition.Namespace)
		}
	}
}

func TestHwmgrGet(t *testing.T) {
	testCases := []struct {
		testBuilder   *HardwareManagerBuilder
		expectedError string
	}{
		{
			testBuilder:   buildValidHwmgrTestBuilder(buildTestClientWithDummyHwmgr()),
			expectedError: "",
		},
		{
			testBuilder:   buildInvalidHwmgrTestBuilder(buildTestClientWithDummyHwmgr()),
			expectedError: "hardwareManager 'nsname' cannot be empty",
		},
		{
			testBuilder:   buildValidHwmgrTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Sprintf("hardwaremanagers.hwmgr-plugin.oran.openshift.io \"%s\" not found", defaultHwmgrName),
		},
	}

	for _, testCase := range testCases {
		hardwareManager, err := testCase.testBuilder.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, testCase.testBuilder.Definition.Name, hardwareManager.Name)
			assert.Equal(t, testCase.testBuilder.Definition.Namespace, hardwareManager.Namespace)
		} else {
			assert.EqualError(t, err, testCase.expectedError)
		}
	}
}

func TestHwmgrExists(t *testing.T) {
	testCases := []struct {
		testBuilder *HardwareManagerBuilder
		exists      bool
	}{
		{
			testBuilder: buildValidHwmgrTestBuilder(buildTestClientWithDummyHwmgr()),
			exists:      true,
		},
		{
			testBuilder: buildInvalidHwmgrTestBuilder(buildTestClientWithDummyHwmgr()),
			exists:      false,
		},
		{
			testBuilder: buildValidHwmgrTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			exists:      false,
		},
	}

	for _, testCase := range testCases {
		exists := testCase.testBuilder.Exists()
		assert.Equal(t, testCase.exists, exists)
	}
}

func TestHwmgrCreate(t *testing.T) {
	testCases := []struct {
		testBuilder   *HardwareManagerBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidHwmgrTestBuilder(buildTestClientWithDummyHwmgr()),
			expectedError: nil,
		},
		{
			testBuilder:   buildValidHwmgrTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidHwmgrTestBuilder(buildTestClientWithDummyHwmgr()),
			expectedError: fmt.Errorf("hardwareManager 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		testBuilder, err := testCase.testBuilder.Create()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testBuilder.Definition.Name, testBuilder.Object.Name)
			assert.Equal(t, testBuilder.Definition.Namespace, testBuilder.Object.Namespace)
		}
	}
}

func TestHwmgrUpdate(t *testing.T) {
	testCases := []struct {
		testBuilder   *HardwareManagerBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidHwmgrTestBuilder(buildTestClientWithDummyHwmgr()),
			expectedError: nil,
		},
		{
			testBuilder:   buildValidHwmgrTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("cannot update non-existent hardwareManager"),
		},
		{
			testBuilder:   buildInvalidHwmgrTestBuilder(buildTestClientWithDummyHwmgr()),
			expectedError: fmt.Errorf("hardwareManager 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		assert.Nil(t, testCase.testBuilder.Definition.Spec.LoopbackData)

		testCase.testBuilder.Definition.Spec.LoopbackData = &pluginv1alpha1.LoopbackData{AddtionalInfo: "test"}
		testCase.testBuilder.Definition.ResourceVersion = "999"

		testBuilder, err := testCase.testBuilder.Update(false)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, pluginv1alpha1.LoopbackData{AddtionalInfo: "test"}, *testBuilder.Definition.Spec.LoopbackData)
		}
	}
}

func TestHwmgrDelete(t *testing.T) {
	testCases := []struct {
		testBuilder   *HardwareManagerBuilder
		expectedError error
	}{
		{
			testBuilder:   buildValidHwmgrTestBuilder(buildTestClientWithDummyHwmgr()),
			expectedError: nil,
		},
		{
			testBuilder:   buildValidHwmgrTestBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
		{
			testBuilder:   buildInvalidHwmgrTestBuilder(buildTestClientWithDummyHwmgr()),
			expectedError: fmt.Errorf("hardwareManager 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		err := testCase.testBuilder.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testBuilder.Object)
		}
	}
}

func TestHwmgrWaitForCondition(t *testing.T) {
	testCases := []struct {
		conditionMet  bool
		exists        bool
		valid         bool
		expectedError error
	}{
		{
			conditionMet:  true,
			exists:        true,
			valid:         true,
			expectedError: nil,
		},
		{
			conditionMet:  false,
			exists:        true,
			valid:         true,
			expectedError: context.DeadlineExceeded,
		},
		{
			conditionMet:  true,
			exists:        false,
			valid:         true,
			expectedError: fmt.Errorf("cannot wait for non-existent HardwareManager"),
		},
		{
			conditionMet:  true,
			exists:        true,
			valid:         false,
			expectedError: fmt.Errorf("hardwareManager 'nsname' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testBuilder    *HardwareManagerBuilder
		)

		if testCase.exists {
			hwmgr := buildDummyHwmgr(defaultHwmgrName, defaultHwmgrNamespace, defaultHwmgrAdaptorID)

			if testCase.conditionMet {
				hwmgr.Status.Conditions = append(hwmgr.Status.Conditions, defaultHwmgrCondition)
			}

			runtimeObjects = append(runtimeObjects, hwmgr)
		}

		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  runtimeObjects,
			SchemeAttachers: pluginTestSchemes,
		})

		if testCase.valid {
			testBuilder = buildValidHwmgrTestBuilder(testSettings)
		} else {
			testBuilder = buildInvalidHwmgrTestBuilder(testSettings)
		}

		_, err := testBuilder.WaitForCondition(defaultHwmgrCondition, time.Second)
		assert.Equal(t, testCase.expectedError, err)
	}
}

// buildDummyHwmgr returns a HardwareManager with the provided parameters. Plugin-specific data is not set.
func buildDummyHwmgr(
	name, nsname string, adaptorID pluginv1alpha1.HardwareManagerAdaptorID) *pluginv1alpha1.HardwareManager {
	return &pluginv1alpha1.HardwareManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: nsname,
		},
		Spec: pluginv1alpha1.HardwareManagerSpec{
			AdaptorID: adaptorID,
		},
	}
}

// buildTestClientWithDummyHwmgr returns an apiClient with the correct schemes and a HardwareManager with all defaults.
func buildTestClientWithDummyHwmgr() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: []runtime.Object{
			buildDummyHwmgr(defaultHwmgrName, defaultHwmgrNamespace, defaultHwmgrAdaptorID),
		},
		SchemeAttachers: pluginTestSchemes,
	})
}

// buildValidHwmgrTestBuilder returns a valid HardwareManagerBuilder with all defaults.
func buildValidHwmgrTestBuilder(apiClient *clients.Settings) *HardwareManagerBuilder {
	return NewHwmgrBuilder(apiClient, defaultHwmgrName, defaultHwmgrNamespace, defaultHwmgrAdaptorID)
}

// buildInvalidHwmgrTestBuilder returns an invalid HardwareManagerBuilder with all defaults, except for nsname which is
// empty.
func buildInvalidHwmgrTestBuilder(apiClient *clients.Settings) *HardwareManagerBuilder {
	return NewHwmgrBuilder(apiClient, defaultHwmgrName, "", defaultHwmgrAdaptorID)
}
