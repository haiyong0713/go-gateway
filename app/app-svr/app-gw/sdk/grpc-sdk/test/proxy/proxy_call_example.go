//nolint:biliautomaxprocs
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/net/rpc/warden"
	wardensdk "go-gateway/app/app-svr/app-gw/sdk/grpc-sdk/warden"
	wardenserversdk "go-gateway/app/app-svr/app-gw/sdk/grpc-sdk/warden/server"

	account "git.bilibili.co/bapis/bapis-go/account/service"

	"github.com/BurntSushi/toml"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

//nolint:deadcode,unused
func jsonify(in interface{}) string {
	out, _ := json.Marshal(in)
	return string(out)
}

func main() {
	tomlGenerate()

	client := warden.NewClient(nil)
	conn, err := client.Dial(context.Background(), "172.23.39.210:9000")
	if err != nil {
		panic(err)
	}
	accountService := account.NewAccountClient(conn)

	ctx := metadata.AppendToOutgoingContext(context.Background(), "color", "testv")
	var header, trailer metadata.MD
	req := &account.MidReq{Mid: 2231364}
	reply, err := accountService.Info3(ctx,
		req,
		grpc.Header(&header),   // will retrieve header
		grpc.Trailer(&trailer), // will retrieve trailer
	)
	if err != nil {
		panic(err)
	}
	fmt.Println("Reply", reply, header, trailer)

	//nolint:gosimple
	return
}

func tomlGenerate() {
	cfg := struct {
		ProxyConfig wardenserversdk.Config
	}{}
	cfg.ProxyConfig.DynService = append(cfg.ProxyConfig.DynService, &wardenserversdk.ServiceMeta{
		ServiceName: "account.service.Account",
		ClientSDKConfig: wardensdk.ClientSDKConfig{
			AppID: "account.service",
			MethodOption: []*wardensdk.MethodOption{
				//nolint:gofmt
				{
					Method: "Info3",
					BackupRetryOption: wardensdk.BackupRetryOption{
						Ratio:        100,
						BackupAction: "ecode",
						BackupECode:  -999,
					},
				},
			},
		},
	})
	tomlBuf := bytes.Buffer{}
	encoder := toml.NewEncoder(&tomlBuf)
	if err := encoder.Encode(cfg); err != nil {
		panic(err)
	}
	//nolint:gosimple
	fmt.Printf("TOML: \n%s\n", string(tomlBuf.Bytes()))
	//nolint:gosimple
	return
}
