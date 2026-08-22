package stats

import (
	"context"
	"keeper/pkg/percent"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/sensors"
)

func Collect(ctx context.Context) (HostMetrics, error) {
	var m HostMetrics

	info, err := CollectInfo(ctx)
	if err != nil {
		return m, err
	}
	m.Info = info

	loadAvg, err := CollectLoad(ctx)
	if err != nil {
		return m, err
	}
	m.Load = loadAvg

	cpuStats, err := CollectCPU(ctx)
	if err != nil {
		return m, err
	}
	m.CPU = cpuStats

	memory, err := CollectMemory(ctx)
	if err != nil {
		return m, err
	}
	m.Memory = memory

	disks, err := CollectDisk(ctx)
	if err != nil {
		return m, err
	}
	m.Disk = disks

	network, err := CollectNetwork(ctx)
	if err != nil {
		return m, err
	}
	m.Network = network

	temperatures, err := CollectTemperature(ctx)
	if err != nil {
		return m, err
	}
	m.Temperature = temperatures

	return m, nil
}

func CollectInfo(ctx context.Context) (InfoStat, error) {
	info, err := host.InfoWithContext(ctx)
	if err != nil {
		return InfoStat{}, err
	}
	return InfoStat(*info), nil
}

func CollectLoad(ctx context.Context) (LoadAvg, error) {
	avg, err := load.AvgWithContext(ctx)
	if err != nil {
		return LoadAvg{}, err
	}
	return LoadAvg{
		One:     avg.Load1,
		Five:    avg.Load5,
		Fifteen: avg.Load15,
	}, nil
}

func CollectCPU(ctx context.Context) (CPUStats, error) {
	percentages, err := cpu.PercentWithContext(ctx, time.Second, false)
	if err != nil {
		return CPUStats{}, err
	}
	cores, err := cpu.CountsWithContext(ctx, true)
	if err != nil {
		return CPUStats{}, err
	}
	times, err := cpu.TimesWithContext(ctx, false)
	if err != nil {
		return CPUStats{}, err
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

	return CPUStats{
		Usage:  percent.NewPercent(usage),
		Cores:  cores,
		User:   percent.NewPercent(user),
		System: percent.NewPercent(system),
		Idle:   percent.NewPercent(idle),
	}, nil
}

func CollectMemory(ctx context.Context) (MemoryStats, error) {
	vm, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return MemoryStats{}, err
	}
	swap, err := mem.SwapMemoryWithContext(ctx)
	if err != nil {
		return MemoryStats{}, err
	}
	return MemoryStats{
		Total:     vm.Total,
		Used:      vm.Used,
		Free:      vm.Free,
		SwapTotal: swap.Total,
		SwapUsed:  swap.Used,
	}, nil
}

func CollectDisk(ctx context.Context) ([]DiskStats, error) {
	partitions, err := disk.PartitionsWithContext(ctx, false)
	if err != nil {
		return nil, err
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
	return disks, nil
}

func CollectNetwork(ctx context.Context) ([]NetStats, error) {
	counters, err := net.IOCountersWithContext(ctx, true)
	if err != nil {
		return nil, err
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
	return nets, nil
}

func CollectTemperature(ctx context.Context) ([]TemperatureStats, error) {
	temps, err := sensors.TemperaturesWithContext(ctx)
	if err != nil {
		return nil, err
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
	return readings, nil
}
