package adresource

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPegasusAdAv(t *testing.T) {
	id, ok := CalcResourceID(context.TODO(), BuildPegasusAdAvScene("ios"))
	assert.True(t, ok)
	assert.Equal(t, PegasusAdAvIOSID, id)

	id, ok = CalcResourceID(context.TODO(), BuildPegasusAdAvScene("android"))
	assert.True(t, ok)
	assert.Equal(t, PegasusAdAvAndroidID, id)

	assert.Panics(t, func() {
		CalcResourceID(context.TODO(), BuildPegasusAdAvScene("qweq"))
	})

	id, ok = CalcResourceID(context.TODO(), BuildPegasusAdAvScene("qweq"), PanicOnNoScene(false))
	assert.False(t, ok)
	assert.Equal(t, EmptyResourceID, id)

	id, ok = CalcResourceID(context.TODO(), PegasusAdAvIOS)
	assert.True(t, ok)
	assert.Equal(t, PegasusAdAvIOSID, id)
}
