package iostat

import (
	"github.com/shirou/gopsutil/v3/net"
)

type IOCountersStat struct {
	Name        string `json:"name"`
	BytesSent   uint64 `json:"bytes_sent"`
	BytesRecv   uint64 `json:"bytes_recv"`
	PacketsSent uint64 `json:"packets_sent"`
	PacketsRecv uint64 `json:"packets_recv"`
	ErrorsIn    uint64 `json:"errors_in"`
	ErrorsOut   uint64 `json:"errors_out"`
	DroppedIn   uint64 `json:"dropped_in"`
	DroppedOut  uint64 `json:"dropped_out"`
	FifoIn      uint64 `json:"fifo_in"`
	FifoOut     uint64 `json:"fifo_out"`
}

func GetData() (output []*IOCountersStat, err error) {
	ioCounters, err := net.IOCounters(true)
	if err != nil {
		return []*IOCountersStat{}, err
	}

	for _, block := range ioCounters {
		output = append(output, &IOCountersStat{
			Name:        block.Name,
			BytesSent:   block.BytesSent,
			BytesRecv:   block.BytesRecv,
			PacketsSent: block.PacketsSent,
			PacketsRecv: block.PacketsRecv,
			ErrorsIn:    block.Errin,
			ErrorsOut:   block.Errout,
			DroppedIn:   block.Dropin,
			DroppedOut:  block.Dropout,
			FifoIn:      block.Fifoin,
			FifoOut:     block.Fifoout,
		})
	}

	return output, nil
}
