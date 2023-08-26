package tool

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func TestInt64Append(t *testing.T) {
	assert.Equal(t, int64(123456), Int64Append(123, 456))
	assert.Equal(t, int64(-123456), Int64Append(-123, 456))
	assert.Equal(t, int64(-123456), Int64Append(-123, -456))
	assert.Equal(t, int64(-1), Int64Append(1232222222223123, 123123123213123))
	assert.Equal(t, int64(123214), Int64Append(0, 123214))
}

func BenchmarkInt64Append(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Int64Append(rand.Int63n(760096753), rand.Int63n(760096753))
		//Int64Append(1232222222223123, 123123123213123)
	}
}
