package thread_sampler

import (
	"math"
	"sync"

	"go-gateway/app/app-svr/app-feed/interface/conf"
	"go-gateway/app/app-svr/app-feed/ng-clarify-job/sample"

	"github.com/pkg/errors"
)

const (
	maxProbability = 0.001
)

var (
	errInvalidProbability = errors.New("sample probability is invalid")
)

type ThreadSampler struct {
	sync.RWMutex
	inner sample.Sampler
	cfg   *conf.Config
}

func NewThreadSampler(cfg *conf.Config) (*ThreadSampler, error) {
	sampler := &ThreadSampler{cfg: cfg}
	if !sampler.isValidConfig() {
		return nil, errors.WithStack(errInvalidProbability)
	}
	sampler.RWMutex = sync.RWMutex{}
	s, err := sample.NewSampler(cfg.SampleConfig)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	sampler.inner = s
	return sampler, nil
}

func (s *ThreadSampler) Reload() error {
	if !s.isValidConfig() {
		return errors.WithStack(errInvalidProbability)
	}
	if floatEqual(s.GetProbability(), s.GetCfgProbability()) {
		return nil
	}
	s.Lock()
	defer s.Unlock()
	if floatEqual(s.inner.GetProbability(), s.GetCfgProbability()) {
		return nil
	}
	sampler, err := sample.NewSampler(s.cfg.SampleConfig)
	if err != nil {
		return errors.WithStack(err)
	}
	s.inner = sampler
	return nil
}

func (s *ThreadSampler) IsSampled(index string) (bool, float64) {
	s.RLock()
	defer s.RUnlock()
	return s.inner.IsSampled(index)
}

func (s *ThreadSampler) GetProbability() float64 {
	s.RLock()
	defer s.RUnlock()
	return s.inner.GetProbability()
}

func (s *ThreadSampler) GetCfgProbability() float64 {
	if s.cfg == nil || s.cfg.SampleConfig == nil {
		return 0
	}
	return s.cfg.SampleConfig.Probability
}

func (s *ThreadSampler) isValidConfig() bool {
	probability := s.GetCfgProbability()
	return probability > 0 && probability <= maxProbability
}

func floatEqual(a, b float64) bool {
	//nolint:gomnd
	return !(math.Abs(a-b) > 0.000001)
}
