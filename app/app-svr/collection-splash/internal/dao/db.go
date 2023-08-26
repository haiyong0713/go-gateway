package dao

import (
	"context"

	"go-common/library/conf/paladin.v2"
	xsql "go-common/library/database/sql"
	pb "go-gateway/app/app-svr/collection-splash/api"
)

func NewDB() (db *xsql.DB, cf func(), err error) {
	var (
		cfg xsql.Config
		ct  paladin.TOML
	)
	if err = paladin.Get("db.toml").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("Client").UnmarshalTOML(&cfg); err != nil {
		return
	}
	db = xsql.NewMySQL(&cfg)
	cf = func() { db.Close() }
	return
}

/*
CREATE TABLE `splash_collection_screen_images` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
  `img_name` varchar(32) NOT NULL DEFAULT '' COMMENT '图片名称',
  `mode` int(11) NOT NULL DEFAULT '1' COMMENT '物料类型。1为半屏物料，2为全屏物料',
  `img_url` varchar(1024) NOT NULL DEFAULT '' COMMENT '半屏图片地址',
  `img_url_normal` varchar(1024) NOT NULL DEFAULT '' COMMENT '常规屏端图片。只有mode=全屏物料时可编辑',
  `img_url_full` varchar(1024) NOT NULL DEFAULT '' COMMENT '全面屏端图片。只有mode=全屏物料时可编辑',
  `img_url_pad` varchar(1024) NOT NULL DEFAULT '' COMMENT 'PAD端图片。只有mode=全屏物料时可编辑',
  `logo_hide` tinyint(4) unsigned NOT NULL DEFAULT '0' COMMENT '是否隐藏LOGO。1-隐藏，0-不隐藏',
  `logo_mode` int(11) NOT NULL DEFAULT '1' COMMENT 'LOGO图。1为粉色LOGO，2为白色LOGO，3为自定义',
  `logo_img_url` varchar(1024) NOT NULL DEFAULT '' COMMENT 'LOGO图资源URL。只有当logo_mode=自定义时必需',
  `ctime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `mtime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  `is_deleted` tinyint(4) unsigned NOT NULL DEFAULT '0' COMMENT '0-未删除 1-已删除',
  PRIMARY KEY (`id`),
  KEY `ix_mtime` (`mtime`)
) ENGINE = InnoDB COMMENT = '典藏闪屏配置的图片物料';
*/

const (
	_addSplashSql = "INSERT INTO splash_collection_screen_images(img_name,mode,img_url_normal,img_url_full," +
		"img_url_pad,logo_hide) VALUE (?,?,?,?,?,?)"
	_fullMode = 2
	_hideLogo = 1
)

func (d *dao) AddSplash(ctx context.Context, param *pb.AddSplashReq) (int64, error) {
	result, err := d.db.Exec(ctx, _addSplashSql, param.ImgName, _fullMode, param.ImgUrlNormal,
		param.ImgUrlFull, param.ImgUrlPad, _hideLogo)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

const _updateSplashSql = "UPDATE splash_collection_screen_images SET img_name=?,img_url_normal=?," +
	"img_url_full=?,img_url_pad=? WHERE id=?"

func (d *dao) UpdateSplash(ctx context.Context, param *pb.UpdateSplashReq) (int64, error) {
	result, err := d.db.Exec(ctx, _updateSplashSql, param.ImgName, param.ImgUrlNormal, param.ImgUrlFull,
		param.ImgUrlPad, param.Id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const _deleteSplashSql = "UPDATE splash_collection_screen_images SET is_deleted=1 WHERE id=?"

func (d *dao) DeleteSplash(ctx context.Context, param *pb.SplashReq) (int64, error) {
	result, err := d.db.Exec(ctx, _deleteSplashSql, param.Id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const _splashSql = "SELECT id,img_name,mode,img_url,img_url_normal,img_url_full," +
	"img_url_pad,logo_hide,logo_mode,logo_img_url,is_deleted FROM splash_collection_screen_images WHERE id=?"

func (d *dao) Splash(ctx context.Context, param *pb.SplashReq) (*pb.Splash, error) {
	row := d.db.QueryRow(ctx, _splashSql, param.Id)
	splash := &pb.Splash{}
	err := row.Scan(&splash.Id, &splash.ImgName, &splash.Mode, &splash.ImgUrl, &splash.ImgUrlNormal, &splash.ImgUrlFull,
		&splash.ImgUrlPad, &splash.LogoHide, &splash.LogoMode, &splash.LogoImgUrl, &splash.IsDeleted)
	if err == xsql.ErrNoRows {
		return splash, nil
	}
	if err != nil {
		return nil, err
	}
	return splash, nil
}

const _splashAllSql = "SELECT id,img_name,mode,img_url,img_url_normal,img_url_full," +
	"img_url_pad,logo_hide,logo_mode,logo_img_url,is_deleted FROM splash_collection_screen_images"

func (d *dao) RawSplashList(ctx context.Context) ([]*pb.Splash, error) {
	rows, err := d.db.Query(ctx, _splashAllSql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]*pb.Splash, 0)
	for rows.Next() {
		splash := &pb.Splash{}
		if err := rows.Scan(&splash.Id, &splash.ImgName, &splash.Mode, &splash.ImgUrl, &splash.ImgUrlNormal, &splash.ImgUrlFull,
			&splash.ImgUrlPad, &splash.LogoHide, &splash.LogoMode, &splash.LogoImgUrl, &splash.IsDeleted); err != nil {
			return nil, err
		}
		if splash.GetIsDeleted() {
			continue
		}
		out = append(out, splash)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
