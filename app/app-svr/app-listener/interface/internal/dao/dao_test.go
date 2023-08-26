package dao

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"testing"

	"go-common/library/conf/paladin.v2"
	"go-gateway/app/app-svr/app-listener/interface/internal/model"
)

var (
	_dao     *dao
	_daoOnce = sync.Once{}
)

func daoInstance() *dao {
	_daoOnce.Do(func() {
		err := paladin.Init()
		if err != nil {
			panic(err)
		}
		d, _, err := newDao()
		if err != nil {
			panic(err)
		}
		_dao = d
	})
	return _dao
}

func TestFavFolderList(t *testing.T) {
	ins := daoInstance()
	fs, err := ins.FavFolderList(context.TODO(), FavFolderListOpt{FavTypes: []int32{11}, Mid: 14139334})
	if err != nil {
		t.Error(err)
	}
	for _, f := range fs {
		fmt.Printf("%+v\n", *f.Folder)
	}
}

func TestFavFolderDetail(t *testing.T) {
	ins := daoInstance()
	dts, err := ins.FavFoldersDetail(context.TODO(), FavFolderDetailsOpt{
		Mid: 14135892,
		Folders: []model.FavFolderMeta{
			{
				Typ: 2,
				Mid: 14135892,
				Fid: 7090051,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for hs, dt := range dts {
		fmt.Printf("%v -> %+v", hs, dt)
	}
}

func TestBmComposeURI(t *testing.T) {
	data := []struct {
		Host     string
		URI      string
		Expected string
	}{
		{
			"http://api.bilibili.co", "/foo/bar",
			"http://api.bilibili.co/foo/bar",
		},
		{
			"http://api.bilibili.co/good/", "foo/bar",
			"http://api.bilibili.co/good/foo/bar",
		},
		{
			"http://api.bilibili.co?ok=1", "foo/bar",
			"http://api.bilibili.co/foo/bar",
		},
		{
			"http://api.bilibili.co/good", "/foo/bar?ok=1",
			"http://api.bilibili.co/foo/bar?ok=1",
		},
	}

	for _, d := range data {
		u, _ := url.Parse(d.Host)
		c := &bmClient{hostURL: u}
		if out := c.composeURI(d.URI); out != d.Expected {
			t.Errorf("expect %q but got %q", d.Expected, out)
		}
	}
}
