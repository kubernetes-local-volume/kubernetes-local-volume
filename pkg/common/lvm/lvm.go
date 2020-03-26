package lvm

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/logging"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/types"
	"github.com/kubernetes-local-volume/kubernetes-local-volume/pkg/common/utils"
)

// create vg if not exist
func CreateVG(vgName string) (int, error) {
	pvNum := 0

	// check vg is created or not
	vgCmd := fmt.Sprintf("%s vgdisplay %s | grep 'VG Name' | grep %s | grep -v grep | wc -l", types.NsenterCmd, vgName, vgName)
	vgline, err := utils.Run(vgCmd)
	if err != nil {
		return 0, err
	}
	if strings.TrimSpace(vgline) == "1" {
		pvNumCmd := fmt.Sprintf("%s vgdisplay %s | grep 'Cur PV' | grep -v grep | awk '{print $3}'", types.NsenterCmd, vgName)
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
	pvAddCmd := fmt.Sprintf("%s pvcreate %s", types.NsenterCmd, localDeviceStr)
	_, err = utils.Run(pvAddCmd)
	if err != nil {
		logging.GetLogger().Errorf("Add PV from deviceList (%s) error : %s", localDeviceStr, err.Error())
		return 0, err
	}

	// create vg
	vgAddCmd := fmt.Sprintf("%s vgcreate %s %s", types.NsenterCmd, vgName, localDeviceStr)
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

	for index := 0; index < len(types.DeviceChars); index++ {
		devicePath := devicePathPrefix + types.DeviceChars[index]

		// check device exist
		if !utils.IsFileExisting(devicePath) {
			continue
		}

		// check is mounted
		if isMounted(devicePath) {
			continue
		}

		// check is used by other vg
		pvCmd := fmt.Sprintf("%s pvdisplay %s", types.NsenterCmd, devicePath)
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
	cmd := fmt.Sprintf("%s mount | grep %s | grep -v grep | wc -l", types.NsenterCmd, mountPath)
	out, err := utils.Run(cmd)
	if err != nil {
		return false
	}
	if strings.TrimSpace(out) == "0" {
		return false
	}
	return true
}

type VGSOutput struct {
	Report []struct {
		Vg []struct {
			Name              string `json:"vg_name"`
			UUID              string `json:"vg_uuid"`
			VgSize            uint64 `json:"vg_size,string"`
			VgFree            uint64 `json:"vg_free,string"`
			VgExtentSize      uint64 `json:"vg_extent_size,string"`
			VgExtentCount     uint64 `json:"vg_extent_count,string"`
			VgFreeExtentCount uint64 `json:"vg_free_count,string"`
			VgTags            string `json:"vg_tags"`
		} `json:"vg"`
	} `json:"report"`
}

func VGTotalSize(vgName string) (uint64, error) {
	result := new(VGSOutput)
	if err := run("vgs", result, "--options=vg_size", vgName); err != nil {
		return 0, err
	}
	for _, report := range result.Report {
		for _, vg := range report.Vg {
			return vg.VgSize, nil
		}
	}
	return 0, nil
}
