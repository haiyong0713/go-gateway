package native

import (
	"context"

	acccli "git.bilibili.co/bapis/bapis-go/account/service"
	actGRPC "git.bilibili.co/bapis/bapis-go/activity/service"
	dynvotegrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/vote"
	chaGRPC "git.bilibili.co/bapis/bapis-go/hmt-channel/interface"
	spaceGRPC "git.bilibili.co/bapis/bapis-go/space/service"
	"go-common/library/cache/credis"
	"go-common/library/database/orm"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/web-svr/native-page/admin/conf"

	tagrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"

	"github.com/jinzhu/gorm"
)

// Dao struct user of Dao.
type Dao struct {
	c             *conf.Config
	DB            *gorm.DB
	client        *bm.Client
	tagGRPC       tagrpc.TagRPCClient
	redis         credis.Redis
	accGRPC       acccli.AccountClient
	spaceGRPC     spaceGRPC.SpaceClient
	actClient     actGRPC.ActivityClient
	chaClient     chaGRPC.ChannelRPCClient
	dynvoteClient dynvotegrpc.VoteSvrClient
}

// New create a instance of Dao and return.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:      c,
		DB:     orm.NewMySQL(c.ORM),
		client: bm.NewClient(c.HTTPClient),
		redis:  credis.NewRedis(c.Redis.Config),
	}
	d.initORM()
	var err error
	if d.tagGRPC, err = tagrpc.NewClient(c.TagGRPC); err != nil {
		panic(err)
	}
	if d.accGRPC, err = acccli.NewClient(c.AccClient); err != nil {
		panic(err)
	}
	if d.spaceGRPC, err = spaceGRPC.NewClient(c.SpaceClient); err != nil {
		panic(err)
	}
	if d.actClient, err = actGRPC.NewClient(c.ActClient); err != nil {
		panic(err)
	}
	if d.chaClient, err = chaGRPC.NewClient(c.ChaClient); err != nil {
		panic(err)
	}
	if d.dynvoteClient, err = dynvotegrpc.NewClient(c.DynvoteClient); err != nil {
		panic(err)
	}
	return
}

func (d *Dao) initORM() {
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return defaultTableName
	}
	d.DB.LogMode(true)
}

// Ping check connection of db , mc.
func (d *Dao) Ping(c context.Context) (err error) {
	if d.DB != nil {
		err = d.DB.DB().PingContext(c)
	}
	return
}

// Close close connection of db , mc.
func (d *Dao) Close() {
	if d.DB != nil {
		d.DB.Close()
	}
}
