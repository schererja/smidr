package metrics

import (
	"github.com/shirou/gopsutil/v4/mem"
	"google.golang.org/protobuf/types/known/structpb"
)

type MemoryMetrics struct {
	TotalMB      uint64
	UsedMB       uint64
	FreeMB       uint64
	UsagePercent float64
}

func GetMemoryMetrics() (*MemoryMetrics, error) {
	// Placeholder for actual memory metrics retrieval logic
	memMetrics, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	return &MemoryMetrics{
		TotalMB:      bToMb(memMetrics.Total),
		UsedMB:       bToMb(memMetrics.Used),
		FreeMB:       bToMb(memMetrics.Free),
		UsagePercent: memMetrics.UsedPercent / 100.0,
	}, nil
}
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
func GetMemoryUsage() *structpb.Struct {
	v, _ := mem.VirtualMemory()
	s, _ := structpb.NewStruct(map[string]interface{}{
		"total": bToMb(v.Total),
		"used":  bToMb(v.Used),
		"free":  bToMb(v.Free),
	})
	return s
}
