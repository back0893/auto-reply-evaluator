package evaluator

import (
	"math"
	"testing"

	"auto-reply-evaluator/internal/model"
)

func TestCosineSimilarity_IdenticalVectors(t *testing.T) {
	v := []float64{1, 2, 3}
	sim := CosineSimilarity(v, v)
	if math.Abs(sim-1.0) > 0.0001 {
		t.Errorf("identical vectors should have similarity 1.0, got %.4f", sim)
	}
}

func TestCosineSimilarity_OrthogonalVectors(t *testing.T) {
	a := []float64{1, 0, 0}
	b := []float64{0, 1, 0}
	sim := CosineSimilarity(a, b)
	if math.Abs(sim-0.0) > 0.0001 {
		t.Errorf("orthogonal vectors should have similarity 0.0, got %.4f", sim)
	}
}

func TestCosineSimilarity_OppositeVectors(t *testing.T) {
	a := []float64{1, 2, 3}
	b := []float64{-1, -2, -3}
	sim := CosineSimilarity(a, b)
	if math.Abs(sim+1.0) > 0.0001 {
		t.Errorf("opposite vectors should have similarity -1.0, got %.4f", sim)
	}
}

func TestCosineSimilarity_ZeroVector(t *testing.T) {
	a := []float64{0, 0, 0}
	b := []float64{1, 2, 3}
	sim := CosineSimilarity(a, b)
	if sim != 0 {
		t.Errorf("zero vector should return 0, got %.4f", sim)
	}
}

func TestCosineSimilarity_DifferentLengths(t *testing.T) {
	a := []float64{1, 2}
	b := []float64{1, 2, 3}
	sim := CosineSimilarity(a, b)
	if sim != 0 {
		t.Errorf("different length vectors should return 0, got %.4f", sim)
	}
}

func TestCosineSimilarity_EmptyVectors(t *testing.T) {
	sim := CosineSimilarity([]float64{}, []float64{})
	if sim != 0 {
		t.Errorf("empty vectors should return 0, got %.4f", sim)
	}
}

func TestSimilarityToScore_Range(t *testing.T) {
	tests := []struct {
		sim      float64
		expected float64
	}{
		{0.0, 1.0},
		{0.2, 1.0},
		{0.5, 2.5},
		{0.8, 4.0},
		{1.0, 5.0},
		{1.5, 5.0},
		{-0.5, 1.0},
	}

	for _, tt := range tests {
		got := SimilarityToScore(tt.sim)
		if math.Abs(got-tt.expected) > 0.01 {
			t.Errorf("SimilarityToScore(%.1f) = %.2f, want %.2f", tt.sim, got, tt.expected)
		}
	}
}

func TestComputeCorrelation_PerfectPositive(t *testing.T) {
	x := []float64{1, 2, 3, 4, 5}
	y := []float64{2, 4, 6, 8, 10}
	r := ComputeCorrelation(x, y)
	if math.Abs(r-1.0) > 0.0001 {
		t.Errorf("perfect positive correlation should be 1.0, got %.4f", r)
	}
}

func TestComputeCorrelation_PerfectNegative(t *testing.T) {
	x := []float64{1, 2, 3, 4, 5}
	y := []float64{10, 8, 6, 4, 2}
	r := ComputeCorrelation(x, y)
	if math.Abs(r+1.0) > 0.0001 {
		t.Errorf("perfect negative correlation should be -1.0, got %.4f", r)
	}
}

func TestComputeCorrelation_DifferentLengths(t *testing.T) {
	r := ComputeCorrelation([]float64{1, 2}, []float64{1, 2, 3})
	if r != 0 {
		t.Errorf("different length arrays should return 0, got %.4f", r)
	}
}

func TestComputeCorrelation_TooFewValues(t *testing.T) {
	r := ComputeCorrelation([]float64{1}, []float64{1})
	if r != 0 {
		t.Errorf("fewer than 2 values should return 0, got %.4f", r)
	}
}

func TestComputeCorrelation_ConstantValues(t *testing.T) {
	r := ComputeCorrelation([]float64{3, 3, 3}, []float64{1, 2, 3})
	if r != 0 {
		t.Errorf("constant values (zero variance) should return 0, got %.4f", r)
	}
}

func TestMockEvaluator_ReturnsSevenDimensions(t *testing.T) {
	mock := NewMockEvaluator([]model.HumanRef{
		{
			ID:             "case_01",
			AnnotatorNotes: "用户的核心诉求是取不到快递，需要的是'帮我解决'而非'自己去想办法'。自动回复把责任推给了用户，没有体现主动服务意识。",
		},
		{
			ID:             "case_02",
			AnnotatorNotes: "自动回复没有查具体商品，只是泛泛讲了规定，需要用户自己去确认。",
		},
	})

	result, err := mock.Eval(model.AutoReply{
		ID:           "case_01",
		UserQuestion: "我的快递到了但是放错快递柜了，取不出来",
		AutoReply:    "您好，如果快递放错了快递柜，您可以尝试联系快递员重新派送",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s := result.Scores
	if s.Accuracy < 1 || s.Accuracy > 5 {
		t.Errorf("Accuracy out of range: %d", s.Accuracy)
	}
	if s.AntiHallucination < 1 || s.AntiHallucination > 5 {
		t.Errorf("AntiHallucination out of range: %d", s.AntiHallucination)
	}
	if s.HelpfulnessProactive < 1 || s.HelpfulnessProactive > 5 {
		t.Errorf("HelpfulnessProactive out of range: %d", s.HelpfulnessProactive)
	}
	if s.HelpfulnessSpecific < 1 || s.HelpfulnessSpecific > 5 {
		t.Errorf("HelpfulnessSpecific out of range: %d", s.HelpfulnessSpecific)
	}
	if s.HelpfulnessBurden < 1 || s.HelpfulnessBurden > 5 {
		t.Errorf("HelpfulnessBurden out of range: %d", s.HelpfulnessBurden)
	}
	if s.HelpfulnessUnderstanding < 1 || s.HelpfulnessUnderstanding > 5 {
		t.Errorf("HelpfulnessUnderstanding out of range: %d", s.HelpfulnessUnderstanding)
	}
	if s.Tone < 1 || s.Tone > 5 {
		t.Errorf("Tone out of range: %d", s.Tone)
	}
}

func TestMockEvaluator_NegativeDirectionMapsToLowScore(t *testing.T) {
	mock := NewMockEvaluator([]model.HumanRef{
		{
			ID:             "case_01",
			AnnotatorNotes: "没有体现主动服务意识。",
		},
	})

	result, err := mock.Eval(model.AutoReply{
		ID:           "case_01",
		UserQuestion: "test",
		AutoReply:    "test",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Scores.HelpfulnessProactive >= 3 {
		t.Errorf("expected low proactive score for negative note, got %d", result.Scores.HelpfulnessProactive)
	}
}

func TestMockEvaluator_CaseNotFoundReturnsDefaults(t *testing.T) {
	mock := NewMockEvaluator(nil)

	result, err := mock.Eval(model.AutoReply{
		ID:           "case_99",
		UserQuestion: "test",
		AutoReply:    "test",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Scores.Accuracy != 3 {
		t.Errorf("expected default Accuracy=3, got %d", result.Scores.Accuracy)
	}
}

func TestMockEvaluator_WeightedTotal(t *testing.T) {
	mock := NewMockEvaluator([]model.HumanRef{
		{
			ID:             "case_01",
			AnnotatorNotes: "没有体现主动服务意识。",
		},
	})

	result, err := mock.Eval(model.AutoReply{
		ID:           "case_01",
		UserQuestion: "test",
		AutoReply:    "test",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.WeightedTotal < 1.0 || result.WeightedTotal > 5.0 {
		t.Errorf("WeightedTotal out of range: %f", result.WeightedTotal)
	}

	expected := float64(result.Scores.AntiHallucination)*0.25 +
		float64(result.Scores.HelpfulnessProactive)*0.10 +
		float64(result.Scores.HelpfulnessSpecific)*0.10 +
		float64(result.Scores.HelpfulnessBurden)*0.10 +
		float64(result.Scores.HelpfulnessUnderstanding)*0.10 +
		float64(result.Scores.Tone)*0.20 +
		float64(result.Scores.Accuracy)*0.15

	diff := result.WeightedTotal - expected
	if diff < 0 {
		diff = -diff
	}
	if diff > 0.01 {
		t.Errorf("WeightedTotal mismatch: got %f, expected %f", result.WeightedTotal, expected)
	}
}

func TestMockScoreFull_StrongNegativeReturns1(t *testing.T) {
	score := mockScoreFull("答非所问，完全没理解用户诉求",
		[]string{"答非所问", "完全没理解"},
		[]string{"没有体现"},
		[]string{"基本正确"},
		[]string{"主动帮用户"},
	)

	if score != 1 {
		t.Errorf("strong negative should return 1, got %d", score)
	}
}

func TestMockScoreFull_WeakNegativeReturns2(t *testing.T) {
	score := mockScoreFull("没有体现主动服务意识",
		nil,
		[]string{"没有体现", "不足"},
		[]string{"基本正确"},
		[]string{"主动帮用户"},
	)

	if score != 2 {
		t.Errorf("weak negative should return 2, got %d", score)
	}
}

func TestMockScoreFull_NoMatchReturns3(t *testing.T) {
	score := mockScoreFull("回复了用户的问题，中规中矩",
		[]string{"答非所问"},
		[]string{"没有体现"},
		[]string{"基本正确"},
		[]string{"主动帮用户"},
	)

	if score != 3 {
		t.Errorf("no match should return 3, got %d", score)
	}
}

func TestMockScoreFull_WeakPositiveReturns4(t *testing.T) {
	score := mockScoreFull("信息基本正确，中规中矩",
		[]string{"答非所问"},
		[]string{"不够主动"},
		[]string{"基本正确"},
		[]string{"主动帮用户"},
	)

	if score != 4 {
		t.Errorf("weak positive should return 4, got %d", score)
	}
}

func TestMockScoreFull_StrongPositiveReturns5(t *testing.T) {
	score := mockScoreFull("主动帮用户查了物流信息，直接给出明确答案",
		[]string{"答非所问"},
		[]string{"没有体现"},
		[]string{"基本正确"},
		[]string{"主动帮用户", "直接给出明确答案"},
	)

	if score != 5 {
		t.Errorf("strong positive should return 5, got %d", score)
	}
}

func TestMockScoreFull_StrongNegativeOverridesPositive(t *testing.T) {
	score := mockScoreFull("答非所问，但是基本正确",
		[]string{"答非所问"},
		[]string{"没有体现"},
		[]string{"基本正确"},
		[]string{"主动帮用户"},
	)

	if score != 1 {
		t.Errorf("strong negative should override positive, got %d", score)
	}
}

func TestMockScoreFull_WeakNegativeOverridesPositive(t *testing.T) {
	score := mockScoreFull("没有体现主动服务，但信息基本正确",
		[]string{"答非所问"},
		[]string{"没有体现"},
		[]string{"基本正确"},
		[]string{"主动帮用户"},
	)

	if score != 2 {
		t.Errorf("weak negative should override positive, got %d (expected 2)", score)
	}
}

func TestMockEvaluator_DimensionIndependentNegativeKeywords(t *testing.T) {
	mock := NewMockEvaluator([]model.HumanRef{
		{
			ID:             "case_01",
			AnnotatorNotes: "没有体现主动服务意识，把责任推给了用户。",
		},
	})

	result, err := mock.Eval(model.AutoReply{
		ID:           "case_01",
		UserQuestion: "test",
		AutoReply:    "test",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Scores.HelpfulnessProactive >= 3 {
		t.Errorf("HelpfulnessProactive should be low, got %d", result.Scores.HelpfulnessProactive)
	}

	if result.Scores.Accuracy < 3 {
		t.Errorf("Accuracy should not be affected by proactive negative keywords, got %d", result.Scores.Accuracy)
	}

	if result.Scores.AntiHallucination < 3 {
		t.Errorf("AntiHallucination should not be affected by proactive negative keywords, got %d", result.Scores.AntiHallucination)
	}
}

func TestToneWeakPos_NoApology(t *testing.T) {
	kws := DimensionKeywords["Tone"]
	for _, kw := range kws.WeakPos {
		if kw == "道歉" {
			t.Error("\"道歉\" should not be in Tone WeakPos keywords")
		}
	}
}
