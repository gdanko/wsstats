package main

// https://ieftimov.com/posts/four-steps-daemonize-your-golang-programs/

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gdanko/wsstats/internal"
	"github.com/gdanko/wsstats/iostat"
	"github.com/gdanko/wsstats/stats"
	test_runner "github.com/gdanko/wsstats/test-runner"
	"github.com/gdanko/wsstats/util"
	flags "github.com/jessevdk/go-flags"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

type Wezterm struct {
	PrintVersion   bool
	All            bool
	CPU            bool
	Disk           bool
	Load           bool
	Memory         bool
	Net            bool
	Swap           bool
	Lockfile       string
	OutputFile     string
	StartTime      uint64
	Logger         *logrus.Logger
	Logfile        string
	LogfileHandle  *os.File
	RunTimeCurrent uint64
	IostatDataOld  test_runner.IOStatData
}

type Options struct {
	All          bool `short:"a" long:"all" description:"Report all available system info (default)"`
	CPU          bool `short:"c" long:"cpu" description:"Report system CPU usage"`
	Disk         bool `short:"d" long:"disk" description:"Report system disk usage"`
	Load         bool `short:"l" long:"load" description:"Report system load averages"`
	Memory       bool `short:"m" long:"memory" description:"Report system memory usage"`
	Net          bool `short:"n" long:"network" description:"Report network throughput information"`
	Swap         bool `short:"s" long:"swap" description:"Report swap memory usage"`
	PrintVersion bool `short:"V" long:"version" description:"Print program version"`
}

func (w *Wezterm) init(args []string) error {
	var (
		err    error
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
	w.Load = opts.Load
	w.Memory = opts.Memory
	w.Net = opts.Net
	w.Swap = opts.Swap
	w.Lockfile = "/tmp/wsstats.lock"
	w.Logfile = "/tmp/wsstats.log"
	w.Logger = logrus.New()
	w.OutputFile = "/tmp/wsstats.json"
	w.PrintVersion = opts.PrintVersion
	w.StartTime = util.GetTimestamp()

	if len(args) == 1 || w.All {
		w.CPU, w.Disk, w.Load, w.Memory, w.Net, w.Swap = true, true, true, true, true, true
	}

	w.LogfileHandle, err = os.OpenFile(w.Logfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("failed to create the log file handle: %s", err.Error())
	}

	w.Logger.SetFormatter(&prefixed.TextFormatter{
		DisableColors:   true,
		ForceFormatting: true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	w.Logger.SetOutput(w.LogfileHandle)
	w.Logger.SetReportCaller(true)
	w.Logger.Info("Starting")

	return nil
}

func (w *Wezterm) ExitError(errorMessage error) {
	w.CleanUp()
	w.Logger.Error(errorMessage.Error())
	w.LogfileHandle.Close()
	os.Exit(1)
}

func (w *Wezterm) ExitCleanly() {
	w.CleanUp()
	w.LogfileHandle.Close()
	os.Exit(0)
}

func (w *Wezterm) CreateLockfile() (err error) {
	f, err := os.Create(w.Lockfile)
	if err != nil {
		return fmt.Errorf("failed to create the lockfile \"%s\"", w.Lockfile)
	}
	defer f.Close()

	return nil
}

func (w *Wezterm) ShowVersion() {
	fmt.Fprintf(os.Stdout, "wsstats version %s\n", internal.Version(false, true))
}

func (w *Wezterm) ProcessOutput(WeztermStatsData map[string]interface{}) {
	jsonBytes, err := json.MarshalIndent(WeztermStatsData, "", "    ")
	if err != nil {
		w.ExitError(err)
	}

	// fmt.Println(string(jsonBytes))

	err = os.WriteFile(w.OutputFile, jsonBytes, 0644)
	if err != nil {
		w.ExitError(err)
	}
}

func (w *Wezterm) CleanUp() {
	for _, filename := range []string{w.OutputFile, w.Lockfile} {
		err := util.DeleteFile(filename)
		if err != nil {
			w.Logger.Warn(err.Error())
		}
	}
}

func (w *Wezterm) ParallelTester() (output map[string]interface{}) {
	// Fetch all data at once using Go channels
	output = make(map[string]interface{})
	if w.CPU {
		cpuPercentChannel := make(chan func() ([]stats.PercentStat, error))
		go test_runner.GetCpuPercent(cpuPercentChannel)
		cpuPercent, err := (<-cpuPercentChannel)()
		if err == nil {
			output["cpu"] = cpuPercent
		}
	}

	if w.Disk {
		diskUsageChannel := make(chan func() ([]stats.DiskUsageData, error))
		go test_runner.GetDiskUsage(diskUsageChannel)
		diskUsage, err := (<-diskUsageChannel)()
		if err == nil {
			output["disk"] = diskUsage
		}
	}

	if w.Load {
		loadAveragesChannel := make(chan func() (*load.AvgStat, error))
		go test_runner.GetLoadAverages(loadAveragesChannel)
		loadAverages, err := (<-loadAveragesChannel)()
		if err == nil {
			output["load"] = loadAverages
		}
	}

	if w.Memory {
		memoryUsageChannel := make(chan func() (*mem.VirtualMemoryStat, error))
		go test_runner.GetMemoryUsage(memoryUsageChannel)
		memoryUsage, err := (<-memoryUsageChannel)()
		if err == nil {
			output["memory"] = memoryUsage
		}
	}

	if w.Net {
		networkThroughputChannel := make(chan func(logger *logrus.Logger, iostatDataOld test_runner.IOStatData) ([]test_runner.NetworkInterfaceData, test_runner.IOStatData, error))
		go test_runner.GetNetworkThroughput(networkThroughputChannel)
		networkThroughput, iostatDataNew, err := (<-networkThroughputChannel)(w.Logger, w.IostatDataOld)
		if err == nil {
			w.IostatDataOld = iostatDataNew
			output["network"] = networkThroughput
		}
	}

	if w.Swap {
		swapUsageChannel := make(chan func() (*mem.SwapMemoryStat, error))
		go test_runner.GetSwapUsage(swapUsageChannel)
		swapUsage, err := (<-swapUsageChannel)()
		if err == nil {
			output["swap"] = swapUsage
		}
	}
	return output
}

func Run(ctx context.Context, w *Wezterm) error {
	if w.PrintVersion {
		w.ShowVersion()
		w.ExitCleanly()
	}

	if util.FileExists(w.Lockfile) {
		return fmt.Errorf("the lockfile \"%s\" already exists - the program is probably already running", w.Lockfile)
	}

	err := w.CreateLockfile()
	if err != nil {
		return err
	}

	// Get the first network sample
	if w.Net {
		data, err := iostat.GetData()
		if err != nil {
			return err
		}
		w.IostatDataOld.Interfaces = data
		time.Sleep(1 * time.Second)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			w.RunTimeCurrent = util.GetTimestamp()

			output := w.ParallelTester()
			output["timestamp"] = w.RunTimeCurrent
			output["start_time"] = w.StartTime
			output["run_time"] = w.RunTimeCurrent - w.StartTime

			w.ProcessOutput(output)

			if w.RunTimeCurrent == util.GetTimestamp() {
				time.Sleep(1 * time.Second)
			}
		}
	}
}

func main() {
	var err error
	w := &Wezterm{}
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case s := <-signalChan:
			switch s {
			case syscall.SIGINT:
				w.Logger.Info("Got SIGINT, exiting.")
				w.ExitCleanly()
			case syscall.SIGTERM:
				w.Logger.Info("Got SIGTERM, exiting.")
				w.ExitCleanly()
			case syscall.SIGHUP:
				w.Logger.Info("Got SIGHUP, reloading.")
				w.init(os.Args)
			}
		case <-ctx.Done():
			w.Logger.Info("Exiting normally.")
			w.ExitCleanly()
		}
	}()

	defer func() {
		cancel()
	}()

	err = w.init(os.Args)
	if err != nil {
		w.ExitError(err)
	}

	if err := Run(ctx, w); err != nil {
		w.ExitError(err)
	}
}
