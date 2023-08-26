package metadata

import (
	"encoding/base64"
	"testing"
)

func TestMetadata(t *testing.T) {
	md := Metadata{
		AccessKey: "\n",
		MobiApp:   "iphone",
		Device:    "phone",
		Build:     8721,
		Channel:   "apple",
		Buvid:     "ZF43F1FE144C207A4EBF8D0EE63322BEC34D",
		Platform:  "ios",
	}
	b, err := md.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	bs := base64.StdEncoding.EncodeToString(b)
	t.Logf("metadata: %+v", md)
	t.Logf("result: %s", bs)

	mm := Device{
		MobiApp:  "iphone",
		Device:   "phone",
		Build:    8721,
		Channel:  "apple",
		Buvid:    "ZF43F1FE144C207A4EBF8D0EE63322BEC34D",
		Platform: "ios",
	}
	bb, err := mm.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	bss := base64.StdEncoding.EncodeToString(bb)
	t.Logf("device: %+v", mm)
	t.Logf("result: %s", bss)
}
