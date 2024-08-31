package main

// https://ieftimov.com/posts/four-steps-daemonize-your-golang-programs/

import (
	"encoding/json"
	"fmt"
	"os"

	test_runner "github.com/gdanko/wsstats/gather"
	"github.com/gdanko/wsstats/internal"
	"github.com/gdanko/wsstats/iostat"
	"github.com/gdanko/wsstats/stats"
	"github.com/gdanko/wsstats/util"
	flags "github.com/jessevdk/go-flags"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/sirupsen/logrus"
)

type Wezterm struct {
	PrintVersion   bool
	All            bool
	CPU            bool
	Disk           bool
	Host           bool
	Load           bool
	Memory         bool
	Net            bool
	Swap           bool
	OutputData     map[string]interface{}
	OutputFile     string
	StartTime      uint64
	Logger         *logrus.Logger
	RunTimeCurrent uint64
	IostatDataOld  test_runner.IOStatData
}

type Options struct {
	All          bool `short:"a" long:"all" description:"Report all available system info (default)"`
	CPU          bool `short:"c" long:"cpu" description:"Report system CPU usage"`
	Disk         bool `short:"d" long:"disk" description:"Report system disk usage"`
	Host         bool `long:"host" description:"Report system host information"`
	Load         bool `short:"l" long:"load" description:"Report system load averages"`
	Memory       bool `short:"m" long:"memory" description:"Report system memory usage"`
	Net          bool `short:"n" long:"network" description:"Report network throughput information"`
	Swap         bool `short:"s" long:"swap" description:"Report swap memory usage"`
	PrintVersion bool `short:"V" long:"version" description:"Print program version"`
}

func (w *Wezterm) init(args []string) error {
	var (
		opts   Options
		parser *flags.Parser
	)

	opts = Options{}
	parser = flags.NewParser(&opts, flags.Default)
	parser.Usage = `[OPTIONS] 
  wsstats gathers and writes system statistics in a way easily consumable by WezTerm`
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	w.All = opts.All
	w.CPU = opts.CPU
	w.Disk = opts.Disk
	w.Host = opts.Host
	w.Load = opts.Load
	w.Memory = opts.Memory
	w.Net = opts.Net
	w.Swap = opts.Swap
	w.Logger = logrus.New()
	w.OutputFile = "/tmp/wsstats.json"
	w.PrintVersion = opts.PrintVersion
	w.StartTime = util.GetTimestamp()

	if len(args) == 1 || w.All {
		w.CPU, w.Disk, w.Host, w.Load, w.Memory, w.Net, w.Swap = true, true, true, true, true, true, true
	}
	return nil
}

func (w *Wezterm) ExitError(errorMessage error) {
	os.Exit(1)
}

func (w *Wezterm) ExitCleanly() {
	os.Exit(0)
}

func (w *Wezterm) ShowVersion() {
	fmt.Fprintf(os.Stdout, "wsstats version %s\n", internal.Version(false, true))
}

func (w *Wezterm) ProcessOutput() {
	jsonBytes, err := json.MarshalIndent(w.OutputData, "", "    ")
	if err != nil {
		w.ExitError(err)
	}

	err = os.WriteFile(w.OutputFile, jsonBytes, 0644)
	if err != nil {
		w.ExitError(err)
	}
}

func (w *Wezterm) ParallelTester() {
	// Fetch all data at once using Go channels
	w.OutputData = make(map[string]interface{})
	if w.CPU {
		cpuPercentChannel := make(chan func() ([]stats.PercentStat, error))
		go test_runner.GetCpuPercent(cpuPercentChannel)
		cpuPercent, err := (<-cpuPercentChannel)()
		if err == nil {
			w.OutputData["cpu"] = cpuPercent
		}
	}

	if w.Disk {
		diskUsageChannel := make(chan func() ([]stats.DiskUsageData, error))
		go test_runner.GetDiskUsage(diskUsageChannel)
		diskUsage, err := (<-diskUsageChannel)()
		if err == nil {
			w.OutputData["disk"] = diskUsage
		}
	}

	if w.Host {
		hostInformationChannel := make(chan func() (stats.HostInformation, error))
		go test_runner.GetHostInformation(hostInformationChannel)
		hostInformation, err := (<-hostInformationChannel)()
		if err == nil {
			w.OutputData["host"] = hostInformation
		}
	}

	if w.Load {
		loadAveragesChannel := make(chan func() (*load.AvgStat, error))
		go test_runner.GetLoadAverages(loadAveragesChannel)
		loadAverages, err := (<-loadAveragesChannel)()
		if err == nil {
			w.OutputData["load"] = loadAverages
		}
	}

	if w.Memory {
		memoryUsageChannel := make(chan func() (*mem.VirtualMemoryStat, error))
		go test_runner.GetMemoryUsage(memoryUsageChannel)
		memoryUsage, err := (<-memoryUsageChannel)()
		if err == nil {
			w.OutputData["memory"] = memoryUsage
		}
	}

	if w.Net {
		networkThroughputChannel := make(chan func() ([]*iostat.IOCountersStat, error))
		go test_runner.GetNetworkThroughput(networkThroughputChannel)
		networkThroughput, err := (<-networkThroughputChannel)()
		if err == nil {
			w.OutputData["network"] = networkThroughput
		}

	}

	if w.Swap {
		swapUsageChannel := make(chan func() (*mem.SwapMemoryStat, error))
		go test_runner.GetSwapUsage(swapUsageChannel)
		swapUsage, err := (<-swapUsageChannel)()
		if err == nil {
			w.OutputData["swap"] = swapUsage
		}
	}
}

func main() {
	w := &Wezterm{}
	err := w.init(os.Args)
	if err != nil {
		w.ExitError(err)
	}
	w.ParallelTester()
	w.ProcessOutput()
}
