CREATE TABLE `rewards_activity_config` (
`id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '活动ID',
`is_deleted` tinyint(4) NOT NULL DEFAULT '0' COMMENT '0 未删除 1 已删除',
`name` varchar(50) NOT NULL DEFAULT '' COMMENT '活动名称',
`notify_sender_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '通知发送方ID',
`notify_message` varchar(100) NOT NULL DEFAULT '' COMMENT '通知模板',
`notify_jump_url` varchar(1000) NOT NULL DEFAULT '' COMMENT '通知跳转链接',
`ctime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
`mtime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
PRIMARY KEY (`id`),
KEY `ix_mtime` (`mtime`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8 COMMENT='奖励活动信息表';


CREATE TABLE `rewards_award_config` (
`id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '配置ID',
`activity_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '关联活动ID',
`is_deleted` tinyint(4) NOT NULL DEFAULT '0' COMMENT '0 未删除 1 已删除',
`award_type` varchar(20) NOT NULL DEFAULT '' COMMENT '奖励类型',
`display_name` varchar(20) NOT NULL DEFAULT '' COMMENT '展示名称',
`icon_url` varchar(100) NOT NULL DEFAULT '' COMMENT '奖品图标',
`notify_sender_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '通知发送方ID',
`notify_message` varchar(100) NOT NULL DEFAULT '' COMMENT '通知模板',
`notify_jump_url` varchar(1000) NOT NULL DEFAULT '' COMMENT '通知跳转链接',
`config_content` varchar(10000) NOT NULL DEFAULT '' COMMENT '配置内容',
`ctime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
`mtime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
`extra_info` varchar(2000) NOT NULL DEFAULT '' COMMENT '自定义tag',
PRIMARY KEY (`id`),
KEY `ix_mtime` (`mtime`)
) ENGINE=InnoDB AUTO_INCREMENT=18 DEFAULT CHARSET=utf8 COMMENT='奖励奖品信息表';


CREATE TABLE `rewards_award_record_01` (
`mid` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '用户id',
`activity_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '关联活动ID',
`unique_id` varchar(50) NOT NULL DEFAULT '0' COMMENT '幂等ID',
`state` tinyint(4) NOT NULL DEFAULT '0' COMMENT '0 开始领取 1 领取完成 2 领取失败',
`award_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '奖励id',
`award_type` varchar(20) NOT NULL DEFAULT '0' COMMENT '奖励类型',
`award_name` varchar(50) NOT NULL DEFAULT '0' COMMENT '奖励名称',
`award_config_content` varchar(10000) NOT NULL DEFAULT '' COMMENT '奖励配置内容',
`business` varchar(50) NOT NULL DEFAULT '0' COMMENT '业务标识',
`ctime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
`mtime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
PRIMARY KEY (`mid`,`activity_id`,`unique_id`),
KEY `ix_mtime` (`mtime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='奖品用户发放记录表';

CREATE TABLE rewards_award_record_02 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_03 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_04 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_05 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_06 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_07 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_08 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_09 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_10 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_11 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_12 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_13 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_14 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_15 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_16 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_17 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_18 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_19 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_20 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_21 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_22 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_23 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_24 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_25 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_26 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_27 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_28 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_29 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_30 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_31 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_32 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_33 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_34 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_35 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_36 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_37 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_38 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_39 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_40 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_41 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_42 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_43 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_44 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_45 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_46 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_47 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_48 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_49 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_50 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_51 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_52 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_53 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_54 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_55 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_56 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_57 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_58 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_59 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_60 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_61 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_62 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_63 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_64 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_65 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_66 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_67 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_68 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_69 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_70 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_71 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_72 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_73 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_74 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_75 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_76 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_77 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_78 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_79 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_80 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_81 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_82 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_83 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_84 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_85 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_86 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_87 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_88 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_89 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_90 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_91 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_92 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_93 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_94 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_95 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_96 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_97 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_98 LIKE rewards_award_record_01;
CREATE TABLE rewards_award_record_99 LIKE rewards_award_record_01;