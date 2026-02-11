//go:build linux

package signals

import "testing"

func TestParseMemInfoValue(t *testing.T) {
	tests := []struct {
		name string
		line string
		want uint64
	}{
		{
			name: "standard format",
			line: "MemTotal:       16384000 kB",
			want: 16384000,
		},
		{
			name: "with extra spaces",
			line: "MemAvailable:   8192000   kB",
			want: 8192000,
		},
		{
			name: "invalid format",
			line: "InvalidLine",
			want: 0,
		},
		{
			name: "not a number",
			line: "MemTotal: abc kB",
			want: 0,
		},
		{
			name: "empty",
			line: "",
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseMemInfoValue(tt.line)
			if got != tt.want {
				t.Errorf("parseMemInfoValue(%q) = %d, want %d", tt.line, got, tt.want)
			}
		})
	}
}
