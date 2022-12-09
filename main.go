package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
)

func setDNSServerForWindows(primaryDNS string, secondaryDNS string) {
	// Gets all the Physical Interfaces with the following Output
	// Output is of the following Format:
	/*
		Admin State    State          Type             Interface Name
		-------------------------------------------------------------------------
		Disabled       Disconnected   Dedicated        Ethernet 2
		Enabled        Disconnected   Dedicated        Ethernet
		Enabled        Connected      Dedicated        Wi-Fi
	*/
	getInterfaces := exec.Command("netsh", "interface", "show", "interface")

	// Filters only the "Connected" (Active) Interfaces' lines. will output the following:
	// Enabled        Connected      Dedicated        Wi-Fi
	filterConnectedOnly := exec.Command("findstr", "\\<Connected")

	// Workaround to run piped (with "|") commands
	filterConnectedOnly.Stdin, _ = getInterfaces.StdoutPipe()

	var filterConnectedOnlyStdout bytes.Buffer
	filterConnectedOnly.Stdout = &filterConnectedOnlyStdout

	filterConnectedOnly.Start()
	if err := getInterfaces.Run(); err != nil {
		log.Fatal(err)
	}
	if err := filterConnectedOnly.Wait(); err != nil {
		log.Fatal(err)
	}

	interfacesOutput := filterConnectedOnlyStdout.String()
	fmt.Println("Found Interfaces: ", interfacesOutput)
	scanner := bufio.NewScanner(strings.NewReader(interfacesOutput))
	// Iterate over the output line-by-line
	for scanner.Scan() {
		line := scanner.Text()

		// The Actual Name of the interface is in the last column
		// so splitting on " " and then accessing the last element of the array
		interfaceInfo := strings.Split(line, " ")
		interfaceName := interfaceInfo[len(interfaceInfo)-1]

		// Command to set the Primary DNS server using "netsh"
		setPrimaryDNSCmd := exec.Command("netsh", "interface", "ip", "set", "dns", interfaceName, "static", primaryDNS)
		fmt.Println("Set Primary DNS Command: ", setPrimaryDNSCmd)

		if err := setPrimaryDNSCmd.Run(); err != nil {
			log.Fatal("Error Updating Primary DNS ", err)
		}

		if secondaryDNS != "" {
			// Command to set the Secondary DNS server using "netsh"
			setSecondaryDNSCmd := exec.Command("netsh", "interface", "ip", "add", "dns", interfaceName, secondaryDNS, "index=2")
			if err := setSecondaryDNSCmd.Run(); err != nil {
				log.Fatal("Error Updating Secondary DNS ", err)
			}
		}

		fmt.Println("Custom DNS server updated for Interface: ", interfaceName)
	}
}

func setDNSServerForMacOS(primaryDNS string, secondaryDNS string) {
	// Command to Get the currently Active Interface IP and Device Name
	getActiveInterfaceNameCommand := "netstat -rn | awk '($1 == \"default\") {print $4; exit}'"
	cmd := exec.Command("bash", "-c", getActiveInterfaceNameCommand)

	var stderr, stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		fmt.Println("Error" + stderr.String())
		log.Fatal(stderr.String())
	}

	activeInterfaceDevice := strings.TrimSpace(stdout.String())
	if activeInterfaceDevice == "" {
		log.Fatal("Unable to get primary network interface")
	}

	// Get the Actual Name of the Interface using the device name
	getInterfaceNameCommand := fmt.Sprintf("networksetup -listallhardwareports | grep -B1 \"Device: %s\\$\" | sed -n 's/^Hardware Port: //p'", activeInterfaceDevice)
	fmt.Println(getInterfaceNameCommand)
	cmdRun := exec.Command("bash", "-c", getInterfaceNameCommand)

	var interfaceNameStderr, interfaceNameStdout bytes.Buffer
	cmdRun.Stderr = &interfaceNameStderr
	cmdRun.Stdout = &interfaceNameStdout
	if err := cmdRun.Run(); err != nil {
		log.Fatal(stderr.String())
	}
	activeInterfaceName := strings.TrimSpace(interfaceNameStdout.String())
	fmt.Println("Output ", activeInterfaceName)

	setDNSCmd := exec.Command("networksetup", "-setdnsservers", activeInterfaceName, primaryDNS, secondaryDNS)
	if err := setDNSCmd.Run(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Custom DNS server updated for Interface: ", activeInterfaceName)
}

func main() {
	primaryDNS := "8.8.8.8"
	secondaryDNS := "8.8.4.4"

	fmt.Println("Current OS: ", runtime.GOOS)
	switch runtime.GOOS {
	case "windows":
		setDNSServerForWindows(primaryDNS, secondaryDNS)
	case "darwin":
		setDNSServerForMacOS(primaryDNS, secondaryDNS)
	default:
		log.Fatal("OS not supported")
	}
}
