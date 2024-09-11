package scc

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	securityV1 "github.com/openshift/api/security/v1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	securityV1Scheme = []clients.SchemeAttacher{
		securityV1.Install,
	}
)

func TestSCCNewBuilder(t *testing.T) {
	testCases := []struct {
		name           string
		user           string
		seLinuxContent string
		expectedErrMsg string
	}{
		{
			name:           "test-name",
			user:           "test-user",
			seLinuxContent: "default",
			expectedErrMsg: "",
		},
		{
			name:           "",
			user:           "test-user",
			seLinuxContent: "default",
			expectedErrMsg: "securityContextConstraints 'name' cannot be empty",
		},
		{
			name:           "test-name",
			user:           "",
			seLinuxContent: "default",
			expectedErrMsg: "securityContextConstraints 'runAsUser' cannot be empty",
		},
		{
			name:           "test-name",
			user:           "test-user",
			seLinuxContent: "",
			expectedErrMsg: "securityContextConstraints 'selinuxContext' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := NewBuilder(
			clients.GetTestClients(clients.TestClientParams{}), testCase.name, testCase.user, testCase.seLinuxContent)
		assert.Equal(t, testBuilder.errorMsg, testCase.expectedErrMsg)
	}
}

func TestSCCPull(t *testing.T) {
	generateSCC := func(name string) *securityV1.SecurityContextConstraints {
		return &securityV1.SecurityContextConstraints{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		}
	}

	testCases := []struct {
		name                string
		expectedError       bool
		addToRuntimeObjects bool
		expectedErrorText   string
	}{
		{
			name:                "test-name",
			addToRuntimeObjects: true,
			expectedError:       false,
			expectedErrorText:   "",
		},
		{
			name:                "",
			addToRuntimeObjects: true,
			expectedError:       true,
			expectedErrorText:   "securityContextConstraints 'name' cannot be empty",
		},
		{
			name:                "test-name",
			addToRuntimeObjects: false,
			expectedError:       true,
			expectedErrorText:   "securityContextConstraints object test-name does not exist",
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testSCC := generateSCC(testCase.name)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testSCC)
		}

		testSettings = clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  runtimeObjects,
			SchemeAttachers: securityV1Scheme,
		})
		// Test the Pull method
		builderResult, err := Pull(testSettings, testCase.name)

		// Check the error
		if testCase.expectedError {
			assert.NotNil(t, err)

			// Check the error message
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, err.Error())
			}
		} else {
			assert.Nil(t, err)
			assert.Equal(t, testCase.name, builderResult.Object.Name)
		}
	}
}

func TestSCCGet(t *testing.T) {
	testCases := []struct {
		sccBuilder    *Builder
		expectedError string
	}{
		{
			sccBuilder:    buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedError: "",
		},
		{
			sccBuilder:    buildInValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedError: "securityContextConstraints 'selinuxContext' cannot be empty",
		},
		{
			sccBuilder: buildValidSCCBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: securityV1Scheme})),
			expectedError: "securitycontextconstraintses.security.openshift.io \"testscc\" not found",
		},
	}

	for _, testCase := range testCases {
		scc, err := testCase.sccBuilder.Get()

		if testCase.expectedError == "" {
			assert.Nil(t, err)
			assert.Equal(t, scc.Name, testCase.sccBuilder.Definition.Name)
		} else {
			assert.Equal(t, err.Error(), testCase.expectedError)
		}
	}
}

func TestSCCExist(t *testing.T) {
	testCases := []struct {
		sccBuilder     *Builder
		expectedStatus bool
	}{
		{
			sccBuilder:     buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedStatus: true,
		},
		{
			sccBuilder:     buildInValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedStatus: false,
		},
		{
			sccBuilder: buildValidSCCBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: securityV1Scheme})),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.sccBuilder.Exists()
		assert.Equal(t, exist, testCase.expectedStatus)
	}
}

func TestSCCCreate(t *testing.T) {
	testCases := []struct {
		testSCC       *Builder
		expectedError error
	}{
		{
			testSCC:       buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testSCC:       buildInValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedError: fmt.Errorf("securityContextConstraints 'selinuxContext' cannot be empty"),
		},
		{
			testSCC: buildValidSCCBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: securityV1Scheme})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		sccBuilder, err := testCase.testSCC.Create()
		assert.Equal(t, err, testCase.expectedError)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.testSCC.Definition.Name, sccBuilder.Object.Name)
		}
	}
}

func TestSCCDelete(t *testing.T) {
	testCases := []struct {
		scc           *Builder
		expectedError error
	}{
		{
			scc:           buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			scc:           buildInValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedError: fmt.Errorf("securityContextConstraints 'selinuxContext' cannot be empty"),
		},
		{
			scc: buildValidSCCBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: securityV1Scheme})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		err := testCase.scc.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.scc.Object)
		}
	}
}

func TestSCCUpdate(t *testing.T) {
	testCases := []struct {
		scc           *Builder
		expectedError error
		users         []string
		force         bool
	}{
		{
			scc:           buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedError: nil,
			users:         []string{"test"},
		},
		{
			scc:           buildInValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedError: fmt.Errorf("securityContextConstraints 'selinuxContext' cannot be empty"),
			users:         []string{"test"},
		},
		{
			scc: buildValidSCCBuilder(
				clients.GetTestClients(clients.TestClientParams{SchemeAttachers: securityV1Scheme})),
			expectedError: fmt.Errorf("failed to update SecurityContextConstraints, object does not exist on cluster"),
			users:         []string{"test"},
		},
	}

	for _, testCase := range testCases {
		assert.Empty(t, testCase.scc.Definition.Users)
		assert.Nil(t, nil, testCase.scc.Object)
		testCase.scc.Definition.Users = testCase.users
		testCase.scc.Definition.ObjectMeta.ResourceVersion = "999"
		_, err := testCase.scc.Update()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.users, testCase.scc.Definition.Users)
		}
	}
}

func TestSCCWithPrivilegedContainer(t *testing.T) {
	testCases := []struct {
		allowPrivileged   bool
		sccBuilder        *Builder
		expectedErrorText string
	}{
		{
			allowPrivileged:   true,
			sccBuilder:        buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "",
		},
		{
			allowPrivileged:   false,
			sccBuilder:        buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "",
		},
		{
			allowPrivileged:   false,
			sccBuilder:        buildInValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "securityContextConstraints 'selinuxContext' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testCase.sccBuilder.WithPrivilegedContainer(testCase.allowPrivileged)
		assert.Equal(t, testCase.sccBuilder.errorMsg, testCase.expectedErrorText)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, testCase.sccBuilder.Definition.AllowPrivilegedContainer, testCase.allowPrivileged)
		}
	}
}

func TestSCCWithPrivilegedEscalation(t *testing.T) {
	testCases := []struct {
		allowEscalation   bool
		sccBuilder        *Builder
		expectedErrorText string
	}{
		{
			allowEscalation:   true,
			sccBuilder:        buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "",
		},
		{
			allowEscalation:   false,
			sccBuilder:        buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "",
		},
		{
			allowEscalation:   false,
			sccBuilder:        buildInValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "securityContextConstraints 'selinuxContext' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testCase.sccBuilder.WithPrivilegedEscalation(testCase.allowEscalation)
		assert.Equal(t, testCase.sccBuilder.errorMsg, testCase.expectedErrorText)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, testCase.sccBuilder.Definition.DefaultAllowPrivilegeEscalation, &testCase.allowEscalation)
		}
	}
}

func TestSCCWithHostDirVolumePlugin(t *testing.T) {
	testCases := []struct {
		allowPlugin       bool
		sccBuilder        *Builder
		expectedErrorText string
	}{
		{
			allowPlugin:       true,
			sccBuilder:        buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "",
		},
		{
			allowPlugin:       false,
			sccBuilder:        buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "",
		},
		{
			allowPlugin:       false,
			sccBuilder:        buildInValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "securityContextConstraints 'selinuxContext' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testCase.sccBuilder.WithHostDirVolumePlugin(testCase.allowPlugin)
		assert.Equal(t, testCase.sccBuilder.errorMsg, testCase.expectedErrorText)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, testCase.sccBuilder.Definition.AllowHostDirVolumePlugin, testCase.allowPlugin)
		}
	}
}

func TestSCCWithHostIPC(t *testing.T) {
	testCases := []struct {
		allowHostIPC      bool
		sccBuilder        *Builder
		expectedErrorText string
	}{
		{
			allowHostIPC:      true,
			sccBuilder:        buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "",
		},
		{
			allowHostIPC:      false,
			sccBuilder:        buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "",
		},
		{
			allowHostIPC:      false,
			sccBuilder:        buildInValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "securityContextConstraints 'selinuxContext' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testCase.sccBuilder.WithHostIPC(testCase.allowHostIPC)
		assert.Equal(t, testCase.sccBuilder.errorMsg, testCase.expectedErrorText)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, testCase.sccBuilder.Definition.AllowHostIPC, testCase.allowHostIPC)
		}
	}
}

func TestSCCWithHostNetwork(t *testing.T) {
	testCases := []struct {
		allowHostNet      bool
		sccBuilder        *Builder
		expectedErrorText string
	}{
		{
			allowHostNet:      true,
			sccBuilder:        buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "",
		},
		{
			allowHostNet:      false,
			sccBuilder:        buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "",
		},
		{
			allowHostNet:      false,
			sccBuilder:        buildInValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "securityContextConstraints 'selinuxContext' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testCase.sccBuilder.WithHostNetwork(testCase.allowHostNet)
		assert.Equal(t, testCase.sccBuilder.errorMsg, testCase.expectedErrorText)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, testCase.sccBuilder.Definition.AllowHostNetwork, testCase.allowHostNet)
		}
	}
}

func TestSCCWithHostPID(t *testing.T) {
	testCases := []struct {
		allowHostPID      bool
		sccBuilder        *Builder
		expectedErrorText string
	}{
		{
			allowHostPID:      true,
			sccBuilder:        buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "",
		},
		{
			allowHostPID:      false,
			sccBuilder:        buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "",
		},
		{
			allowHostPID:      false,
			sccBuilder:        buildInValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "securityContextConstraints 'selinuxContext' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testCase.sccBuilder.WithHostPID(testCase.allowHostPID)
		assert.Equal(t, testCase.sccBuilder.errorMsg, testCase.expectedErrorText)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, testCase.sccBuilder.Definition.AllowHostPID, testCase.allowHostPID)
		}
	}
}

func TestSCCWithHostPorts(t *testing.T) {
	testCases := []struct {
		allowHostPorts    bool
		sccBuilder        *Builder
		expectedErrorText string
	}{
		{
			allowHostPorts:    true,
			sccBuilder:        buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "",
		},
		{
			allowHostPorts:    false,
			sccBuilder:        buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "",
		},
		{
			allowHostPorts:    false,
			sccBuilder:        buildInValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "securityContextConstraints 'selinuxContext' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testCase.sccBuilder.WithHostPorts(testCase.allowHostPorts)
		assert.Equal(t, testCase.sccBuilder.errorMsg, testCase.expectedErrorText)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, testCase.sccBuilder.Definition.AllowHostPorts, testCase.allowHostPorts)
		}
	}
}

func TestSCCWithReadOnlyRootFilesystem(t *testing.T) {
	testCases := []struct {
		readOnlyFS        bool
		sccBuilder        *Builder
		expectedErrorText string
	}{
		{
			readOnlyFS:        true,
			sccBuilder:        buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "",
		},
		{
			readOnlyFS:        false,
			sccBuilder:        buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "",
		},
		{
			readOnlyFS:        false,
			sccBuilder:        buildInValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "securityContextConstraints 'selinuxContext' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testCase.sccBuilder.WithReadOnlyRootFilesystem(testCase.readOnlyFS)
		assert.Equal(t, testCase.sccBuilder.errorMsg, testCase.expectedErrorText)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, testCase.sccBuilder.Definition.ReadOnlyRootFilesystem, testCase.readOnlyFS)
		}
	}
}

func TestSCCWithDropCapabilities(t *testing.T) {
	testCases := []struct {
		dropCapability    []corev1.Capability
		sccBuilder        *Builder
		expectedErrorText string
	}{
		{
			dropCapability:    []corev1.Capability{"NET_RAW", "NET_ADMIN"},
			sccBuilder:        buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "",
		},
		{
			dropCapability:    []corev1.Capability{},
			sccBuilder:        buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "securityContextConstraints 'requiredDropCapabilities' cannot be empty list",
		},
		{
			dropCapability:    []corev1.Capability{"NET_RAW", "NET_ADMIN"},
			sccBuilder:        buildInValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "securityContextConstraints 'selinuxContext' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testCase.sccBuilder.WithDropCapabilities(testCase.dropCapability)
		assert.Equal(t, testCase.sccBuilder.errorMsg, testCase.expectedErrorText)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, testCase.sccBuilder.Definition.RequiredDropCapabilities, testCase.dropCapability)
		}
	}
}

func TestSCCWithAllowCapabilities(t *testing.T) {
	testCases := []struct {
		allowCapability   []corev1.Capability
		sccBuilder        *Builder
		expectedErrorText string
	}{
		{
			allowCapability:   []corev1.Capability{"NET_RAW", "NET_ADMIN"},
			sccBuilder:        buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "",
		},
		{
			allowCapability:   []corev1.Capability{},
			sccBuilder:        buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "securityContextConstraints 'allowCapabilities' cannot be empty list",
		},
		{
			allowCapability:   []corev1.Capability{"NET_RAW", "NET_ADMIN"},
			sccBuilder:        buildInValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "securityContextConstraints 'selinuxContext' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testCase.sccBuilder.WithAllowCapabilities(testCase.allowCapability)
		assert.Equal(t, testCase.sccBuilder.errorMsg, testCase.expectedErrorText)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, testCase.sccBuilder.Definition.AllowedCapabilities, testCase.allowCapability)
		}
	}
}

func TestSCCWithDefaultAddCapabilities(t *testing.T) {
	testCases := []struct {
		defaultAddCapability []corev1.Capability
		sccBuilder           *Builder
		expectedErrorText    string
	}{
		{
			defaultAddCapability: []corev1.Capability{"NET_RAW", "NET_ADMIN"},
			sccBuilder:           buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText:    "",
		},
		{
			defaultAddCapability: []corev1.Capability{},
			sccBuilder:           buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText:    "securityContextConstraints 'defaultAddCapabilities' cannot be empty list",
		},
		{
			defaultAddCapability: []corev1.Capability{"NET_RAW", "NET_ADMIN"},
			sccBuilder:           buildInValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText:    "securityContextConstraints 'selinuxContext' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testCase.sccBuilder.WithDefaultAddCapabilities(testCase.defaultAddCapability)
		assert.Equal(t, testCase.sccBuilder.errorMsg, testCase.expectedErrorText)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, testCase.sccBuilder.Definition.DefaultAddCapabilities, testCase.defaultAddCapability)
		}
	}
}

func TestSCCWithPriority(t *testing.T) {
	testCases := []struct {
		priority          int32
		sccBuilder        *Builder
		expectedErrorText string
	}{
		{
			priority:          100,
			sccBuilder:        buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "",
		},
		{
			priority:          100,
			sccBuilder:        buildInValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "securityContextConstraints 'selinuxContext' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testCase.sccBuilder.WithPriority(&testCase.priority)
		assert.Equal(t, testCase.sccBuilder.errorMsg, testCase.expectedErrorText)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, testCase.sccBuilder.Definition.Priority, &testCase.priority)
		}
	}
}

func TestSCCWithFSGroup(t *testing.T) {
	testCases := []struct {
		fsGroup           string
		sccBuilder        *Builder
		expectedErrorText string
	}{
		{
			fsGroup:           "default",
			sccBuilder:        buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "",
		},
		{
			fsGroup:           "",
			sccBuilder:        buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "securityContextConstraints 'fsGroup' cannot be empty string",
		},
		{
			fsGroup:           "default",
			sccBuilder:        buildInValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "securityContextConstraints 'selinuxContext' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testCase.sccBuilder.WithFSGroup(testCase.fsGroup)
		assert.Equal(t, testCase.sccBuilder.errorMsg, testCase.expectedErrorText)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, testCase.sccBuilder.Definition.FSGroup.Type, securityV1.FSGroupStrategyType(testCase.fsGroup))
		}
	}
}

func TestSCCWithFSGroupRange(t *testing.T) {
	testCases := []struct {
		fsGroupMin        int64
		fsGroupMax        int64
		sccBuilder        *Builder
		expectedErrorText string
	}{
		{
			fsGroupMin:        1001,
			fsGroupMax:        1005,
			sccBuilder:        buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "",
		},
		{
			fsGroupMin:        1005,
			fsGroupMax:        1001,
			sccBuilder:        buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "securityContextConstraints 'fsGroupMin' argument can not be greater than fsGroupMax",
		},
		{
			fsGroupMin:        1005,
			fsGroupMax:        1001,
			sccBuilder:        buildInValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "securityContextConstraints 'selinuxContext' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testCase.sccBuilder.WithFSGroupRange(testCase.fsGroupMin, testCase.fsGroupMax)
		assert.Equal(t, testCase.sccBuilder.errorMsg, testCase.expectedErrorText)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, testCase.sccBuilder.Definition.FSGroup.Ranges,
				[]securityV1.IDRange{{Min: testCase.fsGroupMin, Max: testCase.fsGroupMax}})
		}
	}
}

func TestSCCWithGroups(t *testing.T) {
	testCases := []struct {
		groups            []string
		sccBuilder        *Builder
		expectedErrorText string
	}{
		{
			groups:            []string{"1001", "1002"},
			sccBuilder:        buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "",
		},
		{
			groups:            []string{},
			sccBuilder:        buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "securityContextConstraints 'fsGroupType' cannot be empty string",
		},
		{
			groups:            []string{"1001", "1002"},
			sccBuilder:        buildInValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "securityContextConstraints 'selinuxContext' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testCase.sccBuilder.WithGroups(testCase.groups)
		assert.Equal(t, testCase.sccBuilder.errorMsg, testCase.expectedErrorText)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, testCase.sccBuilder.Definition.Groups, testCase.groups)
		}
	}
}

func TestSCCWithSeccompProfiles(t *testing.T) {
	testCases := []struct {
		seccompProfiles   []string
		sccBuilder        *Builder
		expectedErrorText string
	}{
		{
			seccompProfiles:   []string{"test1", "test2"},
			sccBuilder:        buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "",
		},
		{
			seccompProfiles:   []string{},
			sccBuilder:        buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "securityContextConstraints 'seccompProfiles' cannot be empty list",
		},
		{
			seccompProfiles:   []string{"test1", "test2"},
			sccBuilder:        buildInValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "securityContextConstraints 'selinuxContext' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testCase.sccBuilder.WithSeccompProfiles(testCase.seccompProfiles)
		assert.Equal(t, testCase.sccBuilder.errorMsg, testCase.expectedErrorText)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, testCase.sccBuilder.Definition.SeccompProfiles, testCase.seccompProfiles)
		}
	}
}

func TestSCCWithSupplementalGroups(t *testing.T) {
	testCases := []struct {
		supplementalGroupsType string
		sccBuilder             *Builder
		expectedErrorText      string
	}{
		{
			supplementalGroupsType: "test",
			sccBuilder:             buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText:      "",
		},
		{
			supplementalGroupsType: "",
			sccBuilder:             buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText:      "securityContextConstraints 'SupplementalGroups' cannot be empty string",
		},
		{
			supplementalGroupsType: "test",
			sccBuilder:             buildInValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText:      "securityContextConstraints 'selinuxContext' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testCase.sccBuilder.WithSupplementalGroups(testCase.supplementalGroupsType)
		assert.Equal(t, testCase.sccBuilder.errorMsg, testCase.expectedErrorText)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, testCase.sccBuilder.Definition.SupplementalGroups.Type,
				securityV1.SupplementalGroupsStrategyType(testCase.supplementalGroupsType))
		}
	}
}

func TestSCCWithUsers(t *testing.T) {
	testCases := []struct {
		users             []string
		sccBuilder        *Builder
		expectedErrorText string
	}{
		{
			users:             []string{"1001", "1002"},
			sccBuilder:        buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "",
		},
		{
			users:             []string{},
			sccBuilder:        buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "securityContextConstraints 'users' cannot be empty list",
		},
		{
			users:             []string{"1001", "1002"},
			sccBuilder:        buildInValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "securityContextConstraints 'selinuxContext' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testCase.sccBuilder.WithUsers(testCase.users)
		assert.Equal(t, testCase.sccBuilder.errorMsg, testCase.expectedErrorText)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, testCase.sccBuilder.Definition.Users, testCase.users)
		}
	}
}

func TestSCCWithVolumes(t *testing.T) {
	testCases := []struct {
		volumes           []securityV1.FSType
		sccBuilder        *Builder
		expectedErrorText string
	}{
		{
			volumes:           []securityV1.FSType{"test1", "test2"},
			sccBuilder:        buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "",
		},
		{
			volumes:           []securityV1.FSType{},
			sccBuilder:        buildValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "securityContextConstraints 'volumes' cannot be empty list",
		},
		{
			volumes:           []securityV1.FSType{"test1", "test2"},
			sccBuilder:        buildInValidSCCBuilder(buildTestClientWithDummyObject()),
			expectedErrorText: "securityContextConstraints 'selinuxContext' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testCase.sccBuilder.WithVolumes(testCase.volumes)
		assert.Equal(t, testCase.sccBuilder.errorMsg, testCase.expectedErrorText)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, testCase.sccBuilder.Definition.Volumes, testCase.volumes)
		}
	}
}

func buildValidSCCBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, "testscc", "user", "default")
}

func buildInValidSCCBuilder(apiClient *clients.Settings) *Builder {
	return NewBuilder(apiClient, "testscc", "user", "")
}

func buildTestClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyScc(),
		SchemeAttachers: securityV1Scheme,
	})
}

func buildDummyScc() []runtime.Object {
	return append([]runtime.Object{}, &securityV1.SecurityContextConstraints{
		ObjectMeta: metav1.ObjectMeta{
			Name: "testscc",
		},
	})
}
