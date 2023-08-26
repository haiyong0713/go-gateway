package history

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"testing"

	"git.bilibili.co/bapis/bapis-go/bilibili/metadata/device"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	grpcmd "google.golang.org/grpc/metadata"
)

func TestCursor(t *testing.T) {
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
	cli := NewHistoryClient(conn)
	req := &CursorReq{Business: "all"}
	md := device.Device{
		AppId:    1,
		Build:    6090000,
		Buvid:    "carlos",
		MobiApp:  "android",
		Platform: "android",
		Device:   "android",
		Channel:  "master",
		Brand:    "samsung",
		Model:    "SM-G9550",
		Osver:    "8.0.0",
	}
	mb, err := md.Marshal()
	a := base64.StdEncoding.EncodeToString(mb)
	fmt.Printf(a)
	t.Log(a)
	if err != nil {
		t.Fatal(err)
	}
	ctx := grpcmd.NewIncomingContext(context.Background(), grpcmd.Pairs("authorization", "identify_v1 05f1987d3f57befae59d5fe4d2b85121"))
	ctx = grpcmd.NewIncomingContext(context.Background(), grpcmd.Pairs("x-bili-device-bin", string(mb)))
	resp, err := cli.Cursor(ctx, req)
	if err != nil {
		t.Fatal(err, urlStr)
	}
	t.Log(resp)
	t.Logf("%v", ctx)
}
