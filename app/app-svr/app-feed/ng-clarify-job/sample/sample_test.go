package sample

import (
	"fmt"
	"testing"
)

func TestProbabilitySampling(t *testing.T) {
	sampler, _ := newProbabilisticSampler(0.0001, 1)
	t.Run("test probability", func(t *testing.T) {
		sampler.IsSampled("small,banner")
		count := 0
		for i := 0; i < 1000000; i++ {
			sampled, _ := sampler.IsSampled("small,banner")
			if sampled {
				count++
			}
		}
		fmt.Printf("count=%d\n", count)
		if count > 120 || count < 80 {
			t.Errorf("expect count between 80~120 get %d", count)
		}
	})
}

func BenchmarkProbabilitySampling(b *testing.B) {
	sampler, _ := newProbabilisticSampler(0.0001, 1)
	for i := 0; i < b.N; i++ {
		sampler.IsSampled("small,banner")
	}
}
