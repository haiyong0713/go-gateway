package api

import (
	"context"
	"crypto/x509"
	"testing"

	"go-gateway/app/app-svr/app-interface/interface/api/metadata"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	grpcmd "google.golang.org/grpc/metadata"
)

func TestSVideo(t *testing.T) {
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
	cli := NewDynamicClient(conn)
	req := &SVideoReq{}
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
	resp, err := cli.SVideo(ctx, req)
	if err != nil {
		t.Fatal(err, urlStr)
	}
	t.Log(resp)
	t.Logf("%v", ctx)
}
