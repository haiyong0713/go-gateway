package ab

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
	var (
		v1, v2, v3, v4 *Version
		err            error
	)
	v1, err = newVersion("10.0.0.1")
	assert.Nil(t, err)

	v2, err = newVersion("10")
	assert.Nil(t, err)

	v3, err = newVersion("9.9.9")
	assert.Nil(t, err)

	_, err = newVersion("abc")
	assert.NotNil(t, err)

	v4, err = newVersion("10.0.0.1")

	assert.True(t, v1.ge(v2))
	assert.True(t, v1.ge(v3))
	assert.True(t, v1.eq(v4))
	assert.True(t, v1.ge(nil))
	assert.Equal(t, KVVersion("v1", v1).Value(), v1)
}
