package evaluator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"

	"auto-reply-evaluator/internal/model"
)

type Embedder struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

func NewEmbedder(apiKey, baseURL, model string) *Embedder {
	return &Embedder{
		apiKey:  apiKey,
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		client:  &http.Client{},
	}
}

type embeddingRequest struct {
	Model string               `json:"model"`
	Input []embeddingInputItem `json:"input"`
}
type embeddingInputItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type embeddingResponse struct {
	Data struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
}

func (e *Embedder) Embed(text string) ([]float64, error) {
	reqBody := embeddingRequest{
		Model: e.model,
		Input: []embeddingInputItem{
			{
				Type: "text",
				Text: text,
			},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := e.baseURL + "/embeddings/multimodal"
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.apiKey)

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("api call: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("api error %d: %s", resp.StatusCode, string(respBody))
	}

	var embResp embeddingResponse
	if err := json.Unmarshal(respBody, &embResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if len(embResp.Data.Embedding) == 0 {
		return nil, fmt.Errorf("no embedding data")
	}

	return embResp.Data.Embedding, nil
}

func CosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

func ComputeSimilarityScores(autoReplies []model.AutoReply, humanRefs []model.HumanRef, embedder *Embedder) (map[string]float64, error) {
	refMap := make(map[string]model.HumanRef)
	for _, r := range humanRefs {
		refMap[r.ID] = r
	}

	scores := make(map[string]float64)
	for _, ar := range autoReplies {
		ref, ok := refMap[ar.ID]
		if !ok {
			scores[ar.ID] = 0
			continue
		}

		autoVec, err := embedder.Embed(ar.AutoReply)
		if err != nil {
			return nil, fmt.Errorf("embed auto reply %s: %w", ar.ID, err)
		}

		refVec, err := embedder.Embed(ref.HumanReference)
		if err != nil {
			return nil, fmt.Errorf("embed human ref %s: %w", ar.ID, err)
		}

		sim := CosineSimilarity(autoVec, refVec)
		scores[ar.ID] = math.Round(sim*10000) / 10000
	}

	return scores, nil
}

func SimilarityToScore(sim float64) float64 {
	score := sim * 5
	if score > 5 {
		score = 5
	}
	if score < 1 {
		score = 1
	}
	return math.Round(score*100) / 100
}

func ComputeCorrelation(llmScores, simScores []float64) float64 {
	if len(llmScores) != len(simScores) || len(llmScores) < 2 {
		return 0
	}

	n := float64(len(llmScores))
	var sumX, sumY, sumXY, sumX2, sumY2 float64
	for i := range llmScores {
		sumX += llmScores[i]
		sumY += simScores[i]
		sumXY += llmScores[i] * simScores[i]
		sumX2 += llmScores[i] * llmScores[i]
		sumY2 += simScores[i] * simScores[i]
	}

	num := n*sumXY - sumX*sumY
	den := math.Sqrt((n*sumX2 - sumX*sumX) * (n*sumY2 - sumY*sumY))

	if den == 0 {
		return 0
	}

	return math.Round(num/den*10000) / 10000
}
