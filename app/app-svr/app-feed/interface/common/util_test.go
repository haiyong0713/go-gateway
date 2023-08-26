package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInt32SliceToInt64Slice(t *testing.T) {
	t1 := []int32{1, 2, 3}
	t2 := Int32SliceToInt64Slice(t1)
	assert.Equal(t, len(t1), len(t2))
	assert.Equal(t, int64(t1[0]), t2[0])
	assert.Equal(t, int64(t1[1]), t2[1])
	assert.Equal(t, int64(t1[2]), t2[2])

}
