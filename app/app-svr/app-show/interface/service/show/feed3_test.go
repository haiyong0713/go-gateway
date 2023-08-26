package show

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSVideoURI(t *testing.T) {
	u1 := buildSVideoURI(1, 23456, 0)
	assert.Equal(t, "bilibili://inline/play_list/1/23456", u1)

	u2 := buildSVideoURI(1, 23456, 9988)
	assert.Equal(t, "bilibili://inline/play_list/1/23456?focus_aid=9988", u2)
}
