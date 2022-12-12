package main

import (
	"bufio"
	"bytes"
	"custom-dns/utils"
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

func TestActiveGetInterfaceNameForMacOS(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Run("Test Get Active Interface Name", func(t *testing.T) {
			// If the system is connected with Wi-Fi,
			// Otherwise change this to Ethernet or any other

			currentlyActiveInterface := "Wi-Fi"

			activeInterfaceName, err := utils.GetActiveInterfaceNameForMacOS()
			if err != nil {
				t.Error(err)
			}
			if currentlyActiveInterface != activeInterfaceName {
				t.Errorf("Mismatching Active Interfaces, expected: %s, got: %s", currentlyActiveInterface, activeInterfaceName)
			}
		})
	}
}

func TestSetCustomDNSForMacOS(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Run("Test Setting Custom DNS Server for Active Interface", func(t *testing.T) {
			testPrimaryAddress := "1.1.1.1"
			testSecondaryAddress := "8.8.8.8"

			setDNSServerForMacOS(testPrimaryAddress, testSecondaryAddress)

			activeInterfaceName, err := utils.GetActiveInterfaceNameForMacOS()
			if err != nil {
				t.Error("Couldn't get active interface")
			}

			cmd := exec.Command("networksetup", "-getdnsservers", activeInterfaceName)

			var stdout bytes.Buffer

			cmd.Stdout = &stdout

			if err := cmd.Run(); err != nil {
				t.Error(err)
			}

			scanner := bufio.NewScanner(strings.NewReader(stdout.String()))
			counter := 1
			for scanner.Scan() {
				line := scanner.Text()
				activeInterfaceDevice := strings.TrimSpace(line)
				if activeInterfaceDevice == "" {
					t.Error("No DNS Servers Found")
				}

				if counter == 1 {
					if activeInterfaceDevice != testPrimaryAddress {
						t.Errorf("Primary DNS not matched. Expected: %s, Got: %s", testPrimaryAddress, activeInterfaceDevice)
					}
				}

				if counter == 2 {
					if activeInterfaceDevice != testSecondaryAddress {
						t.Errorf("Secondary DNS not matched. Expected: %s, Got: %s", testSecondaryAddress, activeInterfaceDevice)
					}
				}

				counter++
			}
		})
	}
}

func TestSetCustomDNSForWindwos(t *testing.T) {
	if runtime.GOOS == "windows" {
		testPrimaryAddress := "8.8.8.8"
		testSecondaryAddress := "8.8.4.4"

		setDNSServerForWindows(testPrimaryAddress, testSecondaryAddress)

		cmd := exec.Command("netsh", "interface", "ip", "show", "dns")

		var stdout bytes.Buffer

		cmd.Stdout = &stdout

		if err := cmd.Run(); err != nil {
			t.Error(err)
		}

		scanner := bufio.NewScanner(strings.NewReader(stdout.String()))

		if !strings.Contains(testPrimaryAddress, scanner.Text()) {
			t.Error("Primary DNS server not found")
		}

		if !strings.Contains(testSecondaryAddress, scanner.Text()) {
			t.Error("Secondary DNS server not found")
		}
	}

}
