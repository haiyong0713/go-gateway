package warden

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"go-common/library/ecode"
	"go-common/library/net/rpc/warden"
	rootsdk "go-gateway/app/app-svr/app-gw/sdk"
	sdk "go-gateway/app/app-svr/app-gw/sdk/grpc-sdk"
	"go-gateway/app/app-svr/app-gw/sdk/grpc-sdk/request"

	vipInforpc "git.bilibili.co/bapis/bapis-go/vip/service/vipinfo"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

const (
	appID       = "vipinfo.service"
	backupAppID = "vipinfo.bck.service"
)

func TestInit(t *testing.T) {
	csc := &ClientSDKConfig{
		AppID: appID,
	}
	err := csc.Init()
	assert.NoError(t, err)
}

func TestInterceptorNew(t *testing.T) {
	csc := ClientSDKConfig{
		AppID: appID,
		SDKConfig: sdk.Config{
			Debug: true,
		},
	}
	clientInterceptor := New(csc)
	assert.NotNil(t, clientInterceptor.cfg)
}

func TestReload(t *testing.T) {
	csc := ClientSDKConfig{
		AppID: appID,
		SDKConfig: sdk.Config{
			Debug: true,
		},
	}
	clientInterceptor := New(csc)
	assert.Equal(t, clientInterceptor.cfg.AppID, appID)
	recsc := ClientSDKConfig{
		AppID: backupAppID,
	}
	clientInterceptor.Reload(recsc)
	assert.Equal(t, clientInterceptor.cfg.AppID, backupAppID)
}

var vipBackupPlaceholder = `{"res":{"type":1,"status":0,"due_date":1581696000001,"vip_pay_type":0}}`
var retryBackupSendCount = 0
var directlyBackupSendCount = 0

func NewClient(backupAction string, cfg *warden.ClientConfig, opts ...grpc.DialOption) (vipInforpc.VipInfoClient, error) {
	client := warden.NewClient(cfg, opts...)
	option := BackupRetryOption{
		Ratio:        100,
		BackupAction: backupAction,
	}
	switch backupAction {
	case "directly_backup":
		option.BackupTarget = "discovery://default/" + backupAppID
	case "retry_backup":
		option.BackupTarget = "discovery://default/" + backupAppID
	case "ecode":
		option.BackupECode = 888
	case "placeholder":
		option.BackupPlaceholder = vipBackupPlaceholder
	}
	config := sdk.Config{
		Debug:      true,
		MaxRetries: rootsdk.Int64(1),
	}
	csc := ClientSDKConfig{
		AppID:     appID,
		SDKConfig: config,
		MethodOption: []*MethodOption{
			{
				Method:            "Info",
				BackupRetryOption: option,
			},
		},
	}
	clientSDK := New(csc)
	if option.BackupAction == "retry_backup" {
		clientSDK.client.Handlers.Send.SetFrontNamed(request.NamedHandler{
			Name: "testSend",
			Fn: func(r *request.Request) {
				retryBackupSendCount = r.RetryCount
				if r.RetryCount < 1 {
					r.Error = errors.New("mock send error")
					return
				}
			},
		})
	}
	if option.BackupAction == "directly_backup" {
		clientSDK.client.Handlers.Send.SetFrontNamed(request.NamedHandler{
			Name: "testSend",
			Fn: func(r *request.Request) {
				directlyBackupSendCount = r.RetryCount
			},
		})
	}
	client.Use(clientSDK.UnaryClientInterceptor())
	conn, err := client.Dial(context.Background(), "discovery://default/"+appID) //grpc.ClientConn
	if err != nil {
		return nil, err
	}
	return vipInforpc.NewVipInfoClient(conn), nil
}

func TestPlaceHolderVipClient(t *testing.T) {
	action := "placeholder"
	vipClient, err := NewClient(action, nil)
	assert.NoError(t, err)
	reply, err := vipClient.Info(context.TODO(), &vipInforpc.InfoReq{Mid: 2231365})
	assert.NoError(t, err)
	r, err := json.Marshal(reply)
	assert.NoError(t, err)
	assert.Equal(t, string(r), vipBackupPlaceholder)
}

func TestEcodeVipClient(t *testing.T) {
	action := "ecode"
	vipClient, err := NewClient(action, nil)
	assert.NoError(t, err)
	_, err = vipClient.Info(context.TODO(), &vipInforpc.InfoReq{Mid: 2231365})
	assert.Equal(t, ecode.Int(888), err)
}

func TestDirectlyBackupVipClient(t *testing.T) {
	action := "directly_backup"
	vipClient, err := NewClient(action, nil)
	assert.NoError(t, err)
	tcases := []struct {
		mid           int64
		exceptMessage string
	}{
		{
			mid:           27515254,
			exceptMessage: `{"res":{"type":2,"status":1,"due_date":1679155200000,"vip_pay_type":1},"control":{}}`,
		},
		{
			mid:           27515257,
			exceptMessage: `{"res":{"type":2,"status":1,"due_date":1912089600000,"vip_pay_type":0},"control":{}}`,
		},
	}
	for _, tcase := range tcases {
		reply, err := vipClient.Info(context.TODO(), &vipInforpc.InfoReq{Mid: tcase.mid})
		assert.NoError(t, err)
		r, err := json.Marshal(reply)
		assert.NoError(t, err)
		assert.Equal(t, string(r), tcase.exceptMessage)
		assert.Equal(t, directlyBackupSendCount, 0)
	}
}

func TestRetryBackupVipClient(t *testing.T) {
	action := "retry_backup"
	vipClient, err := NewClient(action, nil)
	assert.NoError(t, err)
	tcases := []struct {
		mid           int64
		exceptMessage string
	}{
		{
			mid:           27515254,
			exceptMessage: `{"res":{"type":2,"status":1,"due_date":1679155200000,"vip_pay_type":1},"control":{}}`,
		},
		{
			mid:           27515257,
			exceptMessage: `{"res":{"type":2,"status":1,"due_date":1912089600000,"vip_pay_type":0},"control":{}}`,
		},
	}
	for _, tcase := range tcases {
		reply, err := vipClient.Info(context.TODO(), &vipInforpc.InfoReq{Mid: tcase.mid})
		assert.NoError(t, err)
		r, err := json.Marshal(reply)
		assert.NoError(t, err)
		assert.Equal(t, string(r), tcase.exceptMessage)
		assert.Equal(t, retryBackupSendCount, 1)
	}
}
