package api

import (
	"context"
	"crypto/x509"
	"testing"

	"git.bilibili.co/bapis/bapis-go/bilibili/metadata/device"
	"git.bilibili.co/bapis/bapis-go/bilibili/metadata/restriction"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	grpcmd "google.golang.org/grpc/metadata"
)

func TestView(t *testing.T) {
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
	cli := NewViewClient(conn)
	req := &ViewReq{
		Aid: 840074675,
	}
	md := device.Device{
		MobiApp:  "ipad",
		Device:   "pad",
		Platform: "ios",
		Build:    12510,
		Buvid:    "",
	}
	md1 := device.Device{
		MobiApp:  "ipad",
		Device:   "pad",
		Platform: "ios",
		Build:    12511,
		Buvid:    "",
	}
	mb, err := md.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	mb1, err := md1.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(mb))
	mr := restriction.Restriction{}
	mrb, err := mr.Marshal()
	//ctx := grpcmd.NewIncomingContext(context.Background(), grpcmd.Pairs("authorization", "identify_v1 05f1987d3f57befae59d5fe4d2b85121"))
	ctx := grpcmd.AppendToOutgoingContext(context.Background(), "x-bili-device-bin", string(mb))
	ctx = grpcmd.AppendToOutgoingContext(ctx, "x-bili-restriction-bin", string(mrb))
	ctx = grpcmd.AppendToOutgoingContext(ctx, "x1-bilispy-color", "bnjtest")
	ct1 := grpcmd.AppendToOutgoingContext(context.Background(), "x-bili-device-bin", string(mb1))
	ct1 = grpcmd.AppendToOutgoingContext(ct1, "x-bili-restriction-bin", string(mrb))
	ct1 = grpcmd.AppendToOutgoingContext(ct1, "x1-bilispy-color", "bnjtest")
	resp1, err := cli.View(ct1, req)
	if err != nil {
		t.Fatal(err, urlStr)
	}
	t.Log(resp1)
	t.Logf("%v", ct1)
	resp, err := cli.View(ctx, req)
	if err != nil {
		t.Fatal(err, urlStr)
	}
	t.Log(resp)
	t.Logf("%v", ctx)

}
