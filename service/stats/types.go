package stats

import (
	"keeper/pkg/percent"

	"github.com/shirou/gopsutil/v4/host"
)

type HostMetrics struct {
	Info        InfoStat           `json:"info"`
	Load        LoadAvg            `json:"load"`
	CPU         CPUStats           `json:"cpu"`
	Memory      MemoryStats        `json:"memory"`
	Disk        []DiskStats        `json:"disk"`
	Network     []NetStats         `json:"network"`
	Temperature []TemperatureStats `json:"temperature"`
}

type InfoStat host.InfoStat

// LoadAvg holds the 1, 5 and 15 minute load averages
type LoadAvg struct {
	One     float64 `json:"one"`
	Five    float64 `json:"five"`
	Fifteen float64 `json:"fifteen"`
}

type CPUStats struct {
	Usage  percent.Percent `json:"usage"`
	Cores  int             `json:"cores"`
	User   percent.Percent `json:"user"`
	System percent.Percent `json:"system"`
	Idle   percent.Percent `json:"idle"`
}

type MemoryStats struct {
	Total     uint64 `json:"total"`
	Used      uint64 `json:"used"`
	Free      uint64 `json:"free"`
	SwapTotal uint64 `json:"swap_total"`
	SwapUsed  uint64 `json:"swap_used"`
}

type DiskStats struct {
	Mountpoint string `json:"mountpoint"`
	Device     string `json:"device"`
	Total      uint64 `json:"total"`
	Used       uint64 `json:"used"`
	Free       uint64 `json:"free"`
	ReadBytes  uint64 `json:"read_bytes"`
	WriteBytes uint64 `json:"write_bytes"`
}

type NetStats struct {
	Interface string `json:"interface"`
	RxBytes   uint64 `json:"rx_bytes"`
	TxBytes   uint64 `json:"tx_bytes"`
	RxPackets uint64 `json:"rx_packets"`
	TxPackets uint64 `json:"tx_packets"`
	RxErrors  uint64 `json:"rx_errors"`
	TxErrors  uint64 `json:"tx_errors"`
}

type TemperatureStats struct {
	SensorKey   string  `json:"sensor_key"`
	Temperature float64 `json:"temperature"`
	High        float64 `json:"high"`
	Critical    float64 `json:"critical"`
}
