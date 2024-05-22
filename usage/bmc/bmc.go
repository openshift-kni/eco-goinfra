package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/openshift-kni/eco-goinfra/pkg/bmc"
)

// CLI flags.
var (
	hostFlag        *string        = flag.String("host", "", "RedFish host name/ip")
	sshUserFlag     *string        = flag.String("sshuser", "", "SSH user's name")
	sshPassFlag     *string        = flag.String("sshpass", "", "SSH user's password")
	sshPortFlag     *uint          = flag.Uint("sshport", 22, "SSH port")
	redfishUserFlag *string        = flag.String("redfishuser", "", "Redfish user's name")
	redfishPassFlag *string        = flag.String("redfishpass", "", "Redfish user's password")
	timeoutFlag     *time.Duration = flag.Duration("timeout", 0, "Timeout (1s, 1m30s, ...)")
	systemIndexFlag *int           = flag.Int("system-index", 0, "Redfish system index")

	// Defaulted to false. If set, it will open the serial console at the end.
	testSerialConsoleFlag *bool = flag.Bool("serialconsoletest", false, "Read the Serial Console for 10s")
)

// runCLICommand is a helper funct to run cli commands and print some traces.
func runCLICommand(bmc *bmc.BMC, cmd string, combineOutput bool, timeout time.Duration) {
	fmt.Printf("%v - Running CLI command %q:\n", time.Now(), cmd)

	stdout, stderr, err := bmc.RunCLICommand(cmd, combineOutput, timeout)
	if err != nil {
		fmt.Printf("%v - Failed to run cmd: %v\n", time.Now(), err)
	}

	fmt.Printf("%v\nstdout:\n%v\nstderr:%v\n", time.Now(), stdout, stderr)
}

func testSerialConsole(bmc *bmc.BMC, readTime time.Duration) {
	reader, writer, err := bmc.OpenSerialConsole("")
	if err != nil {
		fmt.Printf("Failed to open serial console: %v", err)
		os.Exit(1)
	}

	exitFn := func(reason string) {
		fmt.Printf("\n%s -> Closing serial port...\n", reason)

		_ = bmc.CloseSerialConsole()
	}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signalCh
		exitFn("CTRL+C captured")
		os.Exit(1)
	}()

	defer exitFn("Normal exit")

	// Close serialconsole's input.
	_ = writer.Close()

	fmt.Printf("Reading from Serial Console obtained...\n")

	go func() {
		scanner := bufio.NewScanner(reader)

		for {
			if !scanner.Scan() {
				time.Sleep(1 * time.Second)

				continue
			}

			if scanner.Err() != nil {
				fmt.Printf("Scanner error: %v", scanner.Err())

				break
			}

			line := scanner.Text()
			fmt.Printf("%v\n", line)
		}
	}()

	time.Sleep(readTime)
}

// Use cli flags to create a BMC struct and get the system's manufacturer and secure boot status.
// timeout flag is optional. If not provided, it's defaulted to bmc.DefaultTimeouts.
// systemIndex flag is optional. If not provided, it's defaulted to 0.
func main() {
	flag.Parse()

	fmt.Printf("Getting redfish information from host %s, timeouts: %s\n", *hostFlag, *timeoutFlag)

	bmcClient := bmc.New(*hostFlag).
		WithRedfishUser(*redfishUserFlag, *redfishPassFlag).
		WithSSHUser(*sshUserFlag, *sshPassFlag).
		WithSSHPort(uint16(*sshPortFlag)).
		WithRedfishSystemIndex(*systemIndexFlag)

	if *timeoutFlag != 0 {
		bmcClient = bmcClient.WithRedfishTimeout(*timeoutFlag).WithSSHTimeout(*timeoutFlag)
	} else {
		fmt.Printf("Timeout not set (or set to 0): using defaults %+v\n", bmc.DefaultTimeOuts)
	}

	manufacturer, err := bmcClient.SystemManufacturer()
	if err != nil {
		fmt.Printf("Failed to get manufacturer from redfish api on %v: %v\n", *hostFlag, err)
		os.Exit(1)
	}

	sbEnabled, err := bmcClient.IsSecureBootEnabled()
	if err != nil {
		fmt.Printf("Failed to get secure boot status on %v: %v\n", *hostFlag, err)
	}

	fmt.Printf("System %d Manufacturer       : %v\n", *systemIndexFlag, manufacturer)
	fmt.Printf("System %d SecureBoot enabled : %v\n", *systemIndexFlag, sbEnabled)

	// Run the help command. We should see all the available CLI commands.
	runCLICommand(bmcClient, "help", true, 10*time.Second)

	// Run an invalid command.
	runCLICommand(bmcClient, "wrongcommand", true, 10*time.Second)

	// Another wrong command, but in this case we want to force a timeout.
	runCLICommand(bmcClient, "anotherwrongcommand", false, 1*time.Millisecond)

	// The rest of the code tests the ssh-tunneled serial console.
	if *testSerialConsoleFlag {
		fmt.Printf("Reading from Serial Console for 10 seconds. Use ctrl+c to stop it...\n")
		testSerialConsole(bmcClient, 10*time.Second)
	}
}
