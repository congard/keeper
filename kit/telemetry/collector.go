package telemetry

import (
	"context"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/sensors"

	"keeper/kit/types"
)

type Collector interface {
	Collect(ctx context.Context) (HostMetrics, error)
}

type GopsutilCollector struct{}

func NewGopsutilCollector() (*GopsutilCollector, error) {
	return &GopsutilCollector{}, nil
}

func (c *GopsutilCollector) Collect(ctx context.Context) (HostMetrics, error) {
	m := NewHostMetrics()

	if err := c.collectInfo(ctx, &m); err != nil {
		return m, err
	}
	if err := c.collectLoad(ctx, &m); err != nil {
		return m, err
	}
	if err := c.collectCPU(ctx, &m); err != nil {
		return m, err
	}
	if err := c.collectMemory(ctx, &m); err != nil {
		return m, err
	}
	if err := c.collectDisk(ctx, &m); err != nil {
		return m, err
	}
	if err := c.collectNetwork(ctx, &m); err != nil {
		return m, err
	}
	if err := c.collectTemperature(ctx, &m); err != nil {
		return m, err
	}

	return m, nil
}

func (c *GopsutilCollector) collectInfo(ctx context.Context, m *HostMetrics) error {
	info, err := host.InfoWithContext(ctx)
	if err != nil {
		return err
	}
	m.Info = *info
	return nil
}

func (c *GopsutilCollector) collectLoad(ctx context.Context, m *HostMetrics) error {
	avg, err := load.AvgWithContext(ctx)
	if err != nil {
		return err
	}
	m.Load = LoadAvg{
		One:     avg.Load1,
		Five:    avg.Load5,
		Fifteen: avg.Load15,
	}
	return nil
}

func (c *GopsutilCollector) collectCPU(ctx context.Context, m *HostMetrics) error {
	percentages, err := cpu.PercentWithContext(ctx, time.Second, false)
	if err != nil {
		return err
	}
	cores, err := cpu.CountsWithContext(ctx, true)
	if err != nil {
		return err
	}
	times, err := cpu.TimesWithContext(ctx, false)
	if err != nil {
		return err
	}

	usage := 0.0
	if len(percentages) > 0 {
		usage = percentages[0]
	}

	var user, system, idle float64
	if len(times) > 0 {
		t := times[0]
		user = t.User
		system = t.System
		idle = t.Idle
	}

	m.CPU = CPUStats{
		Usage:  types.NewPercent(usage),
		Cores:  cores,
		User:   types.NewPercent(user),
		System: types.NewPercent(system),
		Idle:   types.NewPercent(idle),
	}
	return nil
}

func (c *GopsutilCollector) collectMemory(ctx context.Context, m *HostMetrics) error {
	vm, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return err
	}
	swap, err := mem.SwapMemoryWithContext(ctx)
	if err != nil {
		return err
	}
	m.Memory = MemoryStats{
		Total:     vm.Total,
		Used:      vm.Used,
		Free:      vm.Free,
		SwapTotal: swap.Total,
		SwapUsed:  swap.Used,
	}
	return nil
}

func (c *GopsutilCollector) collectDisk(ctx context.Context, m *HostMetrics) error {
	partitions, err := disk.PartitionsWithContext(ctx, false)
	if err != nil {
		return err
	}
	disks := make([]DiskStats, 0, len(partitions))
	for _, p := range partitions {
		usage, err := disk.UsageWithContext(ctx, p.Mountpoint)
		if err != nil {
			continue
		}
		io, err := disk.IOCountersWithContext(ctx, p.Device)
		if err != nil {
			io = nil
		}
		var readBytes, writeBytes uint64
		if c, ok := io[p.Device]; ok {
			readBytes = c.ReadBytes
			writeBytes = c.WriteBytes
		}
		disks = append(disks, DiskStats{
			Mountpoint: p.Mountpoint,
			Device:     p.Device,
			Total:      usage.Total,
			Used:       usage.Used,
			Free:       usage.Free,
			ReadBytes:  readBytes,
			WriteBytes: writeBytes,
		})
	}
	m.Disk = disks
	return nil
}

func (c *GopsutilCollector) collectNetwork(ctx context.Context, m *HostMetrics) error {
	counters, err := net.IOCountersWithContext(ctx, true)
	if err != nil {
		return err
	}
	nets := make([]NetStats, 0, len(counters))
	for _, n := range counters {
		nets = append(nets, NetStats{
			Interface: n.Name,
			RxBytes:   n.BytesRecv,
			TxBytes:   n.BytesSent,
			RxPackets: n.PacketsRecv,
			TxPackets: n.PacketsSent,
			RxErrors:  n.Errin,
			TxErrors:  n.Errout,
		})
	}
	m.Network = nets
	return nil
}

func (c *GopsutilCollector) collectTemperature(ctx context.Context, m *HostMetrics) error {
	temps, err := sensors.TemperaturesWithContext(ctx)
	if err != nil {
		return err
	}
	readings := make([]TemperatureStats, 0, len(temps))
	for _, t := range temps {
		readings = append(readings, TemperatureStats{
			SensorKey:   t.SensorKey,
			Temperature: t.Temperature,
			High:        t.High,
			Critical:    t.Critical,
		})
	}
	m.Temperature = readings
	return nil
}
