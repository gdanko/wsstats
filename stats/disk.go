package stats

import (
	"strings"

	"github.com/gdanko/wsstats/util"
	"github.com/shirou/gopsutil/disk"
)

type DiskUsageData struct {
	DeviceName        string  `json:"device"`
	MountPoint        string  `json:"mount_point"`
	FileSystemType    string  `json:"fs_type"`
	FileSystemOptions string  `json:"fs_options"`
	Total             uint64  `json:"total"`
	Free              uint64  `json:"free"`
	Used              uint64  `json:"used"`
	UsedPercent       float64 `json:"used_percent"`
	FreePercent       float64 `json:"free_percent"`
}

func GetDiskUsage() (disks []DiskUsageData, err error) {
	diskPartitions, err := disk.Partitions(true)
	if err != nil {
		return disks, err
	}
	for _, diskItem := range diskPartitions {
		if strings.HasPrefix(diskItem.Device, "/dev/") {
			diskUsage, err := disk.Usage(diskItem.Mountpoint)
			if err != nil {
				return disks, err
			}
			disks = append(disks, DiskUsageData{
				DeviceName:        diskItem.Device,
				MountPoint:        diskItem.Mountpoint,
				FileSystemType:    diskItem.Fstype,
				FileSystemOptions: diskItem.Opts,
				Total:             diskUsage.Total,
				Used:              diskUsage.Used,
				Free:              diskUsage.Free,
				UsedPercent:       util.RoundTo((float64(float64(diskUsage.Used)/float64(diskUsage.Total)) * 100), 2),
				FreePercent:       util.RoundTo((float64(float64(diskUsage.Free)/float64(diskUsage.Total)) * 100), 2),
			})
		}
	}
	return disks, nil
}
