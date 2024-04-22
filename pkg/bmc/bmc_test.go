package bmc

import (
	_ "embed"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"

	"net/http"
	"net/http/httptest"
	"testing"
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

	redfishServer := httptest.NewUnstartedServer(mux)
	redfishServer.EnableHTTP2 = true
	redfishServer.StartTLS()

	return redfishServer
}

func TestBMC_Manufacturer(t *testing.T) {
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
	bmc := New(host, redfishAuth, User{}, DefaultTimeOuts)

	// Get the manufacturer of system index = 0
	const expectedManufacturer = "Dell Inc."

	manufacturer, err := bmc.SystemManufacturer(0)
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

	_, err = bmc.SystemManufacturer(1)
	if err.Error() != expectedErrMsg {
		t.Errorf("Unexpected error when getting manufacturer of non-existent system. Want: %v, Got: %v", expectedErrMsg, err)
	}
}

func TestBMC_ManufacturerTimeout(t *testing.T) {
	respCallbacks := redfishAPIResponseCallbacks{}

	// We'll simulate a 200ms delay in the response to one of the rest endpoints
	respCallbacks.sessions = getDelayResponseCallbackFn(t, 200*time.Millisecond)

	redfishServer := createFakeRedfishLocalServer(false, respCallbacks)
	defer redfishServer.Close()

	host := strings.Split(redfishServer.URL, "//")[1]
	redfishAuth := User{"user1", "pass1"}

	// Set a maximum 100ms timeout from redfish api.
	bmc := New(host, redfishAuth, User{}, TimeOuts{Redfish: 100 * time.Millisecond})

	// Get Manufacturer. Since we've force a response greater than the configured timeout, we
	// should get an error.
	const expectedTimeoutErrMsg = "context deadline exceeded"

	_, err := bmc.SystemManufacturer(0)
	if err == nil {
		t.Errorf("No error found. Expected err: %v", expectedTimeoutErrMsg)
	} else {
		errMsg := errors.Unwrap(errors.Unwrap(err)).Error()
		if errMsg != expectedTimeoutErrMsg {
			t.Errorf("Expected error won't match. Expectd %v, Got: %v", expectedTimeoutErrMsg, errMsg)
		}
	}
}

func Test_BMCSecureBootStatus(t *testing.T) {
	redfishServer := createFakeRedfishLocalServer(false, redfishAPIResponseCallbacks{})

	host := strings.Split(redfishServer.URL, "//")[1]
	redfishAuth := User{"user1", "pass1"}

	// No ssh credentials needed.
	bmc := New(host, redfishAuth, User{}, DefaultTimeOuts)

	// Get twice the SecureBoot status, which should be false.
	expectedSecureBootStatus := false

	sbStatus, err := bmc.IsSecureBootEnabled(0)
	if err != nil {
		t.Errorf("Unexpected error found when getting secure boot status: %v", err)
	}

	if sbStatus != expectedSecureBootStatus {
		t.Errorf("Secure boot status won't match. Want: %v, Got: %v", expectedSecureBootStatus, sbStatus)
	}

	// Try getting the secureboot status from a non-existent system (e.g index 2)
	const expectedErrMsg = "failed to get redfish system: invalid system index 2 (base-index=0, num systems=1)"

	_, err = bmc.IsSecureBootEnabled(2)
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
	bmc = New(host, redfishAuth, User{}, DefaultTimeOuts)

	// Get twice the SecureBoot status, which should be true.
	expectedSecureBootStatus = true

	sbStatus, err = bmc.IsSecureBootEnabled(0)
	if err != nil {
		t.Errorf("Unexpected error found when getting secure boot status: %v", err)
	}

	if sbStatus != expectedSecureBootStatus {
		t.Errorf("Secure boot status won't match. Want: %v, Got: %v", expectedSecureBootStatus, sbStatus)
	}
}

func Test_BMCSecureBootEnable(t *testing.T) {
	// Create de fake redfish api endpoint with secureBoot "disabled"
	redfishServer := createFakeRedfishLocalServer(false, redfishAPIResponseCallbacks{})
	defer redfishServer.Close()

	host := strings.Split(redfishServer.URL, "//")[1]
	redfishAuth := User{"user1", "pass1"}

	// No ssh credentials needed.
	bmc := New(host, redfishAuth, User{}, DefaultTimeOuts)

	// Secure boot is already disabled, so we should get an error if we try to disable it again.
	const expectedErrorMsg = "secure boot is already disabled"

	err := bmc.SecureBootDisable(0)
	if err.Error() != expectedErrorMsg {
		t.Errorf("Unexpected error when disabling secure boot. Want: %v, Got: %v", expectedErrorMsg, err)
	}

	if err := bmc.SecureBootEnable(0); err != nil {
		t.Errorf("Unexpected error found when enabling secure boot: %v", err)
	}

	// Try enabling the secureboot status from a non-existent system (e.g index 2)
	const expectedErrMsg = "failed to get redfish system: invalid system index 2 (base-index=0, num systems=1)"

	err = bmc.SecureBootEnable(2)
	if err.Error() != expectedErrMsg {
		t.Errorf("Unexpected error when getting manufacturer of non-existent system. Want: %v, Got: %v", expectedErrMsg, err)
	}
}

func Test_BMCSecureBootDisable(t *testing.T) {
	// Create de fake redfish api endpoint with secureBoot "enabled"
	redfishServer := createFakeRedfishLocalServer(true, redfishAPIResponseCallbacks{})
	defer redfishServer.Close()

	host := strings.Split(redfishServer.URL, "//")[1]
	redfishAuth := User{"user1", "pass1"}

	// No ssh credentials needed.
	bmc := New(host, redfishAuth, User{}, DefaultTimeOuts)

	// Secure boot is already disabled, so we should get an error if we try to disable it again.
	const expectedErrorMsg = "secure boot is already enabled"

	err := bmc.SecureBootEnable(0)
	if err.Error() != expectedErrorMsg {
		t.Errorf("Unexpected error when disabling secure boot. Want: %v, Got: %v", expectedErrorMsg, err)
	}

	if err := bmc.SecureBootDisable(0); err != nil {
		t.Errorf("Unexpected error found when enabling secure boot: %v", err)
	}

	// Try enabling the secureboot status from a non-existent system (e.g index 2)
	const expectedErrMsg = "failed to get redfish system: invalid system index 2 (base-index=0, num systems=1)"

	err = bmc.SecureBootEnable(2)
	if err.Error() != expectedErrMsg {
		t.Errorf("Unexpected error when getting manufacturer of non-existent system. Want: %v, Got: %v", expectedErrMsg, err)
	}
}

func Test_BMCSystemForceReboot(t *testing.T) {
	// Create de fake redfish api endpoint with secureBoot "enabled"
	redfishServer := createFakeRedfishLocalServer(false, redfishAPIResponseCallbacks{})
	defer redfishServer.Close()

	host := strings.Split(redfishServer.URL, "//")[1]
	redfishAuth := User{"user1", "pass1"}

	// No ssh credentials needed.
	bmc := New(host, redfishAuth, User{}, DefaultTimeOuts)

	// Try ForceReset on system 0
	err := bmc.SystemForceReset(0)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Try enabling the secureboot status from a non-existent system (e.g index 2)
	const expectedErrMsg = "failed to get redfish system: invalid system index 2 (base-index=0, num systems=1)"

	err = bmc.SystemForceReset(2)
	if err.Error() != expectedErrMsg {
		t.Errorf("Unexpected error when getting manufacturer of non-existent system. Want: %v, Got: %v", expectedErrMsg, err)
	}
}

func Test_OpenSerialConsoleTimeout(t *testing.T) {
	bmc := New("1.2.3.4", User{}, User{}, TimeOuts{Redfish: 100 * time.Millisecond, SSH: 100 * time.Millisecond})

	expectedErrorMsg := "failed to get redfish system manufacturer: failed to connect to redfish endpoint: " +
		"Get \"https://1.2.3.4/redfish/v1/\": context deadline exceeded"

	// Try to use redfish's Manufacturer to get the cli command to open the serial console.
	_, _, err := bmc.OpenSerialConsole("")
	if err == nil {
		t.Errorf("An error was expected here.")
	} else if err.Error() != expectedErrorMsg {
		t.Errorf("Unexpected error. Expected: %v, Got: %v", expectedErrorMsg, err.Error())
	}

	expectedErrorMsg = "failed to connect: dial tcp 1.2.3.4:22: i/o timeout"
	// Try with custom cli command to get the ssh-tunneled serial port.
	_, _, err = bmc.OpenSerialConsole("fake-cli-cmd")
	if err == nil {
		t.Errorf("An error was expected here.")
	} else if err.Error() != expectedErrorMsg {
		t.Errorf("Unexpected error. Expected: %v, Got: %v", expectedErrorMsg, err.Error())
	}
}
