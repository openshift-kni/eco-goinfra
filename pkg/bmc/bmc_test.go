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

// redfishAuth is used to unmarshall the received login request redfish credentials.
type redfishAuth struct {
	UserName string
	Password string
}

type redfishAPIResponseCallbacks struct {
	v1         func(r *http.Request)
	sessions   func(r *http.Request)
	system     func(r *http.Request)
	secureBoot func(r *http.Request)
	chassis    func(r *http.Request)
	power      func(r *http.Request)
}

const (
	defaultSSHPort = 22

	//nolint:lll // If the literal is broken in two parts with "+" it will be flagged with goconst...
	secureBootFailFmt = "failed to get secure boot: failed to get redfish system: invalid system index %d (base-index=0, num systems=1)"
)

var (
	validUser     = User{"user1", "pass1"}
	validTimeOuts = TimeOuts{Redfish: 1 * time.Minute, SSH: 15 * time.Second}
)

//nolint:funlen
func TestBMCNew(t *testing.T) {
	testCases := []struct {
		name           string
		host           string
		redfishUser    User
		sshUser        User
		sshPort        uint16
		tiemeouts      TimeOuts
		expectedErrMsg string
	}{
		{
			name:        "All params empty",
			host:        "",
			redfishUser: User{},
			sshUser:     User{},
			sshPort:     defaultSSHPort,
			tiemeouts:   TimeOuts{},
			expectedErrMsg: "host is empty, redfish user's name is empty, redfish user's password is empty, " +
				"ssh user's name is empty, ssh user's password is empty, redfish timeout is 0, ssh timeout is 0",
		},
		{
			name:           "Everything's alright",
			host:           "1.2.3.4",
			redfishUser:    validUser,
			sshUser:        validUser,
			sshPort:        defaultSSHPort,
			tiemeouts:      validTimeOuts,
			expectedErrMsg: "",
		},
		{
			name:           "Host is empty",
			host:           "",
			redfishUser:    validUser,
			sshUser:        validUser,
			sshPort:        defaultSSHPort,
			tiemeouts:      validTimeOuts,
			expectedErrMsg: "host is empty",
		},
		{
			name:           "redfish user's name is empty",
			host:           "1.2.3.4",
			redfishUser:    User{Password: "pass1"},
			sshUser:        validUser,
			sshPort:        defaultSSHPort,
			tiemeouts:      validTimeOuts,
			expectedErrMsg: "redfish user's name is empty",
		},
		{
			name:           "redfish user's password is empty",
			host:           "1.2.3.4",
			redfishUser:    User{Name: "user2"},
			sshUser:        validUser,
			sshPort:        defaultSSHPort,
			tiemeouts:      validTimeOuts,
			expectedErrMsg: "redfish user's password is empty",
		},
		{
			name:           "ssh user's name is empty",
			host:           "1.2.3.4",
			redfishUser:    validUser,
			sshUser:        User{Password: "pass1"},
			sshPort:        defaultSSHPort,
			tiemeouts:      validTimeOuts,
			expectedErrMsg: "ssh user's name is empty",
		},
		{
			name:           "ssh user's password is empty",
			host:           "1.2.3.4",
			redfishUser:    validUser,
			sshUser:        User{Name: "user2"},
			sshPort:        defaultSSHPort,
			tiemeouts:      validTimeOuts,
			expectedErrMsg: "ssh user's password is empty",
		},
		{
			name:           "ssh port is zero",
			host:           "1.2.3.4",
			redfishUser:    validUser,
			sshUser:        validUser,
			sshPort:        0,
			tiemeouts:      validTimeOuts,
			expectedErrMsg: "ssh port is zero",
		},
		{
			name:           "invalid redfish timeout",
			host:           "1.2.3.4",
			redfishUser:    validUser,
			sshUser:        validUser,
			sshPort:        defaultSSHPort,
			tiemeouts:      TimeOuts{Redfish: 0, SSH: 1 * time.Second},
			expectedErrMsg: "redfish timeout is 0",
		},
		{
			name:           "invalid SSH timeout",
			host:           "1.2.3.4",
			redfishUser:    validUser,
			sshUser:        validUser,
			sshPort:        defaultSSHPort,
			tiemeouts:      TimeOuts{Redfish: 1 * time.Second, SSH: 0},
			expectedErrMsg: "ssh timeout is 0",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(newT *testing.T) {
			_, err := New(testCase.host, testCase.redfishUser, testCase.sshUser, testCase.sshPort, testCase.tiemeouts)
			if err != nil {
				if err.Error() != testCase.expectedErrMsg {
					newT.Errorf("Unexpected error. Got: %v, Want: %s", err, testCase.expectedErrMsg)
				}
			} else {
				if testCase.expectedErrMsg != "" {
					newT.Errorf("Error is nil. Expected error: %v", testCase.expectedErrMsg)
				}
			}
		})
	}
}

func TestBMCSetSystemIndex(t *testing.T) {
	testCases := []struct {
		index          int
		expectedErrMsg string
	}{
		{
			index:          0,
			expectedErrMsg: "",
		},
		{
			index:          5,
			expectedErrMsg: "",
		},
		{
			index:          -1,
			expectedErrMsg: "invalid index -1",
		},
		{
			index:          -5,
			expectedErrMsg: "invalid index -5",
		},
	}

	bmc, err := New("1.2.3.4", validUser, validUser, defaultSSHPort, DefaultTimeOuts)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	for _, testCase := range testCases {
		err = bmc.SetSystemIndex(testCase.index)
		if err == nil {
			if testCase.expectedErrMsg != "" {
				t.Errorf("Err is nil. Expected: %v", testCase.expectedErrMsg)
			}
		} else if err.Error() != testCase.expectedErrMsg {
			t.Errorf("Unexpected error. Want: %v, Got: %v", testCase.expectedErrMsg, err.Error())
		}
	}
}

func TestBMCSetPowerControlIndex(t *testing.T) {
	testCases := []struct {
		index          int
		expectedErrMsg string
	}{
		{
			index:          0,
			expectedErrMsg: "",
		},
		{
			index:          5,
			expectedErrMsg: "",
		},
		{
			index:          -1,
			expectedErrMsg: "invalid index -1",
		},
		{
			index:          -5,
			expectedErrMsg: "invalid index -5",
		},
	}

	for _, testCase := range testCases {
		bmc, err := New("1.2.3.4", validUser, validUser, defaultSSHPort, DefaultTimeOuts)
		assert.NoError(t, err, "Failed to instantiate bmc")

		err = bmc.SetPowerControlIndex(testCase.index)

		if testCase.expectedErrMsg == "" {
			assert.NoError(t, err)
		} else {
			assert.EqualError(t, err, testCase.expectedErrMsg)
		}
	}
}

func TestBMCManufacturer(t *testing.T) {
	respCallbacks := redfishAPIResponseCallbacks{}

	// We will check user and password received by the connect/login to be
	// the ones we're using.
	user := User{}
	respCallbacks.sessions = getAuthDataCallbackFn(t, &user)

	redfishServer := createFakeRedfishLocalServer(false, respCallbacks)
	defer redfishServer.Close()

	host := strings.Split(redfishServer.URL, "//")[1]
	redfishAuth := User{"user1", "pass1"}

	// No ssh credentials needed.
	bmc, err := New(host, redfishAuth, validUser, defaultSSHPort, DefaultTimeOuts)
	if err != nil {
		t.Errorf("Failed to instantiate bmc: %v", err)
	}

	// Get the manufacturer of system index = 0
	const expectedManufacturer = "Dell Inc."

	manufacturer, err := bmc.SystemManufacturer()
	if err != nil {
		t.Errorf("Failed to get system manufacturer: %v", err)
	}

	// Check the credentials were the ones we used:
	if user.Name != redfishAuth.Name && user.Password != redfishAuth.Password {
		t.Errorf("Wrong auth received in redfish server. Expected: %+v, Got: %+v", redfishAuth, user)
	}

	if manufacturer != expectedManufacturer {
		t.Errorf("Unexpected manufacturer. Want: %v, Got: %v", expectedManufacturer, manufacturer)
	}

	// Try getting the manufacturer of a non-existent system (e.g. index 1).
	const expectedErrMsg = "failed to get redfish system: invalid system index 1 (base-index=0, num systems=1)"

	err = bmc.SetSystemIndex(1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	_, err = bmc.SystemManufacturer()
	if err.Error() != expectedErrMsg {
		t.Errorf("Unexpected error when getting manufacturer of non-existent system. Want: %v, Got: %v", expectedErrMsg, err)
	}
}

func TestBMCManufacturerTimeout(t *testing.T) {
	respCallbacks := redfishAPIResponseCallbacks{}

	// We'll simulate a 200ms delay in the response to one of the rest endpoints
	respCallbacks.sessions = getDelayResponseCallbackFn(t, 200*time.Millisecond)

	redfishServer := createFakeRedfishLocalServer(false, respCallbacks)
	defer redfishServer.Close()

	host := strings.Split(redfishServer.URL, "//")[1]
	redfishAuth := User{"user1", "pass1"}

	// Set a maximum 100ms timeout from redfish api.
	bmc, err := New(host, redfishAuth, validUser, defaultSSHPort,
		TimeOuts{Redfish: 100 * time.Millisecond, SSH: 100 * time.Millisecond})
	if err != nil {
		t.Errorf("Failed to instantiate bmc: %v", err)
	}

	// Get Manufacturer. Since we've force a response greater than the configured timeout, we
	// should get an error.
	const expectedTimeoutErrMsgRegex = `redfish connection error: failed to connect to redfish endpoint: ` +
		`Post "https://127\.0\.0\.1:\d+/redfish/v1/SessionService/Sessions": context deadline exceeded`

	regex := regexp.MustCompile(expectedTimeoutErrMsgRegex)

	_, err = bmc.SystemManufacturer()
	if err == nil {
		t.Errorf("No error found. Expected err regexp: %v", expectedTimeoutErrMsgRegex)
	} else {
		errMsg := err.Error()

		match := regex.Find([]byte(errMsg))
		if len(match) == 0 {
			t.Errorf("Expected error won't match. Expected regexp %v, Got: %v", expectedTimeoutErrMsgRegex, errMsg)
		}
	}
}

func TestBMCSecureBootStatus(t *testing.T) {
	redfishServer := createFakeRedfishLocalServer(false, redfishAPIResponseCallbacks{})

	host := strings.Split(redfishServer.URL, "//")[1]
	redfishAuth := User{"user1", "pass1"}

	// No ssh credentials needed.
	bmc, err := New(host, redfishAuth, validUser, defaultSSHPort, DefaultTimeOuts)
	if err != nil {
		t.Errorf("Failed to instantiate bmc: %v", err)
	}

	// Get twice the SecureBoot status, which should be false.
	expectedSecureBootStatus := false

	sbStatus, err := bmc.IsSecureBootEnabled()
	if err != nil {
		t.Errorf("Unexpected error found when getting secure boot status: %v", err)
	}

	if sbStatus != expectedSecureBootStatus {
		t.Errorf("Secure boot status won't match. Want: %v, Got: %v", expectedSecureBootStatus, sbStatus)
	}

	// Try getting the secureboot status from a non-existent system (e.g index 2)
	expectedErrMsg := fmt.Sprintf(secureBootFailFmt, 2)

	err = bmc.SetSystemIndex(2)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	_, err = bmc.IsSecureBootEnabled()
	if err.Error() != expectedErrMsg {
		t.Errorf("Unexpected error when getting manufacturer of non-existent system. Want: %v, Got: %v", expectedErrMsg, err)
	}

	redfishServer.Close()

	// Now let's create another fake redfish server where secureboot is enabled.
	redfishServer = createFakeRedfishLocalServer(true, redfishAPIResponseCallbacks{})
	defer redfishServer.Close()

	host = strings.Split(redfishServer.URL, "//")[1]
	redfishAuth = User{"user1", "pass1"}

	// No ssh credentials needed.
	bmc, err = New(host, redfishAuth, validUser, defaultSSHPort, DefaultTimeOuts)
	if err != nil {
		t.Errorf("Failed to instantiate bmc: %v", err)
	}

	// Get twice the SecureBoot status, which should be true.
	expectedSecureBootStatus = true

	sbStatus, err = bmc.IsSecureBootEnabled()
	if err != nil {
		t.Errorf("Unexpected error found when getting secure boot status: %v", err)
	}

	if sbStatus != expectedSecureBootStatus {
		t.Errorf("Secure boot status won't match. Want: %v, Got: %v", expectedSecureBootStatus, sbStatus)
	}
}

func TestBMCSecureBootEnable(t *testing.T) {
	// Create de fake redfish api endpoint with secureBoot "disabled"
	redfishServer := createFakeRedfishLocalServer(false, redfishAPIResponseCallbacks{})
	defer redfishServer.Close()

	host := strings.Split(redfishServer.URL, "//")[1]
	redfishAuth := User{"user1", "pass1"}

	// No ssh credentials needed.
	bmc, err := New(host, redfishAuth, validUser, defaultSSHPort, DefaultTimeOuts)
	if err != nil {
		t.Errorf("Failed to instantiate bmc: %v", err)
	}

	// Secure boot is already disabled, so we should get an error if we try to disable it again.
	const expectedErrorMsg = "secure boot is already disabled"

	err = bmc.SecureBootDisable()
	if err.Error() != expectedErrorMsg {
		t.Errorf("Unexpected error when disabling secure boot. Want: %v, Got: %v", expectedErrorMsg, err)
	}

	if err := bmc.SecureBootEnable(); err != nil {
		t.Errorf("Unexpected error found when enabling secure boot: %v", err)
	}

	// Try enabling the secureboot status from a non-existent system (e.g index 2)
	expectedErrMsg := fmt.Sprintf(secureBootFailFmt, 2)

	err = bmc.SetSystemIndex(2)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	err = bmc.SecureBootEnable()
	if err.Error() != expectedErrMsg {
		t.Errorf("Unexpected error when getting manufacturer of non-existent system. Want: %v, Got: %v", expectedErrMsg, err)
	}
}

func TestBMCSecureBootDisable(t *testing.T) {
	// Create de fake redfish api endpoint with secureBoot "enabled"
	redfishServer := createFakeRedfishLocalServer(true, redfishAPIResponseCallbacks{})
	defer redfishServer.Close()

	host := strings.Split(redfishServer.URL, "//")[1]
	redfishAuth := User{"user1", "pass1"}

	// No ssh credentials needed.
	bmc, err := New(host, redfishAuth, validUser, defaultSSHPort, DefaultTimeOuts)
	if err != nil {
		t.Errorf("Failed to instantiate bmc: %v", err)
	}

	// Secure boot is already disabled, so we should get an error if we try to disable it again.
	const expectedErrorMsg = "secure boot is already enabled"

	err = bmc.SecureBootEnable()
	if err.Error() != expectedErrorMsg {
		t.Errorf("Unexpected error when disabling secure boot. Want: %v, Got: %v", expectedErrorMsg, err)
	}

	if err := bmc.SecureBootDisable(); err != nil {
		t.Errorf("Unexpected error found when enabling secure boot: %v", err)
	}

	// Try enabling the secureboot status from a non-existent system (e.g index 2)
	const expectedErrMsg = "failed to get secure boot: failed to get redfish system: " +
		"invalid system index 2 (base-index=0, num systems=1)"

	err = bmc.SetSystemIndex(2)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	err = bmc.SecureBootEnable()
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
		redfish.PowerCycleResetType,
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

func TestBMCSystemPowerCycle(t *testing.T) {
	testResetAction(t, "PowerCycle", func(bmc *BMC) error {
		return bmc.SystemPowerCycle()
	})
}

func TestBMCPowerUsage(t *testing.T) {
	// Create a fake redfish api endpoint with secureBoot "disabled"
	redfishServer := createFakeRedfishLocalServer(false, redfishAPIResponseCallbacks{})
	defer redfishServer.Close()

	host := strings.Split(redfishServer.URL, "//")[1]
	redfishAuth := User{"user1", "pass1"}

	// No ssh credentials needed.
	bmc, err := New(host, redfishAuth, validUser, defaultSSHPort, DefaultTimeOuts)
	assert.Nil(t, err, "Failed to instantiate bmc")

	const expectedPowerUsage float32 = 360.0

	power, err := bmc.PowerUsage()
	assert.Nil(t, err)
	assert.Equal(t, expectedPowerUsage, power)
}

func TestBMCCreateCLISSHSession(t *testing.T) {
	timeouts := TimeOuts{Redfish: 1 * time.Second, SSH: 10 * time.Millisecond}

	bmc, err := New("1.2.3.4", validUser, validUser, defaultSSHPort, timeouts)
	if err != nil {
		t.Errorf("Failed to instantiate bmc: %v", err)
	}

	const expectedErrMsg = `failed to connect to BMC's SSH server: dial tcp 1.2.3.4:22: i/o timeout`

	_, err = bmc.CreateCLISSHSession()
	if err == nil {
		t.Error("Err should not be nil.")
	} else if err.Error() != expectedErrMsg {
		t.Errorf("Unexpected error. Expected %v, Got: %v", expectedErrMsg, err.Error())
	}
}

func TestBMCRunCLICommand(t *testing.T) {
	// Force SSH timeout to 10ms to make it fail faster.
	timeouts := TimeOuts{Redfish: 1 * time.Second, SSH: 10 * time.Millisecond}

	bmc, err := New("1.2.3.4", validUser, validUser, defaultSSHPort, timeouts)
	if err != nil {
		t.Errorf("Failed to instantiate bmc: %v", err)
	}

	const expectedErrMsg = `failed to connect to CLI: failed to connect to BMC's SSH server: ` +
		`dial tcp 1.2.3.4:22: i/o timeout`

	_, _, err = bmc.RunCLICommand("help", false, 5*time.Second)
	if err == nil {
		t.Errorf("Err should not be nil.")
	} else if err.Error() != expectedErrMsg {
		t.Errorf("Unexpected error. Expected %v, Got: %v", expectedErrMsg, err.Error())
	}
}

func TestBMCSerialConsole(t *testing.T) {
	timeouts := TimeOuts{Redfish: 1 * time.Second, SSH: 10 * time.Millisecond}

	bmc, err := New("1.2.3.4", validUser, validUser, defaultSSHPort, timeouts)
	if err != nil {
		t.Errorf("Failed to instantiate bmc: %v", err)
	}

	var expectedErrMsg = `failed to create underlying ssh session for 1.2.3.4: ` +
		`failed to connect to BMC's SSH server: dial tcp 1.2.3.4:22: i/o timeout`

	_, _, err = bmc.OpenSerialConsole("console com2")
	if err == nil {
		t.Errorf("Err should not be nil.")
	} else if err.Error() != expectedErrMsg {
		t.Errorf("Unexpected error. Expected %v, Got: %v", expectedErrMsg, err.Error())
	}

	// Test without cli command... A best effort is made to open it based on system's manufacturer.
	expectedErrMsg = `failed to get redfish system manufacturer for 1.2.3.4: redfish connection error: ` +
		`failed to connect to redfish endpoint: Get "https://1.2.3.4/redfish/v1/": context deadline exceeded`

	_, _, err = bmc.OpenSerialConsole("")
	if err == nil {
		t.Errorf("Err should not be nil.")
	} else if err.Error() != expectedErrMsg {
		t.Errorf("Unexpected error. Expected %v, Got: %v", expectedErrMsg, err.Error())
	}

	expectedErrMsg = "no underlying ssh session found for 1.2.3.4"

	err = bmc.CloseSerialConsole()
	if err == nil {
		t.Errorf("Err should not be nil.")
	} else if err.Error() != expectedErrMsg {
		t.Errorf("Unexpected error. Expected %v, Got: %v", expectedErrMsg, err.Error())
	}
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

// Helper function that creates a fake redfish REST server in localhost (random port).
// When outputAuthData is provided, it will be filled with the auth credentials received in the
// login request. All the responses, except the login one, are sent using static json data from
// the testdata folder. The flag secureBootEnable is used to load the json response for the
// secure boot api depending on wether we want it to be enabled or disabled for our test.
func createFakeRedfishLocalServer(secureBootEnabled bool, callbacks redfishAPIResponseCallbacks) *httptest.Server {
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
	redfishAuth := User{"user1", "pass1"}

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
			// No ssh credentials needed.
			bmc, err := New(host, redfishAuth, validUser, defaultSSHPort, DefaultTimeOuts)
			assert.NoError(t, err, "Failed to instantiate bmc")

			err = bmc.SetSystemIndex(testCase.systemIndex)
			assert.NoError(t, err, "Failed to set system index")

			err = resetFunction(bmc)
			if testCase.expectedErrorText != "" {
				assert.EqualError(t, err, testCase.expectedErrorText)
			}
		})
	}
}
