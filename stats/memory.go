package stats

import "github.com/shirou/gopsutil/v3/mem"

func GetMemoryUsage() (memory *mem.VirtualMemoryStat, err error) {
	memory, err = mem.VirtualMemory()
	if err != nil {
		return memory, err
	}
	return memory, nil
}

func GetSwapUsage() (swap *mem.SwapMemoryStat, err error) {
	swap, err = mem.SwapMemory()
	if err != nil {
		return swap, err
	}
	return swap, nil
}
