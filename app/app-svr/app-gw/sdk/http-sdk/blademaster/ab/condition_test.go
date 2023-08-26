package ab

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCondition(t *testing.T) {
	Registry.RegisterEnv(
		KVInt("mid", 0),
		KVString("sid", ""),
		KVString("buvid3", ""),
		KVInt("build", 0),
		KVString("buvid", ""),
		KVString("channel", ""),
		KVString("device", ""),
		KVString("rawplatform", ""),
		KVString("rawmobiapp", ""),
		KVString("model", ""),
		KVString("brand", ""),
		KVString("osver", ""),
		KVString("useragent", ""),
		KVInt("plat", 0),
		KVBool("isandroid", false),
		KVBool("isIOS", false),
		KVBool("isweb", false),
		KVBool("isoverseas", false),
		KVString("mobiapp", ""),
		KVString("mobiappbulechange", ""),
	)

	abt := New(KVString("buvid", "3133"), KVInt("mid", 1234), KVBool("isweb", true))
	cond, err := ParseConditionWithError("buvid in (\"333\") || isweb==true")
	assert.NoError(t, err)
	assert.True(t, cond.Matches(abt))

	cond, err = ParseConditionWithError("buvid in [\"333\"] || isweb==true")
	assert.Error(t, err)

	cond, err = ParseConditionWithError("buvid in (333,456) || isweb==true")
	assert.NoError(t, err)

	cond, err = ParseConditionWithError("buvid in (333','456') || isweb==true")
	assert.Error(t, err)

	cond, err = ParseConditionWithError("")
	assert.NoError(t, err)
	assert.NotNil(t, cond)
	assert.True(t, cond.Matches(abt))
}
