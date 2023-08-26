package selected

const (
	DanmuBatchCids = 20
	CidType        = 2
)

// 资源配置
type ResourceMeta struct {
	Icon    string `json:"icon"`
	KeyWord string `json:"keyword"`
}

// 防污染配置
type PurifyExtra struct {
	PurifyEffective    bool `json:"purify_effective"`
	EffectivePeriod    int  `json:"effective_period"`
	EffectiveMax       int  `json:"effective_max"`
	PurifyNonEffective bool `json:"purify_non_effective"`
	NonEffectivePeriod int  `json:"non_effective_period"`
	NonEffectiveMax    int  `json:"non_effective_max"`
}
