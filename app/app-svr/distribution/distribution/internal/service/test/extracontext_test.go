package service

import (
	"context"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"go-common/component/metadata/device"
	pb "go-gateway/app/app-svr/distribution/distribution/api"
	"go-gateway/app/app-svr/distribution/distribution/internal/preferenceproto"
	_ "go-gateway/app/app-svr/distribution/distribution/internal/preferenceproto/prelude"

	ptypes "github.com/gogo/protobuf/types"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/stretchr/testify/assert"
)

var (
	client         pb.DistributionClient
	deviveContenxt = device.NewContext(context.Background(), device.Device{
		Buvid:   "abcd",
		FpLocal: "aaaa",
	})
)

func init() {
	var err error
	client, err = pb.NewClient(nil)
	if err != nil {
		panic(err)
	}
}

func TestSetUserPreferenceWithExtraContext(t *testing.T) {
	meta, ok := preferenceproto.TryGetPreference("bilibili.app.distribution.play.v1.SpecificPlayConfig")
	assert.True(t, ok)
	spm := dynamic.NewMessage(meta.ProtoDesc)
	err := spm.UnmarshalJSON([]byte(`{
    "enableSegmentedSection": {
        "value": false
    }
}`))
	assert.NoError(t, err)

	aid := strconv.FormatInt(rand.Int63()%500, 10)
	cid := strconv.FormatInt(rand.Int63()%500, 10)
	t.Run("SetUserPreference", func(t *testing.T) {
		any, err := ptypes.MarshalAny(spm)
		assert.NoError(t, err)
		_, err = client.SetUserPreference(deviveContenxt, &pb.SetUserPreferenceReq{
			ExtraContext: map[string]string{
				"aid": aid,
				"cid": cid,
			},
			Preference: []*ptypes.Any{
				any,
			},
		})
		assert.NoError(t, err)
	})

	t.Run("GetUserPreference", func(t *testing.T) {
		reply, err := client.GetUserPreference(deviveContenxt, &pb.GetUserPreferenceReq{
			TypeUrl: []string{"type.googleapis.com/bilibili.app.distribution.play.v1.SpecificPlayConfig"},
			ExtraContext: map[string]string{
				"aid": aid,
				"cid": cid,
			},
		})
		assert.NoError(t, err)

		dm := dynamic.NewMessage(meta.ProtoDesc)
		err = ptypes.UnmarshalAny(reply.Value[0], dm)
		assert.NoError(t, err)
		t.Logf("reply value is: %+v", dm.GetFieldByName("enableSegmentedSection"))

		value := dm.GetFieldByName("enableSegmentedSection").(*dynamic.Message).GetFieldByName("value").(bool)
		assert.False(t, value)

		lastMofidied := dm.GetFieldByName("enableSegmentedSection").(*dynamic.Message).GetFieldByName("last_modified").(int64)
		assert.InDelta(t, lastMofidied, time.Now().Unix(), 3)
	})
}
