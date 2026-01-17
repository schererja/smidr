package metrics

import (
	"github.com/shirou/gopsutil/v4/cpu"
	"google.golang.org/protobuf/types/known/structpb"
)

func GetCpuUsage() *structpb.Struct {
	percentages, _ := cpu.Percent(0, false)
	cpuCount, err := cpu.Counts(true)
	if err != nil {
		cpuCount = 0
	}

	s, _ := structpb.NewStruct(map[string]interface{}{
		"percent": percentages[0],
		"cores":   cpuCount,
	})
	return s
}
