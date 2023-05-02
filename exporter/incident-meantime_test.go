package exporter

import (
	"reflect"
	"testing"
)

func Test_scanLines(t *testing.T) {
	type args struct {
		content string
	}
	tests := []struct {
		name      string
		args      args
		wantLines []string
	}{
		{
			name: "Test scanLines",
			args: args{
				content: `# Incident Log

> Use the [mean-time-to-repair.rb] script to view performance metrics

## Q2 2023 (January-March)

- **Mean Time to Repair**: 225h 11m

- **Mean Time to Resolve**: 225h 28m`,
			},
			wantLines: []string{
				"# Incident Log",
				"> Use the [mean-time-to-repair.rb] script to view performance metrics",
				"## Q2 2023 (January-March)",
				"- **Mean Time to Repair**: 225h 11m",
				"- **Mean Time to Resolve**: 225h 28m",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotLines := scanLines(tt.args.content); !reflect.DeepEqual(gotLines, tt.wantLines) {
				t.Errorf("scanLines() = %v, want %v", gotLines, tt.wantLines)
			}
		})
	}
}

func Test_parseMTTRQuarter(t *testing.T) {
	type args struct {
		lines []string
	}
	tests := []struct {
		name           string
		args           args
		wantMttrReport []map[string]float64
		wantErr        bool
	}{
		{
			name: "Test parseMTTRQuarter",
			args: args{
				lines: []string{
					"# Incident Log",
					"> Use the [mean-time-to-repair.rb] script to view performance metrics",
					"## Q2 2023 (January-March)",
					"- **Mean Time to Repair**: 225h 11m",
					"- **Mean Time to Resolve**: 225h 28m",
				},
			},
			wantMttrReport: []map[string]float64{
				{
					"incidents_mean_time_to_repair":  13511,
					"incidents_mean_time_to_resolve": 13528,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMttrReport, err := parseMTTRQuarter(tt.args.lines)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseMTTRQuarter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotMttrReport, tt.wantMttrReport) {
				t.Errorf("parseMTTRQuarter() = %v, want %v", gotMttrReport, tt.wantMttrReport)
			}
		})
	}
}
