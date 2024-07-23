package ibi

import (
	"fmt"
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	ibiv1alpha1 "github.com/openshift-kni/eco-goinfra/pkg/schemes/imagebasedinstall/api/hiveextensions/v1alpha1"
	hivev1 "github.com/openshift-kni/eco-goinfra/pkg/schemes/imagebasedinstall/hive/api/v1"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
)

const (
	testImageClusterInstall = "test-image-cluster-install"
)

var testSchemes = []clients.SchemeAttacher{
	ibiv1alpha1.AddToScheme,
}

func TestNewImageClusterInstallBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		imageset      string
		client        bool
		expectedError string
	}{
		{
			name:          testImageClusterInstall,
			namespace:     testImageClusterInstall,
			imageset:      "4.16",
			client:        true,
			expectedError: "",
		},
		{
			name:          "",
			namespace:     testImageClusterInstall,
			imageset:      "4.16",
			client:        true,
			expectedError: "imageclusterinstall 'name' cannot be empty",
		},
		{
			name:          testImageClusterInstall,
			namespace:     "",
			imageset:      "4.16",
			client:        true,
			expectedError: "imageclusterinstall 'nsname' cannot be empty",
		},
		{
			name:          testImageClusterInstall,
			namespace:     testImageClusterInstall,
			imageset:      "",
			client:        true,
			expectedError: "imageclusterinstall 'imageset' cannot be empty",
		},
		{
			name:          testImageClusterInstall,
			namespace:     testImageClusterInstall,
			imageset:      "4.16",
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

		testBuilder := NewImageClusterInstallBuilder(
			client, testCase.name, testCase.namespace, testCase.imageset)

		if testCase.client {
			assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

			if testCase.expectedError == "" {
				assert.Equal(t, testCase.name, testBuilder.Definition.Name)
				assert.Equal(t, testCase.namespace, testBuilder.Definition.Namespace)
				assert.Equal(t, testCase.imageset, testBuilder.Definition.Spec.ImageSetRef.Name)
			}
		} else {
			assert.Nil(t, testBuilder)
		}
	}
}

func TestImageClusterInstallPull(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		client        bool
		exists        bool
		expectedError error
	}{
		{
			name:          testImageClusterInstall,
			namespace:     testImageClusterInstall,
			client:        true,
			exists:        true,
			expectedError: nil,
		},
		{
			name:          "",
			namespace:     testImageClusterInstall,
			client:        true,
			exists:        true,
			expectedError: fmt.Errorf("imageclusterinstall 'name' cannot be empty"),
		},
		{
			name:          testImageClusterInstall,
			namespace:     "",
			client:        true,
			exists:        true,
			expectedError: fmt.Errorf("imageclusterinstall 'nsname' cannot be empty"),
		},
		{
			name:          testImageClusterInstall,
			namespace:     testImageClusterInstall,
			client:        false,
			exists:        true,
			expectedError: fmt.Errorf("apiClient cannot be nil"),
		},
		{
			name:      testImageClusterInstall,
			namespace: testImageClusterInstall,
			client:    true,
			exists:    false,
			expectedError: fmt.Errorf(
				"imageclusterinstall object %s does not exist in namespace %s",
				testImageClusterInstall, testImageClusterInstall),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testClient     *clients.Settings
		)

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateImageClusterInstall())
		}

		if testCase.client {
			testClient = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: testSchemes,
			})
		}

		testBuilder, err := PullImageClusterInstall(testClient, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, testBuilder.Definition.Name)
			assert.Equal(t, testCase.namespace, testBuilder.Definition.Namespace)
		}
	}
}

func TestImageClusterInstallWithHostname(t *testing.T) {
	testCases := []struct {
		hostname         string
		expectedErrorMsg string
	}{
		{
			hostname:         "sno-0-0",
			expectedErrorMsg: "",
		},
		{
			hostname:         "",
			expectedErrorMsg: "imageclusterinstall hostname cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateImageClusterInstallBuilder()

		testBuilder.WithHostname(testCase.hostname)
		assert.Equal(t, testCase.expectedErrorMsg, testBuilder.errorMsg)

		if testCase.expectedErrorMsg == "" {
			assert.Equal(t, testCase.hostname, testBuilder.Definition.Spec.Hostname)
		}
	}
}

func TestImageClusterInstallWithClusterDeployment(t *testing.T) {
	testCases := []struct {
		clusterdeployment string
		expectedErrorMsg  string
	}{
		{
			clusterdeployment: "ibi-cluster-deployment",
			expectedErrorMsg:  "",
		},
		{
			clusterdeployment: "",
			expectedErrorMsg:  "imageclusterinstall clusterdeployment cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateImageClusterInstallBuilder()

		testBuilder.WithClusterDeployment(testCase.clusterdeployment)
		assert.Equal(t, testCase.expectedErrorMsg, testBuilder.errorMsg)

		if testCase.expectedErrorMsg == "" {
			assert.Equal(t, testCase.clusterdeployment, testBuilder.Definition.Spec.ClusterDeploymentRef.Name)
		}
	}
}
func TestImageClusterInstallWithExtraManifests(t *testing.T) {
	testCases := []struct {
		extramanifest    string
		expectedErrorMsg string
	}{
		{
			extramanifest:    "ibi-extra-manifest",
			expectedErrorMsg: "",
		},
		{
			extramanifest:    "",
			expectedErrorMsg: "imageclusterinstall extramanifest cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateImageClusterInstallBuilder()

		testBuilder.WithExtraManifests(testCase.extramanifest)
		assert.Equal(t, testCase.expectedErrorMsg, testBuilder.errorMsg)

		if testCase.expectedErrorMsg == "" {
			assert.Equal(t, testCase.extramanifest, testBuilder.Definition.Spec.ExtraManifestsRefs[0].Name)
		}
	}
}
func TestImageClusterInstallWithMachineNetwork(t *testing.T) {
	testCases := []struct {
		machineNetwork   string
		expectedErrorMsg string
	}{
		{
			machineNetwork:   "192.168.0.0/24",
			expectedErrorMsg: "",
		},
		{
			machineNetwork:   "fd2e:6f44:5dd8::/64",
			expectedErrorMsg: "",
		},
		{
			machineNetwork:   "192.168.0.0/255.255.255.0",
			expectedErrorMsg: "imageclusterinstall machinenetwork incorrectly formatted",
		},
		{
			machineNetwork:   "fd2e:6f44:5dd8::",
			expectedErrorMsg: "imageclusterinstall machinenetwork incorrectly formatted",
		},
		{
			machineNetwork:   "",
			expectedErrorMsg: "imageclusterinstall machinenetwork incorrectly formatted",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateImageClusterInstallBuilder()

		testBuilder.WithMachineNetwork(testCase.machineNetwork)
		assert.Equal(t, testCase.expectedErrorMsg, testBuilder.errorMsg)

		if testCase.expectedErrorMsg == "" {
			assert.Equal(t, testCase.machineNetwork, testBuilder.Definition.Spec.MachineNetwork)
		}
	}
}

func TestImageClusterInstallWithSSHKey(t *testing.T) {
	testCases := []struct {
		sshkey           string
		expectedErrorMsg string
	}{
		{
			sshkey:           "mysecretsshkey",
			expectedErrorMsg: "",
		},
		{
			sshkey:           "",
			expectedErrorMsg: "imageclusterinstall sshkey cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateImageClusterInstallBuilder()

		testBuilder.WithSSHKey(testCase.sshkey)
		assert.Equal(t, testCase.expectedErrorMsg, testBuilder.errorMsg)

		if testCase.expectedErrorMsg == "" {
			assert.Equal(t, testCase.sshkey, testBuilder.Definition.Spec.SSHKey)
		}
	}
}
func TestImageClusterInstallWithVersion(t *testing.T) {
	testCases := []struct {
		version          string
		expectedErrorMsg string
	}{
		{
			version:          "4.16",
			expectedErrorMsg: "",
		},
		{
			version:          "",
			expectedErrorMsg: "imageclusterinstall version cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateImageClusterInstallBuilder()

		testBuilder.WithVersion(testCase.version)
		assert.Equal(t, testCase.expectedErrorMsg, testBuilder.errorMsg)

		if testCase.expectedErrorMsg == "" {
			assert.Equal(t, testCase.version, testBuilder.Definition.Spec.Version)
		}
	}
}
func TestImageClusterInstallGetCompletedCondition(t *testing.T) {
	testCases := []struct {
		status                   ibiv1alpha1.ImageClusterInstallStatus
		expectedConditionMessage string
		expectedError            error
	}{
		{
			status: ibiv1alpha1.ImageClusterInstallStatus{
				Conditions: []hivev1.ClusterInstallCondition{
					{
						Type:    hivev1.ClusterInstallCompleted,
						Message: "This is a test completed condition",
					},
				},
			},
			expectedConditionMessage: "This is a test completed condition",
			expectedError:            nil,
		},
		{
			status: ibiv1alpha1.ImageClusterInstallStatus{
				Conditions: []hivev1.ClusterInstallCondition{},
			},
			expectedConditionMessage: "",
			expectedError:            fmt.Errorf("cannot find Completed condition in imageclusterinstall status"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		testICI := generateImageClusterInstall()
		testICI.Status = testCase.status
		runtimeObjects = append(runtimeObjects, testICI)

		testBuilder := generateImageClusterInstallBuilderWithFakeObjects(runtimeObjects)

		condition, err := testBuilder.GetCompletedCondition()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.expectedConditionMessage, condition.Message)
		}
	}
}
func TestImageClusterInstallGetFailedCondition(t *testing.T) {
	testCases := []struct {
		status                   ibiv1alpha1.ImageClusterInstallStatus
		expectedConditionMessage string
		expectedError            error
	}{
		{
			status: ibiv1alpha1.ImageClusterInstallStatus{
				Conditions: []hivev1.ClusterInstallCondition{
					{
						Type:    hivev1.ClusterInstallFailed,
						Message: "This is a test failed condition",
					},
				},
			},
			expectedConditionMessage: "This is a test failed condition",
			expectedError:            nil,
		},
		{
			status: ibiv1alpha1.ImageClusterInstallStatus{
				Conditions: []hivev1.ClusterInstallCondition{},
			},
			expectedConditionMessage: "",
			expectedError:            fmt.Errorf("cannot find Failed condition in imageclusterinstall status"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		testICI := generateImageClusterInstall()
		testICI.Status = testCase.status
		runtimeObjects = append(runtimeObjects, testICI)

		testBuilder := generateImageClusterInstallBuilderWithFakeObjects(runtimeObjects)

		condition, err := testBuilder.GetFailedCondition()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.expectedConditionMessage, condition.Message)
		}
	}
}
func TestImageClusterInstallGetRequirementsMetCondition(t *testing.T) {
	testCases := []struct {
		status                   ibiv1alpha1.ImageClusterInstallStatus
		expectedConditionMessage string
		expectedError            error
	}{
		{
			status: ibiv1alpha1.ImageClusterInstallStatus{
				Conditions: []hivev1.ClusterInstallCondition{
					{
						Type:    hivev1.ClusterInstallRequirementsMet,
						Message: "This is a test requirements met condition",
					},
				},
			},
			expectedConditionMessage: "This is a test requirements met condition",
			expectedError:            nil,
		},
		{
			status: ibiv1alpha1.ImageClusterInstallStatus{
				Conditions: []hivev1.ClusterInstallCondition{},
			},
			expectedConditionMessage: "",
			expectedError:            fmt.Errorf("cannot find RequirementsMet condition in imageclusterinstall status"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		testICI := generateImageClusterInstall()
		testICI.Status = testCase.status
		runtimeObjects = append(runtimeObjects, testICI)

		testBuilder := generateImageClusterInstallBuilderWithFakeObjects(runtimeObjects)

		condition, err := testBuilder.GetRequirementsMetCondition()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.expectedConditionMessage, condition.Message)
		}
	}
}
func TestImageClusterInstallGetStoppedCondition(t *testing.T) {
	testCases := []struct {
		status                   ibiv1alpha1.ImageClusterInstallStatus
		expectedConditionMessage string
		expectedError            error
	}{
		{
			status: ibiv1alpha1.ImageClusterInstallStatus{
				Conditions: []hivev1.ClusterInstallCondition{
					{
						Type:    hivev1.ClusterInstallStopped,
						Message: "This is a test stopped condition",
					},
				},
			},
			expectedConditionMessage: "This is a test stopped condition",
			expectedError:            nil,
		},
		{
			status: ibiv1alpha1.ImageClusterInstallStatus{
				Conditions: []hivev1.ClusterInstallCondition{},
			},
			expectedConditionMessage: "",
			expectedError:            fmt.Errorf("cannot find Stopped condition in imageclusterinstall status"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		testICI := generateImageClusterInstall()
		testICI.Status = testCase.status
		runtimeObjects = append(runtimeObjects, testICI)

		testBuilder := generateImageClusterInstallBuilderWithFakeObjects(runtimeObjects)

		condition, err := testBuilder.GetStoppedCondition()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.expectedConditionMessage, condition.Message)
		}
	}
}
func TestImageClusterInstallGet(t *testing.T) {
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
			runtimeObjects = append(runtimeObjects, generateImageClusterInstall())
		}

		testBuilder := generateImageClusterInstallBuilderWithFakeObjects(runtimeObjects)

		aci, err := testBuilder.Get()
		if testCase.exists {
			assert.Nil(t, err)
			assert.NotNil(t, aci)
		} else {
			assert.NotNil(t, err)
			assert.Nil(t, aci)
		}
	}
}

func TestImageClusterInstallCreate(t *testing.T) {
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
			runtimeObjects = append(runtimeObjects, generateImageClusterInstall())
		}

		testBuilder := generateImageClusterInstallBuilderWithFakeObjects(runtimeObjects)

		result, err := testBuilder.Create()
		assert.Nil(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, testImageClusterInstall, result.Definition.Name)
		assert.Equal(t, testImageClusterInstall, result.Definition.Namespace)
	}
}
func TestImageClusterInstallUpdate(t *testing.T) {
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
			expectedError: fmt.Errorf("cannot update non-existent imageclusterinstall"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateImageClusterInstall())
		}

		testBuilder := generateImageClusterInstallBuilderWithFakeObjects(runtimeObjects)

		testBuilder.Definition.Spec.Hostname = "test-hostname"

		ici, err := testBuilder.Update(true)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, ici.Object.Spec.Hostname, "test-hostname")
		}
	}
}
func TestImageClusterInstallDelete(t *testing.T) {
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
			expectedError: fmt.Errorf("imageclusterinstall cannot be deleted because it does not exist"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateImageClusterInstall())
		}

		testBuilder := generateImageClusterInstallBuilderWithFakeObjects(runtimeObjects)

		err := testBuilder.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testBuilder.Object)
		}
	}
}
func TestImageClusterInstallExists(t *testing.T) {
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
			runtimeObjects = append(runtimeObjects, generateImageClusterInstall())
		}

		testBuilder := generateImageClusterInstallBuilderWithFakeObjects(runtimeObjects)

		assert.Equal(t, testCase.exists, testBuilder.Exists())
	}
}

func TestImageClusterInstallValidate(t *testing.T) {
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
			expectedError: "error: received nil ImageClusterInstall builder",
		},
		{
			builderNil:    false,
			definitionNil: true,
			apiClientNil:  false,
			expectedError: "can not redefine the undefined ImageClusterInstall",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  true,
			expectedError: "ImageClusterInstall builder cannot have nil apiClient",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateImageClusterInstallBuilderWithFakeObjects([]runtime.Object{})

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

func generateImageClusterInstallBuilderWithFakeObjects(objects []runtime.Object) *ImageClusterInstallBuilder {
	return &ImageClusterInstallBuilder{
		apiClient: clients.GetTestClients(
			clients.TestClientParams{K8sMockObjects: objects, SchemeAttachers: testSchemes}).Client,
		Definition: generateImageClusterInstall(),
	}
}

func generateImageClusterInstallBuilder() *ImageClusterInstallBuilder {
	return &ImageClusterInstallBuilder{
		apiClient:  clients.GetTestClients(clients.TestClientParams{}).Client,
		Definition: generateImageClusterInstall(),
	}
}

func generateImageClusterInstall() *ibiv1alpha1.ImageClusterInstall {
	return &ibiv1alpha1.ImageClusterInstall{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testImageClusterInstall,
			Namespace: testImageClusterInstall,
		},
		Spec: ibiv1alpha1.ImageClusterInstallSpec{
			ImageSetRef: hivev1.ClusterImageSetReference{
				Name: "4.16",
			},
		},
	}
}
