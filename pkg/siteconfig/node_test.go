package siteconfig

import (
	"testing"

	"github.com/openshift-kni/eco-goinfra/pkg/schemes/assisted/api/v1beta1"
	"github.com/stretchr/testify/assert"
)

func TestClusterInstanceNewNewNodeBuilder(t *testing.T) {
	testCases := []struct {
		name               string
		bmcAddress         string
		bootMACAddress     string
		bmcCredentialsName string
		templateName       string
		templateNamespace  string
		expectedErrorText  string
	}{
		{
			name:               "test-node",
			bmcAddress:         "test-bmc-address",
			bootMACAddress:     "00:00:00:00:00:00",
			bmcCredentialsName: "test-creds",
			templateName:       "test-template",
			templateNamespace:  "test-namespace",
			expectedErrorText:  "",
		},
		{
			name:               "",
			bmcAddress:         "test-bmc-address",
			bootMACAddress:     "00:00:00:00:00:00",
			bmcCredentialsName: "test-creds",
			templateName:       "test-template",
			templateNamespace:  "test-namespace",
			expectedErrorText:  "siteconfig node 'name' cannot be empty",
		},
		{
			name:               "test-node",
			bmcAddress:         "",
			bootMACAddress:     "00:00:00:00:00:00",
			bmcCredentialsName: "test-creds",
			templateName:       "test-template",
			templateNamespace:  "test-namespace",
			expectedErrorText:  "siteconfig node 'bmcAddress' cannot be empty",
		},
		{
			name:               "test-node",
			bmcAddress:         "test-bmc-address",
			bootMACAddress:     "",
			bmcCredentialsName: "test-creds",
			templateName:       "test-template",
			templateNamespace:  "test-namespace",
			expectedErrorText:  "siteconfig node 'bootMACAddress' cannot be empty",
		},
		{
			name:               "test-node",
			bmcAddress:         "test-bmc-address",
			bootMACAddress:     "00:00:00:00:00:00",
			bmcCredentialsName: "",
			templateName:       "test-template",
			templateNamespace:  "test-namespace",
			expectedErrorText:  "siteconfig node 'bmcCredentialsName' cannot be empty",
		},
		{
			name:               "test-node",
			bmcAddress:         "test-bmc-address",
			bootMACAddress:     "00:00:00:00:00:00",
			bmcCredentialsName: "test-creds",
			templateName:       "",
			templateNamespace:  "test-namespace",
			expectedErrorText:  "siteconfig node 'templateName' cannot be empty",
		},
		{
			name:               "test-node",
			bmcAddress:         "test-bmc-address",
			bootMACAddress:     "00:00:00:00:00:00",
			bmcCredentialsName: "test-creds",
			templateName:       "test-template",
			templateNamespace:  "",
			expectedErrorText:  "siteconfig node 'templateNamespace' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := NewNodeBuilder(testCase.name, testCase.bmcAddress, testCase.bootMACAddress,
			testCase.bmcCredentialsName, testCase.templateName, testCase.templateNamespace)

		assert.Equal(t, testCase.expectedErrorText, testBuilder.errorMsg)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, testCase.name, testBuilder.definition.HostName)
			assert.Equal(t, testCase.bmcAddress, testBuilder.definition.BmcAddress)
			assert.Equal(t, testCase.bootMACAddress, testBuilder.definition.BootMACAddress)
			assert.Equal(t, testCase.bmcCredentialsName, testBuilder.definition.BmcCredentialsName.Name)
			assert.Equal(t, testCase.templateName, testBuilder.definition.TemplateRefs[0].Name)
			assert.Equal(t, testCase.templateNamespace, testBuilder.definition.TemplateRefs[0].Namespace)
		}
	}
}

func TestClusterInstanceNodeWithAutomatedCleaningMode(t *testing.T) {
	testCases := []struct {
		cleaningMode      string
		expectedErrorText string
	}{
		{
			cleaningMode:      "disabled",
			expectedErrorText: "",
		},
		{
			cleaningMode:      "metadata",
			expectedErrorText: "",
		},
		{
			cleaningMode:      "off",
			expectedErrorText: "siteconfig node automatedCleaningMode must be one of: disabled, metadata",
		},
		{
			cleaningMode:      "",
			expectedErrorText: "siteconfig node automatedCleaningMode must be one of: disabled, metadata",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateNodeBuilder()

		testBuilder.WithAutomatedCleaningMode(testCase.cleaningMode)
		assert.Equal(t, testCase.expectedErrorText, testBuilder.errorMsg)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, testCase.cleaningMode, string(testBuilder.definition.AutomatedCleaningMode))
		}
	}
}

func TestClusterInstanceNodeWithNodeNetwork(t *testing.T) {
	testCases := []struct {
		networkConfig     *v1beta1.NMStateConfigSpec
		expectedErrorText string
	}{
		{
			networkConfig:     generateNetworkConfig(),
			expectedErrorText: "",
		},
		{
			networkConfig:     nil,
			expectedErrorText: "siteconfig node networkConfig cannot be nil",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateNodeBuilder()

		testBuilder.WithNodeNetwork(testCase.networkConfig)
		assert.Equal(t, testCase.expectedErrorText, testBuilder.errorMsg)

		if testCase.expectedErrorText == "" {
			assert.Equal(t, testCase.networkConfig, testBuilder.definition.NodeNetwork)
		}
	}
}

func TestClusterInstanceNodeGenerate(t *testing.T) {
	testCases := []struct {
		builder *NodeBuilder
	}{
		{
			builder: generateNodeBuilder(),
		},
	}

	for _, testCase := range testCases {
		testNodeSpec, err := testCase.builder.Generate()
		assert.Nil(t, err)
		assert.Equal(t, testCase.builder.definition, testNodeSpec)
	}
}

func TestClusterInstanceNodeValidate(t *testing.T) {
	testCases := []struct {
		builderNil    bool
		definitionNil bool
		expectedError string
	}{
		{
			builderNil:    true,
			definitionNil: false,
			expectedError: "error: received nil siteconfig node builder",
		},
		{
			builderNil:    false,
			definitionNil: true,
			expectedError: "can not redefine the undefined siteconfig node",
		},
		{
			builderNil:    false,
			definitionNil: false,
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := generateNodeBuilder()

		if testCase.builderNil {
			testBuilder = nil
		}

		if testCase.definitionNil {
			testBuilder.definition = nil
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

// buildValidClusterInstanceNodeTestBuilder returns a valid ClusterInstanceBuilder for testing purposes.
func generateNodeBuilder() *NodeBuilder {
	return NewNodeBuilder(
		"test-node",
		"test-bmc-address",
		"00:00:00:00:00:00",
		"test-creds",
		"test-template",
		"test-namespace",
	)
}

func generateNetworkConfig() *v1beta1.NMStateConfigSpec {
	config := ` config:
					interfaces:
					  - name: eno1
						  type: ethernet
						  state: up
						  ipv6:
						  	  enabled: false
						  ipv4:
							  enabled: false
							  address:
							  - ip: 192.168.122.100
								  prefix-length: 24
							  dhcp: false
					dns-resolver:
						config:
						server:
							- 8.8.8.8
					routes:
						config:
						- destination: 0.0.0.0
							next-hop-address: 192.168.122.1
							next-hop-interface: "eno1"
							table-id: 254
				interfaces:
				  - name: eno1
					macAddress: 00:00:00:00:00:00`

	return &v1beta1.NMStateConfigSpec{
		Interfaces: []*v1beta1.Interface{
			{
				Name:       "eno1",
				MacAddress: "00:00:00:00:00:00",
			},
		},
		NetConfig: v1beta1.NetConfig{
			Raw: v1beta1.RawNetConfig([]byte(config)),
		},
	}
}
