package api

import (
	"testing"

	"github.com/tomasino/writing-coach/internal/domain"
)

func TestRequiresCalibrationApprovalOverride(t *testing.T) {
	tests := []struct {
		name  string
		run   domain.CalibrationRun
		notes string
		want  bool
	}{
		{
			name:  "adequate no override",
			run:   domain.CalibrationRun{DataAdequate: true},
			notes: "",
			want:  false,
		},
		{
			name:  "inadequate requires override",
			run:   domain.CalibrationRun{DataAdequate: false},
			notes: "validated",
			want:  true,
		},
		{
			name: "objective track failures require override",
			run: domain.CalibrationRun{
				DataAdequate: true,
				TrackLearnings: []domain.CalibrationTrackLearning{
					{TreeSlug: "academic-essay-track", Issues: []string{"objective_eval_policy_failed"}},
				},
			},
			notes: "validated",
			want:  true,
		},
		{
			name: "override note accepted",
			run: domain.CalibrationRun{
				DataAdequate: false,
			},
			notes: "override: accepting low-sample exploratory run while waiting for more submissions",
			want:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			needsOverride, _ := requiresCalibrationApprovalOverride(tc.run, "approved", tc.notes)
			if needsOverride != tc.want {
				t.Fatalf("needsOverride = %v, want %v", needsOverride, tc.want)
			}
		})
	}
}
