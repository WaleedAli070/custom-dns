package utils

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func GetActiveInterfaceNameForMacOS() (string, error) {
	// Command to Get the currently Active Interface IP and Device Name
	getActiveInterfaceNameCommand := "netstat -rn | awk '($1 == \"default\") {print $4; exit}'"
	cmd := exec.Command("bash", "-c", getActiveInterfaceNameCommand)

	var stderr, stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		fmt.Println("Error" + stderr.String())
		log.Fatal(stderr.String())
		return "", errors.New("error Getting the Active Interface's IP")
	}

	activeInterfaceDevice := strings.TrimSpace(stdout.String())
	if activeInterfaceDevice == "" {
		return "", errors.New("unable to get primary network interface")
	}

	// Get the Actual Name of the Interface using the device name
	getInterfaceNameCommand := fmt.Sprintf("networksetup -listallhardwareports | grep -B1 \"Device: %s\\$\" | sed -n 's/^Hardware Port: //p'", activeInterfaceDevice)
	fmt.Println(getInterfaceNameCommand)
	cmdRun := exec.Command("bash", "-c", getInterfaceNameCommand)

	var interfaceNameStderr, interfaceNameStdout bytes.Buffer
	cmdRun.Stderr = &interfaceNameStderr
	cmdRun.Stdout = &interfaceNameStdout
	if err := cmdRun.Run(); err != nil {
		return "", errors.New("unable to get Hardware Ports")
	}
	activeInterfaceName := strings.TrimSpace(interfaceNameStdout.String())

	return activeInterfaceName, nil
}
