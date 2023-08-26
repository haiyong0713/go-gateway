package deeplink

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testBuvid1 = "XYA98DAA83588E15294C8E0E5A78090AC80F0"
	testBuvid2 = "XU90AFD961E1547E7011414B74DAD1CD7B736"
	testBuvid3 = "XZF9D2DF7A2ADFC3F9E257D9F6ED8EA633C66"
)

func TestResolveDeeplinkMetaAbIdOnline(t *testing.T) {
	assert.Equal(t, resolveDeeplinkMetaAbIdOnline(testBuvid1), "yuz_7")
	assert.Equal(t, resolveDeeplinkMetaAbIdOnline(testBuvid2), "yuz_8")
	assert.Equal(t, resolveDeeplinkMetaAbIdOnline(testBuvid3), "yuz_4")
}
