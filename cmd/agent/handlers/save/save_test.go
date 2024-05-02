package hadnlers

import (
	"runtime"
	"testing"
)

func TestSaveMetrics(t *testing.T) {
	type args struct {
		m *runtime.MemStats
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SaveMetrics(tt.args.m)
		})
	}
}
