package validator

import (
	"encoding/json"
	"fmt"
	"strings"

	ev "auto-reply-evaluator/internal/evaluator"
	"auto-reply-evaluator/internal/model"
	"auto-reply-evaluator/internal/prompt"
)

type Validator struct {
	evaluator interface {
		Chat(system, user string) (string, error)
	}
}

func NewValidator(evaluator interface {
	Chat(system, user string) (string, error)
}) *Validator {
	return &Validator{evaluator: evaluator}
}

func (v *Validator) ExtractAnnotations(notes []model.HumanRef) ([]model.AnnotatorJudgment, error) {
	var texts []string
	for _, n := range notes {
		texts = append(texts, fmt.Sprintf("case_id: %s\nnotes: %s", n.ID, n.AnnotatorNotes))
	}

	userPrompt := fmt.Sprintf(prompt.AnnotationExtractionPrompt, strings.Join(texts, "\n\n---\n\n"))
	resp, err := v.evaluator.Chat("你是一个标注分析专家。请严格按照JSON格式输出。", userPrompt)
	if err != nil {
		return nil, fmt.Errorf("extract annotations: %w", err)
	}

	resp = cleanJSON(resp)
	var judgments []model.AnnotatorJudgment
	if err := json.Unmarshal([]byte(resp), &judgments); err != nil {
		return nil, fmt.Errorf("parse judgments: %w\nresponse: %s", err, resp)
	}

	return judgments, nil
}

func CheckMockConsistency(results []model.EvalResult, refs []model.HumanRef) model.ConsistencyResult {
	annotations := make([]model.AnnotatorJudgment, 0, len(refs))
	for _, ref := range refs {
		judgment := model.AnnotatorJudgment{
			CaseID:     ref.ID,
			Dimensions: make(map[string]model.AnnotatorDim),
		}
		notes := ref.AnnotatorNotes
		notesLower := strings.ToLower(notes)

		for dimName, kws := range ev.DimensionKeywords {
			validatorName := ev.DimensionToValidatorName[dimName]

			dim := model.AnnotatorDim{Mentioned: false, Direction: "neutral"}
			matched := false

			for _, kw := range kws.StrongNeg {
				if strings.Contains(notesLower, strings.ToLower(kw)) {
					dim.Mentioned = true
					dim.Direction = "negative"
					matched = true
					break
				}
			}
			if !matched {
				for _, kw := range kws.WeakNeg {
					if strings.Contains(notesLower, strings.ToLower(kw)) {
						dim.Mentioned = true
						dim.Direction = "negative"
						matched = true
						break
					}
				}
			}
			if !matched {
				for _, kw := range kws.StrongPos {
					if strings.Contains(notesLower, strings.ToLower(kw)) {
						dim.Mentioned = true
						dim.Direction = "positive"
						matched = true
						break
					}
				}
			}
			if !matched {
				for _, kw := range kws.WeakPos {
					if strings.Contains(notesLower, strings.ToLower(kw)) {
						dim.Mentioned = true
						dim.Direction = "positive"
						matched = true
						break
					}
				}
			}

			if dim.Mentioned {
				judgment.Dimensions[validatorName] = dim
			}
		}

		antiHallucMentioned := false
		for _, kw := range ev.AntiHallucinationKeywords {
			if strings.Contains(notesLower, strings.ToLower(kw)) {
				antiHallucMentioned = true
				break
			}
		}
		judgment.Dimensions["反幻觉"] = model.AnnotatorDim{
			Mentioned: true,
			Direction: "positive",
		}
		if antiHallucMentioned {
			judgment.Dimensions["反幻觉"] = model.AnnotatorDim{
				Mentioned: true,
				Direction: "negative",
			}
		}

		annotations = append(annotations, judgment)
	}

	return CheckConsistency(results, annotations)
}

func CheckConsistency(results []model.EvalResult, judgments []model.AnnotatorJudgment) model.ConsistencyResult {
	judgmentMap := make(map[string]model.AnnotatorJudgment)
	for _, j := range judgments {
		judgmentMap[j.CaseID] = j
	}

	dimensions := []string{
		"准确性", "反幻觉",
		"有用-主动解决", "有用-具体针对", "有用-降低负担", "有用-理解诉求",
		"语气",
	}

	dimTotals := make(map[string]int)
	dimMatch := make(map[string]int)
	var mismatches []model.ConsistencyMismatch
	total := 0

	for _, r := range results {
		judgment, ok := judgmentMap[r.CaseID]
		if !ok {
			continue
		}

		if _, ok := judgment.Dimensions["反幻觉"]; !ok {
			judgment.Dimensions["反幻觉"] = model.AnnotatorDim{
				Mentioned: true,
				Direction: "positive",
			}
		}

		for _, dim := range dimensions {
			annDim, hasAnnotation := judgment.Dimensions[dim]
			if !hasAnnotation || !annDim.Mentioned {
				continue
			}

			llmDir := scoreToDirection(getScore(r.Scores, dim))
			dimTotals[dim]++
			total++

			if llmDir == annDim.Direction || llmDir == "neutral" || annDim.Direction == "neutral" {
				dimMatch[dim]++
			} else {
				mismatches = append(mismatches, model.ConsistencyMismatch{
					CaseID:         r.CaseID,
					Dimension:      dim,
					LLMDirection:   llmDir,
					HumanDirection: annDim.Direction,
					LLMScore:       getScore(r.Scores, dim),
				})
			}
		}
	}

	overallRate := float64(0)
	if total > 0 {
		matched := total - len(mismatches)
		overallRate = float64(matched) / float64(total)
	}

	perDim := make(map[string]float64)
	for _, dim := range dimensions {
		if dimTotals[dim] > 0 {
			perDim[dim] = float64(dimMatch[dim]) / float64(dimTotals[dim])
		}
	}

	return model.ConsistencyResult{
		OverallRate:        overallRate,
		PerDimension:       perDim,
		PerDimensionTotals: dimTotals,
		Mismatches:         mismatches,
	}
}

func scoreToDirection(score int) string {
	if score >= 4 {
		return "positive"
	}
	if score <= 2 {
		return "negative"
	}
	return "neutral"
}

func getScore(s model.DimensionScores, dim string) int {
	switch dim {
	case "准确性":
		return s.Accuracy
	case "反幻觉":
		return s.AntiHallucination
	case "有用-主动解决":
		return s.HelpfulnessProactive
	case "有用-具体针对":
		return s.HelpfulnessSpecific
	case "有用-降低负担":
		return s.HelpfulnessBurden
	case "有用-理解诉求":
		return s.HelpfulnessUnderstanding
	case "语气":
		return s.Tone
	}
	return 3
}

func cleanJSON(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSuffix(s, "```")
		s = strings.TrimSpace(s)
	}
	return s
}
