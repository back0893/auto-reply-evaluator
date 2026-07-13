package validator

import (
	"testing"

	"auto-reply-evaluator/internal/evaluator"
	"auto-reply-evaluator/internal/model"
)

func TestCheckMockConsistency_100Percent(t *testing.T) {
	refs := []model.HumanRef{
		{ID: "case_01", AnnotatorNotes: "没有体现主动服务意识，把责任推给了用户。"},
		{ID: "case_02", AnnotatorNotes: "自动回复没有查具体商品，只是泛泛讲了规定，让用户自己去看。"},
		{ID: "case_03", AnnotatorNotes: "回答基本正确，主动帮用户查了物流信息，直接给出明确答案。"},
		{ID: "case_04", AnnotatorNotes: "语气不够友好，情绪安抚不够，让人感觉冷漠。"},
		{ID: "case_05", AnnotatorNotes: "先给了一堆操作步骤，增加了操作负担，没有降低用户负担。"},
	}

	mock := evaluator.NewMockEvaluator(refs)

	var results []model.EvalResult
	for _, ref := range refs {
		reply := model.AutoReply{
			ID:           ref.ID,
			UserQuestion: "test question",
			AutoReply:    "test reply",
		}
		result, err := mock.Eval(reply)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		results = append(results, result)
	}

	consistency := CheckMockConsistency(results, refs)

	if consistency.OverallRate < 1.0 {
		t.Errorf("Mock consistency should be 100%%, got %.2f%%", consistency.OverallRate*100)
		if len(consistency.Mismatches) > 0 {
			for _, m := range consistency.Mismatches {
				t.Logf("  mismatch: %s/%s LLM=%s(%d) Human=%s",
					m.CaseID, m.Dimension, m.LLMDirection, m.LLMScore, m.HumanDirection)
			}
		}
	}

	for dim, rate := range consistency.PerDimension {
		if rate < 1.0 {
			t.Errorf("dimension %s consistency should be 100%%, got %.2f%%", dim, rate*100)
		}
	}
}

func TestCheckConsistency_AntiHallucinationDefaultsToPositive(t *testing.T) {
	results := []model.EvalResult{
		{
			CaseID:        "case_01",
			WeightedTotal: 3.5,
			Scores: model.DimensionScores{
				AntiHallucination: 4,
			},
		},
	}

	judgments := []model.AnnotatorJudgment{
		{
			CaseID: "case_01",
			Dimensions: map[string]model.AnnotatorDim{
				"准确性": {Mentioned: true, Direction: "positive"},
			},
		},
	}

	cr := CheckConsistency(results, judgments)

	rate, ok := cr.PerDimension["反幻觉"]
	if !ok {
		t.Fatal("反幻觉 should be present in PerDimension")
	}
	if rate != 1.0 {
		t.Errorf("反幻觉 should be 100%% when annotator didn't mention, got %.2f%%", rate*100)
	}

	total, ok := cr.PerDimensionTotals["反幻觉"]
	if !ok || total != 1 {
		t.Errorf("反幻觉 total should be 1, got %d", total)
	}
}
