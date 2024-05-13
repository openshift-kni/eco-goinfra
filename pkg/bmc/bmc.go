package bmc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang/glog"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
)

const (
	defaultTimeOut = 5 * time.Second
)

var (
	// DefaultTimeOuts holds the default redfish and ssh timeouts.
	DefaultTimeOuts = TimeOuts{
		Redfish: defaultTimeOut,
		SSH:     defaultTimeOut,
	}
)

// User holds the Name and Password for a user (ssh/redfish).
type User struct {
	// Name holds the user's name
	Name string
	// Password holds the user's password
	Password string
}

// TimeOuts holds the configured timeouts for Redfish and SSH acccess.
type TimeOuts struct {
	// Redfish timeout for the redfish api access.
	Redfish time.Duration
	// SSH timeout for the ssh access.
	SSH time.Duration
}

// BMC is the holder struct for BMC access through redfish & ssh.
type BMC struct {
	host        string
	redfishUser User
	sshUser     User
	systemIndex int

	timeOuts TimeOuts
}

// New returns a new BMC struct. The default system index to be used in redfish requests is 0.
// Use SetSystemIndex to modify it.
func New(host string, redfishUser, sshUser User, timeOuts TimeOuts) (*BMC, error) {
	glog.V(100).Infof("Initializing new BMC structure with the following params: %s, %v, %v, %v (system index = 0)",
		host, redfishUser, sshUser, timeOuts)

	errMsgs := []string{}
	if host == "" {
		errMsgs = append(errMsgs, "host is empty")
	}

	if redfishUser.Name == "" {
		errMsgs = append(errMsgs, "redfish user's name is empty")
	}

	if redfishUser.Password == "" {
		errMsgs = append(errMsgs, "redfish user's password is empty")
	}

	if sshUser.Name == "" {
		errMsgs = append(errMsgs, "ssh user's name is empty")
	}

	if sshUser.Password == "" {
		errMsgs = append(errMsgs, "ssh user's password is empty")
	}

	if timeOuts.Redfish == 0 {
		errMsgs = append(errMsgs, "redfish timeout is 0")
	}

	if timeOuts.SSH == 0 {
		errMsgs = append(errMsgs, "ssh timeout is 0")
	}

	// Build final error msg in case there were some error/s validating the input params.
	if len(errMsgs) > 0 {
		errMsg := ""

		for i, msg := range errMsgs {
			if i != 0 {
				errMsg += ", "
			}

			errMsg += msg
		}

		glog.V(100).Infof("Failed to initialize BMC: %s", errMsg)

		return nil, errors.New(errMsg)
	}

	return &BMC{
		host:        host,
		redfishUser: redfishUser,
		sshUser:     sshUser,
		timeOuts:    timeOuts,
		systemIndex: 0,
	}, nil
}

// SetSystemIndex sets the system index to be used in redfish api requests.
func (bmc *BMC) SetSystemIndex(index int) error {
	glog.V(100).Infof("Setting default Redfish System Index to %d", index)

	if index < 0 {
		glog.V(100).Infof("Invalid system index %d: must be >= 0", index)

		return fmt.Errorf("invalid index %d", index)
	}

	bmc.systemIndex = index

	return nil
}

// SystemManufacturer gets system's manufacturer from the BMC's RedFish API endpoint.
func (bmc *BMC) SystemManufacturer() (string, error) {
	glog.V(100).Infof("Getting SystemManufacturer param from bmc's redfish endpoint")

	redfishClient, cancel, err := redfishConnect(
		bmc.host,
		bmc.redfishUser.Name, bmc.redfishUser.Password,
		bmc.timeOuts.Redfish)
	if err != nil {
		glog.V(100).Infof("Redfish connection error: %v", err)

		return "", fmt.Errorf("redfish connection error: %w", err)
	}

	defer func() {
		redfishClient.Logout()
		cancel()
	}()

	system, err := redfishGetSystem(redfishClient, bmc.systemIndex)
	if err != nil {
		glog.V(100).Infof("Failed to get redfish system: %v", err)

		return "", fmt.Errorf("failed to get redfish system: %w", err)
	}

	return system.Manufacturer, nil
}

// IsSecureBootEnabled returns whether the SecureBoot feature is enabled using the BMC's RedFish API endpoint.
func (bmc *BMC) IsSecureBootEnabled() (bool, error) {
	glog.V(100).Infof("Getting secure boot status from bmc's redfish endpoint")

	redfishClient, cancel, err := redfishConnect(
		bmc.host,
		bmc.redfishUser.Name, bmc.redfishUser.Password,
		bmc.timeOuts.Redfish)
	if err != nil {
		glog.V(100).Infof("Redfish connection error: %v", err)

		return false, fmt.Errorf("redfish connection error: %w", err)
	}

	defer func() {
		redfishClient.Logout()
		cancel()
	}()

	sboot, err := redfishGetSystemSecureBoot(redfishClient, bmc.systemIndex)
	if err != nil {
		glog.V(100).Infof("Failed to get redfish system's secure boot: %v", err)

		return false, fmt.Errorf("failed to get secure boot: %w", err)
	}

	return sboot.SecureBootEnable, nil
}

// SecureBootEnable enables the SecureBoot feature using the BMC's RedFish API endpoint.
func (bmc *BMC) SecureBootEnable() error {
	glog.V(100).Infof("Enabling secure boot from bmc's redfish endpoint")

	redfishClient, cancel, err := redfishConnect(
		bmc.host,
		bmc.redfishUser.Name, bmc.redfishUser.Password,
		bmc.timeOuts.Redfish)
	if err != nil {
		glog.V(100).Infof("Redfish connection error: %v", err)

		return fmt.Errorf("redfish connection error: %w", err)
	}

	defer func() {
		redfishClient.Logout()
		cancel()
	}()

	sboot, err := redfishGetSystemSecureBoot(redfishClient, bmc.systemIndex)
	if err != nil {
		glog.V(100).Infof("Failed to get redfish system's secure boot: %v", err)

		return fmt.Errorf("failed to get secure boot: %w", err)
	}

	if sboot.SecureBootEnable {
		glog.V(100).Infof("Failed to enable secure boot: it is already enabled")

		return fmt.Errorf("secure boot is already enabled")
	}

	sboot.SecureBootEnable = true

	err = sboot.Update()
	if err != nil {
		glog.V(100).Infof("Failed to enable secure boot: %v", err)

		return fmt.Errorf("failed to enable secure boot: %w", err)
	}

	return nil
}

// SecureBootDisable disables the SecureBoot feature using the BMC's RedFish API endpoint.
func (bmc *BMC) SecureBootDisable() error {
	glog.V(100).Infof("Disabling secure boot from bmc's redfish endpoint")

	redfishClient, cancel, err := redfishConnect(
		bmc.host,
		bmc.redfishUser.Name, bmc.redfishUser.Password,
		bmc.timeOuts.Redfish)
	if err != nil {
		glog.V(100).Infof("Redfish connection error: %v", err)

		return fmt.Errorf("redfish connection error: %w", err)
	}

	defer func() {
		redfishClient.Logout()
		cancel()
	}()

	sboot, err := redfishGetSystemSecureBoot(redfishClient, bmc.systemIndex)
	if err != nil {
		glog.V(100).Infof("Failed to get redfish system's secure boot: %v", err)

		return fmt.Errorf("failed to get secure boot: %w", err)
	}

	if !sboot.SecureBootEnable {
		glog.V(100).Infof("Failed to disable secure boot: it is already disabled")

		return fmt.Errorf("secure boot is already disabled")
	}

	sboot.SecureBootEnable = false

	err = sboot.Update()
	if err != nil {
		glog.V(100).Infof("Failed to disable secure boot: %v", err)

		return fmt.Errorf("failed to disable secure boot: %w", err)
	}

	return nil
}

// SystemForceReset performs a (non-graceful) forced system reset using redfish API.
func (bmc *BMC) SystemForceReset() error {
	glog.V(100).Infof("Forcing system reset from bmc's redfish endpoint")

	redfishClient, cancel, err := redfishConnect(
		bmc.host,
		bmc.redfishUser.Name, bmc.redfishUser.Password,
		bmc.timeOuts.Redfish)
	if err != nil {
		glog.V(100).Infof("Redfish connection error: %v", err)

		return fmt.Errorf("redfish connection error: %w", err)
	}

	defer func() {
		redfishClient.Logout()
		cancel()
	}()

	system, err := redfishGetSystem(redfishClient, bmc.systemIndex)
	if err != nil {
		glog.V(100).Infof("Failed to get redfish system: %v", err)

		return fmt.Errorf("failed to get redfish system: %w", err)
	}

	return system.Reset(redfish.ForceRestartResetType)
}

func redfishConnect(host, user, password string, sessionTimeout time.Duration) (
	*gofish.APIClient, context.CancelFunc, error) {
	gofishConfig := gofish.ClientConfig{
		Endpoint: "https://" + host,
		Username: user,
		Password: password,
		Insecure: true,
	}

	ctx, cancel := context.WithTimeout(context.Background(), sessionTimeout)

	client, err := gofish.ConnectContext(ctx, gofishConfig)
	if err != nil {
		cancel()

		return nil, nil, fmt.Errorf("failed to connect to redfish endpoint: %w", err)
	}

	return client, cancel, nil
}

func redfishGetSystem(redfishClient *gofish.APIClient, index int) (*redfish.ComputerSystem, error) {
	systems, err := redfishClient.GetService().Systems()
	if err != nil {
		return nil, fmt.Errorf("failed to get systems: %w", err)
	}

	if len(systems) < index+1 {
		return nil, fmt.Errorf("invalid system index %d (base-index=0, num systems=%d)", index, len(systems))
	}

	return systems[index], nil
}

func redfishGetSystemSecureBoot(redfishClient *gofish.APIClient, systemIndex int) (*redfish.SecureBoot, error) {
	system, err := redfishGetSystem(redfishClient, systemIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to get redfish system: %w", err)
	}

	sboot, err := system.SecureBoot()
	if err != nil {
		return nil, err
	}

	return sboot, nil
}
