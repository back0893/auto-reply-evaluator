package evaluator

type DimKeywords struct {
	StrongNeg []string
	WeakNeg   []string
	WeakPos   []string
	StrongPos []string
}

var DimensionKeywords = map[string]DimKeywords{
	"HelpfulnessProactive": {
		StrongNeg: []string{"推给用户", "责任推给", "全部推给"},
		WeakNeg:   []string{"没有体现主动", "没有主动", "没有帮"},
		WeakPos:   []string{"道歉", "表示歉意", "主动联系"},
		StrongPos: []string{"主动帮用户", "主动查", "直接帮忙"},
	},
	"HelpfulnessSpecific": {
		StrongNeg: []string{"答非所问", "完全不相关"},
		WeakNeg:   []string{"泛泛", "没有查具体", "让用户自己去看", "自己翻详情页", "让用户自己去"},
		WeakPos:   []string{"列举", "具体商品", "具体信息"},
		StrongPos: []string{"直接给出明确答案", "针对具体商品", "查了具体"},
	},
	"HelpfulnessBurden": {
		StrongNeg: []string{"增加负担", "反增负担", "更加麻烦"},
		WeakNeg:   []string{"没有降低", "操作负担", "先给了一堆", "让用户自己"},
		WeakPos:   []string{"简化", "直接给出", "一步到位"},
		StrongPos: []string{"主动帮用户操作", "帮用户去查", "替用户做"},
	},
	"HelpfulnessUnderstanding": {
		StrongNeg: []string{"答非所问", "完全没理解", "理解反了", "完全不理解"},
		WeakNeg:   []string{"没有理解", "没抓住", "没有抓住", "重复"},
		WeakPos:   []string{"基本理解", "理解了基本", "理解用户"},
		StrongPos: []string{"准确理解", "精准抓住", "一针见血"},
	},
	"Accuracy": {
		StrongNeg: []string{"错误", "不准确", "信息错误", "错了"},
		WeakNeg:   []string{"不够准确", "不够精确", "不够具体"},
		WeakPos:   []string{"基本正确", "信息本身", "大致正确"},
		StrongPos: []string{"完全正确", "非常准确", "准确无误"},
	},
	"Tone": {
		StrongNeg: []string{"语气生硬", "冷漠", "态度差"},
		WeakNeg:   []string{"语气不够", "不够友好", "情绪安抚不够", "没有安抚", "太冷漠"},
		WeakPos:   []string{"安抚", "基本友好"},
		StrongPos: []string{"共情", "非常友好", "温暖", "让人舒服"},
	},
}

var DimensionToValidatorName = map[string]string{
	"HelpfulnessProactive":     "有用-主动解决",
	"HelpfulnessSpecific":      "有用-具体针对",
	"HelpfulnessBurden":        "有用-降低负担",
	"HelpfulnessUnderstanding": "有用-理解诉求",
	"Accuracy":                 "准确性",
	"Tone":                     "语气",
}

var AntiHallucinationKeywords = []string{"瞎编", "编造", "幻觉", "虚构", "不实", "虚假"}
