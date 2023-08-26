package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"go-gateway/app/app-svr/distribution/distribution/api"
	"go-gateway/app/app-svr/distribution/distribution/internal/preferenceproto"
	_ "go-gateway/app/app-svr/distribution/distribution/internal/preferenceproto/prelude"
	_ "go-gateway/app/app-svr/distribution/distribution/internal/storagedriver/experimentalflag"

	"git.bilibili.co/bapis/bapis-go/bilibili/metadata/parabox"

	"git.bilibili.co/bapis/bapis-go/bilibili/metadata/device"
	"github.com/golang/protobuf/jsonpb"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/urfave/cli"
	_ "go.uber.org/automaxprocs"
	"google.golang.org/grpc/metadata"
)

func main() {
	flag := &Flag{}
	app := cli.NewApp()
	app.Name = ""
	app.Usage = "distribution cmd工具"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "buvid",
			Usage:       "Replace the UserPreference buvid",
			Destination: &flag.Buvid,
		},
	}
	app.Commands = []cli.Command{
		{
			Name: "exps-bin",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "exps",
					Usage:       "decode exps",
					Destination: &flag.Exps,
				},
			},
			Action: func(ctx *cli.Context) {
				exps := &parabox.Exps{}
				bs, err := base64.RawStdEncoding.DecodeString(flag.Exps)
				if err != nil {
					panic(err)
				}
				if err := exps.Unmarshal(bs); err != nil {
					fmt.Println(err)
				}
				marshaler := &jsonpb.Marshaler{
					EmitDefaults: true,
					Indent:       "\t",
				}
				eb, _ := marshaler.MarshalToString(exps)
				fmt.Println("x-bili-exps-bin:")
				fmt.Println(eb)
			},
		},
		{
			Name: "userpref",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "buvid",
					Usage:       "Replace the UserPreference buvid",
					Destination: &flag.Buvid,
				},
			},
			Action: func(ctx *cli.Context) error {
				return run(ctx, flag)
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}

type Flag struct {
	Buvid string
	Exps  string
}

func run(_ *cli.Context, flag *Flag) error {
	// Step 2: 替换buvid
	dev := &device.Device{
		Buvid: flag.Buvid,
	}
	devMD, _ := dev.Marshal()
	ctx := metadata.AppendToOutgoingContext(context.Background(), "x-bili-device-bin", string(devMD))
	// Step 3: 请求UserPreference
	client, err := api.NewClient(nil)
	if err != nil {
		panic(err)
	}
	reply, err := client.UserPreference(ctx, &api.UserPreferenceReq{})
	if err != nil {
		panic(err)
	}
	// Step 4: 遍历Any
	metas := map[string]*preferenceproto.PreferenceMeta{}
	for _, untyped := range reply.Preference {
		meta, ok := preferenceproto.TryGetPreference(trimGoogleApis(untyped.TypeUrl))
		if !ok {
			continue
		}
		metas[untyped.TypeUrl] = meta
	}
	// Step 5: 输出信息
	for i, untyped := range reply.Preference {
		meta, ok := metas[untyped.TypeUrl]
		if !ok {
			continue
		}
		ctr := dynamic.NewMessage(meta.ProtoDesc)
		if err := ctr.Unmarshal(untyped.Value); err != nil {
			fmt.Println("解析value出错：", err)
		}
		jsonStr, err := ctr.MarshalJSONPB(&jsonpb.Marshaler{
			EmitDefaults: true,
			Indent:       "\t",
		})
		if err != nil {
			fmt.Println("序列化json出错：", err)
			continue
		}
		fmt.Println("-------------------", "config:", i+1, untyped.TypeUrl, "-------------------")
		fmt.Println(string(jsonStr))
	}
	return nil
}

func trimGoogleApis(in string) string {
	const googleApis = "type.googleapis.com/"
	return strings.TrimPrefix(in, googleApis)
}
