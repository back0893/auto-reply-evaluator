# Issues：自动回复质量评估流水线

## 依赖关系图

```
1 ──→ 2 ──→ 3 ──┐
  │              │
  ├──→ 4 ────────┤
  │              ├──→ 6 ──→ 7
  └──→ 5 ────────┘
```

切片 4 和 5 可以并行（各自只依赖 1），切片 6 汇总所有数据。

---

## Issue 1: 项目脚手架 + 数据加载

**Parent**: PRD - 自动回复质量评估流水线

**What to build**

搭建 Go 项目骨架，初始化 `go mod`，定义核心数据结构，实现 JSON 数据加载器。运行后能成功加载 `auto_replies.json` 和 `human_ref.json` 并打印数据摘要。

**Acceptance criteria**

- [ ] `go mod init auto-reply-evaluator` 完成
- [ ] `internal/model/types.go` 定义 AutoReply、HumanRef、Scores 等核心结构体
- [ ] `internal/loader/loader.go` 加载 3 个 JSON 文件，返回结构体切片
- [ ] `cmd/evaluator/main.go` 运行后打印 20 条数据加载成功 + 每条 ID 和问题摘要
- [ ] 项目目录结构符合标准 Go 模块规范

**Blocked by**: None - can start immediately

---

## Issue 2: Mock 评分流水线 + 简单文本报告

**Parent**: PRD - 自动回复质量评估流水线

**What to build**

实现 Mock 模式评分器：从 annotator_notes 提取方向判断，映射为 7 维 1-5 分。实现评分编排逻辑（遍历 20 条、加权计算总分）。输出简单文本报告（整体得分、各维度均值、按总分排序的 case 列表）。

**Acceptance criteria**

- [ ] `internal/evaluator/scorer.go` 实现 Mock 评分逻辑（基于 annotator_notes 方向映射）
- [ ] `internal/evaluator/llm.go` 定义 Evaluator 接口（Eval 方法），Mock 实现返回 7 维分数
- [ ] `internal/config/config.go` 支持 `--mock` 标志切换模式
- [ ] 终端输出：整体加权总分 + 7 个维度各自均值 + 20 条 case 按总分排序
- [ ] 每条 case 输出 7 维分数详情

**Blocked by**: Issue 1

---

## Issue 3: 真实 LLM 评分

**Parent**: PRD - 自动回复质量评估流水线

**What to build**

实现真实 LLM 评分器：调用 OpenAI 兼容接口，一次 Prompt 输出全部 7 维分数。设计评分 Prompt 模板（包含评分标准、维度定义、输出格式）。实现响应解析（JSON 提取 + 校验 1-5 范围）。

**Acceptance criteria**

- [ ] `internal/prompt/prompts.go` 定义评分 Prompt 模板（含 7 维度定义和 1-5 分标准）
- [ ] `internal/evaluator/llm.go` 实现 RealEvaluator（调用 OpenAI Chat Completions API）
- [ ] 响应 JSON 解析 + 校验（每个维度在 1-5 范围内）
- [ ] 配置支持 API Key、Base URL、Model 参数
- [ ] `--mock=false` 时使用真实 API，输出格式与 Mock 一致

**Blocked by**: Issue 2

---

## Issue 4: 语义相似度计算

**Parent**: PRD - 自动回复质量评估流水线

**What to build**

实现语义相似度计算：调用 OpenAI Embeddings API 获取自动回复向量和人工参考回复向量，计算余弦相似度。输出每条 case 的相似度分数，并统计与 LLM 评分的相关性。

**Acceptance criteria**

- [ ] `internal/evaluator/embedding.go` 实现 Embedder 接口（调用 Embeddings API）
- [ ] 实现余弦相似度计算函数
- [ ] 每条 case 输出语义相似度分（0-1，报告时映射到 1-5 便于对比）
- [ ] 输出 Pearson/Spearman 相关系数（LLM 总分 vs 相似度分）

**Blocked by**: Issue 1

---

## Issue 5: 交叉验证（一致性）

**Parent**: PRD - 自动回复质量评估流水线

**What to build**

用 LLM 一次调用将 20 条 annotator_notes 批量提取结构化判断（mentioned + direction）。实现方向一致性比较逻辑：LLM 评分 ≤2=negative, ≥4=positive, 3=neutral。统计整体一致率和各维度一致率。输出不一致 case 列表。

**Acceptance criteria**

- [ ] `internal/validator/validator.go` 实现一致性验证逻辑
- [ ] 一次 LLM 调用提取 20 条 annotator_notes 的结构化判断
- [ ] 方向一致性比较 + 统计（整体一致率、各维度一致率）
- [ ] 输出不一致 case 列表（case ID + 维度 + LLM 评分方向 + 人工方向）
- [ ] Mock 模式下一致性应为 100%（验证逻辑正确性）

**Blocked by**: Issue 2 or Issue 3

---

## Issue 6: HTML/Markdown 报告

**Parent**: PRD - 自动回复质量评估流水线

**What to build**

基于完整评估结果生成 HTML 和 Markdown 报告。HTML 包含：整体得分卡片、各维度分布表、语义相似度对比、最差 3 条 case 分析、一致性验证结果、局限性分析。Markdown 为 HTML 的纯文本版本。

**Acceptance criteria**

- [ ] `internal/report/html.go` 生成 HTML 报告（含内联 CSS，无需外部依赖）
- [ ] `internal/report/markdown.go` 生成 Markdown 报告
- [ ] 报告包含：整体得分、各维度分布、最差 3 条 case 详情及分析
- [ ] 报告包含：语义相似度验证结果
- [ ] 报告包含：一致性验证结果 + 局限性分析
- [ ] 输出到 `output/` 目录

**Blocked by**: Issue 3, Issue 4, Issue 5

---

## Issue 7: CLI 集成 + README

**Parent**: PRD - 自动回复质量评估流水线

**What to build**

完善 CLI 入口：支持 `--mock`、`--output`、`--api-key` 等参数。编写 README 说明指标定义、评估方法、局限性、AI 工具使用情况。

**Acceptance criteria**

- [ ] `cmd/evaluator/main.go` 完整的 CLI 参数解析
- [ ] `--mock` 切换 Mock/Real 模式
- [ ] `--output` 指定输出目录（默认 `output/`）
- [ ] 运行 `go run . --mock` 一键输出完整报告
- [ ] README 包含：指标定义、评估方法、局限性、AI 工具使用情况

**Blocked by**: Issue 6

---

## 汇总

| Issue | 标题 | 阻塞 | 可并行 |
|-------|------|------|--------|
| 1 | 项目脚手架 + 数据加载 | 无 | - |
| 2 | Mock 评分流水线 + 简单文本报告 | 1 | - |
| 3 | 真实 LLM 评分 | 2 | - |
| 4 | 语义相似度计算 | 1 | 5 |
| 5 | 交叉验证（一致性） | 2 或 3 | 4 |
| 6 | HTML/Markdown 报告 | 3, 4, 5 | - |
| 7 | CLI 集成 + README | 6 | - |

---

## 修正轮次：Mock 评分与报告对齐

### 依赖关系图

```
1 ──┐
    ├──→ 3
2 ──┘
```

Issue 1 和 2 可并行开发，Issue 3 汇总验证。

---

## Issue 8: Mock 评分体系重构（1-5 范围 + 维度独立关键词 + 一致性 100%）

**Parent**: PRD 修正 - 自动回复质量评估流水线

**What to build**

重构 Mock 评分逻辑，使其与 PRD 定义完全对齐。当前 `mockScore` 仅返回 2/3/4 三个值，且使用全局负面关键词列表导致维度间分数污染。需要改为：返回 1-5 完整范围，每个维度使用独立的正面/负面关键词列表，`CheckMockConsistency` 与 `mockScore` 共享同一套关键词规则确保 Mock 一致率 100%。

**Acceptance criteria**

- [ ] `mockScore` 返回 1-5 全范围（强负面→1, 弱负面→2, 未提及→3, 弱正面→4, 强正面→5）
- [ ] 每个维度使用独立的负面关键词列表（非全局共享），防止维度间污染
- [ ] 反幻觉维度：未提及幻觉关键词→4 分，提及→2 分
- [ ] `CheckMockConsistency` 与 `mockScore` 使用同一套关键词规则，Mock 一致率 100%
- [ ] 各维度分数应有区分度，不再出现大量 case 全部维度都是 2 分的情况

**Blocked by**: None - can start immediately

---

## Issue 9: 语义相似度 1-5 映射

**Parent**: PRD 修正 - 自动回复质量评估流水线

**What to build**

在报告中，将语义相似度的原始 0-1 余弦相似度值线性映射到 1-5 分制（mappedScore = 1 + similarity × 4），便于与其他 7 个维度的 LLM 评分横向对比。同时保留原始值展示。

**Acceptance criteria**

- [ ] 报告中语义相似度同时展示原始值（0-1）和映射值（1-5）
- [ ] 映射公式：`mappedScore = 1 + similarity × 4`
- [ ] HTML 和 Markdown 报告均展示映射后的分数
- [ ] 语义相似度映射值不参与加权总分计算（仅作参考）

**Blocked by**: None - can start immediately（可与 Issue 8 并行）

---

## Issue 10: 回归测试 + 端到端验证

**Parent**: PRD 修正 - 自动回复质量评估流水线

**What to build**

更新现有测试用例适配新评分范围，新增 Mock 一致性 100% 测试，端到端运行 Mock 模式验证所有修复生效。

**Acceptance criteria**

- [ ] 现有测试适配 1-5 评分范围，全部通过
- [ ] 新增 `TestMockConsistency_100Percent` 测试验证 Mock 一致率 100%
- [ ] 新增 `TestMockScore_ReturnsFullRange` 测试验证评分覆盖 1-5
- [ ] `go run . --mock` 端到端运行，输出评分有明显区分度
- [ ] 报告输出到 `output/` 目录，HTML 和 Markdown 均正确生成

**Blocked by**: Issue 8, Issue 9

---

## 汇总

| Issue | 标题 | 阻塞 | 可并行 |
|-------|------|------|--------|
| 8 | Mock 评分体系重构 | 无 | 9 |
| 9 | 语义相似度 1-5 映射 | 无 | 8 |
| 10 | 回归测试 + 端到端验证 | 8, 9 | - |

---

## 修正轮次：报告展示维度合并

### 依赖关系

```
11 (独立，无阻塞)
```

---

## Issue 11: 报告展示维度合并（7→4）

**Parent**: PRD 修正 - 自动回复质量评估流水线

**What to build**

内部评分仍为 7 维度，但最终报告展示时将 4 个有用性子维度合并为 1 个"有用"均值，形成 4 维度展示：准确性、反幻觉、有用、语气。每条 case 详情中嵌套展示 4 子维度明细。一致性验证报告同步合并展示。

**Acceptance criteria**

- [ ] 报告主题展示 4 维度（准确性、反幻觉、有用、语气），有用 = 4 子维度算术均值
- [ ] 每条 case 详情中嵌套展示 4 有用性子维度明细（主动解决、具体针对、降低负担、理解诉求）
- [ ] 终端文本报告同步调整为 4 维度展示 + 子维度明细
- [ ] 一致性验证报告展示有用性一致率 = 4 子维度一致率均值
- [ ] HTML 和 Markdown 报告均正确展示
- [ ] `go run . --mock` 端到端输出正确

**Blocked by**: None - can start immediately