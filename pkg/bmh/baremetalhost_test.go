package bmh

import (
	"context"
	"fmt"
	"testing"
	"time"

	bmhv1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	defaultBmHostName       = "metallbio"
	defaultBmHostNsName     = "test-namespace"
	defaultBmHostAddress    = "1.1.1.1"
	defaultBmHostSecretName = "testsecret"
	defaultBmHostMacAddress = "AA:BB:CC:11:22:33"
	defaultBmHostBootMode   = "UEFISecureBoot"
)

func TestBareMetalHostPull(t *testing.T) {
	generateBaremetalHost := func(name, namespace string) *bmhv1alpha1.BareMetalHost {
		return &bmhv1alpha1.BareMetalHost{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: bmhv1alpha1.BareMetalHostSpec{},
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
			name:                "metallbio",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("baremetalhost 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                "metallbio",
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("baremetalhost 'namespace' cannot be empty"),
			client:              true,
		},
		{
			name:                "metallbio",
			namespace:           "test-namespace",
			addToRuntimeObjects: false,
			expectedError:       fmt.Errorf("baremetalhost object metallbio does not exist in namespace test-namespace"),
			client:              true,
		},
		{
			name:                "metallbio",
			namespace:           "test-namespace",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("baremetalhost 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testBmHost := generateBaremetalHost(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testBmHost)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})
		}

		builderResult, err := Pull(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.name, builderResult.Object.Name)
			assert.Equal(t, testCase.namespace, builderResult.Object.Namespace)
		}
	}
}

//nolint:funlen
func TestBareMetalHostPullNewBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		bmcAddress    string
		bmcSecretName string
		bmcMacAddress string
		bootMode      string
		label         map[string]string
		expectedError string
	}{
		{
			name:          "metallbio",
			namespace:     "test-namespace",
			bmcAddress:    "1.1.1.1",
			bmcSecretName: "test-secret",
			bmcMacAddress: "AA:BB:CC:DD:11:22",
			bootMode:      "UEFISecureBoot",
			label:         map[string]string{"test": "test"},
			expectedError: "",
		},
		{
			name:          "",
			namespace:     "test-namespace",
			bmcAddress:    "1.1.1.1",
			bmcSecretName: "test-secret",
			bmcMacAddress: "AA:BB:CC:DD:11:22",
			bootMode:      "UEFISecureBoot",
			label:         map[string]string{"test": "test"},
			expectedError: "BMH 'name' cannot be empty",
		},
		{
			name:          "metallbio",
			namespace:     "",
			bmcAddress:    "1.1.1.1",
			bmcSecretName: "test-secret",
			bmcMacAddress: "AA:BB:CC:DD:11:22",
			bootMode:      "UEFISecureBoot",
			label:         map[string]string{"test": "test"},
			expectedError: "BMH 'nsname' cannot be empty",
		},
		{
			name:          "metallbio",
			namespace:     "test-namespace",
			bmcAddress:    "",
			bmcSecretName: "test-secret",
			bmcMacAddress: "AA:BB:CC:DD:11:22",
			bootMode:      "UEFISecureBoot",
			label:         map[string]string{"test": "test"},
			expectedError: "BMH 'bmcAddress' cannot be empty",
		},
		{
			name:          "metallbio",
			namespace:     "test-namespace",
			bmcAddress:    "1.1.1.1",
			bmcSecretName: "",
			bmcMacAddress: "AA:BB:CC:DD:11:22",
			bootMode:      "UEFISecureBoot",
			expectedError: "BMH 'bmcSecretName' cannot be empty",
		},
		{
			name:          "metallbio",
			namespace:     "test-namespace",
			bmcAddress:    "1.1.1.1",
			bmcSecretName: "test-secret",
			bmcMacAddress: "",
			bootMode:      "UEFISecureBoot",
			expectedError: "BMH 'bootMacAddress' cannot be empty",
		},
		{
			name:          "metallbio",
			namespace:     "test-namespace",
			bmcAddress:    "1.1.1.1",
			bmcSecretName: "test-secret",
			bmcMacAddress: "AA:BB:CC:DD:11:22",
			bootMode:      "",
			expectedError: "not acceptable 'bootMode' value",
		},
		{
			name:          "metallbio",
			namespace:     "test-namespace",
			bmcAddress:    "1.1.1.1",
			bmcSecretName: "test-secret",
			bmcMacAddress: "AA:BB:CC:DD:11:22",
			bootMode:      "UEFISecureBoot",
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testSettings := clients.GetTestClients(clients.TestClientParams{})
		testMetalLbBuilder := NewBuilder(
			testSettings,
			testCase.name,
			testCase.namespace,
			testCase.bmcAddress,
			testCase.bmcSecretName,
			testCase.bmcMacAddress,
			testCase.bootMode)
		assert.Equal(t, testCase.expectedError, testMetalLbBuilder.errorMsg)
		assert.NotNil(t, testMetalLbBuilder.Definition)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.name, testMetalLbBuilder.Definition.Name)
			assert.Equal(t, testCase.namespace, testMetalLbBuilder.Definition.Namespace)
		}
	}
}

func TestBareMetalHostExists(t *testing.T) {
	testCases := []struct {
		testBmHost     *BmhBuilder
		expectedStatus bool
	}{
		{
			testBmHost:     buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedStatus: true,
		},
		{
			testBmHost:     buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus: false,
		},
		{
			testBmHost:     buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testBmHost.Exists()
		assert.Equal(t, exist, testCase.expectedStatus)
	}
}

func TestBareMetalHostGet(t *testing.T) {
	testCases := []struct {
		testBmHost    *BmhBuilder
		expectedError error
	}{
		{
			testBmHost:    buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testBmHost:    buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("baremetalhosts.metal3.io \"metallbio\" not found"),
		},
		{
			testBmHost:    buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError: fmt.Errorf("not acceptable 'bootMode' value"),
		},
	}

	for _, testCase := range testCases {
		bmHost, err := testCase.testBmHost.Get()

		if testCase.expectedError == nil {
			assert.Equal(t, bmHost.Name, testCase.testBmHost.Definition.Name)
			assert.Equal(t, bmHost.Namespace, testCase.testBmHost.Definition.Namespace)
		} else {
			assert.Equal(t, err.Error(), testCase.expectedError.Error())
		}
	}
}

func TestBareMetalHostCreate(t *testing.T) {
	testCases := []struct {
		testBmHost    *BmhBuilder
		expectedError error
	}{
		{
			testBmHost:    buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testBmHost:    buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
		{
			testBmHost:    buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError: fmt.Errorf("not acceptable 'bootMode' value"),
		},
	}

	for _, testCase := range testCases {
		ipAddressPoolBuilder, err := testCase.testBmHost.Create()

		if testCase.expectedError == nil {
			assert.Equal(t, ipAddressPoolBuilder.Definition.Name, ipAddressPoolBuilder.Object.Name)
			assert.Equal(t, ipAddressPoolBuilder.Definition.Namespace, ipAddressPoolBuilder.Object.Namespace)
		} else {
			assert.Equal(t, err.Error(), testCase.expectedError.Error())
		}
	}
}

func TestBareMetalHostDelete(t *testing.T) {
	testCases := []struct {
		testBmHost    *BmhBuilder
		expectedError error
	}{
		{
			testBmHost:    buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testBmHost:    buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("bmh cannot be deleted because it does not exist"),
		},
		{
			testBmHost:    buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError: fmt.Errorf("not acceptable 'bootMode' value"),
		},
	}

	for _, testCase := range testCases {
		_, err := testCase.testBmHost.Delete()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testBmHost.Object)
		}
	}
}

func TestBareMetalHostWithRootDeviceDeviceName(t *testing.T) {
	testCases := []struct {
		testBmHost       *BmhBuilder
		expectedError    string
		deviceDeviceName string
	}{
		{
			testBmHost:       buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError:    "",
			deviceDeviceName: "123",
		},
		{
			testBmHost:       buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			deviceDeviceName: "",
			expectedError:    "the baremetalhost rootDeviceHint deviceName cannot be empty",
		},
		{
			testBmHost:       buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			deviceDeviceName: "123",
			expectedError:    "not acceptable 'bootMode' value",
		},
	}

	for _, testCase := range testCases {
		testBmHostBuilder := testCase.testBmHost.WithRootDeviceDeviceName(testCase.deviceDeviceName)
		assert.Equal(t, testCase.expectedError, testBmHostBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.deviceDeviceName, testBmHostBuilder.Definition.Spec.RootDeviceHints.DeviceName)
		}
	}
}

func TestBareMetalHostWithRootDeviceHTCL(t *testing.T) {
	testCases := []struct {
		testBmHost     *BmhBuilder
		expectedError  string
		rootDeviceHTCL string
	}{
		{
			testBmHost:     buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError:  "",
			rootDeviceHTCL: "123",
		},
		{
			testBmHost:     buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			rootDeviceHTCL: "",
			expectedError:  "the baremetalhost rootDeviceHint hctl cannot be empty",
		},
		{
			testBmHost:     buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			rootDeviceHTCL: "123",
			expectedError:  "not acceptable 'bootMode' value",
		},
	}

	for _, testCase := range testCases {
		testBmHostBuilder := testCase.testBmHost.WithRootDeviceHTCL(testCase.rootDeviceHTCL)
		assert.Equal(t, testCase.expectedError, testBmHostBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.rootDeviceHTCL, testBmHostBuilder.Definition.Spec.RootDeviceHints.HCTL)
		}
	}
}

func TestBareMetalHostWithRootDeviceModel(t *testing.T) {
	testCases := []struct {
		testBmHost      *BmhBuilder
		expectedError   string
		rootDeviceModel string
	}{
		{
			testBmHost:      buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError:   "",
			rootDeviceModel: "123",
		},
		{
			testBmHost:      buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			rootDeviceModel: "",
			expectedError:   "the baremetalhost rootDeviceHint model cannot be empty",
		},
		{
			testBmHost:      buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			rootDeviceModel: "123",
			expectedError:   "not acceptable 'bootMode' value",
		},
	}

	for _, testCase := range testCases {
		testBmHostBuilder := testCase.testBmHost.WithRootDeviceModel(testCase.rootDeviceModel)
		assert.Equal(t, testCase.expectedError, testBmHostBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.rootDeviceModel, testBmHostBuilder.Definition.Spec.RootDeviceHints.Model)
		}
	}
}

func TestBareMetalHostWithRootDeviceVendor(t *testing.T) {
	testCases := []struct {
		testBmHost       *BmhBuilder
		expectedError    string
		rootDeviceVendor string
	}{
		{
			testBmHost:       buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError:    "",
			rootDeviceVendor: "123",
		},
		{
			testBmHost:       buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			rootDeviceVendor: "",
			expectedError:    "the baremetalhost rootDeviceHint vendor cannot be empty",
		},
		{
			testBmHost:       buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			rootDeviceVendor: "123",
			expectedError:    "not acceptable 'bootMode' value",
		},
	}

	for _, testCase := range testCases {
		testBmHostBuilder := testCase.testBmHost.WithRootDeviceVendor(testCase.rootDeviceVendor)
		assert.Equal(t, testCase.expectedError, testBmHostBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(t, testCase.rootDeviceVendor, testBmHostBuilder.Definition.Spec.RootDeviceHints.Model)
		}
	}
}

func TestBareMetalHostWithRootDeviceSerialNumber(t *testing.T) {
	testCases := []struct {
		testBmHost             *BmhBuilder
		expectedError          string
		rootDeviceSerialNumber string
	}{
		{
			testBmHost:             buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError:          "",
			rootDeviceSerialNumber: "123",
		},
		{
			testBmHost:             buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			rootDeviceSerialNumber: "",
			expectedError:          "the baremetalhost rootDeviceHint serialNumber cannot be empty",
		},
		{
			testBmHost:             buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			rootDeviceSerialNumber: "123",
			expectedError:          "not acceptable 'bootMode' value",
		},
	}

	for _, testCase := range testCases {
		testBmHostBuilder := testCase.testBmHost.WithRootDeviceSerialNumber(testCase.rootDeviceSerialNumber)
		assert.Equal(t, testCase.expectedError, testBmHostBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(
				t, testCase.rootDeviceSerialNumber, testBmHostBuilder.Definition.Spec.RootDeviceHints.SerialNumber)
		}
	}
}

func TestBareMetalHostWithRootDeviceMinSizeGigabytes(t *testing.T) {
	testCases := []struct {
		testBmHost                 *BmhBuilder
		expectedError              string
		rootDeviceMinSizeGigabytes int
	}{
		{
			testBmHost:                 buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError:              "",
			rootDeviceMinSizeGigabytes: 12,
		},
		{
			testBmHost:                 buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			rootDeviceMinSizeGigabytes: -1,
			expectedError:              "the baremetalhost rootDeviceHint size cannot be less than 0",
		},
		{
			testBmHost:                 buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			rootDeviceMinSizeGigabytes: 123,
			expectedError:              "not acceptable 'bootMode' value",
		},
	}

	for _, testCase := range testCases {
		testBmHostBuilder := testCase.testBmHost.WithRootDeviceMinSizeGigabytes(testCase.rootDeviceMinSizeGigabytes)
		assert.Equal(t, testCase.expectedError, testBmHostBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(
				t, testCase.rootDeviceMinSizeGigabytes, testBmHostBuilder.Definition.Spec.RootDeviceHints.MinSizeGigabytes)
		}
	}
}

func TestBareMetalHostWithRootDeviceWWN(t *testing.T) {
	testCases := []struct {
		testBmHost    *BmhBuilder
		expectedError string
		rootDeviceWwn string
	}{
		{
			testBmHost:    buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError: "",
			rootDeviceWwn: "test",
		},
		{
			testBmHost:    buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			rootDeviceWwn: "",
			expectedError: "the baremetalhost rootDeviceHint wwn cannot be empty",
		},
		{
			testBmHost:    buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			rootDeviceWwn: "test",
			expectedError: "not acceptable 'bootMode' value",
		},
	}

	for _, testCase := range testCases {
		testBmHostBuilder := testCase.testBmHost.WithRootDeviceWWN(testCase.rootDeviceWwn)
		assert.Equal(t, testCase.expectedError, testBmHostBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(
				t, testCase.rootDeviceWwn, testBmHostBuilder.Definition.Spec.RootDeviceHints.WWN)
		}
	}
}

func TestBareMetalHostWithRootDeviceWWNWithExtension(t *testing.T) {
	testCases := []struct {
		testBmHost                 *BmhBuilder
		expectedError              string
		rootDeviceWWNWithExtension string
	}{
		{
			testBmHost:                 buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError:              "",
			rootDeviceWWNWithExtension: "test",
		},
		{
			testBmHost:                 buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			rootDeviceWWNWithExtension: "",
			expectedError:              "the baremetalhost rootDeviceHint wwnWithExtension cannot be empty",
		},
		{
			testBmHost:                 buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			rootDeviceWWNWithExtension: "test",
			expectedError:              "not acceptable 'bootMode' value",
		},
	}

	for _, testCase := range testCases {
		testBmHostBuilder := testCase.testBmHost.WithRootDeviceWWNWithExtension(testCase.rootDeviceWWNWithExtension)
		assert.Equal(t, testCase.expectedError, testBmHostBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(
				t, testCase.rootDeviceWWNWithExtension, testBmHostBuilder.Definition.Spec.RootDeviceHints.WWNWithExtension)
		}
	}
}

func TestBareMetalHostWithRootDeviceWWNVendorExtension(t *testing.T) {
	testCases := []struct {
		testBmHost                   *BmhBuilder
		expectedError                string
		rootDeviceWWNVendorExtension string
	}{
		{
			testBmHost:                   buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError:                "",
			rootDeviceWWNVendorExtension: "test",
		},
		{
			testBmHost:                   buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			rootDeviceWWNVendorExtension: "",
			expectedError:                "the baremetalhost rootDeviceHint wwnVendorExtension cannot be empty",
		},
		{
			testBmHost:                   buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			rootDeviceWWNVendorExtension: "test",
			expectedError:                "not acceptable 'bootMode' value",
		},
	}

	for _, testCase := range testCases {
		testBmHostBuilder := testCase.testBmHost.WithRootDeviceWWNVendorExtension(testCase.rootDeviceWWNVendorExtension)
		assert.Equal(t, testCase.expectedError, testBmHostBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(
				t, testCase.rootDeviceWWNVendorExtension, testBmHostBuilder.Definition.Spec.RootDeviceHints.WWNVendorExtension)
		}
	}
}

func TestBareMetalHostWithRootDeviceRotationalDisk(t *testing.T) {
	testCases := []struct {
		testBmHost    *BmhBuilder
		expectedError string
		rotational    bool
	}{
		{
			testBmHost:    buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError: "",
			rotational:    true,
		},
		{
			testBmHost:    buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: "",
			rotational:    false,
		},
		{
			testBmHost:    buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError: "not acceptable 'bootMode' value",
			rotational:    false,
		},
	}

	for _, testCase := range testCases {
		testBmHostBuilder := testCase.testBmHost.WithRootDeviceRotationalDisk(testCase.rotational)
		assert.Equal(t, testCase.expectedError, testBmHostBuilder.errorMsg)

		if testCase.expectedError == "" {
			assert.Equal(
				t, &testCase.rotational, testBmHostBuilder.Definition.Spec.RootDeviceHints.Rotational)
		}
	}
}

func TestBareMetalHostWithOptions(t *testing.T) {
	testSettings := buildBareMetalHostTestClientWithDummyObject()
	testBuilder := buildValidBmHostBuilder(testSettings).WithOptions(
		func(builder *BmhBuilder) (*BmhBuilder, error) {
			return builder, nil
		})

	assert.Equal(t, "", testBuilder.errorMsg)
	testBuilder = buildValidBmHostBuilder(testSettings).WithOptions(
		func(builder *BmhBuilder) (*BmhBuilder, error) {
			return builder, fmt.Errorf("error")
		})
	assert.Equal(t, "error", testBuilder.errorMsg)
}

func TestBareMetalHostGetBmhOperationalState(t *testing.T) {
	testCases := []struct {
		testBmHost    *BmhBuilder
		expectedState bmhv1alpha1.OperationalStatus
	}{
		{
			testBmHost:    buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedState: bmhv1alpha1.OperationalStatusOK,
		},
		{
			testBmHost:    buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedState: "",
		},
		{
			testBmHost:    buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedState: "",
		},
	}

	for _, testCase := range testCases {
		bmhOperationalState := testCase.testBmHost.GetBmhOperationalState()
		assert.Equal(t, testCase.expectedState, bmhOperationalState)
	}
}

func TestBareMetalHostGetBmhPowerOnStatus(t *testing.T) {
	testCases := []struct {
		testBmHost    *BmhBuilder
		expectedState bool
	}{
		{
			testBmHost:    buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedState: true,
		},
		{
			testBmHost:    buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedState: false,
		},
		{
			testBmHost:    buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedState: false,
		},
	}

	for _, testCase := range testCases {
		bmhPowerOnStatus := testCase.testBmHost.GetBmhPowerOnStatus()
		assert.Equal(t, testCase.expectedState, bmhPowerOnStatus)
	}
}

func TestBareMetalHostCreateAndWaitUntilProvisioned(t *testing.T) {
	testCases := []struct {
		testBmHost    *BmhBuilder
		expectedError error
	}{
		{
			testBmHost:    buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testBmHost:    buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("context deadline exceeded"),
		},
		{
			testBmHost:    buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError: fmt.Errorf("not acceptable 'bootMode' value"),
		},
	}

	for _, testCase := range testCases {
		ipAddressPoolBuilder, err := testCase.testBmHost.CreateAndWaitUntilProvisioned(1 * time.Millisecond)
		if testCase.expectedError == nil {
			assert.Nil(t, err)
			assert.Equal(t, ipAddressPoolBuilder.Definition.Name, ipAddressPoolBuilder.Object.Name)
			assert.Equal(t, ipAddressPoolBuilder.Definition.Namespace, ipAddressPoolBuilder.Object.Namespace)
		} else {
			assert.Equal(t, err.Error(), testCase.expectedError.Error())
		}
	}
}

func TestBareMetalHostWaitUntilProvisioned(t *testing.T) {
	testCases := []struct {
		testBmHost    *BmhBuilder
		expectedError error
	}{
		{
			testBmHost:    buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testBmHost:    buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject(bmhv1alpha1.StateDeprovisioning)),
			expectedError: fmt.Errorf("context deadline exceeded"),
		},
		{
			testBmHost:    buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("context deadline exceeded"),
		},
		{
			testBmHost:    buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError: fmt.Errorf("not acceptable 'bootMode' value"),
		},
	}

	for _, testCase := range testCases {
		err := testCase.testBmHost.WaitUntilProvisioned(1 * time.Millisecond)
		if testCase.expectedError != nil {
			assert.Equal(t, err.Error(), testCase.expectedError.Error())
		} else {
			assert.Nil(t, err)
		}
	}
}

func TestBareMetalHostWaitUntilProvisioning(t *testing.T) {
	testCases := []struct {
		testBmHost    *BmhBuilder
		expectedError error
	}{
		{
			testBmHost:    buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject(bmhv1alpha1.StateProvisioning)),
			expectedError: nil,
		},
		{
			testBmHost:    buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError: fmt.Errorf("context deadline exceeded"),
		},
		{
			testBmHost:    buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("context deadline exceeded"),
		},
		{
			testBmHost:    buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError: fmt.Errorf("not acceptable 'bootMode' value"),
		},
	}

	for _, testCase := range testCases {
		err := testCase.testBmHost.WaitUntilProvisioning(1 * time.Millisecond)
		if testCase.expectedError != nil {
			assert.Equal(t, err.Error(), testCase.expectedError.Error())
		} else {
			assert.Nil(t, err)
		}
	}
}

func TestBareMetalHostWaitUntilReady(t *testing.T) {
	testCases := []struct {
		testBmHost    *BmhBuilder
		expectedError error
	}{
		{
			testBmHost:    buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject(bmhv1alpha1.StateReady)),
			expectedError: nil,
		},
		{
			testBmHost:    buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError: fmt.Errorf("context deadline exceeded"),
		},
		{
			testBmHost:    buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("context deadline exceeded"),
		},
		{
			testBmHost:    buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError: fmt.Errorf("not acceptable 'bootMode' value"),
		},
	}

	for _, testCase := range testCases {
		err := testCase.testBmHost.WaitUntilReady(1 * time.Millisecond)
		if testCase.expectedError != nil {
			assert.Equal(t, err.Error(), testCase.expectedError.Error())
		} else {
			assert.Nil(t, err)
		}
	}
}

func TestBareMetalHostWaitUntilAvailable(t *testing.T) {
	testCases := []struct {
		testBmHost    *BmhBuilder
		expectedError error
	}{
		{
			testBmHost:    buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject(bmhv1alpha1.StateAvailable)),
			expectedError: nil,
		},
		{
			testBmHost:    buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError: fmt.Errorf("context deadline exceeded"),
		},
		{
			testBmHost:    buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("context deadline exceeded"),
		},
		{
			testBmHost:    buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError: fmt.Errorf("not acceptable 'bootMode' value"),
		},
	}

	for _, testCase := range testCases {
		err := testCase.testBmHost.WaitUntilAvailable(1 * time.Millisecond)
		if testCase.expectedError != nil {
			assert.Equal(t, err.Error(), testCase.expectedError.Error())
		} else {
			assert.Nil(t, err)
		}
	}
}

func TestBareMetalHostWaitUntilInStatus(t *testing.T) {
	testCases := []struct {
		testBmHost    *BmhBuilder
		expectedError error
	}{
		{
			testBmHost:    buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject(bmhv1alpha1.StateAvailable)),
			expectedError: nil,
		},
		{
			testBmHost:    buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject(bmhv1alpha1.StateProvisioning)),
			expectedError: fmt.Errorf("context deadline exceeded"),
		},
		{
			testBmHost:    buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError: fmt.Errorf("context deadline exceeded"),
		},
		{
			testBmHost:    buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("context deadline exceeded"),
		},
		{
			testBmHost:    buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError: fmt.Errorf("not acceptable 'bootMode' value"),
		},
	}

	for _, testCase := range testCases {
		err := testCase.testBmHost.WaitUntilInStatus(bmhv1alpha1.StateAvailable, 1*time.Millisecond)
		if testCase.expectedError != nil {
			assert.Equal(t, err.Error(), testCase.expectedError.Error())
		} else {
			assert.Nil(t, err)
		}
	}
}

func TestBareMetalHostDeleteAndWaitUntilDeleted(t *testing.T) {
	testCases := []struct {
		testBmHost    *BmhBuilder
		expectedError error
	}{
		{
			testBmHost:    buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testBmHost:    buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("bmh cannot be deleted because it does not exist"),
		},
		{
			testBmHost:    buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError: fmt.Errorf("not acceptable 'bootMode' value"),
		},
	}

	for _, testCase := range testCases {
		builder, err := testCase.testBmHost.DeleteAndWaitUntilDeleted(1 * time.Second)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testBmHost.Object)
			assert.Nil(t, builder)
		}
	}
}

func TestBareMetalHostWaitUntilDeleted(t *testing.T) {
	testCases := []struct {
		testBmHost    *BmhBuilder
		expectedError error
	}{
		{
			testBmHost:    buildValidBmHostBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
		{
			testBmHost:    buildValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError: context.DeadlineExceeded,
		},
		{
			testBmHost:    buildInValidBmHostBuilder(buildBareMetalHostTestClientWithDummyObject()),
			expectedError: fmt.Errorf("not acceptable 'bootMode' value"),
		},
	}

	for _, testCase := range testCases {
		err := testCase.testBmHost.WaitUntilDeleted(1 * time.Second)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testBmHost.Object)
		}
	}
}

func buildValidBmHostBuilder(apiClient *clients.Settings) *BmhBuilder {
	return NewBuilder(
		apiClient,
		defaultBmHostName,
		defaultBmHostNsName,
		defaultBmHostAddress,
		defaultBmHostSecretName,
		defaultBmHostMacAddress,
		defaultBmHostBootMode,
	)
}

func buildInValidBmHostBuilder(apiClient *clients.Settings) *BmhBuilder {
	return NewBuilder(
		apiClient,
		defaultBmHostName,
		defaultBmHostNsName,
		defaultBmHostAddress,
		defaultBmHostSecretName,
		defaultBmHostMacAddress,
		"test",
	)
}

func buildBareMetalHostTestClientWithDummyObject(state ...bmhv1alpha1.ProvisioningState) *clients.Settings {
	provisionState := bmhv1alpha1.StateProvisioned
	if len(state) > 0 {
		provisionState = state[0]
	}

	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: buildDummyBmHost(provisionState),
	})
}

func buildDummyBmHost(state bmhv1alpha1.ProvisioningState) []runtime.Object {
	return append([]runtime.Object{}, &bmhv1alpha1.BareMetalHost{
		Spec: bmhv1alpha1.BareMetalHostSpec{
			BMC: bmhv1alpha1.BMCDetails{
				Address:                        defaultBmHostAddress,
				CredentialsName:                defaultBmHostSecretName,
				DisableCertificateVerification: true,
			},
			BootMode:              bmhv1alpha1.BootMode(defaultBmHostBootMode),
			BootMACAddress:        defaultBmHostMacAddress,
			Online:                true,
			ExternallyProvisioned: false,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultBmHostName,
			Namespace: defaultBmHostNsName,
		},
		Status: bmhv1alpha1.BareMetalHostStatus{
			OperationalStatus: bmhv1alpha1.OperationalStatusOK,
			PoweredOn:         true,
			Provisioning: bmhv1alpha1.ProvisionStatus{
				State: state,
			},
		},
	})
}
