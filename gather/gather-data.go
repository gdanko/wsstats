package test_runner

import (
	"time"

	"github.com/gdanko/wsstats/iostat"
	"github.com/gdanko/wsstats/stats"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
)

type IOStatData struct {
	Interfaces []*iostat.IOCountersStat `json:"interfaces"`
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

func GetHostInformation(c chan func() (stats.HostInformation, error)) {
	c <- (func() (stats.HostInformation, error) {
		hostInformation, err := stats.GetHostInformation()
		return hostInformation, err
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

func GetNetworkThroughput(c chan func() ([]*iostat.IOCountersStat, error)) {
	c <- (func() ([]*iostat.IOCountersStat, error) {
		firstSample, err := iostat.GetData()
		if err != nil {
			return []*iostat.IOCountersStat{}, err
		}
		time.Sleep(1 * time.Second)
		secondSample, err := iostat.GetData()
		if err != nil {
			return []*iostat.IOCountersStat{}, err
		}

		output := []*iostat.IOCountersStat{}
		for i, _ := range firstSample {
			output = append(output, &iostat.IOCountersStat{
				Name:        firstSample[i].Name,
				BytesSent:   secondSample[i].BytesSent - firstSample[i].BytesSent,
				BytesRecv:   secondSample[i].BytesRecv - firstSample[i].BytesRecv,
				PacketsSent: secondSample[i].PacketsSent - firstSample[i].PacketsSent,
				PacketsRecv: secondSample[i].PacketsRecv - firstSample[i].PacketsRecv,
				ErrorsIn:    secondSample[i].ErrorsIn - firstSample[i].ErrorsIn,
				ErrorsOut:   secondSample[i].ErrorsOut - firstSample[i].ErrorsOut,
				DroppedIn:   secondSample[i].DroppedIn - firstSample[i].DroppedIn,
				DroppedOut:  secondSample[i].DroppedOut - firstSample[i].DroppedOut,
				FifoIn:      secondSample[i].FifoIn - firstSample[i].FifoIn,
				FifoOut:     secondSample[i].FifoOut - firstSample[i].FifoOut,
			})
		}
		return output, err
	})
}
