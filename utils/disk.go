/*
 * @Author: bill
 * @Date: 2021-10-28 18:27:16
 * @LastEditors: bill
 * @LastEditTime: 2021-11-05 14:42:45
 * @Description:
 * @FilePath: /device-sector-migration/utils/disk.go
 */
package utils

import (
	"errors"
	"strconv"
	"strings"

	"device-sector-migration/global"
	"device-sector-migration/utils/exec"

	"github.com/shirou/gopsutil/v3/disk"
)

const (
	DISK_TYPE_SSD  = "ssd"
	DISK_TYPE_HDD  = "hdd"
	DISK_TYPE_NVME = "nvme"
)

type DiskStat struct {
	DevName    string `json:"dev_name"`
	Mountpoint string `json:"mountpoint"`
	Fstype     string `json:"fstype"`
	DiskType   string `json:"disk_type"`

	Total       uint64  `json:"total"`
	Free        uint64  `json:"free"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"usedPercent"`
}

// GetAllDiskInfo 获取所有磁盘基本信息
func GetAllDiskInfo() ([]DiskStat, error) {
	diskStat := make([]DiskStat, 0)
	parts, err := disk.Partitions(false)
	if err != nil {
		global.Logger.Error(err.Error())
		return diskStat, errors.New(err.Error())
	}
	for _, part := range parts {
		if part.Fstype == "ext4" && part.Mountpoint != "/boot/efi" {
			if info, err := disk.Usage(part.Mountpoint); err != nil {
				continue
			} else {
				var diskType string
				if part.Mountpoint == "/" {
					diskType = "system"
				} else {
					diskType = DiskType(part.Device)
				}
				diskStat = append(diskStat, DiskStat{
					DevName:    part.Device,
					Mountpoint: part.Mountpoint,
					Fstype:     part.Fstype,
					DiskType:   diskType,

					Total:       info.Total,
					Free:        info.Free,
					Used:        info.Used,
					UsedPercent: info.UsedPercent,
				})
			}
		}
	}
	return diskStat, nil
}

// GetSystemDiskInfo 获取系统盘基本信息
func GetSystemDiskInfo() (DiskStat, error) {
	diskStat := DiskStat{}
	parts, err := disk.Partitions(false)
	if err != nil {
		global.Logger.Error(err.Error())
		return diskStat, errors.New(err.Error())
	}
	for _, part := range parts {
		if part.Fstype == "ext4" && part.Mountpoint == "/" {
			if info, err := disk.Usage(part.Mountpoint); err != nil {
				return diskStat, err
			} else {
				diskStat.DevName = part.Device
				diskStat.Mountpoint = part.Mountpoint
				diskStat.Fstype = part.Fstype
				diskStat.DiskType = DiskType(part.Device)

				diskStat.Total = info.Total
				diskStat.Free = info.Free
				diskStat.Used = info.Used
				diskStat.UsedPercent = info.UsedPercent
				break
			}
		}
	}
	return diskStat, nil
}

// 判断磁盘是否是SSD型磁盘
// 如果返回1,则表示磁盘可旋转，那么就是HDD(机械硬盘)
// 如果返回0,则表示磁盘不可旋转，那么就是SSD(固态硬盘)
// nvme固态硬盘     M2接口      挂载nvme_
// ssd固态硬盘    sata接口      挂载ssd_
// hdd机械硬盘    sata接口      挂载hdd_
func DiskType(name string) string {
	dn := strings.Split(name, "/")
	if len(dn) < 1 {
		global.Logger.Error("传入设备名不符合规格")
		return ""
	}

	devname := dn[len(dn)-1]
	if strings.Contains(devname, DISK_TYPE_NVME) {
		return DISK_TYPE_NVME
	} else {
		out, err := exec.GetOutByNoninteractionExec("cat", "/sys/block/"+dn[len(dn)-1]+"/queue/rotational")
		if err != nil {
			global.Logger.Error("获取磁盘类型失败")
			return ""
		}
		lines := strings.Split(string(out), "\n")
		if i, err := strconv.Atoi(lines[0]); err == nil {
			if 0 == i {
				return DISK_TYPE_SSD
			} else if 1 == i {
				return DISK_TYPE_HDD
			}
		}
	}
	return ""
}
