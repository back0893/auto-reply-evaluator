package evaluator

import (
	"fmt"
	"sort"

	"auto-reply-evaluator/internal/model"
)

type Scorer struct {
	evaluator Evaluator
}

func NewScorer(e Evaluator) *Scorer {
	return &Scorer{evaluator: e}
}

func (s *Scorer) ScoreAll(replies []model.AutoReply) ([]model.EvalResult, error) {
	var results []model.EvalResult
	for _, r := range replies {
		result, err := s.evaluator.Eval(r)
		if err != nil {
			return nil, fmt.Errorf("evaluate %s: %w", r.ID, err)
		}
		results = append(results, result)
	}
	return results, nil
}

func PrintTextReport(results []model.EvalResult) {
	sorted := make([]model.EvalResult, len(results))
	copy(sorted, results)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].WeightedTotal > sorted[j].WeightedTotal
	})

	fmt.Println("========================================")
	fmt.Println("  自动回复质量评估报告")
	fmt.Println("========================================")

	avg := computeAverages(results)
	fmt.Printf("\n整体加权均分: %.2f / 5.00\n\n", avg.WeightedTotal)

	fmt.Println("各维度均值:")
	usefulAvg := (avg.HelpfulnessProactive + avg.HelpfulnessSpecific + avg.HelpfulnessBurden + avg.HelpfulnessUnderstanding) / 4
	fmt.Printf("  反幻觉:          %.2f\n", avg.AntiHallucination)
	fmt.Printf("  有用:            %.2f\n", usefulAvg)
	fmt.Printf("  语气:            %.2f\n", avg.Tone)
	fmt.Printf("  准确性:          %.2f\n", avg.Accuracy)
	fmt.Printf("  (有用子维度: 主动=%.2f 具体=%.2f 负担=%.2f 理解=%.2f)\n", avg.HelpfulnessProactive, avg.HelpfulnessSpecific, avg.HelpfulnessBurden, avg.HelpfulnessUnderstanding)

	fmt.Println("\n----------------------------------------")
	fmt.Println("  各 Case 详情（按总分排序）")
	fmt.Println("----------------------------------------")

	for i, r := range sorted {
		useful := float64(r.Scores.HelpfulnessProactive+r.Scores.HelpfulnessSpecific+r.Scores.HelpfulnessBurden+r.Scores.HelpfulnessUnderstanding) / 4
		fmt.Printf("\n[%d] %s | 加权得分: %.2f\n", i+1, r.CaseID, r.WeightedTotal)
		fmt.Printf("  Q: %s\n", truncateStr(r.UserQuestion, 60))
		fmt.Printf("  反幻觉:%d 有用:%.1f (主动:%d 具体:%d 负担:%d 理解:%d) 语气:%d 准确:%d\n",
			r.Scores.AntiHallucination,
			useful,
			r.Scores.HelpfulnessProactive,
			r.Scores.HelpfulnessSpecific,
			r.Scores.HelpfulnessBurden,
			r.Scores.HelpfulnessUnderstanding,
			r.Scores.Tone,
			r.Scores.Accuracy,
		)
	}

	fmt.Println("\n========================================")
	fmt.Println("  最差 3 条 Case")
	fmt.Println("========================================")
	for i := len(sorted) - 1; i >= len(sorted)-3 && i >= 0; i-- {
		r := sorted[i]
		fmt.Printf("\n[%d] %s | 加权得分: %.2f\n", i+1, r.CaseID, r.WeightedTotal)
		fmt.Printf("  Q: %s\n", r.UserQuestion)
		fmt.Printf("  A: %s\n", truncateStr(r.AutoReply, 80))
	}
}

type avgScores struct {
	Accuracy                 float64
	AntiHallucination        float64
	HelpfulnessProactive     float64
	HelpfulnessSpecific      float64
	HelpfulnessBurden        float64
	HelpfulnessUnderstanding float64
	Tone                     float64
	WeightedTotal            float64
}

func computeAverages(results []model.EvalResult) avgScores {
	n := float64(len(results))
	var a avgScores
	for _, r := range results {
		a.Accuracy += float64(r.Scores.Accuracy)
		a.AntiHallucination += float64(r.Scores.AntiHallucination)
		a.HelpfulnessProactive += float64(r.Scores.HelpfulnessProactive)
		a.HelpfulnessSpecific += float64(r.Scores.HelpfulnessSpecific)
		a.HelpfulnessBurden += float64(r.Scores.HelpfulnessBurden)
		a.HelpfulnessUnderstanding += float64(r.Scores.HelpfulnessUnderstanding)
		a.Tone += float64(r.Scores.Tone)
		a.WeightedTotal += r.WeightedTotal
	}
	a.Accuracy /= n
	a.AntiHallucination /= n
	a.HelpfulnessProactive /= n
	a.HelpfulnessSpecific /= n
	a.HelpfulnessBurden /= n
	a.HelpfulnessUnderstanding /= n
	a.Tone /= n
	a.WeightedTotal /= n
	return a
}

func truncateStr(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
