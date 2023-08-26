package newyear2021

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserPaid(t *testing.T) {
	ctx := context.Background()
	testMid := int64(216761)
	paid, err := testDao.IsUserPaid(ctx, testMid, 33385)
	assert.Equal(t, nil, err)
	assert.Equal(t, false, paid)
}
