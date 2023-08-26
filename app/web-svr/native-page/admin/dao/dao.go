package dao

import (
	"context"

	"github.com/jinzhu/gorm"
	"go-common/library/database/orm"
	xhttp "go-common/library/net/http/blademaster"

	"go-gateway/app/web-svr/native-page/admin/conf"
)

// Dao struct user of Dao.
type Dao struct {
	c                   *conf.Config
	DB                  *gorm.DB
	client              *xhttp.Client
	gameClient          *xhttp.Client
	actAdminClient      *xhttp.Client
	gameInfoURL         string
	gameListURL         string
	ComicInfosURL       string
	addActSubjectURL    string
	updateActSubjectURL string
}

// New create a instance of Dao and return.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:                   c,
		DB:                  orm.NewMySQL(c.ORM),
		client:              xhttp.NewClient(c.HTTPClient),
		gameClient:          xhttp.NewClient(c.HTTPGameClient),
		actAdminClient:      xhttp.NewClient(c.HTTPActAdminClient),
		gameInfoURL:         c.Host.GameCo + _gameURI,
		gameListURL:         c.Host.GameCo + _gameListURI,
		ComicInfosURL:       c.Host.ManGaCo + _comicInfosURI,
		addActSubjectURL:    c.Host.ActAdmin + _addActSubjectURI,
		updateActSubjectURL: c.Host.ActAdmin + _updateActSubjectURI,
	}
	d.initORM()
	return
}

func (d *Dao) initORM() {
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		if defaultTableName == "act_matchs" {
			return defaultTableName
		}
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
