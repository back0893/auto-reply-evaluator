package evaluator

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"auto-reply-evaluator/internal/model"
	"auto-reply-evaluator/internal/prompt"
)

type Evaluator interface {
	Eval(reply model.AutoReply) (model.EvalResult, error)
}

type RealEvaluator struct {
	client *LLMClient
}

func NewRealEvaluator(apiKey, baseURL, model string) *RealEvaluator {
	return &RealEvaluator{
		client: NewLLMClient(apiKey, baseURL, model),
	}
}

func (r *RealEvaluator) Chat(system, user string) (string, error) {
	return r.client.Chat(system, user)
}

func (r *RealEvaluator) Eval(reply model.AutoReply) (model.EvalResult, error) {
	userPrompt := fmt.Sprintf(prompt.ScoringPrompt, reply.UserQuestion, reply.AutoReply)
	content, err := r.client.Chat(prompt.ScoringSystem, userPrompt)
	if err != nil {
		return model.EvalResult{}, fmt.Errorf("api call: %w", err)
	}

	scores, err := parseScores(content)
	if err != nil {
		return model.EvalResult{}, fmt.Errorf("parse scores: %w", err)
	}

	weightedTotal := float64(scores.AntiHallucination)*0.25 +
		float64(scores.HelpfulnessProactive)*0.10 +
		float64(scores.HelpfulnessSpecific)*0.10 +
		float64(scores.HelpfulnessBurden)*0.10 +
		float64(scores.HelpfulnessUnderstanding)*0.10 +
		float64(scores.Tone)*0.20 +
		float64(scores.Accuracy)*0.15
	weightedTotal = math.Round(weightedTotal*100) / 100

	return model.EvalResult{
		CaseID:        reply.ID,
		UserQuestion:  reply.UserQuestion,
		AutoReply:     reply.AutoReply,
		Scores:        scores,
		WeightedTotal: weightedTotal,
	}, nil
}

func parseScores(content string) (model.DimensionScores, error) {
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	}

	var s model.DimensionScores
	if err := json.Unmarshal([]byte(content), &s); err != nil {
		return model.DimensionScores{}, fmt.Errorf("parse JSON: %w, content: %s", err, content)
	}

	clamp(&s.Accuracy)
	clamp(&s.AntiHallucination)
	clamp(&s.HelpfulnessProactive)
	clamp(&s.HelpfulnessSpecific)
	clamp(&s.HelpfulnessBurden)
	clamp(&s.HelpfulnessUnderstanding)
	clamp(&s.Tone)

	return s, nil
}

func clamp(v *int) {
	if *v < 1 {
		*v = 1
	}
	if *v > 5 {
		*v = 5
	}
}

type MockEvaluator struct {
	refs map[string]model.HumanRef
}

func NewMockEvaluator(refs []model.HumanRef) *MockEvaluator {
	m := &MockEvaluator{
		refs: make(map[string]model.HumanRef),
	}
	for _, r := range refs {
		m.refs[r.ID] = r
	}
	return m
}

func (m *MockEvaluator) Eval(reply model.AutoReply) (model.EvalResult, error) {
	ref, ok := m.refs[reply.ID]
	scores := model.DimensionScores{
		Accuracy:                 3,
		AntiHallucination:        3,
		HelpfulnessProactive:     3,
		HelpfulnessSpecific:      3,
		HelpfulnessBurden:        3,
		HelpfulnessUnderstanding: 3,
		Tone:                     3,
	}

	if ok {
		notes := ref.AnnotatorNotes
		for dimName, kws := range DimensionKeywords {
			score := mockScoreFull(notes, kws.StrongNeg, kws.WeakNeg, kws.WeakPos, kws.StrongPos)
			setScore(&scores, dimName, score)
		}
		scores.AntiHallucination = mockScoreAntiHallucination(notes)
	}

	weightedTotal := float64(scores.AntiHallucination)*0.25 +
		float64(scores.HelpfulnessProactive)*0.10 +
		float64(scores.HelpfulnessSpecific)*0.10 +
		float64(scores.HelpfulnessBurden)*0.10 +
		float64(scores.HelpfulnessUnderstanding)*0.10 +
		float64(scores.Tone)*0.20 +
		float64(scores.Accuracy)*0.15
	weightedTotal = math.Round(weightedTotal*100) / 100

	return model.EvalResult{
		CaseID:        reply.ID,
		UserQuestion:  reply.UserQuestion,
		AutoReply:     reply.AutoReply,
		Scores:        scores,
		WeightedTotal: weightedTotal,
	}, nil
}

func mockScoreAntiHallucination(notes string) int {
	notesLower := strings.ToLower(notes)

	for _, kw := range AntiHallucinationKeywords {
		if strings.Contains(notesLower, strings.ToLower(kw)) {
			return 2
		}
	}

	return 4
}

func setScore(scores *model.DimensionScores, dim string, value int) {
	switch dim {
	case "Accuracy":
		scores.Accuracy = value
	case "AntiHallucination":
		scores.AntiHallucination = value
	case "HelpfulnessProactive":
		scores.HelpfulnessProactive = value
	case "HelpfulnessSpecific":
		scores.HelpfulnessSpecific = value
	case "HelpfulnessBurden":
		scores.HelpfulnessBurden = value
	case "HelpfulnessUnderstanding":
		scores.HelpfulnessUnderstanding = value
	case "Tone":
		scores.Tone = value
	}
}

func mockScore(notes string, positiveKeywords ...string) int {
	notesLower := strings.ToLower(notes)

	negativeKeywords := []string{
		"没有体现", "不足", "不够", "缺乏", "没体现",
		"推给用户", "责任推", "没有主动", "没有查", "没有帮",
		"泛泛", "让用户自己", "自己翻", "自己去看", "让用户自己去",
		"增加", "操作负担", "先给了一堆",
		"答非所问", "重复", "说了跟没说",
	}
	for _, kw := range negativeKeywords {
		if strings.Contains(notesLower, strings.ToLower(kw)) {
			return 2
		}
	}

	for _, kw := range positiveKeywords {
		if strings.Contains(notesLower, strings.ToLower(kw)) {
			return 4
		}
	}

	return 3
}

func mockScoreFull(notes string, strongNeg, weakNeg, weakPos, strongPos []string) int {
	notesLower := strings.ToLower(notes)

	for _, kw := range strongNeg {
		if strings.Contains(notesLower, strings.ToLower(kw)) {
			return 1
		}
	}

	for _, kw := range weakNeg {
		if strings.Contains(notesLower, strings.ToLower(kw)) {
			return 2
		}
	}

	for _, kw := range strongPos {
		if strings.Contains(notesLower, strings.ToLower(kw)) {
			return 5
		}
	}

	for _, kw := range weakPos {
		if strings.Contains(notesLower, strings.ToLower(kw)) {
			return 4
		}
	}

	return 3
}
