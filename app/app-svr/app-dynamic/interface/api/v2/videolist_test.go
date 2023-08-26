package v2

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	io "io"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"git.bilibili.co/bapis/bapis-go/bilibili/metadata/device"
)

func TestH1PB(t *testing.T) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	cli := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 10,
	}
	sh := &DynVideoReq{
		AssistBaseline: "",
		Offset:         "",
		Page:           1,
		RefreshType:    Refresh_refresh_new,
		UpdateBaseline: "",
		PlayurlParam: &PlayurlParam{
			Fnval:     16,
			Fnver:     0,
			ForceHost: 0,
			Fourk:     0,
			Qn:        32,
		},
	}
	b, err := sh.Marshal()
	if err != nil {
		t.Error(err)
		t.Failed()
	}
	// grpc body:
	//   1byte 0-uncompressed, 1-compressed using the mechanism declared by the Message-Encoding header.
	//   4byte message length
	//   bytes message data
	var head [5]byte
	binary.BigEndian.PutUint32(head[1:], uint32(len(b)))
	buff := bytes.NewBuffer(nil)
	buff.Write(head[:])
	buff.Write(b)
	req, err := http.NewRequest("POST", "https://pre-grpc.biliapi.net/bilibili.app.dynamic.v2.Dynamic/DynVideo", buff)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	req.Header = http.Header{}
	req.Header.Set("Accept", "application/grpc")
	req.Header.Set("Content-Type", "application/grpc")
	req.Header.Set("Accept-Encoding", "gzip")
	md := device.Device{
		MobiApp:  "iphone",
		Device:   "phone",
		Build:    89933,
		Channel:  "apple",
		Buvid:    "65afa745b8141ba6b42c315a0031539d",
		Platform: "ios",
	}
	bb, err := md.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	bs := base64.StdEncoding.EncodeToString(bb)
	req.Header.Set("authorization", "identify_v1 bbf90f3c2d2d07265e351ac74c0d0971")
	req.Header.Set("x-bili-device-bin", bs)
	resp, err := cli.Do(req)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Log("req body:", append(head[:], b...))
	t.Log("req header:", resp.Request.Header)
	t.Log("resp header:", resp.Header)
	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
	default:
		reader = resp.Body
	}
	defer reader.Close()
	b, err = ioutil.ReadAll(reader)
	if err != nil {
		t.Error(err, string(b))
		t.FailNow()
	}
	t.Log("resp body:", b)
	r := DynVideoReply{}
	if err = r.Unmarshal(b[5:]); err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Logf("%+v", r)
	bjson, _ := json.Marshal(&r)
	t.Logf("index json: %+s", bjson)
	t.Logf("%+v", resp.Header)
}
