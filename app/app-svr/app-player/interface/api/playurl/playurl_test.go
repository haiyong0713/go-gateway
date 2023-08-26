package api

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-interface/interface/api/metadata"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	grpcmd "google.golang.org/grpc/metadata"
)

func TestGRPC(t *testing.T) {
	pool, err := x509.SystemCertPool()
	if err != nil {
		t.Fatal(err)
	}
	urlStr := "uat-grpc.biliapi.net:443"
	conn, err := grpc.Dial(urlStr,
		grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(pool, "")),
		//		grpc.WithCompressor(grpc.NewGZIPCompressor()),
	)
	if err != nil {
		t.Fatalf("grpc.Dial(%s) err: %v", urlStr, err)
	}
	defer conn.Close()
	cli := NewPlayURLClient(conn)
	req := &PlayURLReq{Aid: 880003594, Cid: 10167904, Fnval: 16}
	md := metadata.Metadata{
		AccessKey: "\n",
		MobiApp:   "android",
		Device:    "phone",
		Build:     8721,
		Channel:   "apple",
		Buvid:     "ZF43F1FE144C207A4EBF8D0EE63322BEC34D",
		Platform:  "android",
	}
	mb, err := md.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	ctx := grpcmd.AppendToOutgoingContext(context.Background(), "x-bili-metadata-bin", string(mb))
	ctx = grpcmd.AppendToOutgoingContext(ctx, "X-BACKEND-BILI-REAL-IP", "222.73.196.18")
	resp, err := cli.PlayURL(ctx, req)
	if err != nil {
		t.Fatal(err, urlStr)
	}
	t.Log(resp)
	t.Logf("%v", ctx)
}

func TestProject(t *testing.T) {
	pool, err := x509.SystemCertPool()
	if err != nil {
		t.Fatal(err)
	}
	urlStr := "uat-grpc.biliapi.net:443"
	conn, err := grpc.Dial(urlStr,
		grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(pool, "")),
		//		grpc.WithCompressor(grpc.NewGZIPCompressor()),
	)
	if err != nil {
		t.Fatalf("grpc.Dial(%s) err: %v", urlStr, err)
	}
	defer conn.Close()
	cli := NewPlayURLClient(conn)
	req := &ProjectReq{Aid: 10099579, Cid: 10113485, Fnval: 16, Protocol: 1, DeviceType: 1}
	md := metadata.Metadata{
		AccessKey: "\n",
		MobiApp:   "android",
		Device:    "phone",
		Build:     8721,
		Channel:   "apple",
		Buvid:     "ZF43F1FE144C207A4EBF8D0EE63322BEC34D",
		Platform:  "android",
	}
	mb, err := md.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	ctx := grpcmd.AppendToOutgoingContext(context.Background(), "x-bili-metadata-bin", string(mb))
	ctx = grpcmd.AppendToOutgoingContext(ctx, "X-BACKEND-BILI-REAL-IP", "222.73.196.18")
	resp, err := cli.Project(ctx, req)
	if err != nil {
		t.Fatal(err, urlStr)
	}
	t.Log(resp)
	t.Logf("%v", ctx)
}

func TestH1PB(t *testing.T) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	cli := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 10,
	}
	sh := &PlayURLReq{Aid: 1, Cid: 2}
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
	req, err := http.NewRequest("POST", "https://app.bilibili.com/bilibili.app.playurl.v1.PlayURL/PlayURL", buff)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	req.Header = http.Header{}
	req.Header.Set("Accept", "application/grpc")
	req.Header.Set("Content-Type", "application/grpc")
	req.Header.Set("Accept-Encoding", "gzip")
	md := metadata.Metadata{
		AccessKey: "\n",
		MobiApp:   "iphone",
		Device:    "phone",
		Build:     8721,
		Channel:   "apple",
		Buvid:     "ZF43F1FE144C207A4EBF8D0EE63322BEC34D",
		Platform:  "ios",
	}
	bb, err := md.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	bs := base64.StdEncoding.EncodeToString(bb)
	req.Header.Set("x-bili-metadata-bin", bs)
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
	r := PlayURLReply{}
	if err = r.Unmarshal(b[5:]); err != nil {
		t.Error(err)
		t.FailNow()
	}
	t.Logf("%+v", r)
	t.Logf("%+v", resp.Header)
}

func TestPlayConf(t *testing.T) {
	pool, err := x509.SystemCertPool()
	if err != nil {
		t.Fatal(err)
	}
	urlStr := "uat-grpc.biliapi.net:443"
	conn, err := grpc.Dial(urlStr,
		grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(pool, "")),
		//		grpc.WithCompressor(grpc.NewGZIPCompressor()),
	)
	if err != nil {
		t.Fatalf("grpc.Dial(%s) err: %v", urlStr, err)
	}
	defer conn.Close()
	cli := NewPlayURLClient(conn)
	req := &PlayConfReq{}
	md := metadata.Metadata{
		AccessKey: "\n",
		MobiApp:   "android",
		Device:    "phone",
		Build:     8721,
		Channel:   "apple",
		Buvid:     "ZF43F1FE144C207A4EBF8D0EE63322BEC34D",
		Platform:  "android",
	}
	mb, err := md.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	ctx := grpcmd.AppendToOutgoingContext(context.Background(), "x-bili-metadata-bin", string(mb))
	ctx = grpcmd.AppendToOutgoingContext(ctx, "X-BACKEND-BILI-REAL-IP", "222.73.196.18")
	resp, err := cli.PlayConf(ctx, req)
	if err != nil {
		t.Fatal(err, urlStr)
	}
	t.Log(resp)
	t.Logf("%v", ctx)
}

func TestPlayConfEdit(t *testing.T) {
	pool, err := x509.SystemCertPool()
	if err != nil {
		t.Fatal(err)
	}
	urlStr := "uat-grpc.biliapi.net:443"
	conn, err := grpc.Dial(urlStr,
		grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(pool, "")),
		//		grpc.WithCompressor(grpc.NewGZIPCompressor()),
	)
	if err != nil {
		t.Fatalf("grpc.Dial(%s) err: %v", urlStr, err)
	}
	defer conn.Close()
	cli := NewPlayURLClient(conn)
	var playconf []*PlayConfState
	playconf = append(playconf, &PlayConfState{ConfType: ConfType_BACKGROUNDPLAY, Show: true})
	req := &PlayConfEditReq{PlayConf: playconf}
	md := metadata.Metadata{
		AccessKey: "\n",
		MobiApp:   "android",
		Device:    "phone",
		Build:     8721,
		Channel:   "apple",
		Buvid:     "ZF43F1FE144C207A4EBF8D0EE63322BEC34D",
		Platform:  "android",
	}
	mb, err := md.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	ctx := grpcmd.AppendToOutgoingContext(context.Background(), "x-bili-metadata-bin", string(mb))
	ctx = grpcmd.AppendToOutgoingContext(ctx, "X-BACKEND-BILI-REAL-IP", "222.73.196.18")
	resp, err := cli.PlayConfEdit(ctx, req)
	if err != nil {
		t.Fatal(err, urlStr)
	}
	t.Log(resp)
	t.Logf("%v", ctx)
}

func TestPlayView(t *testing.T) {
	pool, err := x509.SystemCertPool()
	if err != nil {
		t.Fatal(err)
	}
	urlStr := "uat-grpc.biliapi.net:443"
	conn, err := grpc.Dial(urlStr,
		grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(pool, "")),
		//		grpc.WithCompressor(grpc.NewGZIPCompressor()),
	)
	if err != nil {
		t.Fatalf("grpc.Dial(%s) err: %v", urlStr, err)
	}
	defer conn.Close()
	cli := NewPlayURLClient(conn)
	req := &PlayViewReq{Aid: 10318888, Cid: 10211630, Fnval: 16, ForceHost: 0, Fourk: false, Spmid: ""}
	md := metadata.Metadata{
		AccessKey: "b8842eb1d0c4ac28549368cac6b9d111",
		MobiApp:   "iphone",
		Device:    "phone",
		Build:     10080,
		Channel:   "apple",
		Buvid:     "ZF43F1FE144C207A4EBF8D0EE63322BEC34D",
		Platform:  "ios",
	}
	mb, err := md.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	ctx := grpcmd.AppendToOutgoingContext(context.Background(), "x-bili-metadata-bin", string(mb))
	ctx = grpcmd.AppendToOutgoingContext(ctx, "X-BACKEND-BILI-REAL-IP", "222.73.196.18")
	resp, err := cli.PlayView(ctx, req)
	if err != nil {
		t.Fatal(err, urlStr)
	}
	t.Log(resp)
	t.Logf("%v", ctx)
}
