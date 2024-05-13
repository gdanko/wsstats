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

	"github.com/gdanko/wezterm-system-stats/internal"
	"github.com/gdanko/wezterm-system-stats/iostat"
	"github.com/gdanko/wezterm-system-stats/util"
	flags "github.com/jessevdk/go-flags"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

type Netspeed struct {
	JSON          bool
	PrintVersion  bool
	Lockfile      string
	OutputFile    string
	StartTime     uint64
	Logger        *logrus.Logger
	Logfile       string
	LogfileHandle *os.File
}

type Options struct {
	JSON         bool `short:"j" long:"json" description:"Save the output to /tmp/netspeed.json instead of to STDOUT.\nOnly the current iteration is saved to file."`
	PrintVersion bool `short:"V" long:"version" description:"Print program version"`
}

type NetspeedInterfaceData struct {
	Interface   string  `json:"interface"`
	BytesRecv   float64 `json:"bytes_recv"`
	BytesSent   float64 `json:"bytes_sent"`
	PacketsRecv uint64  `json:"packets_recv"`
	PacketsSent uint64  `json:"packets_sent"`
}

type NetspeedData struct {
	Timestamp  uint64                  `json:"timestamp"`
	Interfaces []NetspeedInterfaceData `json:"interfaces"`
	StartTime  uint64                  `json:"start_time"`
	RunTime    uint64                  `json:"run_time"`
}

type IOStatData struct {
	Interfaces []iostat.IOStatData `json:"interfaces"`
}

func (n *Netspeed) init(args []string) error {
	var (
		err    error
		opts   Options
		parser *flags.Parser
	)

	opts = Options{}
	parser = flags.NewParser(&opts, flags.Default)
	parser.Usage = `[-j, --json] [-V, --version] 
  netspeed prints bytes in/out per second and packets sent/received per second for all interfaces`
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	n.JSON = opts.JSON
	n.PrintVersion = opts.PrintVersion
	n.Lockfile = "/tmp/netspeed.lock"
	n.OutputFile = "/tmp/netspeed.json"
	n.StartTime = util.GetTimestamp()
	n.Logfile = "/tmp/netspeed.log"
	n.Logger = logrus.New()

	n.LogfileHandle, err = os.OpenFile(n.Logfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("failed to create the log file handle: %s", err.Error())
	}

	n.Logger.SetFormatter(&prefixed.TextFormatter{
		DisableColors:   true,
		ForceFormatting: true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	n.Logger.SetOutput(n.LogfileHandle)
	n.Logger.SetReportCaller(true)
	n.Logger.Info("Starting")

	return nil
}

func (n *Netspeed) ExitError(errorMessage error) {
	n.CleanUp()
	n.Logger.Error(errorMessage.Error())
	n.LogfileHandle.Close()
	os.Exit(1)
}

func (n *Netspeed) ExitCleanly() {
	n.CleanUp()
	n.LogfileHandle.Close()
	os.Exit(0)
}

func (n *Netspeed) CreateLockfile() (err error) {
	f, err := os.Create(n.Lockfile)
	if err != nil {
		return fmt.Errorf("failed to create the lockfile \"%s\"", n.Lockfile)
	}
	defer f.Close()

	return nil
}

func (n *Netspeed) ShowVersion() {
	fmt.Fprintf(os.Stdout, "netspeed version %s\n", internal.Version(false, true))
}

func (n *Netspeed) ProcessOutput(netspeedData NetspeedData) {
	jsonBytes, err := json.Marshal(netspeedData)
	if err != nil {
		n.ExitError(err)
	}

	if !n.JSON {
		fmt.Fprintln(os.Stdout, string(jsonBytes))
	} else {
		err = os.WriteFile(n.OutputFile, jsonBytes, 0644)
		if err != nil {
			n.ExitError(err)
		}
	}
}

func (n *Netspeed) CleanUp() {
	for _, filename := range []string{n.OutputFile, n.Lockfile} {
		err := util.DeleteFile(filename)
		if err != nil {
			n.Logger.Warn(err.Error())
		}
	}
}

func (n *Netspeed) FindInterface(interfaceName string, interfaceList []iostat.IOStatData) (iostatEntry iostat.IOStatData, err error) {
	for _, iostatEntry = range interfaceList {
		if interfaceName == iostatEntry.Interface {
			return iostatEntry, nil
		}
	}
	return iostat.IOStatData{}, fmt.Errorf("the interface \"%s\" was not found in this block", interfaceName)
}

func Run(ctx context.Context, n *Netspeed) error {
	if n.PrintVersion {
		n.ShowVersion()
		n.ExitCleanly()
	}

	var iostatDataOld = IOStatData{}
	var iostatDataNew = IOStatData{}
	var netspeedData = NetspeedData{}

	if util.FileExists(n.Lockfile) {
		return fmt.Errorf("the lockfile \"%s\" already exists - the program is probably already running", n.Lockfile)
	}

	err := n.CreateLockfile()
	if err != nil {
		return err
	}

	// Get the first sample
	data, err := iostat.GetData()
	if err != nil {
		return err
	}
	iostatDataOld.Interfaces = data
	time.Sleep(1 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			// Clear out New at each iteration
			currentTimestamp := util.GetTimestamp()
			netspeedData = NetspeedData{
				Timestamp: currentTimestamp,
				StartTime: n.StartTime,
				RunTime:   currentTimestamp - n.StartTime,
			}

			data, err := iostat.GetData()
			if err != nil {
				// breaking out here will cause disruption for one iteration but
				// should normalize iterself naturally
				break
			}
			iostatDataNew.Interfaces = data

			for _, iostatBlock := range iostatDataNew.Interfaces {
				var foundInOld, foundInNew = true, true

				interfaceName := iostatBlock.Interface
				interfaceOld, err := n.FindInterface(interfaceName, iostatDataOld.Interfaces)
				if err != nil {
					n.Logger.Warnf("interface \"%s\" not found in the old data set", interfaceName)
					foundInOld = false
				}
				foundInOld = true

				interfaceNew, err := n.FindInterface(interfaceName, iostatDataNew.Interfaces)
				if err != nil {
					n.Logger.Warnf("interface \"%s\" not found in the new data set", interfaceName)
					foundInNew = false
				}

				// Only add the block if the interface name was found in both old and new blocks
				if foundInOld && foundInNew {
					netspeedData.Interfaces = append(netspeedData.Interfaces, NetspeedInterfaceData{
						Interface:   interfaceNew.Interface,
						BytesSent:   interfaceNew.BytesSent - interfaceOld.BytesSent,
						BytesRecv:   interfaceNew.BytesRecv - interfaceOld.BytesRecv,
						PacketsSent: interfaceNew.PacketsSent - interfaceOld.PacketsSent,
						PacketsRecv: interfaceNew.PacketsRecv - interfaceOld.PacketsRecv,
					})
				}
			}

			n.ProcessOutput(netspeedData)

			iostatDataOld.Interfaces = iostatDataNew.Interfaces
			time.Sleep(1 * time.Second)
		}
	}
}

func main() {
	var err error

	n := &Netspeed{}
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case s := <-signalChan:
			switch s {
			case syscall.SIGINT:
				n.Logger.Info("Got SIGINT, exiting.")
				n.ExitCleanly()
			case syscall.SIGTERM:
				n.Logger.Info("Got SIGTERM, exiting.")
				n.ExitCleanly()
			case syscall.SIGHUP:
				n.Logger.Info("Got SIGHUP, reloading.")
				n.init(os.Args)
			}
		case <-ctx.Done():
			n.Logger.Info("Exiting normally.")
			n.ExitCleanly()
		}
	}()

	defer func() {
		cancel()
	}()

	err = n.init(os.Args)
	if err != nil {
		n.ExitError(err)
	}

	if err := Run(ctx, n); err != nil {
		n.ExitError(err)
	}
}
