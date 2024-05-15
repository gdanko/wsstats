package test_runner

import (
	"fmt"

	"github.com/gdanko/wezterm-system-stats/iostat"
	"github.com/gdanko/wezterm-system-stats/stats"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/sirupsen/logrus"
)

type NetworkInterfaceData struct {
	Interface   string  `json:"interface"`
	BytesRecv   float64 `json:"bytes_recv"`
	BytesSent   float64 `json:"bytes_sent"`
	PacketsRecv uint64  `json:"packets_recv"`
	PacketsSent uint64  `json:"packets_sent"`
}

type IOStatData struct {
	Interfaces []iostat.IOStatData `json:"interfaces"`
}

func GetCpuPercent(c chan func() ([]stats.PercentStat, error)) {
	c <- (func() ([]stats.PercentStat, error) {
		cpuPercent, err := stats.GetCpuPercent(false)
		return cpuPercent, err
	})
}

func GetDiskUsage(c chan func() ([]stats.DiskUsageData, error)) {
	c <- (func() ([]stats.DiskUsageData, error) {
		diskUsage, err := stats.GetDiskUsage()
		return diskUsage, err
	})
}

func GetLoadAverages(c chan func() (*load.AvgStat, error)) {
	c <- (func() (*load.AvgStat, error) {
		loadAverages, err := stats.GetLoadAverages()
		return loadAverages, err
	})
}

func GetMemoryUsage(c chan func() (*mem.VirtualMemoryStat, error)) {
	c <- (func() (*mem.VirtualMemoryStat, error) {
		memoryUsage, err := stats.GetMemoryUsage()
		return memoryUsage, err
	})
}

func GetSwapUsage(c chan func() (*mem.SwapMemoryStat, error)) {
	c <- (func() (*mem.SwapMemoryStat, error) {
		swapUsage, err := stats.GetSwapUsage()
		return swapUsage, err
	})
}

func GetNetworkThroughput(c chan func(logger *logrus.Logger, iostatDataOld IOStatData) (interfaces []NetworkInterfaceData, iostatDataNew IOStatData, err error)) {
	c <- func(logger *logrus.Logger, iostatDataOld IOStatData) (interfaces []NetworkInterfaceData, iostatDataNew IOStatData, err error) {
		data, err := iostat.GetData()
		if err != nil {
			return interfaces, iostatDataNew, err
		}
		iostatDataNew.Interfaces = data

		for _, iostatBlock := range iostatDataNew.Interfaces {
			var foundInOld, foundInNew = true, true

			interfaceName := iostatBlock.Interface
			interfaceOld, err := findInterface(interfaceName, iostatDataOld.Interfaces)
			if err != nil {
				logger.Warnf("interface \"%s\" not found in the old data set", interfaceName)
				foundInOld = false
			}
			foundInOld = true

			interfaceNew, err := findInterface(interfaceName, iostatDataNew.Interfaces)
			if err != nil {
				logger.Warnf("interface \"%s\" not found in the new data set", interfaceName)
				foundInNew = false
			}

			if foundInOld && foundInNew {
				interfaces = append(interfaces, NetworkInterfaceData{
					Interface:   interfaceNew.Interface,
					BytesSent:   interfaceNew.BytesSent - interfaceOld.BytesSent,
					BytesRecv:   interfaceNew.BytesRecv - interfaceOld.BytesRecv,
					PacketsSent: interfaceNew.PacketsSent - interfaceOld.PacketsSent,
					PacketsRecv: interfaceNew.PacketsRecv - interfaceOld.PacketsRecv,
				})
			}
		}
		return interfaces, iostatDataNew, nil
	}
}

func findInterface(interfaceName string, interfaceList []iostat.IOStatData) (iostatEntry iostat.IOStatData, err error) {
	for _, iostatEntry = range interfaceList {
		if interfaceName == iostatEntry.Interface {
			return iostatEntry, nil
		}
	}
	return iostat.IOStatData{}, fmt.Errorf("the interface \"%s\" was not found in this block", interfaceName)
}
