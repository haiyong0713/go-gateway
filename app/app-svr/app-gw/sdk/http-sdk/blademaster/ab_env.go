package blademaster

import (
	"net/http"
	"strconv"

	"go-common/component/metadata/device"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/blademaster/ab"
)

// TODO: static variable for now
func init() {
	ab.Registry.RegisterEnv(
		ab.KVInt("mid", 0),
		ab.KVString("sid", ""),
		ab.KVString("buvid3", ""),
		ab.KVInt("build", 0),
		ab.KVString("buvid", ""),
		ab.KVString("channel", ""),
		ab.KVString("device", ""),
		ab.KVString("rawplatform", ""),
		ab.KVString("rawmobiapp", ""),
		ab.KVString("model", ""),
		ab.KVString("brand", ""),
		ab.KVString("osver", ""),
		ab.KVString("useragent", ""),
		ab.KVInt("plat", 0),
		ab.KVBool("isandroid", false),
		ab.KVBool("isIOS", false),
		ab.KVBool("isweb", false),
		ab.KVBool("isoverseas", false),
		ab.KVString("mobiapp", ""),
		ab.KVString("mobiappbulechange", ""),
	)
}

func midFromRequest(req *http.Request) (int64, bool) {
	midCookie, err := req.Cookie("DedeUserID")
	if err != nil {
		return 0, false
	}
	mid, err := strconv.ParseInt(midCookie.Value, 10, 64)
	if err != nil {
		return 0, false
	}
	return mid, true
}

func midFromCtx(ctx *bm.Context) (int64, bool) {
	v, ok := ctx.Get("mid")
	if ok {
		return v.(int64), true
	}
	return midFromRequest(ctx.Request)
}

func ABEnvExtract(ctx *bm.Context) []ab.KV {
	kv := []ab.KV{}
	mid, _ := midFromCtx(ctx)
	kv = append(kv, ab.KVInt("mid", mid))

	device, ok := device.FromContext(ctx)
	if ok {
		kv = append(kv,
			ab.KVString("sid", device.Sid),
			ab.KVString("buvid3", device.Buvid3),
			ab.KVInt("build", device.Build),
			ab.KVString("buvid", device.Buvid),
			ab.KVString("channel", device.Channel),
			ab.KVString("device", device.Device),
			ab.KVString("rawplatform", device.RawPlatform),
			ab.KVString("rawmobiapp", device.RawMobiApp),
			ab.KVString("model", device.Model),
			ab.KVString("brand", device.Brand),
			ab.KVString("osver", device.Osver),
			ab.KVString("useragent", device.UserAgent),
			ab.KVInt("plat", int64(device.Plat())),
			ab.KVBool("isandroid", device.IsAndroid()),
			ab.KVBool("isIOS", device.IsIOS()),
			ab.KVBool("isweb", device.IsWeb()),
			ab.KVBool("isoverseas", device.IsOverseas()),
			ab.KVString("mobiapp", device.MobiApp()),
			ab.KVString("mobiappbulechange", device.MobiAPPBuleChange()),
		)
	}

	return kv
}

type ABEnv struct{}

func (ABEnv) ServeHTTP(ctx *bm.Context) {
	kv := ABEnvExtract(ctx)
	t := ab.New(kv...)
	ctx.Context = ab.NewContext(ctx.Context, t)
}
