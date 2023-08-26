package api

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
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
	sh := &SuggestionResult3Req{Keyword: "test"}
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
	req, err := http.NewRequest("POST", "https://app.bilibili.com/bilibili.app.interface.v1.Search/Suggest3", buff)
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
		Build:    8721,
		Channel:  "apple",
		Buvid:    "ZF43F1FE144C207A4EBF8D0EE63322BEC34D",
		Platform: "ios",
	}
	bb, err := md.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	bs := base64.StdEncoding.EncodeToString(bb)
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
	r := SuggestionResult3Reply{}
	if err = r.Unmarshal(b[5:]); err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Logf("%+v", r)
	t.Logf("%+v", resp.Header)
}
