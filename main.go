package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func Contains(vs []string, t string, caseSensitive bool) int {
	for i, v := range vs {
		if caseSensitive {
			if strings.Contains(v, t) {
				return i
			}
		} else if strings.Contains(strings.ToLower(v), strings.ToLower(t)) {
			return i
		}
	}
	return -1
}

func SimctlExec(command string, validationLog string, deviceUUID string) {
	for {
		var tryCount = 0
		commandSimLog, _ := exec.Command("xcrun", "simctl", command, deviceUUID).CombinedOutput()
		commandSimLogList := strings.Split(string(commandSimLog), "\n")

		// Wait for simulator to be booted
		if Contains(commandSimLogList, validationLog, false) > -1 {
			break
		}
		if tryCount == 10 {
			break
		}
		tryCount++
	}
}

func main() {
	commandEnv := os.Getenv("simctl_command")
	deviceEnv := os.Getenv("simctl_device")
	// commandEnv := "boot"
	// deviceEnv := "iphone 8"

	xcrunSimctlShutdownStateLog := "Unable to shutdown device in current state: Shutdown"
	xcrunSimctlBootedStateLog := "Unable to boot device in current state: Booted"

	fmt.Println("This is the value specified for the input 'example_step_input':", os.Getenv("example_step_input"))

	devicesListLog, devicesListErr := exec.Command("xcrun", "simctl", "list", "devices").CombinedOutput()
	if devicesListErr != nil {
		fmt.Printf("XCode devices couldn't be called - error: %#v | output: %s", devicesListErr, devicesListLog)
		os.Exit(1)
	}
	devicesListLogSplitted := strings.Split(string(devicesListLog), "\n")
	fmt.Print(devicesListLogSplitted)

	indexOfSearchedDevice := Contains(devicesListLogSplitted, deviceEnv, false)
	if indexOfSearchedDevice == -1 {
		fmt.Printf("Device not found - devicesListLog: %s", devicesListLog)
		os.Exit(1)
	}

	deviceUUIDLog := devicesListLogSplitted[indexOfSearchedDevice]

	re := regexp.MustCompile("[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}")
	deviceUUID := re.FindString(deviceUUIDLog)

	if deviceUUID == "" {
		fmt.Printf("Device not found - deviceUuidLog: %s", deviceUUIDLog)
		os.Exit(1)
	}

	// boot/shutdown simulator
	if commandEnv == "boot" {
		SimctlExec(commandEnv, xcrunSimctlBootedStateLog, deviceUUID)
		fmt.Printf("XCode device %s is booted", deviceUUID)
	} else if commandEnv == "shutdown" {
		SimctlExec(commandEnv, xcrunSimctlShutdownStateLog, deviceUUID)
		fmt.Printf("XCode device %s is shutdown", deviceUUID)
	} else {
		fmt.Printf("Command %s is unknown", commandEnv)
		os.Exit(1)
	}

	// Exec simulator command successful
	os.Exit(0)
}
