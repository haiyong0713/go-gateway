package dao

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	accountGRPC "git.bilibili.co/bapis/bapis-go/account/service"
	activityGRPC "git.bilibili.co/bapis/bapis-go/activity/service"
	archiveGRPC "git.bilibili.co/bapis/bapis-go/archive/service"
	tagGRPC "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	"go-common/library/conf/paladin"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc/warden"
)

func NewBMClient() (client *bm.Client, cf func(), err error) {
	var (
		cfg *bm.ClientConfig
		ct  paladin.TOML
	)
	if err = paladin.Get("http.toml").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("Client").UnmarshalTOML(&cfg); err != nil {
		return
	}
	client = bm.NewClient(cfg)
	cf = func() {}

	return
}

// NewHTTPClient new a http client.
func NewHTTPClient() (client *http.Client) {
	var (
		transport *http.Transport
		dialer    *net.Dialer
	)
	dialer = &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 20 * time.Second,
	}
	transport = &http.Transport{
		DialContext:     dialer.DialContext,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client = &http.Client{
		Transport: transport,
	}
	return
}

func NewArchiveGRPC() (client archiveGRPC.ArchiveClient, cf func(), err error) {
	var (
		cfg *warden.ClientConfig
		ct  paladin.TOML
	)
	if err = paladin.Get("grpc.toml").Unmarshal(&ct); err != nil {
		return
	}
	if exists := ct.Exist("Client"); exists {
		if err = ct.Get("Client").UnmarshalTOML(&cfg); err != nil {
			return
		}
	}
	client, err = archiveGRPC.NewClient(cfg)
	cf = func() {}

	return
}

func NewTagGRPC() (client tagGRPC.TagRPCClient, cf func(), err error) {
	var (
		cfg *warden.ClientConfig
		ct  paladin.TOML
	)
	if err = paladin.Get("grpc.toml").Unmarshal(&ct); err != nil {
		return
	}
	if exists := ct.Exist("Client"); exists {
		if err = ct.Get("Client").UnmarshalTOML(&cfg); err != nil {
			return
		}
	}
	client, err = tagGRPC.NewClient(cfg)
	cf = func() {}

	return
}

func NewAccountGRPC() (client accountGRPC.AccountClient, cf func(), err error) {
	var (
		cfg *warden.ClientConfig
		ct  paladin.TOML
	)
	if err = paladin.Get("grpc.toml").Unmarshal(&ct); err != nil {
		return
	}
	if exists := ct.Exist("Client"); exists {
		if err = ct.Get("Client").UnmarshalTOML(&cfg); err != nil {
			return
		}
	}
	client, err = accountGRPC.NewClient(cfg)
	cf = func() {}

	return
}

func NewActivityGRPC() (client activityGRPC.ActivityClient, cf func(), err error) {
	var (
		cfg *warden.ClientConfig
		ct  paladin.TOML
	)
	if err = paladin.Get("grpc.toml").Unmarshal(&ct); err != nil {
		return
	}
	if exists := ct.Exist("Client"); exists {
		if err = ct.Get("Client").UnmarshalTOML(&cfg); err != nil {
			return
		}
	}
	client, err = activityGRPC.NewClient(cfg)
	cf = func() {}

	return
}
