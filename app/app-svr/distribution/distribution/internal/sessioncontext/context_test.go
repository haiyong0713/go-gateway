package sessioncontext

import (
	"testing"

	pb "go-gateway/app/app-svr/distribution/distribution/api"

	"github.com/stretchr/testify/assert"
)

func TestSessionContextImpl_ExtraContext(t *testing.T) {
	eMap := map[string]string{
		"aid": "123",
		"cid": "234",
	}
	getReq := &pb.GetUserPreferenceReq{
		ExtraContext: eMap,
	}
	getRes, ok := extractExtraContext(getReq)
	assert.True(t, ok)
	assert.Equal(t, getRes, eMap)
	setReq := &pb.SetUserPreferenceReq{
		ExtraContext: eMap,
	}
	setRes, ok := extractExtraContext(setReq)
	assert.True(t, ok)
	assert.Equal(t, setRes, eMap)
}
