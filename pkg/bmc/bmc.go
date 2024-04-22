package bmc

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
	"golang.org/x/crypto/ssh"
)

const (
	manufacturerDell = "Dell Inc."
	manufacturerHPE  = "HPE"

	defaultTimeOut = 5 * time.Second
)

var (
	// CLI command to get the serial console (virtual serial port).
	cliCmdSerialConsole = map[string]string{
		manufacturerHPE:  "VSP",
		manufacturerDell: "console com2",
	}

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

	timeOuts TimeOuts

	sshSessionForSerialConsole *ssh.Session
}

// New returns a new BMC struct.
func New(host string, redfishUser, sshUser User, timeOuts TimeOuts) *BMC {
	return &BMC{
		host:        host,
		redfishUser: redfishUser,
		sshUser:     sshUser,
		timeOuts:    timeOuts,
	}
}

// SystemManufacturer gets system's manufacturer from the BMC's RedFish API endpoint.
func (bmc *BMC) SystemManufacturer(systemIndex int) (string, error) {
	redfishClient, cancel, err := redfishConnect(
		bmc.host,
		bmc.redfishUser.Name, bmc.redfishUser.Password,
		bmc.timeOuts.Redfish)
	if err != nil {
		return "", err
	}

	defer func() {
		redfishClient.Logout()
		cancel()
	}()

	system, err := redfishGetSystem(redfishClient, systemIndex)
	if err != nil {
		return "", fmt.Errorf("failed to get redfish system: %w", err)
	}

	return system.Manufacturer, nil
}

// IsSecureBootEnabled returns whether the SecureBoot feature is enabled using the BMC's RedFish API endpoint.
func (bmc *BMC) IsSecureBootEnabled(systemIndex int) (bool, error) {
	redfishClient, cancel, err := redfishConnect(
		bmc.host,
		bmc.redfishUser.Name, bmc.redfishUser.Password,
		bmc.timeOuts.Redfish)
	if err != nil {
		return false, err
	}

	defer func() {
		redfishClient.Logout()
		cancel()
	}()

	sboot, err := redfishGetSystemSecureBoot(redfishClient, systemIndex)
	if err != nil {
		return false, err
	}

	return sboot.SecureBootEnable, nil
}

// SecureBootEnable enables the SecureBoot feature using the BMC's RedFish API endpoint.
func (bmc *BMC) SecureBootEnable(systemIndex int) error {
	redfishClient, cancel, err := redfishConnect(
		bmc.host,
		bmc.redfishUser.Name, bmc.redfishUser.Password,
		bmc.timeOuts.Redfish)
	if err != nil {
		return err
	}

	defer func() {
		redfishClient.Logout()
		cancel()
	}()

	sboot, err := redfishGetSystemSecureBoot(redfishClient, systemIndex)
	if err != nil {
		return err
	}

	if sboot.SecureBootEnable {
		return fmt.Errorf("secure boot is already enabled")
	}

	sboot.SecureBootEnable = true

	err = sboot.Update()
	if err != nil {
		return fmt.Errorf("failed to enable secure boot %w", err)
	}

	return sboot.Update()
}

// SecureBootDisable disables the SecureBoot feature using the BMC's RedFish API endpoint.
func (bmc *BMC) SecureBootDisable(systemIndex int) error {
	redfishClient, cancel, err := redfishConnect(
		bmc.host,
		bmc.redfishUser.Name, bmc.redfishUser.Password,
		bmc.timeOuts.Redfish)
	if err != nil {
		return err
	}

	defer func() {
		redfishClient.Logout()
		cancel()
	}()

	sboot, err := redfishGetSystemSecureBoot(redfishClient, systemIndex)
	if err != nil {
		return err
	}

	if !sboot.SecureBootEnable {
		return fmt.Errorf("secure boot is already disabled")
	}

	sboot.SecureBootEnable = true

	err = sboot.Update()
	if err != nil {
		return fmt.Errorf("failed to disable secure boot %w", err)
	}

	return sboot.Update()
}

// SystemForceReset performs a (non-graceful) forced system reset using redfish API.
func (bmc *BMC) SystemForceReset(systemIndex int) error {
	redfishClient, cancel, err := redfishConnect(
		bmc.host,
		bmc.redfishUser.Name, bmc.redfishUser.Password,
		bmc.timeOuts.Redfish)
	if err != nil {
		return err
	}

	defer func() {
		redfishClient.Logout()
		cancel()
	}()

	system, err := redfishGetSystem(redfishClient, systemIndex)
	if err != nil {
		return fmt.Errorf("failed to get redfish system: %w", err)
	}

	return system.Reset(redfish.ForceRestartResetType)
}

// OpenSerialConsole opens the serial console port. The console is tunneled in an underlying (CLI)
// ssh session that is opened in the BMC's ssh server. If openConsoleCliCmd is
// provided, it will be sent to the BMC's cli. Otherwise, a best effort will
// be made to run the appropriate cli command based on the system manufacturer.
func (bmc *BMC) OpenSerialConsole(openConsoleCliCmd string) (io.Reader, io.WriteCloser, error) {
	const SSHPort = 22

	if bmc.sshSessionForSerialConsole != nil {
		return nil, nil, fmt.Errorf("there is already a serial console opened for this BMC")
	}

	cliCmd := openConsoleCliCmd
	if cliCmd == "" {
		// no cli command to get console port was provided, try to guess based on
		// manufacturer.
		manufacturer, err := bmc.SystemManufacturer(0)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get redfish system manufacturer: %w", err)
		}

		var found bool
		if cliCmd, found = cliCmdSerialConsole[manufacturer]; !found {
			return nil, nil, fmt.Errorf("cli command to get serial console not found for manufacturer %v", manufacturer)
		}
	}

	sshSession, err := createSSHSession(
		bmc.host, SSHPort,
		bmc.sshUser.Name,
		bmc.sshUser.Password,
		bmc.timeOuts.SSH)
	if err != nil {
		return nil, nil, err
	}

	// Pipes need to be retrieved before session.Start()
	reader, err := sshSession.StdoutPipe()
	if err != nil {
		_ = sshSession.Close()

		return nil, nil, fmt.Errorf("failed to get stdout pipe from ssh session: %w", err)
	}

	writer, err := sshSession.StdinPipe()
	if err != nil {
		_ = sshSession.Close()

		return nil, nil, fmt.Errorf("failed to get stdin pipe from ssh session: %w", err)
	}

	err = sshSession.Start(cliCmd)
	if err != nil {
		_ = sshSession.Close()

		return nil, nil, fmt.Errorf("failed to start serial console with cli command %q: %w", cliCmd, err)
	}

	go func() { _ = sshSession.Wait() }()

	bmc.sshSessionForSerialConsole = sshSession

	return reader, writer, nil
}

// CloseSerialConsole closes the serial console's underlying ssh session.
func (bmc *BMC) CloseSerialConsole() error {
	if bmc.sshSessionForSerialConsole == nil {
		return fmt.Errorf("no underlying ssh session found")
	}

	err := bmc.sshSessionForSerialConsole.Close()
	if err != nil {
		return fmt.Errorf("failed to close underlying ssh session: %w", err)
	}

	bmc.sshSessionForSerialConsole = nil

	return nil
}

func createSSHSession(host string, port int, user, password string, timeout time.Duration) (*ssh.Session, error) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
			ssh.KeyboardInteractive(func(user, instruction string, questions []string,
				echos []bool) (answers []string, err error) {
				answers = make([]string, len(questions))
				// The second parameter is unused
				for n := range questions {
					answers[n] = password
				}

				return answers, nil
			}),
		},
		Timeout:         timeout,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Establish SSH connection
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	// Create a session
	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create new ssh session: %w", err)
	}

	return session, nil
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
