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

func GetDeviceId(deviceNameEnv string) string {
	devicesListLog, devicesListErr := exec.Command("xcrun", "simctl", "list", "devices").CombinedOutput()
	if devicesListErr != nil {
		fmt.Printf("XCode devices couldn't be called - error: %#v | output: %s", devicesListErr, devicesListLog)
		os.Exit(1)
	}
	devicesListLogSplitted := strings.Split(string(devicesListLog), "\n")
	// fmt.Print(devicesListLogSplitted)

	indexOfSearchedDevice := Contains(devicesListLogSplitted, deviceNameEnv, false)
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
	return deviceUUID
}

func SimctlExec(command string, validationLog string, deviceUUID string) {
	for {
		var tryCount = 0
		commandSimLog, _ := exec.Command(
			"xcrun", "simctl", command, deviceUUID).CombinedOutput()
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

func SimctlExecPermission(command string, deviceUUID string,
	action string, service string, bundleId string) {
	exec.Command("xcrun", "simctl", command, deviceUUID, action, service, bundleId).CombinedOutput()
}

func SetEnv(key string, value string) {
	cmdLog, err := exec.Command("bitrise", "envman", "add", "--key", key, "--value", value).CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to expose output with envman, error: %#v | output: %s", err, cmdLog)
		// os.Exit(1)
	}
}

func main() {
	localTesting := false

	commandEnv := os.Getenv("simctl_command")
	deviceNameEnv := os.Getenv("simctl_device")
	simctlAction := os.Getenv("BITRISE_SIMCTL_ACTION")
	simctlService := os.Getenv("BITRISE_SIMCTL_SERVICE")
	iosBundleId := os.Getenv("BITRISE_IOS_BUNDLE_ID")
	
	prevDeviceIdEnv := os.Getenv("BITRISE_SIMCTL_PREVIOUS_DEVICE_ID")
	prevDeviceNameEnv := os.Getenv("BITRISE_SIMCTL_PREVIOUS_DEVICE_NAME")
	
	if localTesting == true {
		// commandEnv = "boot"
		// commandEnv = "shutdown"
		// commandEnv = "erase"
		commandEnv = "privacy"
		deviceNameEnv = "iphone 11"
		// deviceNameEnv := ""
		simctlAction = "grant"
		simctlService = "location-always"
		iosBundleId = "de.dealog.mobile.pilot"
		prevDeviceNameEnv = "FEF73DCA-1D03-4847-9D87-77CFCD977977"
	}

	if deviceNameEnv == "" {
		deviceNameEnv = prevDeviceNameEnv
	}
	
	deviceUUID := prevDeviceIdEnv
	if prevDeviceIdEnv == "" {
		deviceUUID = GetDeviceId(deviceNameEnv)
	}

	if deviceUUID == "" {
		fmt.Printf("DeviceUUID is empty")
		os.Exit(1)
	}

	xcrunSimctlShutdownStateLog := "Unable to shutdown device in current state: Shutdown"
	xcrunSimctlBootedStateLog := "Unable to boot device in current state: Booted"

	// boot/shutdown simulator
	if commandEnv == "boot" {
		SimctlExec(commandEnv, xcrunSimctlBootedStateLog, deviceUUID)
		SetEnv("BITRISE_SIMCTL_PREVIOUS_DEVICE_ID", deviceUUID)
		SetEnv("BITRISE_SIMCTL_PREVIOUS_DEVICE_NAME", deviceNameEnv)

		fmt.Printf("XCode device %s is booted", deviceUUID)
	} else if commandEnv == "shutdown" {
		SimctlExec(commandEnv, xcrunSimctlShutdownStateLog, deviceUUID)
		fmt.Printf("XCode device %s is shutdown", deviceUUID)
	} else if commandEnv == "erase" {
		SimctlExec(commandEnv, xcrunSimctlShutdownStateLog, deviceUUID)
		fmt.Printf("XCode device %s is shutdown", deviceUUID)
	} else if commandEnv == "privacy" {
		SimctlExecPermission(commandEnv, deviceUUID, simctlAction, simctlService, iosBundleId)
		fmt.Printf("Privacy set for XCode device %s", deviceUUID)
	} else {
		fmt.Printf("Command %s is unknown", commandEnv)
		os.Exit(1)
	}

	fmt.Printf("4")
	// Exec simulator command successful
	os.Exit(0)
}
