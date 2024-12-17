package bmc

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stmcginnis/gofish/redfish"
	"github.com/stretchr/testify/assert"
)

//go:embed testdata/redfish_v1.json
var redfishRootJSONResponse string

//go:embed testdata/redfish_v1_systems.json
var redfishSystemsJSONResponse string

//go:embed testdata/redfish_v1_system.json
var redfishSystemJSONResponse string

//go:embed testdata/redfish_v1_system_secureboot_disabled.json
var redfishSystemSecureBootDisabledJSONResponse string

//go:embed testdata/redfish_v1_system_secureboot_enabled.json
var redfishSystemSecureBootEnabledJSONResponse string

//go:embed testdata/redfish_v1_chassiscollection.json
var redfishChassisCollectionJSONResponse string

//go:embed testdata/redfish_v1_chassis.json
var redfishChassisJSONResponse string

// redfishChassisNoPowerJSONResponse is the response a chassis that does not contain a power link.
//
//go:embed testdata/redfish_v1_chassis_nopower.json
var redfishChassisNoPowerJSONResponse string

//go:embed testdata/redfish_v1_power.json
var redfishPowerJSONResponse string

//go:embed testdata/redfish_v1_system_boot_options.json
var redfishSystemBootOptionsJSONResponse string

//go:embed testdata/redfish_v1_system_boot_option_Boot0000.json
var redfishSystemBootOption0000JSONResponse string

//go:embed testdata/redfish_v1_system_boot_option_Boot0003.json
var redfishSystemBootOption0003JSONResponse string

// redfishAuth is used to unmarshall the received login request redfish credentials.
type redfishAuth struct {
	UserName string
	Password string
}

type redfishAPIResponseCallbacks struct {
	v1          func(r *http.Request)
	sessions    func(r *http.Request)
	system      func(r *http.Request)
	secureBoot  func(r *http.Request)
	bootOptions func(r *http.Request)
	chassis     func(r *http.Request)
	power       func(r *http.Request)
}

const (
	// These are valid defaults to use in test cases.
	defaultHost     = "1.2.3.4"
	defaultUsername = "user1"
	defaultPassword = "pass1"

	//nolint:lll // If the literal is broken in two parts with "+" it will be flagged with goconst...
	secureBootFailFmt = "failed to get secure boot: failed to get redfish system: invalid system index %d (base-index=0, num systems=1)"
)

func TestBMCNew(t *testing.T) {
	testCases := []struct {
		name           string
		host           string
		expectedErrMsg string
	}{
		{
			name:           "empty host",
			host:           "",
			expectedErrMsg: "bmc 'host' cannot be empty",
		},
		{
			name:           "valid host",
			host:           defaultHost,
			expectedErrMsg: "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			bmc := New(testCase.host)

			assert.Equal(t, testCase.expectedErrMsg, bmc.errorMsg)

			if testCase.expectedErrMsg == "" {
				assert.Equal(t, testCase.host, bmc.host)
			}
		})
	}
}

func TestBMCWithRedfishUser(t *testing.T) {
	testCases := []struct {
		name           string
		username       string
		password       string
		expectedErrMsg string
	}{
		{
			name:           "everything alright",
			username:       defaultUsername,
			password:       defaultPassword,
			expectedErrMsg: "",
		},
		{
			name:           "all params empty",
			username:       "",
			password:       "",
			expectedErrMsg: "redfish 'username' cannot be empty",
		},
		{
			name:           "username empty",
			username:       "",
			password:       defaultPassword,
			expectedErrMsg: "redfish 'username' cannot be empty",
		},
		{
			name:           "password empty",
			username:       defaultUsername,
			password:       "",
			expectedErrMsg: "redfish 'password' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			bmc := New(defaultHost).WithRedfishUser(testCase.username, testCase.password)

			assert.Equal(t, testCase.expectedErrMsg, bmc.errorMsg)

			if testCase.expectedErrMsg == "" {
				assert.Equal(t, testCase.username, bmc.redfishUser.Name)
				assert.Equal(t, testCase.password, bmc.redfishUser.Password)
			}
		})
	}
}

func TestBMCWithRedfishTimeout(t *testing.T) {
	testCases := []struct {
		name           string
		timeout        time.Duration
		expectedErrMsg string
	}{
		{
			name:           "everything alright",
			timeout:        defaultTimeOut,
			expectedErrMsg: "",
		},
		{
			name:           "zero timeout",
			timeout:        0,
			expectedErrMsg: "redfish 'timeout' cannot be less than or equal to zero",
		},
		{
			name:           "negative timeout",
			timeout:        -1 * time.Minute,
			expectedErrMsg: "redfish 'timeout' cannot be less than or equal to zero",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			bmc := New(defaultHost).WithRedfishTimeout(testCase.timeout)

			assert.Equal(t, testCase.expectedErrMsg, bmc.errorMsg)

			if testCase.expectedErrMsg == "" {
				assert.Equal(t, testCase.timeout, bmc.timeOuts.Redfish)
			}
		})
	}
}

func TestBMCWithRedfishSystemIndex(t *testing.T) {
	testCases := []struct {
		name           string
		index          int
		expectedErrMsg string
	}{
		{
			name:           "everything alright",
			index:          1,
			expectedErrMsg: "",
		},
		{
			name:           "negative index",
			index:          -1,
			expectedErrMsg: "redfish 'systemIndex' cannot be negative",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			bmc := New(defaultHost).WithRedfishSystemIndex(testCase.index)

			assert.Equal(t, testCase.expectedErrMsg, bmc.errorMsg)

			if testCase.expectedErrMsg == "" {
				assert.Equal(t, testCase.index, bmc.systemIndex)
			}
		})
	}
}

func TestBMCWithRedfishPowerControlIndex(t *testing.T) {
	testCases := []struct {
		name           string
		index          int
		expectedErrMsg string
	}{
		{
			name:           "everything alright",
			index:          1,
			expectedErrMsg: "",
		},
		{
			name:           "negative index",
			index:          -1,
			expectedErrMsg: "redfish 'powerControlIndex' cannot be negative",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			bmc := New(defaultHost).WithRedfishPowerControlIndex(testCase.index)

			assert.Equal(t, testCase.expectedErrMsg, bmc.errorMsg)

			if testCase.expectedErrMsg == "" {
				assert.Equal(t, testCase.index, bmc.powerControlIndex)
			}
		})
	}
}

func TestBMCWithSSHUser(t *testing.T) {
	testCases := []struct {
		name           string
		username       string
		password       string
		expectedErrMsg string
	}{
		{
			name:           "everything alright",
			username:       defaultUsername,
			password:       defaultPassword,
			expectedErrMsg: "",
		},
		{
			name:           "all params empty",
			username:       "",
			password:       "",
			expectedErrMsg: "ssh 'username' cannot be empty",
		},
		{
			name:           "username empty",
			username:       "",
			password:       defaultPassword,
			expectedErrMsg: "ssh 'username' cannot be empty",
		},
		{
			name:           "password empty",
			username:       defaultUsername,
			password:       "",
			expectedErrMsg: "ssh 'password' cannot be empty",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			bmc := New(defaultHost).WithSSHUser(testCase.username, testCase.password)

			assert.Equal(t, testCase.expectedErrMsg, bmc.errorMsg)

			if testCase.expectedErrMsg == "" {
				assert.Equal(t, testCase.username, bmc.sshUser.Name)
				assert.Equal(t, testCase.password, bmc.sshUser.Password)
			}
		})
	}
}
func TestBMCWithSSHPort(t *testing.T) {
	testCases := []struct {
		name           string
		port           uint16
		expectedErrMsg string
	}{
		{
			name:           "everything alright",
			port:           1234,
			expectedErrMsg: "",
		},
		{
			name:           "port zero",
			port:           0,
			expectedErrMsg: "ssh 'port' cannot be zero",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			bmc := New(defaultHost).WithSSHPort(testCase.port)

			assert.Equal(t, testCase.expectedErrMsg, bmc.errorMsg)

			if testCase.expectedErrMsg == "" {
				assert.Equal(t, testCase.port, bmc.sshPort)
			}
		})
	}
}

func TestBMCWithSSHTimeout(t *testing.T) {
	testCases := []struct {
		name           string
		timeout        time.Duration
		expectedErrMsg string
	}{
		{
			name:           "everything alright",
			timeout:        defaultTimeOut,
			expectedErrMsg: "",
		},
		{
			name:           "zero timeout",
			timeout:        0,
			expectedErrMsg: "ssh 'timeout' cannot be less than or equal to zero",
		},
		{
			name:           "negative timeout",
			timeout:        -1 * time.Minute,
			expectedErrMsg: "ssh 'timeout' cannot be less than or equal to zero",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			bmc := New(defaultHost).WithSSHTimeout(testCase.timeout)

			assert.Equal(t, testCase.expectedErrMsg, bmc.errorMsg)

			if testCase.expectedErrMsg == "" {
				assert.Equal(t, testCase.timeout, bmc.timeOuts.Redfish)
			}
		})
	}
}

func TestBMCSystemManufacturer(t *testing.T) {
	respCallbacks := redfishAPIResponseCallbacks{}

	// We will check user and password received by the connect/login to be
	// the ones we're using.
	user := User{}
	respCallbacks.sessions = getAuthDataCallbackFn(t, &user)

	redfishServer := createFakeRedfishLocalServer(false, respCallbacks)
	defer redfishServer.Close()

	host := strings.Split(redfishServer.URL, "//")[1]
	bmc := New(host).WithRedfishUser(defaultUsername, defaultPassword)

	// Get the manufacturer of system index = 0
	const expectedManufacturer = "Dell Inc."

	manufacturer, err := bmc.SystemManufacturer()
	assert.NoError(t, err, "Failed to get system manufacturer")

	// Check the credentials were the ones we used.
	assert.Equal(t, defaultUsername, user.Name, "Wrong auth username received in Redfish server")
	assert.Equal(t, defaultPassword, user.Password, "Wrong auth password received in Redfish server")

	assert.Equal(t, expectedManufacturer, manufacturer)

	// Try getting the manufacturer of a non-existent system (e.g. index 1).
	const expectedErrMsg = "failed to get redfish system: invalid system index 1 (base-index=0, num systems=1)"

	_, err = bmc.WithRedfishSystemIndex(1).SystemManufacturer()
	assert.EqualError(t, err, expectedErrMsg)
}

func TestBMCManufacturerTimeout(t *testing.T) {
	respCallbacks := redfishAPIResponseCallbacks{}

	// We'll simulate a 200ms delay in the response to one of the rest endpoints
	respCallbacks.sessions = getDelayResponseCallbackFn(t, 200*time.Millisecond)

	redfishServer := createFakeRedfishLocalServer(false, respCallbacks)
	defer redfishServer.Close()

	host := strings.Split(redfishServer.URL, "//")[1]
	bmc := New(host).WithRedfishUser(defaultUsername, defaultPassword).WithRedfishTimeout(100 * time.Millisecond)

	// Get Manufacturer. Since we've force a response greater than the configured timeout, we
	// should get an error.
	const expectedTimeoutErrMsgRegex = `redfish connection error: failed to connect to redfish endpoint: ` +
		`Post "https://127\.0\.0\.1:\d+/redfish/v1/SessionService/Sessions": context deadline exceeded`

	regex := regexp.MustCompile(expectedTimeoutErrMsgRegex)

	_, err := bmc.SystemManufacturer()
	assert.Errorf(t, err, "No error found, expected err regexp: %v", expectedTimeoutErrMsgRegex)

	errMsg := err.Error()
	match := regex.Find([]byte(errMsg))

	assert.NotEmptyf(t, match, "Error did not match. Expected regexp: %v, Got: %s", expectedTimeoutErrMsgRegex, errMsg)
}

func TestBMCSecureBootStatus(t *testing.T) {
	redfishServer := createFakeRedfishLocalServer(false, redfishAPIResponseCallbacks{})

	host := strings.Split(redfishServer.URL, "//")[1]
	bmc := New(host).WithRedfishUser(defaultUsername, defaultPassword)

	// Get twice the SecureBoot status, which should be false.
	expectedSecureBootStatus := false

	sbStatus, err := bmc.IsSecureBootEnabled()
	assert.NoError(t, err, "Failed to get secure boot status")
	assert.Equal(t, expectedSecureBootStatus, sbStatus)

	// Try getting the secureboot status from a non-existent system (e.g index 2)
	expectedErrMsg := fmt.Sprintf(secureBootFailFmt, 2)
	bmc = bmc.WithRedfishSystemIndex(2)

	_, err = bmc.IsSecureBootEnabled()
	assert.EqualError(t, err, expectedErrMsg)

	redfishServer.Close()

	// Now let's create another fake redfish server where secureboot is enabled.
	redfishServer = createFakeRedfishLocalServer(true, redfishAPIResponseCallbacks{})
	defer redfishServer.Close()

	host = strings.Split(redfishServer.URL, "//")[1]
	bmc = New(host).WithRedfishUser(defaultUsername, defaultPassword)

	// Get twice the SecureBoot status, which should be true.
	expectedSecureBootStatus = true

	sbStatus, err = bmc.IsSecureBootEnabled()
	assert.NoError(t, err, "Failed to get secure boot status")
	assert.Equal(t, expectedSecureBootStatus, sbStatus)
}

func TestBMCSecureBootEnable(t *testing.T) {
	// Create de fake redfish api endpoint with secureBoot "disabled"
	redfishServer := createFakeRedfishLocalServer(false, redfishAPIResponseCallbacks{})
	defer redfishServer.Close()

	host := strings.Split(redfishServer.URL, "//")[1]
	bmc := New(host).WithRedfishUser(defaultUsername, defaultPassword)

	// Secure boot is already disabled, so we should get an error if we try to disable it again.
	const expectedErrorMsg = "secure boot is already disabled"

	err := bmc.SecureBootDisable()
	if err.Error() != expectedErrorMsg {
		t.Errorf("Unexpected error when disabling secure boot. Want: %v, Got: %v", expectedErrorMsg, err)
	}

	if err := bmc.SecureBootEnable(); err != nil {
		t.Errorf("Unexpected error found when enabling secure boot: %v", err)
	}

	// Try enabling the secureboot status from a non-existent system (e.g index 2)
	expectedErrMsg := fmt.Sprintf(secureBootFailFmt, 2)

	err = bmc.WithRedfishSystemIndex(2).SecureBootEnable()
	if err.Error() != expectedErrMsg {
		t.Errorf("Unexpected error when getting manufacturer of non-existent system. Want: %v, Got: %v", expectedErrMsg, err)
	}
}

func TestBMCSecureBootDisable(t *testing.T) {
	// Create de fake redfish api endpoint with secureBoot "enabled"
	redfishServer := createFakeRedfishLocalServer(true, redfishAPIResponseCallbacks{})
	defer redfishServer.Close()

	host := strings.Split(redfishServer.URL, "//")[1]
	bmc := New(host).WithRedfishUser(defaultUsername, defaultPassword)

	// Secure boot is already disabled, so we should get an error if we try to disable it again.
	const expectedErrorMsg = "secure boot is already enabled"

	err := bmc.SecureBootEnable()
	if err.Error() != expectedErrorMsg {
		t.Errorf("Unexpected error when disabling secure boot. Want: %v, Got: %v", expectedErrorMsg, err)
	}

	if err := bmc.SecureBootDisable(); err != nil {
		t.Errorf("Unexpected error found when enabling secure boot: %v", err)
	}

	// Try enabling the secureboot status from a non-existent system (e.g index 2)
	const expectedErrMsg = "failed to get secure boot: failed to get redfish system: " +
		"invalid system index 2 (base-index=0, num systems=1)"

	err = bmc.WithRedfishSystemIndex(2).SecureBootEnable()
	if err.Error() != expectedErrMsg {
		t.Errorf("Unexpected error when getting manufacturer of non-existent system. Want: %v, Got: %v", expectedErrMsg, err)
	}
}

func TestBMCSystemResetAction(t *testing.T) {
	resetActions := []redfish.ResetType{
		redfish.OnResetType,
		redfish.ForceOnResetType,
		redfish.ForceOffResetType,
		redfish.ForceRestartResetType,
		redfish.GracefulRestartResetType,
		redfish.GracefulShutdownResetType,
		redfish.PushPowerButtonResetType,
		redfish.NmiResetType,
		redfish.PauseResetType,
		redfish.ResumeResetType,
		redfish.SuspendResetType,
	}

	for _, resetAction := range resetActions {
		testResetAction(t, string(resetAction), func(bmc *BMC) error {
			return bmc.SystemResetAction(resetAction)
		})
	}
}

func TestBMCSystemForceReset(t *testing.T) {
	testResetAction(t, "ForceReset", func(bmc *BMC) error {
		return bmc.SystemForceReset()
	})
}

func TestBMCSystemGracefulShutdown(t *testing.T) {
	testResetAction(t, "GracefulShutdown", func(bmc *BMC) error {
		return bmc.SystemGracefulShutdown()
	})
}

func TestBMCSystemPowerOn(t *testing.T) {
	testResetAction(t, "PowerOn", func(bmc *BMC) error {
		return bmc.SystemPowerOn()
	})
}

func TestBMCSystemPowerOff(t *testing.T) {
	testResetAction(t, "PowerOff", func(bmc *BMC) error {
		return bmc.SystemPowerOff()
	})
}

func TestBMCSystemPowerCycle(t *testing.T) {
	testResetAction(t, "PowerCycle", func(bmc *BMC) error {
		return bmc.SystemPowerCycle()
	})
}

func TestBMCSystemPowerState(t *testing.T) {
	// Create fake redfish endpoint.
	redfishServer := createFakeRedfishLocalServer(false, redfishAPIResponseCallbacks{})
	defer redfishServer.Close()

	host := strings.Split(redfishServer.URL, "//")[1]
	bmc := New(host).WithRedfishUser(defaultUsername, defaultPassword)

	const expectedPowerState = "On"

	powerState, err := bmc.SystemPowerState()
	assert.NoError(t, err)
	assert.Equal(t, expectedPowerState, powerState)
}

func TestBMCPowerUsage(t *testing.T) {
	// Create a fake redfish api endpoint with secureBoot "disabled"
	redfishServer := createFakeRedfishLocalServer(false, redfishAPIResponseCallbacks{})
	defer redfishServer.Close()

	host := strings.Split(redfishServer.URL, "//")[1]
	bmc := New(host).WithRedfishUser(defaultUsername, defaultPassword)

	const expectedPowerUsage float32 = 360.0

	power, err := bmc.PowerUsage()
	assert.NoError(t, err)
	assert.Equal(t, expectedPowerUsage, power)
}

func TestBMCCreateCLISSHSession(t *testing.T) {
	bmc := New(defaultHost).WithRedfishUser(defaultUsername, defaultPassword)

	// Check that the session creation fails with no SSH user.
	expectedErrMsg := "cannot access ssh with nil user"

	session, err := bmc.createCLISSHClient()
	assert.Nil(t, session)
	assert.EqualError(t, err, expectedErrMsg)

	// Now set SSH user to test when session created with SSH user. Also set the timeout so this test fails quickly.
	bmc = bmc.WithSSHUser(defaultUsername, defaultPassword).WithSSHTimeout(10 * time.Millisecond)

	expectedErrMsg = "failed to connect to BMC's SSH server: dial tcp 1.2.3.4:22: i/o timeout"

	session, err = bmc.createCLISSHClient()
	assert.Nil(t, session)
	assert.EqualError(t, err, expectedErrMsg)
}

func TestBMCRunCLICommand(t *testing.T) {
	bmc := New(defaultHost).WithSSHUser(defaultUsername, defaultPassword).WithSSHTimeout(10 * time.Millisecond)

	const expectedErrMsg = `failed to connect to CLI: failed to connect to BMC's SSH server: ` +
		`dial tcp 1.2.3.4:22: i/o timeout`

	_, _, err := bmc.RunCLICommand("help", false, 5*time.Second)
	assert.EqualError(t, err, expectedErrMsg)
}

func TestBMCSerialConsole(t *testing.T) {
	bmc := New(defaultHost).
		WithRedfishUser(defaultUsername, defaultPassword).
		WithSSHUser(defaultUsername, defaultPassword).
		WithRedfishTimeout(10 * time.Millisecond).WithSSHTimeout(10 * time.Millisecond)

	var expectedErrMsg = `failed to create underlying ssh session for 1.2.3.4: ` +
		`failed to connect to BMC's SSH server: dial tcp 1.2.3.4:22: i/o timeout`

	_, _, err := bmc.OpenSerialConsole("console com2")
	assert.EqualError(t, err, expectedErrMsg)

	// This can sometimes return two different messages, but both signify a timeout as expected.
	expectedErrMsg = `failed to get redfish system manufacturer for 1.2.3.4: redfish connection error: ` +
		`failed to connect to redfish endpoint: Get "https://1.2.3.4/redfish/v1/": `
	deadlineErrMsg := "context deadline exceeded"
	timeoutErrMsg := "dial tcp 1.2.3.4:443: i/o timeout"

	// Test without cli command. A best effort is made to open it based on system's manufacturer.
	_, _, err = bmc.OpenSerialConsole("")
	assert.Contains(t, []string{expectedErrMsg + deadlineErrMsg, expectedErrMsg + timeoutErrMsg}, err.Error())

	expectedErrMsg = "no underlying ssh session found for 1.2.3.4"

	err = bmc.CloseSerialConsole()
	assert.EqualError(t, err, expectedErrMsg)
}

func TestBMCSystemBootOptions(t *testing.T) {
	// Create fake redfish endpoint.
	redfishServer := createFakeRedfishLocalServer(false, redfishAPIResponseCallbacks{})
	defer redfishServer.Close()

	host := strings.Split(redfishServer.URL, "//")[1]
	bmc := New(host).WithRedfishUser(defaultUsername, defaultPassword).WithRedfishTimeout(1 * time.Hour)

	expectedOptions := map[string]string{
		"Boot0000": "PXE Device 1: Embedded NIC 1 Port 1 Partition 1",
		"Boot0003": "RAID Controller in SL 3: Red Hat Enterprise Linux",
	}

	bootOptions, err := bmc.SystemBootOptions()
	assert.NoError(t, err)
	assert.Equal(t, expectedOptions, bootOptions)
}

func TestBMCSystemBootOrderReferences(t *testing.T) {
	// Create fake redfish endpoint.
	redfishServer := createFakeRedfishLocalServer(false, redfishAPIResponseCallbacks{})
	defer redfishServer.Close()

	host := strings.Split(redfishServer.URL, "//")[1]
	bmc := New(host).WithRedfishUser(defaultUsername, defaultPassword)

	expectedBootOrder := []string{"Boot0003", "Boot0000"}

	bootOrder, err := bmc.SystemBootOrderReferences()
	assert.NoError(t, err)
	assert.Equal(t, expectedBootOrder, bootOrder)
}

func TestBMCSetSystemBootOrderReferences(t *testing.T) {
	// Create fake redfish endpoint.
	redfishServer := createFakeRedfishLocalServer(false, redfishAPIResponseCallbacks{})
	defer redfishServer.Close()

	host := strings.Split(redfishServer.URL, "//")[1]
	bmc := New(host).WithRedfishUser(defaultUsername, defaultPassword)

	// Read the current boot order.
	expectedBootOrder := []string{"Boot0003", "Boot0000"}

	bootOrder, err := bmc.SystemBootOrderReferences()
	assert.NoError(t, err)
	assert.Equal(t, expectedBootOrder, bootOrder)

	// Switch first and second boot references and remove the last two
	newBootOrder := []string{expectedBootOrder[1], expectedBootOrder[0]}

	err = bmc.SetSystemBootOrderReferences(newBootOrder)
	assert.NoError(t, err)
}

func getDelayResponseCallbackFn(t *testing.T, respDelay time.Duration) func(r *http.Request) {
	t.Helper()

	return func(*http.Request) {
		time.Sleep(respDelay)
	}
}

func getAuthDataCallbackFn(t *testing.T, user *User) func(r *http.Request) {
	t.Helper()

	return func(r *http.Request) {
		buff, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}

		auth := redfishAuth{}

		err = json.Unmarshal(buff, &auth)
		if err != nil {
			t.Errorf("Failed to unmarshal redfish auth data: %v", err)
		} else {
			user.Name = auth.UserName
			user.Password = auth.Password
		}
	}
}

// Helper function that creates a fake redfish REST server in localhost (random port). When outputAuthData is provided,
// it will be filled with the auth credentials received in the login request. All the responses, except the login one,
// are sent using static json data from the testdata folder. The flag secureBootEnable is used to load the json response
// for the secure boot api depending on wether we want it to be enabled or disabled for our test.
func createFakeRedfishLocalServer(secureBootEnabled bool, callbacks redfishAPIResponseCallbacks) *httptest.Server { //nolint:funlen,lll
	sbEnabled := secureBootEnabled
	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if callbacks.v1 != nil {
			callbacks.v1(r)
		}

		_, _ = w.Write([]byte(redfishRootJSONResponse))
	}))

	mux.HandleFunc("/redfish/v1/SessionService/Sessions",
		http.HandlerFunc(func(writer http.ResponseWriter, reader *http.Request) {
			if callbacks.sessions != nil {
				callbacks.sessions(reader)
			}

			// fake empty response
			_, _ = writer.Write([]byte("{}"))
		}))

	mux.HandleFunc("/redfish/v1/Systems", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if callbacks.system != nil {
			callbacks.system(r)
		}

		_, _ = w.Write([]byte(redfishSystemsJSONResponse))
	}))

	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if callbacks.secureBoot != nil {
			callbacks.secureBoot(r)
		}

		_, _ = w.Write([]byte(redfishSystemJSONResponse))
	}))

	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/SecureBoot",
		http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)

			if sbEnabled {
				_, _ = writer.Write([]byte(redfishSystemSecureBootEnabledJSONResponse))
			} else {
				_, _ = writer.Write([]byte(redfishSystemSecureBootDisabledJSONResponse))
			}
		}))

	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/BootOptions",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if callbacks.bootOptions != nil {
				callbacks.bootOptions(r)
			}

			_, _ = w.Write([]byte(redfishSystemBootOptionsJSONResponse))
		}))

	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/BootOptions/Boot0000",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(redfishSystemBootOption0000JSONResponse))
		}))

	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/BootOptions/Boot0003",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(redfishSystemBootOption0003JSONResponse))
		}))

	mux.HandleFunc("GET /redfish/v1/Chassis", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if callbacks.chassis != nil {
			callbacks.chassis(r)
		}

		_, _ = w.Write([]byte(redfishChassisCollectionJSONResponse))
	}))

	mux.HandleFunc("GET /redfish/v1/Chassis/System.Embedded.1",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(redfishChassisJSONResponse))
		}))

	mux.HandleFunc("GET /redfish/v1/Chassis/Enclosure.Internal.0-1",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(redfishChassisNoPowerJSONResponse))
		}))

	mux.HandleFunc("GET /redfish/v1/Chassis/System.Embedded.1/Power",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if callbacks.power != nil {
				callbacks.power(r)
			}

			_, _ = w.Write([]byte(redfishPowerJSONResponse))
		}))

	redfishServer := httptest.NewUnstartedServer(mux)
	redfishServer.EnableHTTP2 = true
	redfishServer.StartTLS()

	return redfishServer
}

// testResetAction performs unit testing for a provided function that performs a reset action on the BMC.
func testResetAction(t *testing.T, name string, resetFunction func(bmc *BMC) error) {
	t.Helper()

	// Create a fake redfish api endpoint with secureBoot "disabled"
	redfishServer := createFakeRedfishLocalServer(false, redfishAPIResponseCallbacks{})
	defer redfishServer.Close()

	host := strings.Split(redfishServer.URL, "//")[1]

	testCases := []struct {
		name              string
		systemIndex       int
		expectedErrorText string
	}{
		{
			name:              fmt.Sprintf("%s valid system index", name),
			systemIndex:       0,
			expectedErrorText: "",
		},
		{
			name:              fmt.Sprintf("%s invalid system index", name),
			systemIndex:       2,
			expectedErrorText: "failed to get redfish system: invalid system index 2 (base-index=0, num systems=1)",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			bmc := New(host).WithRedfishUser(defaultUsername, defaultPassword).WithRedfishSystemIndex(testCase.systemIndex)

			expectedErrorTest := testCase.expectedErrorText

			err := resetFunction(bmc)

			if testCase.expectedErrorText != "" {
				if name == "PowerCycle" {
					// PowerCycle is special as it needs to get the supported reset types first and it will also fail.
					expectedErrorTest = "failed to get system's supported reset types: " + testCase.expectedErrorText
				}

				assert.EqualError(t, err, expectedErrorTest)
			}
		})
	}
}
