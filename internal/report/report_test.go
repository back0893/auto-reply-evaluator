package report

import (
	"math"
	"testing"

	"auto-reply-evaluator/internal/model"
)

func TestSemanticSimilarity1to5(t *testing.T) {
	tests := []struct {
		raw      float64
		expected float64
	}{
		{0.0, 1.0},
		{0.25, 2.0},
		{0.5, 3.0},
		{0.75, 4.0},
		{1.0, 5.0},
		{0.123, 1.0 + 0.123*4},
	}

	for _, tt := range tests {
		got := simTo1to5(tt.raw)
		diff := math.Abs(got - tt.expected)
		if diff > 0.01 {
			t.Errorf("simTo1to5(%f) = %f, want %f", tt.raw, got, tt.expected)
		}
	}
}

func TestBuildReportData_MergesHelpfulnessIntoFourDimensions(t *testing.T) {
	results := []model.EvalResult{
		{
			CaseID:        "case_01",
			UserQuestion:  "test Q",
			AutoReply:     "test A",
			WeightedTotal: 3.5,
			Scores: model.DimensionScores{
				Accuracy:                 4,
				AntiHallucination:        4,
				HelpfulnessProactive:     2,
				HelpfulnessSpecific:      3,
				HelpfulnessBurden:        4,
				HelpfulnessUnderstanding: 3,
				Tone:                     4,
			},
		},
	}

	data := buildReportData(results, nil, nil)

	if len(data.DimensionAvgs) != 4 {
		t.Errorf("expected 4 dimensions, got %d", len(data.DimensionAvgs))
	}

	dimNames := []string{"反幻觉", "有用", "语气", "准确性"}
	for i, name := range dimNames {
		if data.DimensionAvgs[i].Name != name {
			t.Errorf("dimension %d: expected %s, got %s", i, name, data.DimensionAvgs[i].Name)
		}
	}

	expectedUseful := (2.0 + 3.0 + 4.0 + 3.0) / 4.0
	diff := math.Abs(data.DimensionAvgs[1].Value - expectedUseful)
	if diff > 0.01 {
		t.Errorf("usefulness avg: expected %.2f, got %.2f", expectedUseful, data.DimensionAvgs[1].Value)
	}

	if len(data.Cases) != 1 {
		t.Fatalf("expected 1 case, got %d", len(data.Cases))
	}

	c := data.Cases[0]
	if c.HelpfulnessProactive != 2 {
		t.Errorf("sub-dim proactive: expected 2, got %d", c.HelpfulnessProactive)
	}
	if c.HelpfulnessSpecific != 3 {
		t.Errorf("sub-dim specific: expected 3, got %d", c.HelpfulnessSpecific)
	}
	if c.HelpfulnessBurden != 4 {
		t.Errorf("sub-dim burden: expected 4, got %d", c.HelpfulnessBurden)
	}
	if c.HelpfulnessUnderstanding != 3 {
		t.Errorf("sub-dim understanding: expected 3, got %d", c.HelpfulnessUnderstanding)
	}
}

func TestConsistencyDim_NoAnnotationShowsNoData(t *testing.T) {
	cr := model.ConsistencyResult{
		OverallRate: 0.75,
		PerDimension: map[string]float64{
			"反幻觉":     0,
			"语气":      1.0,
			"准确性":     1.0,
			"有用-主动解决": 1.0,
			"有用-具体针对": 1.0,
			"有用-降低负担": 1.0,
			"有用-理解诉求": 1.0,
		},
		PerDimensionTotals: map[string]int{
			"反幻觉":     0,
			"语气":      10,
			"准确性":     10,
			"有用-主动解决": 5,
			"有用-具体针对": 5,
			"有用-降低负担": 5,
			"有用-理解诉求": 5,
		},
	}

	data := buildReportData(nil, nil, &cr)

	if !data.HasConsistency {
		t.Error("expected HasConsistency to be true")
	}

	for _, cd := range data.ConsistencyPerDim {
		if cd.Name == "反幻觉" {
			if cd.HasData {
				t.Error("反幻觉 should have HasData=false when total is 0")
			}
		} else {
			if !cd.HasData {
				t.Errorf("%s should have HasData=true when total > 0", cd.Name)
			}
		}
	}
}
