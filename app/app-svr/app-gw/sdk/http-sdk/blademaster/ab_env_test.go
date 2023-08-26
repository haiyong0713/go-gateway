package blademaster

import (
	"net/http"
	"testing"

	bm "go-common/library/net/http/blademaster"

	"github.com/stretchr/testify/assert"
)

func TestMidFromCookie(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	req.Header.Set("Cookie", `_uuid=CCF31385-B17E-AA9C-07CC-80B2617039B682784infoc; buvid3=116ADB36-E762-4D4E-A785-B266FB8EBB9B190954infoc; LIVE_BUVID=AUTO9915694131875097; sid=8w3goheg; UM_distinctid=16d6b7cbb347f8-087c06ca41910f-38607501-fa000-16d6b7cbb35118; fts=1570518232; rpdid=|(RY|lkJk|R0J'ul~uYRYJ|u; CURRENT_FNVAL=16; stardustvideo=1; laboratory=1-1; im_notify_type_2231365=0; CURRENT_QUALITY=112; bp_t_offset_2231365=360875388572500425; INTVER=1; DedeUserID=2231365; DedeUserID__ckMd5=36976f7a5cb6e4a6; SESSDATA=196b576a%2C1599277945%2Cb3e39*31; bili_jct=51c822156b21b376564cffed75653e81`)
	mid, ok := midFromRequest(req)
	assert.True(t, ok)
	assert.Equal(t, int64(2231365), mid)
}

func TestMidFromContext(t *testing.T) {
	ctx := &bm.Context{}
	ctx.Set("mid", int64(2231365))

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	req.Header.Set("Cookie", `_uuid=CCF31385-B17E-AA9C-07CC-80B2617039B682784infoc; buvid3=116ADB36-E762-4D4E-A785-B266FB8EBB9B190954infoc; LIVE_BUVID=AUTO9915694131875097; sid=8w3goheg; UM_distinctid=16d6b7cbb347f8-087c06ca41910f-38607501-fa000-16d6b7cbb35118; fts=1570518232; rpdid=|(RY|lkJk|R0J'ul~uYRYJ|u; CURRENT_FNVAL=16; stardustvideo=1; laboratory=1-1; im_notify_type_2231365=0; CURRENT_QUALITY=112; bp_t_offset_2231365=360875388572500425; INTVER=1; DedeUserID=2231366; DedeUserID__ckMd5=36976f7a5cb6e4a6; SESSDATA=196b576a%2C1599277945%2Cb3e39*31; bili_jct=51c822156b21b376564cffed75653e81`)
	ctx.Request = req

	mid, ok := midFromCtx(ctx)
	assert.True(t, ok)
	assert.Equal(t, int64(2231365), mid)
}
