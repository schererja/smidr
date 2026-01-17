package metrics

import "google.golang.org/protobuf/types/known/structpb"

func GetNetworkLatency(latencyMs int64) *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]interface{}{
		"latency":  latencyMs,
		"endpoint": "control-plane",
	})
	return s
}
