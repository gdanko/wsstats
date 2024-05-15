package stats

import (
	"math"
	"time"

	"github.com/shirou/gopsutil/cpu"
)

type PercentStat struct {
	CPU       string  `json:"cpu"`
	User      float64 `json:"user"`
	System    float64 `json:"system"`
	Idle      float64 `json:"idle"`
	Nice      float64 `json:"nice"`
	Iowait    float64 `json:"iowait"`
	Irq       float64 `json:"irq"`
	Softirq   float64 `json:"softirq"`
	Steal     float64 `json:"steal"`
	Guest     float64 `json:"guest"`
	GuestNice float64 `json:"guestNice"`
}

func cpuTimeDeltas(t1, t2 cpu.TimesStat) cpu.TimesStat {
	return cpu.TimesStat{
		CPU:       t1.CPU,
		User:      math.Max(0, (t2.User - t1.User)),
		System:    math.Max(0, (t2.System - t1.System)),
		Idle:      math.Max(0, (t2.Idle - t1.Idle)),
		Nice:      math.Max(0, (t2.Nice - t1.Nice)),
		Iowait:    math.Max(0, (t2.Iowait - t1.Iowait)),
		Irq:       math.Max(0, (t2.Irq - t1.Irq)),
		Softirq:   math.Max(0, (t2.Softirq - t1.Softirq)),
		Steal:     math.Max(0, (t2.Steal - t1.Steal)),
		Guest:     math.Max(0, (t2.Guest - t1.Guest)),
		GuestNice: math.Max(0, (t2.GuestNice - t1.GuestNice)),
	}
}

func cpuTotalTime(timesDelta cpu.TimesStat) (total float64) {
	return timesDelta.User + timesDelta.System + timesDelta.Idle + timesDelta.Nice + timesDelta.Iowait + timesDelta.Irq + timesDelta.Softirq + timesDelta.Steal + timesDelta.Guest + timesDelta.GuestNice
}

func calculate(t1, t2 cpu.TimesStat) (percentStat PercentStat) {
	timesDelta := cpuTimeDeltas(t1, t2)
	allDelta := cpuTotalTime(timesDelta)
	scale := 100.0 / math.Max(1, allDelta)

	// fieldPercent := value * scale
	// fieldPercent = math.Min(math.Max(0.0, fieldPercent), 100.0)
	// cpuTimesMap[key] = math.Ceil(fieldPercent*100) / 100

	return PercentStat{
		CPU:       t1.CPU,
		User:      math.Ceil(math.Min(math.Max(0.0, (timesDelta.User*scale)), 100.0)*100) / 100,
		System:    math.Ceil(math.Min(math.Max(0.0, (timesDelta.System*scale)), 100.0)*100) / 100,
		Idle:      math.Ceil(math.Min(math.Max(0.0, (timesDelta.Idle*scale)), 100.0)*100) / 100,
		Nice:      math.Ceil(math.Min(math.Max(0.0, (timesDelta.Nice*scale)), 100.0)*100) / 100,
		Iowait:    math.Ceil(math.Min(math.Max(0.0, (timesDelta.Iowait*scale)), 100.0)*100) / 100,
		Irq:       math.Ceil(math.Min(math.Max(0.0, (timesDelta.Irq*scale)), 100.0)*100) / 100,
		Softirq:   math.Ceil(math.Min(math.Max(0.0, (timesDelta.Softirq*scale)), 100.0)*100) / 100,
		Steal:     math.Ceil(math.Min(math.Max(0.0, (timesDelta.Steal*scale)), 100.0)*100) / 100,
		Guest:     math.Ceil(math.Min(math.Max(0.0, (timesDelta.Guest*scale)), 100.0)*100) / 100,
		GuestNice: math.Ceil(math.Min(math.Max(0.0, (timesDelta.GuestNice*scale)), 100.0)*100) / 100,
	}
}

func GetCpuPercent(perCpu bool) ([]PercentStat, error) {
	var (
		lastPerCpuTimes  []cpu.TimesStat
		lastPerCpuTimes2 []cpu.TimesStat
		blocking         bool = false
		err              error
		interval         float64 = 1.0
		lastCpuTimes     []cpu.TimesStat
		lastCpuTimes2    []cpu.TimesStat
		output           []PercentStat
		t1               []cpu.TimesStat
	)

	lastCpuTimes, err = cpu.Times(false)
	if err != nil {
		return nil, err
	}
	lastCpuTimes2 = lastCpuTimes

	lastPerCpuTimes, err = cpu.Times(true)
	if err != nil {
		return nil, err
	}
	lastPerCpuTimes2 = lastPerCpuTimes

	if interval > 0.0 {
		blocking = true
	}

	if !perCpu {
		if blocking {
			t1, err = cpu.Times(false)
			if err != nil {
				return nil, err
			}
			time.Sleep(time.Duration(interval) * time.Second)
		} else {
			t1 = lastCpuTimes2
			if t1 == nil {
				t1, err = cpu.Times(false)
				if err != nil {
					return nil, err
				}
			}
		}
		lastCpuTimes2, err = cpu.Times(false)
		if err != nil {
			return nil, err
		}
		output = append(output, calculate(t1[0], lastCpuTimes2[0]))

	} else {
		if blocking {
			t1, err = cpu.Times(true)
			if err != nil {
				return nil, err
			}
			time.Sleep(time.Duration(interval) * time.Second)
		} else {
			t1 = lastPerCpuTimes2
			if t1 == nil {
				t1, err = cpu.Times(true)
				if err != nil {
					return nil, err
				}
			}
		}
		lastPerCpuTimes2, err = cpu.Times(true)
		if err != nil {
			return nil, err
		}

		for count := 0; count < len(t1); count++ {
			output = append(output, calculate(t1[count], lastPerCpuTimes2[count]))
		}
	}
	return output, nil
}
