package report

import (
	"fmt"
	"html/template"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"auto-reply-evaluator/internal/model"
)

const htmlTemplate = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>自动回复质量评估报告</title>
<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; background: #f5f7fa; color: #333; line-height: 1.6; padding: 20px; }
.container { max-width: 1000px; margin: 0 auto; }
h1 { text-align: center; color: #1a1a2e; margin-bottom: 10px; }
h2 { color: #1a1a2e; margin: 30px 0 15px; border-bottom: 2px solid #4a90d9; padding-bottom: 8px; }
.card { background: #fff; border-radius: 8px; padding: 20px; margin-bottom: 16px; box-shadow: 0 2px 8px rgba(0,0,0,.08); }
.overall { text-align: center; font-size: 48px; font-weight: bold; color: #4a90d9; }
.overall-label { color: #888; font-size: 14px; }
table { width: 100%; border-collapse: collapse; margin: 10px 0; }
th, td { padding: 10px 12px; text-align: left; border-bottom: 1px solid #eee; }
th { background: #f0f4f8; font-weight: 600; color: #555; }
.bar { height: 20px; border-radius: 10px; background: #4a90d9; min-width: 4px; }
.bar-bg { background: #e8ecf0; border-radius: 10px; width: 100%; }
.score-good { color: #27ae60; font-weight: bold; }
.score-mid { color: #f39c12; font-weight: bold; }
.score-bad { color: #e74c3c; font-weight: bold; }
.case { margin-bottom: 20px; }
.case-header { font-weight: bold; margin-bottom: 6px; }
.case-q { background: #fff3cd; padding: 8px 12px; border-radius: 4px; margin-bottom: 4px; font-size: 14px; }
.case-a { background: #d4edda; padding: 8px 12px; border-radius: 4px; margin-bottom: 8px; font-size: 14px; }
.case-scores { font-size: 13px; color: #666; }
.tag { display: inline-block; padding: 2px 8px; border-radius: 12px; font-size: 12px; margin: 2px; }
.tag-neg { background: #fde8e8; color: #c0392b; }
.tag-pos { background: #e8f8e8; color: #27ae60; }
.tag-neu { background: #eee; color: #888; }
.mismatch { background: #fff3cd; padding: 8px 12px; border-radius: 4px; margin-bottom: 6px; font-size: 13px; }
</style>
</head>
<body>
<div class="container">
<h1>自动回复质量评估报告</h1>

<div class="card">
<div class="overall">{{printf "%.2f" .OverallScore}}<span class="overall-label"> / 5.00</span></div>
<div style="text-align:center;color:#888;">整体加权均分</div>
</div>

<h2>各维度均值</h2>
<div class="card">
<table>
<tr><th>维度</th><th>均分</th><th>分布</th></tr>
{{range .DimensionAvgs}}
<tr>
<td>{{.Name}}</td>
<td class="{{.ScoreClass}}">{{printf "%.2f" .Value}}</td>
<td><div class="bar-bg"><div class="bar" style="width:{{.BarWidth}}%"></div></div></td>
</tr>
{{end}}
</table>
</div>

<h2>各 Case 详情（按总分排序）</h2>
{{range .Cases}}
<div class="card case">
<div class="case-header">{{.Rank}}. {{.CaseID}} | 加权得分: <span class="{{.ScoreClass}}">{{printf "%.2f" .WeightedTotal}}</span></div>
<div class="case-q"><strong>Q:</strong> {{.UserQuestion}}</div>
<div class="case-a"><strong>A:</strong> {{.AutoReply}}</div>
<div class="case-scores">
反幻觉:{{.AntiHallucination}} | 有用:{{printf "%.1f" .Usefulness}} |
<details style="display:inline;margin-left:4px;">
<summary style="cursor:pointer;color:#666;font-size:12px;">子维度明细</summary>
<span style="font-size:12px;color:#888;">主动:{{.HelpfulnessProactive}} 具体:{{.HelpfulnessSpecific}} 负担:{{.HelpfulnessBurden}} 理解:{{.HelpfulnessUnderstanding}}</span>
</details>
 | 语气:{{.Tone}} | 准确:{{.Accuracy}}
{{if .SemanticSimilarity}} | 语义相似度: {{printf "%.2f" .SemanticSimilarity}} (1-5分: {{printf "%.1f" .SemanticSimilarity1to5}}){{end}}
</div>
</div>
{{end}}

<h2>最差 3 条 Case 分析</h2>
{{range .WorstCases}}
<div class="card case">
<div class="case-header">{{.CaseID}} | 加权得分: <span class="score-bad">{{printf "%.2f" .WeightedTotal}}</span></div>
<div class="case-q"><strong>Q:</strong> {{.UserQuestion}}</div>
<div class="case-a"><strong>A:</strong> {{.AutoReply}}</div>
<div class="case-scores">
反幻觉:{{.AntiHallucination}} | 有用:{{printf "%.1f" .Usefulness}} |
<details style="display:inline;margin-left:4px;">
<summary style="cursor:pointer;color:#666;font-size:12px;">子维度明细</summary>
<span style="font-size:12px;color:#888;">主动:{{.HelpfulnessProactive}} 具体:{{.HelpfulnessSpecific}} 负担:{{.HelpfulnessBurden}} 理解:{{.HelpfulnessUnderstanding}}</span>
</details>
 | 语气:{{.Tone}} | 准确:{{.Accuracy}}
</div>
<div style="margin-top:8px;font-size:13px;color:#888;">
<strong>弱点分析:</strong> {{.WeaknessAnalysis}}
</div>
</div>
{{end}}

{{if .HasSemantic}}
<h2>语义相似度验证</h2>
<div class="card">
<p>语义相似度（自动回复 vs 人工参考回复）与 LLM 评分的相关系数: <strong>{{printf "%.3f" .Correlation}}</strong></p>
<p style="font-size:13px;color:#888;">注: 语义相似度仅作为参考指标，不参与总分计算。高相似度不等于高质量。</p>
</div>
{{end}}

{{if .HasConsistency}}
<h2>一致性验证结果</h2>
<div class="card">
<p>LLM 评分与人工标注的方向一致率: <strong>{{printf "%.1f%%" .ConsistencyRate}}</strong></p>
<table>
<tr><th>维度</th><th>一致率</th></tr>
{{range .ConsistencyPerDim}}
<tr><td>{{.Name}}</td><td>{{if .HasData}}{{printf "%.1f%%" .Rate}}{{else}}无标注数据{{end}}</td></tr>
{{end}}
</table>
{{if .Mismatches}}
<p style="margin-top:10px;font-weight:bold;">不一致 Case:</p>
{{range .Mismatches}}
<div class="mismatch">
{{.CaseID}} | {{.Dimension}}: LLM={{.LLMDirection}} 人工={{.HumanDirection}} (LLM评分:{{.LLMScore}})
</div>
{{end}}
{{end}}
</div>
{{end}}

<h2>局限性分析</h2>
<div class="card">
<ul style="padding-left:20px;">
<li><strong>LLM-as-Judge 的主观性:</strong> 评分结果受 LLM 自身偏差影响，不同模型可能给出不同分数。</li>
<li><strong>Mock 模式的粗糙性:</strong> 基于关键词匹配的 Mock 评分无法捕捉语义细节，仅适合验证流水线。</li>
<li><strong>语义相似度的局限:</strong> 高相似度不等于高质量（如 case_20 自动回复和人工回复都提及"退货流程"，但自动回复是答非所问）。</li>
<li><strong>人工标注的覆盖度:</strong> annotator_notes 并非对所有维度都有判断，一致性验证只能覆盖部分维度。</li>
<li><strong>权重设定的主观性:</strong> 当前权重基于讨论共识，但不同场景可能需要调整。</li>
<li><strong>改进方向:</strong> 可使用多模型 ensemble 评分、引入更多人工标注作为 ground truth 校准、支持自定义权重配置。</li>
</ul>
</div>

</div>
</body>
</html>`

type DimAvg struct {
	Name       string
	Value      float64
	BarWidth   float64
	ScoreClass string
}

type CaseView struct {
	Rank                     int
	CaseID                   string
	UserQuestion             string
	AutoReply                string
	WeightedTotal            float64
	ScoreClass               string
	AntiHallucination        int
	Usefulness               float64
	HelpfulnessProactive     int
	HelpfulnessSpecific      int
	HelpfulnessBurden        int
	HelpfulnessUnderstanding int
	Tone                     int
	Accuracy                 int
	SemanticSimilarity       float64
	SemanticSimilarity1to5   float64
	WeaknessAnalysis         string
}

type ConsDim struct {
	Name    string
	Rate    float64
	HasData bool
}

type MismatchView struct {
	CaseID         string
	Dimension      string
	LLMDirection   string
	HumanDirection string
	LLMScore       int
}

type ReportData struct {
	OverallScore      float64
	DimensionAvgs     []DimAvg
	Cases             []CaseView
	WorstCases        []CaseView
	HasSemantic       bool
	Correlation       float64
	HasConsistency    bool
	ConsistencyRate   float64
	ConsistencyPerDim []ConsDim
	Mismatches        []MismatchView
}

func GenerateHTML(results []model.EvalResult, simScores map[string]float64, consistency *model.ConsistencyResult, outputDir string) error {
	os.MkdirAll(outputDir, 0755)

	data := buildReportData(results, simScores, consistency)

	tmpl, err := template.New("report").Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	f, err := os.Create(filepath.Join(outputDir, "report.html"))
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	return nil
}

func GenerateMarkdown(results []model.EvalResult, simScores map[string]float64, consistency *model.ConsistencyResult, outputDir string) error {
	os.MkdirAll(outputDir, 0755)

	data := buildReportData(results, simScores, consistency)

	var sb strings.Builder
	sb.WriteString("# 自动回复质量评估报告\n\n")
	sb.WriteString(fmt.Sprintf("## 整体加权均分\n\n**%.2f / 5.00**\n\n", data.OverallScore))

	sb.WriteString("## 各维度均值\n\n")
	sb.WriteString("| 维度 | 均分 |\n")
	sb.WriteString("|------|------|\n")
	for _, d := range data.DimensionAvgs {
		sb.WriteString(fmt.Sprintf("| %s | %.2f |\n", d.Name, d.Value))
	}

	sb.WriteString("\n## 各 Case 详情\n\n")
	for _, c := range data.Cases {
		sb.WriteString(fmt.Sprintf("### %d. %s (%.2f)\n\n", c.Rank, c.CaseID, c.WeightedTotal))
		sb.WriteString(fmt.Sprintf("- **Q:** %s\n", c.UserQuestion))
		sb.WriteString(fmt.Sprintf("- **A:** %s\n", truncateMD(c.AutoReply, 100)))
		sb.WriteString(fmt.Sprintf("- 反幻觉:%d | 有用:%.1f (主动:%d 具体:%d 负担:%d 理解:%d) | 语气:%d | 准确:%d\n",
			c.AntiHallucination, c.Usefulness, c.HelpfulnessProactive, c.HelpfulnessSpecific, c.HelpfulnessBurden, c.HelpfulnessUnderstanding, c.Tone, c.Accuracy))
		if c.SemanticSimilarity > 0 {
			sb.WriteString(fmt.Sprintf("- 语义相似度: %.2f (1-5分: %.1f)\n", c.SemanticSimilarity, c.SemanticSimilarity1to5))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## 最差 3 条 Case 分析\n\n")
	for _, c := range data.WorstCases {
		sb.WriteString(fmt.Sprintf("### %s (%.2f)\n\n", c.CaseID, c.WeightedTotal))
		sb.WriteString(fmt.Sprintf("- **Q:** %s\n", c.UserQuestion))
		sb.WriteString(fmt.Sprintf("- **A:** %s\n", c.AutoReply))
		sb.WriteString(fmt.Sprintf("- **弱点分析:** %s\n\n", c.WeaknessAnalysis))
	}

	if data.HasSemantic {
		sb.WriteString("## 语义相似度验证\n\n")
		sb.WriteString(fmt.Sprintf("相关系数: **%.3f**\n\n", data.Correlation))
	}

	if data.HasConsistency {
		sb.WriteString("## 一致性验证结果\n\n")
		sb.WriteString(fmt.Sprintf("方向一致率: **%.1f%%**\n\n", data.ConsistencyRate))
		sb.WriteString("| 维度 | 一致率 |\n")
		sb.WriteString("|------|--------|\n")
		for _, d := range data.ConsistencyPerDim {
			if d.HasData {
				sb.WriteString(fmt.Sprintf("| %s | %.1f%% |\n", d.Name, d.Rate))
			} else {
				sb.WriteString(fmt.Sprintf("| %s | 无标注数据 |\n", d.Name))
			}
		}
		if len(data.Mismatches) > 0 {
			sb.WriteString("\n不一致 Case:\n\n")
			for _, m := range data.Mismatches {
				sb.WriteString(fmt.Sprintf("- %s | %s: LLM=%s 人工=%s\n", m.CaseID, m.Dimension, m.LLMDirection, m.HumanDirection))
			}
		}
	}

	sb.WriteString("\n## 局限性分析\n\n")
	sb.WriteString("- **LLM-as-Judge 的主观性:** 评分结果受 LLM 自身偏差影响\n")
	sb.WriteString("- **Mock 模式的粗糙性:** 基于关键词匹配，仅适合验证流水线\n")
	sb.WriteString("- **语义相似度的局限:** 高相似度不等于高质量\n")
	sb.WriteString("- **人工标注的覆盖度:** annotator_notes 并非对所有维度都有判断\n")
	sb.WriteString("- **改进方向:** 多模型 ensemble、更多人工标注校准、自定义权重\n")

	f, err := os.Create(filepath.Join(outputDir, "report.md"))
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()
	f.WriteString(sb.String())

	return nil
}

func buildReportData(results []model.EvalResult, simScores map[string]float64, consistency *model.ConsistencyResult) ReportData {
	sorted := make([]model.EvalResult, len(results))
	copy(sorted, results)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].WeightedTotal > sorted[j].WeightedTotal
	})

	var overall float64
	for _, r := range results {
		overall += r.WeightedTotal
	}
	overall /= float64(len(results))

	dimNames := []string{"反幻觉", "有用-主动解决", "有用-具体针对", "有用-降低负担", "有用-理解诉求", "语气", "准确性"}
	dimAvgs := computeDimAvgs(results, dimNames)

	displayAvgs := []DimAvg{
		{Name: "反幻觉", Value: dimAvgs[0].Value, BarWidth: dimAvgs[0].Value * 20, ScoreClass: scoreClass(dimAvgs[0].Value)},
		{Name: "有用", Value: (dimAvgs[1].Value + dimAvgs[2].Value + dimAvgs[3].Value + dimAvgs[4].Value) / 4,
			BarWidth:   ((dimAvgs[1].Value + dimAvgs[2].Value + dimAvgs[3].Value + dimAvgs[4].Value) / 4) * 20,
			ScoreClass: scoreClass((dimAvgs[1].Value + dimAvgs[2].Value + dimAvgs[3].Value + dimAvgs[4].Value) / 4)},
		{Name: "语气", Value: dimAvgs[5].Value, BarWidth: dimAvgs[5].Value * 20, ScoreClass: scoreClass(dimAvgs[5].Value)},
		{Name: "准确性", Value: dimAvgs[6].Value, BarWidth: dimAvgs[6].Value * 20, ScoreClass: scoreClass(dimAvgs[6].Value)},
	}

	var cases []CaseView
	for i, r := range sorted {
		sim := simScores[r.CaseID]
		cases = append(cases, CaseView{
			Rank:                     i + 1,
			CaseID:                   r.CaseID,
			UserQuestion:             r.UserQuestion,
			AutoReply:                r.AutoReply,
			WeightedTotal:            r.WeightedTotal,
			ScoreClass:               scoreClass(r.WeightedTotal),
			AntiHallucination:        r.Scores.AntiHallucination,
			Usefulness:               float64(r.Scores.HelpfulnessProactive+r.Scores.HelpfulnessSpecific+r.Scores.HelpfulnessBurden+r.Scores.HelpfulnessUnderstanding) / 4,
			HelpfulnessProactive:     r.Scores.HelpfulnessProactive,
			HelpfulnessSpecific:      r.Scores.HelpfulnessSpecific,
			HelpfulnessBurden:        r.Scores.HelpfulnessBurden,
			HelpfulnessUnderstanding: r.Scores.HelpfulnessUnderstanding,
			Tone:                     r.Scores.Tone,
			Accuracy:                 r.Scores.Accuracy,
			SemanticSimilarity:       sim,
			SemanticSimilarity1to5:   simTo1to5(sim),
		})
	}

	worst := make([]CaseView, 0)
	for i := len(sorted) - 1; i >= 0 && len(worst) < 3; i-- {
		cv := cases[i]
		cv.WeaknessAnalysis = analyzeWeakness(sorted[i].Scores)
		worst = append(worst, cv)
	}

	data := ReportData{
		OverallScore:  overall,
		DimensionAvgs: displayAvgs,
		Cases:         cases,
		WorstCases:    worst,
	}

	if len(simScores) > 0 {
		var llmScores, simVals []float64
		for _, r := range results {
			if s, ok := simScores[r.CaseID]; ok {
				llmScores = append(llmScores, r.WeightedTotal)
				simVals = append(simVals, s)
			}
		}
		if len(llmScores) > 1 {
			data.HasSemantic = true
			data.Correlation = computePearson(llmScores, simVals)
		}
	}

	if consistency != nil {
		data.HasConsistency = true
		data.ConsistencyRate = consistency.OverallRate * 100

		displayConsistencyDims := []struct {
			name string
			dims []string
		}{
			{"反幻觉", []string{"反幻觉"}},
			{"有用", []string{"有用-主动解决", "有用-具体针对", "有用-降低负担", "有用-理解诉求"}},
			{"语气", []string{"语气"}},
			{"准确性", []string{"准确性"}},
		}

		for _, dcd := range displayConsistencyDims {
			var sum float64
			var totalAnnotations int
			for _, dim := range dcd.dims {
				if rate, ok := consistency.PerDimension[dim]; ok {
					sum += rate
				}
				if t, ok := consistency.PerDimensionTotals[dim]; ok {
					totalAnnotations += t
				}
			}
			rate := (sum / float64(len(dcd.dims))) * 100
			hasData := totalAnnotations > 0
			data.ConsistencyPerDim = append(data.ConsistencyPerDim, ConsDim{Name: dcd.name, Rate: rate, HasData: hasData})
		}
		for _, m := range consistency.Mismatches {
			data.Mismatches = append(data.Mismatches, MismatchView{
				CaseID:         m.CaseID,
				Dimension:      m.Dimension,
				LLMDirection:   m.LLMDirection,
				HumanDirection: m.HumanDirection,
				LLMScore:       m.LLMScore,
			})
		}
	}

	return data
}

func computeDimAvgs(results []model.EvalResult, dimNames []string) []DimAvg {
	n := float64(len(results))
	var sumAcc, sumAH, sumHP, sumHS, sumHB, sumHU, sumTone float64
	for _, r := range results {
		sumAcc += float64(r.Scores.Accuracy)
		sumAH += float64(r.Scores.AntiHallucination)
		sumHP += float64(r.Scores.HelpfulnessProactive)
		sumHS += float64(r.Scores.HelpfulnessSpecific)
		sumHB += float64(r.Scores.HelpfulnessBurden)
		sumHU += float64(r.Scores.HelpfulnessUnderstanding)
		sumTone += float64(r.Scores.Tone)
	}
	vals := []float64{sumAH / n, sumHP / n, sumHS / n, sumHB / n, sumHU / n, sumTone / n, sumAcc / n}
	var avgs []DimAvg
	for i, name := range dimNames {
		avgs = append(avgs, DimAvg{
			Name:       name,
			Value:      vals[i],
			BarWidth:   vals[i] * 20,
			ScoreClass: scoreClass(vals[i]),
		})
	}
	return avgs
}

func scoreClass(score float64) string {
	if score >= 3.5 {
		return "score-good"
	}
	if score >= 2.5 {
		return "score-mid"
	}
	return "score-bad"
}

func analyzeWeakness(s model.DimensionScores) string {
	type kv struct {
		name  string
		score int
	}
	pairs := []kv{
		{"反幻觉", s.AntiHallucination},
		{"主动解决", s.HelpfulnessProactive},
		{"具体针对", s.HelpfulnessSpecific},
		{"降低负担", s.HelpfulnessBurden},
		{"理解诉求", s.HelpfulnessUnderstanding},
		{"语气", s.Tone},
		{"准确性", s.Accuracy},
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].score < pairs[j].score })

	var low []string
	for _, p := range pairs {
		if p.score <= 2 {
			low = append(low, p.name)
		}
	}
	if len(low) == 0 {
		return "无显著弱项"
	}
	return strings.Join(low, "、") + "得分偏低"
}

func computePearson(x, y []float64) float64 {
	n := float64(len(x))
	var sx, sy, sxy, sx2, sy2 float64
	for i := range x {
		sx += x[i]
		sy += y[i]
		sxy += x[i] * y[i]
		sx2 += x[i] * x[i]
		sy2 += y[i] * y[i]
	}
	num := n*sxy - sx*sy
	den := (n*sx2 - sx*sx) * (n*sy2 - sy*sy)
	if den <= 0 {
		return 0
	}
	return num / math.Sqrt(den)
}

func simTo1to5(raw float64) float64 {
	return 1.0 + raw*4.0
}

func truncateMD(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
