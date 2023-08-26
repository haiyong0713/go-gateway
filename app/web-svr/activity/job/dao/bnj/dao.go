package bnj

import (
	"go-gateway/app/web-svr/activity/tools/lib/initialize"
	"time"

	"go-common/library/cache/memcache"
	"go-common/library/database/sql"
	"go-common/library/net/http/blademaster"
	"go-common/library/queue/databus"
	"go-gateway/app/web-svr/activity/job/conf"
)

// Dao .
type Dao struct {
	c                *conf.Config
	db               *sql.DB
	client           *blademaster.Client
	comicClient      *blademaster.Client
	mc               *memcache.Memcache
	liveItemPub      *databus.Databus
	messagePub       *databus.Databus
	broadcastURL     string
	messageURL       string
	msgKeyURL        string
	normalMsgURL     string
	comicCouponURL   string
	mallCouponURL    string
	timeFinishExpire int32
	lessTimeExpire   int32
}

// New .
func New(c *conf.Config) *Dao {
	d := &Dao{
		c:           c,
		db:          sql.NewMySQL(c.MySQL.Like),
		client:      blademaster.NewClient(c.HTTPClient),
		comicClient: blademaster.NewClient(c.HTTPClientComic),
		mc:          memcache.New(c.Memcache.Like),
		liveItemPub: initialize.NewDatabusV1(c.LiveItemPub),
		messagePub:  initialize.NewDatabusV1(c.MessageDatabusPub),
	}
	d.broadcastURL = d.c.Host.APICo + _broadURL
	d.messageURL = d.c.Host.MsgCo + _messageURL
	d.msgKeyURL = d.c.Host.ApiVcCo + _msgKeyURI
	d.normalMsgURL = d.c.Host.ApiVcCo + _sendMsgURI
	d.comicCouponURL = d.c.Host.Comic + _comicCouponURI
	d.mallCouponURL = c.Host.Mall + _mallCouponURI
	d.timeFinishExpire = int32(time.Duration(c.Memcache.TimeFinishExpire) / time.Second)
	d.lessTimeExpire = int32(time.Duration(c.Memcache.LessTimeExpire) / time.Second)
	return d
}
