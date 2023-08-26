package dao

import (
	"context"
	"flag"
	"fmt"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-dynamic/interface/conf"
	"go-gateway/app/app-svr/app-dynamic/interface/service/draw"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDaoGetUserLatestFollowTopK(t *testing.T) {
	flag.Set("conf", "../../cmd/app-dynamic-test.toml")
	flag.Parse()
	if err := conf.Init(); err != nil {
		log.Error("conf.Init() error(%v)", err)
		panic(err)
	} // init log

	d = New(conf.Conf)
	s = draw.New(conf.Conf)
	Convey("GetUserLatestFollowTopK", t, func() {
		var (
			ctx    = context.Background()
			uid    = uint64(28271978)
			k      = int(10)
			userIP = ""
		)
		Convey("When everything goes positive", func() {
			users, err := d.GetUserLatestFollowTopK(ctx, uid, k, userIP)
			Convey("Then err should be nil.users should not be nil.", func() {
				So(err, ShouldBeNil)
				So(users, ShouldNotBeNil)
			})
		})
	})
}

func TestDaoGetUserLatestAtUsers(t *testing.T) {
	Convey("GetUserLatestAtUsers", t, func() {
		var (
			ctx    = context.Background()
			uid    = uint64(88895133)
			userIP = ""
		)
		Convey("When everything goes positive", func() {
			users, err := d.GetUserLatestAtUsers(ctx, uid, userIP)
			Convey("Then err should be nil.users should not be nil.", func() {
				So(err, ShouldBeNil)
				So(users, ShouldNotBeNil)
			})
		})
	})
}

func TestDaoGetUserSearchItems(t *testing.T) {
	Convey("GetUserSearchItems", t, func() {
		var (
			ctx  = context.Background()
			mids = []int64{}
		)
		Convey("When everything goes positive", func() {
			users, err := d.GetUserSearchItems(ctx, mids)
			Convey("Then err should be nil.users should not be nil.", func() {
				So(err, ShouldBeNil)
				So(users, ShouldNotBeNil)
			})
		})
	})
}

func TestDaoSearchUser(t *testing.T) {
	Convey("SearchUser", t, func() {
		var (
			ctx      = context.Background()
			uid      = uint64(88895133)
			word     = "bili"
			page     = int(1)
			pageSize = int(20)
		)
		Convey("When everything goes positive", func() {
			users, hasMore, err := d.SearchUser(ctx, uid, word, page, pageSize)
			Convey("Then err should be nil.users should not be nil.", func() {
				So(err, ShouldBeNil)
				So(users, ShouldNotBeNil)
				So(hasMore, ShouldBeTrue)
			})
		})
	})
}

func TestCloseChan(t *testing.T) {
	ch := make(chan int)
	go func() {
		defer func() {
			fmt.Println("1 exit")
		}()

	}()
	time.Sleep(2 * time.Second)
	close(ch)
	time.Sleep(5 * time.Second)
}

func TestHttpCancel(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(2500*time.Millisecond))
	fmt.Println(all(ctx, cancel))
	time.Sleep(10 * time.Second)
}

type Ret struct {
	One   string
	Two   string
	Three string
}

func all(ctx context.Context, cancel func()) (ret *Ret, err error) {
	type Return struct {
		Msg string `json:"msg"`
	}

	ret = new(Ret)
	var (
		errCh   = make(chan error)
		oneCh   = make(chan string, 1)
		twoCh   = make(chan string, 1)
		threeCh = make(chan string, 1)

		isOneRecved   = false
		isTwoRecved   = false
		isThreeRecved = false
	)
	go func() {
		defer func() {
			fmt.Println("one exit")
		}()
		var ret Return
		err := d.client.Get(ctx, "http://127.0.0.1:8080/one", "", nil, &ret)
		if err != nil {
			errCh <- err
		}
		oneCh <- ret.Msg
	}()
	go func() {
		defer func() {
			fmt.Println("two exit")
		}()
		var ret Return
		err := d.client.Get(ctx, "http://127.0.0.1:8080/two", "", nil, &ret)
		if err != nil {
			errCh <- err
		}
		twoCh <- ret.Msg
	}()
	go func() {
		defer func() {
			fmt.Println("three exit")
		}()
		var ret Return
		err := d.client.Get(ctx, "http://127.0.0.1:8080/three", "", nil, &ret)
		if err != nil {
			errCh <- err
		}
		threeCh <- ret.Msg
	}()
	defer func() {
		fmt.Println(isOneRecved, isTwoRecved, isThreeRecved)
		fmt.Println("exit")
		cancel()
	}()
	for {
		select {
		case err = <-errCh:
			fmt.Println(err)
		case msg := <-oneCh:
			fmt.Printf("one recv %s\n", msg)
			ret.One = msg
			isOneRecved = true
		case msg := <-twoCh:
			fmt.Printf("two recv %s\n", msg)
			ret.Two = msg
			isTwoRecved = true
		case msg := <-threeCh:
			fmt.Printf("three recv %s\n", msg)
			ret.Three = msg
			isThreeRecved = true
		case <-ctx.Done():
			return
		}
	}
}
