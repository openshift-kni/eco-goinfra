package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/openshift-kni/eco-goinfra/pkg/bmc"
)

// CLI flags.
var (
	hostFlag        *string        = flag.String("host", "", "RedFish host name/ip")
	sshUserFlag     *string        = flag.String("sshuser", "", "SSH user's name")
	sshPassFlag     *string        = flag.String("sshpass", "", "SSH user's password")
	redfishUserFlag *string        = flag.String("redfishuser", "", "Redfish user's name")
	redfishPassFlag *string        = flag.String("redfishpass", "", "Redfish user's password")
	timeoutFlag     *time.Duration = flag.Duration("timeout", 0, "Timeout (1s, 1m30s, ...)")
	systemIndex     *int           = flag.Int("system-index", 0, "Redfish system index")
)

// Use cli flags to create a BMC struct and get the system's manufacturer and secure boot status.
// timeout flag is optional. If not provided, it's defaulted to bmc.DefaultTimeouts.
// systemIndex flag is optional. If not provided, it's defaulted to 0.
func main() {
	flag.Parse()

	// Create users and timeout structs.
	redfishUser := bmc.User{Name: *redfishUserFlag, Password: *redfishPassFlag}
	sshUser := bmc.User{Name: *sshUserFlag, Password: *sshPassFlag}
	timeOuts := bmc.TimeOuts{Redfish: *timeoutFlag, SSH: *timeoutFlag}

	if *timeoutFlag == 0 {
		fmt.Printf("Timeout not set (or set to 0): using defaults %+v\n", bmc.DefaultTimeOuts)
		timeOuts = bmc.DefaultTimeOuts
	}

	fmt.Printf("Getting redfish information from host %s, timeouts: %+v\n", *hostFlag, timeOuts)

	bmc, err := bmc.New(*hostFlag, redfishUser, sshUser, timeOuts)
	if err != nil {
		fmt.Printf("Failed to create BMC struct: %v\n", err)
		os.Exit(1)
	}

	err = bmc.SetSystemIndex(*systemIndex)
	if err != nil {
		fmt.Printf("Failed to set system index: %v\n", err)
		os.Exit(1)
	}

	manufacturer, err := bmc.SystemManufacturer()
	if err != nil {
		fmt.Printf("Failed to get manufacturer from redfish api on %v: %v\n", *hostFlag, err)
		os.Exit(1)
	}

	sbEnabled, err := bmc.IsSecureBootEnabled()
	if err != nil {
		fmt.Printf("Failed to get secure boot status on %v: %v\n", *hostFlag, err)
	}

	fmt.Printf("System %d Manufacturer       : %v\n", *systemIndex, manufacturer)
	fmt.Printf("System %d SecureBoot enabled : %v\n", *systemIndex, sbEnabled)
}
