package sample

const (
	// SamplerTypeProbabilistic is the type of sampler that samples traces with a certain fixed probability.
	SamplerTypeProbabilistic = "probabilistic"
)

const (
	DefaultProbability = 0.0001
	DefaultTimeWindow  = 10
)

var defaultConfig = &Config{Probability: DefaultProbability, TimeWindow: DefaultTimeWindow}

// sampler decides whether a request should be sampled or not.
type Sampler interface {
	IsSampled(index string) (bool, float64)
	GetProbability() float64
}

type Config struct {
	Probability float64
	TimeWindow  int64
}

func NewSampler(cfg *Config) (Sampler, error) {
	if cfg == nil {
		cfg = defaultConfig
	}
	return newProbabilisticSampler(cfg.Probability, cfg.TimeWindow)
}
