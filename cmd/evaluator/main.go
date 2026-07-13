package main

import (
	"fmt"
	"os"
	"strings"

	"auto-reply-evaluator/internal/config"
	"auto-reply-evaluator/internal/evaluator"
	"auto-reply-evaluator/internal/loader"
	"auto-reply-evaluator/internal/model"
	"auto-reply-evaluator/internal/report"
	"auto-reply-evaluator/internal/validator"
)

func main() {
	cfg := config.Parse()

	autoPath := "task/task3_auto_replies.json"
	humanPath := "task/task3_human_ref.json"

	replies, err := loader.LoadAutoReplies(autoPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading auto replies: %v\n", err)
		os.Exit(1)
	}

	refs, err := loader.LoadHumanRefs(humanPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading human refs: %v\n", err)
		os.Exit(1)
	}

	var ev evaluator.Evaluator
	var llmClient *evaluator.LLMClient
	if cfg.Mock {
		ev = evaluator.NewMockEvaluator(refs)
		fmt.Println("[Mock Mode] 基于人工标注方向生成评分")
	} else {
		llmClient = evaluator.NewLLMClient(cfg.APIKey, cfg.BaseURL, cfg.Model)
		ev = evaluator.NewRealEvaluator(cfg.APIKey, cfg.BaseURL, cfg.Model)
		fmt.Printf("[Real Mode] 使用模型: %s\n", cfg.Model)
	}

	scorer := evaluator.NewScorer(ev)
	results, err := scorer.ScoreAll(replies)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scoring: %v\n", err)
		os.Exit(1)
	}

	evaluator.PrintTextReport(results)

	var simScores map[string]float64
	if !cfg.Mock && llmClient != nil {
		emb := evaluator.NewEmbedder(cfg.APIKey, cfg.BaseURL, cfg.EmbeddingModel)
		simScores, err = evaluator.ComputeSimilarityScores(replies, refs, emb)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: semantic similarity failed: %v\n", err)
		}
	} else {
		simScores = computeMockSimilarity(replies, refs)
	}

	var consistency *model.ConsistencyResult
	if !cfg.Mock && llmClient != nil {
		v := validator.NewValidator(llmClient)
		judgments, err := v.ExtractAnnotations(refs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: annotation extraction failed: %v\n", err)
		} else {
			c := validator.CheckConsistency(results, judgments)
			consistency = &c
		}
	} else {
		c := validator.CheckMockConsistency(results, refs)
		consistency = &c
	}

	if err := report.GenerateHTML(results, simScores, consistency, cfg.Output); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating HTML report: %v\n", err)
	}
	if err := report.GenerateMarkdown(results, simScores, consistency, cfg.Output); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating Markdown report: %v\n", err)
	}

	fmt.Printf("\nReports saved to %s/\n", cfg.Output)
}

func computeMockSimilarity(replies []model.AutoReply, refs []model.HumanRef) map[string]float64 {
	refMap := make(map[string]model.HumanRef)
	for _, r := range refs {
		refMap[r.ID] = r
	}
	scores := make(map[string]float64)
	for _, ar := range replies {
		ref, ok := refMap[ar.ID]
		if !ok {
			scores[ar.ID] = 0
			continue
		}
		sim := simpleSimilarity(ar.AutoReply, ref.HumanReference)
		scores[ar.ID] = sim
	}
	return scores
}

func simpleSimilarity(a, b string) float64 {
	wordsA := make(map[string]int)
	wordsB := make(map[string]int)
	for _, w := range strings.Fields(a) {
		wordsA[w]++
	}
	for _, w := range strings.Fields(b) {
		wordsB[w]++
	}
	var intersection int
	var union int
	allWords := make(map[string]bool)
	for w := range wordsA {
		allWords[w] = true
	}
	for w := range wordsB {
		allWords[w] = true
	}
	for w := range allWords {
		aCount := wordsA[w]
		bCount := wordsB[w]
		if aCount < bCount {
			intersection += aCount
		} else {
			intersection += bCount
		}
		if aCount > bCount {
			union += aCount
		} else {
			union += bCount
		}
	}
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}
