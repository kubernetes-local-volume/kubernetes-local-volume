/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package driver

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/logging"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/utils"
)

// ErrParse is an error that is returned when parse operation fails
var ErrParse = errors.New("Cannot parse output of blkid")

func formatDevice(devicePath, fstype string) error {
	output, err := exec.Command("mkfs", "-t", fstype, devicePath).CombinedOutput()
	if err != nil {
		return errors.New("FormatDevice error: " + string(output))
	}
	return nil
}

func checkFSType(devicePath string) (string, error) {
	// We use `file -bsL` to determine whether any filesystem type is detected.
	// If a filesystem is detected (ie., the output is not "data", we use
	// `blkid` to determine what the filesystem is. We use `blkid` as `file`
	// has inconvenient output.
	// We do *not* use `lsblk` as that requires udev to be up-to-date which
	// is often not the case when a device is erased using `dd`.
	output, err := exec.Command("file", "-bsL", devicePath).CombinedOutput()
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(string(output)) == "data" {
		return "", nil
	}
	output, err = exec.Command("blkid", "-c", "/dev/null", "-o", "export", devicePath).CombinedOutput()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Split(strings.TrimSpace(line), "=")
		if len(fields) != 2 {
			return "", ErrParse
		}
		if fields[0] == "TYPE" {
			return fields[1], nil
		}
	}
	return "", ErrParse
}

// create vg if not exist
func createVG(vgName string) (int, error) {
	pvNum := 0

	// check vg is created or not
	vgCmd := fmt.Sprintf("%s vgdisplay %s | grep 'VG Name' | grep %s | grep -v grep | wc -l", NsenterCmd, vgName, vgName)
	vgline, err := utils.Run(vgCmd)
	if err != nil {
		return 0, err
	}
	if strings.TrimSpace(vgline) == "1" {
		pvNumCmd := fmt.Sprintf("%s vgdisplay %s | grep 'Cur PV' | grep -v grep | awk '{print $3}'", NsenterCmd, vgName)
		if pvNumStr, err := utils.Run(pvNumCmd); err != nil {
			return 0, err
		} else if pvNum, err = strconv.Atoi(strings.TrimSpace(pvNumStr)); err != nil {
			return 0, err
		}
		return pvNum, nil
	}

	// device list
	localDeviceList := getDeviceList()
	localDeviceStr := strings.Join(localDeviceList, " ")

	logging.GetLogger().Infof("Find available device list : %s", localDeviceStr)

	// create pv
	pvAddCmd := fmt.Sprintf("%s pvcreate %s", NsenterCmd, localDeviceStr)
	_, err = utils.Run(pvAddCmd)
	if err != nil {
		logging.GetLogger().Errorf("Add PV from deviceList (%s) error : %s", localDeviceStr, err.Error())
		return 0, err
	}

	// create vg
	vgAddCmd := fmt.Sprintf("%s vgcreate %s %s", NsenterCmd, vgName, localDeviceStr)
	_, err = utils.Run(vgAddCmd)
	if err != nil {
		logging.GetLogger().Errorf("Add PV (%s) to VG: %s error: %s", localDeviceStr, strings.TrimSpace(vgName), err.Error())
		return 0, err
	}

	logging.GetLogger().Infof("Successful add Local Disks to VG (%s): %v", vgName, localDeviceList)
	return len(localDeviceList), nil
}

func getDeviceList() []string {
	devicePathPrefix := "/dev/vd"
	result := make([]string, 0)

	for index := 0; index < len(DeviceChars); index++ {
		devicePath := devicePathPrefix + DeviceChars[index]

		// check device exist
		if !utils.IsFileExisting(devicePath) {
			continue
		}

		// check is mounted
		if isMounted(devicePath) {
			continue
		}

		// check is used by other vg
		pvCmd := fmt.Sprintf("%s pvdisplay %s", NsenterCmd, devicePath)
		_, err := utils.Run(pvCmd)
		if err == nil {
			continue
		}

		result = append(result, devicePath)
	}
	return result
}

// isMounted return status of mount operation
func isMounted(mountPath string) bool {
	cmd := fmt.Sprintf("%s mount | grep %s | grep -v grep | wc -l", NsenterCmd, mountPath)
	out, err := utils.Run(cmd)
	if err != nil {
		return false
	}
	if strings.TrimSpace(out) == "0" {
		return false
	}
	return true
}

func isVgExist(vgName string) (bool, error) {
	vgCmd := fmt.Sprintf("%s vgdisplay %s | grep 'VG Name' | grep %s | grep -v grep | wc -l", NsenterCmd, vgName, vgName)
	vgline, err := utils.Run(vgCmd)
	if err != nil {
		return false, err
	}
	if strings.TrimSpace(vgline) == "1" {
		return true, nil
	}
	return false, nil
}
