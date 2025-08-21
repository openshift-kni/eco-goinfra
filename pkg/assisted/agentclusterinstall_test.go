package assisted

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	hiveextV1Beta1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/assisted/api/hiveextension/v1beta1"
	hivev1 "github.com/rh-ecosystem-edge/eco-goinfra/pkg/schemes/assisted/hive/api/v1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	aciTestName      = "aci-test-name"
	aciTestNamespace = "aci-test-namespace"
)

var testSchemes = []clients.SchemeAttacher{
	hiveextV1Beta1.AddToScheme,
}

func TestNewAgentClusterInstallBuilder(t *testing.T) {
	testCases := []struct {
		name              string
		namespace         string
		clusterDeployment string
		masterCount       int
		workerCount       int
		network           hiveextV1Beta1.Networking
		client            bool
		expectedError     string
	}{
		{
			name:              aciTestName,
			namespace:         aciTestNamespace,
			clusterDeployment: "aci-test-clusterdeployment",
			masterCount:       3,
			workerCount:       2,
			network:           dummyTestNetwork(),
			client:            true,
			expectedError:     "",
		},
		{
			name:              "",
			namespace:         aciTestNamespace,
			clusterDeployment: "aci-test-clusterdeployment",
			masterCount:       3,
			workerCount:       2,
			network:           dummyTestNetwork(),
			client:            true,
			expectedError:     "agentclusterinstall 'name' cannot be empty",
		},
		{
			name:              aciTestName,
			namespace:         "",
			clusterDeployment: "aci-test-clusterdeployment",
			masterCount:       3,
			workerCount:       2,
			network:           dummyTestNetwork(),
			client:            true,
			expectedError:     "agentclusterinstall 'namespace' cannot be empty",
		},
		{
			name:              aciTestName,
			namespace:         aciTestNamespace,
			clusterDeployment: "",
			masterCount:       3,
			workerCount:       2,
			network:           dummyTestNetwork(),
			client:            true,
			expectedError:     "agentclusterinstall 'clusterDeployment' cannot be empty",
		},
		{
			name:              aciTestName,
			namespace:         aciTestNamespace,
			clusterDeployment: "aci-test-clusterdeployment",
			masterCount:       3,
			workerCount:       2,
			network:           dummyTestNetwork(),
			client:            false,
			expectedError:     "",
		},
	}

	for _, testcase := range testCases {
		var testSettings *clients.Settings
		if testcase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		testBuilder := NewAgentClusterInstallBuilder(
			testSettings, testcase.name, testcase.namespace, testcase.clusterDeployment,
			testcase.masterCount, testcase.workerCount, testcase.network)

		if testcase.client {
			assert.NotNil(t, testBuilder)
			assert.Equal(t, testcase.expectedError, testBuilder.errorMsg)
			assert.Equal(t, testcase.name, testBuilder.Definition.Name)
			assert.Equal(t, testcase.namespace, testBuilder.Definition.Namespace)
			assert.Equal(t, testcase.clusterDeployment, testBuilder.Definition.Spec.ClusterDeploymentRef.Name)
			assert.Equal(t, testcase.masterCount, testBuilder.Definition.Spec.ProvisionRequirements.ControlPlaneAgents)
			assert.Equal(t, testcase.workerCount, testBuilder.Definition.Spec.ProvisionRequirements.WorkerAgents)
			assert.Equal(t, testcase.network, testBuilder.Definition.Spec.Networking)
		} else {
			assert.Nil(t, testBuilder)
		}
	}
}

func TestPullAgentClusterInstall(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		client        bool
		exists        bool
		expectedError error
	}{
		{
			name:          aciTestName,
			namespace:     aciTestNamespace,
			client:        true,
			exists:        true,
			expectedError: nil,
		},
		{
			name:          "",
			namespace:     aciTestNamespace,
			client:        true,
			exists:        true,
			expectedError: errors.New("agentclusterinstall 'name' cannot be empty"),
		},
		{
			name:          aciTestName,
			namespace:     "",
			client:        true,
			exists:        true,
			expectedError: errors.New("agentclusterinstall 'namespace' cannot be empty"),
		},
		{
			name:          aciTestName,
			namespace:     aciTestNamespace,
			client:        false,
			exists:        true,
			expectedError: errors.New("the apiClient is nil"),
		},
		{
			name:          aciTestName,
			namespace:     aciTestNamespace,
			client:        true,
			exists:        false,
			expectedError: errors.New("agentclusterinstall object aci-test-name does not exist in namespace aci-test-namespace"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
			testSettings   *clients.Settings
		)

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateAgentClusterInstall())
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects, SchemeAttachers: testSchemes})
		}

		testBuilder, err := PullAgentClusterInstall(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

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

func TestAgentClusterInstallCreate(t *testing.T) {
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
			runtimeObjects = append(runtimeObjects, generateAgentClusterInstall())
		}

		testBuilder := buildTestBuilderWithFakeObjects(runtimeObjects)

		result, err := testBuilder.Create()
		assert.Nil(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, aciTestName, result.Definition.Name)
		assert.Equal(t, aciTestNamespace, result.Definition.Namespace)
	}
}

func TestAgentClusterInstallDelete(t *testing.T) {
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
			runtimeObjects = append(runtimeObjects, generateAgentClusterInstall())
		}

		testBuilder := buildTestBuilderWithFakeObjects(runtimeObjects)

		err := testBuilder.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testBuilder.Object)
		}
	}
}

func TestAgentClusterInstallDeleteAndWait(t *testing.T) {
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
			runtimeObjects = append(runtimeObjects, generateAgentClusterInstall())
		}

		testBuilder := buildTestBuilderWithFakeObjects(runtimeObjects)

		err := testBuilder.DeleteAndWait(time.Second * 1)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testBuilder.Object)
		}
	}
}

func TestAgentClusterInstallGet(t *testing.T) {
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
			runtimeObjects = append(runtimeObjects, generateAgentClusterInstall())
		}

		testBuilder := buildTestBuilderWithFakeObjects(runtimeObjects)

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

func TestAgentClusterInstallExists(t *testing.T) {
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
			runtimeObjects = append(runtimeObjects, generateAgentClusterInstall())
		}

		testBuilder := buildTestBuilderWithFakeObjects(runtimeObjects)

		assert.Equal(t, testCase.exists, testBuilder.Exists())
	}
}

func TestAgentClusterInstallUpdate(t *testing.T) {
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
			expectedError: errors.New("cannot update non-existent agentclusterinstall"),
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		if testCase.exists {
			runtimeObjects = append(runtimeObjects, generateAgentClusterInstall())
		}

		testBuilder := buildTestBuilderWithFakeObjects(runtimeObjects)

		testBuilder.Definition.Spec.ProvisionRequirements.WorkerAgents = 5

		aci, err := testBuilder.Update(true)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, aci.Object.Spec.ProvisionRequirements.WorkerAgents, 5)
		}
	}
}

func TestAgentClusterInstallWithOptions(t *testing.T) {
	testCases := []struct {
		option        AgentClusterInstallAdditionalOptions
		validator     func(*AgentClusterInstallBuilder) bool
		expectedError string
	}{
		{
			option: func(builder *AgentClusterInstallBuilder) (*AgentClusterInstallBuilder, error) {
				builder.Definition.Spec.HoldInstallation = true

				return builder, nil
			},
			validator: func(acib *AgentClusterInstallBuilder) bool {
				return acib.Definition.Spec.HoldInstallation == true
			},
			expectedError: "",
		},
		{
			option: func(builder *AgentClusterInstallBuilder) (*AgentClusterInstallBuilder, error) {
				builder.Definition.Spec.IgnitionEndpoint = &hiveextV1Beta1.IgnitionEndpoint{
					Url: "https://some.ign.endpoint.com",
				}

				return builder, nil
			},
			validator: func(acib *AgentClusterInstallBuilder) bool {

				return acib.Definition.Spec.IgnitionEndpoint.Url == "https://some.ign.endpoint.com"
			},
			expectedError: "",
		},
		{
			option: func(builder *AgentClusterInstallBuilder) (*AgentClusterInstallBuilder, error) {

				return builder, fmt.Errorf("agentclusterinstallbuilder contains error")
			},
			validator: func(acib *AgentClusterInstallBuilder) bool {

				return acib.errorMsg != ""
			},
			expectedError: "agentclusterinstallbuilder contains error",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateAgentClusterInstallTestBuilder()

		testBuilder.WithOptions(testCase.option)
		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)
		assert.True(t, testCase.validator(testBuilder))
	}
}

func TestAgentClusterInstallWithApiVip(t *testing.T) {
	testCases := []struct {
		ip          string
		expectError string
	}{
		{
			ip:          "192.168.1.5",
			expectError: "",
		},
		{
			ip:          "fd2e::5",
			expectError: "",
		},
		{
			ip:          "notanip",
			expectError: "agentclusterinstall apiVIP incorrectly formatted",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateAgentClusterInstallTestBuilder()
		testBuilder.WithAPIVip(testCase.ip)
		assert.Equal(t, testCase.expectError, testBuilder.errorMsg)

		if testCase.expectError == "" {
			assert.Equal(t, testCase.ip, testBuilder.Definition.Spec.APIVIP)
		}
	}
}
func TestAgentClusterInstallWithAdditionalApiVip(t *testing.T) {
	testCases := []struct {
		ip          string
		expectError string
	}{
		{
			ip:          "192.168.1.5",
			expectError: "",
		},
		{
			ip:          "fd2e::5",
			expectError: "",
		},
		{
			ip:          "notanip",
			expectError: "agentclusterinstall apiVIP incorrectly formatted",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateAgentClusterInstallTestBuilder()
		testBuilder.WithAdditionalAPIVip(testCase.ip)
		assert.Equal(t, testCase.expectError, testBuilder.errorMsg)

		if testCase.expectError == "" {
			assert.Contains(t, testBuilder.Definition.Spec.APIVIPs, testCase.ip)
		}
	}
}

func TestAgentClusterInstallWithIngressVip(t *testing.T) {
	testCases := []struct {
		ip          string
		expectError string
	}{
		{
			ip:          "192.168.1.10",
			expectError: "",
		},
		{
			ip:          "fd2e::10",
			expectError: "",
		},
		{
			ip:          "notanip",
			expectError: "agentclusterinstall ingressVIP incorrectly formatted",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateAgentClusterInstallTestBuilder()
		testBuilder.WithIngressVip(testCase.ip)
		assert.Equal(t, testCase.expectError, testBuilder.errorMsg)

		if testCase.expectError == "" {
			assert.Equal(t, testCase.ip, testBuilder.Definition.Spec.IngressVIP)
		}
	}
}
func TestAgentClusterInstallWithAdditionalIngressVip(t *testing.T) {
	testCases := []struct {
		ip          string
		expectError string
	}{
		{
			ip:          "192.168.1.10",
			expectError: "",
		},
		{
			ip:          "fd2e::10",
			expectError: "",
		},
		{
			ip:          "notanip",
			expectError: "agentclusterinstall ingressVIP incorrectly formatted",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateAgentClusterInstallTestBuilder()
		testBuilder.WithAdditionalIngressVip(testCase.ip)
		assert.Equal(t, testCase.expectError, testBuilder.errorMsg)

		if testCase.expectError == "" {
			assert.Contains(t, testBuilder.Definition.Spec.IngressVIPs, testCase.ip)
		}
	}
}
func TestAgentClusterInstallWithUserManagedNetworking(t *testing.T) {
	testCases := []struct {
		umn bool
	}{
		{
			umn: true,
		},
		{
			umn: false,
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateAgentClusterInstallTestBuilder()
		testBuilder.WithUserManagedNetworking(testCase.umn)

		assert.Equal(t, testCase.umn, *testBuilder.Definition.Spec.Networking.UserManagedNetworking)
	}
}
func TestAgentClusterInstallWithPlatformType(t *testing.T) {
	testCases := []struct {
		platform hiveextV1Beta1.PlatformType
	}{
		{
			platform: "BareMetal",
		},
		{
			platform: "None",
		},
		{
			platform: "VSphere",
		},
		{
			platform: "Nutanix",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateAgentClusterInstallTestBuilder()
		testBuilder.WithPlatformType(testCase.platform)

		assert.Equal(t, testCase.platform, testBuilder.Definition.Spec.PlatformType)
	}
}
func TestAgentClusterInstallWithControlPlaneAgents(t *testing.T) {
	testCases := []struct {
		count         int
		expectedError string
	}{
		{
			count:         3,
			expectedError: "",
		},
		{
			count:         1,
			expectedError: "",
		},
		{
			count:         0,
			expectedError: "agentclusterinstall controlplane agents must be greater than 0",
		},
		{
			count:         -1,
			expectedError: "agentclusterinstall controlplane agents must be greater than 0",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateAgentClusterInstallTestBuilder()
		testBuilder.WithControlPlaneAgents(testCase.count)

		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)
		assert.Equal(t, testCase.count, testBuilder.Definition.Spec.ProvisionRequirements.ControlPlaneAgents)
	}
}

func TestAgentClusterInstallWithImageSet(t *testing.T) {
	testCases := []struct {
		image         string
		expectedError string
	}{
		{
			image:         "4.16",
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateAgentClusterInstallTestBuilder()
		testBuilder.WithImageSet(testCase.image)

		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)
		assert.Equal(t, testCase.image, testBuilder.Definition.Spec.ImageSetRef.Name)
	}
}

func TestAgentClusterInstallWithWorkerAgents(t *testing.T) {
	testCases := []struct {
		count         int
		expectedError string
	}{
		{
			count:         3,
			expectedError: "",
		},
		{
			count:         1,
			expectedError: "",
		},
		{
			count:         0,
			expectedError: "",
		},
		{
			count:         -1,
			expectedError: "agentclusterinstall worker agents cannot be less that 0",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateAgentClusterInstallTestBuilder()
		testBuilder.WithWorkerAgents(testCase.count)

		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)
		assert.Equal(t, testCase.count, testBuilder.Definition.Spec.ProvisionRequirements.WorkerAgents)
	}
}

//nolint:lll
func TestAgentClusterInstallWithSSHPublicKey(t *testing.T) {
	testCases := []struct {
		sshKey string
	}{
		{
			sshKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCaX4UsUhbLbeTGx6wob/6Y2qQzBDi2VhjFaFGugNk9zuNgPB4Um4VNa8TgRr/nrJxLWduuM0gbBgAA9719cCV6FsTq6e0TTrtFAy498vNrE1gyCp8gX2/+60NWwjeK3iglVym0WxmXhWpZCPk6EBsBQgclt6iQhVH20f60lWAD6+8I5iXCm5u63HNEzuTmJUVaVCEU62zamwwJUR+6yddnu+rIrLqw7xYuuATKMBuL+AUnB3zR08qv65Tyhf/YpyjlW/w7kCFwTsE0Xrbtumd1irF49DiY4Xb5WGPRTjrhVhzpAvtvEhfrMGGnbwnQ0P1pUTAQz8eYKSN4M0oxaGPObtjkEM4EAjBo5vyzN1xi9H/tekOQCmVDY2XOCVvIuxUPgq/8Uc3oGkoh3Eay+KSRIhQC3+kqXX889H//SmENoejIAxnETXPnVygKMaoJbh2d44RjIC5LkJsNSwcK24+hv4VU0uJjwXMPPE5YTUxcVPukATd4i22FOSiIL/ULuos=",
		},
		{
			sshKey: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateAgentClusterInstallTestBuilder()
		testBuilder.WithSSHPublicKey(testCase.sshKey)

		assert.Equal(t, testCase.sshKey, testBuilder.Definition.Spec.SSHPublicKey)
	}
}
func TestAgentClusterInstallWithNetworkType(t *testing.T) {
	testCases := []struct {
		network string
	}{
		{
			network: "OpenShiftSDN",
		},
		{
			network: "OVNKubernetes",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateAgentClusterInstallTestBuilder()
		testBuilder.WithNetworkType(testCase.network)

		assert.Equal(t, testCase.network, testBuilder.Definition.Spec.Networking.NetworkType)
	}
}
func TestAgentClusterInstallWithAdditionalClusterNetwork(t *testing.T) {
	testCases := []struct {
		cidr          string
		prefix        int32
		expectedError string
	}{
		{
			cidr:          "10.128.0.0/14",
			prefix:        23,
			expectedError: "",
		},
		{
			cidr:          "10.128.0.0",
			prefix:        23,
			expectedError: "agentclusterinstall contains invalid clusterNetwork cidr",
		},
		{
			cidr:          "10.128.0.0/14",
			prefix:        0,
			expectedError: "agentclusterinstall contains invalid clusterNetwork prefix",
		},
		{
			cidr:          "10.128.0.0/14",
			prefix:        -1,
			expectedError: "agentclusterinstall contains invalid clusterNetwork prefix",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateAgentClusterInstallTestBuilder()
		testBuilder.WithAdditionalClusterNetwork(testCase.cidr, testCase.prefix)
		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.cidr, testBuilder.Definition.Spec.Networking.ClusterNetwork[1].CIDR)
			assert.Equal(t, testCase.prefix, testBuilder.Definition.Spec.Networking.ClusterNetwork[1].HostPrefix)
		}
	}
}

func TestAgentClusterInstallWithAdditionalServiceNetwork(t *testing.T) {
	testCases := []struct {
		cidr          string
		expectedError string
	}{
		{
			cidr:          "172.30.0.0/16",
			expectedError: "",
		},
		{
			cidr:          "",
			expectedError: "agentclusterinstall contains invalid serviceNetwork cidr",
		},
		{
			cidr:          "172.30.0.0/162",
			expectedError: "agentclusterinstall contains invalid serviceNetwork cidr",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateAgentClusterInstallTestBuilder()
		testBuilder.WithAdditionalServiceNetwork(testCase.cidr)
		assert.Equal(t, testCase.expectedError, testBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.cidr, testBuilder.Definition.Spec.Networking.ServiceNetwork[1])
		}
	}
}
func TestAgentClusterInstallWaitForState(t *testing.T) {
	testCases := []struct {
		status hiveextV1Beta1.AgentClusterInstallStatus
		state  string
	}{
		{
			status: hiveextV1Beta1.AgentClusterInstallStatus{
				DebugInfo: hiveextV1Beta1.DebugInfo{
					State: "adding-hosts",
				},
			},
			state: "adding-hosts",
		},
		{
			status: hiveextV1Beta1.AgentClusterInstallStatus{
				DebugInfo: hiveextV1Beta1.DebugInfo{
					State: "installing",
				},
			},
			state: "installing",
		},
		{
			status: hiveextV1Beta1.AgentClusterInstallStatus{
				DebugInfo: hiveextV1Beta1.DebugInfo{
					State: "finalizing",
				},
			},
			state: "finalizing",
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		testACI := generateAgentClusterInstall()
		testACI.Status = testCase.status

		runtimeObjects = append(runtimeObjects, testACI)

		testBuilder := buildTestBuilderWithFakeObjects(runtimeObjects)

		aci, err := testBuilder.WaitForState(testCase.state, time.Second*2)

		assert.Nil(t, err)
		assert.NotNil(t, aci)
	}
}
func TestAgentClusterInstallWaitForStateInfo(t *testing.T) {
	testCases := []struct {
		status hiveextV1Beta1.AgentClusterInstallStatus
		info   string
	}{
		{
			status: hiveextV1Beta1.AgentClusterInstallStatus{
				DebugInfo: hiveextV1Beta1.DebugInfo{
					StateInfo: "Cluster is installed",
				},
			},
			info: "Cluster is installed",
		},
		{
			status: hiveextV1Beta1.AgentClusterInstallStatus{
				DebugInfo: hiveextV1Beta1.DebugInfo{
					StateInfo: "Cluster is finalizing",
				},
			},
			info: "Cluster is finalizing",
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		testACI := generateAgentClusterInstall()
		testACI.Status = testCase.status

		runtimeObjects = append(runtimeObjects, testACI)

		testBuilder := buildTestBuilderWithFakeObjects(runtimeObjects)

		aci, err := testBuilder.WaitForStateInfo(testCase.info, time.Second*2)

		assert.Nil(t, err)
		assert.NotNil(t, aci)
	}
}

func TestAgentClusterInstallWaitForConditionMessage(t *testing.T) {
	testCases := []struct {
		status        hiveextV1Beta1.AgentClusterInstallStatus
		conditionType hivev1.ClusterInstallConditionType
		message       string
	}{
		{
			status: hiveextV1Beta1.AgentClusterInstallStatus{
				Conditions: []hivev1.ClusterInstallCondition{
					{
						Type:    "Stopped",
						Status:  corev1.ConditionTrue,
						Reason:  "InstallationCompleted",
						Message: "The installation has stopped because it completed successfully",
					},
				},
			},
			conditionType: "Stopped",
			message:       "The installation has stopped because it completed successfully",
		},
		{
			status: hiveextV1Beta1.AgentClusterInstallStatus{
				Conditions: []hivev1.ClusterInstallCondition{
					{
						Type:    "SpecSynced",
						Status:  corev1.ConditionTrue,
						Reason:  "SyncOK",
						Message: "SyncOK",
					},
				},
			},
			conditionType: "SpecSynced",
			message:       "SyncOK",
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		testACI := generateAgentClusterInstall()
		testACI.Status = testCase.status

		runtimeObjects = append(runtimeObjects, testACI)

		testBuilder := buildTestBuilderWithFakeObjects(runtimeObjects)

		err := testBuilder.WaitForConditionMessage(testCase.conditionType, testCase.message, time.Second*5)

		assert.Nil(t, err)
	}
}
func TestAgentClusterInstallWaitForConditionStatus(t *testing.T) {
	testCases := []struct {
		status          hiveextV1Beta1.AgentClusterInstallStatus
		conditionType   hivev1.ClusterInstallConditionType
		conditionStatus corev1.ConditionStatus
	}{
		{
			status: hiveextV1Beta1.AgentClusterInstallStatus{
				Conditions: []hivev1.ClusterInstallCondition{
					{
						Type:    "Stopped",
						Status:  corev1.ConditionTrue,
						Reason:  "InstallationCompleted",
						Message: "The installation has stopped because it completed successfully",
					},
				},
			},
			conditionType:   "Stopped",
			conditionStatus: corev1.ConditionTrue,
		},
		{
			status: hiveextV1Beta1.AgentClusterInstallStatus{
				Conditions: []hivev1.ClusterInstallCondition{
					{
						Type:    "Failed",
						Status:  corev1.ConditionFalse,
						Reason:  "InstallationNotFailed",
						Message: "The installation has not failed",
					},
				},
			},
			conditionType:   "Failed",
			conditionStatus: corev1.ConditionFalse,
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		testACI := generateAgentClusterInstall()
		testACI.Status = testCase.status

		runtimeObjects = append(runtimeObjects, testACI)

		testBuilder := buildTestBuilderWithFakeObjects(runtimeObjects)

		err := testBuilder.WaitForConditionStatus(testCase.conditionType, testCase.conditionStatus, time.Second*2)

		assert.Nil(t, err)
	}
}
func TestAgentClusterInstallWaitForConditionReason(t *testing.T) {
	testCases := []struct {
		status        hiveextV1Beta1.AgentClusterInstallStatus
		conditionType hivev1.ClusterInstallConditionType
		reason        string
	}{
		{
			status: hiveextV1Beta1.AgentClusterInstallStatus{
				Conditions: []hivev1.ClusterInstallCondition{
					{
						Type:    "Stopped",
						Status:  corev1.ConditionTrue,
						Reason:  "InstallationCompleted",
						Message: "The installation has stopped because it completed successfully",
					},
				},
			},
			conditionType: "Stopped",
			reason:        "InstallationCompleted",
		},
		{
			status: hiveextV1Beta1.AgentClusterInstallStatus{
				Conditions: []hivev1.ClusterInstallCondition{
					{
						Type:    "Failed",
						Status:  corev1.ConditionFalse,
						Reason:  "InstallationNotFailed",
						Message: "The installation has not failed",
					},
				},
			},
			conditionType: "Failed",
			reason:        "InstallationNotFailed",
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		testACI := generateAgentClusterInstall()
		testACI.Status = testCase.status

		runtimeObjects = append(runtimeObjects, testACI)

		testBuilder := buildTestBuilderWithFakeObjects(runtimeObjects)

		err := testBuilder.WaitForConditionReason(testCase.conditionType, testCase.reason, time.Second*2)

		assert.Nil(t, err)
	}
}
func TestAgentClusterInstallGetEvents(t *testing.T) {
	path, err := os.Getwd()
	assert.Nil(t, err)

	testCases := []struct {
		status hiveextV1Beta1.AgentClusterInstallStatus
	}{
		{
			status: hiveextV1Beta1.AgentClusterInstallStatus{
				DebugInfo: hiveextV1Beta1.DebugInfo{
					EventsURL: fmt.Sprintf("file://%s/testdata/aci_events.json", path),
				},
			},
		},
	}

	for _, testCase := range testCases {
		var (
			runtimeObjects []runtime.Object
		)

		transport := &http.Transport{}
		transport.RegisterProtocol("file", http.NewFileTransport(http.Dir("/")))
		eventsTransport = transport

		testACI := generateAgentClusterInstall()
		testACI.Status = testCase.status

		runtimeObjects = append(runtimeObjects, testACI)

		testBuilder := buildTestBuilderWithFakeObjects(runtimeObjects)

		events, err := testBuilder.GetEvents(true)

		assert.Nil(t, err)
		assert.NotNil(t, events)
		assert.Len(t, events, 63)
	}
}

func TestAgentClusterInstallValidate(t *testing.T) {
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
			expectedError: "error: received nil AgentClusterInstall builder",
		},
		{
			builderNil:    false,
			definitionNil: true,
			apiClientNil:  false,
			expectedError: "can not redefine the undefined AgentClusterInstall",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  true,
			expectedError: "AgentClusterInstall builder cannot have nil apiClient",
		},
		{
			builderNil:    false,
			definitionNil: false,
			apiClientNil:  false,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildTestBuilderWithFakeObjects([]runtime.Object{})

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

func dummyTestNetwork() hiveextV1Beta1.Networking {
	return hiveextV1Beta1.Networking{
		ClusterNetwork: []hiveextV1Beta1.ClusterNetworkEntry{{
			CIDR:       "10.128.0.0/14",
			HostPrefix: 23,
		}},
		ServiceNetwork: []string{"172.30.0.0/16"},
	}
}

func buildTestBuilderWithFakeObjects(objects []runtime.Object) *AgentClusterInstallBuilder {
	fakeClientScheme := runtime.NewScheme()

	err := clients.SetScheme(fakeClientScheme)
	if err != nil {
		return nil
	}

	apiClient := clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  objects,
		SchemeAttachers: testSchemes,
	})

	testBuilder := NewAgentClusterInstallBuilder(
		apiClient,
		aciTestName,
		aciTestNamespace,
		"aci-test-clusterdeployment",
		3,
		2,
		dummyTestNetwork())

	return testBuilder
}

func generateAgentClusterInstallTestBuilder() *AgentClusterInstallBuilder {
	return &AgentClusterInstallBuilder{
		apiClient:  clients.GetTestClients(clients.TestClientParams{}).Client,
		Definition: generateAgentClusterInstall(),
	}
}

func generateAgentClusterInstall() *hiveextV1Beta1.AgentClusterInstall {
	return &hiveextV1Beta1.AgentClusterInstall{
		ObjectMeta: metav1.ObjectMeta{
			Name:      aciTestName,
			Namespace: aciTestNamespace,
		},
		Spec: hiveextV1Beta1.AgentClusterInstallSpec{
			ClusterDeploymentRef: corev1.LocalObjectReference{
				Name: aciTestName,
			},
			Networking: dummyTestNetwork(),
			ProvisionRequirements: hiveextV1Beta1.ProvisionRequirements{
				ControlPlaneAgents: 3,
				WorkerAgents:       2,
			},
		},
	}
}
