package iostat

import (
	"github.com/shirou/gopsutil/net"
)

type IOStatData struct {
	Interface   string  `json:"interface"`
	BytesRecv   float64 `json:"bytes_recv"`
	BytesSent   float64 `json:"bytes_sent"`
	PacketsRecv uint64  `json:"packets_recv"`
	PacketsSent uint64  `json:"packets_sent"`
}

func GetData() (output []IOStatData, err error) {
	ioCounters, err := net.IOCounters(true)
	if err != nil {
		return []IOStatData{}, err
	}

	for _, ifaceBlock := range ioCounters {
		output = append(output, IOStatData{
			Interface:   ifaceBlock.Name,
			BytesSent:   float64(ifaceBlock.BytesSent),
			BytesRecv:   float64(ifaceBlock.BytesRecv),
			PacketsSent: uint64(ifaceBlock.PacketsSent),
			PacketsRecv: uint64(ifaceBlock.PacketsRecv),
		})
	}
	return output, nil
}
