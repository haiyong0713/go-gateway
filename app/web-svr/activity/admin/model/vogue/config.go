package vogue

/**
CREATE TABLE `act_vogue` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `name` varchar(150) NOT NULL DEFAULT '' COMMENT '键',
  `config` varchar(1000) NOT NULL DEFAULT '' COMMENT '值',
  `ctime` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `mtime` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `ix_name` (`name`) USING BTREE,
  KEY `ix_mtime` (`mtime`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8 COMMENT='活动配置表';
*/

type ConfigItem struct {
	Id     int    `json:"-" gorm:"column:id"`
	Name   string `json:"name" gorm:"column:name"`
	Config string `json:"config" gorm:"column:config"`
}

type ConfigResponse struct {
	InviteScore    string `json:"invite_score" form:"invite_score" gorm:"column:invite_score" validate:"required"`
	ViewScore      string `json:"view_score" form:"view_score" gorm:"column:view_score" validate:"required"`
	ActStart       string `json:"act_start" form:"act_start" gorm:"column:act_start" validate:"required"`
	ActEnd         string `json:"act_end" form:"act_end" gorm:"column:act_end" validate:"required"`
	ActDoubleStart string `json:"act_double_start" form:"act_double_start" gorm:"column:act_double_start" validate:"required"`
	ActDoubleEnd   string `json:"act_double_end" form:"act_double_end" gorm:"column:act_double_end" validate:"required"`
	ScoreList      string `json:"score_list" form:"score_list" gorm:"column:score_list" validate:"required"`
	PlayList       string `json:"play_list" form:"play_list" gorm:"column:play_list" validate:"required"`
	TodayLimit     string `json:"today_limit" form:"today_limit" gorm:"column:tody_limit" validate:"required"`
}

type CritItem struct {
	Num  int64
	Min  int64
	Max  int64
	Show bool
}

type ConfigCreditLimit struct {
	DailyLimit           int64
	ActDoubleStart       int64
	ActDoubleEnd         int64
	ActSecondDoubleStart int64
	ActSecondDoubleEnd   int64
}

type CritList = []CritItem

func (*ConfigItem) TableName() string {
	return "act_vogue"
}
