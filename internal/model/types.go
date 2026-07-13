package model

type AutoReply struct {
	ID           string `json:"id"`
	UserQuestion string `json:"user_question"`
	AutoReply    string `json:"auto_reply"`
}

type HumanRef struct {
	ID             string `json:"id"`
	HumanReference string `json:"human_reference"`
	AnnotatorNotes string `json:"annotator_notes"`
}

type DimensionScores struct {
	Accuracy                 int `json:"accuracy"`
	AntiHallucination        int `json:"anti_hallucination"`
	HelpfulnessProactive     int `json:"helpfulness_proactive"`
	HelpfulnessSpecific      int `json:"helpfulness_specific"`
	HelpfulnessBurden        int `json:"helpfulness_burden"`
	HelpfulnessUnderstanding int `json:"helpfulness_understanding"`
	Tone                     int `json:"tone"`
}

type EvalResult struct {
	CaseID             string          `json:"case_id"`
	UserQuestion       string          `json:"user_question"`
	AutoReply          string          `json:"auto_reply"`
	Scores             DimensionScores `json:"scores"`
	WeightedTotal      float64         `json:"weighted_total"`
	SemanticSimilarity float64         `json:"semantic_similarity,omitempty"`
}

type AnnotatorJudgment struct {
	CaseID     string                  `json:"case_id"`
	Dimensions map[string]AnnotatorDim `json:"dimensions"`
}

type AnnotatorDim struct {
	Mentioned bool   `json:"mentioned"`
	Direction string `json:"direction"`
}

type ConsistencyResult struct {
	OverallRate        float64               `json:"overall_rate"`
	PerDimension       map[string]float64    `json:"per_dimension"`
	PerDimensionTotals map[string]int        `json:"per_dimension_totals"`
	Mismatches         []ConsistencyMismatch `json:"mismatches"`
}

type ConsistencyMismatch struct {
	CaseID         string `json:"case_id"`
	Dimension      string `json:"dimension"`
	LLMDirection   string `json:"llm_direction"`
	HumanDirection string `json:"human_direction"`
	LLMScore       int    `json:"llm_score"`
}
