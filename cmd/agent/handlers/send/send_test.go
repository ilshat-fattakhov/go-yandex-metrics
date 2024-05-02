package handlers

import "testing"

func TestSendMetrics(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SendMetrics()
		})
	}
}
