package sample

import (
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
)

const (
	slotLength = 2048
)

var (
	errInvalidProbability = errors.New("probability P ∈ [0, 1]")
	errInvalidTimeWindow  = errors.New("time_window P ∈ [0, +∞)")
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type probabilitySampling struct {
	probability float64
	timeWindow  int64
	slot        [slotLength]int64
}

// newSampler new probability sampler
func newProbabilisticSampler(probability float64, timeWindow int64) (Sampler, error) {
	if probability < 0 || probability > 1 {
		return nil, errors.WithStack(errInvalidProbability)
	}
	if timeWindow < 0 {
		return nil, errors.WithStack(errInvalidTimeWindow)
	}
	return &probabilitySampling{probability: probability, timeWindow: timeWindow}, nil
}

func (p *probabilitySampling) IsSampled(index string) (bool, float64) {
	if p.probability == 0 {
		return false, 0
	} else if p.probability == 1 {
		return true, 1
	}
	now := time.Now().Unix()
	idx := oneAtTimeHash(index) % slotLength
	old := atomic.LoadInt64(&p.slot[idx])
	if now-old >= p.timeWindow {
		if atomic.CompareAndSwapInt64(&p.slot[idx], old, now) {
			return true, 1
		}
	}
	return rand.Float64() < p.probability, p.probability
}

func (p *probabilitySampling) GetProbability() float64 {
	return p.probability
}

func oneAtTimeHash(s string) (hash uint32) {
	b := []byte(s)
	for i := range b {
		hash += uint32(b[i])
		//nolint:gomnd
		hash += hash << 10
		//nolint:gomnd
		hash ^= hash >> 6
	}
	//nolint:gomnd
	hash += hash << 3
	//nolint:gomnd
	hash ^= hash >> 11
	//nolint:gomnd
	hash += hash << 15
	return
}
