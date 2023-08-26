package mogul

import (
	"context"

	"go-common/library/database/sql"

	"go-gateway/app/app-svr/app-job/job/conf"
	mogulmdl "go-gateway/app/app-svr/app-job/job/model/mogul"
)

// CREATE TABLE `app_mogul_log` (
// 	`id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
// 	`mid` bigint(11) NOT NULL DEFAULT 0 COMMENT '用户mid',
// 	`buvid` varchar(50) NOT NULL DEFAULT '' COMMENT 'Buvid',
// 	`path` varchar(100)NOT NULL DEFAULT '' COMMENT '接口请求路径',
// 	`method` varchar(10)NOT NULL DEFAULT '' COMMENT '接口请求方法',
// 	`header` varchar(5000)NOT NULL DEFAULT '' COMMENT'接口请求header',
// 	`param` varchar(3000)NOT NULL DEFAULT '' COMMENT'接口请求参数',
// 	`body`varchar(10000)NOT NULL DEFAULT '' COMMENT'接口请求body',
// 	`response_header` varchar(5000)NOT NULL DEFAULT '' COMMENT'接口响应header',
// 	`response` MEDIUMTEXT COMMENT '接口响应',
// 	`status_code` varchar(10) NOT NULL DEFAULT '' COMMENT '接口HTTP响应状态码',
// 	`err_code` varchar(10) NOT NULL DEFAULT '' COMMENT '接口响应错误码',
// 	`request_time` TIMESTAMP NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT '请求时间',
// 	`duration` bigint(20) NOT NULL DEFAULT '0' COMMENT '接口响应时长',
// 	`ctime` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
// 	`mtime` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
// 	PRIMARY KEY (`id`),
// 	KEY `ix_mid_request_time` (`mid`,`request_time`),
// 	KEY `ix_buvid` (`buvid`),
// 	KEY `ix_mtime` (`mtime`)
// 	) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='app网关接口大佬行为日志';

const _inAppMogulLogSQL = "INSERT INTO app_mogul_log (mid,buvid,path,method,header,param,body,response_header,response,status_code,err_code,request_time,duration) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)"

type Dao struct {
	c  *conf.Config
	db *sql.DB
}

func New(c *conf.Config) *Dao {
	d := &Dao{
		c:  c,
		db: sql.NewMySQL(c.MySQL.Manager),
	}
	return d
}

func (d *Dao) AddAppUserLog(ctx context.Context, m *mogulmdl.AppMogulLog) error {
	_, err := d.db.Exec(ctx, _inAppMogulLogSQL, m.Mid, m.Buvid, m.Path, m.Method, m.Header, m.Param, m.Body, m.ResponseHeader, m.Response, m.StatusCode, m.ErrCode, m.RequestTime, m.Duration)
	return err
}
