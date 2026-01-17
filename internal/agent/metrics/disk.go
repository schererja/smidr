package metrics

import (
	"context"

	byteto "github.com/schererja/smidr/internal/agent/byteTo"
	"github.com/shirou/gopsutil/v4/disk"
	"google.golang.org/protobuf/types/known/structpb"
)

func GetDiskUsage() *structpb.Struct {
	ctx := context.Background()
	diskPartitions, _ := disk.PartitionsWithContext(ctx, true)
	disks := []interface{}{}
	for _, partition := range diskPartitions {
		usageStat, err := disk.UsageWithContext(ctx, partition.Mountpoint)
		if err != nil {
			continue
		}
		disks = append(disks, map[string]interface{}{
			"device":     partition.Device,
			"mountpoint": partition.Mountpoint,
			"fstype":     partition.Fstype,
			"total_gb":   byteto.Gb(usageStat.Total),
			"used_gb":    byteto.Gb(usageStat.Used),
			"free_gb":    byteto.Gb(usageStat.Free),
			"used_pct":   usageStat.UsedPercent,
		})
	}
	s, _ := structpb.NewStruct(map[string]interface{}{
		"disks": disks,
	})
	return s
}
